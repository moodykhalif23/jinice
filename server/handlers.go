package server

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

var (
	requestCount = 0
	requestMutex sync.Mutex
	startTime    = time.Now()
)

func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		requestMutex.Lock()
		requestCount++
		requestMutex.Unlock()

		next(w, r)
	}
}

func NewRouter() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", corsMiddleware(healthHandler))
	mux.HandleFunc("/hello", corsMiddleware(helloHandler))
	mux.HandleFunc("/time", corsMiddleware(timeHandler))
	mux.HandleFunc("/echo", corsMiddleware(echoHandler))
	mux.HandleFunc("/stats", corsMiddleware(statsHandler))
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

func timeHandler(w http.ResponseWriter, r *http.Request) {
	resp := map[string]interface{}{
		"now":       time.Now(),
		"unix":      time.Now().Unix(),
		"formatted": time.Now().Format("2006-01-02 15:04:05 MST"),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func echoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "POST method required"})
		return
	}

	var data map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid JSON"})
		return
	}

	resp := map[string]interface{}{
		"echo":      data,
		"timestamp": time.Now().Unix(),
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
	requestMutex.Lock()
	count := requestCount
	requestMutex.Unlock()

	uptime := time.Since(startTime).Seconds()
	resp := map[string]interface{}{
		"total_requests": count,
		"uptime_seconds": uptime,
		"start_time":     startTime.Format("2006-01-02 15:04:05 MST"),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
