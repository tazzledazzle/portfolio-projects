package com.example.eda.query;

import java.util.UUID;

// Query — read-only intent; never mutates state.
// In CQRS the read model is updated by event consumers, not by the command handler.
public record OrderSummaryQuery(UUID customerId) {}
