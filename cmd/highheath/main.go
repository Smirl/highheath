package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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

	// Setup a context which will be notified on SIGINT to capture user interupts
	// when going through the authorization flow
	ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Println("Starting server on 0.0.0.0:8080")
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe error: %v", err)
		}
	}()

	<-ctx.Done()

	// Gracefully (15s) shutdown servers
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()
	log.Println("Shutdown down servers with 15s timeout")
	err := server.Shutdown(ctx)
	if err != nil {
		log.Fatalf("Http Server Shutdown error: %v", err)
	}
	log.Println("Shutdown complete")
}
