package accrual

import (
	"sync"
)

type WorkerTimeoutMap struct {
	workers map[int]chan int
	mu      sync.RWMutex
}

func NewWorkerTimeoutMap(rateLimit int) *WorkerTimeoutMap {
	return &WorkerTimeoutMap{
		workers: make(map[int]chan int, rateLimit),
		mu:      sync.RWMutex{},
	}
}

func (m *WorkerTimeoutMap) AddWorker(key int, value chan int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.workers[key] = value
}

func (m *WorkerTimeoutMap) GetWorker(key int) chan int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.workers[key]
}

func (m *WorkerTimeoutMap) Broadcast(timeout int) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, ch := range m.workers {
		select {
		case ch <- timeout:
			continue
		default:
			continue
		}
	}
}
