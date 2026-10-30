package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"gofm-server/internal/server"
)

func main() {
	configPath := flag.String("config", "config.json", "path to the JSON configuration file")
	flag.Parse()

	config, err := server.LoadConfig(*configPath)
	if err != nil {
		log.Printf("configuration error: %v", err)
		os.Exit(1)
	}

	httpServer := &http.Server{
		Addr:              config.Address,
		Handler:           server.New(config),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       90 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}
	log.Printf("goFM Server listening on %s", config.Address)
	if err := httpServer.ListenAndServe(); err != nil {
		log.Printf("server stopped: %v", err)
		os.Exit(1)
	}
}
