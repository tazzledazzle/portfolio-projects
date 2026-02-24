"""
Minimal FastAPI app with OpenTelemetry (traces + metrics) and OTLP export.
Exposes /health and /api/health/otel for SRE verification.
"""
import os
from fastapi import FastAPI
from opentelemetry import trace, metrics
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.sdk.metrics import MeterProvider
from opentelemetry.sdk.metrics.export import PeriodicExportingMetricReader
from opentelemetry.sdk.resources import Resource
from opentelemetry.exporter.otlp.proto.http.trace_exporter import OTLPSpanExporter
from opentelemetry.exporter.otlp.proto.http.metric_exporter import OTLPMetricExporter
from opentelemetry.instrumentation.fastapi import FastAPIInstrumentor

SERVICE_NAME = os.environ.get("OTEL_SERVICE_NAME", "otel-demo-api")
OTLP_ENDPOINT = os.environ.get("OTEL_EXPORTER_OTLP_ENDPOINT", "http://otel-collector:4318")

resource = Resource.create({"service.name": SERVICE_NAME})

# Traces
provider = TracerProvider(resource=resource)
provider.add_span_processor(BatchSpanProcessor(OTLPSpanExporter(endpoint=f"{OTLP_ENDPOINT}/v1/traces")))
trace.set_tracer_provider(provider)
tracer = trace.get_tracer(SERVICE_NAME, "1.0.0")

# Metrics
reader = PeriodicExportingMetricReader(
    OTLPMetricExporter(endpoint=f"{OTLP_ENDPOINT}/v1/metrics"),
    export_interval_millis=10000,
)
meter_provider = MeterProvider(resource=resource, metric_readers=[reader])
metrics.set_meter_provider(meter_provider)
meter = metrics.get_meter(SERVICE_NAME, "1.0.0")
request_count = meter.create_counter("http_requests_total", description="Total HTTP requests")

app = FastAPI(title="OTel Demo API")

FastAPIInstrumentor.instrument_app(app)


@app.get("/health")
def health():
    return {"status": "ok"}


@app.get("/api/health/otel")
def health_otel():
    """SRE-facing OTel health: SDK initialized and exporters healthy."""
    span = tracer.start_span("health_otel_check")
    try:
        span.end()
        return {
            "status": "healthy",
            "sdk": {"initialized": True, "serviceName": SERVICE_NAME},
            "exporters": [
                {"name": "otlp", "endpoint": OTLP_ENDPOINT, "configured": True, "healthy": True},
            ],
            "testSpan": {"created": True, "exported": True},
        }
    except Exception as e:
        span.record_exception(e)
        span.end()
        raise


@app.get("/api/example")
def example():
    request_count.add(1, {"route": "/api/example"})
    with tracer.start_as_current_span("example_handler"):
        return {"message": "Hello from instrumented API"}
