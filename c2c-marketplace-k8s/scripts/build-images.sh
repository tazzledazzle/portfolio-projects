#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."

# Optional: IMAGE_TAG=v2 ./scripts/build-images.sh listings-service
# builds only listings-service tagged c2c/listings-service:v2 (and :local).
IMAGE_TAG="${IMAGE_TAG:-local}"
SERVICES=("$@")
if [[ ${#SERVICES[@]} -eq 0 ]]; then
  SERVICES=(listings-service search-service messaging-service payments-service)
fi

echo "==> Building JARs with Gradle"
./gradlew installDist --no-daemon

for service in "${SERVICES[@]}"; do
  echo "==> Building image c2c/${service}:${IMAGE_TAG}"
  docker build -t "c2c/${service}:${IMAGE_TAG}" -f "${service}/Dockerfile" .
  # Always keep :local as an alias when using a custom tag so default manifests work.
  if [[ "${IMAGE_TAG}" != "local" ]]; then
    docker tag "c2c/${service}:${IMAGE_TAG}" "c2c/${service}:local"
  fi
done

echo "==> Done. Images built:"
docker images | grep '^c2c/' || true
