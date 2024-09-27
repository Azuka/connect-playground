package telemetry

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/MrAlias/redact"
	pkgerrors "github.com/pkg/errors"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/bridge/opencensus"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.25.0"
)

const shutDownTimeout = time.Second * 10

func Build(serviceName string) func() {

	// Setup opencensus tracer
	defer func() {
		opencensus.InstallTraceBridge()
	}()

	http.DefaultClient.Transport = otelhttp.NewTransport(http.DefaultTransport)

	client := otlptracehttp.NewClient()
	exp, err := otlptrace.New(context.Background(), client)
	if err != nil {
		return func() {
			log.Println("Unable to initialize opentelemetry", pkgerrors.WithStack(err))
		}
	}

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		b3.New(),
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// Create and install Jaeger export pipeline.
	tp := tracesdk.NewTracerProvider(
		redact.Attributes("key"),
		// Always be sure to batch in production.
		tracesdk.WithBatcher(exp),
		// Record information about this application in an Resource.
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
		)),
	)

	otel.SetTracerProvider(tp)

	return func() {
		ctx, cancel := context.WithTimeout(context.Background(), shutDownTimeout)
		defer cancel()
		_ = tp.Shutdown(ctx) // nolint
	}
}
