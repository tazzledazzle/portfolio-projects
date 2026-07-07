package com.example.eda.consumer;

import com.example.eda.bus.EventBus;
import com.example.eda.event.OrderPlacedEvent;

import java.util.Collections;
import java.util.Set;
import java.util.UUID;
import java.util.concurrent.ConcurrentHashMap;

// Idempotent consumer — safe to receive the same event more than once (at-least-once delivery).
// Deduplication uses the event's stable eventId as the key.
public class NotificationConsumer {

    // In production this would be a persistent store (Redis, DB) shared across instances.
    private final Set<UUID> processedEvents = Collections.newSetFromMap(new ConcurrentHashMap<>());

    public NotificationConsumer(EventBus bus) {
        bus.subscribe(OrderPlacedEvent.class, this::handle);
    }

    private void handle(OrderPlacedEvent event) {
        // Idempotency check — duplicate delivery is silently skipped
        if (!processedEvents.add(event.eventId())) {
            System.out.println("Duplicate event ignored: " + event.eventId());
            return;
        }

        System.out.printf("[NOTIFY] Order %s placed by customer %s — sending confirmation email%n",
            event.orderId(), event.customerId());
    }
}
