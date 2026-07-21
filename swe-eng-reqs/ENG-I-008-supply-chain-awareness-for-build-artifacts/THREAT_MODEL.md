# Threat / failure notes — ENG-I-008

| Threat | Disposition | Mitigation |
|--------|-------------|------------|
| T-5-17 Unsigned accept | mitigate | `Push` requires Verify before accept; scopes default-deny without `artifacts:push` |
| T-5-18 Key leakage | mitigate | Fixture keys only in `testdata/keys/`; never log private key material |
| T-5-SC Sigstore/cosign deps | mitigate | Stdlib `crypto/ed25519` only; `sigstore=false` in Info/demo/README |

- Authn/z for production registry would bind to H-004-style RBAC (out of scope here).
- SBOM is SPDX-inspired, not a signed SPDX attestation.
