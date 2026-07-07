package com.example.guice;

import com.example.guice.order.OrderProcessor;
import com.google.inject.Guice;
import java.math.BigDecimal;

public class Main {
    public static void main(String[] args) {
        var injector = Guice.createInjector(
            new AppModule("smtp.example.com", "sms-key-abc123")
        );

        var processor = injector.getInstance(OrderProcessor.class);
        processor.process("tok_visa_4242", new BigDecimal("49.99"), "user@example.com", "+15550001234");
    }
}
