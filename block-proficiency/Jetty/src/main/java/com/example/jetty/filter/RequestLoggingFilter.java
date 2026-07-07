package com.example.jetty.filter;

import jakarta.servlet.*;
import jakarta.servlet.http.*;
import java.io.IOException;
import java.time.Instant;

// Filter — intercepts every request before it reaches a servlet;
// chain.doFilter() passes control to the next filter or the target servlet.
public class RequestLoggingFilter implements Filter {

    @Override
    public void doFilter(ServletRequest request, ServletResponse response, FilterChain chain)
            throws IOException, ServletException {

        HttpServletRequest  req = (HttpServletRequest)  request;
        HttpServletResponse res = (HttpServletResponse) response;

        long start = System.currentTimeMillis();

        chain.doFilter(request, response);   // delegate to next filter / servlet

        long elapsed = System.currentTimeMillis() - start;
        System.out.printf("[%s] %s %s → %d (%dms)%n",
            Instant.now(), req.getMethod(), req.getRequestURI(), res.getStatus(), elapsed);
    }
}
