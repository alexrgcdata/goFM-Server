package main

import (
	"flag"
	"log"
	"net/http"
	"os"
<<<<<<< HEAD
	"time"
=======
>>>>>>> 0d8871589256cb66840deaca1805331e2759ccc6

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

<<<<<<< HEAD
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
=======
	log.Printf("goFM Server listening on %s", config.Address)
	if err := http.ListenAndServe(config.Address, server.New(config)); err != nil {
>>>>>>> 0d8871589256cb66840deaca1805331e2759ccc6
		log.Printf("server stopped: %v", err)
		os.Exit(1)
	}
}
