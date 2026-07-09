# Zero Trust in Kubernetes: Pod Security, RBAC, and Admission Control

Zero trust is a security model built on a single principle: **never trust, always verify**. No implicit trust based on network location — a pod inside the cluster is not inherently trusted just because it made it past the perimeter. Every access request is verified against identity and policy, regardless of where it originates.

This guide covers the zero-trust controls implemented at the Kubernetes level: admission policies enforced by Kyverno, RBAC-scoped ServiceAccounts, pod security contexts, and how these controls connect to the service mesh mTLS layer.

---

## Why Perimeter Security Fails in Cloud-Native

Traditional network security draws a perimeter — a firewall around the data center — and trusts everything inside it. Once inside, services communicate freely. This model has two fatal flaws in cloud-native environments:

**The perimeter is porous**: Pods run across shared nodes, shared networks, and public cloud infrastructure. A compromised pod, a misconfigured ingress, or a supply chain attack can put a malicious workload inside the perimeter with full trust.

**Lateral movement is unrestricted**: Once an attacker is inside, they can reach any other service on the internal network. A vulnerability in a low-value service becomes a path to high-value data.

Zero trust eliminates the implicit trust inside the perimeter. Every service must authenticate, every access must be authorized against explicit policy, every workload runs with minimum necessary privileges.

---

## Admission Control: Policies Enforced at Admission Time

Kubernetes admission controllers intercept API requests before resources are created or modified. Kyverno is a policy engine that lets you write policies as YAML that are enforced by a validating webhook.

The key property: **policies run at admission time**, not at runtime. A pod that violates a policy is rejected before it's scheduled. This is fundamentally different from scanning running containers for violations — by the time you detect a violation at runtime, the workload is already running.

### Deny Privileged Containers

```yaml
apiVersion: kyverno.io/v1
kind: ClusterPolicy
metadata:
  name: deny-privileged-pods
spec:
  validationFailureAction: Enforce
  rules:
    - name: deny-privileged-containers
      validate:
        foreach:
          - list: "request.object.spec.[containers, initContainers, ephemeralContainers][]"
            deny:
              conditions:
                any:
                  - key: "{{ element.securityContext.privileged || false }}"
                    operator: Equals
                    value: true
    - name: deny-host-namespaces
      validate:
        deny:
          conditions:
            any:
              - key: "{{ request.object.spec.hostPID || false }}"
                operator: Equals
                value: true
              - key: "{{ request.object.spec.hostIPC || false }}"
                operator: Equals
                value: true
              - key: "{{ request.object.spec.hostNetwork || false }}"
                operator: Equals
                value: true
```

`securityContext.privileged: true` gives the container root-level access to the host kernel — it can mount filesystems, modify network interfaces, load kernel modules. A privileged container escape is a full node compromise.

`hostPID`, `hostIPC`, and `hostNetwork` share host namespaces with the container. `hostNetwork: true` is the most commonly misconfigured — it lets the pod see all network traffic on the node and bypass network policies. These are occasionally legitimately needed (certain CNI plugins, node monitoring agents) but never in application workloads.

`validationFailureAction: Enforce` rejects the API request with an error message. `Audit` logs violations without blocking — useful when rolling out policies to existing clusters to understand impact before enforcing.

The `background: true` setting also evaluates policies against already-running resources periodically, not just at admission time. This catches resources created before the policy was installed.

---

### Require Non-Root Execution

```yaml
- name: require-run-as-non-root
  validate:
    foreach:
      - list: "request.object.spec.containers"
        deny:
          conditions:
            all:
              - key: "{{ request.object.spec.securityContext.runAsNonRoot || false }}"
                operator: Equals
                value: false
              - key: "{{ element.securityContext.runAsNonRoot || false }}"
                operator: Equals
                value: false
```

Running as root (UID 0) means a container escape gives the attacker root on the host. This policy requires either the pod-level or container-level `runAsNonRoot: true` to be set. The condition uses `all` — both pod and container level must be false to trigger the deny, meaning either one being true satisfies the policy.

The associated policy description says it plainly: *"Build your images with a non-root USER in the Dockerfile."* This is the source — the Dockerfile sets a non-root user, and the Kubernetes policy verifies that the deployed pod respects it.

```yaml
- name: require-read-only-root-filesystem
  validate:
    foreach:
      - list: "request.object.spec.containers"
        deny:
          conditions:
            any:
              - key: "{{ element.securityContext.readOnlyRootFilesystem || false }}"
                operator: Equals
                value: false
```

A read-only root filesystem prevents malware from writing binaries into the container's filesystem at runtime. If an attacker achieves code execution inside a container, they can't persist a backdoor binary. Applications that need to write files should use explicitly mounted volumes (`emptyDir` for temp files, persistent volumes for data).

---

### Resource Limits as a Security Control

```yaml
apiVersion: kyverno.io/v1
kind: ClusterPolicy
metadata:
  name: require-resource-limits
spec:
  validationFailureAction: Enforce
  rules:
    - name: check-container-limits
      validate:
        foreach:
          - list: "request.object.spec.containers"
            deny:
              conditions:
                any:
                  - key: "{{ element.resources.limits.cpu || '' }}"
                    operator: Equals
                    value: ""
                  - key: "{{ element.resources.limits.memory || '' }}"
                    operator: Equals
                    value: ""
```

Resource limits are typically discussed as a reliability concern — preventing noisy neighbors. They're also a security control. A compromised pod that wants to run a cryptocurrency miner or a DoS attack wants unlimited CPU. Resource limits bound the blast radius.

The policy description makes this explicit: *"Pods without resource limits are a noisy-neighbor DoS vector — one misbehaving pod can starve the node."* A security incident that degrades the entire node through resource exhaustion is as damaging as one that exfiltrates data.

---

## Least-Privilege ServiceAccounts

Kubernetes ServiceAccounts are the identity mechanism for pods. By default, every pod gets a token for the `default` ServiceAccount, which often has more permissions than needed. The most common finding in cloud security audits is ServiceAccounts with cluster-admin or wide RBAC permissions.

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: orders-service
  namespace: orders
  annotations:
    eks.amazonaws.com/role-arn: "arn:aws:iam::123456789012:role/orders-service-role"
automountServiceAccountToken: false
```

`automountServiceAccountToken: false` prevents the default service account token from being mounted at `/var/run/secrets/kubernetes.io/serviceaccount/token`. Most application pods don't need to call the Kubernetes API — they shouldn't have credentials to do so.

The IRSA annotation (`eks.amazonaws.com/role-arn`) maps this ServiceAccount to an AWS IAM role scoped to exactly the S3 bucket the service legitimately needs. No `AdministratorAccess`, no wildcard policies.

The RBAC binding is minimally scoped:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: orders-service-role
  namespace: orders
rules:
  - apiGroups: [""]
    resources: ["configmaps"]
    resourceNames: ["orders-config", "orders-feature-flags"]
    verbs: ["get", "watch"]
  - apiGroups: [""]
    resources: ["secrets"]
    resourceNames: ["orders-db-credentials"]
    verbs: ["get"]
```

This ServiceAccount can only read two specific ConfigMaps and one specific Secret in the `orders` namespace. Nothing else. The `resourceNames` constraint is important — it prevents the ServiceAccount from reading *all* ConfigMaps or *all* Secrets, just the named ones.

---

## Projected Service Account Tokens

The pod spec uses projected tokens instead of the auto-mounted long-lived token:

```yaml
volumes:
  - name: service-account-token
    projected:
      sources:
        - serviceAccountToken:
            path: token
            expirationSeconds: 3600
            audience: "payment-service"
```

The differences from the legacy token:

- **Audience-scoped**: This token is only valid when presented to `payment-service`. A compromised pod that steals the token can't use it to call other services.
- **Short-lived**: Expires in 1 hour. The kubelet rotates it automatically. Legacy tokens are long-lived and don't expire.
- **Pod-bound**: Automatically revoked when the pod is deleted.

---

## Pod Security Context: Defense in Depth

The example pod spec in `least-privilege-sa.yaml` shows the full defense-in-depth security context:

```yaml
securityContext:
  runAsNonRoot: true
  runAsUser: 1000
  runAsGroup: 1000
  readOnlyRootFilesystem: true
  allowPrivilegeEscalation: false
  capabilities:
    drop: ["ALL"]
```

- `runAsNonRoot: true` + `runAsUser: 1000`: Container runs as a non-root user with UID 1000
- `readOnlyRootFilesystem: true`: Can't write to the container filesystem
- `allowPrivilegeEscalation: false`: The process can't gain more privileges than its parent (prevents `setuid` escalation)
- `capabilities.drop: ["ALL"]`: Drops all Linux capabilities. This removes `NET_BIND_SERVICE`, `CHOWN`, `SETUID`, `SETGID`, and dozens of others. Applications running on port >1024 don't need any capabilities.

These settings are enforced by Kyverno policies at admission time (non-root, read-only filesystem). Kyverno can also validate that `allowPrivilegeEscalation: false` and capabilities are dropped — this is where the policy-as-code model pays for itself: you write the policy once and every deployment across all namespaces must comply.

---

## Connecting to the Service Mesh

The zero-trust controls in this module operate at the workload identity layer. The service mesh mTLS layer (pattern 08) extends this to network communication:

- **Kyverno + RBAC**: Controls what workloads can do (Kubernetes API access, file system, process privileges, AWS resources)
- **Istio AuthorizationPolicy**: Controls which services can call which endpoints, verified by SPIFFE certificate
- **PeerAuthentication STRICT**: Rejects plaintext traffic, enforcing that all communication uses mTLS

Together, these layers implement defense in depth:

1. Admission control prevents dangerous pod configurations from being scheduled
2. RBAC limits what the pod can do in the Kubernetes and AWS control planes
3. Network policy limits which pods can initiate connections
4. mTLS + AuthorizationPolicy enforces which services can talk to which endpoints, verified by cryptographic identity

No single layer is sufficient. A pod running without resource limits (bypassing policy) could still be blocked by network policy. A service with cluster-admin RBAC might be constrained by Istio authorization. Each layer catches failures in the others.

---

## Operational Complexity

Zero trust is not free:

- **Policy management**: ClusterPolicies need to be maintained and updated as requirements change. Exclusions for system namespaces must be correct or critical infrastructure breaks.
- **Developer friction**: Developers used to running as root in containers will get rejected. Requires documentation, example pod specs, and a process for exceptions.
- **Audit and reporting**: `validationFailureAction: Audit` mode and Kyverno policy reports help understand compliance posture before switching to Enforce.
- **Compatibility**: Some third-party tools and operators require elevated privileges. Maintain a documented exemption list.

The recommendation for new clusters: start with `Audit` mode to understand your current violation rate, fix violations namespace by namespace, then switch to `Enforce`.

---

## Key Takeaways

- Zero trust assumes breach: no implicit trust from network location, every access verified against identity and policy
- Admission control (Kyverno) enforces policies at deploy time, not at runtime — misconfigurations are rejected before scheduling
- `validationFailureAction: Enforce` blocks violations; `Audit` logs them without blocking — use Audit when rolling out new policies
- `background: true` evaluates policies against already-running resources, not just new ones
- Privileged containers, host namespaces, and root UIDs are the three most critical controls — block all of them for application workloads
- Resource limits are both a reliability and a security control — they bound the blast radius of a compromised container
- `automountServiceAccountToken: false` + projected tokens with audience scoping and 1-hour TTL replaces long-lived default tokens
- `capabilities.drop: ["ALL"]` is achievable for most application workloads and removes dozens of potential escalation vectors
- Kyverno + RBAC + Network Policy + Istio mTLS are complementary layers — no single layer is sufficient
- Roll out with Audit mode first; fix violations before switching to Enforce to avoid breaking running workloads
