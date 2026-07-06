package com.example.guice.notify;

import com.google.inject.Inject;
import com.google.inject.name.Named;

public class EmailNotifier implements Notifier {

    // Field injection — used here to show the pattern; constructor injection preferred in prod
    @Inject
    @Named("smtpHost")
    private String smtpHost;

    @Override
    public void send(String recipient, String message) {
        System.out.printf("[EMAIL via %s] To: %s — %s%n", smtpHost, recipient, message);
    }
}
