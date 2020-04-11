package main

import (
	"fmt"
	"log"
	"sync/atomic"
)

type RoundRobin struct {
	current uint64
}

func NewRoundRobin() *RoundRobin {
	return &RoundRobin{
		current: uint64(0),
	}
}

func (r *RoundRobin) GetNextBackend(backends []*Backend) (*Backend, error) {

	lenBackends := len(backends)
	nextBackend := int(atomic.AddUint64(&r.current, uint64(1)) % uint64(lenBackends))

	for i := nextBackend; i < (nextBackend + lenBackends); i++ {
		idx := i % lenBackends
		backend := backends[idx]
		if backend.IsAlive() {

			if idx != nextBackend {
				atomic.StoreUint64(&r.current, uint64(idx))
			}

			log.Printf("Backend selected to response the request is %s\n", backend.URL.String())
			return backend, nil
		}
	}

	log.Print("Error, No backend is alive\n")
	return nil, fmt.Errorf("no backend is alive")

}
