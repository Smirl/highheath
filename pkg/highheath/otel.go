package highheath

import (
	"context"
	"log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/trace"
)

// SetupOTEL bootstraps the OpenTelemetry pipeline.
// If it does not return an error, make sure to call shutdown for proper cleanup.
func SetupOTEL(ctx context.Context) (shutdown func(context.Context), err error) {
	// Set up propagator.
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// Set up trace provider.
	traceExporter, err := otlptracehttp.New(ctx)
	if err != nil {
		return nil, err
	}
	// Set up tracer provider.
	tracerProvider := trace.NewTracerProvider(trace.WithBatcher(traceExporter))
	otel.SetTracerProvider(tracerProvider)

	// Set up shutdown function.
	shutdown = func(ctx context.Context) {
		err := tracerProvider.Shutdown(ctx)
		if err != nil {
			log.Fatalf("OTEL shutdown error: %v", err)
		}
	}
	return shutdown, nil
}
