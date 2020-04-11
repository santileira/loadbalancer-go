package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
)

type Backend struct {
	URL          *url.URL
	Alive        bool
	ReverseProxy *httputil.ReverseProxy
	Mutex        sync.RWMutex
}

func NewBackend(url *url.URL, reverseProxy *httputil.ReverseProxy) *Backend {
	return &Backend{
		URL:          url,
		Alive:        true,
		ReverseProxy: reverseProxy,
	}
}

func (b *Backend) UpdateStatus() {
	resp, err := http.Get(fmt.Sprintf("%s/healthcheck", b.URL.String()))

	if err != nil || !(http.StatusOK <= resp.StatusCode && resp.StatusCode <= http.StatusMultipleChoices) {
		log.Printf("Backend %s is DOWN\n", b.URL.String())
		b.SetAlive(false)
		return
	}

	log.Printf("Backend %s is UP\n", b.URL.String())
	b.SetAlive(true)
	return
}

func (b *Backend) SetAlive(alive bool) {
	b.Mutex.Lock()
	defer b.Mutex.Unlock()
	b.Alive = alive
}

func (b *Backend) IsAlive() bool {
	b.Mutex.Lock()
	defer b.Mutex.Unlock()
	return b.Alive
}
