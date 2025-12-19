package com.example.demo;

// User service implementation
public class UserService {
    
    // Database connection pool
    private DataSource dataSource;
    
    /**
     * Find user by ID
     * @param id user identifier
     * @return user object or null
     */
    public User findById(Long id) {
        // Execute query
        return database.query("SELECT * FROM users WHERE id = ?", id);
    }
    
    /* 
     * Helper method for validation 
     */
    private boolean validate(User user) {
        return user != null && user.getName() != null;
    }
}
