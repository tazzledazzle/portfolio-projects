package com.example.jetty;

import com.example.jetty.filter.RequestLoggingFilter;
import com.example.jetty.servlet.HealthServlet;
import com.example.jetty.servlet.OrderServlet;
import org.eclipse.jetty.server.*;
import org.eclipse.jetty.server.handler.ErrorHandler;
import org.eclipse.jetty.servlet.*;
import org.eclipse.jetty.util.ssl.SslContextFactory;
import org.eclipse.jetty.util.thread.QueuedThreadPool;

public class JettyServer {

    private final Server server;

    public JettyServer(int httpPort, int httpsPort, String keystorePath, String keystorePassword) {
        // Thread pool — bound queue prevents unbounded memory growth under load
        QueuedThreadPool pool = new QueuedThreadPool(200, 10, 60_000);
        pool.setName("jetty");

        server = new Server(pool);

        // ── HTTP connector ────────────────────────────────────────────────
        ServerConnector http = new ServerConnector(server);
        http.setPort(httpPort);
        http.setIdleTimeout(30_000);

        // ── HTTPS connector ───────────────────────────────────────────────
        SslContextFactory.Server ssl = new SslContextFactory.Server();
        ssl.setKeyStorePath(keystorePath);
        ssl.setKeyStorePassword(keystorePassword);
        ssl.setProtocol("TLS");

        HttpConfiguration httpsConfig = new HttpConfiguration();
        httpsConfig.addCustomizer(new SecureRequestCustomizer());

        ServerConnector https = new ServerConnector(server,
            new SslConnectionFactory(ssl, "http/1.1"),
            new HttpConnectionFactory(httpsConfig));
        https.setPort(httpsPort);

        server.addConnector(http);
        server.addConnector(https);

        // ── Servlet context ───────────────────────────────────────────────
        ServletContextHandler ctx = new ServletContextHandler(ServletContextHandler.SESSIONS);
        ctx.setContextPath("/");

        // Servlets
        ctx.addServlet(new ServletHolder(new HealthServlet()), "/health");
        ctx.addServlet(new ServletHolder(new OrderServlet()),  "/api/orders/*");

        // Filter — runs before every servlet in this context
        ctx.addFilter(new FilterHolder(new RequestLoggingFilter()), "/*",
            java.util.EnumSet.of(jakarta.servlet.DispatcherType.REQUEST));

        server.setHandler(ctx);
    }

    public void start() throws Exception { server.start(); }
    public void join()  throws Exception { server.join(); }
    public void stop()  throws Exception { server.stop(); }

    public static void main(String[] args) throws Exception {
        var s = new JettyServer(8080, 8443, "/etc/jetty/keystore.p12", "changeit");
        s.start();
        s.join();
    }
}
