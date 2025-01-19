### Итоговый файл `README.md`

---

#### **На русском языке**

# File Management Server

Простой сервер для загрузки, скачивания, удаления и управления файлами через веб-интерфейс. Сервер поддерживает авторизацию, может хранить данные пользователей в формате JSON или в базе данных PostgreSQL. Выбор хранилища настраивается в конфигурационном файле.

---

## Возможности
- Авторизация пользователей.
- Загрузка файлов на сервер.
- Скачивание файлов с сервера.
- Удаление файлов с сервера.
- Выбор типа хранилища данных пользователей (JSON или PostgreSQL).
- Отображение прогресса загрузки файла.

---

## Сферы применения
- Личные серверы для управления файлами.
- Внутренние системы обмена файлами в компании.
- Хранилища данных для небольших команд.

---

## Требования
1. Golang (v1.19+).
2. PostgreSQL (если используется как хранилище пользователей).
3. Современный браузер для доступа к веб-интерфейсу.

---

## Структура проекта

```plaintext
/project
├── main.go              # Основной файл сервера.
├── config.json          # Файл конфигурации проекта.
├── users.json           # Хранилище пользователей (для режима JSON).
├── templates/
│   └── main.html        # HTML-шаблон веб-интерфейса.
├── static/
│   └── styles.css       # Стили для интерфейса.
├── files/               # Директория для хранения загруженных файлов.
```

### Описание файлов
- **`main.go`**: Основной файл сервера. Реализует все маршруты, обработчики и логику работы.
- **`config.json`**: Конфигурационный файл. Настраивает сервер (хост, порт, режим работы, подключение к PostgreSQL).
- **`users.json`**: JSON-файл для хранения пользователей (если выбрано хранилище JSON).
- **`templates/main.html`**: Шаблон для веб-интерфейса.
- **`static/styles.css`**: Стили для веб-интерфейса.
- **`files/`**: Папка для хранения загруженных файлов.

---

## Установка и настройка

### Шаг 1: Установка PostgreSQL
1. Установите PostgreSQL:
   ```bash
   sudo apt update
   sudo apt install postgresql postgresql-contrib
   ```
2. Запустите PostgreSQL:
   ```bash
   sudo systemctl start postgresql
   ```

3. Создайте пользователя и базу данных:
   ```bash
   sudo -u postgres psql
   CREATE DATABASE file_server;
   CREATE USER file_user WITH PASSWORD 'password';
   GRANT ALL PRIVILEGES ON DATABASE file_server TO file_user;
   ```

### Шаг 2: Настройка проекта
1. Клонируйте репозиторий:
   ```bash
   git clone https://github.com/your-repo/file-management-server.git
   cd file-management-server
   ```

2. Настройте файл `config.json`:
   ```json
   {
       "server_host": "localhost",
       "server_port": "8080",
       "file_dir": "./files",
       "user_file": "./users.json",
       "storage_mode": "postgresql",
       "db_host": "localhost",
       "db_port": "5432",
       "db_user": "file_user",
       "db_password": "password",
       "db_name": "file_server"
   }
   ```

3. Установите зависимости PostgreSQL для Go:
   ```bash
   go get github.com/lib/pq
   ```

### Шаг 3: Запуск сервера
Соберите и запустите проект:
```bash
go run main.go
```

Сервер будет доступен по адресу: [http://localhost:8080](http://localhost:8080).

---

## Описание работы функций

### Функции в `main.go`

1. **`main()`**:
   - Инициализирует конфигурацию и подключение к базе данных (если выбрана PostgreSQL).
   - Устанавливает маршруты и запускает сервер.

2. **`initPostgres()`**:
   - Подключается к PostgreSQL и создаёт таблицу пользователей, если она отсутствует.

3. **`authMiddleware()`**:
   - Проверяет авторизацию через cookie и перенаправляет неавторизованных пользователей на страницу входа.

4. **`uploadHandler()`**:
   - Обрабатывает загрузку файлов и сохраняет их в директорию `file_dir`.

5. **`filesHandler()`**:
   - Возвращает список файлов в формате JSON.

6. **`deleteHandler()`**:
   - Удаляет выбранный файл с сервера.

7. **`downloadHandler()`**:
   - Возвращает файл для скачивания.

8. **`logoutHandler()`**:
   - Удаляет cookie авторизации и перенаправляет на страницу входа.

---

### JavaScript в `main.html`

1. **Функция для загрузки файла**:
   - Отправляет файл на сервер с отображением прогресса загрузки.

2. **Функция для обновления списка файлов**:
   - Получает обновлённый список файлов с сервера после загрузки.

---

#### **In English**

# File Management Server

A simple server for uploading, downloading, deleting, and managing files through a web interface. The server supports user authentication and allows storing user data in JSON or PostgreSQL, configurable via a settings file.

---

## Features
- User authentication.
- File upload to the server.
- File download from the server.
- File deletion from the server.
- Selectable storage mode for user data (JSON or PostgreSQL).
- File upload progress bar.

---

## Usage Areas
- Personal file management servers.
- Internal file-sharing systems within a company.
- Small team data storage.

---

## Project Structure

```plaintext
/project
├── main.go              # Main server file.
├── config.json          # Project configuration file.
├── users.json           # User storage file (for JSON mode).
├── templates/
│   └── main.html        # HTML template for the web interface.
├── static/
│   └── styles.css       # Styles for the interface.
├── files/               # Directory for storing uploaded files.
```

### File Descriptions
- **`main.go`**: The main server file implementing routes, handlers, and logic.
- **`config.json`**: Configures the server (host, port, storage mode, PostgreSQL connection).
- **`users.json`**: JSON file for storing users (if JSON storage is selected).
- **`templates/main.html`**: Web interface template.
- **`static/styles.css`**: Styles for the web interface.
- **`files/`**: Directory for storing uploaded files.

---

## Installation and Setup

### Step 1: PostgreSQL Installation
1. Install PostgreSQL:
   ```bash
   sudo apt update
   sudo apt install postgresql postgresql-contrib
   ```
2. Start PostgreSQL:
   ```bash
   sudo systemctl start postgresql
   ```

3. Create a user and database:
   ```bash
   sudo -u postgres psql
   CREATE DATABASE file_server;
   CREATE USER file_user WITH PASSWORD 'password';
   GRANT ALL PRIVILEGES ON DATABASE file_server TO file_user;
   ```

### Step 2: Project Setup
1. Clone the repository:
   ```bash
   git clone https://github.com/your-repo/file-management-server.git
   cd file-management-server
   ```

2. Configure `config.json`:
   ```json
   {
       "server_host": "localhost",
       "server_port": "8080",
       "file_dir": "./files",
       "user_file": "./users.json",
       "storage_mode": "postgresql",
       "db_host": "localhost",
       "db_port": "5432",
       "db_user": "file_user",
       "db_password": "password",
       "db_name": "file_server"
   }
   ```

3. Install PostgreSQL dependencies for Go:
   ```bash
   go get github.com/lib/pq
   ```

### Step 3: Start the Server
Build and run the project:
```bash
go run main.go
```

The server will be accessible at: [http://localhost:8080](http://localhost:8080).

---

Feel free to ask if any clarifications or additional features are needed! 😊