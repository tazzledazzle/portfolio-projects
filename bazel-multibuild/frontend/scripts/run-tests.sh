#!/usr/bin/env bash
set -euo pipefail

cd "${TEST_SRCDIR}/_main"

node --input-type=module -e "
import { greet } from './src/index.js';
const message = greet('Bazel');
if (message !== 'Hello, Bazel!') {
  throw new Error(\`unexpected greeting: \${message}\`);
}
"
