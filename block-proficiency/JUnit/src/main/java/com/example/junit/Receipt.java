package com.example.junit;

import java.time.Instant;

public record Receipt(long orderId, String transactionId, Instant placedAt) {}
