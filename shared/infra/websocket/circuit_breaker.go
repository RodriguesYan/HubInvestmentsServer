package websocket

import (
	"errors"
	"sync"
	"time"
)

type CircuitBreakerState int

const (
	CircuitBreakerClosed CircuitBreakerState = iota
	CircuitBreakerOpen
	CircuitBreakerHalfOpen
)

// CircuitBreaker implements the circuit breaker pattern for WebSocket connections
type CircuitBreaker struct {
	config          CircuitBreakerConfig
	state           CircuitBreakerState
	failureCount    int
	lastFailureTime time.Time
	halfOpenCalls   int
	mutex           sync.RWMutex
}

func NewCircuitBreaker(config CircuitBreakerConfig) *CircuitBreaker {
	return &CircuitBreaker{
		config: config,
		state:  CircuitBreakerClosed,
	}
}

func (cb *CircuitBreaker) Execute(fn func() error) error {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	switch cb.state {
	case CircuitBreakerOpen:
		if time.Since(cb.lastFailureTime) > cb.config.RecoveryTimeout {
			cb.state = CircuitBreakerHalfOpen
			cb.halfOpenCalls = 0
		} else {
			return errors.New("circuit breaker is open")
		}
	case CircuitBreakerHalfOpen:
		if cb.halfOpenCalls >= cb.config.HalfOpenMaxCalls {
			return errors.New("circuit breaker half-open call limit exceeded")
		}
		cb.halfOpenCalls++
	}

	err := fn()

	if err != nil {
		cb.onFailure()
	} else {
		cb.onSuccess()
	}

	return err
}

// onFailure handles failure cases
func (cb *CircuitBreaker) onFailure() {
	cb.failureCount++
	cb.lastFailureTime = time.Now()

	if cb.state == CircuitBreakerHalfOpen {
		cb.state = CircuitBreakerOpen
		cb.halfOpenCalls = 0
	} else if cb.failureCount >= cb.config.FailureThreshold {
		cb.state = CircuitBreakerOpen
	}
}

func (cb *CircuitBreaker) onSuccess() {
	if cb.state == CircuitBreakerHalfOpen {
		cb.state = CircuitBreakerClosed
		cb.failureCount = 0
		cb.halfOpenCalls = 0
	} else if cb.state == CircuitBreakerClosed {
		cb.failureCount = 0
	}
}

func (cb *CircuitBreaker) GetState() CircuitBreakerState {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.state
}

func (cb *CircuitBreaker) GetFailureCount() int {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.failureCount
}

func (cb *CircuitBreaker) Reset() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	cb.state = CircuitBreakerClosed
	cb.failureCount = 0
	cb.halfOpenCalls = 0
}
