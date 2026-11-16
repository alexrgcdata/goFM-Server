package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"gofm-server/internal/api"
)

func main() {
	path := flag.String("config", "config.local.json", "private JSON configuration file")
	dashboard := flag.String("dashboard", "", "optional directory containing the built admin dashboard")
	init := flag.Bool("init-config", false, "create a private configuration with a random token, then exit")
	flag.Parse()

	if *init {
		if err := initConfig(*path); err != nil {
			log.Printf("initialize config: %v", err)
			os.Exit(1)
		}
		log.Print("Private config created. Read its token locally; existing files are never overwritten.")
		return
	}
	if err := run(*path, *dashboard); err != nil {
		log.Printf("server: %v", err)
		os.Exit(1)
	}
}

func initConfig(path string) error {
	token := make([]byte, 32)
	if _, err := rand.Read(token); err != nil {
		return err
	}
	config := api.Config{
		Address:         "127.0.0.1:8080",
		Token:           hex.EncodeToString(token),
		MaxRequestBytes: 65536,
	}

	// Exclusive creation leaves an existing token in place.

	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")
	writeErr := encoder.Encode(config)
	closeErr := f.Close()
	return errors.Join(writeErr, closeErr)
}

func run(path string, dashboard string) error {
	config, err := api.LoadConfig(path)
	if err != nil {
		return err
	}
	handler, err := buildHandler(config)
	if err != nil {
		return err
	}
	mux := http.NewServeMux()
	mux.Handle("/", handler)
	if dashboard != "" {
		files := http.FileServer(http.Dir(dashboard))
		mux.Handle("/admin/dashboard/", http.StripPrefix("/admin/dashboard/", files))
		mux.HandleFunc("/admin/node/metrics", readNodeMetrics)
	}
	srv := &http.Server{
		Addr:              config.Address,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    16 << 10,
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	finished := make(chan struct{})
	defer close(finished)
	shutdownDone := make(chan struct{})
	go func() {
		defer close(shutdownDone)
		select {
		case <-ctx.Done():
			shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			if err := srv.Shutdown(shutdown); err != nil {
				_ = srv.Close()
			}
		case <-finished:
		}
	}()
	log.Printf("goFM dispatcher listening on %s", config.Address)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("listen: %w", err)
	}
	// ListenAndServe returns as soon as shutdown starts. Wait for in-flight
	// requests to drain before allowing the process to exit.
	<-shutdownDone
	return nil
}

func buildHandler(config api.Config) (*api.Server, error) {
	dsn := os.Getenv("WORKSHOP_POSTGRES_DSN")
	if dsn == "" {
		return api.New(config)
	}

	database, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, errors.New("invalid Workshop PostgreSQL connection settings")
	}
	// Keep the pool small on the Workshop VPS.
	database.SetMaxOpenConns(2)
	database.SetMaxIdleConns(1)
	return api.NewWithDemoJobs(config, database)
}
