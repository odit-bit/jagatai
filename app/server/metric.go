package main

import (
	"context"
	"errors"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	otelMetric "go.opentelemetry.io/otel/sdk/metric"
)

func setupOtel(ctx context.Context) (func(ctx context.Context) error, error) {
	var shutdownFuncs []func(context.Context) error

	// shutdown calls cleanup functions registered via shutdownFuncs.
	// The errors from the calls are joined.
	// Each registered cleanup will be invoked once.
	shutdown := func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}
		shutdownFuncs = nil
		return err
	}

	meterProvider, err := setupMeterProvider()
	if err != nil {
		return nil, err
	}
	shutdownFuncs = append(shutdownFuncs, meterProvider.Shutdown)

	return shutdown, nil
}

func setupMeterProvider() (*otelMetric.MeterProvider, error) {
	exporter, err := prometheus.New()
	if err != nil {
		return nil, err
	}

	provider := otelMetric.NewMeterProvider(
		otelMetric.WithReader(exporter),
	)
	otel.SetMeterProvider(provider)
	return provider, nil
}

func RamUsage(ctx context.Context, cb func() int64) (metric.Int64ObservableGauge, error) {
	meter := otel.Meter("jagatAI_rest_server_meter")
	ramUsage, err := meter.Int64ObservableGauge(
		"jagatAI_ram_usage_bytes",
		metric.WithDescription("Ram usage of the app in bytes"),
		metric.WithInt64Callback(func(ctx context.Context, io metric.Int64Observer) error {
			io.Observe(cb())
			return nil
		}),
	)
	if err != nil {
		return nil, err
	}
	_ = ramUsage
	return ramUsage, nil
}
