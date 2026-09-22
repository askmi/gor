package internal

import (
	"context"
	"log/slog"
	"net/http"

	gor "gor/pkg/server"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/otlptranslator"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	otelprometheus "go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

func SetupMeter() (*sdkmetric.MeterProvider, http.Handler, error) {
	registry := prometheus.NewRegistry()
	registerer := prometheus.WrapRegistererWith(
		prometheus.Labels{"app": AppName},
		registry,
	)
	// registerer.MustRegister(
	// collectors.NewGoCollector(),
	// collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	// )
	/*
		When Prometheus requests /metrics:
		1. promhttp.Handler() asks the registry for metrics.
		2. The registry calls the OTel collector’s Collect().
		3. The collector asks the OTel Reader for current measurements.
		4. The reader retrieves aggregated values from the MeterProvider.
		5. The collector converts them into Prometheus metrics.
		6. promhttp writes them as text in the HTTP response.
		In short:
		- Reader: reads metrics from OpenTelemetry.
		- Collector: gives metrics to Prometheus.
		- OTel Prometheus exporter: bridges the Reader and Collector.

		Internally, this approximately does:

			reader := sdkmetric.NewManualReader()
			collector := newPrometheusCollector(reader)
			prometheus.DefaultRegisterer.Register(collector)
	*/

	exporter, err := otelprometheus.New(
		otelprometheus.WithRegisterer(registerer),
		otelprometheus.WithTranslationStrategy(otlptranslator.UnderscoreEscapingWithoutSuffixes),
		otelprometheus.WithoutScopeInfo(),
		otelprometheus.WithoutTargetInfo(),
	)
	if err != nil {
		return nil, nil, err
	}

	provider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(exporter),
	)
	otel.SetMeterProvider(provider)

	handler := promhttp.InstrumentMetricHandler(
		registerer,
		promhttp.HandlerFor(registry, promhttp.HandlerOpts{}),
	)

	return provider, handler, nil
}

func SetupTracer() *sdktrace.TracerProvider {
	provider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(
			sdktrace.ParentBased(sdktrace.AlwaysSample()),
		),
	)

	// Set global provider
	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	return provider
}

type TraceLogHandler struct {
	slog.Handler
	traceIDBound bool
}

func hasAttr(record slog.Record, key string) bool {
	found := false

	record.Attrs(func(attr slog.Attr) bool {
		if attr.Key == key {
			found = true
			return false
		}
		return true
	})

	return found
}

func (h TraceLogHandler) Handle(ctx context.Context, record slog.Record) error {
	sc := trace.SpanContextFromContext(ctx)

	if sc.IsValid() && !h.traceIDBound {
		if !hasAttr(record, "trace_id") {
			record.AddAttrs(
				slog.String("trace_id", sc.TraceID().String()),
			)
		}
	}

	return h.Handler.Handle(ctx, record)
}

func (h TraceLogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	traceIDBound := h.traceIDBound
	for _, attr := range attrs {
		if attr.Key == "trace_id" {
			traceIDBound = true
			break
		}
	}

	return TraceLogHandler{
		Handler:      h.Handler.WithAttrs(attrs),
		traceIDBound: traceIDBound,
	}
}

func (h TraceLogHandler) WithGroup(name string) slog.Handler {
	return TraceLogHandler{
		Handler:      h.Handler.WithGroup(name),
		traceIDBound: h.traceIDBound,
	}
}

func UserCounter[Req, Resp any](f gor.RouterFunc[Req, Resp]) gor.RouterFunc[Req, Resp] {
	m := otel.Meter("users")
	c, err := m.Int64Counter("users_total")
	if err != nil {
		panic("can not create user.counter " + err.Error())
	}
	return func(ctx context.Context, req Req) (Resp, error) {
		resp, err := f(ctx, req)
		op := "success"
		if err != nil {
			op = "fail"
		}
		c.Add(ctx, 1, metric.WithAttributes(
			attribute.String("operation", "get_user"),
			attribute.String("status", op),
		))
		return resp, err
	}
}
