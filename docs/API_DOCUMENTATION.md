# HubInvestments API Documentation

## Overview

HubInvestments is a comprehensive financial investment platform API that provides portfolio management, market data access, and user authentication.

## Accessing the API Documentation

The interactive Swagger documentation is available at:
**http://192.168.0.6:8080/swagger/index.html**

## Available Endpoints

### 1. Authentication
- **POST /login** - User authentication with email and password

### 2. Balance Management
- **GET /getBalance** - Retrieve user's available balance

### 3. Portfolio Management
- **GET /getPortfolioSummary** - Get complete portfolio summary including balance and positions
- **GET /getAucAggregation** - Get aggregated position data by category

## Authentication

Most endpoints require Bearer token authentication. To use protected endpoints:

1. First, authenticate using the `/login` endpoint with valid credentials
2. Include the returned JWT token in the Authorization header:
   ```
   Authorization: Bearer your_jwt_token_here
   ```

## API Response Format

All endpoints return JSON responses with appropriate HTTP status codes:
- **200**: Success
- **401**: Unauthorized (missing or invalid token)
- **500**: Internal server error

## Example Usage

### 1. Login
```bash
curl -X POST http://192.168.0.6:8080/login \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "password": "password123"}'
```

### 2. Get Balance (with authentication)
```bash
curl -X GET http://192.168.0.6:8080/getBalance \
  -H "Authorization: Bearer your_jwt_token_here"
```

### 3. Get Portfolio Summary (with authentication)
```bash
curl -X GET http://192.168.0.6:8080/getPortfolioSummary \
  -H "Authorization: Bearer your_jwt_token_here"
```

## Response Examples

### Login Response
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

### Balance Response
```json
{
  "availableBalance": 15000.50
}
```

### Portfolio Summary Response
```json
{
  "Balance": {
    "availableBalance": 5000.0
  },
  "TotalPortfolio": 17000.0,
  "LastUpdatedDate": "",
  "PositionAggregation": {
    "totalInvested": 11500.0,
    "currentTotal": 12000.0,
    "positionAggregation": [
      {
        "category": 1,
        "totalInvested": 6500.0,
        "currentTotal": 6750.0,
        "pnl": 250.0,
        "pnlPercentage": 3.85,
        "assets": [
          {
            "symbol": "AAPL",
            "quantity": 10.0,
            "averagePrice": 150.0,
            "currentPrice": 155.0,
            "category": 1
          }
        ]
      }
    ]
  }
}
```

## Contact Information

- **Development Team**: HubInvestments Development Team
- **Email**: support@hubinvestments.com

## Documentation Files

- `swagger.json` - OpenAPI 2.0 specification in JSON format
- `swagger.yaml` - OpenAPI 2.0 specification in YAML format
- `docs.go` - Generated Go documentation file 