package home

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mocking the auth package
type MockAuth struct {
	mock.Mock
}

func (m *MockAuth) VerifyToken(token string, w http.ResponseWriter) (string, error) {
	args := m.Called(token, w)
	return args.String(0), args.Error(1)
}

func TestGetAucAggregation(t *testing.T) {
	// Mocking the auth package
	mockAuth := new(MockAuth)
	mockAuth.On("VerifyToken", "valid-token", mock.Anything).Return("user-id", nil)

	// Mocking the database connection
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	rows := sqlmock.NewRows([]string{"symbol", "average_price", "quantity", "category", "last_price", "current_value"}).
		AddRow("AAPL", 150.0, 10, 1, 155.0, 1000.0)
	mock.ExpectQuery("SELECT i.symbol, p.average_price, p.quantity, i.category, i.last_price, b.current_value").
		WithArgs("user-id").
		WillReturnRows(rows)

	// Creating a new HTTP request
	req, err := http.NewRequest("GET", "/getAucAggregationBalance", nil)
	assert.NoError(t, err)
	req.Header.Set("Authorization", "valid-token")

	// Creating a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Calling the handler function
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		GetAucAggregation(w, r, mockAuth.VerifyToken, func() (*sqlx.DB, error) { return sqlxDB, nil })
	})
	handler.ServeHTTP(rr, req)

	// Checking the status code
	assert.Equal(t, http.StatusOK, rr.Code)

	// Checking the response body
	var response AucAggregationModel
	err = json.NewDecoder(rr.Body).Decode(&response)
	assert.NoError(t, err)

	// Add more assertions based on the expected response
	assert.Equal(t, float32(1000.0), response.TotalBalance) // Example assertion
}

func TestGetAucAggregation_Unauthorized(t *testing.T) {
	// Mocking the auth package
	mockAuth := new(MockAuth)
	mockAuth.On("VerifyToken", "invalid-token", mock.Anything).Return("", fmt.Errorf("unauthorized"))

	// Creating a new HTTP request
	req, err := http.NewRequest("GET", "/getAucAggregationBalance", nil)
	assert.NoError(t, err)
	req.Header.Set("Authorization", "invalid-token")

	// Creating a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Calling the handler function
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		GetAucAggregation(w, r, mockAuth.VerifyToken, func() (*sqlx.DB, error) { return nil, nil })
	})
	handler.ServeHTTP(rr, req)

	// Checking the status code
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}
