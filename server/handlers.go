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
	todosMutex   sync.Mutex
	todos        = make(map[int]Todo)
	nextTodoID   = 1
	usersMutex   sync.Mutex
	users        = make(map[int]User)
	nextUserID   = 1
)

type Todo struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Completed bool      `json:"completed"`
	CreatedAt time.Time `json:"created_at"`
}

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
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
	mux.HandleFunc("/hello", corsMiddleware(helloHandler))
	mux.HandleFunc("/time", corsMiddleware(timeHandler))
	mux.HandleFunc("/echo", corsMiddleware(echoHandler))
	mux.HandleFunc("/stats", corsMiddleware(statsHandler))
	mux.HandleFunc("/todos", corsMiddleware(todosRouter))
	mux.HandleFunc("/users", corsMiddleware(usersRouter))
	return mux
}

func todosRouter(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getTodosHandler(w, r)
	case http.MethodPost:
		createTodoHandler(w, r)
	case http.MethodPut:
		updateTodoHandler(w, r)
	case http.MethodDelete:
		deleteTodoHandler(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func usersRouter(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getUsersHandler(w, r)
	case http.MethodPost:
		createUserHandler(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
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

func getTodosHandler(w http.ResponseWriter, r *http.Request) {
	todosMutex.Lock()
	todoList := make([]Todo, 0, len(todos))
	for _, todo := range todos {
		todoList = append(todoList, todo)
	}
	todosMutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todoList)
}

func createTodoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Title string `json:"title"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request"})
		return
	}

	todosMutex.Lock()
	todo := Todo{
		ID:        nextTodoID,
		Title:     req.Title,
		Completed: false,
		CreatedAt: time.Now(),
	}
	todos[nextTodoID] = todo
	nextTodoID++
	todosMutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(todo)
}

func updateTodoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ID        int    `json:"id"`
		Title     string `json:"title"`
		Completed bool   `json:"completed"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	todosMutex.Lock()
	if todo, ok := todos[req.ID]; ok {
		if req.Title != "" {
			todo.Title = req.Title
		}
		todo.Completed = req.Completed
		todos[req.ID] = todo
	}
	todo := todos[req.ID]
	todosMutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todo)
}

func deleteTodoHandler(w http.ResponseWriter, r *http.Request) {
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

	todosMutex.Lock()
	delete(todos, req.ID)
	todosMutex.Unlock()

	w.WriteHeader(http.StatusOK)
}

func getUsersHandler(w http.ResponseWriter, r *http.Request) {
	usersMutex.Lock()
	userList := make([]User, 0, len(users))
	for _, user := range users {
		userList = append(userList, user)
	}
	usersMutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userList)
}

func createUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request"})
		return
	}

	usersMutex.Lock()
	user := User{
		ID:    nextUserID,
		Name:  req.Name,
		Email: req.Email,
	}
	users[nextUserID] = user
	nextUserID++
	usersMutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}
