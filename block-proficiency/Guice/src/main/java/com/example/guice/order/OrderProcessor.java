package com.example.guice.order;

import com.example.guice.billing.BillingService;
import com.example.guice.notify.Notifier;
import com.google.inject.Inject;
import com.google.inject.name.Named;
import java.math.BigDecimal;

// Pulls everything together: constructor injection, @Named qualifiers, @Singleton lifecycle
public class OrderProcessor {

    private final BillingService billing;
    private final Notifier emailNotifier;
    private final Notifier smsNotifier;

    @Inject
    public OrderProcessor(
            BillingService billing,
            @Named("email") Notifier emailNotifier,
            @Named("sms")   Notifier smsNotifier) {
        this.billing = billing;
        this.emailNotifier = emailNotifier;
        this.smsNotifier = smsNotifier;
    }

    public void process(String cardToken, BigDecimal amount, String userEmail, String userPhone) {
        var receipt = billing.charge(cardToken, amount);
        emailNotifier.send(userEmail, "Your order was charged: $" + receipt.amount());
        smsNotifier.send(userPhone,  "Order confirmed. Status: " + receipt.status());
    }
}
