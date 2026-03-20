package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
)

type response struct {
	Version    string `json:"version"`
	User       string `json:"user"`
	ExtraField bool   `json:"extra_field,omitempty"`
}

func main() {
	port := flag.String("port", "3000", "port to listen on")
	mode := flag.String("mode", "live", "response mode: live or shadow")
	flag.Parse()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		user := r.Header.Get("X-User-ID")
		if user == "" {
			user = "anonymous"
		}

		payload := response{
			Version: *mode,
			User:    user,
		}
		if *mode == "shadow" {
			payload.ExtraField = true
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(payload); err != nil {
			http.Error(w, "failed to write response", http.StatusInternalServerError)
		}
	})

	log.Printf("test server (%s) on :%s", *mode, *port)
	if err := http.ListenAndServe(":"+*port, nil); err != nil {
		log.Fatal(err)
	}
}
