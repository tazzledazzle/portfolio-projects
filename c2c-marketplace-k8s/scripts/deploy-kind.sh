#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."

CLUSTER_NAME="c2c-marketplace"
# gloo (default) | istio
PROGRESSIVE_PROVIDER="${PROGRESSIVE_PROVIDER:-gloo}"
FLAGGER_CRD_URL="${FLAGGER_CRD_URL:-https://raw.githubusercontent.com/fluxcd/flagger/main/artifacts/flagger/crd.yaml}"

need_cmd() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "ERROR: required command not found: $1" >&2
    exit 1
  }
}

need_cmd kind
need_cmd kubectl
need_cmd helm
need_cmd docker

# kind load docker-image can fail on Colima/containerd with multi-arch digests.
# Fall back to docker save | ctr import when that happens.
load_image_into_kind() {
  local image="$1"
  if kind load docker-image "$image" --name "${CLUSTER_NAME}"; then
    return 0
  fi
  echo "    kind load failed for ${image}; falling back to docker save | ctr import"
  docker save "$image" | docker exec -i "${CLUSTER_NAME}-control-plane" \
    ctr --namespace=k8s.io images import -
}

ensure_image() {
  local image="$1"
  if ! docker image inspect "$image" >/dev/null 2>&1; then
    docker pull --platform linux/arm64 "$image" 2>/dev/null || docker pull "$image"
  fi
  load_image_into_kind "$image"
}

if ! kind get clusters 2>/dev/null | grep -q "^${CLUSTER_NAME}$"; then
  echo "==> Creating kind cluster ${CLUSTER_NAME}"
  kind create cluster --config infra/k8s/kind-config.yaml
else
  echo "==> Reusing existing kind cluster ${CLUSTER_NAME}"
  echo "    Note: extraPortMappings only apply on create; recreate the cluster if host ports changed."
fi

echo "==> Loading images into kind"
for service in listings-service search-service messaging-service payments-service; do
  load_image_into_kind "c2c/${service}:local"
done

# Infra images are pulled from registries by default; on a cold kind node
# that often lands in ImagePullBackOff. Pre-load from the host daemon.
echo "==> Ensuring infra images are available in kind"
for image in \
  postgres:16-alpine \
  redis:7-alpine \
  opensearchproject/opensearch:2.14.0 \
  redpandadata/redpanda:v24.1.9
do
  ensure_image "$image"
done

echo "==> Ensuring observability images are available in kind"
for image in \
  prom/prometheus:v2.54.1 \
  grafana/loki:3.1.1 \
  grafana/tempo:2.6.1 \
  grafana/grafana:11.2.0 \
  grafana/alloy:v1.3.1 \
  registry.k8s.io/kube-state-metrics/kube-state-metrics:v2.13.0
do
  ensure_image "$image"
done

echo "==> Applying manifests"
kubectl apply -f infra/k8s/00-namespace.yaml
kubectl apply -f infra/k8s/01-postgres.yaml
kubectl apply -f infra/k8s/02-redis.yaml
kubectl apply -f infra/k8s/03-opensearch.yaml
kubectl apply -f infra/k8s/04-redpanda.yaml

echo "==> Waiting for infra to be ready before starting services"
kubectl -n c2c wait --for=condition=available --timeout=180s deployment/postgres deployment/redis deployment/opensearch deployment/redpanda

echo "==> Applying observability stack (Prometheus/Loki/Tempo/Grafana/Alloy)"
# Dashboard JSON lives as files; materialize ConfigMap so Grafana can provision them.
kubectl -n c2c create configmap grafana-dashboards \
  --from-file=infra/k8s/observability/grafana/dashboards/ \
  --dry-run=client -o yaml | kubectl apply -f -
kubectl apply -f infra/k8s/observability/

echo "==> Waiting for Prometheus (Flagger metrics provider)"
kubectl -n c2c wait --for=condition=available --timeout=180s deployment/prometheus

install_gloo() {
  echo "==> Installing Gloo Edge (progressive delivery provider)"
  helm repo add gloo https://storage.googleapis.com/solo-public-helm >/dev/null 2>&1 || true
  helm repo update gloo >/dev/null
  helm upgrade -i gloo gloo/gloo \
    --namespace gloo-system \
    --create-namespace \
    -f infra/k8s/progressive/values-gloo.yaml \
    --wait --timeout 5m

  echo "==> Installing Flagger CRDs + controller (meshProvider=gloo)"
  kubectl apply -f "${FLAGGER_CRD_URL}"
  helm repo add flagger https://flagger.app >/dev/null 2>&1 || true
  helm repo update flagger >/dev/null
  helm upgrade -i flagger flagger/flagger \
    --namespace gloo-system \
    -f infra/k8s/progressive/values-flagger.yaml \
    --wait --timeout 3m

  # Loadtester image
  ensure_image ghcr.io/fluxcd/flagger-loadtester:0.34.0

  echo "==> Applying progressive delivery manifests (Gloo)"
  kubectl apply -f infra/k8s/progressive/13-loadtester.yaml
  kubectl apply -f infra/k8s/progressive/12-metric-templates.yaml
  kubectl apply -f infra/k8s/progressive/10-listings-virtualservice.yaml
}

install_istio() {
  echo "==> Installing Istio (progressive delivery fallback)"
  helm repo add istio https://istio-release.storage.googleapis.com/charts >/dev/null 2>&1 || true
  helm repo update istio >/dev/null
  kubectl create namespace istio-system --dry-run=client -o yaml | kubectl apply -f -
  helm upgrade -i istio-base istio/base \
    --namespace istio-system \
    --set defaultRevision=default \
    --wait --timeout 3m
  helm upgrade -i istiod istio/istiod \
    --namespace istio-system \
    --wait --timeout 5m
  helm upgrade -i istio-ingress istio/gateway \
    --namespace istio-system \
    -f infra/k8s/progressive/values-istio-gateway.yaml \
    --wait --timeout 3m

  kubectl label namespace c2c istio-injection=enabled --overwrite

  echo "==> Installing Flagger CRDs + controller (meshProvider=istio)"
  kubectl apply -f "${FLAGGER_CRD_URL}"
  helm repo add flagger https://flagger.app >/dev/null 2>&1 || true
  helm repo update flagger >/dev/null
  helm upgrade -i flagger flagger/flagger \
    --namespace istio-system \
    -f infra/k8s/progressive/values-flagger-istio.yaml \
    --wait --timeout 3m

  ensure_image ghcr.io/fluxcd/flagger-loadtester:0.34.0

  echo "==> Applying progressive delivery manifests (Istio)"
  kubectl apply -f infra/k8s/progressive/13-loadtester.yaml
  kubectl apply -f infra/k8s/progressive/12-metric-templates.yaml
}

case "${PROGRESSIVE_PROVIDER}" in
  gloo)
    install_gloo
    ;;
  istio)
    install_istio
    ;;
  *)
    echo "ERROR: PROGRESSIVE_PROVIDER must be 'gloo' or 'istio' (got: ${PROGRESSIVE_PROVIDER})" >&2
    exit 1
    ;;
esac

echo "==> Applying application manifests"
kubectl apply -f infra/k8s/10-listings.yaml
kubectl apply -f infra/k8s/11-search.yaml
kubectl apply -f infra/k8s/12-messaging.yaml
kubectl apply -f infra/k8s/13-payments.yaml

echo "==> Applying listings Canary"
if [[ "${PROGRESSIVE_PROVIDER}" == "gloo" ]]; then
  kubectl apply -f infra/k8s/progressive/11-listings-canary.yaml
else
  kubectl apply -f infra/k8s/progressive/20-listings-istio.yaml
fi

echo "==> Waiting for listings Canary to initialize"
# Flagger sets condition Promoted=True with reason Initialized/Succeeded after bootstrap.
for i in $(seq 1 60); do
  phase="$(kubectl -n c2c get canary listings-service -o jsonpath='{.status.phase}' 2>/dev/null || true)"
  if [[ "${phase}" == "Initialized" || "${phase}" == "Succeeded" ]]; then
    echo "    Canary phase: ${phase}"
    break
  fi
  if [[ "${i}" -eq 60 ]]; then
    echo "WARNING: Canary not Initialized/Succeeded yet (phase=${phase:-unknown}). Check: kubectl -n c2c describe canary listings-service"
  fi
  sleep 5
done

echo "==> Done. Check status with: kubectl -n c2c get pods -w"
echo "==> Listings via gateway: curl -H 'Host: listings.local' http://localhost:8081/healthz"
echo "==> Other services: localhost:8082-8084 (direct NodePorts)"
echo "==> Grafana: http://localhost:3000 (anonymous Viewer)"
echo "==> Progressive provider: ${PROGRESSIVE_PROVIDER}"
echo "==> Canary demo: ./scripts/canary-listings.sh"
