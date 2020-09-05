package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/handlers"

	"github.com/smirl/highheath/pkg/highheath"
)

func main() {
	handler := http.NewServeMux()
	wrappedHandler := handlers.RecoveryHandler()(handlers.CombinedLoggingHandler(os.Stdout, handler))
	server := &http.Server{
		Handler: wrappedHandler,
		Addr:    "0.0.0.0:8080",
		// Good practice: enforce timeouts
		WriteTimeout:      15 * time.Second,
		ReadTimeout:       15 * time.Second,
		IdleTimeout:       30 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
	}

	handler.HandleFunc("/health", highheath.HandleHealth)
	handler.HandleFunc("/api/booking", highheath.HandleBookingForm)
	handler.HandleFunc("/api/contact", highheath.HandleContactForm)
	handler.HandleFunc("/api/comment", highheath.HandleCommentForm)
	handler.Handle("/", http.FileServer(http.Dir("./public")))
	log.Println("Starting server on 0.0.0.0:8080")
	server.ListenAndServe()
}
