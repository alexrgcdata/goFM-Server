package main

import (
	"context"
	"crypto/rand"
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

	"gofm-server/internal/api"
)

func main() {
	path := flag.String("config", "config.local.json", "private JSON configuration file")
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
	if err := run(*path); err != nil {
		log.Printf("server: %v", err)
		os.Exit(1)
	}
}

func initConfig(path string) error {
	token := make([]byte, 32)
	if _, err := rand.Read(token); err != nil {
		return err
	}
	config := api.Config{Address: "127.0.0.1:8080", Token: hex.EncodeToString(token), MaxRequestBytes: 65536}
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

func run(path string) error {
	config, err := api.LoadConfig(path)
	if err != nil {
		return err
	}
	handler, err := api.New(config)
	if err != nil {
		return err
	}
	srv := &http.Server{
		Addr: config.Address, Handler: handler,
		ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second,
		WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 16 << 10,
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
