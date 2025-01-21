# Принцип работы FileServer

## Структура проекта

```
fileserver/
├── main.go           # Основной файл сервера
├── config.json       # Конфигурация
├── users.json        # Данные пользователей
├── files/           # Директория с файлами
│   └── .fileinfo.json  # Метаданные файлов
├── templates/        # HTML шаблоны
│   └── main.html     # Основной шаблон
└── static/          # Статические файлы
    └── styles.css    # Стили
```

## Основные компоненты

### 1. Конфигурация (Config)
```go
type Config struct {
    ServerHost  string // Хост сервера
    ServerPort  string // Порт сервера
    FileDir     string // Директория для файлов
    UserFile    string // Файл с пользователями
    StorageMode string // Режим хранения (json/postgres)
    DBHost      string // Хост базы данных
    DBPort      string // Порт базы данных
    DBUser      string // Пользователь БД
    DBPassword  string // Пароль БД
    DBName      string // Имя БД
}
```

### 2. Структуры данных

#### Пользователь (User)
```go
type User struct {
    Username string // Имя пользователя
    Password string // Пароль
    IsAdmin  bool   // Права администратора
}
```

#### Информация о файле (FileInfo)
```go
type FileInfo struct {
    Name    string // Имя файла
    Owner   string // Владелец
    Created string // Дата создания
}
```

## Основные процессы

### 1. Инициализация сервера
1. Загрузка конфигурации из config.json
2. Инициализация хранилища пользователей (JSON/PostgreSQL)
3. Создание необходимых директорий
4. Запуск HTTP-сервера

### 2. Аутентификация

#### Процесс входа:
1. Получение данных формы
2. Проверка учетных данных
3. Создание сессии
4. Установка cookie

```go
func loginHandler(w http.ResponseWriter, r *http.Request) {
    // Проверка метода
    // Получение данных
    // Валидация
    // Установка cookie
}
```

### 3. Работа с файлами

#### Загрузка файла:
1. Проверка авторизации
2. Валидация файла
3. Сохранение файла
4. Обновление метаданных

```go
func uploadHandler(w http.ResponseWriter, r *http.Request) {
    // Проверка метода и авторизации
    // Получение файла
    // Проверка ограничений
    // Сохранение
}
```

#### Скачивание файла:
1. Проверка прав доступа
2. Отправка файла

```go
func downloadHandler(w http.ResponseWriter, r *http.Request) {
    // Проверка прав
    // Отправка файла
}
```

### 4. Текстовый редактор

#### Открытие файла:
1. Проверка прав доступа
2. Чтение содержимого
3. Проверка ограничений
4. Отправка содержимого

```javascript
function openFile(filename) {
    // Запрос к серверу
    // Обработка ответа
    // Инициализация редактора
}
```

#### Сохранение файла:
1. Проверка изменений
2. Отправка содержимого
3. Обновление статуса

```javascript
function saveFile() {
    // Получение содержимого
    // Отправка на сервер
    // Обновление UI
}
```

### 5. Поиск и фильтрация

```javascript
function filterFiles() {
    // Получение текста поиска
    // Фильтрация списка
    // Обновление UI
}
```

### 6. Темная тема

```javascript
// Переключение темы
themeSwitch.addEventListener('change', function(e) {
    // Изменение темы
    // Сохранение настроек
});
```

## Безопасность

### 1. Проверка прав доступа
```go
func checkFileAccess(filename, username string) bool {
    // Проверка владельца
    // Проверка прав администратора
}
```

### 2. Валидация файлов
```go
func validateFile(file *multipart.FileHeader) error {
    // Проверка размера
    // Проверка типа
}
```

## Обработка ошибок

### 1. Серверные ошибки
```go
func handleError(w http.ResponseWriter, err error, status int) {
    // Логирование
    // Ответ клиенту
}
```

### 2. Клиентские ошибки
```javascript
function handleError(error) {
    // Отображение ошибки
    // Обновление UI
}
```

## События и обработчики

### 1. Редактор
```javascript
// Обработчики событий редактора
editor.addEventListener('keyup', updateStats);
editor.addEventListener('scroll', syncScroll);
editor.addEventListener('input', handleChange);
```

### 2. Файловые операции
```javascript
// Обработчики файловых операций
uploadForm.addEventListener('submit', handleUpload);
deleteButton.addEventListener('click', handleDelete);
```

## Утилиты

### 1. Обновление UI
```javascript
function updateUI() {
    updateLineNumbers();
    updateStats();
    updateCursorPosition();
}
```

### 2. Работа с файлами
```go
func getFileInfo(filename string) (*FileInfo, error) {
    // Чтение метаданных
    // Возврат информации
}
```

## Ограничения системы

1. Максимальный размер файла: 4GB
2. Максимальная длина строки: 4294967296 символов
3. Максимальное количество строк: 4294967296
4. Поддерживаемые типы файлов: текстовые файлы

## Производительность

1. Использование буферизации при чтении/записи файлов
2. Кэширование метаданных
3. Оптимизация JavaScript-кода
4. Минимизация запросов к серверу

## Расширение функциональности

Для добавления новых функций:
1. Добавить обработчик в main.go
2. Обновить шаблоны
3. Добавить клиентский код
4. Обновить документацию
