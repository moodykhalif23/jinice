package server

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

var (
	requestCount   = 0
	requestMutex   sync.Mutex
	startTime      = time.Now()
	businessMutex  sync.Mutex
	businesses     = make(map[int]Business)
	nextBusinessID = 1
	eventLog       = make([]Event, 0)
	eventMutex     sync.Mutex
)

type Business struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Category    string    `json:"category"`
	Description string    `json:"description"`
	Phone       string    `json:"phone"`
	Email       string    `json:"email"`
	Address     string    `json:"address"`
	Rating      float64   `json:"rating"`
	CreatedAt   time.Time `json:"created_at"`
}

type Event struct {
	Type      string      `json:"type"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

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
	mux.HandleFunc("/businesses", corsMiddleware(businessesRouter))
	mux.HandleFunc("/stats", corsMiddleware(statsHandler))
	mux.HandleFunc("/events", corsMiddleware(eventsHandler))
	return mux
}

func businessesRouter(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getBusinessesHandler(w, r)
	case http.MethodPost:
		createBusinessHandler(w, r)
	case http.MethodPut:
		updateBusinessHandler(w, r)
	case http.MethodDelete:
		deleteBusinessHandler(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func getBusinessesHandler(w http.ResponseWriter, _ *http.Request) {
	businessMutex.Lock()
	businessList := make([]Business, 0, len(businesses))
	for _, business := range businesses {
		businessList = append(businessList, business)
	}
	businessMutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(businessList)
}

func createBusinessHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Name        string  `json:"name"`
		Category    string  `json:"category"`
		Description string  `json:"description"`
		Phone       string  `json:"phone"`
		Email       string  `json:"email"`
		Address     string  `json:"address"`
		Rating      float64 `json:"rating"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request"})
		return
	}

	businessMutex.Lock()
	business := Business{
		ID:          nextBusinessID,
		Name:        req.Name,
		Category:    req.Category,
		Description: req.Description,
		Phone:       req.Phone,
		Email:       req.Email,
		Address:     req.Address,
		Rating:      req.Rating,
		CreatedAt:   time.Now(),
	}
	businesses[nextBusinessID] = business
	nextBusinessID++
	businessMutex.Unlock()

	logEvent("business_created", "Business "+business.Name+" added to directory", business)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(business)
}

func updateBusinessHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ID          int     `json:"id"`
		Name        string  `json:"name"`
		Category    string  `json:"category"`
		Description string  `json:"description"`
		Phone       string  `json:"phone"`
		Email       string  `json:"email"`
		Address     string  `json:"address"`
		Rating      float64 `json:"rating"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	businessMutex.Lock()
	if business, ok := businesses[req.ID]; ok {
		if req.Name != "" {
			business.Name = req.Name
		}
		if req.Category != "" {
			business.Category = req.Category
		}
		if req.Description != "" {
			business.Description = req.Description
		}
		if req.Phone != "" {
			business.Phone = req.Phone
		}
		if req.Email != "" {
			business.Email = req.Email
		}
		if req.Address != "" {
			business.Address = req.Address
		}
		if req.Rating > 0 {
			business.Rating = req.Rating
		}
		businesses[req.ID] = business
		logEvent("business_updated", "Business "+business.Name+" updated", business)
	}
	business := businesses[req.ID]
	businessMutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(business)
}

func deleteBusinessHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ID int `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	businessMutex.Lock()
	if business, ok := businesses[req.ID]; ok {
		delete(businesses, req.ID)
		logEvent("business_deleted", "Business "+business.Name+" removed from directory", business)
	}
	businessMutex.Unlock()

	w.WriteHeader(http.StatusOK)
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

func logEvent(eventType, message string, data interface{}) {
	eventMutex.Lock()
	event := Event{
		Type:      eventType,
		Message:   message,
		Data:      data,
		Timestamp: time.Now(),
	}
	eventLog = append(eventLog, event)
	if len(eventLog) > 100 {
		eventLog = eventLog[1:]
	}
	eventMutex.Unlock()
}

func eventsHandler(w http.ResponseWriter, _ *http.Request) {
	eventMutex.Lock()
	events := make([]Event, len(eventLog))
	copy(events, eventLog)
	eventMutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}
