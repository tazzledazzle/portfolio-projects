# SRE Communications Checklist

Practical checklists for SREs (and adjacent roles) to improve consistency around time zones, handoffs, incidents, and daily communication. Adapt to your team’s tools (Slack, PagerDuty, email) and norms.

---

## 1. Time-zone awareness (distributed teams)

**Before your workday**

- [ ] Confirm today’s required support or overlap window (e.g. 6 AM–2 PM PT).
- [ ] Check calendar for early meetings (e.g. before 8 AM local).
- [ ] Pre-check Slack/email for anything that arrived from other-timezone teammates.
- [ ] If early start is required, set alarms and prep the night before (workstation, quick breakfast).

**During the day**

- [ ] Convert times when scheduling or reading messages (e.g. “12 PM ET = 9 AM PT”).
- [ ] Note when other-timezone teammates typically go offline; surface blockers or requests before they leave.
- [ ] Use status (e.g. “On PT — online early, async welcome”) so others know when to expect replies.
- [ ] Use scheduled send for non-urgent messages so they land in the recipient’s working hours.

**End of day**

- [ ] Review unresolved threads; respond or hand off before others log off.
- [ ] Document anything handed over or pending.
- [ ] Set clear status at sign-off (e.g. “Done for today — back at 6 AM PT”).
- [ ] Prepare tomorrow’s first tasks to reduce morning load.

---

## 2. Leaving desk briefly (5–30 minutes)

**Before leaving**

- [ ] Check if you’re in an incident or time-sensitive work.
- [ ] If yes, notify incident channel or partner explicitly.
- [ ] Update status (e.g. “BRB ~10 min — reachable on mobile if urgent”).

**Template**

> Stepping away for ~10 minutes. Ping me if something is urgent — I’ll keep notifications on.

---

## 3. Incident communication

**When you join an incident**

- [ ] Announce in the incident channel (e.g. “Joining, taking comms” or “Joining, focusing on [area]”).
- [ ] Confirm comms lead and where updates will be posted (Slack, status page, email).

**When posting updates**

- [ ] Use a consistent format: **What changed / What we’re doing / ETA or next update.**
- [ ] Post on a regular cadence (e.g. every 15–30 min) even if “no change.”
- [ ] When handing off, summarize state and next steps in one short message.

**When you leave an incident**

- [ ] Hand off explicitly: “Stepping back; [Name] has comms. Current state: … Next: …”
- [ ] Ensure someone else is clearly responsible before you go.

---

## 4. Handoffs and PTO

**Before PTO or long absence**

- [ ] Document in-progress work and owners in a shared doc or ticket.
- [ ] Designate a backup for your primary responsibilities.
- [ ] Set OOO and update status; point to the handoff doc or backup in your status message.

**When receiving a handoff**

- [ ] Read the handoff doc; ask clarifying questions before the other person leaves.
- [ ] Confirm you’re the backup in the same place the team checks (e.g. runbook, on-call schedule).

---

## 5. Async-friendly habits

- Prefer written summaries for decisions and context so others can catch up asynchronously.
- Use threads for follow-ups so the main channel stays scannable.
- Tag people only when you need a response from them; use @channel sparingly and per team policy.
