package main

import (
	"flag"
	"log"
	"net/http"
	"os"

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

	log.Printf("goFM Server listening on %s", config.Address)
	if err := http.ListenAndServe(config.Address, server.New(config)); err != nil {
		log.Printf("server stopped: %v", err)
		os.Exit(1)
	}
}
