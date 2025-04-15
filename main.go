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
	"time"

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
	IsAdmin  bool   `json:"is_admin"`
}

// Структура для хранения информации о файле
type FileInfo struct {
	Name    string `json:"name"`    // Имя файла
	Owner   string `json:"owner"`   // Владелец файла
	Created string `json:"created"` // Дата создания
}

type UserStore interface {
	LoadUsers() error
	ValidateCredentials(username, password string) bool
	AddUser(username, password string) error
}

type JSONUserStore struct {
	Users []User `json:"users"`
	mu    sync.RWMutex
}

type PostgresUserStore struct{}

var (
	config    Config
	userStore UserStore
	templates = template.Must(template.New("main.html").Funcs(template.FuncMap{
		"lower": strings.ToLower,
	}).ParseFiles("templates/main.html", "templates/game.html"))
	db *sql.DB
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

func (store *JSONUserStore) AddUser(username, password string) error {
	store.mu.Lock()
	defer store.mu.Unlock()

	// Проверка на существующего пользователя
	for _, user := range store.Users {
		if user.Username == username {
			return fmt.Errorf("пользователь уже существует")
		}
	}

	// Добавление нового пользователя
	store.Users = append(store.Users, User{
		Username: username,
		Password: password,
		IsAdmin:  false,
	})

	// Сохранение в файл с форматированием
	file, err := os.Create(config.UserFile)
	if err != nil {
		return err
	}
	defer file.Close()

	// Создаем encoder с отступами
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "    ")

	return encoder.Encode(store)
}

func (store *PostgresUserStore) AddUser(username, password string) error {
	_, err := db.Exec("INSERT INTO users (username, password) VALUES ($1, $2)", username, password)
	return err
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
	http.HandleFunc("/upload", uploadHandler)
	http.HandleFunc("/files", filesHandler)
	http.HandleFunc("/delete", deleteHandler)
	http.HandleFunc("/download", downloadHandler)
	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/openFile", openFileHandler)
	http.HandleFunc("/saveFile", saveFileHandler)
	http.HandleFunc("/game", gamePageHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Start server
	address := fmt.Sprintf("%s:%s", config.ServerHost, config.ServerPort)
	log.Printf("Starting server on %s", address)
	if err := http.ListenAndServe(address, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
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

	// Получаем информацию о файлах
	fileInfos, err := getFileInfos()
	if err != nil {
		http.Error(w, "Failed to get file info", http.StatusInternalServerError)
		return
	}

	// Получаем информацию о текущем пользователе
	var isAdmin bool
	var currentUsername string
	if cookie, err := r.Cookie("auth"); err == nil {
		currentUsername = strings.Split(cookie.Value, ":")[0]
		for _, user := range userStore.(*JSONUserStore).Users {
			if user.Username == currentUsername {
				isAdmin = user.IsAdmin
				break
			}
		}
	}

	// Рендеринг страницы
	data := struct {
		IsAuthorized bool
		Files        []FileInfo
		Username     string
		IsAdmin      bool
	}{
		IsAuthorized: isAuthorized,
		Files:        fileInfos,
		Username:     currentUsername,
		IsAdmin:      isAdmin,
	}

	err = templates.ExecuteTemplate(w, "main.html", data)
	if err != nil {
		http.Error(w, "Failed to render page", http.StatusInternalServerError)
	}
}

// Обработчик загрузки файлов
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	// Проверяем метод запроса
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Получаем имя пользователя из cookie
	cookie, _ := r.Cookie("auth")
	username := strings.Split(cookie.Value, ":")[0]

	// Получаем файл из формы
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to read file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Создаем путь для сохранения файла
	filePath := filepath.Join(config.FileDir, header.Filename)
	out, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}
	defer out.Close()

	// Копируем содержимое файла
	if _, err := io.Copy(out, file); err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}

	// Создаем информацию о файле
	fileInfo := FileInfo{
		Name:    header.Filename,
		Owner:   username,
		Created: time.Now().Format(time.RFC3339),
	}

	// Сохраняем информацию о файле
	saveFileInfo(fileInfo)

	http.Redirect(w, r, "/", http.StatusFound)
}

// Функция для сохранения информации о файле
func saveFileInfo(info FileInfo) error {
	filePath := filepath.Join(config.FileDir, ".fileinfo.json")
	var fileInfos []FileInfo

	// Читаем существующие данные
	data, err := os.ReadFile(filePath)
	if err == nil {
		json.Unmarshal(data, &fileInfos)
	}

	// Добавляем новую информацию
	fileInfos = append(fileInfos, info)

	// Сохраняем обновленные данные
	data, err = json.Marshal(fileInfos)
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0644)
}

// Функция для получения информации о файлах
func getFileInfos() ([]FileInfo, error) {
	filePath := filepath.Join(config.FileDir, ".fileinfo.json")
	var fileInfos []FileInfo

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []FileInfo{}, nil
		}
		return nil, err
	}

	err = json.Unmarshal(data, &fileInfos)
	return fileInfos, err
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

	// Получаем информацию о файлах из .fileinfo.json
	fileInfos, err := getFileInfos()
	if err != nil {
		http.Error(w, "Failed to get file info", http.StatusInternalServerError)
		return
	}

	// Получаем имя текущего пользователя из cookie
	cookie, _ := r.Cookie("auth")
	username := strings.Split(cookie.Value, ":")[0]

	// Проверяем, является ли пользователь администратором
	var isAdmin bool
	for _, user := range userStore.(*JSONUserStore).Users {
		if user.Username == username {
			isAdmin = user.IsAdmin
			break
		}
	}

	// Формируем ответ с информацией о правах доступа
	type FileResponse struct {
		FileInfo
		CanDelete bool `json:"can_delete"`
	}

	response := make([]FileResponse, 0)
	for _, info := range fileInfos {
		response = append(response, FileResponse{
			FileInfo:  info,
			CanDelete: isAdmin || info.Owner == username,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// File delete handler
func deleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Получаем имя пользователя из cookie
	cookie, _ := r.Cookie("auth")
	username := strings.Split(cookie.Value, ":")[0]

	fileName := r.FormValue("filename")

	// Проверяем права на удаление
	fileInfos, err := getFileInfos()
	if err != nil {
		http.Error(w, "Failed to get file info", http.StatusInternalServerError)
		return
	}

	var isOwner bool
	var isAdmin bool

	// Проверяем, является ли пользователь владельцем файла или админом
	for _, user := range userStore.(*JSONUserStore).Users {
		if user.Username == username {
			isAdmin = user.IsAdmin
			break
		}
	}

	for _, info := range fileInfos {
		if info.Name == fileName {
			isOwner = info.Owner == username
			break
		}
	}

	if !isOwner && !isAdmin {
		http.Error(w, "Permission denied", http.StatusForbidden)
		return
	}

	filePath := filepath.Join(config.FileDir, fileName)
	if err := os.Remove(filePath); err != nil {
		http.Error(w, "Failed to delete file", http.StatusInternalServerError)
		return
	}

	// Обновляем информацию о файлах
	updateFileInfos(fileName)

	http.Redirect(w, r, "/", http.StatusFound)
}

// Функция для обновления информации о файлах после удаления
func updateFileInfos(deletedFileName string) error {
	filePath := filepath.Join(config.FileDir, ".fileinfo.json")
	var fileInfos []FileInfo

	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	json.Unmarshal(data, &fileInfos)

	// Удаляем информацию об удаленном файле
	var updatedInfos []FileInfo
	for _, info := range fileInfos {
		if info.Name != deletedFileName {
			updatedInfos = append(updatedInfos, info)
		}
	}

	data, err = json.Marshal(updatedInfos)
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0644)
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

func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не разрешен", http.StatusMethodNotAllowed)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	if username == "" || password == "" {
		http.Error(w, "Имя пользователя и пароль обязательны", http.StatusBadRequest)
		return
	}

	err := userStore.AddUser(username, password)
	if err != nil {
		http.Error(w, "Ошибка при регистрации: "+err.Error(), http.StatusBadRequest)
		return
	}

	// После успешной регистрации сразу авторизуем пользователя
	http.SetCookie(w, &http.Cookie{
		Name:  "auth",
		Value: username + ":" + password,
		Path:  "/",
	})

	http.Redirect(w, r, "/", http.StatusFound)
}

// Добавим новый обработчик для открытия файла
func openFileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	fileName := r.URL.Query().Get("filename")
	filePath := filepath.Join(config.FileDir, fileName)

	// Проверяем права доступа
	cookie, _ := r.Cookie("auth")
	username := strings.Split(cookie.Value, ":")[0]

	fileInfos, err := getFileInfos()
	if err != nil {
		http.Error(w, "Failed to get file info", http.StatusInternalServerError)
		return
	}

	var isOwner bool
	var isAdmin bool

	for _, user := range userStore.(*JSONUserStore).Users {
		if user.Username == username {
			isAdmin = user.IsAdmin
			break
		}
	}

	for _, info := range fileInfos {
		if info.Name == fileName {
			isOwner = info.Owner == username
			break
		}
	}

	if !isOwner && !isAdmin {
		http.Error(w, "Permission denied", http.StatusForbidden)
		return
	}

	// Читаем содержимое файла
	content, err := os.ReadFile(filePath)
	if err != nil {
		http.Error(w, "Failed to read file", http.StatusInternalServerError)
		return
	}

	// Проверяем ограничения
	lines := strings.Split(string(content), "\n")
	if len(lines) > 4294967296 {
		http.Error(w, "File has too many lines", http.StatusBadRequest)
		return
	}

	for _, line := range lines {
		if len(line) > 4294967296 {
			http.Error(w, "File contains lines that are too long", http.StatusBadRequest)
			return
		}
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write(content)
}

// Добавим обработчик для сохранения файла
func saveFileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	fileName := r.URL.Query().Get("filename")
	filePath := filepath.Join(config.FileDir, fileName)

	// Проверяем права доступа
	cookie, _ := r.Cookie("auth")
	username := strings.Split(cookie.Value, ":")[0]

	fileInfos, err := getFileInfos()
	if err != nil {
		http.Error(w, "Failed to get file info", http.StatusInternalServerError)
		return
	}

	var isOwner bool
	var isAdmin bool

	for _, user := range userStore.(*JSONUserStore).Users {
		if user.Username == username {
			isAdmin = user.IsAdmin
			break
		}
	}

	for _, info := range fileInfos {
		if info.Name == fileName {
			isOwner = info.Owner == username
			break
		}
	}

	if !isOwner && !isAdmin {
		http.Error(w, "Permission denied", http.StatusForbidden)
		return
	}

	// Читаем новое содержимое
	content, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}

	// Проверяем ограничения
	lines := strings.Split(string(content), "\n")
	if len(lines) > 4294967296 {
		http.Error(w, "File has too many lines", http.StatusBadRequest)
		return
	}

	for _, line := range lines {
		if len(line) > 4294967296 {
			http.Error(w, "File contains lines that are too long", http.StatusBadRequest)
			return
		}
	}

	// Сохраняем файл
	err = os.WriteFile(filePath, content, 0644)
	if err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// New handler for the game page
func gamePageHandler(w http.ResponseWriter, r *http.Request) {
	// Render the game page template
	err := templates.ExecuteTemplate(w, "game.html", nil)
	if err != nil {
		http.Error(w, "Failed to render game page", http.StatusInternalServerError)
	}
}
