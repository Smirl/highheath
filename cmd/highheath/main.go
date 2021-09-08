package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/schema"

	"github.com/smirl/highheath/pkg/highheath"
)

func main() {
	handler := http.NewServeMux()
	wrappedHandler := handlers.RecoveryHandler()(handlers.ProxyHeaders(handlers.CombinedLoggingHandler(os.Stdout, handler)))
	server := &http.Server{
		Handler: wrappedHandler,
		Addr:    "0.0.0.0:8080",
		// Good practice: enforce timeouts
		WriteTimeout:      15 * time.Second,
		ReadTimeout:       15 * time.Second,
		IdleTimeout:       30 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
	}
	ac := highheath.AppContext{
		Decoder:      schema.NewDecoder(),
		GmailClient:  highheath.GmailClient(),
		GithubClient: highheath.GithubClient(),
		Recaptcha:    highheath.NewRecaptcha(),
	}
	ac.Decoder.IgnoreUnknownKeys(true)
	handler.HandleFunc("/health", highheath.HandleHealth)
	handler.Handle("/api/booking", &highheath.Handler{AppContext: ac, H: highheath.HandleBookingForm})
	handler.Handle("/api/contact", &highheath.Handler{AppContext: ac, H: highheath.HandleContactForm})
	handler.Handle("/api/comment", &highheath.Handler{AppContext: ac, H: highheath.HandleCommentForm})
	handler.Handle("/", http.FileServer(http.Dir("./public")))
	log.Println("Starting server on 0.0.0.0:8080")
	server.ListenAndServe()
}
