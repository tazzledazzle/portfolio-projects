package com.example.junit;

import java.math.BigDecimal;

public interface PaymentGateway {
    String charge(String customerId, BigDecimal amount);
}
