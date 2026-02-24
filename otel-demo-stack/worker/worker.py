"""
Minimal worker that runs a periodic task and emits OTel traces/metrics (W3C context).
Sends OTLP to the same collector as the API.
"""
import os
import time
from opentelemetry import trace, metrics
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.sdk.metrics import MeterProvider
from opentelemetry.sdk.metrics.export import PeriodicExportingMetricReader
from opentelemetry.sdk.resources import Resource
from opentelemetry.exporter.otlp.proto.http.trace_exporter import OTLPSpanExporter
from opentelemetry.exporter.otlp.proto.http.metric_exporter import OTLPMetricExporter

SERVICE_NAME = os.environ.get("OTEL_SERVICE_NAME", "otel-demo-worker")
OTLP_ENDPOINT = os.environ.get("OTEL_EXPORTER_OTLP_ENDPOINT", "http://otel-collector:4318")

resource = Resource.create({"service.name": SERVICE_NAME})

provider = TracerProvider(resource=resource)
provider.add_span_processor(BatchSpanProcessor(OTLPSpanExporter(endpoint=f"{OTLP_ENDPOINT}/v1/traces")))
trace.set_tracer_provider(provider)
tracer = trace.get_tracer(SERVICE_NAME, "1.0.0")

reader = PeriodicExportingMetricReader(
    OTLPMetricExporter(endpoint=f"{OTLP_ENDPOINT}/v1/metrics"),
    export_interval_millis=10000,
)
meter_provider = MeterProvider(resource=resource, metric_readers=[reader])
metrics.set_meter_provider(meter_provider)
meter = metrics.get_meter(SERVICE_NAME, "1.0.0")
tasks_done = meter.create_counter("worker_tasks_total", description="Total tasks completed")


def do_work() -> None:
    with tracer.start_as_current_span("worker_task"):
        time.sleep(0.5)
        tasks_done.add(1, {"task": "tick"})


if __name__ == "__main__":
    import logging
    logging.basicConfig(level=logging.INFO)
    logger = logging.getLogger(__name__)
    logger.info("Worker starting; OTLP endpoint=%s", OTLP_ENDPOINT)
    while True:
        do_work()
        time.sleep(5)
