# Error Budget Response

Use this procedure when a service is consuming its error budget faster than
planned. It is an original operational checklist for this demo; it does not
reproduce text from an external book.

## Symptoms

- The error ratio or latency objective is outside its target.
- Recent budget consumption is materially above the expected pace.
- Operators are considering whether delivery work should pause.

## Check burn

1. Confirm the SLI query covers the intended service and request population.
2. Compare short and longer observation windows to distinguish a spike from a
   sustained regression.
3. Check deploy, dependency, capacity, and traffic changes near the start of
   the burn.

## Mitigate

1. Reduce impact first: roll back the suspected change, shed optional work, or
   add safe capacity.
2. Assign an incident owner and record each mitigation with its observed
   result.
3. Recheck the SLI after every action; avoid stacking unmeasured changes.

## Communicate and recover

1. Tell affected teams the current impact, owner, and next update time.
2. Resume normal delivery only after the burn stabilizes and the recovery plan
   has an owner.
3. Capture follow-up work for detection, capacity, testing, and runbook
   improvements without embedding credentials or customer data.
