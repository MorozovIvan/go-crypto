package market

import (
	"errors"
	"sync"
	"time"
)

// CircuitState represents the state of the circuit breaker
type CircuitState int

const (
	StateClosed CircuitState = iota
	StateHalfOpen
	StateOpen
)

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	name          string
	maxFailures   int
	resetTimeout  time.Duration
	state         CircuitState
	failures      int
	lastFailTime  time.Time
	mutex         sync.RWMutex
	onStateChange func(name string, from CircuitState, to CircuitState)
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(name string, maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		name:         name,
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
		state:        StateClosed,
	}
}

// SetOnStateChange sets the callback for state changes
func (cb *CircuitBreaker) SetOnStateChange(fn func(name string, from CircuitState, to CircuitState)) {
	cb.onStateChange = fn
}

// Execute runs the given function with circuit breaker protection
func (cb *CircuitBreaker) Execute(fn func() (interface{}, error)) (interface{}, error) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	if cb.state == StateOpen {
		if time.Since(cb.lastFailTime) > cb.resetTimeout {
			cb.setState(StateHalfOpen)
		} else {
			return nil, errors.New("circuit breaker is open")
		}
	}

	result, err := fn()

	if err != nil {
		cb.onFailure()
		return nil, err
	}

	cb.onSuccess()
	return result, nil
}

// onSuccess handles successful execution
func (cb *CircuitBreaker) onSuccess() {
	cb.failures = 0
	if cb.state == StateHalfOpen {
		cb.setState(StateClosed)
	}
}

// onFailure handles failed execution
func (cb *CircuitBreaker) onFailure() {
	cb.failures++
	cb.lastFailTime = time.Now()

	if cb.failures >= cb.maxFailures {
		cb.setState(StateOpen)
	}
}

// setState changes the circuit breaker state
func (cb *CircuitBreaker) setState(state CircuitState) {
	if cb.state == state {
		return
	}

	oldState := cb.state
	cb.state = state

	if cb.onStateChange != nil {
		cb.onStateChange(cb.name, oldState, state)
	}
}

// GetState returns the current state
func (cb *CircuitBreaker) GetState() CircuitState {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.state
}

// GetStats returns circuit breaker statistics
func (cb *CircuitBreaker) GetStats() map[string]interface{} {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()

	return map[string]interface{}{
		"name":           cb.name,
		"state":          cb.state,
		"failures":       cb.failures,
		"max_failures":   cb.maxFailures,
		"last_fail_time": cb.lastFailTime,
	}
}

// CircuitBreakerManager manages multiple circuit breakers
type CircuitBreakerManager struct {
	breakers map[string]*CircuitBreaker
	mutex    sync.RWMutex
}

// NewCircuitBreakerManager creates a new manager
func NewCircuitBreakerManager() *CircuitBreakerManager {
	return &CircuitBreakerManager{
		breakers: make(map[string]*CircuitBreaker),
	}
}

// GetBreaker returns or creates a circuit breaker
func (cbm *CircuitBreakerManager) GetBreaker(name string, maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
	cbm.mutex.Lock()
	defer cbm.mutex.Unlock()

	if breaker, exists := cbm.breakers[name]; exists {
		return breaker
	}

	breaker := NewCircuitBreaker(name, maxFailures, resetTimeout)
	breaker.SetOnStateChange(func(name string, from CircuitState, to CircuitState) {
		// Log state changes
		// You can add logging here
	})

	cbm.breakers[name] = breaker
	return breaker
}

// GetAllStats returns statistics for all circuit breakers
func (cbm *CircuitBreakerManager) GetAllStats() map[string]interface{} {
	cbm.mutex.RLock()
	defer cbm.mutex.RUnlock()

	stats := make(map[string]interface{})
	for name, breaker := range cbm.breakers {
		stats[name] = breaker.GetStats()
	}

	return stats
}

// Global circuit breaker manager
var globalCBManager *CircuitBreakerManager

func init() {
	globalCBManager = NewCircuitBreakerManager()
}

// GetCircuitBreaker returns a circuit breaker for the given API
func GetCircuitBreaker(apiName string) *CircuitBreaker {
	return globalCBManager.GetBreaker(apiName, 5, 30*time.Second)
}

// GetCircuitBreakerStats returns all circuit breaker statistics
func GetCircuitBreakerStats() map[string]interface{} {
	return globalCBManager.GetAllStats()
}
