package server

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var (
	db           *sql.DB
	requestCount = 0
	requestMutex sync.Mutex
	startTime    = time.Now()
	eventLog     = make([]Event, 0)
	eventMutex   sync.Mutex
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
	OwnerID     int       `json:"owner_id,omitempty"`
}

type BusinessOwner struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	Company   string    `json:"company"`
	CreatedAt time.Time `json:"created_at"`
}

type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  string    `json:"password,omitempty"`
	Type      string    `json:"type"` // "user" or "business_owner"
	CreatedAt time.Time `json:"created_at"`
}

type Event struct {
	Type      string      `json:"type"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

func InitDB() error {
	var err error
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "3306"
	}
	user := os.Getenv("DB_USER")
	if user == "" {
		user = "root"
	}
	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		password = "password"
	}
	dbname := os.Getenv("DB_NAME")
	if dbname == "" {
		dbname = "business_directory"
	}

	dsn := user + ":" + password + "@tcp(" + host + ":" + port + ")/" + dbname + "?parseTime=true"
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		return err
	}

	if err = db.Ping(); err != nil {
		return err
	}

	// Create tables
	if err = createTables(); err != nil {
		return err
	}

	// Seed initial data
	if err = seedData(); err != nil {
		log.Printf("Warning: Could not seed initial data: %v", err)
	}

	return nil
}

func createTables() error {
	// Users table
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			email VARCHAR(255) UNIQUE NOT NULL,
			password VARCHAR(255) NOT NULL,
			type ENUM('user', 'business_owner') NOT NULL DEFAULT 'user',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return err
	}

	// Business owners additional info
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS business_owners (
			id INT PRIMARY KEY,
			company VARCHAR(255),
			phone VARCHAR(50),
			FOREIGN KEY (id) REFERENCES users(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return err
	}

	// Businesses table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS businesses (
			id INT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			category VARCHAR(100) NOT NULL,
			description TEXT,
			phone VARCHAR(50),
			email VARCHAR(255),
			address TEXT,
			rating DECIMAL(3,1) DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			owner_id INT,
			FOREIGN KEY (owner_id) REFERENCES business_owners(id)
		)
	`)
	if err != nil {
		return err
	}

	return nil
}

func seedData() error {
	// Check if data already exists
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM businesses").Scan(&count)
	if err == nil && count > 0 {
		log.Println("Data already exists, skipping seed")
		return nil
	}

	// Sample business owners
	businessOwners := []struct {
		name    string
		email   string
		company string
		phone   string
	}{
		{"John Smith", "john@coffee.com", "Coffee Corner", "+1234567890"},
		{"Maria Garcia", "maria@techhub.com", "Tech Hub", "+1234567891"},
		{"David Chen", "david@fitnessfirst.com", "Fitness First", "+1234567892"},
		{"Sarah Johnson", "sarah@bookstore.com", "City Bookstore", "+1234567893"},
		{"Mike Wilson", "mike@autoshop.com", "Wilson Auto Service", "+1234567894"},
	}

	businesses := []struct {
		name        string
		category    string
		description string
		phone       string
		email       string
		address     string
		rating      float64
	}{
		{"Coffee Corner", "Restaurant", "Best coffee in town with fresh pastries", "+1234567801", "info@coffeecorner.com", "123 Main St, Downtown", 4.5},
		{"Tech Hub", "Technology", "Latest gadgets and computer repair services", "+1234567802", "support@techhub.com", "456 Tech Ave", 4.2},
		{"Fitness First Gym", "Healthcare", "Complete fitness center with personal trainers", "+1234567803", "fitness@first.com", "789 Health Blvd", 4.7},
		{"City Bookstore", "Retail", "Books for all ages with quiet reading areas", "+1234567804", "books@city.com", "321 Reading St", 4.3},
		{"Wilson Auto Service", "Services", "Full auto repair and maintenance services", "+1234567805", "service@wilsonauto.com", "654 Car Lane", 4.1},
		{"Bella Pizza", "Restaurant", "Authentic Italian pizza with fresh ingredients", "+1234567806", "bella@pizza.com", "987 Food Court", 4.8},
		{"Green Garden Spa", "Services", "Relaxing spa treatments and massages", "+1234567807", "spa@greengarden.com", "159 Wellness Rd", 4.6},
		{"Kids Play Center", "Entertainment", "Safe and fun environment for children", "+1234567808", "play@kidscenter.com", "753 Fun St", 4.4},
		{"Quick Hair Studio", "Services", "Modern hair styling and beauty services", "+1234567809", "hair@quick-studio.com", "852 Style Ave", 4.0},
		{"Fresh Market", "Retail", "Local farm fresh produce and groceries", "+1234567810", "fresh@market.com", "951 Organic Way", 4.2},
	}

	// Insert business owners and their businesses
	for i, owner := range businessOwners {
		result, err := db.Exec("INSERT INTO users (name, email, password, type) VALUES (?, ?, ?, 'business_owner')",
			owner.name, owner.email, "password123")

		if err != nil {
			log.Printf("Error seeding user %s: %v", owner.name, err)
			continue
		}

		userID, _ := result.LastInsertId()

		_, err = db.Exec("INSERT INTO business_owners (id, company, phone) VALUES (?, ?, ?)",
			userID, owner.company, owner.phone)

		if err != nil {
			log.Printf("Error seeding business owner %s: %v", owner.company, err)
			continue
		}

		// Insert corresponding business
		if i < len(businesses) {
			business := businesses[i]
			_, err = db.Exec("INSERT INTO businesses (name, category, description, phone, email, address, rating, owner_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
				business.name, business.category, business.description, business.phone, business.email, business.address, business.rating, userID)

			if err != nil {
				log.Printf("Error seeding business %s: %v", business.name, err)
			}
		}
	}

	log.Println("Sample data seeded successfully")
	return nil
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

	// API routes
	mux.HandleFunc("/health", corsMiddleware(healthHandler))
	mux.HandleFunc("/businesses", corsMiddleware(businessesRouter))
	mux.HandleFunc("/stats", corsMiddleware(statsHandler))
	mux.HandleFunc("/events", corsMiddleware(eventsHandler))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Apply CORS for static files
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Serve static files from ./web/ directory
		http.StripPrefix("/", http.FileServer(http.Dir("./web/"))).ServeHTTP(w, r)
	})

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
	rows, err := db.Query("SELECT id, name, category, description, phone, email, address, rating, created_at, owner_id FROM businesses ORDER BY created_at DESC")
	if err != nil {
		log.Printf("Error querying businesses: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "internal server error"})
		return
	}
	defer rows.Close()

	var businesses []Business
	for rows.Next() {
		var b Business
		err := rows.Scan(&b.ID, &b.Name, &b.Category, &b.Description, &b.Phone, &b.Email, &b.Address, &b.Rating, &b.CreatedAt, &b.OwnerID)
		if err != nil {
			log.Printf("Error scanning business: %v", err)
			continue
		}
		businesses = append(businesses, b)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(businesses)
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
		OwnerID     *int    `json:"owner_id,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request"})
		return
	}

	if req.Name == "" || req.Category == "" || req.Description == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "name, category, and description are required"})
		return
	}

	result, err := db.Exec("INSERT INTO businesses (name, category, description, phone, email, address, rating, owner_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		req.Name, req.Category, req.Description, req.Phone, req.Email, req.Address, req.Rating, req.OwnerID)

	if err != nil {
		log.Printf("Error creating business: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to create business"})
		return
	}

	id, _ := result.LastInsertId()
	business := Business{
		ID:          int(id),
		Name:        req.Name,
		Category:    req.Category,
		Description: req.Description,
		Phone:       req.Phone,
		Email:       req.Email,
		Address:     req.Address,
		Rating:      req.Rating,
		CreatedAt:   time.Now(),
	}
	if req.OwnerID != nil {
		business.OwnerID = *req.OwnerID
	}

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
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request"})
		return
	}

	// Build update query dynamically
	setParts := []string{}
	args := []interface{}{}

	if req.Name != "" {
		setParts = append(setParts, "name = ?")
		args = append(args, req.Name)
	}
	if req.Category != "" {
		setParts = append(setParts, "category = ?")
		args = append(args, req.Category)
	}
	if req.Description != "" {
		setParts = append(setParts, "description = ?")
		args = append(args, req.Description)
	}
	if req.Phone != "" {
		setParts = append(setParts, "phone = ?")
		args = append(args, req.Phone)
	}
	if req.Email != "" {
		setParts = append(setParts, "email = ?")
		args = append(args, req.Email)
	}
	if req.Address != "" {
		setParts = append(setParts, "address = ?")
		args = append(args, req.Address)
	}
	if req.Rating > 0 {
		setParts = append(setParts, "rating = ?")
		args = append(args, req.Rating)
	}

	if len(setParts) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "no valid fields to update"})
		return
	}

	query := "UPDATE businesses SET " + setParts[0]
	for i := 1; i < len(setParts); i++ {
		query += ", " + setParts[i]
	}
	query += " WHERE id = ?"
	args = append(args, req.ID)

	_, err := db.Exec(query, args...)
	if err != nil {
		log.Printf("Error updating business: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to update business"})
		return
	}

	// Get updated business
	var business Business
	err = db.QueryRow("SELECT id, name, category, description, phone, email, address, rating, created_at, owner_id FROM businesses WHERE id = ?", req.ID).
		Scan(&business.ID, &business.Name, &business.Category, &business.Description, &business.Phone, &business.Email, &business.Address, &business.Rating, &business.CreatedAt, &business.OwnerID)

	if err != nil {
		log.Printf("Error fetching updated business: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to fetch updated business"})
		return
	}

	logEvent("business_updated", "Business "+business.Name+" updated", business)

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
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request"})
		return
	}

	// Get business before deletion for logging
	var business Business
	err := db.QueryRow("SELECT id, name FROM businesses WHERE id = ?", req.ID).
		Scan(&business.ID, &business.Name)

	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "business not found"})
			return
		}
		log.Printf("Error fetching business for deletion: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "internal server error"})
		return
	}

	_, err = db.Exec("DELETE FROM businesses WHERE id = ?", req.ID)
	if err != nil {
		log.Printf("Error deleting business: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to delete business"})
		return
	}

	logEvent("business_deleted", "Business "+business.Name+" removed from directory", business)

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
