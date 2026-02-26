package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// contextKey is a custom type for context keys to avoid collisions.
type contextKey string

const (
	// UserIDKey is the context key for storing the authenticated user's ID.
	UserIDKey contextKey = "user_id"
	// UserUUIDKey is the context key for storing the authenticated user's UUID from Supabase.
	UserUUIDKey contextKey = "user_uuid"
)

// SupabaseClaims represents the claims in a Supabase JWT token.
type SupabaseClaims struct {
	Sub   string `json:"sub"`   // User UUID
	Email string `json:"email"` // User email
	Role  string `json:"role"`  // User role (authenticated, anon, etc.)
	jwt.RegisteredClaims
}

// AuthMiddleware validates Supabase JWT tokens and extracts user information.
// It adds the user's UUID to the request context for downstream handlers.
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get JWT secret from environment
		jwtSecret := os.Getenv("SUPABASE_JWT_SECRET")
		if jwtSecret == "" {
			log.Println("[Auth] SUPABASE_JWT_SECRET not set, skipping authentication")
			// In development, allow requests without auth
			next.ServeHTTP(w, r)
			return
		}

		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			log.Println("[Auth] No Authorization header")
			http.Error(w, "Unauthorized: missing token", http.StatusUnauthorized)
			return
		}

		// Check for "Bearer " prefix
		const bearerPrefix = "Bearer "
		if !strings.HasPrefix(authHeader, bearerPrefix) {
			log.Println("[Auth] Invalid Authorization header format")
			http.Error(w, "Unauthorized: invalid token format", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, bearerPrefix)

		// Parse and validate the JWT
		token, err := jwt.ParseWithClaims(tokenString, &SupabaseClaims{}, func(token *jwt.Token) (interface{}, error) {
			// Verify signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(jwtSecret), nil
		})

		if err != nil {
			log.Printf("[Auth] Failed to parse JWT: %v", err)
			http.Error(w, "Unauthorized: invalid token", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(*SupabaseClaims)
		if !ok || !token.Valid {
			log.Println("[Auth] Invalid token claims")
			http.Error(w, "Unauthorized: invalid token", http.StatusUnauthorized)
			return
		}

		// Verify the user has authenticated role
		if claims.Role != "authenticated" {
			log.Printf("[Auth] User does not have authenticated role: %s", claims.Role)
			http.Error(w, "Unauthorized: insufficient permissions", http.StatusForbidden)
			return
		}

		// Store user UUID in context
		ctx := context.WithValue(r.Context(), UserUUIDKey, claims.Sub)
		ctx = context.WithValue(ctx, contextKey("user_email"), claims.Email)

		log.Printf("[Auth] Authenticated user: %s (%s)", claims.Email, claims.Sub)

		// Continue with the updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserUUID extracts the user UUID from the request context.
// Returns empty string if not found.
func GetUserUUID(ctx context.Context) string {
	if uuid, ok := ctx.Value(UserUUIDKey).(string); ok {
		return uuid
	}
	return ""
}

// GetUserEmail extracts the user email from the request context.
// Returns empty string if not found.
func GetUserEmail(ctx context.Context) string {
	if email, ok := ctx.Value(contextKey("user_email")).(string); ok {
		return email
	}
	return ""
}

// GetUserID extracts the integer user ID from the request context.
// Returns 0 if not found.
func GetUserID(ctx context.Context) int {
	if userID, ok := ctx.Value(UserIDKey).(int); ok {
		return userID
	}
	return 0
}

// SetUserID stores the integer user ID in the context.
func SetUserID(ctx context.Context, userID int) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

// RequireAuth is a middleware that requires authentication.
// It validates the JWT, looks up the user, and stores user_id in context.
// This combines JWT validation with database user lookup.
func RequireAuth(getUserByAuthID func(ctx context.Context, authUserID, email string) (int, error)) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// In development mode (secret not set), bypass authentication
			if os.Getenv("SUPABASE_JWT_SECRET") == "" {
				next.ServeHTTP(w, r)
				return
			}

			// First, validate JWT and extract UUID
			authHandler := AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// JWT is valid, now lookup user in database
				userUUID := GetUserUUID(r.Context())
				userEmail := GetUserEmail(r.Context())

				if userUUID == "" {
					log.Println("[Auth] No user UUID in context after JWT validation")
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}

				// Check if userUUID is actually an integer ID (backend-managed auth)
				var userID int
				var err error
				if id, parseErr := strconv.Atoi(userUUID); parseErr == nil {
					// It's an integer ID, use it directly
					userID = id
				} else {
					// It's likely a UUID (Supabase), get or create user in database
					userID, err = getUserByAuthID(r.Context(), userUUID, userEmail)
					if err != nil {
						log.Printf("[Auth] Failed to get/create user: %v", err)
						http.Error(w, "Internal server error", http.StatusInternalServerError)
						return
					}
				}

				// Store integer user_id in context
				ctx := SetUserID(r.Context(), userID)
				log.Printf("[Auth] User authenticated: id=%d uuid=%s email=%s", userID, userUUID, userEmail)

				// Continue with authenticated context
				next.ServeHTTP(w, r.WithContext(ctx))
			}))

			authHandler.ServeHTTP(w, r)
		})
	}
}
