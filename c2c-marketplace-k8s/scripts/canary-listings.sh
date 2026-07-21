#!/usr/bin/env bash
# Trigger a Flagger canary for listings-service by deploying a new image tag.
#
# Usage:
#   ./scripts/canary-listings.sh              # build IMAGE_TAG=canary-<ts>, load, set image, watch
#   IMAGE_TAG=v2 ./scripts/canary-listings.sh # use an existing/custom tag
#   ./scripts/canary-listings.sh --watch-only # only watch an in-flight canary
set -euo pipefail
cd "$(dirname "$0")/.."

CLUSTER_NAME="c2c-marketplace"
WATCH_ONLY=0
if [[ "${1:-}" == "--watch-only" ]]; then
  WATCH_ONLY=1
fi

IMAGE_TAG="${IMAGE_TAG:-canary-$(date +%Y%m%d%H%M%S)}"
IMAGE="c2c/listings-service:${IMAGE_TAG}"

watch_canary() {
  echo "==> Watching canary listings-service (Ctrl-C to stop)"
  echo "    kubectl -n c2c get canary listings-service -w"
  kubectl -n c2c get canary listings-service
  echo
  echo "Tips:"
  echo "  - Events:  kubectl -n c2c describe canary listings-service"
  echo "  - Flagger: kubectl -n gloo-system logs deploy/flagger -f   # or istio-system"
  echo "  - Traffic: curl -H 'Host: listings.local' http://localhost:8081/healthz"
  echo "  - Rollback demo: generate 5xx during Progressing (or scale canary to crash)"
  kubectl -n c2c get canary listings-service -w
}

if [[ "${WATCH_ONLY}" -eq 1 ]]; then
  watch_canary
  exit 0
fi

echo "==> Building ${IMAGE}"
IMAGE_TAG="${IMAGE_TAG}" ./scripts/build-images.sh listings-service

echo "==> Loading ${IMAGE} into kind"
if ! kind load docker-image "${IMAGE}" --name "${CLUSTER_NAME}"; then
  echo "    kind load failed; falling back to docker save | ctr import"
  docker save "${IMAGE}" | docker exec -i "${CLUSTER_NAME}-control-plane" \
    ctr --namespace=k8s.io images import -
fi
# Also ensure :local alias exists if scripts expect it
kind load docker-image "c2c/listings-service:local" --name "${CLUSTER_NAME}" 2>/dev/null || true

echo "==> Setting listings-service image to ${IMAGE}"
kubectl -n c2c set image deployment/listings-service \
  listings-service="${IMAGE}"

echo "==> Waiting for Flagger to detect Progressing"
ok=0
for _ in $(seq 1 36); do
  phase="$(kubectl -n c2c get canary listings-service -o jsonpath='{.status.phase}' 2>/dev/null || true)"
  if [[ "${phase}" == "Progressing" || "${phase}" == "WaitingPromotion" || "${phase}" == "Promoting" ]]; then
    echo "    Canary phase: ${phase}"
    ok=1
    break
  fi
  sleep 5
done
if [[ "${ok}" -ne 1 ]]; then
  echo "WARNING: Did not observe Progressing yet (phase=${phase:-unknown})."
  echo "         If the Deployment template did not change, rebuild with a new IMAGE_TAG."
fi

watch_canary
