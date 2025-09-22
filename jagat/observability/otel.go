package observability

import (
	"context"
	"errors"
	"log/slog"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

type Config struct {
	Enable bool
	// if not set but enable will use stdout
	Exporter string
	// http endpoint exporter
	TraceEndpoint   string
	MetricsEndpoint string
	// secure endpoint (https)
	Secure bool
}

// Initializes and configures OpenTelemetry for the application.
// It returns a shutdown function that must be called on application exit.
func Init(ctx context.Context, serviceName string, cfg Config) (shutdown func(context.Context) error) {
	noopShutdown := func(context.Context) error { return nil }
	if !cfg.Enable {
		slog.Info("Observability is disabled")
		return noopShutdown
	}

	res, err := resource.New(
		ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
		),
	)
	if err != nil {
		slog.Error("failed to create otel resource", "error", err)
		return func(context.Context) error { return nil }
	}

	// --- TRACER PROVIDER ---
	var traceExporter trace.SpanExporter
	switch cfg.Exporter {
	case "http":
		slog.Info("Initializing Jaeger exporter", "endpoint", cfg.TraceEndpoint)

		//tracer option.
		otlpOpts := []otlptracehttp.Option{}
		otlpOpts = append(otlpOpts, otlptracehttp.WithEndpoint(cfg.TraceEndpoint))
		if !cfg.Secure {
			otlpOpts = append(otlpOpts, otlptracehttp.WithInsecure())
		}
		traceExporter, err = otlptracehttp.New(ctx, otlpOpts...)

		if err != nil {
			slog.Error("failed to create otlp http trace exporter ", "error", err)
			return noopShutdown
		}

	default:
		slog.Info("Initializing stdout exporter")
		traceExporter, err = stdouttrace.New(stdouttrace.WithPrettyPrint())
		if err != nil {
			slog.Error("failed to create trace exporter", "error", err)
			return func(context.Context) error { return nil }
		}

	}

	tracerProvider := trace.NewTracerProvider(
		trace.WithBatcher(traceExporter),
		trace.WithResource(res),
	)
	otel.SetTracerProvider(tracerProvider)

	// --- METER PROVIDER ---

	var metricExporter metric.Exporter
	switch cfg.Exporter {
	case "http":
		//metric option
		opts := []otlpmetrichttp.Option{}
		opts = append(opts, otlpmetrichttp.WithEndpoint(cfg.MetricsEndpoint))
		if !cfg.Secure {
			opts = append(opts, otlpmetrichttp.WithInsecure())
		}
		metricExporter, err = otlpmetrichttp.New(ctx, opts...)
		if err != nil {
			slog.Error("failed to create otlp http metric exporter", "error", err)
			return noopShutdown
		}

	default:
		metricExporter, err = stdoutmetric.New()
		if err != nil {
			slog.Error("failed to create metric exporter", "error", err)
			return func(context.Context) error { return nil }
		}
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(metricExporter)),
		metric.WithResource(res),
	)
	otel.SetMeterProvider(meterProvider)

	// Set the global propagator to tracecontext.
	otel.SetTextMapPropagator(propagation.TraceContext{})
	slog.Info("Observability initialized", "exporter", cfg.Exporter)

	// The returned shutdown function will be called on application exit
	// to ensure all telemetry data is flushed.
	return func(ctx context.Context) error {
		slog.Info("Shutting down observability providers...")
		var shutdownErr error
		if err := tracerProvider.Shutdown(ctx); err != nil {
			shutdownErr = errors.Join(shutdownErr, err)
		}
		if err := meterProvider.Shutdown(ctx); err != nil {
			shutdownErr = errors.Join(shutdownErr, err)
		}
		return shutdownErr
	}
}
