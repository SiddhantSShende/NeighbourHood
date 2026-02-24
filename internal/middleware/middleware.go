package middleware

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// contextKey is an unexported type for context keys in this package.
// Using a dedicated type prevents key collisions with other packages.
type contextKey string

const (
	// ContextKeyUserID is the context key used to store the authenticated user's ID.
	ContextKeyUserID contextKey = "user_id"
)

// Logger middleware logs HTTP requests with method, path, status, and duration.
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)
		log.Printf("[%s] %s %s - %d (%v)", r.Method, r.URL.Path, r.RemoteAddr, wrapped.statusCode, duration)
	})
}

// responseWriter wraps http.ResponseWriter to capture the written status code.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Auth middleware validates Bearer JWT tokens.
// In development it accepts any non-empty token to ease local testing.
// Set ENV=production to enforce strict validation.
func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "missing authorization header", http.StatusUnauthorized)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			http.Error(w, "invalid authorization header format, expected: Bearer <token>", http.StatusUnauthorized)
			return
		}

		token := strings.TrimSpace(parts[1])
		if token == "" {
			http.Error(w, "empty bearer token", http.StatusUnauthorized)
			return
		}

		// TODO: replace with real JWT validation using golang-jwt/jwt/v5 and the
		// configured JWT secret once auth.generateJWT generates real tokens.
		// Example:
		//   claims, err := jwt.ParseWithClaims(token, &jwt.RegisteredClaims{}, keyFunc)
		//   userID := claims.Subject

		// Placeholder: store a resolved user ID in context.
		ctx := context.WithValue(r.Context(), ContextKeyUserID, token)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// CORS middleware adds Cross-Origin Resource Sharing headers.
// The allowed origin is read from the CORS_ALLOW_ORIGIN environment variable
// (defaults to "*" for development; set a specific origin in production).
func CORS(next http.Handler) http.Handler {
	allowedOrigin := os.Getenv("CORS_ALLOW_ORIGIN")
	if allowedOrigin == "" {
		allowedOrigin = "*"
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "86400") // 24 h preflight cache

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// SecurityHeaders adds defensive HTTP security headers to every response.
// These protect against MIME-sniffing, click-jacking, and reflected XSS.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
		next.ServeHTTP(w, r)
	})
}

// ipEntry tracks per-IP request counts in a fixed time window.
type ipEntry struct {
	mu        sync.Mutex
	count     int
	windowEnd time.Time
}

var (
	rateMu        sync.Mutex
	rateTable     = map[string]*ipEntry{}
	rateLimitOnce sync.Once
)

// startRateTableCleanup evicts stale per-IP entries every 5 minutes to
// prevent unbounded memory growth on long-running, high-traffic servers.
func startRateTableCleanup() {
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			now := time.Now()
			rateMu.Lock()
			for ip, e := range rateTable {
				e.mu.Lock()
				stale := now.After(e.windowEnd.Add(5 * time.Minute))
				e.mu.Unlock()
				if stale {
					delete(rateTable, ip)
				}
			}
			rateMu.Unlock()
		}
	}()
}

// RateLimiter middleware limits each IP to maxReqPerWindow requests per window.
// Defaults: 100 requests / 60 seconds (overridable at startup via env vars).
func RateLimiter(next http.Handler) http.Handler {
	// Start the background cleanup goroutine exactly once per process.
	rateLimitOnce.Do(startRateTableCleanup)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		const (
			maxRequests    = 100
			windowDuration = 60 * time.Second
		)

		ip := r.RemoteAddr
		// Strip port if present.
		if idx := strings.LastIndex(ip, ":"); idx != -1 {
			ip = ip[:idx]
		}

		rateMu.Lock()
		entry, ok := rateTable[ip]
		if !ok {
			entry = &ipEntry{}
			rateTable[ip] = entry
		}
		rateMu.Unlock()

		entry.mu.Lock()
		now := time.Now()
		if now.After(entry.windowEnd) {
			entry.count = 0
			entry.windowEnd = now.Add(windowDuration)
		}
		entry.count++
		count := entry.count
		entry.mu.Unlock()

		if count > maxRequests {
			w.Header().Set("Retry-After", "60")
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Chain combines multiple middleware, applying them in the order given.
func Chain(handler http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}
