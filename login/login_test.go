package login

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"HubInvestments/shared/infra/database"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockDatabase implements the database.Database interface for testing
type MockDatabase struct {
	mock.Mock
}

func (m *MockDatabase) Query(query string, args ...interface{}) (database.Rows, error) {
	arguments := m.Called(query, args)
	return arguments.Get(0).(database.Rows), arguments.Error(1)
}

func (m *MockDatabase) QueryContext(ctx context.Context, query string, args ...interface{}) (database.Rows, error) {
	arguments := m.Called(ctx, query, args)
	return arguments.Get(0).(database.Rows), arguments.Error(1)
}

func (m *MockDatabase) QueryRow(query string, args ...interface{}) database.Row {
	arguments := m.Called(query, args)
	return arguments.Get(0).(database.Row)
}

func (m *MockDatabase) QueryRowContext(ctx context.Context, query string, args ...interface{}) database.Row {
	arguments := m.Called(ctx, query, args)
	return arguments.Get(0).(database.Row)
}

func (m *MockDatabase) Exec(query string, args ...interface{}) (database.Result, error) {
	arguments := m.Called(query, args)
	return arguments.Get(0).(database.Result), arguments.Error(1)
}

func (m *MockDatabase) ExecContext(ctx context.Context, query string, args ...interface{}) (database.Result, error) {
	arguments := m.Called(ctx, query, args)
	return arguments.Get(0).(database.Result), arguments.Error(1)
}

func (m *MockDatabase) Begin() (database.Transaction, error) {
	arguments := m.Called()
	return arguments.Get(0).(database.Transaction), arguments.Error(1)
}

func (m *MockDatabase) BeginTx(ctx context.Context, opts *sql.TxOptions) (database.Transaction, error) {
	arguments := m.Called(ctx, opts)
	return arguments.Get(0).(database.Transaction), arguments.Error(1)
}

func (m *MockDatabase) Get(dest interface{}, query string, args ...interface{}) error {
	arguments := m.Called(dest, query, args)

	// If we're expecting a successful result, populate the destination
	if arguments.Error(0) == nil && len(arguments) > 1 {
		if user, ok := arguments.Get(1).(*UserCredentials); ok {
			if destUser, ok := dest.(*UserCredentials); ok {
				*destUser = *user
			}
		}
	}

	return arguments.Error(0)
}

func (m *MockDatabase) Select(dest interface{}, query string, args ...interface{}) error {
	arguments := m.Called(dest, query, args)
	return arguments.Error(0)
}

func (m *MockDatabase) Ping() error {
	arguments := m.Called()
	return arguments.Error(0)
}

func (m *MockDatabase) Close() error {
	arguments := m.Called()
	return arguments.Error(0)
}

func TestLoginModel_JSONMarshaling(t *testing.T) {
	// Test that LoginModel can be properly marshaled and unmarshaled
	original := LoginModel{
		Email:    "test@example.com",
		Password: "testpassword",
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(original)
	require.NoError(t, err)

	// Unmarshal back
	var unmarshaled LoginModel
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)

	// Compare
	assert.Equal(t, original.Email, unmarshaled.Email)
	assert.Equal(t, original.Password, unmarshaled.Password)
}

func TestLoginHandler_Success(t *testing.T) {
	// Setup mock database
	mockDB := &MockDatabase{}
	defer mockDB.AssertExpectations(t)

	expectedUser := &UserCredentials{
		ID:       "user-123",
		Email:    "test@example.com",
		Password: "testpassword",
	}

	// Mock successful database query
	mockDB.On("Get",
		mock.AnythingOfType("*login.UserCredentials"),
		"SELECT id, email, password FROM users WHERE email = $1",
		[]interface{}{"test@example.com"},
	).Return(nil, expectedUser)

	// Create login handler with mock database
	handler := NewLoginHandler(mockDB)

	// Create request
	loginData := LoginModel{
		Email:    "test@example.com",
		Password: "testpassword",
	}
	jsonData, _ := json.Marshal(loginData)
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call the handler
	handler.Login(rr, req)

	// Check response
	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "token")
	assert.NotEmpty(t, response["token"])
}

func TestLoginHandler_InvalidCredentials(t *testing.T) {
	// Setup mock database
	mockDB := &MockDatabase{}
	defer mockDB.AssertExpectations(t)

	// Mock database query that returns a user with different password
	expectedUser := &UserCredentials{
		ID:       "user-123",
		Email:    "test@example.com",
		Password: "differentpassword",
	}

	mockDB.On("Get",
		mock.AnythingOfType("*login.UserCredentials"),
		"SELECT id, email, password FROM users WHERE email = $1",
		[]interface{}{"test@example.com"},
	).Return(nil, expectedUser)

	// Create login handler with mock database
	handler := NewLoginHandler(mockDB)

	// Create request with wrong password
	loginData := LoginModel{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}
	jsonData, _ := json.Marshal(loginData)
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call the handler
	handler.Login(rr, req)

	// Check response
	assert.Equal(t, http.StatusUnauthorized, rr.Code)

	var response map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "error")
	assert.Equal(t, "Invalid credentials", response["error"])
}

func TestLoginHandler_UserNotFound(t *testing.T) {
	// Setup mock database
	mockDB := &MockDatabase{}
	defer mockDB.AssertExpectations(t)

	// Mock database query that returns no user found error
	mockDB.On("Get",
		mock.AnythingOfType("*login.UserCredentials"),
		"SELECT id, email, password FROM users WHERE email = $1",
		[]interface{}{"nonexistent@example.com"},
	).Return(errors.New("sql: no rows in result set"))

	// Create login handler with mock database
	handler := NewLoginHandler(mockDB)

	// Create request
	loginData := LoginModel{
		Email:    "nonexistent@example.com",
		Password: "testpassword",
	}
	jsonData, _ := json.Marshal(loginData)
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call the handler
	handler.Login(rr, req)

	// Check response
	assert.Equal(t, http.StatusUnauthorized, rr.Code)

	var response map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "error")
	assert.Equal(t, "Invalid credentials", response["error"])
}

func TestLoginHandler_InvalidJSON(t *testing.T) {
	// Setup mock database (won't be called)
	mockDB := &MockDatabase{}

	// Create login handler with mock database
	handler := NewLoginHandler(mockDB)

	// Create request with invalid JSON
	req, _ := http.NewRequest("POST", "/login", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call the handler
	handler.Login(rr, req)

	// Check response
	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var response map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "error")
	assert.Equal(t, "Invalid request body", response["error"])
}

func TestNewLoginHandler(t *testing.T) {
	mockDB := &MockDatabase{}

	handler := NewLoginHandler(mockDB)

	assert.NotNil(t, handler)
	assert.Equal(t, mockDB, handler.db)
	assert.NotNil(t, handler.authService)
}

func TestUserCredentials(t *testing.T) {
	user := UserCredentials{
		ID:       "123",
		Email:    "test@example.com",
		Password: "password",
	}

	assert.Equal(t, "123", user.ID)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, "password", user.Password)
}
