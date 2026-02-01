package com.gateway.serviceb.controller;

import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.autoconfigure.web.servlet.WebMvcTest;
import org.springframework.http.MediaType;
import org.springframework.test.web.servlet.MockMvc;

import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.*;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.*;

/**
 * Unit tests for Service B
 * 
 * Run locally (from services/service-b/):
 *     mvn test
 * 
 * Run via Docker (from project root):
 *     docker build -t service-b ./services/service-b
 *     docker run --rm service-b mvn test
 */
@WebMvcTest(ApiController.class)
class ApiControllerTest {

    @Autowired
    private MockMvc mockMvc;

    // Health endpoint tests
    
    @Test
    void health_returns200() throws Exception {
        mockMvc.perform(get("/health"))
                .andExpect(status().isOk());
    }

    @Test
    void health_returnsCorrectJson() throws Exception {
        mockMvc.perform(get("/health"))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.status").value("healthy"))
                .andExpect(jsonPath("$.service").value("service-b"));
    }

    // Hello endpoint tests
    
    @Test
    void hello_returns200() throws Exception {
        mockMvc.perform(get("/hello"))
                .andExpect(status().isOk());
    }

    @Test
    void hello_returnsMessage() throws Exception {
        mockMvc.perform(get("/hello"))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.message").value("Hello from Java Service B"));
    }

    // Echo endpoint tests
    
    @Test
    void echo_postReturns200() throws Exception {
        mockMvc.perform(post("/echo")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content("{\"test\": \"data\"}"))
                .andExpect(status().isOk());
    }

    @Test
    void echo_returnsMethod() throws Exception {
        mockMvc.perform(post("/echo")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content("{\"test\": \"data\"}"))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.method").value("POST"));
    }

    @Test
    void echo_returnsBody() throws Exception {
        mockMvc.perform(post("/echo")
                        .contentType(MediaType.APPLICATION_JSON)
                        .content("{\"key\": \"value\"}"))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.body.key").value("value"));
    }

    @Test
    void echo_getWorks() throws Exception {
        mockMvc.perform(get("/echo"))
                .andExpect(status().isOk())
                .andExpect(jsonPath("$.method").value("GET"));
    }
}
