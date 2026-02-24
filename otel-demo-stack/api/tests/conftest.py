"""Set OTLP endpoint to localhost for tests so collector hostname is not required."""
import os

# Set before main is imported so the app uses localhost instead of 'otel-collector'
os.environ.setdefault("OTEL_EXPORTER_OTLP_ENDPOINT", "http://127.0.0.1:4318")
