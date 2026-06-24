package main

import "net/http"

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /wallet/info", walletInfo)
	mux.Handle("POST /order/create", http.HandlerFunc(createOrder))
	mux.Handle("/metrics", metricsHandler)
	http.HandleFunc("/login", login)
	http.ListenAndServe(":8080", mux)
}

func login(w http.ResponseWriter, r *http.Request) {}

func walletInfo(w http.ResponseWriter, r *http.Request) {}

func createOrder(w http.ResponseWriter, r *http.Request) {}

var metricsHandler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
