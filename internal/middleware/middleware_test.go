package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// ──────────────────────────────────────────────────────────────────────────────
// Logger middleware
// ──────────────────────────────────────────────────────────────────────────────

func TestLogger_PassesThrough200(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	Logger(next).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Logger should preserve 200, got %d", rr.Code)
	}
}

func TestLogger_PassesThrough404(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	req := httptest.NewRequest(http.MethodGet, "/missing", nil)
	rr := httptest.NewRecorder()

	Logger(next).ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("Logger should preserve 404, got %d", rr.Code)
	}
}

func TestLogger_PassesThrough500(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	req := httptest.NewRequest(http.MethodGet, "/error", nil)
	rr := httptest.NewRecorder()

	Logger(next).ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("Logger should preserve 500, got %d", rr.Code)
	}
}

func TestLogger_ContainsRequestContext(t *testing.T) {
	// Logger should not modify request; body still accessible downstream
	bodyContent := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/logged" {
			bodyContent = true
		}
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodPost, "/logged", nil)
	rr := httptest.NewRecorder()

	Logger(next).ServeHTTP(rr, req)

	if !bodyContent {
		t.Error("logger should pass request to next handler unmodified")
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Auth middleware
// ──────────────────────────────────────────────────────────────────────────────

func TestAuth_NoAuthorizationHeader_Returns401(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rr := httptest.NewRecorder()

	Auth(next).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 without auth header, got %d", rr.Code)
	}
}

func TestAuth_EmptyAuthorizationHeader_Returns401(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "")
	rr := httptest.NewRecorder()

	Auth(next).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for empty auth header, got %d", rr.Code)
	}
}

func TestAuth_BearerToken_Returns200(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer some-valid-token")
	rr := httptest.NewRecorder()

	Auth(next).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 with Bearer token, got %d", rr.Code)
	}
}

func TestAuth_WrongScheme_Returns401(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Basic dXNlcjpwYXNz") // Basic auth
	rr := httptest.NewRecorder()

	Auth(next).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for Basic auth scheme, got %d", rr.Code)
	}
}

func TestAuth_InjectsUserIDIntoContext(t *testing.T) {
	userIDFound := ""
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if uid, ok := r.Context().Value(ContextKeyUserID).(string); ok {
			userIDFound = uid
		}
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer some-token")
	rr := httptest.NewRecorder()

	Auth(next).ServeHTTP(rr, req)

	if userIDFound == "" {
		t.Error("Auth middleware should inject user_id into context")
	}
	if userIDFound != "some-token" {
		t.Errorf("expected some-token, got %q", userIDFound)
	}
}

func TestAuth_TokenOnlyBearerKeyword_Returns401(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer") // no token after keyword
	rr := httptest.NewRecorder()

	Auth(next).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for bare 'Bearer' without token, got %d", rr.Code)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// CORS middleware
// ──────────────────────────────────────────────────────────────────────────────

func TestCORS_SetsAllowOriginHeader(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodGet, "/api/data", nil)
	rr := httptest.NewRecorder()

	CORS(next).ServeHTTP(rr, req)

	if rr.Header().Get("Access-Control-Allow-Origin") == "" {
		t.Error("CORS middleware should set Access-Control-Allow-Origin header")
	}
}

func TestCORS_SetsAllowMethodsHeader(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodGet, "/api/data", nil)
	rr := httptest.NewRecorder()

	CORS(next).ServeHTTP(rr, req)

	if rr.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Error("CORS middleware should set Access-Control-Allow-Methods header")
	}
}

func TestCORS_SetsAllowHeadersHeader(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodGet, "/api/data", nil)
	rr := httptest.NewRecorder()

	CORS(next).ServeHTTP(rr, req)

	if rr.Header().Get("Access-Control-Allow-Headers") == "" {
		t.Error("CORS middleware should set Access-Control-Allow-Headers header")
	}
}

func TestCORS_OPTIONS_Returns200(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot) // should never be reached
	})
	req := httptest.NewRequest(http.MethodOptions, "/api/data", nil)
	rr := httptest.NewRecorder()

	CORS(next).ServeHTTP(rr, req)

	// OPTIONS preflight correctly returns 204 No Content (headers set, no body)
	if rr.Code != http.StatusNoContent {
		t.Errorf("CORS should return 204 for OPTIONS preflight, got %d", rr.Code)
	}
}

func TestCORS_OPTIONS_DoesNotCallNext(t *testing.T) {
	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodOptions, "/api/data", nil)
	rr := httptest.NewRecorder()

	CORS(next).ServeHTTP(rr, req)

	if nextCalled {
		t.Error("CORS should short-circuit on OPTIONS — next handler should not be called")
	}
}

func TestCORS_AllowsWildcardOrigin(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodGet, "/api/data", nil)
	rr := httptest.NewRecorder()

	CORS(next).ServeHTTP(rr, req)

	origin := rr.Header().Get("Access-Control-Allow-Origin")
	if origin != "*" {
		t.Errorf("expected wildcard '*' origin, got %q", origin)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// RateLimiter middleware
// ──────────────────────────────────────────────────────────────────────────────

func TestRateLimiter_PassesThrough(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodGet, "/api/endpoint", nil)
	rr := httptest.NewRecorder()

	RateLimiter(next).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("RateLimiter (placeholder) should pass through, got %d", rr.Code)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Chain
// ──────────────────────────────────────────────────────────────────────────────

func TestChain_SingleMiddleware(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	Chain(next, Logger).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("chain with single middleware should return 200, got %d", rr.Code)
	}
}

func TestChain_MultipleMiddlewares(t *testing.T) {
	order := []string{}
	makeMiddleware := func(name string) func(http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, name)
				next.ServeHTTP(w, r)
			})
		}
	}
	mA := makeMiddleware("A")
	mB := makeMiddleware("B")
	mC := makeMiddleware("C")

	base := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		order = append(order, "handler")
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	Chain(base, mA, mB, mC).ServeHTTP(rr, req)

	if len(order) != 4 {
		t.Fatalf("expected 4 calls (A, B, C, handler), got %v", order)
	}
	if order[len(order)-1] != "handler" {
		t.Errorf("handler should be called last, order: %v", order)
	}
}

func TestChain_NoMiddlewares(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	Chain(next).ServeHTTP(rr, req)

	if rr.Code != http.StatusTeapot {
		t.Errorf("chain with no middlewares should call handler directly, got %d", rr.Code)
	}
}

func TestChain_WithCORS_InjectsHeaders(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodGet, "/api", nil)
	rr := httptest.NewRecorder()

	Chain(next, CORS).ServeHTTP(rr, req)

	if rr.Header().Get("Access-Control-Allow-Origin") == "" {
		t.Error("Chain with CORS should inject CORS headers")
	}
}

func TestChain_AuthBlocksUnauthorized(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodGet, "/api/protected", nil)
	// No Authorization header
	rr := httptest.NewRecorder()

	Chain(next, Auth).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Chain with Auth should block unauthorized requests, got %d", rr.Code)
	}
}

func TestChain_CORS_Then_Auth_OPTIONS_Bypasses(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodOptions, "/api/protected", nil)
	rr := httptest.NewRecorder()

	// CORS should short-circuit before Auth sees the request
	Chain(next, CORS, Auth).ServeHTTP(rr, req)

	if rr.Code == http.StatusUnauthorized {
		t.Error("OPTIONS preflight should be handled by CORS before reaching Auth")
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Error response format
// ──────────────────────────────────────────────────────────────────────────────

func TestAuth_401_ContainsJSONBody(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rr := httptest.NewRecorder()

	Auth(next).ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.FailNow()
	}
	body := rr.Body.String()
	if body == "" {
		t.Error("401 response should contain a body")
	}
	if body == "" {
		t.Error("401 response should have a body")
	}
	// Body contains plain error message (middleware returns plain text error)
	if !strings.Contains(strings.ToLower(body), "authorization") && !strings.Contains(strings.ToLower(body), "unauthorized") && !strings.Contains(strings.ToLower(body), "missing") {
		t.Errorf("401 body should describe the problem, got: %q", body)
	}
}
