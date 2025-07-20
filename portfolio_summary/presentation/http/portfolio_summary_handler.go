package http

import (
	"HubInvestments/middleware"
	di "HubInvestments/pck"
	"encoding/json"
	"fmt"
	"net/http"
)

// GetPortfolioSummary handles portfolio summary retrieval for authenticated users
//
// Endpoint: GET /getPortfolioSummary
// Authentication: Bearer token required in Authorization header
// Content-Type: application/json
//
// Request Headers:
// Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
//
// Success Response (200 OK):
//
//	{
//	  "Balance": {
//	    "availableBalance": 5000.0
//	  },
//	  "TotalPortfolio": 17000.0,
//	  "LastUpdatedDate": "",
//	  "PositionAggregation": {
//	    "totalInvested": 11500.0,
//	    "currentTotal": 12000.0,
//	    "positionAggregation": [
//	      {
//	        "category": 1,
//	        "totalInvested": 6500.0,
//	        "currentTotal": 6750.0,
//	        "pnl": 250.0,
//	        "pnlPercentage": 3.85,
//	        "assets": [
//	          {
//	            "symbol": "AAPL",
//	            "quantity": 10.0,
//	            "averagePrice": 150.0,
//	            "currentPrice": 155.0,
//	            "category": 1
//	          },
//	          {
//	            "symbol": "GOOGL",
//	            "quantity": 2.0,
//	            "averagePrice": 2500.0,
//	            "currentPrice": 2600.0,
//	            "category": 1
//	          }
//	        ]
//	      }
//	    ]
//	  }
//	}
//
// Error Responses:
// 401 Unauthorized - Missing or invalid token:
//
//	{
//	  "error": "Missing authorization header"
//	}
//
// 500 Internal Server Error - Failed to retrieve portfolio data:
// "Failed to get portfolio summary: balance service unavailable"
func GetPortfolioSummary(w http.ResponseWriter, r *http.Request, userId string, container di.Container) {
	aggregation, err := container.GetPortfolioSummaryUsecase().Execute(userId)

	if err != nil {
		http.Error(w, "Failed to get portfolio summary: "+err.Error(), http.StatusInternalServerError)
		return
	}

	result, err := json.Marshal(aggregation)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, string(result))
}

// GetPortfolioSummaryWithAuth returns a handler wrapped with authentication middleware
func GetPortfolioSummaryWithAuth(verifyToken middleware.TokenVerifier, container di.Container) http.HandlerFunc {
	return middleware.WithAuthentication(verifyToken, func(w http.ResponseWriter, r *http.Request, userId string) {
		GetPortfolioSummary(w, r, userId, container)
	})
}
