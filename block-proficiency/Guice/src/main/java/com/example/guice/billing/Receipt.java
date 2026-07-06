package com.example.guice.billing;

import java.math.BigDecimal;

public record Receipt(String cardToken, BigDecimal amount, String status) {}
