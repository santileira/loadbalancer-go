package main

import (
	"log"
	"net/http"
	"net/url"
	"time"
)

type ServerPool struct {
	backends  []*Backend
	algorithm Algorithm
}

func NewServerPool(algorithm Algorithm) *ServerPool {
	return &ServerPool{
		algorithm: algorithm,
	}
}

func (s *ServerPool) AddBackend(backend *Backend) {
	s.backends = append(s.backends, backend)
}

func (s *ServerPool) HealthCheck() {

	ticker := time.NewTicker(1 * time.Minute)

	for {
		select {
		case <-ticker.C:
			log.Print("Starting health check")

			for _, backend := range s.backends {
				backend.UpdateStatus()
			}

			log.Print("Finishing health check")
		}
	}
}

func (s *ServerPool) UpdateBackendStatus(backendURL *url.URL, alive bool) {

	for _, backend := range s.backends {
		if backend.URL.String() == backendURL.String() {
			backend.SetAlive(alive)
			break
		}
	}

}

func (s *ServerPool) PingHandler(writer http.ResponseWriter, request *http.Request) {
	attempts := getValueFromContext(request.Context(), Attempts)
	if attempts > MaxAttempts {
		log.Printf("%s(%s) Max attempts reached, terminating\n", request.RemoteAddr, request.URL.Path)
		http.Error(writer, "Service not available", http.StatusServiceUnavailable)
		return
	}

	backend, err := s.algorithm.GetNextBackend(s.backends)
	if err != nil {
		log.Printf("Error getting next backend %s", err.Error())
		http.Error(writer, "Service not available", http.StatusServiceUnavailable)
		return
	}

	backend.ReverseProxy.ServeHTTP(writer, request)
	return
}
