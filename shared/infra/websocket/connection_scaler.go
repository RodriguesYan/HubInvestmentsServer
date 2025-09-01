package websocket

import (
	"context"
	"log"
	"runtime"
	"sync"
	"time"
)

type ConnectionScaler struct {
	pool           *ConnectionPool
	config         ConnectionPoolConfig
	metrics        *ScalerMetrics
	mutex          sync.RWMutex
	ctx            context.Context
	cancel         context.CancelFunc
	lastScaleEvent time.Time
}

type ScalerMetrics struct {
	ScaleUpEvents   int64     `json:"scale_up_events"`
	ScaleDownEvents int64     `json:"scale_down_events"`
	CPUUsage        float64   `json:"cpu_usage_percent"`
	MemoryUsage     float64   `json:"memory_usage_percent"`
	ConnectionLoad  float64   `json:"connection_load_percent"`
	LastScaleTime   time.Time `json:"last_scale_time"`
	mutex           sync.RWMutex
}

type SystemMetrics struct {
	CPUUsagePercent    float64
	MemoryUsagePercent float64
	GoroutineCount     int
	HeapSizeMB         float64
	ActiveConnections  int
}

func NewConnectionScaler(pool *ConnectionPool, config ConnectionPoolConfig) *ConnectionScaler {
	ctx, cancel := context.WithCancel(context.Background())

	scaler := &ConnectionScaler{
		pool:           pool,
		config:         config,
		metrics:        &ScalerMetrics{},
		ctx:            ctx,
		cancel:         cancel,
		lastScaleEvent: time.Now(),
	}

	go scaler.startMonitoring()

	return scaler
}

func (s *ConnectionScaler) EvaluateScaling() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Prevent too frequent scaling events
	if time.Since(s.lastScaleEvent) < 30*time.Second {
		return
	}

	systemMetrics := s.getSystemMetrics()
	s.updateMetrics(systemMetrics)

	currentConnections := s.pool.GetConnectionCount()
	connectionLoad := float64(currentConnections) / float64(s.config.MaxConnections)

	// Determine if scaling is needed
	shouldScaleUp := s.shouldScaleUp(systemMetrics, connectionLoad)
	shouldScaleDown := s.shouldScaleDown(systemMetrics, connectionLoad)

	if shouldScaleUp {
		s.scaleUp(systemMetrics)
	} else if shouldScaleDown {
		s.scaleDown(systemMetrics)
	}
}

func (s *ConnectionScaler) shouldScaleUp(metrics SystemMetrics, connectionLoad float64) bool {
	// Scale up if:
	// 1. Connection load is above threshold
	// 2. CPU usage is manageable
	// 3. Memory usage is manageable
	// 4. We haven't reached max connections

	return connectionLoad > s.config.ScaleUpThreshold &&
		metrics.CPUUsagePercent < 85.0 &&
		metrics.MemoryUsagePercent < 85.0 &&
		metrics.ActiveConnections < s.config.MaxConnections
}

// shouldScaleDown determines if scaling down is needed
func (s *ConnectionScaler) shouldScaleDown(metrics SystemMetrics, connectionLoad float64) bool {
	// Scale down if:
	// 1. Connection load is below threshold
	// 2. We have more than minimum connections
	// 3. System resources are underutilized

	return connectionLoad < s.config.ScaleDownThreshold &&
		metrics.ActiveConnections > s.config.MinConnections &&
		metrics.CPUUsagePercent < 50.0 &&
		metrics.MemoryUsagePercent < 50.0
}

// scaleUp handles scaling up operations
func (s *ConnectionScaler) scaleUp(metrics SystemMetrics) {
	log.Printf("Scaling up WebSocket connections - Current: %d, CPU: %.1f%%, Memory: %.1f%%",
		metrics.ActiveConnections, metrics.CPUUsagePercent, metrics.MemoryUsagePercent)

	// Increase buffer sizes for better performance
	s.optimizeForHighLoad()

	s.lastScaleEvent = time.Now()

	s.metrics.mutex.Lock()
	s.metrics.ScaleUpEvents++
	s.metrics.LastScaleTime = time.Now()
	s.metrics.mutex.Unlock()
}

// scaleDown handles scaling down operations
func (s *ConnectionScaler) scaleDown(metrics SystemMetrics) {
	log.Printf("Scaling down WebSocket connections - Current: %d, CPU: %.1f%%, Memory: %.1f%%",
		metrics.ActiveConnections, metrics.CPUUsagePercent, metrics.MemoryUsagePercent)

	// Clean up idle connections more aggressively
	s.optimizeForLowLoad()

	s.lastScaleEvent = time.Now()

	s.metrics.mutex.Lock()
	s.metrics.ScaleDownEvents++
	s.metrics.LastScaleTime = time.Now()
	s.metrics.mutex.Unlock()
}

// optimizeForHighLoad optimizes settings for high connection load
func (s *ConnectionScaler) optimizeForHighLoad() {
	// Reduce cleanup frequency to save CPU
	// Increase buffer sizes
	// Adjust timeouts for better throughput

	log.Println("Optimizing WebSocket settings for high load")
}

func (s *ConnectionScaler) optimizeForLowLoad() {
	// Increase cleanup frequency to save memory
	// Reduce buffer sizes
	// Adjust timeouts for better resource utilization

	log.Println("Optimizing WebSocket settings for low load")

	// Trigger aggressive cleanup of idle connections
	s.pool.cleanupIdleConnections()
}

func (s *ConnectionScaler) getSystemMetrics() SystemMetrics {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// Calculate memory usage percentage (simplified)
	heapSizeMB := float64(memStats.HeapInuse) / 1024 / 1024
	memoryUsagePercent := (float64(memStats.HeapInuse) / float64(memStats.Sys)) * 100

	// CPU usage would require additional monitoring (simplified here)
	cpuUsagePercent := s.estimateCPUUsage()

	return SystemMetrics{
		CPUUsagePercent:    cpuUsagePercent,
		MemoryUsagePercent: memoryUsagePercent,
		GoroutineCount:     runtime.NumGoroutine(),
		HeapSizeMB:         heapSizeMB,
		ActiveConnections:  s.pool.GetConnectionCount(),
	}
}

func (s *ConnectionScaler) estimateCPUUsage() float64 {
	// In a real implementation, you would use proper CPU monitoring
	// This is a simplified estimation based on goroutine count and connections

	goroutines := runtime.NumGoroutine()
	connections := s.pool.GetConnectionCount()

	// Rough estimation: more goroutines and connections = higher CPU usage
	baseUsage := float64(goroutines) / 1000 * 10       // 10% per 1000 goroutines
	connectionUsage := float64(connections) / 1000 * 5 // 5% per 1000 connections

	totalUsage := baseUsage + connectionUsage
	if totalUsage > 100 {
		totalUsage = 100
	}

	return totalUsage
}

func (s *ConnectionScaler) updateMetrics(systemMetrics SystemMetrics) {
	s.metrics.mutex.Lock()
	defer s.metrics.mutex.Unlock()

	s.metrics.CPUUsage = systemMetrics.CPUUsagePercent
	s.metrics.MemoryUsage = systemMetrics.MemoryUsagePercent
	s.metrics.ConnectionLoad = float64(systemMetrics.ActiveConnections) / float64(s.config.MaxConnections) * 100
}

func (s *ConnectionScaler) startMonitoring() {
	ticker := time.NewTicker(15 * time.Second) // Monitor every 15 seconds
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.EvaluateScaling()
		}
	}
}

// GetMetrics returns current scaler metrics
func (s *ConnectionScaler) GetMetrics() ScalerMetrics {
	s.metrics.mutex.RLock()
	defer s.metrics.mutex.RUnlock()

	// Return a copy to avoid mutex copying
	return ScalerMetrics{
		ScaleUpEvents:   s.metrics.ScaleUpEvents,
		ScaleDownEvents: s.metrics.ScaleDownEvents,
		CPUUsage:        s.metrics.CPUUsage,
		MemoryUsage:     s.metrics.MemoryUsage,
		ConnectionLoad:  s.metrics.ConnectionLoad,
		LastScaleTime:   s.metrics.LastScaleTime,
	}
}

// Stop stops the connection scaler
func (s *ConnectionScaler) Stop() {
	s.cancel()
}
