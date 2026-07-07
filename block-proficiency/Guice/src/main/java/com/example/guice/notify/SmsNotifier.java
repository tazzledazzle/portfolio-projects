package com.example.guice.notify;

import com.google.inject.Inject;
import com.google.inject.name.Named;

public class SmsNotifier implements Notifier {

    private String apiKey;

    // Method injection — shows the third injection style; useful when you don't control the constructor
    @Inject
    public void init(@Named("smsApiKey") String apiKey) {
        this.apiKey = apiKey;
    }

    @Override
    public void send(String recipient, String message) {
        System.out.printf("[SMS apiKey=%s] To: %s — %s%n", apiKey, recipient, message);
    }
}
