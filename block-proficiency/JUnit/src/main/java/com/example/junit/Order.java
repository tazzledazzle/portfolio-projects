package com.example.junit;

import java.math.BigDecimal;

public record Order(String customerId, BigDecimal amount) {
    public Order withId(long id) {
        return new OrderWithId(id, customerId, amount);
    }
}
