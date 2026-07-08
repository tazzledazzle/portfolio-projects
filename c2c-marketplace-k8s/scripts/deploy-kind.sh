#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."

CLUSTER_NAME="c2c-marketplace"

if ! kind get clusters | grep -q "^${CLUSTER_NAME}$"; then
  echo "==> Creating kind cluster ${CLUSTER_NAME}"
  kind create cluster --config infra/k8s/kind-config.yaml
else
  echo "==> Reusing existing kind cluster ${CLUSTER_NAME}"
fi

echo "==> Loading images into kind"
for service in listings-service search-service messaging-service payments-service; do
  kind load docker-image "c2c/${service}:local" --name "${CLUSTER_NAME}"
done

echo "==> Applying manifests"
kubectl apply -f infra/k8s/00-namespace.yaml
kubectl apply -f infra/k8s/01-postgres.yaml
kubectl apply -f infra/k8s/02-redis.yaml
kubectl apply -f infra/k8s/03-opensearch.yaml
kubectl apply -f infra/k8s/04-redpanda.yaml

echo "==> Waiting for infra to be ready before starting services"
kubectl -n c2c wait --for=condition=available --timeout=180s deployment/postgres deployment/redis deployment/opensearch deployment/redpanda

kubectl apply -f infra/k8s/10-listings.yaml
kubectl apply -f infra/k8s/11-search.yaml
kubectl apply -f infra/k8s/12-messaging.yaml
kubectl apply -f infra/k8s/13-payments.yaml

echo "==> Done. Check status with: kubectl -n c2c get pods -w"
echo "==> Services reachable at localhost:8081-8084 (via kind's extraPortMappings)"
