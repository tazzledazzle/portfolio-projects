package com.example.eda.command;

import java.math.BigDecimal;
import java.util.UUID;

// Command — imperative intent ("do this"); point-to-point, exactly one handler.
// Distinct from an Event ("this happened"); events fan out to N consumers.
public record PlaceOrderCommand(
    UUID commandId,
    UUID customerId,
    BigDecimal amount
) {
    public static PlaceOrderCommand of(UUID customerId, BigDecimal amount) {
        return new PlaceOrderCommand(UUID.randomUUID(), customerId, amount);
    }
}
