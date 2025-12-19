package com.example.demo;

// Application main class
public class Application {
    
    // Default port number
    private static final int DEFAULT_PORT = 8080;
    
    /**
     * Main entry point
     * @param args command line arguments
     */
    public static void main(String[] args) {
        System.out.println("Starting application...");
    }
    
    // Configuration holder
    public static class Config {
        /* Server timeout in milliseconds */
        private int timeout = 5000;
        
        /**
         * Get the configured timeout
         * @return timeout value
         */
        public int getTimeout() {
            return timeout;
        }
    }
}
