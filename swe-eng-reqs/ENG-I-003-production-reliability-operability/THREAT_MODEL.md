# Threat / failure notes — ENG-I-003

- Authn/z: demo binds locally; production would require OIDC/RBAC.
- SLO input: identifiers, objectives, and SLI descriptions are validated before storage.
- T-4-08 (runbook information disclosure): the API returns file names and paths,
  not file contents. The checked-in runbook contains no credentials, customer
  data, or environment-specific secrets.
- Copyright: runbook procedures are original paraphrases and do not reproduce
  copyrighted book passages.
- Scope: this local kit does not expose a release gate API or an OTel exporter;
  those trust boundaries belong to ENG-N-005 and ENG-I-005.
