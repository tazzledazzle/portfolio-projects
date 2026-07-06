package com.example.eda.handler;

import com.example.eda.bus.EventBus;
import com.example.eda.command.PlaceOrderCommand;
import com.example.eda.event.OrderPlacedEvent;

import java.util.UUID;

// Command handler — point-to-point; exactly one handler owns this command.
// It mutates state, then publishes an event so consumers can react without coupling.
public class PlaceOrderHandler {

    private final EventBus bus;

    public PlaceOrderHandler(EventBus bus) {
        this.bus = bus;
    }

    public UUID handle(PlaceOrderCommand cmd) {
        // Domain logic: persist the order (omitted — focus is on EDA wiring)
        UUID orderId = UUID.randomUUID();

        // Publish event — consumers decide what to do next (notifications, audit, read model)
        bus.publish(OrderPlacedEvent.of(orderId, cmd.customerId(), "PENDING"));

        return orderId;
    }
}
