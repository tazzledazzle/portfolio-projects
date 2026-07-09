# CQRS and Event Sourcing: Commands, Events, and Projections

CQRS (Command Query Responsibility Segregation) and event sourcing are two patterns that are often paired but are independent concepts. CQRS separates the model used to update data from the model used to read it. Event sourcing replaces the current-state database model with an immutable log of events — the current state of anything is always derived by replaying those events.

Together, they give you a complete audit trail, the ability to rebuild any past state, and independent optimization of read and write paths. They also add significant complexity. This guide covers the Kotlin implementation of both patterns using an inventory management domain.

---

## CQRS: Separate Models for Reads and Writes

In a traditional CRUD system, a single `Inventory` table serves both reads and writes. You insert and update rows, and you query the same rows to display current inventory.

CQRS splits this into two explicit models:

**Write side (command model)**: Accepts commands, enforces business rules, produces events. Optimized for correctness and consistency. In this codebase, this is `InventoryAggregate`.

**Read side (query model)**: Projects a view of the data optimized for reading. Optimized for query performance and flexibility. In this codebase, this is `InventoryProjection`.

The two sides communicate through events: the write side produces them, the read side consumes them to update its view.

Why bother? Because the access patterns are different. Writes need strong consistency and invariant enforcement. Reads need fast queries across many fields, often with different groupings. Optimizing a single model for both produces a model that's mediocre at both.

---

## Event Sourcing: State as an Immutable Event Log

Standard persistence stores the *current state* of an entity. An `Inventory` row has a `quantity` column. When inventory is deducted, you run `UPDATE inventory SET quantity = quantity - 10`. The previous state is gone.

Event sourcing stores the *history of changes* instead. Inventory deductions, restocks, and initializations are each recorded as separate immutable event records. The current quantity is computed by replaying them:

```
InventoryInitialized(sku=WIDGET, quantity=100)
InventoryDeducted(sku=WIDGET, quantity=15, remaining=85)
InventoryRestocked(sku=WIDGET, quantity=50, new=135)
InventoryDeducted(sku=WIDGET, quantity=20, remaining=115)
```

Current quantity: 115. You can also compute what the quantity was at any point in time. You can't do that with a mutable row.

This is the source of truth. Not the projection. Not the read model. The event log.

---

## Domain Events

```kotlin
sealed interface DomainEvent {
    val aggregateId: String
    val occurredAt: Instant
    var sequenceNumber: Long  // Set by EventStore on append; 0 until stored
}

data class InventoryDeducted(
    override val aggregateId: String,
    val sku: String,
    val quantity: Int,
    val remainingQuantity: Int,
    val reason: String,
    override val occurredAt: Instant = Instant.now(),
    override var sequenceNumber: Long = 0,
) : DomainEvent

data class InventoryBelowReorderThreshold(
    override val aggregateId: String,
    val sku: String,
    val currentQuantity: Int,
    val reorderThreshold: Int,
    // ...
) : DomainEvent
```

Events are named in the past tense — they record facts that have already happened (`InventoryDeducted`, not `DeductInventory`). They are immutable once created. The `sequenceNumber` is assigned by the EventStore when appended, giving each event a globally ordered position.

`InventoryBelowReorderThreshold` is worth noting: it's a derived event, not caused by a command directly. When a deduction causes stock to fall below the reorder threshold, a second event is produced automatically. Downstream services (a purchasing system) can subscribe to this event without the inventory service needing to call them.

---

## The Aggregate: Enforcing Invariants

```kotlin
class InventoryAggregate(val aggregateId: String) {
    var quantity: Int = 0
        private set
    var initialized: Boolean = false
        private set

    fun handle(cmd: DeductInventoryCommand): List<DomainEvent> {
        require(initialized) { "Inventory for ${cmd.sku} is not initialized" }
        require(cmd.quantity > 0) { "Deduction quantity must be positive" }
        check(quantity >= cmd.quantity) {
            "Insufficient inventory: have $quantity units of ${cmd.sku}, requested ${cmd.quantity}"
        }

        val remaining = quantity - cmd.quantity
        val deducted = InventoryDeducted(
            aggregateId = aggregateId,
            sku = sku,
            quantity = cmd.quantity,
            remainingQuantity = remaining,
            reason = cmd.reason,
        )
        apply(deducted)
        pendingEvents.add(deducted)

        if (remaining < reorderThreshold) {
            val belowThreshold = InventoryBelowReorderThreshold(...)
            apply(belowThreshold)
            pendingEvents.add(belowThreshold)
        }

        return drainPendingEvents()
    }
```

The pattern for command handlers:
1. Validate the command against current state (throw if invalid)
2. Produce one or more events
3. Call `apply()` to update state
4. Add the event to `pendingEvents`

The `apply()` methods are the state mutators:

```kotlin
fun apply(event: DomainEvent) {
    when (event) {
        is InventoryInitialized -> {
            sku = event.sku
            quantity = event.initialQuantity
            reorderThreshold = event.reorderThreshold
            initialized = true
        }
        is InventoryDeducted -> {
            quantity -= event.quantity
        }
        is InventoryRestocked -> {
            quantity = event.newQuantity
        }
        is InventoryBelowReorderThreshold -> {
            // Informational — no state change in the aggregate
        }
    }
}
```

`apply()` is called both when handling a new command AND when replaying events from the event store on load. This is the key invariant: **`apply()` must be pure**. No external calls, no I/O, no side effects — just state mutations. The same sequence of events must always produce the same state.

---

## The Event Store

```kotlin
class EventStore(private val snapshotFrequency: Int = 10) {

    fun append(aggregateId: String, newEvents: List<DomainEvent>, aggregate: InventoryAggregate? = null) {
        val stream = events.getOrPut(aggregateId) { mutableListOf() }
        for (event in newEvents) {
            event.sequenceNumber = globalSequence.incrementAndGet()
            stream.add(event)
        }

        // Auto-snapshot every N events
        if (stream.size % snapshotFrequency == 0 && aggregate != null) {
            snapshots[aggregateId] = aggregate.toSnapshot(stream.last().sequenceNumber)
        }
    }

    fun load(aggregateId: String): InventoryAggregate {
        val aggregate = InventoryAggregate(aggregateId)
        val stream = events[aggregateId] ?: return aggregate

        val snapshot = snapshots[aggregateId]
        if (snapshot != null) {
            aggregate.restoreFromSnapshot(snapshot)
        }

        val replayFrom = snapshot?.snapshotAtSequence ?: 0L
        val eventsToReplay = stream.filter { it.sequenceNumber > replayFrom }
        for (event in eventsToReplay) {
            aggregate.apply(event)
        }
        return aggregate
    }
}
```

The EventStore has two responsibilities: appending new events and loading aggregates by replaying events.

The load algorithm:
1. Check for a snapshot. If one exists, restore the aggregate state from it (skipping all events up to the snapshot sequence)
2. Replay only events after the snapshot sequence
3. Return the aggregate in its current state

Without snapshots, loading an aggregate with 10,000 events means replaying all 10,000. The snapshot every 10 events means at most 9 events need to be replayed in addition to snapshot restoration. In production with EventStoreDB, snapshots are a first-class feature with configurable strategies.

### Production EventStore Options

- **EventStoreDB**: Purpose-built for event sourcing. Native subscriptions, projections, and snapshots. The richest feature set.
- **PostgreSQL events table**: Simple append-only table. Works well for moderate volume. Subscriptions require polling or LISTEN/NOTIFY.
- **Kafka as the event log**: Excellent for fan-out to many consumers. Lacks efficient per-aggregate point-in-time queries — querying "all events for aggregate X" requires scanning a partition.

---

## Projections: The Read Side

```kotlin
class InventoryProjection {
    private val views = mutableMapOf<String, InventoryView>()

    fun on(event: DomainEvent) {
        when (event) {
            is InventoryInitialized -> {
                views[event.sku] = InventoryView(
                    sku = event.sku,
                    currentQuantity = event.initialQuantity,
                    reorderThreshold = event.reorderThreshold,
                    lastEventSequence = event.sequenceNumber,
                )
            }
            is InventoryDeducted -> {
                views.compute(event.sku) { _, existing ->
                    existing?.copy(
                        currentQuantity = event.remainingQuantity,
                        lastEventSequence = event.sequenceNumber,
                    )
                }
            }
            // ...
        }
    }

    fun getInventory(sku: String): InventoryView? = views[sku]
    fun getLowStockSkus(): List<InventoryView> =
        views.values.filter { it.belowReorderThreshold }
}
```

The projection listens to events and maintains a denormalized view optimized for reading. `InventoryView` is the query model — a snapshot of current state derived from the event stream.

The docstring notes the most important property to surface to clients:

```kotlin
// Consistency note: projections are EVENTUALLY CONSISTENT.
// A command that just succeeded may not yet be reflected here.
// Surface this in your API: "Inventory levels as of [lastUpdatedAt]"
```

`lastEventSequence` in `InventoryView` lets callers know how fresh the data is. If a client needs to confirm its just-submitted command is reflected, it can poll until the sequence number advances.

### Event Replay: A Superpower

Because the event log is the source of truth, you can:

- **Debug past failures**: Replay events up to the point of failure and inspect state
- **Create new projections**: Add a new read model without migrating data — just replay all events through it
- **Fix a bad projection**: If a projection bug corrupts the read model, delete and rebuild from the event log
- **Audit trail**: Every change to inventory is recorded with who did it, when, and why

This is the core value proposition of event sourcing beyond CQRS. The audit log is not an afterthought appended to the main system — it *is* the main system.

---

## Tradeoffs

**Complexity**: Two models instead of one. Command handling, event production, event store writes, projection updates. More moving parts than CRUD.

**Eventual consistency on reads**: The projection may lag behind the aggregate by milliseconds or more. Applications need to handle this gracefully.

**Schema evolution**: Events are immutable, but their interpretation may need to change. Upcasting (transforming old events to new schema on replay) is a real engineering concern.

**Debugging unfamiliar**: Developers used to `SELECT * FROM inventory WHERE sku=?` need to shift to thinking in event streams.

**When it's worth it**: Systems with complex business rules and invariants, audit requirements, temporal queries ("what was the state at 3pm last Tuesday?"), and multiple consumers of state changes.

**When to use simpler CRUD**: Internal admin tools, simple CRUD entities without business rules, greenfield with a small team unfamiliar with the pattern.

---

## Key Takeaways

- CQRS separates the write model (commands → events → aggregate state) from the read model (projections for queries) — each optimized for its access pattern
- Event sourcing stores immutable events as the source of truth; current state is always derived by replaying events
- `apply()` methods are pure state mutators — they must produce the same state given the same events, with no side effects
- The EventStore `load()` algorithm: restore from snapshot (if any), then replay only events after the snapshot sequence
- Snapshots prevent O(n) replay on every load — take one every N events
- Projections are eventually consistent — surface `lastEventSequence` so clients can know data freshness
- Event replay is a debugging and migration superpower: rebuild any past state, fix projection bugs, or add new read models without migrating data
- CQRS+ES is genuinely complex — use it when audit requirements, temporal queries, or complex invariants justify the overhead
- In production, prefer EventStoreDB for purpose-built event sourcing or PostgreSQL for simplicity; Kafka works for fan-out but struggles with per-aggregate queries
