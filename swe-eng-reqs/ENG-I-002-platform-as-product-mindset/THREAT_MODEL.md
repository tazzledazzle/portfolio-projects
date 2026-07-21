# Threat / failure notes — ENG-I-002

## Trust boundary
client → product metrics API (untrusted team labels for adoption)

## Mitigations
- Demo binds locally (`demo-local` on `:18602`); production would require authn before mutating adoption.
- Golden-path content is read from local template files — no remote fetch.
- Empty team names ignored on `RecordAdoption`.

## Out of scope
- IDP project/pipeline/environment catalog (ENG-E-005)
- Ticket-removal automation metrics (ENG-I-006)
