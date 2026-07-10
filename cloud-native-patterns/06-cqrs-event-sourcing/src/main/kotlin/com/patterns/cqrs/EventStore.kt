package com.patterns.cqrs

import com.patterns.cqrs.aggregate.InventoryAggregate
import com.patterns.cqrs.aggregate.InventorySnapshot
import com.patterns.cqrs.events.DomainEvent
import org.slf4j.LoggerFactory
import java.util.concurrent.ConcurrentHashMap
import java.util.concurrent.atomic.AtomicLong

/**
 * In-memory event store.
 *
 * Production alternatives:
 *  - EventStoreDB (purpose-built, native event sourcing, subscriptions, projections)
 *  - PostgreSQL with an events table (simple, but manual subscription/projection wiring)
 *  - Kafka as event log (great for fan-out, but lacks point-in-time query by aggregate ID)
 *
 * This implementation provides:
 *  - Append-only event storage per aggregate ID
 *  - Monotonic sequence numbers across all aggregates (global ordering)
 *  - Snapshot support: stores a snapshot after every [snapshotFrequency] events
 *  - Load with snapshot: fetch snapshot + replay only events after it
 */
class EventStore(private val snapshotFrequency: Int = 10) {
    private val log = LoggerFactory.getLogger(EventStore::class.java)

    // Events per aggregate: aggregateId → ordered list of DomainEvent
    private val events = ConcurrentHashMap<String, MutableList<DomainEvent>>()

    // Latest snapshot per aggregate
    private val snapshots = ConcurrentHashMap<String, InventorySnapshot>()

    // Global sequence counter — each event gets a unique monotonic sequence number
    private val globalSequence = AtomicLong(0)

    /**
     * Appends events to the event log for the given aggregate.
     *
     * Each event is assigned a global sequence number. If the aggregate has reached
     * the [snapshotFrequency] threshold, a snapshot is automatically taken.
     *
     * In production, this would include optimistic concurrency control:
     * append only if current version matches expectedVersion, else throw concurrency exception.
     */
    fun append(aggregateId: String, newEvents: List<DomainEvent>, aggregate: InventoryAggregate? = null) {
        val stream = events.getOrPut(aggregateId) { mutableListOf() }

        for (event in newEvents) {
            event.sequenceNumber = globalSequence.incrementAndGet()
            stream.add(event)
            log.debug(
                "[EventStore] Appended {} seq={} aggregateId={}",
                event::class.simpleName,
                event.sequenceNumber,
                aggregateId,
            )
        }

        // Auto-snapshot every N events
        if (stream.size % snapshotFrequency == 0 && aggregate != null) {
            val snap = aggregate.toSnapshot(stream.last().sequenceNumber)
            snapshots[aggregateId] = snap
            log.info(
                "[EventStore] Snapshot taken for {} at seq={} qty={}",
                aggregateId,
                snap.snapshotAtSequence,
                snap.quantity,
            )
        }
    }

    /**
     * Loads an [InventoryAggregate] by replaying its events.
     *
     * Algorithm:
     * 1. Check for a snapshot. If found, restore aggregate state from it.
     * 2. Load only events AFTER the snapshot's sequence number.
     * 3. Replay those events to bring the aggregate to current state.
     *
     * Without snapshots, step 2 replays all events. For aggregates with thousands
     * of events this can be slow — the snapshot prevents O(n) replay on every load.
     */
    fun load(aggregateId: String): InventoryAggregate {
        val aggregate = InventoryAggregate(aggregateId)
        val stream = events[aggregateId] ?: return aggregate

        val snapshot = snapshots[aggregateId]
        if (snapshot != null) {
            aggregate.restoreFromSnapshot(snapshot)
            log.debug(
                "[EventStore] Restored {} from snapshot at seq={} (skipping {} events)",
                aggregateId,
                snapshot.snapshotAtSequence,
                stream.count { it.sequenceNumber <= snapshot.snapshotAtSequence },
            )
        }

        val replayFrom = snapshot?.snapshotAtSequence ?: 0L
        val eventsToReplay = stream.filter { it.sequenceNumber > replayFrom }

        log.debug(
            "[EventStore] Replaying {} event(s) for {} (after seq={})",
            eventsToReplay.size,
            aggregateId,
            replayFrom,
        )

        for (event in eventsToReplay) {
            aggregate.apply(event)
        }

        return aggregate
    }

    /** Returns the full event log for an aggregate (for debugging / audit). */
    fun getEventLog(aggregateId: String): List<DomainEvent> =
        events[aggregateId]?.toList() ?: emptyList()

    /** Returns all events across all aggregates in global sequence order. */
    fun getAllEvents(): List<DomainEvent> =
        events.values.flatten().sortedBy { it.sequenceNumber }

    fun getSnapshot(aggregateId: String): InventorySnapshot? = snapshots[aggregateId]

    fun aggregateCount(): Int = events.size

    fun totalEventCount(): Int = events.values.sumOf { it.size }
}
