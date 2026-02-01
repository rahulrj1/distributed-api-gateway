package com.gateway.serviceb.controller;

import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import jakarta.servlet.http.HttpServletRequest;
import java.util.HashMap;
import java.util.Map;
import java.util.Enumeration;

/**
 * API Controller for Service B
 * 
 * Endpoints:
 * - GET /health - Health check
 * - GET /hello - Hello message
 * - POST /echo - Echo request details
 */
@RestController
public class ApiController {

    @GetMapping("/health")
    public ResponseEntity<Map<String, String>> health() {
        Map<String, String> response = new HashMap<>();
        response.put("status", "healthy");
        response.put("service", "service-b");
        return ResponseEntity.ok(response);
    }

    @GetMapping("/hello")
    public ResponseEntity<Map<String, String>> hello(HttpServletRequest request) {
        Map<String, String> response = new HashMap<>();
        response.put("message", "Hello from Java Service B");
        response.put("request_id", getHeader(request, "X-Request-ID", "unknown"));
        response.put("user_id", getHeader(request, "X-User-ID", "unknown"));
        return ResponseEntity.ok(response);
    }

    @RequestMapping(value = "/echo", method = {RequestMethod.GET, RequestMethod.POST, 
                                                RequestMethod.PUT, RequestMethod.DELETE, 
                                                RequestMethod.PATCH})
    public ResponseEntity<Map<String, Object>> echo(
            HttpServletRequest request,
            @RequestBody(required = false) Map<String, Object> body) {
        
        Map<String, Object> response = new HashMap<>();
        response.put("service", "service-b");
        response.put("method", request.getMethod());
        response.put("path", request.getRequestURI());
        response.put("query_params", getQueryParams(request));
        response.put("headers", getHeaders(request));
        response.put("body", body);
        
        return ResponseEntity.ok(response);
    }

    private String getHeader(HttpServletRequest request, String name, String defaultValue) {
        String value = request.getHeader(name);
        return value != null ? value : defaultValue;
    }

    private Map<String, String> getQueryParams(HttpServletRequest request) {
        Map<String, String> params = new HashMap<>();
        request.getParameterMap().forEach((key, values) -> {
            if (values.length > 0) {
                params.put(key, values[0]);
            }
        });
        return params;
    }

    private Map<String, String> getHeaders(HttpServletRequest request) {
        Map<String, String> headers = new HashMap<>();
        Enumeration<String> headerNames = request.getHeaderNames();
        while (headerNames.hasMoreElements()) {
            String name = headerNames.nextElement();
            if (!name.equalsIgnoreCase("host") && !name.equalsIgnoreCase("content-length")) {
                headers.put(name, request.getHeader(name));
            }
        }
        return headers;
    }
}
