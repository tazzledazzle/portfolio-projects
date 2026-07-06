package com.example.junit;

import java.math.BigDecimal;
import java.time.Instant;
import java.util.Set;

public class OrderService {

    private static final Set<String> REFUNDABLE_STATUSES = Set.of("PENDING", "CONFIRMED");
    private static final BigDecimal LARGE_ORDER_THRESHOLD_1 = new BigDecimal("500.00");
    private static final BigDecimal LARGE_ORDER_THRESHOLD_2 = new BigDecimal("1000.00");

    private final OrderRepository repository;
    private final PaymentGateway  paymentGateway;

    public OrderService(OrderRepository repository, PaymentGateway paymentGateway) {
        this.repository     = repository;
        this.paymentGateway = paymentGateway;
    }

    public Receipt placeOrder(Order order) {
        if (order.customerId() == null || order.customerId().isBlank()) {
            throw new IllegalArgumentException("customerId must not be blank");
        }
        if (order.amount().compareTo(BigDecimal.ZERO) <= 0) {
            throw new IllegalArgumentException("amount must be positive");
        }

        Order saved = repository.save(order);
        long savedId = ((OrderWithId) saved).id();

        BigDecimal charged = applyDiscount(order.amount());
        String txnId = paymentGateway.charge(order.customerId(), charged);

        return new Receipt(savedId, txnId, Instant.now());
    }

    public boolean isRefundable(String status) {
        return REFUNDABLE_STATUSES.contains(status);
    }

    private BigDecimal applyDiscount(BigDecimal amount) {
        if (amount.compareTo(LARGE_ORDER_THRESHOLD_2) >= 0) {
            return amount.multiply(new BigDecimal("0.90"));  // 10 % off
        }
        if (amount.compareTo(LARGE_ORDER_THRESHOLD_1) >= 0) {
            return amount.multiply(new BigDecimal("0.95"));  // 5 % off
        }
        return amount;
    }
}
