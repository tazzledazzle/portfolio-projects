# Local Service Mesh

## Overview
A local networking layer based on Docker Compose and Caddy that preserves Kubernetes-like service naming so teams can validate service-to-service behavior without a full cluster.

## Architecture
Compose manages service lifecycles while Caddy acts as the local routing gateway. Route definitions in `gateway/routes.json` map stable service identities to local containers and ports, mimicking in-cluster DNS and ingress behavior.

## Use Cases
- Reproduce cross-service request flows locally.
- Test route rewrites, auth propagation, and upstream retries.
- Validate local integration scenarios before deploying to shared environments.

## Usage
1. Configure routes in `gateway/routes.json` and `caddy/Caddyfile`.
2. Start the mesh with `scripts/up.sh`.
3. Run local service calls through mesh hostnames.
4. Tear down cleanly with `scripts/down.sh`.

## Control Flow
1. Compose boots gateway, proxy, and service containers
2. Caddy receives incoming mesh requests.
3. Routing rules resolve destination service aliases.
4. Requests are proxied to the target local service.
5. Integration tests validate expected end-to-end behavior.

## Project Structure
```text
.
├── caddy/
├── compose/
├── docs/
│   ├── architecture/
│   └── routing/
├── gateway/
├── scripts/
├── services/
│   ├── order-service/
│   ├── payment-service/
│   └── user-service/
└── tests/
    ├── e2e/
    └── integration/
```
