#!/usr/bin/env bash
set -euo pipefail

./gradlew detekt \
  -Pdetekt.config=config/detekt-custom-rules.yml
