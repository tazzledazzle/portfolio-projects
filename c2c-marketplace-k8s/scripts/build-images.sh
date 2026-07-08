#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."

echo "==> Building JARs with Gradle"
./gradlew installDist --no-daemon

for service in listings-service search-service messaging-service payments-service; do
  echo "==> Building image c2c/${service}:local"
  docker build -t "c2c/${service}:local" -f "${service}/Dockerfile" .
done

echo "==> Done. Images built:"
docker images | grep '^c2c/'
