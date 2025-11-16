package server

import (
    "encoding/json"
    "net/http"
)

func NewRouter() http.Handler {
    mux := http.NewServeMux()
    mux.HandleFunc("/health", healthHandler)
    mux.HandleFunc("/hello", helloHandler)
    return mux
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("OK"))
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
    resp := map[string]string{"message": "Hello, world!"}
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(resp)
}
