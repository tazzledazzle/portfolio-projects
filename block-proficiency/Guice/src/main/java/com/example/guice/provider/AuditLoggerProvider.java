package com.example.guice.provider;

import com.example.guice.audit.AuditLogger;
import com.example.guice.audit.FileAuditLogger;
import com.google.inject.Provider;

// toProvider — used when construction requires logic that doesn't fit a simple binding
public class AuditLoggerProvider implements Provider<AuditLogger> {

    @Override
    public AuditLogger get() {
        String path = System.getenv().getOrDefault("AUDIT_LOG_PATH", "/var/log/app/audit.log");
        return new FileAuditLogger(path);
    }
}
