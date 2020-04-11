package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

const (
	Attempts    = "Attempts"
	Retries     = "Retries"
	MaxRetries  = 3
	MaxAttempts = 3
)

func main() {

	var backends string
	var port int
	var algorithm string

	flag.StringVar(&backends, "backends", "", "Servers available, use comma to separate")
	flag.IntVar(&port, "port", 8080, "Port of the load balancer")
	flag.StringVar(&algorithm, "algorithm", "round-robin", "Strategy to select the backend")
	flag.Parse()

	if len(backends) == 0 {
		log.Fatal("Error, Please provide one or more backends")
	}

	serverPool := generateServerPool(algorithm, backends)

	go serverPool.HealthCheck()

	http.HandleFunc("/ping", serverPool.PingHandler)
	log.Printf("Load Balancer started at :%d\n", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		log.Fatal(err)
	}
}

func generateServerPool(algorithmName string, backends string) *ServerPool {

	algorithm := generateAlgorithm(algorithmName)

	serverPool := NewServerPool(algorithm)

	backendURLStrList := strings.Split(backends, ",")

	for _, backendURLStr := range backendURLStrList {

		backend, err := generateBackend(backendURLStr, serverPool)
		if err != nil {
			continue
		}
		serverPool.AddBackend(backend)
	}

	return serverPool
}

func generateAlgorithm(algorithmName string) Algorithm {
	var algorithm Algorithm
	switch algorithmName {
	case AlgorithmRoundRobin:
		algorithm = NewRoundRobin()
	default:
		log.Printf("Warning, Algorithm %s doesn't exist in our system\n", algorithm)
		algorithm = NewRoundRobin()
	}
	return algorithm
}

func generateBackend(backendURLStr string, serverPool *ServerPool) (*Backend, error) {

	backendURL, err := url.Parse(backendURLStr)
	if err != nil {
		message := fmt.Sprintf("Error generating backend url (%s), err: %s\n", backendURLStr, err.Error())
		log.Printf(message)
		return nil, fmt.Errorf(message)
	}

	reverseProxy := generateReverseProxy(backendURL, serverPool)

	return NewBackend(backendURL, reverseProxy), nil
}

func generateReverseProxy(backendURL *url.URL, serverPool *ServerPool) *httputil.ReverseProxy {

	reverseProxy := httputil.NewSingleHostReverseProxy(backendURL)
	reverseProxy.ErrorHandler = func(writer http.ResponseWriter, request *http.Request, err error) {
		log.Printf("Error requesting response to backend %s, err: %s\n", request.URL.String(), err.Error())
		retries := getValueFromContext(request.Context(), Retries)
		if retries < MaxRetries {
			select {
			case <-time.After(10 * time.Millisecond):
				log.Printf("Executing again the request in backend %s, retry %d", request.URL.String(), retries)
				ctx := context.WithValue(request.Context(), Retries, retries+1)
				reverseProxy.ServeHTTP(writer, request.WithContext(ctx))
			}

			return
		}

		attempts := getValueFromContext(request.Context(), Attempts)
		ctx := context.WithValue(request.Context(), Attempts, attempts+1)

		serverPool.UpdateBackendStatus(request.URL, false)
		serverPool.PingHandler(writer, request.WithContext(ctx))
	}
	return reverseProxy
}

func getValueFromContext(context context.Context, key string) int {
	if value, ok := context.Value(key).(int); ok {
		return value
	}

	return 0
}
