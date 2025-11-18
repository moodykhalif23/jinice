package server

import (
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	db           *sql.DB
	jwtSecret    = generateJWTSecret()
	requestCount = 0
	requestMutex sync.Mutex
	startTime    = time.Now()
	eventLog     = make([]SystemEvent, 0)
	eventMutex   sync.Mutex
)

// generateJWTSecret generates a random JWT secret key
func generateJWTSecret() []byte {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		// Fallback to a static key if random generation fails
		key = []byte("fallback-secret-key-change-in-production-12345678")
	}
	return key
}

// hashPassword hashes a password using bcrypt
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// checkPasswordHash checks if a password matches its hash
func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// generateToken generates a JWT token for a user and stores it in the database
func generateToken(user User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"type":    user.Type,
		"exp":     time.Now().Add(time.Hour * 24).Unix(), // 24 hours
	})
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	// Store token in database
	expiresAt := time.Now().Add(time.Hour * 24)
	_, err = db.Exec("INSERT INTO sessions (user_id, token, expires_at) VALUES (?, ?, ?)",
		user.ID, tokenString, expiresAt)
	if err != nil {
		log.Printf("Error storing session: %v", err)
		return "", err
	}

	return tokenString, nil
}

// authenticateUser verifies JWT token from database and returns user claims
func authenticateToken(r *http.Request) (jwt.MapClaims, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return nil, fmt.Errorf("authorization header required")
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	// Check if token exists in database and is not expired
	var userID int
	var expiresAt time.Time
	err := db.QueryRow("SELECT user_id, expires_at FROM sessions WHERE token = ? AND expires_at > NOW()",
		tokenString).Scan(&userID, &expiresAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("invalid or expired token")
		}
		return nil, err
	}

	// Verify JWT signature
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// authMiddleware wraps handlers to require authentication
func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, err := authenticateToken(r)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
			return
		}

		// Store claims in request context for handlers to use
		r.Header.Set("X-User-ID", fmt.Sprintf("%.0f", claims["user_id"]))
		r.Header.Set("X-User-Type", claims["type"].(string))

		next(w, r)
	}
}

// businessOwnerOnly middleware ensures only business owners can access
func businessOwnerOnly(next http.HandlerFunc) http.HandlerFunc {
	return authMiddleware(func(w http.ResponseWriter, r *http.Request) {
		userType := r.Header.Get("X-User-Type")
		if userType != "business_owner" {
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]string{"error": "business owner access required"})
			return
		}
		next(w, r)
	})
}

// eventOwnerOnly middleware ensures only event owners can access
func eventOwnerOnly(next http.HandlerFunc) http.HandlerFunc {
	return authMiddleware(func(w http.ResponseWriter, r *http.Request) {
		userType := r.Header.Get("X-User-Type")
		if userType != "event_owner" && userType != "business_owner" {
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]string{"error": "event owner or business owner access required"})
			return
		}
		next(w, r)
	})
}

type Business struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Category    string    `json:"category"`
	Description string    `json:"description"`
	Phone       string    `json:"phone"`
	Email       string    `json:"email"`
	Address     string    `json:"address"`
	ImageURL    string    `json:"image_url,omitempty"`
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

type SystemEvent struct {
	Type      string      `json:"type"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

type BusinessEvent struct {
	ID          int       `json:"id"`
	OwnerID     int       `json:"owner_id"`
	BusinessID  *int      `json:"business_id,omitempty"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	EventDate   time.Time `json:"event_date"`
	Location    string    `json:"location"`
	ImageURL    string    `json:"image_url,omitempty"`
	Price       float64   `json:"price"`
	Category    string    `json:"category"`
	CreatedAt   time.Time `json:"created_at"`
}

type Booking struct {
	ID        int       `json:"id"`
	EventID   int       `json:"event_id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	Tickets   int       `json:"tickets"`
	Notes     string    `json:"notes"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
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

	// Initialize image storage
	if err = InitImageStorage(); err != nil {
		log.Printf("Warning: Could not initialize image storage: %v", err)
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
			type ENUM('user', 'business_owner', 'event_owner') NOT NULL DEFAULT 'user',
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

	// Event owners additional info
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS event_owners (
			id INT PRIMARY KEY,
			organization VARCHAR(255),
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

	// Business views table for observability
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS business_views (
			id INT AUTO_INCREMENT PRIMARY KEY,
			business_id INT NOT NULL,
			user_ip VARCHAR(45),
			user_agent TEXT,
			viewed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (business_id) REFERENCES businesses(id) ON DELETE CASCADE,
			INDEX idx_business_views_business_id (business_id),
			INDEX idx_business_views_viewed_at (viewed_at)
		)
	`)
	if err != nil {
		return err
	}

	// Sessions table for storing auth tokens
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS sessions (
			id INT AUTO_INCREMENT PRIMARY KEY,
			user_id INT NOT NULL,
			token VARCHAR(500) NOT NULL UNIQUE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			expires_at TIMESTAMP NOT NULL,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			INDEX idx_sessions_token (token),
			INDEX idx_sessions_expires_at (expires_at)
		)
	`)
	if err != nil {
		return err
	}

	// Events table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS events (
			id INT AUTO_INCREMENT PRIMARY KEY,
			owner_id INT NOT NULL,
			business_id INT,
			title VARCHAR(255) NOT NULL,
			description TEXT,
			event_date DATETIME NOT NULL,
			location VARCHAR(255),
			price DECIMAL(10,2) DEFAULT 0,
			category VARCHAR(100),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (business_id) REFERENCES businesses(id) ON DELETE SET NULL,
			INDEX idx_events_owner_id (owner_id),
			INDEX idx_events_business_id (business_id),
			INDEX idx_events_event_date (event_date)
		)
	`)
	if err != nil {
		return err
	}

	// Bookings table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS bookings (
			id INT AUTO_INCREMENT PRIMARY KEY,
			event_id INT NOT NULL,
			name VARCHAR(255) NOT NULL,
			email VARCHAR(255) NOT NULL,
			phone VARCHAR(50),
			tickets INT NOT NULL DEFAULT 1,
			notes TEXT,
			status ENUM('pending', 'confirmed', 'cancelled') DEFAULT 'pending',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE CASCADE,
			INDEX idx_bookings_event_id (event_id),
			INDEX idx_bookings_email (email),
			INDEX idx_bookings_status (status)
		)
	`)
	if err != nil {
		return err
	}

	// Images table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS images (
			id INT AUTO_INCREMENT PRIMARY KEY,
			entity_type VARCHAR(50) NOT NULL,
			entity_id INT NOT NULL,
			image_url VARCHAR(500) NOT NULL,
			storage_path VARCHAR(500),
			caption VARCHAR(255),
			display_order INT DEFAULT 0,
			is_primary BOOLEAN DEFAULT FALSE,
			uploaded_by INT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (uploaded_by) REFERENCES users(id) ON DELETE SET NULL,
			INDEX idx_images_entity (entity_type, entity_id),
			INDEX idx_images_created_at (created_at)
		)
	`)
	if err != nil {
		return err
	}

	// Image metadata table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS image_metadata (
			id INT AUTO_INCREMENT PRIMARY KEY,
			image_id INT NOT NULL,
			file_size INT,
			width INT,
			height INT,
			mime_type VARCHAR(100),
			original_filename VARCHAR(255),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (image_id) REFERENCES images(id) ON DELETE CASCADE,
			INDEX idx_image_metadata_image_id (image_id)
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
		hashedPassword, err := hashPassword("password123")
		if err != nil {
			log.Printf("Error hashing password for %s: %v", owner.name, err)
			continue
		}

		result, err := db.Exec("INSERT INTO users (name, email, password, type) VALUES (?, ?, ?, 'business_owner')",
			owner.name, owner.email, hashedPassword)

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
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

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

	// Auth routes (no auth required)
	mux.HandleFunc("/register", corsMiddleware(registerHandler))
	mux.HandleFunc("/login", corsMiddleware(loginHandler))
	mux.HandleFunc("/logout", corsMiddleware(authMiddleware(logoutHandler)))

	// API routes
	mux.HandleFunc("/health", corsMiddleware(healthHandler))

	// Business routes
	mux.HandleFunc("/businesses", corsMiddleware(businessesRouter))
	mux.HandleFunc("/business/", corsMiddleware(getBusinessByIDHandler))
	mux.HandleFunc("/my-businesses", corsMiddleware(businessOwnerOnly(getMyBusinessesHandler)))
	mux.HandleFunc("/my-business-stats", corsMiddleware(businessOwnerOnly(getMyBusinessStatsHandler)))

	// Event routes
	mux.HandleFunc("/business-events", corsMiddleware(businessEventsRouter))
	mux.HandleFunc("/event/", corsMiddleware(getEventByIDHandler))
	mux.HandleFunc("/my-events", corsMiddleware(eventOwnerOnly(getMyEventsHandler)))

	// Booking routes
	mux.HandleFunc("/bookings", corsMiddleware(bookingsRouter))

	// Global stats (no auth required)
	mux.HandleFunc("/stats", corsMiddleware(statsHandler))
	mux.HandleFunc("/system-events", corsMiddleware(systemEventsHandler))

	// Image routes
	mux.HandleFunc("/images", corsMiddleware(getImagesHandler))
	mux.HandleFunc("/images/upload", corsMiddleware(authMiddleware(uploadImageHandler)))
	mux.HandleFunc("/images/add-url", corsMiddleware(authMiddleware(addImageURLHandler)))
	mux.HandleFunc("/images/update", corsMiddleware(authMiddleware(updateImageHandler)))
	mux.HandleFunc("/images/delete", corsMiddleware(authMiddleware(deleteImageHandler)))

	// Serve uploaded files
	mux.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir(uploadDir))))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Apply CORS for static files
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

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
		// GET is public - no auth required
		getBusinessesHandler(w, r)
	case http.MethodPost:
		// POST requires business owner auth
		businessOwnerOnly(createBusinessHandler)(w, r)
	case http.MethodPut:
		// PUT requires business owner auth
		businessOwnerOnly(updateBusinessHandler)(w, r)
	case http.MethodDelete:
		// DELETE requires business owner auth
		businessOwnerOnly(deleteBusinessHandler)(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
		Type     string `json:"type"` // "business_owner" or "event_owner"
		Company  string `json:"company"`
		Phone    string `json:"phone"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request"})
		return
	}

	if req.Name == "" || req.Email == "" || req.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "name, email, and password are required"})
		return
	}

	// Hash the password
	hashedPassword, err := hashPassword(req.Password)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "internal server error"})
		return
	}

	// Determine user type (default to business_owner for backward compatibility)
	userType := "business_owner"
	if req.Type != "" {
		userType = req.Type
	}

	// Insert user
	result, err := db.Exec("INSERT INTO users (name, email, password, type) VALUES (?, ?, ?, ?)",
		req.Name, req.Email, hashedPassword, userType)

	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(map[string]string{"error": "email already exists"})
			return
		}
		log.Printf("Error creating user: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to create user"})
		return
	}

	userID, _ := result.LastInsertId()

	// Insert business owner info if type is business_owner
	if userType == "business_owner" {
		_, err = db.Exec("INSERT INTO business_owners (id, company, phone) VALUES (?, ?, ?)",
			userID, req.Company, req.Phone)

		if err != nil {
			log.Printf("Error creating business owner: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "failed to create business owner profile"})
			return
		}
	}

	// Insert event owner info if type is event_owner
	if userType == "event_owner" {
		_, err = db.Exec("INSERT INTO event_owners (id, organization, phone) VALUES (?, ?, ?)",
			userID, req.Company, req.Phone)

		if err != nil {
			log.Printf("Error creating event owner: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "failed to create event owner profile"})
			return
		}
	}

	// Get the created user for response
	var user User
	err = db.QueryRow("SELECT id, name, email, type, created_at FROM users WHERE id = ?", userID).
		Scan(&user.ID, &user.Name, &user.Email, &user.Type, &user.CreatedAt)

	if err != nil {
		log.Printf("Error fetching created user: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "user created but could not retrieve"})
		return
	}

	// Generate JWT token and store in database
	token, err := generateToken(user)
	if err != nil {
		log.Printf("Error generating token: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to generate token"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user":    user,
		"token":   token,
		"message": "User registered successfully",
	})
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request"})
		return
	}

	if req.Email == "" || req.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "email and password are required"})
		return
	}

	// Get user from database
	var user User
	err := db.QueryRow("SELECT id, name, email, password, type, created_at FROM users WHERE email = ?", req.Email).
		Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.Type, &user.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid credentials"})
			return
		}
		log.Printf("Error fetching user: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "internal server error"})
		return
	}

	// Check password
	if !checkPasswordHash(req.Password, user.Password) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid credentials"})
		return
	}

	// Generate JWT token and store in database
	token, err := generateToken(user)
	if err != nil {
		log.Printf("Error generating token: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to generate token"})
		return
	}

	// Don't send password in response
	user.Password = ""

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user":  user,
		"token": token,
	})
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "authorization header required"})
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	// Delete token from database
	_, err := db.Exec("DELETE FROM sessions WHERE token = ?", tokenString)
	if err != nil {
		log.Printf("Error deleting session: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to logout"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "logged out successfully"})
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func getBusinessesHandler(w http.ResponseWriter, _ *http.Request) {
	rows, err := db.Query(`
		SELECT id, name, category, description, phone, email, address,
		  (SELECT image_url FROM images WHERE entity_type = 'business' AND entity_id = businesses.id ORDER BY is_primary DESC, display_order ASC, created_at ASC LIMIT 1) as image_url,
		  rating, created_at, owner_id
		FROM businesses
		ORDER BY created_at DESC
	`)
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
		var imageURL sql.NullString
		err := rows.Scan(&b.ID, &b.Name, &b.Category, &b.Description, &b.Phone, &b.Email, &b.Address, &imageURL, &b.Rating, &b.CreatedAt, &b.OwnerID)
		if err != nil {
			log.Printf("Error scanning business: %v", err)
			continue
		}
		if imageURL.Valid {
			b.ImageURL = imageURL.String
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

	userID := r.Header.Get("X-User-ID")
	ownerID, err := strconv.Atoi(userID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid user ID"})
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

	if req.Name == "" || req.Category == "" || req.Description == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "name, category, and description are required"})
		return
	}

	result, err := db.Exec("INSERT INTO businesses (name, category, description, phone, email, address, rating, owner_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		req.Name, req.Category, req.Description, req.Phone, req.Email, req.Address, req.Rating, ownerID)

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
		OwnerID:     ownerID,
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

func getBusinessByIDHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from URL path - trim "/business/"
	idStr := strings.TrimPrefix(r.URL.Path, "/business/")
	if idStr == r.URL.Path { // Path didn't contain /business/
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "business ID required"})
		return
	}

	if idStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "business ID required"})
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid business ID"})
		return
	}

	var business Business
	err = db.QueryRow(`
		SELECT id, name, category, description, phone, email, address,
		  (SELECT image_url FROM images WHERE entity_type = 'business' AND entity_id = businesses.id ORDER BY is_primary DESC, display_order ASC, created_at ASC LIMIT 1) as image_url,
		  rating, created_at, owner_id
		FROM businesses WHERE id = ?
	`, id).
		Scan(&business.ID, &business.Name, &business.Category, &business.Description, &business.Phone, &business.Email, &business.Address, &business.ImageURL, &business.Rating, &business.CreatedAt, &business.OwnerID)

	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "business not found"})
			return
		}
		log.Printf("Error fetching business: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "internal server error"})
		return
	}

	// Track business view (optional - don't fail if it errors)
	userIP := r.RemoteAddr
	userAgent := r.UserAgent()
	_, _ = db.Exec("INSERT INTO business_views (business_id, user_ip, user_agent) VALUES (?, ?, ?)", id, userIP, userAgent)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(business)
}

func getMyBusinessesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	userID := r.Header.Get("X-User-ID")
	ownerID, err := strconv.Atoi(userID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid user ID"})
		return
	}

	rows, err := db.Query(`
				SELECT id, name, category, description, phone, email, address,
					(SELECT image_url FROM images WHERE entity_type = 'business' AND entity_id = businesses.id ORDER BY is_primary DESC, display_order ASC, created_at ASC LIMIT 1) as image_url,
					rating, created_at, owner_id
				FROM businesses
				WHERE owner_id = ?
				ORDER BY created_at DESC
		`, ownerID)
	if err != nil {
		log.Printf("Error querying user businesses: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "internal server error"})
		return
	}
	defer rows.Close()

	var businesses []Business
	for rows.Next() {
		var b Business
		var imageURL sql.NullString
		err := rows.Scan(&b.ID, &b.Name, &b.Category, &b.Description, &b.Phone, &b.Email, &b.Address, &imageURL, &b.Rating, &b.CreatedAt, &b.OwnerID)
		if err == nil && imageURL.Valid {
			b.ImageURL = imageURL.String
		}
		if err != nil {
			log.Printf("Error scanning business: %v", err)
			continue
		}
		businesses = append(businesses, b)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(businesses)
}

func getMyBusinessStatsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	userID := r.Header.Get("X-User-ID")
	ownerID, err := strconv.Atoi(userID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid user ID"})
		return
	}

	// Get business count
	var businessCount int
	err = db.QueryRow("SELECT COUNT(*) FROM businesses WHERE owner_id = ?", ownerID).Scan(&businessCount)
	if err != nil {
		log.Printf("Error counting businesses: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "internal server error"})
		return
	}

	// Get total views for all owner's businesses
	var totalViews int
	err = db.QueryRow("SELECT COUNT(*) FROM business_views WHERE business_id IN (SELECT id FROM businesses WHERE owner_id = ?)", ownerID).Scan(&totalViews)
	if err != nil {
		log.Printf("Error counting business views: %v", err)
		totalViews = 0 // Don't fail request if views table is unavailable
	}

	// Get average rating
	var avgRating sql.NullFloat64
	err = db.QueryRow("SELECT AVG(rating) FROM businesses WHERE owner_id = ? AND rating > 0", ownerID).Scan(&avgRating)
	if err != nil {
		log.Printf("Error calculating average rating: %v", err)
	}

	// Get views per business
	rows, err := db.Query(`
		SELECT b.id, b.name, COUNT(bv.id) as view_count
		FROM businesses b
		LEFT JOIN business_views bv ON b.id = bv.business_id
		WHERE b.owner_id = ?
		GROUP BY b.id, b.name
		ORDER BY view_count DESC
	`, ownerID)
	if err != nil {
		log.Printf("Error querying business views: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "internal server error"})
		return
	}
	defer rows.Close()

	var businessViews []map[string]interface{}
	for rows.Next() {
		var b map[string]interface{}
		var id int
		var name string
		var viewCount int
		err := rows.Scan(&id, &name, &viewCount)
		if err != nil {
			log.Printf("Error scanning business view: %v", err)
			continue
		}
		b = map[string]interface{}{
			"id":         id,
			"name":       name,
			"view_count": viewCount,
		}
		businessViews = append(businessViews, b)
	}

	resp := map[string]interface{}{
		"business_count": businessCount,
		"total_views":    totalViews,
		"average_rating": avgRating.Float64,
		"business_views": businessViews,
	}

	w.Header().Set("Content-Type", "application/json")
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

func logEvent(eventType, message string, data interface{}) {
	eventMutex.Lock()
	event := SystemEvent{
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

func systemEventsHandler(w http.ResponseWriter, _ *http.Request) {
	eventMutex.Lock()
	events := make([]SystemEvent, len(eventLog))
	copy(events, eventLog)
	eventMutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}

// Business Events Handlers

func businessEventsRouter(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// GET is public - no auth required
		getBusinessEventsHandler(w, r)
	case http.MethodPost:
		// POST requires event owner or business owner auth
		eventOwnerOnly(createBusinessEventHandler)(w, r)
	case http.MethodPut:
		// PUT requires event owner or business owner auth
		eventOwnerOnly(updateBusinessEventHandler)(w, r)
	case http.MethodDelete:
		// DELETE requires event owner or business owner auth
		eventOwnerOnly(deleteBusinessEventHandler)(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func getBusinessEventsHandler(w http.ResponseWriter, r *http.Request) {
	// Get optional business_id filter
	businessIDStr := r.URL.Query().Get("business_id")

	var rows *sql.Rows
	var err error

	if businessIDStr != "" {
		businessID, err := strconv.Atoi(businessIDStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid business ID"})
			return
		}
		rows, err = db.Query(`
			SELECT id, owner_id, business_id, title, description, event_date, location, price, category,
			  (SELECT image_url FROM images WHERE entity_type = 'event' AND entity_id = events.id ORDER BY is_primary DESC, display_order ASC, created_at ASC LIMIT 1) as image_url,
			  created_at
			FROM events
			WHERE business_id = ? AND event_date >= NOW()
			ORDER BY event_date ASC
		`, businessID)
	} else {
		rows, err = db.Query(`
			SELECT id, owner_id, business_id, title, description, event_date, location, price, category,
			  (SELECT image_url FROM images WHERE entity_type = 'event' AND entity_id = events.id ORDER BY is_primary DESC, display_order ASC, created_at ASC LIMIT 1) as image_url,
			  created_at
			FROM events
			WHERE event_date >= NOW()
			ORDER BY event_date ASC
		`)
	}

	if err != nil {
		log.Printf("Error querying events: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "internal server error"})
		return
	}
	defer rows.Close()

	var events []BusinessEvent
	for rows.Next() {
		var e BusinessEvent
		var businessID sql.NullInt64
		var imageURL sql.NullString
		err := rows.Scan(&e.ID, &e.OwnerID, &businessID, &e.Title, &e.Description, &e.EventDate, &e.Location, &e.Price, &e.Category, &imageURL, &e.CreatedAt)
		if err != nil {
			log.Printf("Error scanning event: %v", err)
			continue
		}
		if businessID.Valid {
			bid := int(businessID.Int64)
			e.BusinessID = &bid
		}
		if imageURL.Valid {
			e.ImageURL = imageURL.String
		}
		events = append(events, e)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}

func createBusinessEventHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	userID := r.Header.Get("X-User-ID")
	ownerID, err := strconv.Atoi(userID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid user ID"})
		return
	}

	userType := r.Header.Get("X-User-Type")

	var req struct {
		BusinessID  *int    `json:"business_id"`
		Title       string  `json:"title"`
		Description string  `json:"description"`
		EventDate   string  `json:"event_date"`
		Location    string  `json:"location"`
		Price       float64 `json:"price"`
		Category    string  `json:"category"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request"})
		return
	}

	if req.Title == "" || req.EventDate == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "title and event_date are required"})
		return
	}

	// If business_id is provided, verify ownership (only for business owners)
	if req.BusinessID != nil && *req.BusinessID > 0 {
		if userType == "business_owner" {
			var businessOwnerID int
			err = db.QueryRow("SELECT owner_id FROM businesses WHERE id = ?", *req.BusinessID).Scan(&businessOwnerID)
			if err != nil {
				if err == sql.ErrNoRows {
					w.WriteHeader(http.StatusNotFound)
					json.NewEncoder(w).Encode(map[string]string{"error": "business not found"})
					return
				}
				log.Printf("Error checking business ownership: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{"error": "internal server error"})
				return
			}

			if businessOwnerID != ownerID {
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(map[string]string{"error": "you can only create events for your own businesses"})
				return
			}
		} else {
			// Event owners can't link to businesses they don't own
			req.BusinessID = nil
		}
	}

	// Parse event date
	eventDate, err := time.Parse("2006-01-02T15:04", req.EventDate)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid event_date format, use YYYY-MM-DDTHH:MM"})
		return
	}

	result, err := db.Exec(`
		INSERT INTO events (owner_id, business_id, title, description, event_date, location, price, category)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, ownerID, req.BusinessID, req.Title, req.Description, eventDate, req.Location, req.Price, req.Category)

	if err != nil {
		log.Printf("Error creating event: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to create event"})
		return
	}

	id, _ := result.LastInsertId()
	event := BusinessEvent{
		ID:          int(id),
		OwnerID:     ownerID,
		BusinessID:  req.BusinessID,
		Title:       req.Title,
		Description: req.Description,
		EventDate:   eventDate,
		Location:    req.Location,
		Price:       req.Price,
		Category:    req.Category,
		CreatedAt:   time.Now(),
	}

	logEvent("event_created", "Event "+event.Title+" created", event)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(event)
}

func updateBusinessEventHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	userID := r.Header.Get("X-User-ID")
	ownerID, err := strconv.Atoi(userID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid user ID"})
		return
	}

	var req struct {
		ID          int     `json:"id"`
		Title       string  `json:"title"`
		Description string  `json:"description"`
		EventDate   string  `json:"event_date"`
		Location    string  `json:"location"`
		Price       float64 `json:"price"`
		Category    string  `json:"category"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request"})
		return
	}

	// Verify event belongs to owner
	var eventOwnerID int
	err = db.QueryRow("SELECT owner_id FROM events WHERE id = ?", req.ID).Scan(&eventOwnerID)

	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "event not found"})
			return
		}
		log.Printf("Error checking event ownership: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "internal server error"})
		return
	}

	if eventOwnerID != ownerID {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": "you can only update your own events"})
		return
	}

	// Build update query dynamically
	setParts := []string{}
	args := []interface{}{}

	if req.Title != "" {
		setParts = append(setParts, "title = ?")
		args = append(args, req.Title)
	}
	if req.Description != "" {
		setParts = append(setParts, "description = ?")
		args = append(args, req.Description)
	}
	if req.EventDate != "" {
		eventDate, err := time.Parse("2006-01-02T15:04", req.EventDate)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid event_date format"})
			return
		}
		setParts = append(setParts, "event_date = ?")
		args = append(args, eventDate)
	}
	if req.Location != "" {
		setParts = append(setParts, "location = ?")
		args = append(args, req.Location)
	}
	if req.Price >= 0 {
		setParts = append(setParts, "price = ?")
		args = append(args, req.Price)
	}
	if req.Category != "" {
		setParts = append(setParts, "category = ?")
		args = append(args, req.Category)
	}

	if len(setParts) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "no valid fields to update"})
		return
	}

	query := "UPDATE events SET " + setParts[0]
	for i := 1; i < len(setParts); i++ {
		query += ", " + setParts[i]
	}
	query += " WHERE id = ?"
	args = append(args, req.ID)

	_, err = db.Exec(query, args...)
	if err != nil {
		log.Printf("Error updating event: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to update event"})
		return
	}

	// Get updated event
	var event BusinessEvent
	var businessID sql.NullInt64
	err = db.QueryRow(`
		SELECT id, owner_id, business_id, title, description, event_date, location, price, category, created_at
		FROM events
		WHERE id = ?
	`, req.ID).Scan(&event.ID, &event.OwnerID, &businessID, &event.Title, &event.Description, &event.EventDate, &event.Location, &event.Price, &event.Category, &event.CreatedAt)

	if businessID.Valid {
		bid := int(businessID.Int64)
		event.BusinessID = &bid
	}

	if err != nil {
		log.Printf("Error fetching updated event: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to fetch updated event"})
		return
	}

	logEvent("event_updated", "Event "+event.Title+" updated", event)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(event)
}

func deleteBusinessEventHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	userID := r.Header.Get("X-User-ID")
	ownerID, err := strconv.Atoi(userID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid user ID"})
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

	// Get event before deletion for logging and verification
	var event BusinessEvent
	err = db.QueryRow("SELECT id, title, owner_id FROM events WHERE id = ?", req.ID).Scan(&event.ID, &event.Title, &event.OwnerID)

	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "event not found"})
			return
		}
		log.Printf("Error fetching event for deletion: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "internal server error"})
		return
	}

	if event.OwnerID != ownerID {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": "you can only delete your own events"})
		return
	}

	_, err = db.Exec("DELETE FROM events WHERE id = ?", req.ID)
	if err != nil {
		log.Printf("Error deleting event: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to delete event"})
		return
	}

	logEvent("event_deleted", "Event "+event.Title+" deleted", event)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "event deleted successfully"})
}

func getEventByIDHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/event/")
	if idStr == r.URL.Path {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "event ID required"})
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid event ID"})
		return
	}

	var event BusinessEvent
	var businessID sql.NullInt64
	err = db.QueryRow(`
		SELECT id, owner_id, business_id, title, description, event_date, location, price, category, created_at
		FROM events
		WHERE id = ?
	`, id).Scan(&event.ID, &event.OwnerID, &businessID, &event.Title, &event.Description, &event.EventDate, &event.Location, &event.Price, &event.Category, &event.CreatedAt)

	if businessID.Valid {
		bid := int(businessID.Int64)
		event.BusinessID = &bid
	}

	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "event not found"})
			return
		}
		log.Printf("Error fetching event: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "internal server error"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(event)
}

func getMyEventsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	userID := r.Header.Get("X-User-ID")
	ownerID, err := strconv.Atoi(userID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid user ID"})
		return
	}

	rows, err := db.Query(`
		SELECT id, owner_id, business_id, title, description, event_date, location, price, category,
		  (SELECT image_url FROM images WHERE entity_type = 'event' AND entity_id = events.id ORDER BY is_primary DESC, display_order ASC, created_at ASC LIMIT 1) as image_url,
		  created_at
		FROM events
		WHERE owner_id = ?
		ORDER BY event_date ASC
	`, ownerID)

	if err != nil {
		log.Printf("Error querying user events: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "internal server error"})
		return
	}
	defer rows.Close()

	var events []BusinessEvent
	for rows.Next() {
		var e BusinessEvent
		var businessID sql.NullInt64
		var imageURL sql.NullString
		err := rows.Scan(&e.ID, &e.OwnerID, &businessID, &e.Title, &e.Description, &e.EventDate, &e.Location, &e.Price, &e.Category, &imageURL, &e.CreatedAt)
		if err != nil {
			log.Printf("Error scanning event: %v", err)
			continue
		}
		if businessID.Valid {
			bid := int(businessID.Int64)
			e.BusinessID = &bid
		}
		if imageURL.Valid {
			e.ImageURL = imageURL.String
		}
		events = append(events, e)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}

// Booking Handlers

func bookingsRouter(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// GET requires auth to view bookings
		authMiddleware(getBookingsHandler)(w, r)
	case http.MethodPost:
		// POST is public - anyone can book
		createBookingHandler(w, r)
	case http.MethodPut:
		// PUT requires auth to update booking status
		authMiddleware(updateBookingHandler)(w, r)
	case http.MethodDelete:
		// DELETE requires auth
		authMiddleware(deleteBookingHandler)(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func createBookingHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		EventID int    `json:"event_id"`
		Name    string `json:"name"`
		Email   string `json:"email"`
		Phone   string `json:"phone"`
		Tickets int    `json:"tickets"`
		Notes   string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request"})
		return
	}

	if req.EventID == 0 || req.Name == "" || req.Email == "" || req.Tickets < 1 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "event_id, name, email, and tickets are required"})
		return
	}

	// Verify event exists
	var eventID int
	err := db.QueryRow("SELECT id FROM events WHERE id = ?", req.EventID).Scan(&eventID)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "event not found"})
			return
		}
		log.Printf("Error checking event: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "internal server error"})
		return
	}

	result, err := db.Exec(`
		INSERT INTO bookings (event_id, name, email, phone, tickets, notes, status)
		VALUES (?, ?, ?, ?, ?, ?, 'pending')
	`, req.EventID, req.Name, req.Email, req.Phone, req.Tickets, req.Notes)

	if err != nil {
		log.Printf("Error creating booking: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to create booking"})
		return
	}

	id, _ := result.LastInsertId()
	booking := Booking{
		ID:        int(id),
		EventID:   req.EventID,
		Name:      req.Name,
		Email:     req.Email,
		Phone:     req.Phone,
		Tickets:   req.Tickets,
		Notes:     req.Notes,
		Status:    "pending",
		CreatedAt: time.Now(),
	}

	logEvent("booking_created", fmt.Sprintf("Booking created for event %d by %s", req.EventID, req.Name), booking)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"booking": booking,
		"message": "Booking created successfully",
	})
}

func getBookingsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	userID := r.Header.Get("X-User-ID")
	ownerID, err := strconv.Atoi(userID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid user ID"})
		return
	}

	// Get bookings for events owned by this user
	rows, err := db.Query(`
		SELECT b.id, b.event_id, b.name, b.email, b.phone, b.tickets, b.notes, b.status, b.created_at
		FROM bookings b
		INNER JOIN events e ON b.event_id = e.id
		WHERE e.owner_id = ?
		ORDER BY b.created_at DESC
	`, ownerID)

	if err != nil {
		log.Printf("Error querying bookings: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "internal server error"})
		return
	}
	defer rows.Close()

	var bookings []Booking
	for rows.Next() {
		var b Booking
		err := rows.Scan(&b.ID, &b.EventID, &b.Name, &b.Email, &b.Phone, &b.Tickets, &b.Notes, &b.Status, &b.CreatedAt)
		if err != nil {
			log.Printf("Error scanning booking: %v", err)
			continue
		}
		bookings = append(bookings, b)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bookings)
}

func updateBookingHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	userID := r.Header.Get("X-User-ID")
	ownerID, err := strconv.Atoi(userID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid user ID"})
		return
	}

	var req struct {
		ID     int    `json:"id"`
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request"})
		return
	}

	// Verify booking belongs to user's event
	var eventOwnerID int
	err = db.QueryRow(`
		SELECT e.owner_id
		FROM bookings b
		INNER JOIN events e ON b.event_id = e.id
		WHERE b.id = ?
	`, req.ID).Scan(&eventOwnerID)

	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "booking not found"})
			return
		}
		log.Printf("Error checking booking ownership: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "internal server error"})
		return
	}

	if eventOwnerID != ownerID {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": "you can only update bookings for your own events"})
		return
	}

	// Update booking status
	_, err = db.Exec("UPDATE bookings SET status = ? WHERE id = ?", req.Status, req.ID)
	if err != nil {
		log.Printf("Error updating booking: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to update booking"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "booking updated successfully"})
}

func deleteBookingHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	userID := r.Header.Get("X-User-ID")
	ownerID, err := strconv.Atoi(userID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid user ID"})
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

	// Verify booking belongs to user's event
	var eventOwnerID int
	err = db.QueryRow(`
		SELECT e.owner_id
		FROM bookings b
		INNER JOIN events e ON b.event_id = e.id
		WHERE b.id = ?
	`, req.ID).Scan(&eventOwnerID)

	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "booking not found"})
			return
		}
		log.Printf("Error checking booking ownership: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "internal server error"})
		return
	}

	if eventOwnerID != ownerID {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": "you can only delete bookings for your own events"})
		return
	}

	_, err = db.Exec("DELETE FROM bookings WHERE id = ?", req.ID)
	if err != nil {
		log.Printf("Error deleting booking: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to delete booking"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "booking deleted successfully"})
}
