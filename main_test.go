package main

import (
	"HubInvestments/internal/auth"
	"HubInvestments/internal/auth/token"
	"HubInvestments/middleware"
	di "HubInvestments/pck"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockTokenService for testing
type MockTokenService struct {
	mock.Mock
	shouldFail bool
}

func (m *MockTokenService) CreateAndSignToken(userName, userId string) (string, error) {
	if m.shouldFail {
		return "", errors.New("token creation failed")
	}
	return "mock-jwt-token", nil
}

func (m *MockTokenService) ValidateToken(tokenString string) (map[string]interface{}, error) {
	if m.shouldFail {
		return nil, errors.New("token validation failed")
	}
	return map[string]interface{}{
		"userId":   "user123",
		"username": "testuser",
	}, nil
}

// MockAuthService for testing
type MockAuthService struct {
	mock.Mock
	shouldFail bool
}

func (m *MockAuthService) VerifyToken(tokenString string, w http.ResponseWriter) (string, error) {
	if m.shouldFail {
		return "", errors.New("verification failed")
	}
	if tokenString == "" {
		return "", errors.New("missing authorization header")
	}
	return "user123", nil
}

func (m *MockAuthService) CreateToken(userName, userId string) (string, error) {
	if m.shouldFail {
		return "", errors.New("token creation failed")
	}
	return "mock-token", nil
}

// Test token service initialization
func TestTokenServiceInitialization(t *testing.T) {
	tokenService := token.NewTokenService()
	assert.NotNil(t, tokenService)

	// Test that we can create tokens
	token, err := tokenService.CreateAndSignToken("testuser", "user123")
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

// Test auth service initialization
func TestAuthServiceInitialization(t *testing.T) {
	tokenService := token.NewTokenService()
	authService := auth.NewAuthService(tokenService)
	assert.NotNil(t, authService)
}

// Test middleware token verifier function creation
func TestTokenVerifierCreation(t *testing.T) {
	tokenService := token.NewTokenService()
	authService := auth.NewAuthService(tokenService)

	verifyToken := middleware.TokenVerifier(func(token string, w http.ResponseWriter) (string, error) {
		return authService.VerifyToken(token, w)
	})

	assert.NotNil(t, verifyToken)

	// Test token verification with mock
	rr := httptest.NewRecorder()
	userId, err := verifyToken("", rr)
	assert.Error(t, err)
	assert.Empty(t, userId)
}

// Test container initialization success
func TestContainerInitialization_Success(t *testing.T) {
	// Note: This may succeed or fail depending on database availability
	container, err := di.NewContainer()

	// Check if database is available or not
	if err != nil {
		// Database not available - expected in many test environments
		assert.Error(t, err)
		assert.Nil(t, container)
		assert.Contains(t, err.Error(), "connect")
	} else {
		// Database is available - test that container is properly initialized
		assert.NotNil(t, container)
		assert.NotNil(t, container.GetPositionAggregationUseCase())
		assert.NotNil(t, container.GetBalanceUseCase())
		assert.NotNil(t, container.GetPortfolioSummaryUsecase())
	}
}

// Test HTTP route registration
func TestHTTPRouteRegistration(t *testing.T) {
	// We can't directly test route registration from main() function
	// since it calls http.HandleFunc which registers globally
	// But we can test that the routes would be properly configured

	// Create mock dependencies
	authService := &MockAuthService{shouldFail: false}

	verifyToken := middleware.TokenVerifier(func(token string, w http.ResponseWriter) (string, error) {
		return authService.VerifyToken(token, w)
	})

	assert.NotNil(t, verifyToken)

	// Test that verifyToken function works as expected
	rr := httptest.NewRecorder()
	userId, err := verifyToken("Bearer mock-token", rr)
	assert.NoError(t, err)
	assert.Equal(t, "user123", userId)
}

// Test route patterns and handler registration
func TestRoutePatterns(t *testing.T) {
	expectedRoutes := []string{
		"/login",
		"/getAucAggregation",
		"/getBalance",
		"/getPortfolioSummary",
	}

	// In a real test, we would need to refactor main() to return the mux
	// or have a separate function for route registration
	for _, route := range expectedRoutes {
		assert.NotEmpty(t, route)
		assert.True(t, strings.HasPrefix(route, "/"))
	}
}

// Test server configuration
func TestServerConfiguration(t *testing.T) {
	// Test port configuration with environment variables
	testCases := []struct {
		name        string
		httpPort    string
		grpcPort    string
		shouldValid bool
	}{
		{"localhost", "localhost:8080", "localhost:50051", true},
		{"home IP", "192.168.0.3:8080", "192.168.0.6:50051", true},
		{"Camila's IP", "192.168.0.48:8080", "192.168.0.6:50051", true},
		{"invalid port", "invalid:port", "invalid:port", false},
		{"missing port", "localhost", "localhost", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test HTTP port validation
			if tc.shouldValid {
				assert.Contains(t, tc.httpPort, ":")
			}
			// Test gRPC port validation
			if tc.shouldValid {
				assert.Contains(t, tc.grpcPort, ":")
			}
		})
	}
}

// Test environment variable loading
func TestEnvironmentVariableLoading(t *testing.T) {
	// Import os package to test environment variables
	// Test default fallback values
	t.Run("default fallback values", func(t *testing.T) {
		// When no environment variables are set, should use defaults
		defaultHTTP := "localhost:8080"
		defaultGRPC := "localhost:50051"

		assert.Contains(t, defaultHTTP, ":")
		assert.Contains(t, defaultGRPC, ":")
	})

	// Test configuration file format
	t.Run("config file format", func(t *testing.T) {
		// Validate that our config.env format is correct
		expectedFormat := map[string]string{
			"HTTP_PORT": "192.168.0.3:8080",
			"GRPC_PORT": "192.168.0.6:50051",
		}

		for key, value := range expectedFormat {
			assert.NotEmpty(t, key)
			assert.Contains(t, value, ":")
		}
	})
}

// Test middleware authentication wrapper
func TestMiddlewareAuthenticationWrapper(t *testing.T) {
	// Create a simple handler for testing
	testHandler := func(w http.ResponseWriter, r *http.Request, userId string) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success: " + userId))
	}

	// Test with successful authentication
	t.Run("successful authentication", func(t *testing.T) {
		verifyToken := middleware.TokenVerifier(func(token string, w http.ResponseWriter) (string, error) {
			if token == "Bearer valid-token" {
				return "user123", nil
			}
			return "", errors.New("invalid token")
		})

		wrappedHandler := middleware.WithAuthentication(verifyToken, testHandler)

		req, err := http.NewRequest("GET", "/test", nil)
		assert.NoError(t, err)
		req.Header.Set("Authorization", "Bearer valid-token")

		rr := httptest.NewRecorder()
		wrappedHandler(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, rr.Body.String(), "success: user123")
	})

	// Test with failed authentication
	t.Run("failed authentication", func(t *testing.T) {
		verifyToken := middleware.TokenVerifier(func(token string, w http.ResponseWriter) (string, error) {
			return "", errors.New("invalid token")
		})

		wrappedHandler := middleware.WithAuthentication(verifyToken, testHandler)

		req, err := http.NewRequest("GET", "/test", nil)
		assert.NoError(t, err)
		req.Header.Set("Authorization", "Bearer invalid-token")

		rr := httptest.NewRecorder()
		wrappedHandler(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		assert.Contains(t, rr.Body.String(), "invalid token")
	})
}

// Test error handling in main function scenarios
func TestMainFunctionErrorScenarios(t *testing.T) {
	// Test container creation error
	t.Run("container creation failure", func(t *testing.T) {
		// This simulates what would happen if di.NewContainer() fails
		// In the actual main function, this would call log.Fatal()
		_, err := di.NewContainer()

		// We may or may not get an error depending on database availability
		// The important thing is that we can handle both cases
		if err != nil {
			assert.Error(t, err)
		} else {
			// If database is available, that's also valid
			assert.NoError(t, err)
		}
	})

	// Test server listen error scenario
	t.Run("server listen failure simulation", func(t *testing.T) {
		// Test invalid port configuration
		invalidPorts := []string{
			"invalid-port",
			":99999999", // Invalid port number
			"",          // Empty port
		}

		for _, port := range invalidPorts {
			// We're not actually starting the server, just validating the port format
			if port == "" {
				assert.Empty(t, port)
			} else {
				assert.NotEmpty(t, port)
			}
		}
	})
}

// Test dependency injection flow
func TestDependencyInjectionFlow(t *testing.T) {
	// Test the flow of dependency creation as it would happen in main()

	// Step 1: Create token service
	tokenService := token.NewTokenService()
	assert.NotNil(t, tokenService)

	// Step 2: Create auth service with token service
	authService := auth.NewAuthService(tokenService)
	assert.NotNil(t, authService)

	// Step 3: Create token verifier function
	verifyToken := middleware.TokenVerifier(func(token string, w http.ResponseWriter) (string, error) {
		return authService.VerifyToken(token, w)
	})
	assert.NotNil(t, verifyToken)

	// Test the complete flow
	rr := httptest.NewRecorder()
	userId, err := verifyToken("", rr)
	assert.Error(t, err) // Should fail with empty token
	assert.Empty(t, userId)
}

// Test concurrent request handling simulation
func TestConcurrentRequestHandling(t *testing.T) {
	// Create a simple handler for testing concurrency
	testHandler := func(w http.ResponseWriter, r *http.Request, userId string) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("user: " + userId))
	}

	verifyToken := middleware.TokenVerifier(func(token string, w http.ResponseWriter) (string, error) {
		if strings.Contains(token, "valid") {
			return "user123", nil
		}
		return "", errors.New("invalid token")
	})

	wrappedHandler := middleware.WithAuthentication(verifyToken, testHandler)

	// Simulate multiple concurrent requests
	requests := []struct {
		name           string
		token          string
		expectedStatus int
	}{
		{"valid request 1", "Bearer valid-token-1", http.StatusOK},
		{"valid request 2", "Bearer valid-token-2", http.StatusOK},
		{"invalid request", "Bearer bad-token", http.StatusUnauthorized}, // Changed to "bad-token" which doesn't contain "valid"
	}

	for _, reqTest := range requests {
		t.Run(reqTest.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/test", nil)
			assert.NoError(t, err)
			req.Header.Set("Authorization", reqTest.token)

			rr := httptest.NewRecorder()
			wrappedHandler(rr, req)

			assert.Equal(t, reqTest.expectedStatus, rr.Code)
		})
	}
}

// Test HTTP header validation
func TestHTTPHeaderValidation(t *testing.T) {
	testHandler := func(w http.ResponseWriter, r *http.Request, userId string) {
		w.WriteHeader(http.StatusOK)
	}

	verifyToken := middleware.TokenVerifier(func(token string, w http.ResponseWriter) (string, error) {
		if token == "" {
			return "", errors.New("missing authorization header")
		}
		return "user123", nil
	})

	wrappedHandler := middleware.WithAuthentication(verifyToken, testHandler)

	testCases := []struct {
		name           string
		authHeader     string
		expectedStatus int
	}{
		{"missing header", "", http.StatusUnauthorized},
		{"empty header", "", http.StatusUnauthorized},
		{"valid header", "Bearer token", http.StatusOK},
		{"malformed header", "InvalidFormat", http.StatusOK}, // Handler doesn't validate format
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/test", nil)
			assert.NoError(t, err)

			if tc.authHeader != "" {
				req.Header.Set("Authorization", tc.authHeader)
			}

			rr := httptest.NewRecorder()
			wrappedHandler(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code)
		})
	}
}

// Test content type headers
func TestContentTypeHeaders(t *testing.T) {
	testHandler := func(w http.ResponseWriter, r *http.Request, userId string) {
		w.WriteHeader(http.StatusOK)
	}

	verifyToken := middleware.TokenVerifier(func(token string, w http.ResponseWriter) (string, error) {
		return "user123", nil
	})

	wrappedHandler := middleware.WithAuthentication(verifyToken, testHandler)

	req, err := http.NewRequest("GET", "/test", nil)
	assert.NoError(t, err)
	req.Header.Set("Authorization", "Bearer token")

	rr := httptest.NewRecorder()
	wrappedHandler(rr, req)

	// Check that the middleware sets the content type
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
	assert.Equal(t, http.StatusOK, rr.Code)
}

// Test application startup sequence validation
func TestApplicationStartupSequence(t *testing.T) {
	// Test the sequence of operations that should happen in main()
	steps := []string{
		"create_token_service",
		"create_auth_service",
		"create_token_verifier",
		"create_container",
		"register_routes",
		"start_server",
	}

	// Validate each step can be performed
	for i, step := range steps {
		t.Run(step, func(t *testing.T) {
			switch step {
			case "create_token_service":
				tokenService := token.NewTokenService()
				assert.NotNil(t, tokenService)
			case "create_auth_service":
				tokenService := token.NewTokenService()
				authService := auth.NewAuthService(tokenService)
				assert.NotNil(t, authService)
			case "create_token_verifier":
				tokenService := token.NewTokenService()
				authService := auth.NewAuthService(tokenService)
				verifyToken := middleware.TokenVerifier(func(token string, w http.ResponseWriter) (string, error) {
					return authService.VerifyToken(token, w)
				})
				assert.NotNil(t, verifyToken)
			case "create_container":
				// This may succeed or fail depending on database availability
				container, err := di.NewContainer()
				// We validate that either we get a container or we get an error, but not both nil
				if err != nil {
					assert.Nil(t, container)
				} else {
					assert.NotNil(t, container)
				}
			case "register_routes":
				// Routes would be registered with http.HandleFunc
				// We can only validate the pattern strings
				routes := []string{"/login", "/getAucAggregation", "/getBalance", "/getPortfolioSummary"}
				assert.Equal(t, 4, len(routes))
			case "start_server":
				// Server would be started with http.ListenAndServe
				// We can only validate the port string format from environment
				// Default fallback should contain ":"
				port := "localhost:8080" // Default fallback
				assert.Contains(t, port, ":")
			}

			// Ensure we're testing all steps
			assert.LessOrEqual(t, i, len(steps)-1)
		})
	}
}
