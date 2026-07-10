# API Gateway with Kubernetes Gateway API: Routing, Rate Limiting, and TLS

An API gateway is the controlled entry point for external traffic into your microservices cluster. It handles cross-cutting concerns — TLS termination, authentication, rate limiting, routing — that would otherwise need to be duplicated across every service. In Kubernetes, the modern way to configure a gateway is the **Gateway API**, a set of official Kubernetes resource types that replaces the older Ingress resource with a richer, more expressive model.

This guide covers the Gateway API implementation in this codebase: a two-listener gateway with HTTPS termination, HTTPRoute traffic routing including weighted canary splits, rate limiting with global counters, and JWT authentication — all configured declaratively as Kubernetes resources.

---

## Why Not Ingress?

The Ingress resource was Kubernetes' original answer to external traffic routing. It works, but it has real limitations:

- A single resource type tries to cover all concerns: routing, TLS, load balancing behavior, annotations-as-escape-hatch for everything else
- Annotations became the de facto extension point, but they're controller-specific — `nginx.ingress.kubernetes.io/rate-limit` means nothing to an Envoy-based controller
- No separation between infrastructure concerns (who manages the load balancer) and application concerns (which paths route where)
- Limited support for advanced routing (traffic splitting, header-based routing, gRPC)

Gateway API separates these concerns across three resource types:

- **GatewayClass**: defines the controller implementation (Envoy Gateway, Nginx, Cilium, etc.)
- **Gateway**: defines the infrastructure — listeners, ports, TLS certificates, which namespaces can attach routes
- **HTTPRoute / GRPCRoute**: defines application routing — which paths go where, timeouts, header manipulation

---

## GatewayClass and Gateway: Infrastructure Layer

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: GatewayClass
metadata:
  name: envoy-gateway
spec:
  controllerName: gateway.envoyproxy.io/gatewayclass-controller
```

The GatewayClass names the implementation. When you create a Gateway that references this class, the Envoy Gateway controller provisions the actual infrastructure — a dedicated Envoy fleet in this case.

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: main-gateway
  namespace: gateway-system
spec:
  gatewayClassName: envoy-gateway
  listeners:
    - name: http
      protocol: HTTP
      port: 80
      allowedRoutes:
        namespaces:
          from: Selector
          selector:
            matchLabels:
              gateway-access: "true"
    - name: https
      protocol: HTTPS
      port: 443
      tls:
        mode: Terminate
        certificateRefs:
          - name: api-tls-cert
            namespace: gateway-system
      allowedRoutes:
        namespaces:
          from: Selector
          selector:
            matchLabels:
              gateway-access: "true"
```

Two listeners: HTTP on 80 (for redirect) and HTTPS on 443 (with TLS termination). The certificate reference points to a Kubernetes Secret managed by cert-manager — the gateway handles the TLS handshake, and traffic inside the cluster is unencrypted HTTP. If you need encryption inside the cluster too, that's where a service mesh comes in.

`allowedRoutes` with a namespace selector is the access control mechanism. Only namespaces with the label `gateway-access: "true"` can attach HTTPRoutes to this Gateway. This prevents any team from accidentally (or intentionally) exposing their service through the production gateway without going through an approval process that adds the label.

---

## HTTPRoute: Application Routing Layer

HTTPRoutes are created by application teams in their own namespaces. They attach to the Gateway and define routing rules:

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: orders-route
  namespace: ecommerce
spec:
  parentRefs:
    - name: main-gateway
      namespace: gateway-system
      sectionName: https
  hostnames:
    - "api.example.com"
  rules:
    - matches:
        - path:
            type: PathPrefix
            value: /v1/orders
          headers:
            - name: x-api-version
              type: Exact
              value: "2024"
      backendRefs:
        - name: orders-service
          port: 8080
          weight: 90
        - name: orders-service-canary
          port: 8080
          weight: 10
      timeouts:
        request: 5s
        backendRequest: 4s
```

A few things worth examining:

**Combined path and header matching** (`/v1/orders` + `x-api-version: 2024`) routes requests to the new API version only when clients send the version header. Older clients that don't send the header can be routed to a legacy backend in a separate rule. This enables version migration without changing the URL.

**Weighted traffic split** (90/10 between `orders-service` and `orders-service-canary`) is how you do canary deployments at the gateway layer. Route 10% of production traffic to the canary, observe error rates and latency, then roll the weight to 100 when satisfied. This is more controlled than a Deployment rollout because you choose the percentage explicitly.

**Explicit timeouts** (`request: 5s`, `backendRequest: 4s`) deserve a dedicated mention. A gateway without timeouts will hold connections open indefinitely when a backend is slow. This is how a slow backend takes down the gateway under load. The 1-second gap between request and backendRequest gives the gateway time to return a 504 to the client before the backend request times out.

### Payment Route: Different Timeouts for Different SLAs

```yaml
- matches:
    - path:
        type: PathPrefix
        value: /v1/payments
  backendRefs:
    - name: payment-service
      namespace: payments
      port: 8080
  timeouts:
    request: 10s
    backendRequest: 9s
  filters:
    - type: RequestHeaderModifier
      requestHeaderModifier:
        add:
          - name: x-forwarded-gateway
            value: "main-gateway"
          - name: x-request-id
            value: "$(request.id)"
```

Payment operations legitimately take longer — a 5-second timeout would generate false failures during peak load. Separate routes for different services mean each gets timeouts appropriate to its SLA.

The `RequestHeaderModifier` filter injects headers before forwarding. `x-forwarded-gateway` tells downstream services which gateway processed the request — useful for audit logs that span multiple gateways. `x-request-id` propagates a unique request identifier for distributed tracing correlation.

### gRPC Routing

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: GRPCRoute
metadata:
  name: orders-grpc-route
  namespace: ecommerce
spec:
  rules:
    - matches:
        - method:
            service: orders.v1.OrderService
      backendRefs:
        - name: orders-service
          port: 9090
```

Gateway API natively understands gRPC. The `GRPCRoute` matches on the fully qualified service name (`orders.v1.OrderService`), which is derived from the Protocol Buffers package and service definition. This is cleaner than HTTP-level path matching for gRPC, which uses `/package.Service/Method` paths.

---

## Rate Limiting

The `BackendTrafficPolicy` attaches to the `orders-route` HTTPRoute and applies a global rate limit:

```yaml
rateLimit:
  type: Global
  global:
    rules:
      - clientSelectors:
          - headers:
              - name: x-api-key
                type: Distinct
        limit:
          requests: 1000
          unit: Hour
      - clientSelectors:
          - headers:
              - name: authorization
                type: Distinct
        limit:
          requests: 500
          unit: Minute
      - limit:
          requests: 100
          unit: Minute
```

Three tiers: per-API-key (1000/hour), per-JWT-subject (500/minute), and a global fallback for unauthenticated traffic (100/minute).

`type: Global` means the rate limit counter is shared across all Envoy instances via Redis. Without this, each Envoy pod maintains its own counter — with 3 pods, a client could make 3× the intended limit by hitting different pods.

`type: Distinct` per API key means each unique key gets its own counter. The alternative is a shared counter for all clients using the same rule, which would pool everyone together. Distinct is almost always what you want for per-client limits.

IP-based rate limiting — a common alternative — is broken for mobile carriers and corporate NATs where thousands of users share one egress IP. Identity-based limiting (API key or JWT subject) targets the actual client.

---

## JWT Authentication

```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: SecurityPolicy
metadata:
  name: orders-jwt-auth
spec:
  targetRef:
    kind: HTTPRoute
    name: orders-route
  jwt:
    providers:
      - name: auth0
        issuer: "https://auth.example.com/"
        remoteJWKS:
          uri: "https://auth.example.com/.well-known/jwks.json"
          cacheDuration: 5m
        claimToHeaders:
          - claim: sub
            header: x-jwt-subject
          - claim: email
            header: x-jwt-email
```

The gateway validates JWTs before forwarding requests. Downstream services can trust the `x-jwt-subject` and `x-jwt-email` headers injected by the gateway — they don't need to verify the JWT themselves. This shifts the authentication boundary to the edge, where it belongs, and avoids distributing JWT verification logic across every service.

`cacheDuration: 5m` caches the JWKS (public keys) for 5 minutes. Without caching, every request would hit the auth server for key material, which adds latency and creates a dependency on the auth server's availability.

---

## Gateway vs. Service Mesh

A gateway and a service mesh address different layers of traffic:

| Concern | API Gateway | Service Mesh |
|---------|------------|--------------|
| External traffic (north-south) | Yes | No |
| Internal traffic (east-west) | No | Yes |
| TLS termination at edge | Yes | Not applicable |
| mTLS between services | No | Yes |
| JWT auth at edge | Yes | Can validate, but not the primary use case |
| Request routing and canary | Yes | Yes (for internal traffic) |
| Observability (access logs, metrics) | Yes | Yes |

They're complementary. The gateway controls what enters the cluster and authenticates clients. The service mesh controls how services talk to each other inside the cluster. In this stack, the `03-api-gateway` pattern (Kubernetes Gateway API with Envoy Gateway) handles edge traffic, while the `08-service-mesh` pattern (Istio Ambient) handles east-west.

---

## Key Takeaways

- Kubernetes Gateway API separates concerns: GatewayClass (implementation), Gateway (infrastructure), HTTPRoute (application routing) — different teams own different layers
- Namespace selectors on `allowedRoutes` control which teams can attach to the gateway — a first line of defense against accidental exposure
- Always set explicit timeouts on routes — an absent timeout means connections hold open indefinitely under backend failure
- Weighted traffic splits in HTTPRoute enable canary deployments at the gateway without touching Deployment specs
- Global rate limiting requires a shared counter store (Redis) — per-pod counters multiply the effective limit by replica count
- Rate limit by identity (API key, JWT subject), not by IP — IP-based limiting breaks under NAT
- JWT validation at the gateway shifts auth to the edge and lets downstream services trust injected headers
- A gateway handles north-south (external) traffic; a service mesh handles east-west (internal) — they're complementary, not competing
- The `request` timeout should be slightly longer than `backendRequest` to give the gateway time to return a 504 before the backend times out
