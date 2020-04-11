package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

func main() {

	var port int
	flag.IntVar(&port, "port", 8080, "Port of the load balancer")
	flag.Parse()

	http.HandleFunc("/ping", pingHandler)
	http.HandleFunc("/healthcheck", healthCheckHandler)

	log.Printf("Backend started at :%d\n", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		log.Fatal(err)
	}

}

func pingHandler(w http.ResponseWriter, r *http.Request) {
	message := fmt.Sprintf("Pong on the %s", r.Host)
	log.Print(message)
	fmt.Fprintln(w, message)
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	message := fmt.Sprintf("The service is running on %s", r.Host)
	log.Print(message)
	fmt.Fprintln(w, message)
}
