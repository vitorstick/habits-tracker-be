package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"habit-tracker-be/internal/database"
	"habit-tracker-be/internal/models"

	"habit-tracker-be/internal/middleware"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Register handles user registration.
func Register(w http.ResponseWriter, r *http.Request) {
	log.Println("[Register] Handling POST /api/auth/register")

	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" || req.Name == "" {
		http.Error(w, "Email, name, and password are required", http.StatusBadRequest)
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("[Register] Failed to hash password: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	var user models.User
	err = database.DB.QueryRow(r.Context(),
		`INSERT INTO users (email, name, password_hash)
		 VALUES ($1, $2, $3)
		 RETURNING id, email, name, created_at`,
		req.Email, req.Name, string(hashedPassword)).
		Scan(&user.ID, &user.Email, &user.Name, &user.CreatedAt)

	if err != nil {
		log.Printf("[Register] Failed to insert user: %v", err)
		http.Error(w, "User already exists or database error", http.StatusConflict)
		return
	}

	// Generate token
	token, err := generateToken(user)
	if err != nil {
		log.Printf("[Register] Failed to generate token: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(models.AuthResponse{
		Token: token,
		User:  user,
	})
}

// Login handles user login.
func Login(w http.ResponseWriter, r *http.Request) {
	log.Println("[Login] Handling POST /api/auth/login")

	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var user models.User
	err := database.DB.QueryRow(r.Context(),
		`SELECT id, email, name, password_hash, created_at FROM users WHERE email = $1`,
		req.Email).
		Scan(&user.ID, &user.Email, &user.Name, &user.PasswordHash, &user.CreatedAt)

	if err != nil {
		log.Printf("[Login] User not found: %s", req.Email)
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	// Compare password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		log.Printf("[Login] Invalid password for user: %s", req.Email)
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	// Generate token
	token, err := generateToken(user)
	if err != nil {
		log.Printf("[Login] Failed to generate token: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.AuthResponse{
		Token: token,
		User:  user,
	})
}

// GetCurrentUser returns the currently authenticated user's details.
func GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	log.Println("[GetCurrentUser] Handling GET /api/auth/me")

	userID := middleware.GetUserID(r.Context())
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var user models.User
	err := database.DB.QueryRow(r.Context(),
		`SELECT id, email, name, created_at FROM users WHERE id = $1`,
		userID).
		Scan(&user.ID, &user.Email, &user.Name, &user.CreatedAt)

	if err != nil {
		log.Printf("[GetCurrentUser] User not found: %d", userID)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// generateToken creates a JWT for the user.
// We use the same format as Supabase to minimize changes to the middleware.
func generateToken(user models.User) (string, error) {
	jwtSecret := os.Getenv("SUPABASE_JWT_SECRET")
	if jwtSecret == "" {
		return "", fmt.Errorf("SUPABASE_JWT_SECRET not set")
	}

	// Supabase-like claims
	claims := jwt.MapClaims{
		"sub":   fmt.Sprintf("%d", user.ID), // Store ID as sub
		"email": user.Email,
		"role":  "authenticated",
		"exp":   time.Now().Add(time.Hour * 72).Unix(), // 3 days
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}
