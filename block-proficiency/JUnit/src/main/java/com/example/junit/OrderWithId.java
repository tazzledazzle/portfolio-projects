package com.example.junit;

import java.math.BigDecimal;

public final class OrderWithId extends Order {
    private final long id;

    public OrderWithId(long id, String customerId, BigDecimal amount) {
        super(customerId, amount);
        this.id = id;
    }

    public long id() { return id; }
}
