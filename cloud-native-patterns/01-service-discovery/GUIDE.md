# Service Discovery in Kubernetes: DNS, ClusterIP, and Headless Services

Service discovery is the mechanism that lets one microservice find and talk to another without hardcoding IP addresses. In Kubernetes, this is largely solved at the platform level — but understanding *how* it works determines whether your services are observable, reliable, and correctly load-balanced.

This guide walks through Kubernetes service discovery using a real e-commerce workload: an `orders-service` that needs to reach a `payment-service` and `inventory-service`, backed by a Kafka cluster running as a StatefulSet.

---

## The Core Problem

In a containerized environment, pod IPs are ephemeral. A pod is rescheduled, it gets a new IP. Three replicas of a service each have different IPs. The set of healthy pods changes constantly as deployments roll and autoscalers act.

Hardcoding IPs is obviously wrong. But what replaces them?

In Kubernetes, the answer is: **DNS names backed by Service resources**. The platform takes responsibility for mapping a stable name to the current set of healthy pods, and for routing traffic to them.

---

## DNS-Based Discovery in Practice

The `orders-service` Deployment makes this concrete. Rather than configuring downstream IPs, it uses fully qualified DNS names:

```yaml
env:
  - name: PAYMENT_SERVICE_URL
    value: "http://payment-service.payments.svc.cluster.local:8080"
  - name: INVENTORY_SERVICE_URL
    value: "http://inventory-service.inventory.svc.cluster.local:8080"
```

The format is `<service>.<namespace>.svc.cluster.local`. Within the same namespace, the short form (`payment-service`) resolves via the ndots search path — CoreDNS appends the cluster domain suffix automatically.

What's actually happening when `orders-service` sends a request to `payment-service.payments.svc.cluster.local`:

1. The pod's DNS resolver (pointing at CoreDNS) resolves the name
2. CoreDNS returns the ClusterIP of the `payment-service` Service
3. The kernel routes the packet to that ClusterIP
4. kube-proxy rewrites the destination IP to one of the healthy pod IPs (using iptables or IPVS rules)
5. The packet arrives at an actual pod

The application never sees a pod IP. It sees a stable virtual IP that exists for the lifetime of the Service resource.

---

## ClusterIP Services: The Standard Pattern

The `orders-service` Service resource is a ClusterIP:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: orders-service
  namespace: ecommerce
spec:
  type: ClusterIP
  selector:
    app: orders-service
  ports:
    - name: http
      port: 8080
      targetPort: http
    - name: grpc
      port: 9090
      targetPort: grpc
```

A ClusterIP Service gets a virtual IP that's reachable from anywhere in the cluster. The `selector` tells Kubernetes which pods receive traffic — pods with the label `app: orders-service`. As pods are added, removed, or fail readiness checks, the endpoints list updates automatically.

Named ports (`targetPort: http`) are worth using. They decouple the port number the Service listens on from the container's actual port. If the application ever needs to change its listening port, you update the container spec without touching the Service.

The commented-out `sessionAffinity: ClientIP` setting is worth noting: sticky sessions break load balancing. For stateless services, round-robin distribution (the default) gives better utilization. Only enable session affinity if the application genuinely requires it and you've accepted the tradeoff.

---

## Readiness Probes: The Real Key to Reliable Discovery

Here is the comment from the Deployment that captures the most important insight in this entire pattern:

```yaml
# Readiness probe is the critical signal — a pod is only included in
# DNS / iptables routing AFTER this passes. Getting this right matters
# more than which service discovery mechanism you pick.
readinessProbe:
  httpGet:
    path: /healthz/ready
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 5
  failureThreshold: 3
  successThreshold: 1
```

A pod is only added to the Service's endpoint list — and therefore to DNS responses and kube-proxy rules — after its readiness probe passes. This means traffic is never routed to a pod that hasn't finished starting up, hasn't connected to its database, or has entered a degraded state.

The three-probe setup in this Deployment is the production standard:

- **startupProbe**: prevents liveness from killing the pod during a slow initial start (JVM warmup, schema migration). Fails for up to 60 seconds (`failureThreshold=30` × `periodSeconds=2`) before liveness takes over.
- **readinessProbe**: controls whether the pod receives traffic. Checked every 5 seconds; a pod is removed from rotation after 3 consecutive failures.
- **livenessProbe**: restarts the pod if it's deadlocked or otherwise unresponsive to recovery. Uses a longer check interval (10s) to avoid restart storms.

Getting readiness wrong is the most common cause of 502 errors during rolling deployments. If your readiness check is too permissive (passes before the app is ready), you'll send traffic to pods that aren't ready. If it's too strict (fails during normal operation), healthy pods get removed from rotation unnecessarily.

---

## Headless Services: When Clients Need Pod Identity

Not all services want a ClusterIP. Stateful systems like Kafka, Cassandra, and Elasticsearch need clients to address specific pod instances — not an arbitrary one behind a VIP. The headless Service pattern handles this:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: kafka-broker
  namespace: kafka
spec:
  clusterIP: None   # This makes it headless
  selector:
    app: kafka-broker
  ports:
    - name: kafka
      port: 9092
```

`clusterIP: None` tells Kubernetes not to allocate a virtual IP. Instead, DNS queries for `kafka-broker.kafka.svc.cluster.local` return *all* pod IPs directly — one A record per ready pod. The client is responsible for deciding which pod to talk to.

For a three-broker Kafka cluster running as a StatefulSet, the DNS layout is:

```
kafka-broker-0.kafka-broker.kafka.svc.cluster.local → pod IP of broker 0
kafka-broker-1.kafka-broker.kafka.svc.cluster.local → pod IP of broker 1
kafka-broker-2.kafka-broker.kafka.svc.cluster.local → pod IP of broker 2
```

This is the format Kafka clients use for `bootstrap.servers`. The client connects to one of the listed brokers, learns the full cluster topology from it, and then connects directly to the leader for each partition.

---

## StatefulSets and Stable Network Identity

Headless Services work in concert with StatefulSets to give each pod a stable, predictable DNS name. The StatefulSet must reference the headless Service by name:

```yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: kafka-broker
spec:
  serviceName: "kafka-broker"   # Must match the headless Service name
  replicas: 3
```

This binding is what causes Kubernetes to assign the stable DNS names. Without `serviceName`, pods would get random DNS names that change on restart. With it, `kafka-broker-0` always resolves to the pod with ordinal 0, even after that pod is rescheduled to a different node.

The pod ordinal also feeds directly into Kafka broker identity:

```yaml
- name: KAFKA_BROKER_ID
  valueFrom:
    fieldRef:
      fieldPath: metadata.labels['apps.kubernetes.io/pod-index']
- name: KAFKA_ADVERTISED_LISTENERS
  value: "INTERNAL://$(POD_NAME).kafka-broker.$(POD_NAMESPACE).svc.cluster.local:9093,
          EXTERNAL://$(POD_NAME).kafka-broker.$(POD_NAMESPACE).svc.cluster.local:9092"
```

The broker advertises itself using its stable DNS name. Clients and peer brokers can always reach it at that name, regardless of which node it's running on.

### VolumeClaimTemplates: Storage Identity

StatefulSets give each pod a persistent volume with a name tied to the pod's ordinal:

```yaml
volumeClaimTemplates:
  - metadata:
      name: kafka-data
    spec:
      accessModes: ["ReadWriteOnce"]
      storageClassName: gp3-encrypted
      resources:
        requests:
          storage: 100Gi
```

`kafka-broker-0` gets PVC `kafka-data-kafka-broker-0`, `kafka-broker-1` gets `kafka-data-kafka-broker-1`. When a pod is rescheduled, it rebinds to its own PVC — not another broker's data. This is what makes StatefulSets suitable for storage systems.

### Anti-Affinity: Spread Across Nodes

The StatefulSet also enforces broker spread across nodes:

```yaml
affinity:
  podAntiAffinity:
    requiredDuringSchedulingIgnoredDuringExecution:
      - labelSelector:
          matchExpressions:
            - key: app
              operator: In
              values: [kafka-broker]
        topologyKey: "kubernetes.io/hostname"
```

`requiredDuringScheduling` means the scheduler will refuse to place two Kafka brokers on the same node. `topologyKey: kubernetes.io/hostname` means the isolation boundary is the node. This prevents a single node failure from taking out all Kafka replicas.

---

## Graceful Termination and the preStop Hook

DNS discovery and readiness probes handle pod startup. But there's a symmetrical concern at shutdown: traffic can be routed to a pod for a brief window after it starts terminating, because kube-proxy's iptables rules and DNS caches don't update instantaneously.

The preStop hook in the Deployment addresses this:

```yaml
lifecycle:
  preStop:
    exec:
      command: ["/bin/sh", "-c", "sleep 5"]
```

When a pod is deleted, Kubernetes sends SIGTERM and simultaneously starts removing the pod from Service endpoints. The `sleep 5` gives the endpoint removal time to propagate before the process begins its own shutdown. Combined with the 30-second `terminationGracePeriodSeconds`, the pod has time to finish in-flight requests before the process exits.

---

## Client-Side vs. Server-Side Discovery

This implementation uses **server-side discovery**: the platform (CoreDNS + kube-proxy) handles routing. The client calls a stable DNS name; the platform handles the rest.

The alternative is **client-side discovery**: the client queries a service registry (Eureka, Consul), gets a list of healthy pod IPs, and picks one itself. This gives the client control over load balancing algorithms (round-robin, least-connections, consistent hashing) but introduces a dependency on the registry and additional complexity in every service.

For most Kubernetes workloads, server-side discovery is the right default. The platform handles health checking, load distribution, and endpoint management. Client-side discovery is worth considering when:

- You need advanced load balancing (consistent hashing for cache locality)
- You're running across multiple clusters where DNS zones don't overlap
- You need to spread load unevenly (canary weights) without a service mesh

---

## Key Takeaways

- Kubernetes DNS-based service discovery gives services stable names that survive pod churn — configure downstream addresses as `<service>.<namespace>.svc.cluster.local`
- ClusterIP Services provide a stable virtual IP; kube-proxy handles actual pod selection via iptables/IPVS
- Readiness probes control whether a pod receives traffic — getting these right matters more than which discovery mechanism you choose
- Headless Services (`clusterIP: None`) return individual pod IPs from DNS — required when clients need to address specific instances (Kafka, Cassandra)
- StatefulSets + headless Services give each pod a stable, predictable DNS name tied to its ordinal
- VolumeClaimTemplates bind storage to pod identity — `pod-0` always gets its own PVC on restart
- Pod anti-affinity spreads replicas across nodes — `requiredDuringScheduling` makes this a hard constraint
- The preStop hook + terminationGracePeriodSeconds prevents traffic hitting pods that are mid-shutdown
- For most Kubernetes workloads, platform-level service discovery (ClusterIP + CoreDNS) is the right default; client-side discovery is for advanced scenarios
