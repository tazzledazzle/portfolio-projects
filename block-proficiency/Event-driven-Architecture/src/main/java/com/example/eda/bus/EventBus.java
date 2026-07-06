package com.example.eda.bus;

import com.example.eda.event.DomainEvent;

import java.util.*;
import java.util.concurrent.CopyOnWriteArrayList;

// Simple in-process pub/sub bus — illustrates fan-out to multiple consumers.
// A real system would use Kafka, RabbitMQ, etc.; the contract stays the same.
public class EventBus {

    private final Map<Class<?>, List<EventConsumer<DomainEvent>>> handlers = new HashMap<>();

    @SuppressWarnings("unchecked")
    public <E extends DomainEvent> void subscribe(Class<E> eventType, EventConsumer<E> consumer) {
        handlers
            .computeIfAbsent(eventType, k -> new CopyOnWriteArrayList<>())
            .add((EventConsumer<DomainEvent>) consumer);
    }

    public void publish(DomainEvent event) {
        List<EventConsumer<DomainEvent>> consumers =
            handlers.getOrDefault(event.getClass(), List.of());
        for (EventConsumer<DomainEvent> consumer : consumers) {
            consumer.handle(event);
        }
    }

    @FunctionalInterface
    public interface EventConsumer<E extends DomainEvent> {
        void handle(E event);
    }
}
