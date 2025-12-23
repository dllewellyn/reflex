package telemetry

import (
	"context"
	"fmt"
	"io"
	"time"

	mexporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/metric"
	texporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// NewTracerProvider creates a new tracer provider.
func NewTracerProvider(ctx context.Context, projectID, serviceName string, w io.Writer) (*trace.TracerProvider, error) {
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			"",
			semconv.ServiceNameKey.String(serviceName),
		),
	)
	if err != nil {
		return nil, err
	}

	var exporter trace.SpanExporter
	if projectID != "" {
		exporter, err = texporter.New(texporter.WithProjectID(projectID))
		if err != nil {
			return nil, fmt.Errorf("failed to create google cloud trace exporter: %w", err)
		}
	} else {
		exporter, err = stdouttrace.New(stdouttrace.WithWriter(w), stdouttrace.WithPrettyPrint())
		if err != nil {
			return nil, err
		}
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(res),
	)
	return tp, nil
}

// NewMeterProvider creates a new meter provider.
func NewMeterProvider(projectID string) (*metric.MeterProvider, error) {
	if projectID == "" {
		return nil, nil
	}

	exporter, err := mexporter.New(mexporter.WithProjectID(projectID))
	if err != nil {
		return nil, fmt.Errorf("failed to create google cloud metric exporter: %w", err)
	}

	mp := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(exporter, metric.WithInterval(60*time.Second))),
	)
	return mp, nil
}

// SetupTracer initializes and registers a global tracer provider.
func SetupTracer(ctx context.Context, projectID, serviceName string, w io.Writer) (func(), error) {
	tp, err := NewTracerProvider(ctx, projectID, serviceName, w)
	if err != nil {
		return nil, err
	}
	otel.SetTracerProvider(tp)

	mp, err := NewMeterProvider(projectID)
	if err != nil {
		return nil, err
	}
	if mp != nil {
		otel.SetMeterProvider(mp)
	}

	cleanup := func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			otel.Handle(err)
		}
		if mp != nil {
			if err := mp.Shutdown(context.Background()); err != nil {
				otel.Handle(err)
			}
		}
	}
	return cleanup, nil
}
