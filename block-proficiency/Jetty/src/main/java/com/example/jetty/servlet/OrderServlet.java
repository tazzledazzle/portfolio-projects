package com.example.jetty.servlet;

import jakarta.servlet.http.*;
import java.io.IOException;

// Demonstrates reading from HttpServletRequest and writing to HttpServletResponse
public class OrderServlet extends HttpServlet {

    @Override
    protected void doGet(HttpServletRequest req, HttpServletResponse res) throws IOException {
        // pathInfo gives the part after the servlet mapping (/api/orders/*)
        String pathInfo = req.getPathInfo();   // e.g. "/42"
        String orderId  = pathInfo == null ? null : pathInfo.replaceFirst("^/", "");

        if (orderId == null || orderId.isBlank()) {
            res.setStatus(HttpServletResponse.SC_BAD_REQUEST);
            res.getWriter().write("{\"error\":\"orderId required\"}");
            return;
        }

        res.setContentType("application/json");
        res.setStatus(HttpServletResponse.SC_OK);
        res.getWriter().write("{\"orderId\":\"" + orderId + "\",\"status\":\"PENDING\"}");
    }

    @Override
    protected void doPost(HttpServletRequest req, HttpServletResponse res) throws IOException {
        // Read raw body from request input stream
        String body = new String(req.getInputStream().readAllBytes());

        res.setContentType("application/json");
        res.setStatus(HttpServletResponse.SC_CREATED);
        res.getWriter().write("{\"created\":true,\"received\":" + body.length() + "}");
    }
}
