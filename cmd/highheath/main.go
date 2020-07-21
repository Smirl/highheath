package main

import (
	"log"
	"net/http"
	"time"

	"github.com/smirl/highheath/pkg/highheath"
)

func main() {
	handler := http.NewServeMux()
	server := &http.Server{
		Handler: highheath.LogRequest(handler),
		Addr:    "0.0.0.0:8080",
		// Good practice: enforce timeouts
		WriteTimeout:      15 * time.Second,
		ReadTimeout:       15 * time.Second,
		IdleTimeout:       30 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
	}

	fs := http.FileServer(http.Dir("./public"))
	handler.Handle("/", fs)
	handler.HandleFunc("/contact", highheath.HandleContactForm)
	log.Println("Starting server on 0.0.0.0:8080")
	server.ListenAndServe()
}
