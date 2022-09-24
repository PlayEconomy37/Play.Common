package opentelemetry

import (
	"log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

// Sets up Opentelemetry with a Jaeger exporter or an in-memory exporter (when running tests)
func SetupTracer(isTest bool) *sdktrace.TracerProvider {
	var tracerProvider *sdktrace.TracerProvider

	if isTest {
		// Setup in-memory exporter
		inMemoryExporter := tracetest.NewInMemoryExporter()

		tracerProvider = sdktrace.NewTracerProvider(
			sdktrace.WithSampler(sdktrace.AlwaysSample()),
			sdktrace.WithSyncer(inMemoryExporter),
			sdktrace.WithResource(resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceNameKey.String("catalog"),
			)),
		)
	} else {
		// Setup Jaeger
		jaegerExporter, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint("http://localhost:14268/api/traces")))
		if err != nil {
			log.Fatalf("could not initialize Jaeger exporter: %v", err)
		}

		tracerProvider = sdktrace.NewTracerProvider(
			sdktrace.WithSampler(sdktrace.AlwaysSample()),
			sdktrace.WithBatcher(jaegerExporter),
			sdktrace.WithResource(resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceNameKey.String("catalog"),
			)),
		)
	}

	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	return tracerProvider
}
