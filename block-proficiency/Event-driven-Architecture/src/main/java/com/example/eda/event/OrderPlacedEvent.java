package com.example.eda.event;

import java.time.Instant;
import java.util.UUID;

// Stable event schema: every event carries a unique ID, timestamp, and aggregate ID.
// Fields are immutable — consumers must not mutate received events.
public record OrderPlacedEvent(
    UUID eventId,          // deduplication key; consumers use this to detect replays
    Instant occurredAt,    // wall-clock time the domain fact happened
    UUID orderId,          // aggregate ID — which Order this event is about
    UUID customerId,
    String status
) implements DomainEvent {

    // Factory — always generates a fresh eventId so callers can't accidentally reuse one
    public static OrderPlacedEvent of(UUID orderId, UUID customerId, String status) {
        return new OrderPlacedEvent(UUID.randomUUID(), Instant.now(), orderId, customerId, status);
    }
}
