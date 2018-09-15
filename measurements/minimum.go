package measurements

import (
	"sync"
)

// MinimumMeasurement implements a minimum value measurement
type MinimumMeasurement struct {
	value float64
	mu    sync.RWMutex
}

func (m *MinimumMeasurement) Add(sample float64) (float64, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	oldValue := float64(m.value)
	if oldValue == 0.0 || sample < oldValue {
		m.value = sample
	}
	return m.value, oldValue == m.value
}

func (m *MinimumMeasurement) Get() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.value
}

func (m *MinimumMeasurement) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.value = 0.0
}
