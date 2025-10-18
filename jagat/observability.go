package jagat

import (
	"context"
	"errors"
	"fmt"
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

type Metric struct {
	// Enable     bool
	Stdout     bool
	Prometheus bool
	// if not set but enable will use stdout
	Endpoint string
	// secure endpoint (https)
	Secure bool
}

type MetricShutdownFn func(context.Context) error

func StartMetric(ctx context.Context, res *resource.Resource, mcfg Metric) (*metric.MeterProvider, error) {
	// noopShutdown := func(_ context.Context) error { return nil }
	if !mcfg.Prometheus && !mcfg.Stdout {
		slog.Info("Metric is disabled")
		return nil, nil
	}

	// res, err := resource.New(
	// 	ctx,
	// 	resource.WithAttributes(
	// 		semconv.ServiceNameKey.String(serviceName),
	// 	),
	// )
	// if err != nil {
	// 	err = fmt.Errorf("failed to create otel resource: %w", err)
	// 	return nil, err //func(context.Context) error { return nil }, err
	// }

	// --- METER PROVIDER ---

	var metricExporter metric.Exporter
	var err error
	// prometheus exporter
	if mcfg.Prometheus {
		opts := []otlpmetrichttp.Option{}

		// endpoint setup
		opts = append(opts, otlpmetrichttp.WithEndpoint(mcfg.Endpoint))
		if !mcfg.Secure {
			opts = append(opts, otlpmetrichttp.WithInsecure())
		}

		metricExporter, err = otlpmetrichttp.New(ctx, opts...)
		if err != nil {
			err = fmt.Errorf("failed to create otlp http metric exporter: %w", err)
			return nil, err
		}

	} else {
		slog.Debug("Initilize stdout metric")
		metricExporter, err = stdoutmetric.New()
		if err != nil {
			err = fmt.Errorf("failed to create metric exporter: %w", err)
			return nil, err
		}
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(metricExporter)),
		metric.WithResource(res),
	)

	return meterProvider, nil

	// otel.SetMeterProvider(meterProvider)

	// shutDownFunc := func(ctx context.Context) error {
	// 	if err := meterProvider.Shutdown(ctx); err != nil {
	// 		return err
	// 	}
	// 	return nil
	// }
	// slog.Info("Metric Observability Initialized", "endpoint", mcfg.Endpoint)
	// return shutDownFunc, nil
}

type Trace struct {
	// Enable     bool
	Stdout bool
	Jaeger bool
	// if not set but enable will use stdout
	Endpoint string
	// secure endpoint (https)
	Secure bool
}

type TraceShutdownFn func(context.Context) error

func StartTrace(ctx context.Context, res *resource.Resource, tcfg Trace) (*trace.TracerProvider, error) {
	// noopShutdown := func(context.Context) error { return nil }
	// if !tcfg.Jaeger && !tcfg.Stdout {
	// 	slog.Info("tracer  disabled")
	// 	return nil, nil //noopShutdown, nil
	// }

	// --- TRACER PROVIDER ---

	// res, err := resource.New(
	// 	ctx,
	// 	resource.WithAttributes(
	// 		semconv.ServiceNameKey.String(serviceName),
	// 	),
	// )
	// if err != nil {
	// 	err = fmt.Errorf("failed to create otel resource: %w", err)
	// 	return nil, err //func(context.Context) error { return nil }, err
	// }

	var traceExporter trace.SpanExporter
	var err error
	if tcfg.Jaeger {
		//tracer option.
		otlpOpts := []otlptracehttp.Option{}
		if !tcfg.Secure {
			otlpOpts = append(otlpOpts, otlptracehttp.WithInsecure())
		}
		traceExporter, err = otlptracehttp.New(ctx, otlpOpts...)
		if err != nil {
			err = fmt.Errorf("failed to create otlp http trace exporter: %w", err)
			return nil, err
		}
	} else {
		slog.Debug("Initilize stdout trace")
		traceExporter, err = stdouttrace.New(stdouttrace.WithPrettyPrint())
		if err != nil {
			err = fmt.Errorf("failed to create trace exporter: %w", err)
			return nil, err
		}
	}

	tracerProvider := trace.NewTracerProvider(
		trace.WithBatcher(traceExporter),
		trace.WithResource(res),
	)

	return tracerProvider, nil
	// otel.SetTracerProvider(tracerProvider)
	// otel.SetTextMapPropagator(propagation.TraceContext{})

	// slog.Debug("Tracer exporter Initialized", "endpoint", tcfg.Endpoint)
	// // to ensure all telemetry data is flushed.
	// return func(ctx context.Context) error {
	// 	return tracerProvider.Shutdown(ctx)
	// }, nil
}

// Initializes and configures OpenTelemetry for the application.
// It returns a shutdown function that must be called on application exit.
func InitObservability(ctx context.Context, serviceName string, cfg Config) (shutdown func(context.Context) error, err error) {

	res, err := resource.New(
		ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
		),
	)
	if err != nil {
		err = fmt.Errorf("failed to create otel resource: %w", err)
		return func(context.Context) error { return nil }, err
	}

	sfn := []func(context.Context) error{}

	// --- METER PROVIDER ---
	if cfg.Metric.Prometheus {

		meterProvider, err := StartMetric(ctx, res, cfg.Metric)
		if err != nil {
			return nil, err
		}
		otel.SetMeterProvider(meterProvider)
		sfn = append(sfn, meterProvider.Shutdown)
	}

	// --- TRACER PROVIDER ---
	if cfg.Trace.Jaeger {
		tracerProvider, err := StartTrace(ctx, res, cfg.Trace)
		if err != nil {
			return nil, err
		}
		otel.SetTracerProvider(tracerProvider)
		sfn = append(sfn, tracerProvider.Shutdown)
	}

	// Set the global propagator to tracecontext.
	otel.SetTextMapPropagator(propagation.TraceContext{})

	// The returned shutdown function will be called on application exit
	// to ensure all telemetry data is flushed.
	return func(ctx context.Context) error {
		var shutdownErr error
		for _, fn := range sfn {
			if xerr := fn(ctx); err != nil {
				shutdownErr = errors.Join(err, xerr)
			}
		}
		return shutdownErr
	}, nil
}
