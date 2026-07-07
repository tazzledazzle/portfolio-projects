package com.example.guice.audit;

public class FileAuditLogger implements AuditLogger {

    private final String filePath;

    public FileAuditLogger(String filePath) {
        this.filePath = filePath;
    }

    @Override
    public void log(String event) {
        System.out.printf("[AUDIT -> %s] %s%n", filePath, event);
    }
}
