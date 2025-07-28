package http

import (
	di "HubInvestments/pck"
	"encoding/json"
	"net/http"
)

func DoLogin(w http.ResponseWriter, r *http.Request, container di.Container) {
	w.Header().Set("Content-Type", "application/json")

	var loginRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := json.NewDecoder(r.Body).Decode(&loginRequest)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Authenticate user
	user, err := container.DoLoginUsecase().Execute(loginRequest.Email, loginRequest.Password)

	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Generate token
	tokenString, err := container.GetAuthService().CreateToken(user.Email.Value(), user.ID)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// Return success response with proper JSON format
	w.WriteHeader(http.StatusOK)
	response := map[string]string{"token": tokenString}
	json.NewEncoder(w).Encode(response)
}
