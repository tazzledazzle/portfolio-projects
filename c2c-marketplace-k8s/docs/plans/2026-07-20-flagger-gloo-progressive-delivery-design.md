# Flagger + Gloo Edge Progressive Delivery (listings pilot)

**Date:** 2026-07-20  
**Status:** Approved for implementation  
**Scope:** `listings-service` only on kind; existing LGTM Prometheus stands in for Datadog

## Goal

Automate canary analysis and promotion for `listings-service` using Flagger, with Gloo Edge as the traffic provider and the existing `prometheus.c2c:9090` instance as the metrics backend.

## Why Gloo (not Istio by default)

The pilot is ingress-only: external HTTP to listings. Gloo Edge avoids per-pod sidecars on an already memory-heavy one-node kind cluster (OpenSearch + discrete LGTM). Istio remains an explicit fallback via `PROGRESSIVE_PROVIDER=istio` if Gloo install or Envoy scraping fails locally.

## Architecture

```
Host :8081 → kind NodePort 30081 → Gloo gateway-proxy
  → VirtualService (listings.local)
  → RouteTable (Flagger-owned weights)
  → listings-service-primary | listings-service-canary

Alloy scrapes app + gateway-proxy → remote_write → prometheus.c2c:9090
Flagger queries Prometheus via MetricTemplates during analysis
```

## Design decisions

1. **Reuse host port 8081** — Gloo gateway-proxy takes NodePort `30081` so `localhost:8081` still reaches listings (via the gateway). Search/messaging/payments keep direct NodePorts.
2. **Flagger owns listings Services** — remove the hand-written NodePort Service; Flagger generates apex / primary / canary ClusterIP Services and a `-primary` Deployment.
3. **No Flagger-bundled Prometheus** — `metricsServer=http://prometheus.c2c:9090`, `prometheus.install=false`.
4. **Micrometer MetricTemplates first** — builtin Gloo Envoy queries (`envoy_cluster_upstream_rq`) are version-sensitive. Primary checks use existing `http_server_requests_seconds_*` series, filtered by canary pod name regex.
5. **Load tests through the gateway** — Flagger loadtester hits `gateway-proxy.gloo-system` with Host `listings.local` so gateway and app metrics both see traffic.
6. **Tagged images for demos** — `IMAGE_TAG=v2 ./scripts/build-images.sh listings-service` so Flagger detects a Deployment revision change (fixed `:local` alone does not).

## Analysis policy

| Setting | Value |
|---|---|
| interval | 30s |
| stepWeight | 10 |
| maxWeight | 50 |
| threshold | 5 |
| success rate | ≥ 99% (non-5xx) |
| latency p99 | ≤ 500ms |

## Out of scope

- Canarying search, payments, or WebSocket messaging
- Production HA Gloo/Istio, mTLS, auth gateway
- Changing global SLO recording-rule schema

## References

- Flagger Gloo tutorial: https://docs.flagger.app/main/tutorials/gloo-progressive-delivery
- Flagger metrics: https://docs.flagger.app/main/usage/metrics
- Gloo kind install: https://docs.solo.io/gloo-edge/latest/installation/gateway/kubernetes/
