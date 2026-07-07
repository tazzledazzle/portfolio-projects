package com.example.eda.event;

import java.time.Instant;
import java.util.UUID;

// Marker interface + minimal contract every event must satisfy
public interface DomainEvent {
    UUID eventId();      // deduplication / idempotency key
    Instant occurredAt();
    UUID orderId();      // aggregate ID (could be generified; kept concrete for clarity)
}
