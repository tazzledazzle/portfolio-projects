package com.example.guice.billing;

import java.math.BigDecimal;

public interface BillingService {
    Receipt charge(String cardToken, BigDecimal amount);
}
