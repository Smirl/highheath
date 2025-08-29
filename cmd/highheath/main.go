package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/schema"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/smirl/highheath/pkg/highheath"
)

func init() {
	log.Println("Instrumenting the default HTTP transport")
	http.DefaultTransport = otelhttp.NewTransport(http.DefaultTransport)
}

func main() {
	// Setup a context which will be notified on SIGINT to capture user interupts
	// when going through the authorization flow
	ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	// Create an AppContext with the necessary dependencies.
	// AppContect implements http.Handler
	app := highheath.AppContext{
		ServeMux:     http.NewServeMux(),
		Decoder:      schema.NewDecoder(),
		GmailClient:  highheath.GmailClient(ctx),
		GithubClient: highheath.GithubClient(),
		Recaptcha:    highheath.NewRecaptcha(),
	}
	app.Decoder.IgnoreUnknownKeys(true)
	// Register the handlers with the AppContext.
	app.ServeMux.HandleFunc("/health", highheath.HandleHealth)
	app.HandleFunc("/api/booking", highheath.HandleBookingForm)
	app.HandleFunc("/api/contact", highheath.HandleContactForm)
	app.HandleFunc("/api/comment", highheath.HandleCommentForm)
	app.ServeMux.Handle("/", http.FileServer(http.Dir("./public")))

	server := &http.Server{
		Handler: app.WrapHandler(),
		Addr:    "0.0.0.0:8080",
		// Good practice: enforce timeouts
		WriteTimeout:      15 * time.Second,
		ReadTimeout:       15 * time.Second,
		IdleTimeout:       30 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
	}

	// Set up OpenTelemetry.
	otelShutdown, err := highheath.SetupOTEL(ctx)
	if err != nil {
		log.Fatalf("SetupOTEL error: %v", err)
	}
	defer otelShutdown(ctx)

	// Start the server
	go func() {
		log.Println("Starting server on 0.0.0.0:8080")
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe error: %v", err)
		}
	}()

	// Wait
	<-ctx.Done()

	// Gracefully (15s) shutdown servers
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()
	log.Println("Shutdown down servers with 15s timeout")
	err = server.Shutdown(ctx)
	if err != nil {
		log.Fatalf("Http Server Shutdown error: %v", err)
	}
	log.Println("Shutdown complete")
}
