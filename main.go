package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	_ "github.com/lib/pq"
)

type Config struct {
	ServerHost  string `json:"server_host"`
	ServerPort  string `json:"server_port"`
	FileDir     string `json:"file_dir"`
	UserFile    string `json:"user_file"`
	StorageMode string `json:"storage_mode"`
	DBHost      string `json:"db_host"`
	DBPort      string `json:"db_port"`
	DBUser      string `json:"db_user"`
	DBPassword  string `json:"db_password"`
	DBName      string `json:"db_name"`
}

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type UserStore interface {
	LoadUsers() error
	ValidateCredentials(username, password string) bool
}

type JSONUserStore struct {
	Users []User `json:"users"`
	mu    sync.RWMutex
}

type PostgresUserStore struct{}

var (
	config    Config
	userStore UserStore
	templates = template.Must(template.ParseFiles("templates/main.html"))
	db        *sql.DB
)

func initPostgres() {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.DBHost, config.DBPort, config.DBUser, config.DBPassword, config.DBName)

	var err error
	db, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("PostgreSQL connection failed: %v", err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		username VARCHAR(255) UNIQUE NOT NULL,
		password VARCHAR(255) NOT NULL
	)`)
	if err != nil {
		log.Fatalf("Failed to create users table: %v", err)
	}
}

func (store *JSONUserStore) LoadUsers() error {
	file, err := os.Open(config.UserFile)
	if err != nil {
		return err
	}
	defer file.Close()

	store.mu.Lock()
	defer store.mu.Unlock()

	return json.NewDecoder(file).Decode(&store)
}

func (store *JSONUserStore) ValidateCredentials(username, password string) bool {
	store.mu.RLock()
	defer store.mu.RUnlock()

	for _, user := range store.Users {
		if user.Username == username && user.Password == password {
			return true
		}
	}
	return false
}

func (store *PostgresUserStore) LoadUsers() error {
	return nil
}

func (store *PostgresUserStore) ValidateCredentials(username, password string) bool {
	var dbPassword string
	err := db.QueryRow("SELECT password FROM users WHERE username = $1", username).Scan(&dbPassword)
	if err != nil {
		return false
	}
	return dbPassword == password
}

func main() {
	// Load config
	configFile, err := os.Open("config.json")
	if err != nil {
		log.Fatalf("Failed to load config.json: %v", err)
	}
	defer configFile.Close()

	if err := json.NewDecoder(configFile).Decode(&config); err != nil {
		log.Fatalf("Failed to parse config.json: %v", err)
	}

	// Load users from file
	switch config.StorageMode {
	case "json":
		jsonStore := &JSONUserStore{}
		if err := jsonStore.LoadUsers(); err != nil {
			log.Fatalf("Failed to load users from JSON: %v", err)
		}
		userStore = jsonStore
	case "postgresql":
		initPostgres()
		userStore = &PostgresUserStore{}
	default:
		log.Fatalf("Invalid storage mode: %s", config.StorageMode)
	}

	// Ensure file directory exists
	if _, err := os.Stat(config.FileDir); os.IsNotExist(err) {
		if err := os.Mkdir(config.FileDir, os.ModePerm); err != nil {
			log.Fatalf("Failed to create files directory: %v", err)
		}
	}

	// HTTP routes
	http.HandleFunc("/", mainPageHandler)
	http.HandleFunc("/upload", authMiddleware(uploadHandler))
	http.HandleFunc("/files", authMiddleware(filesHandler))
	http.HandleFunc("/delete", authMiddleware(deleteHandler))
	http.HandleFunc("/download", authMiddleware(downloadHandler))
	http.HandleFunc("/logout", logoutHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Start server
	address := fmt.Sprintf("%s:%s", config.ServerHost, config.ServerPort)
	log.Printf("Starting server on %s", address)
	if err := http.ListenAndServe(address, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

// Load users from JSON file
func loadUsers() {
	file, err := os.Open(config.UserFile)
	if err != nil {
		log.Fatalf("Failed to open user file: %v", err)
	}
	defer file.Close()

	if err := json.NewDecoder(file).Decode(&userStore); err != nil {
		log.Fatalf("Failed to parse user file: %v", err)
	}
}

// Authentication middleware
func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("auth")
		if err != nil {
			log.Printf("Unauthorized access attempt: %v", r.RemoteAddr)
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		authParts := strings.Split(cookie.Value, ":")
		if len(authParts) != 2 || !userStore.ValidateCredentials(authParts[0], authParts[1]) {
			log.Printf("Invalid cookie or credentials: %v", r.RemoteAddr)
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		// Pass to the next handler if authorized
		next(w, r)
	}
}

// Main page handler (login + dashboard)
func mainPageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		password := r.FormValue("password")

		if userStore.ValidateCredentials(username, password) {
			// Установка cookie
			http.SetCookie(w, &http.Cookie{
				Name:  "auth",
				Value: username + ":" + password,
				Path:  "/",
			})

			// Перенаправление на главную страницу
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		// Ошибка авторизации
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Проверка авторизации через cookie
	cookie, err := r.Cookie("auth")
	var isAuthorized bool
	if err == nil {
		authParts := strings.Split(cookie.Value, ":")
		if len(authParts) == 2 {
			isAuthorized = userStore.ValidateCredentials(authParts[0], authParts[1])
		}
	}

	// Список файлов для авторизованных пользователей
	var fileNames []string
	if isAuthorized {
		files, _ := os.ReadDir(config.FileDir)
		for _, file := range files {
			fileNames = append(fileNames, file.Name())
		}
	}

	// Рендеринг страницы
	data := struct {
		IsAuthorized bool
		Files        []string
	}{
		IsAuthorized: isAuthorized,
		Files:        fileNames,
	}

	err = templates.ExecuteTemplate(w, "main.html", data)
	if err != nil {
		http.Error(w, "Failed to render page", http.StatusInternalServerError)
	}
}

// Upload handler
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to read file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	filePath := filepath.Join(config.FileDir, header.Filename)
	out, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

// File listing handler
func fileListHandler(w http.ResponseWriter, r *http.Request) {
	// Logic for listing files
}

func filesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	files, err := os.ReadDir(config.FileDir)
	if err != nil {
		http.Error(w, "Failed to list files", http.StatusInternalServerError)
		return
	}

	fileNames := []string{}
	for _, file := range files {
		fileNames = append(fileNames, file.Name())
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(fileNames)
}

// File delete handler
func deleteHandler(w http.ResponseWriter, r *http.Request) {
	// Logic for deleting files
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	fileName := r.FormValue("filename")
	filePath := filepath.Join(config.FileDir, fileName)

	// Попытка удалить файл
	if err := os.Remove(filePath); err != nil {
		http.Error(w, "Failed to delete file", http.StatusInternalServerError)
		return
	}

	// Перенаправление на главную страницу после успешного удаления
	http.Redirect(w, r, "/", http.StatusFound)
}

// File download handler
func downloadHandler(w http.ResponseWriter, r *http.Request) {
	fileName := r.URL.Query().Get("filename")
	filePath := filepath.Join(config.FileDir, fileName)

	file, err := os.Open(filePath)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	defer file.Close()

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileName))
	w.Header().Set("Content-Type", "application/octet-stream")
	if _, err := io.Copy(w, file); err != nil {
		http.Error(w, "Failed to download file", http.StatusInternalServerError)
	}
}
func logoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:   "auth",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
	http.Redirect(w, r, "/", http.StatusFound)
}
