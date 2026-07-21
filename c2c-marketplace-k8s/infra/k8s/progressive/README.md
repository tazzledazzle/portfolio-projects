# Progressive delivery (kind)

Helm-installed controllers live outside this directory:

| Component | Install | Values |
|---|---|---|
| Gloo Edge OSS | `helm upgrade -i gloo gloo/gloo -n gloo-system` | [`values-gloo.yaml`](values-gloo.yaml) |
| Flagger (Gloo) | `helm upgrade -i flagger flagger/flagger -n gloo-system` | [`values-flagger.yaml`](values-flagger.yaml) |
| Flagger (Istio) | same chart in `istio-system` | [`values-flagger-istio.yaml`](values-flagger-istio.yaml) |

`./scripts/deploy-kind.sh` performs those installs. Set `PROGRESSIVE_PROVIDER=istio` for the Istio path (`20-listings-istio.yaml` instead of the Gloo VirtualService + Canary).

Manifests applied with kubectl:

- `10-listings-virtualservice.yaml` — Gloo VS → Flagger RouteTable
- `11-listings-canary.yaml` — Flagger Canary (Gloo)
- `12-metric-templates.yaml` — Micrometer PromQL → `prometheus.c2c:9090`
- `13-loadtester.yaml` — Flagger loadtester
- `20-listings-istio.yaml` — Istio Gateway + Canary (fallback only)
