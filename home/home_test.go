package home

import (
	"HubInvestments/home/application/service"
	domain "HubInvestments/home/domain/model"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

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

type MockContainer struct {
	aucService *service.AucService
}

func (m *MockContainer) GetAucService() *service.AucService {
	return m.aucService
}

type MockAucRepository struct {
	aggregations []domain.AssetsModel
	err          error
}

func (m *MockAucRepository) GetPositionAggregation(userId string) ([]domain.AssetsModel, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.aggregations, nil
}

func TestGetAucAggregation_Success(t *testing.T) {
	// Mock dependencies
	verifyToken := func(token string, w http.ResponseWriter) (string, error) {
		return "user123", nil
	}

	mockRepo := &MockAucRepository{
		aggregations: []domain.AssetsModel{
			{Symbol: "AAPL", Category: 1, AveragePrice: 150, LastPrice: 155, Quantity: 10},
			{Symbol: "AMZN", Category: 1, AveragePrice: 350, LastPrice: 385, Quantity: 5},
			{Symbol: "VOO", Category: 2, AveragePrice: 450, LastPrice: 555, Quantity: 15},
		},
	}

	mockContainer := &MockContainer{
		aucService: service.NewAucService(mockRepo),
	}

	req, err := http.NewRequest("GET", "/auc-aggregation", nil)
	assert.NoError(t, err)
	req.Header.Set("Authorization", "Bearer token")

	rr := httptest.NewRecorder()
	GetAucAggregation(rr, req, verifyToken, mockContainer)

	// Check the response
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	var response domain.AucAggregationModel
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)

	//Stocks
	assert.Equal(t, float32(3475), response.PositionAggregation[0].CurrentTotal)
	assert.Equal(t, float32(3250), response.PositionAggregation[0].TotalInvested)
	assert.Equal(t, float32(225), response.PositionAggregation[0].Pnl)
	assert.Equal(t, float32(10), response.PositionAggregation[0].PnlPercentage)
	assert.Equal(t, int(2), len(response.PositionAggregation[0].Assets))

	//ETFs
	assert.Equal(t, float32(8325), response.PositionAggregation[1].CurrentTotal)
	assert.Equal(t, float32(6750), response.PositionAggregation[1].TotalInvested)
	assert.Equal(t, float32(1575), response.PositionAggregation[1].Pnl)
	assert.Equal(t, float32(23.333334), response.PositionAggregation[1].PnlPercentage)
	assert.Equal(t, int(1), len(response.PositionAggregation[1].Assets))
}
