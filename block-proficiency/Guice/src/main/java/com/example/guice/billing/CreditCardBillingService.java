package com.example.guice.billing;

import com.example.guice.audit.AuditLogger;
import com.google.inject.Inject;
import java.math.BigDecimal;

// Constructor injection — preferred style; dependencies are explicit and testable
public class CreditCardBillingService implements BillingService {

    private final AuditLogger auditLogger;

    @Inject
    public CreditCardBillingService(AuditLogger auditLogger) {
        this.auditLogger = auditLogger;
    }

    @Override
    public Receipt charge(String cardToken, BigDecimal amount) {
        auditLogger.log("Charging card token=" + cardToken + " amount=" + amount);
        return new Receipt(cardToken, amount, "SUCCESS");
    }
}
