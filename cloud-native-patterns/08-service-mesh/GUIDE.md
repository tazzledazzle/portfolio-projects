# Service Mesh with Istio Ambient Mode: mTLS, Authorization, and the Waypoint Architecture

A service mesh moves cross-cutting network concerns — mutual TLS, observability, authorization policies, retry logic — out of application code and into the infrastructure layer. Every service gets consistent security and observability without writing a line of networking code.

This guide covers Istio in **Ambient mode**, a significant architectural departure from the traditional sidecar model. Rather than injecting a proxy container into every pod, Ambient uses shared node-level proxies (ztunnels) and per-namespace waypoint proxies — giving you the same features with lower resource overhead and no pod restarts when adding or removing the mesh.

---

## What the Service Mesh Solves

Without a mesh, each service is responsible for:

- **mTLS**: Verifying the identity of callers, encrypting traffic
- **Observability**: Emitting request latency, error rates, connection counts
- **Retries and timeouts**: Implementing consistent retry behavior
- **Authorization**: Enforcing which services can call which endpoints

Teams either implement these per-service (duplicated across languages and stacks) or skip them (vulnerable and unobservable). A service mesh implements them once, at the infrastructure layer, for every service in the cluster.

---

## Sidecar vs. Ambient: Two Models

**Sidecar mode** (Istio's original architecture): An Envoy proxy container is injected into every pod. All traffic into and out of the pod flows through this sidecar. The sidecar handles mTLS, observability, and policy.

Problems with sidecars:
- Every pod carries a proxy that consumes CPU and memory (~50MB RAM per sidecar)
- Adding the mesh requires rolling restarts of all pods (to inject the sidecar)
- Proxy upgrades require rolling restarts of all application pods
- The sidecar and application share the pod lifecycle

**Ambient mode** (Istio 1.22+, GA): Two layers of proxies, separate from application pods.

**Layer 1 — ztunnel** (zero-trust tunnel): A DaemonSet running one pod per node. Handles L4 concerns: mTLS termination, basic L4 authorization, telemetry. All traffic on the node passes through ztunnel. No pod restarts needed to enable.

**Layer 2 — Waypoint proxy**: A per-namespace (or per-service) Envoy proxy. Handles L7 concerns: HTTP routing, header-based authorization, JWT validation, retries, traffic shaping. Only deployed when L7 features are needed.

---

## Enabling Ambient Mode

The namespace label is all it takes:

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: payments
  labels:
    istio.io/dataplane-mode: ambient
```

Every pod in the `payments` namespace now has its traffic handled by the node's ztunnel. No annotation per pod, no sidecar injection webhook, no application restarts. Services that were already running pick up mTLS automatically.

The annotation in the YAML is worth reading:

```yaml
annotations:
  # Ambient mode was GA in Istio 1.22 (May 2024).
  # Sidecar injection label (istio-injection: enabled) is mutually exclusive with ambient.
```

Sidecar and ambient are incompatible within the same namespace. You pick one model per namespace.

---

## Enforcing mTLS: PeerAuthentication

```yaml
apiVersion: security.istio.io/v1
kind: PeerAuthentication
metadata:
  name: payments-mtls
  namespace: payments
spec:
  mtls:
    mode: STRICT
```

`STRICT` mode means the ztunnel will reject any plaintext connection to pods in the `payments` namespace. Only mTLS connections with valid SPIFFE/X.509 certificates are accepted.

Certificates are issued by Istiod (the Istio control plane) to each workload identity. The identity is derived from the pod's Kubernetes ServiceAccount:

```
spiffe://cluster.local/ns/payments/sa/payment-service
```

This is workload identity, not IP-based identity. An attacker who steals an IP address and sends traffic claiming to be `orders-service` will be rejected because they don't have the certificate for the `orders-service` identity.

The mesh-wide default:

```yaml
apiVersion: security.istio.io/v1
kind: PeerAuthentication
metadata:
  name: default-mtls
  namespace: istio-system
spec:
  mtls:
    mode: STRICT
```

Applying this to `istio-system` makes STRICT the cluster-wide default for all ambient-enabled namespaces. This is the right production posture — opt out explicitly if needed, rather than opting in.

---

## The Waypoint: L7 Policy for the Payments Namespace

The ztunnel handles L4 — it knows about connections, not HTTP methods or headers. For L7 authorization (allow POST to `/v1/payments/charge` but not GET), a Waypoint proxy is needed.

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: payments-waypoint
  namespace: payments
  labels:
    istio.io/waypoint-for: service
spec:
  gatewayClassName: istio-waypoint
  listeners:
    - name: mesh
      port: 15008
      protocol: HBONE    # HTTP/2 tunnel used by Istio ambient ztunnel
```

The `istio-waypoint` GatewayClass is an Istio-provided class. Creating this Gateway resource causes Istiod to provision an Envoy proxy in the `payments` namespace that handles L7 traffic for services labeled to use it.

Attaching a service to the waypoint:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: payment-service
  namespace: payments
  labels:
    istio.io/use-waypoint: payments-waypoint
```

Traffic flow: external request → ztunnel (L4 mTLS) → waypoint (L7 policy, routing) → payment-service pod. Services without the waypoint label get L4-only (mTLS via ztunnel, no L7 policy).

HBONE (HTTP/2-Based Overlay Network Encapsulation) is the tunneling protocol ztunnel uses to forward traffic to waypoints. It's an implementation detail — application code is unaware of it.

---

## Authorization Policies: Zero-Trust in Practice

```yaml
apiVersion: security.istio.io/v1
kind: AuthorizationPolicy
metadata:
  name: payment-service-authz
  namespace: payments
spec:
  selector:
    matchLabels:
      app: payment-service
  action: ALLOW
  rules:
    - from:
        - source:
            principals:
              - "cluster.local/ns/orders/sa/orders-service"
              - "cluster.local/ns/payment-ops/sa/payment-admin"
      to:
        - operation:
            methods: ["POST"]
            paths: ["/v1/payments/charge", "/v1/payments/refund"]
        - operation:
            methods: ["GET"]
            paths: ["/v1/payments/*"]
    - from:
        - source:
            namespaces: ["kube-system"]
      to:
        - operation:
            paths: ["/healthz/*"]
---
apiVersion: security.istio.io/v1
kind: AuthorizationPolicy
metadata:
  name: deny-all-payments
  namespace: payments
spec:
  action: DENY
  rules:
    - {}
```

The two policies work together:

**`deny-all-payments`** has an empty `selector` (matches all pods in the namespace) and a single empty rule that matches all traffic. This is the default-deny baseline.

**`payment-service-authz`** selects only `payment-service` pods and explicitly allows specific principals for specific methods and paths.

The principal format is the SPIFFE URI: `cluster.local/ns/orders/sa/orders-service`. Only the pod running with the `orders-service` ServiceAccount in the `orders` namespace can call the payment charge and refund endpoints. Not pods with the orders-service IP. Not pods with the orders-service label. The ServiceAccount identity, verified by mTLS certificate.

This authorization model has a specific boundary: it's service-to-service authorization, not user authorization. The user identity (who placed the order) is typically carried in a JWT validated at the gateway. The mesh validates that the calling *service* is authorized to make the call.

The health check rule uses namespace (`kube-system`) rather than principal because kubelet health checks come from the node, not from a workload with a SPIFFE identity. Namespace-based allowances are wider — use them only for infrastructure access patterns.

---

## Sidecar vs. Ambient: The Decision

| Concern | Sidecar | Ambient |
|---------|---------|---------|
| Resource overhead | ~50MB RAM per pod | One ztunnel per node |
| Pod restarts to enable | Yes | No |
| Proxy upgrade rollout | Rolling restart of apps | ztunnel DaemonSet update |
| L4 features (mTLS, telemetry) | Sidecar | ztunnel |
| L7 features (HTTP policy, routing) | Sidecar | Waypoint proxy |
| Multi-cluster | Mature | Newer |
| Maturity | GA since 2017 | GA since May 2024 |

For new deployments, Ambient is the better default. The operational simplicity (no sidecar injection, no rolling restarts) is a real advantage, and the security model is equivalent.

For existing sidecar deployments, migration is possible namespace by namespace. The two modes can coexist in the same cluster.

---

## When to Introduce a Service Mesh

A service mesh makes sense when:

- You need mTLS between services and don't want to implement certificate rotation per service
- You need consistent observability (request latency, error rates) across services in multiple languages
- You need L7 authorization policies (which service can call which endpoint) enforced by infrastructure rather than application code
- You want traffic management (retries, circuit breaking, canary routing) at the mesh level

A mesh adds operational complexity: Istio control plane components to manage, certificate rotation infrastructure, new CRD types for your team to learn. If you have two or three services and a small team, start without a mesh and add one when the cross-cutting concerns become painful.

---

## Key Takeaways

- A service mesh moves mTLS, observability, authorization, and traffic management into infrastructure — no per-service implementation needed
- Istio Ambient mode replaces per-pod sidecars with shared node-level ztunnels and optional per-namespace waypoints
- Enabling ambient requires only a namespace label — no pod restarts, no sidecar injection
- `PeerAuthentication STRICT` rejects plaintext connections — workload identity (SPIFFE certificate) is verified, not IP addresses
- Waypoint proxies handle L7 concerns (HTTP method/path policies, header-based routing); ztunnel handles L4 (mTLS, telemetry)
- AuthorizationPolicy default-deny (`action: DENY, rules: [{}]`) + explicit ALLOW is the zero-trust posture
- Authorization principals use SPIFFE URIs (`cluster.local/ns/<ns>/sa/<sa>`) — identity-based, not IP-based
- Sidecar and ambient modes are mutually exclusive per namespace; both can coexist in the same cluster
- Introduce a mesh when cross-cutting network concerns become painful to implement per-service — not as a first step
