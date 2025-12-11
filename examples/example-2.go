package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Simulate a simple in-memory auth service
type AuthService struct {
	mu            sync.RWMutex
	lastLoginTime time.Time
	isHealthy     bool
	tokenCache    map[string]time.Time // token -> expiry
}

func NewAuthService() *AuthService {
	return &AuthService{
		isHealthy:  true,
		tokenCache: make(map[string]time.Time),
	}
}

// Track when login was last called
func (a *AuthService) RecordLogin() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.lastLoginTime = time.Now()
	a.isHealthy = true
}

// Check if auth service is "healthy" (login called recently)
func (a *AuthService) IsHealthy() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()

	// If login hasn't been called in last 5 seconds, consider auth unhealthy
	if time.Since(a.lastLoginTime) > 5*time.Second {
		return false
	}
	return a.isHealthy
}

// Validate a token (simulated)
func (a *AuthService) ValidateToken(token string) bool {
	// If auth service is unhealthy, token validation fails
	if !a.IsHealthy() {
		return false
	}

	a.mu.RLock()
	defer a.mu.RUnlock()

	if expiry, exists := a.tokenCache[token]; exists {
		return time.Now().Before(expiry)
	}
	return false
}

// Create a new token
func (a *AuthService) CreateToken(token string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.tokenCache[token] = time.Now().Add(10 * time.Minute)
}

var authService = NewAuthService()

func main() {
	// =================================================================
	// ENDPOINT 1: /login - The critical dependency
	// =================================================================
	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		// Simulate auth processing time
		time.Sleep(50 * time.Millisecond)

		// Record that login was called (keeps auth service "healthy")
		authService.RecordLogin()

		// Create a token
		token := fmt.Sprintf("token-%d", time.Now().Unix())
		authService.CreateToken(token)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status": "ok",
			"token":  token,
		})
		fmt.Println("[LOGIN] Auth token created")
	})

	// =================================================================
	// ENDPOINT 2: /signup - Depends on auth service health
	// =================================================================
	http.HandleFunc("/signup", func(w http.ResponseWriter, r *http.Request) {
		// HIDDEN DEPENDENCY: Signup calls auth service to validate
		// if the system can accept new users
		if !authService.IsHealthy() {
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "auth service unavailable, cannot signup",
			})
			fmt.Println("[SIGNUP] FAILED - auth service unhealthy")
			return
		}

		// Simulate signup processing
		time.Sleep(30 * time.Millisecond)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status": "signed up successfully",
		})
		fmt.Println("[SIGNUP] Success")
	})

	// =================================================================
	// ENDPOINT 3: /orders - Requires valid auth token
	// =================================================================
	http.HandleFunc("/orders", func(w http.ResponseWriter, r *http.Request) {
		// HIDDEN DEPENDENCY: Orders requires authentication
		authHeader := r.Header.Get("Authorization")

		if authHeader == "" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "missing auth token",
			})
			fmt.Println("[ORDERS] FAILED - no auth token")
			return
		}

		// Validate token (will fail if auth service is unhealthy)
		if !authService.ValidateToken(authHeader) {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "invalid or expired token, auth service may be down",
			})
			fmt.Println("[ORDERS] FAILED - token validation failed")
			return
		}

		// Simulate order processing
		time.Sleep(20 * time.Millisecond)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]interface{}{
			{"id": 1, "item": "Widget", "price": 29.99},
			{"id": 2, "item": "Gadget", "price": 49.99},
		})
		fmt.Println("[ORDERS] Success")
	})

	// =================================================================
	// ENDPOINT 4: /checkout - Depends on both auth AND orders
	// =================================================================
	http.HandleFunc("/checkout", func(w http.ResponseWriter, r *http.Request) {
		// HIDDEN DEPENDENCY: Checkout requires auth
		authHeader := r.Header.Get("Authorization")

		if authHeader == "" || !authService.ValidateToken(authHeader) {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "authentication required for checkout",
			})
			fmt.Println("[CHECKOUT] FAILED - auth required")
			return
		}

		// HIDDEN DEPENDENCY: If auth service is slow, checkout times out
		if !authService.IsHealthy() {
			w.WriteHeader(http.StatusGatewayTimeout)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "checkout timeout waiting for auth service",
			})
			fmt.Println("[CHECKOUT] FAILED - auth service timeout")
			return
		}

		// Simulate payment processing
		time.Sleep(100 * time.Millisecond)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":         "payment processed",
			"transaction_id": fmt.Sprintf("txn-%d", time.Now().Unix()),
		})
		fmt.Println("[CHECKOUT] Success")
	})

	// =================================================================
	// ENDPOINT 5: /products - Independent (no dependencies)
	// =================================================================
	http.HandleFunc("/products", func(w http.ResponseWriter, r *http.Request) {
		// This endpoint has NO dependencies - should never be affected by chaos
		time.Sleep(10 * time.Millisecond)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]interface{}{
			{"id": 1, "name": "Widget", "price": 29.99},
			{"id": 2, "name": "Gadget", "price": 49.99},
			{"id": 3, "name": "Doohickey", "price": 19.99},
		})
		fmt.Println("[PRODUCTS] Success (independent)")
	})

	// =================================================================
	// ENDPOINT 6: /health - Simple health check
	// =================================================================
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		status := "healthy"
		statusCode := http.StatusOK

		if !authService.IsHealthy() {
			status = "degraded - auth service unhealthy"
			statusCode = http.StatusServiceUnavailable
		}

		w.WriteHeader(statusCode)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status": status,
		})
	})

	fmt.Println("🚀 Demo backend listening on :3000")
	fmt.Println("")
	fmt.Println("Endpoints:")
	fmt.Println("  POST /login      - Creates auth token (critical service)")
	fmt.Println("  POST /signup     - Depends on auth service health")
	fmt.Println("  GET  /orders     - Requires auth token")
	fmt.Println("  POST /checkout   - Requires auth token + auth health")
	fmt.Println("  GET  /products   - Independent (no dependencies)")
	fmt.Println("  GET  /health     - Health check")
	fmt.Println("")
	fmt.Println("Hidden Dependencies:")
	fmt.Println("  ❌ /signup → /login (auth health)")
	fmt.Println("  ❌ /orders → /login (token validation)")
	fmt.Println("  ❌ /checkout → /login (token + health)")
	fmt.Println("  ✅ /products → (independent)")
	fmt.Println("")

	http.ListenAndServe(":3000", nil)
}
