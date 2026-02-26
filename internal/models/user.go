package models

import "time"

// User represents a user in the system.
type User struct {
	ID           int       `json:"id"`
	Email        string    `json:"email"`
	Name         string    `json:"name"`
	PasswordHash string    `json:"-"` // Never expose password hash in JSON
	CreatedAt    time.Time `json:"createdAt"`
}

// RegisterRequest is the payload for user registration.
type RegisterRequest struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

// LoginRequest is the payload for user login.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse is the response sent to the client after successful login/registration.
type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}
