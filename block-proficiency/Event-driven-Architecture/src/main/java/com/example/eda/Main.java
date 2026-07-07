package com.example.eda;

import com.example.eda.bus.EventBus;
import com.example.eda.command.PlaceOrderCommand;
import com.example.eda.consumer.NotificationConsumer;
import com.example.eda.event.OrderPlacedEvent;
import com.example.eda.handler.PlaceOrderHandler;

import java.math.BigDecimal;
import java.util.UUID;

public class Main {
    public static void main(String[] args) {
        EventBus bus = new EventBus();

        // Wire consumers (pub/sub — multiple subscribers to the same event)
        new NotificationConsumer(bus);
        bus.subscribe(OrderPlacedEvent.class, e ->
            System.out.println("[AUDIT] Event " + e.eventId() + " at " + e.occurredAt()));

        // Wire command handler (point-to-point)
        var handler = new PlaceOrderHandler(bus);

        // Fire a command → handler mutates state → publishes event → consumers react
        var cmd = PlaceOrderCommand.of(UUID.randomUUID(), new BigDecimal("99.00"));
        UUID orderId = handler.handle(cmd);
        System.out.println("Order created: " + orderId);

        // Simulate duplicate delivery — idempotent consumer must ignore it
        var duplicate = OrderPlacedEvent.of(orderId, cmd.customerId(), "PENDING");
        bus.publish(duplicate);
        bus.publish(duplicate);  // second publish has the same eventId → skipped
    }
}
