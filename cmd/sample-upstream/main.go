package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"time"
)

func main() {
	address := flag.String("addr", ":3001", "listen address")
	name := flag.String("name", "sample-api", "sample service name")
	flag.Parse()
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"service": *name, "method": r.Method, "path": r.URL.Path, "received": true, "message": "sample upstream reached"})
	})
	mux.HandleFunc("/inventory", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"service": *name, "records": []map[string]any{{"sku": "DEMO-001", "quantity": 12}, {"sku": "DEMO-002", "quantity": 4}}})
	})
	mux.HandleFunc("/slow", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(150 * time.Millisecond)
		_ = json.NewEncoder(w).Encode(map[string]string{"service": *name, "status": "slow-but-successful"})
	})
	log.Printf("sample upstream %s listening on %s", *name, *address)
	if err := http.ListenAndServe(*address, mux); err != nil {
		log.Fatal(err)
	}
}
