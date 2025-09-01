package websocket

import (
	"context"
	"log"
	"sync"
	"time"
)

type HealthStatus int

const (
	HealthStatusHealthy HealthStatus = iota
	HealthStatusDegraded
	HealthStatusUnhealthy
	HealthStatusCritical
)

func (hs HealthStatus) String() string {
	switch hs {
	case HealthStatusHealthy:
		return "healthy"
	case HealthStatusDegraded:
		return "degraded"
	case HealthStatusUnhealthy:
		return "unhealthy"
	case HealthStatusCritical:
		return "critical"
	default:
		return "unknown"
	}
}

type HealthMonitor struct {
	pool          *ConnectionPool
	config        ConnectionPoolConfig
	status        HealthStatus
	lastCheck     time.Time
	healthHistory []HealthCheckResult
	mutex         sync.RWMutex
	ctx           context.Context
	cancel        context.CancelFunc
	alerts        chan HealthAlert
}

type HealthCheckResult struct {
	Timestamp         time.Time     `json:"timestamp"`
	Status            HealthStatus  `json:"status"`
	ActiveConnections int           `json:"active_connections"`
	FailedConnections int           `json:"failed_connections"`
	ResponseTime      time.Duration `json:"response_time_ms"`
	ErrorRate         float64       `json:"error_rate_percent"`
	Details           string        `json:"details,omitempty"`
}

type HealthAlert struct {
	Timestamp time.Time              `json:"timestamp"`
	Severity  HealthStatus           `json:"severity"`
	Message   string                 `json:"message"`
	Component string                 `json:"component"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// NewHealthMonitor creates a new health monitor
func NewHealthMonitor(pool *ConnectionPool, config ConnectionPoolConfig) *HealthMonitor {
	ctx, cancel := context.WithCancel(context.Background())

	monitor := &HealthMonitor{
		pool:          pool,
		config:        config,
		status:        HealthStatusHealthy,
		lastCheck:     time.Now(),
		healthHistory: make([]HealthCheckResult, 0, 100), // Keep last 100 results
		ctx:           ctx,
		cancel:        cancel,
		alerts:        make(chan HealthAlert, 100),
	}

	go monitor.startHealthChecking()
	go monitor.startAlertProcessing()

	return monitor
}

func (hm *HealthMonitor) CheckHealth() HealthCheckResult {
	startTime := time.Now()

	hm.mutex.Lock()
	defer hm.mutex.Unlock()

	activeConnections := hm.pool.GetConnectionCount()
	poolMetrics := hm.pool.GetMetrics()

	// Calculate error rate
	totalConnections := poolMetrics.TotalConnections
	failedConnections := poolMetrics.FailedConnections
	errorRate := float64(0)
	if totalConnections > 0 {
		errorRate = (float64(failedConnections) / float64(totalConnections)) * 100
	}

	status := hm.determineHealthStatus(activeConnections, errorRate)

	result := HealthCheckResult{
		Timestamp:         startTime,
		Status:            status,
		ActiveConnections: activeConnections,
		FailedConnections: int(failedConnections),
		ResponseTime:      time.Since(startTime),
		ErrorRate:         errorRate,
		Details:           hm.generateHealthDetails(status, activeConnections, errorRate),
	}

	previousStatus := hm.status
	hm.status = status
	hm.lastCheck = startTime
	hm.addToHistory(result)

	// Generate alerts if status changed
	if previousStatus != status {
		hm.generateStatusChangeAlert(previousStatus, status, result)
	}

	return result
}

func (hm *HealthMonitor) determineHealthStatus(activeConnections int, errorRate float64) HealthStatus {
	// Critical: System is failing
	if errorRate > 50 || activeConnections == 0 {
		return HealthStatusCritical
	}

	// Unhealthy: High error rate or resource exhaustion
	if errorRate > 20 || activeConnections > int(float64(hm.config.MaxConnections)*0.95) {
		return HealthStatusUnhealthy
	}

	// Degraded: Moderate issues
	if errorRate > 5 || activeConnections > int(float64(hm.config.MaxConnections)*0.8) {
		return HealthStatusDegraded
	}

	// Healthy: All systems normal
	return HealthStatusHealthy
}

// generateHealthDetails generates detailed health information
func (hm *HealthMonitor) generateHealthDetails(status HealthStatus, activeConnections int, errorRate float64) string {
	switch status {
	case HealthStatusCritical:
		return "Critical: System failure detected. Immediate attention required."
	case HealthStatusUnhealthy:
		return "Unhealthy: High error rate or resource exhaustion detected."
	case HealthStatusDegraded:
		return "Degraded: Performance issues detected. Monitoring required."
	case HealthStatusHealthy:
		return "Healthy: All systems operating normally."
	default:
		return "Unknown status"
	}
}

// addToHistory adds a health check result to history
func (hm *HealthMonitor) addToHistory(result HealthCheckResult) {
	// Keep only last 100 results
	if len(hm.healthHistory) >= 100 {
		hm.healthHistory = hm.healthHistory[1:]
	}
	hm.healthHistory = append(hm.healthHistory, result)
}

func (hm *HealthMonitor) generateStatusChangeAlert(oldStatus, newStatus HealthStatus, result HealthCheckResult) {
	alert := HealthAlert{
		Timestamp: time.Now(),
		Severity:  newStatus,
		Component: "websocket_pool",
		Message:   "Health status changed from " + oldStatus.String() + " to " + newStatus.String(),
		Metadata: map[string]interface{}{
			"active_connections": result.ActiveConnections,
			"error_rate":         result.ErrorRate,
			"response_time_ms":   result.ResponseTime.Milliseconds(),
		},
	}

	select {
	case hm.alerts <- alert:
	default:
		// Alert channel is full, log the alert
		log.Printf("Health alert (channel full): %s - %s", alert.Severity.String(), alert.Message)
	}
}

func (hm *HealthMonitor) startHealthChecking() {
	ticker := time.NewTicker(hm.config.HealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-hm.ctx.Done():
			return
		case <-ticker.C:
			hm.CheckHealth()
		}
	}
}

func (hm *HealthMonitor) startAlertProcessing() {
	for {
		select {
		case <-hm.ctx.Done():
			return
		case alert := <-hm.alerts:
			hm.processAlert(alert)
		}
	}
}

func (hm *HealthMonitor) processAlert(alert HealthAlert) {
	log.Printf("Health Alert [%s] %s: %s", alert.Component, alert.Severity.String(), alert.Message)

	// In a production system, you would:
	// 1. Send alerts to monitoring systems (Prometheus, Grafana, etc.)
	// 2. Send notifications (email, Slack, PagerDuty, etc.)
	// 3. Trigger automated recovery actions
	// 4. Log to centralized logging system

	switch alert.Severity {
	case HealthStatusCritical:
		hm.handleCriticalAlert(alert)
	case HealthStatusUnhealthy:
		hm.handleUnhealthyAlert(alert)
	case HealthStatusDegraded:
		hm.handleDegradedAlert(alert)
	}
}

func (hm *HealthMonitor) handleCriticalAlert(alert HealthAlert) {
	log.Printf("CRITICAL ALERT: %s", alert.Message)

	// Trigger emergency procedures:
	// 1. Notify on-call engineers
	// 2. Initiate failover procedures
	// 3. Scale up resources immediately
	// 4. Enable circuit breakers
}

func (hm *HealthMonitor) handleUnhealthyAlert(alert HealthAlert) {
	log.Printf("UNHEALTHY ALERT: %s", alert.Message)

	// Trigger recovery procedures:
	// 1. Increase monitoring frequency
	// 2. Prepare for scaling
	// 3. Check dependent services
	// 4. Enable degraded mode if necessary
}

func (hm *HealthMonitor) handleDegradedAlert(alert HealthAlert) {
	log.Printf("DEGRADED ALERT: %s", alert.Message)

	// Trigger preventive measures:
	// 1. Increase resource allocation
	// 2. Optimize performance
	// 3. Monitor trends
	// 4. Prepare scaling strategies
}

func (hm *HealthMonitor) GetCurrentStatus() HealthStatus {
	hm.mutex.RLock()
	defer hm.mutex.RUnlock()
	return hm.status
}

func (hm *HealthMonitor) GetHealthHistory() []HealthCheckResult {
	hm.mutex.RLock()
	defer hm.mutex.RUnlock()

	// Return a copy to prevent external modification
	history := make([]HealthCheckResult, len(hm.healthHistory))
	copy(history, hm.healthHistory)
	return history
}

func (hm *HealthMonitor) GetLastCheck() time.Time {
	hm.mutex.RLock()
	defer hm.mutex.RUnlock()
	return hm.lastCheck
}

func (hm *HealthMonitor) Stop() {
	hm.cancel()
	close(hm.alerts)
}
