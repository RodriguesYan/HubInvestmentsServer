package di

import (
	balService "HubInvestments/balance/application/service"
	posService "HubInvestments/position/application/service"
)

// TestContainer is a simple mock container for testing
// It implements the Container interface with configurable services
type TestContainer struct {
	aucService     *posService.AucService
	balanceService *balService.BalanceService
}

// NewTestContainer creates a new test container with optional services
func NewTestContainer() *TestContainer {
	return &TestContainer{}
}

// WithAucService sets the AucService for testing
func (c *TestContainer) WithAucService(service *posService.AucService) *TestContainer {
	c.aucService = service
	return c
}

// WithBalanceService sets the BalanceService for testing
func (c *TestContainer) WithBalanceService(service *balService.BalanceService) *TestContainer {
	c.balanceService = service
	return c
}

// GetAucService returns the configured AucService or nil
func (c *TestContainer) GetAucService() *posService.AucService {
	return c.aucService
}

// GetBalanceService returns the configured BalanceService or nil
func (c *TestContainer) GetBalanceService() *balService.BalanceService {
	return c.balanceService
}

// Add new methods here as you add them to the Container interface
// Example:
// func (c *TestContainer) GetNewService() *NewService {
//     return c.newService
// }
