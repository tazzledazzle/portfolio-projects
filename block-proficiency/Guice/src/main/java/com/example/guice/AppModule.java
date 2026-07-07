package com.example.guice;

import com.example.guice.billing.BillingService;
import com.example.guice.billing.CreditCardBillingService;
import com.example.guice.notify.EmailNotifier;
import com.example.guice.notify.Notifier;
import com.example.guice.notify.SmsNotifier;
import com.example.guice.provider.AuditLoggerProvider;
import com.example.guice.audit.AuditLogger;
import com.google.inject.AbstractModule;
import com.google.inject.Singleton;
import com.google.inject.name.Names;

public class AppModule extends AbstractModule {

    private final String smtpHost;
    private final String smsApiKey;

    public AppModule(String smtpHost, String smsApiKey) {
        this.smtpHost = smtpHost;
        this.smsApiKey = smsApiKey;
    }

    @Override
    protected void configure() {
        // bind(Interface).to(Impl) — standard binding
        bind(BillingService.class).to(CreditCardBillingService.class).in(Singleton.class);

        // @Named disambiguation — two Notifier implementations, selected by name
        bind(Notifier.class)
            .annotatedWith(Names.named("email"))
            .to(EmailNotifier.class)
            .in(Singleton.class);

        bind(Notifier.class)
            .annotatedWith(Names.named("sms"))
            .to(SmsNotifier.class)
            .in(Singleton.class);

        // toInstance — bind a pre-built value directly
        bind(String.class).annotatedWith(Names.named("smtpHost")).toInstance(smtpHost);
        bind(String.class).annotatedWith(Names.named("smsApiKey")).toInstance(smsApiKey);

        // toProvider — delegate construction to a Provider when logic is needed
        bind(AuditLogger.class).toProvider(AuditLoggerProvider.class).in(Singleton.class);
    }
}
