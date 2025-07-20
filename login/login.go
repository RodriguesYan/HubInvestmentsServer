package login

import (
	"HubInvestments/auth"
	"HubInvestments/auth/token"
	"HubInvestments/shared/infra/database"
	"encoding/json"
	"fmt"
	"net/http"
)

// LoginModel represents the login request payload
type LoginModel struct {
	Email    string
	Password string
}

// UserCredentials represents user data from the database
type UserCredentials struct {
	ID       string
	Email    string
	Password string
}

// LoginHandler handles user authentication using the database abstraction
type LoginHandler struct {
	db          database.Database
	authService auth.IAuthService
}

// NewLoginHandler creates a new login handler with database abstraction
func NewLoginHandler(db database.Database) *LoginHandler {
	tokenService := token.NewTokenService()
	authService := auth.NewAuthService(tokenService)

	return &LoginHandler{
		db:          db,
		authService: authService,
	}
}

// Login handles the login HTTP request
func (h *LoginHandler) Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Parse request body
	loginRequest, err := h.parseLoginRequest(r)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Authenticate user
	user, err := h.authenticateUser(loginRequest.Email, loginRequest.Password)
	if err != nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// Generate token
	tokenString, err := h.authService.CreateToken(user.Email, user.ID)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	// Return success response
	h.writeSuccessResponse(w, tokenString)
}

// parseLoginRequest parses the login request from the HTTP body
func (h *LoginHandler) parseLoginRequest(r *http.Request) (*LoginModel, error) {
	decoder := json.NewDecoder(r.Body)
	var loginRequest LoginModel

	if err := decoder.Decode(&loginRequest); err != nil {
		return nil, fmt.Errorf("failed to decode request body: %w", err)
	}

	return &loginRequest, nil
}

// authenticateUser validates user credentials against the database
func (h *LoginHandler) authenticateUser(email, password string) (*UserCredentials, error) {
	query := "SELECT id, email, password FROM users WHERE email = $1"

	var user UserCredentials
	err := h.db.Get(&user, query, email)
	if err != nil {
		return nil, fmt.Errorf("user not found or database error: %w", err)
	}

	// Validate password (in a real application, passwords should be hashed)
	if user.Password != password {
		return nil, fmt.Errorf("invalid password")
	}

	return &user, nil
}

// writeErrorResponse writes an error response to the HTTP writer
func (h *LoginHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.WriteHeader(statusCode)
	response := map[string]string{"error": message}
	json.NewEncoder(w).Encode(response)
}

// writeSuccessResponse writes a success response with the token
func (h *LoginHandler) writeSuccessResponse(w http.ResponseWriter, token string) {
	w.WriteHeader(http.StatusOK)
	response := map[string]string{"token": token}
	json.NewEncoder(w).Encode(response)
}

// Login is the main login function that creates a handler and processes the login
// This is the main entry point for login functionality using the database abstraction
func Login(w http.ResponseWriter, r *http.Request) {
	db, err := database.CreateDatabaseConnection()
	if err != nil {
		http.Error(w, "Database connection failed", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	handler := NewLoginHandler(db)
	handler.Login(w, r)
}

// LoginWithDatabase is an alias for Login for backward compatibility
func LoginWithDatabase(w http.ResponseWriter, r *http.Request) {
	Login(w, r)
}
