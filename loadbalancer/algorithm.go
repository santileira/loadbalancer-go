package main

const (
	AlgorithmRoundRobin = "round-robin"
)

type Algorithm interface {
	GetNextBackend(backends []*Backend) (*Backend, error)
}
