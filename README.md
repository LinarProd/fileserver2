# fileserver
This is a mini website written in go, thanks to which you can transfer files using your website. The site does not have complex security systems, so you should not use it in complex situations.
### Инструкция по использованию и настройке сайта (на русском и английском языках)

---

#### **Настройка и использование (русский язык)**

1. **Установка необходимых инструментов**:
   - Установите Go (Golang) с [официального сайта](https://golang.org/dl/).
   - Убедитесь, что на вашем компьютере установлен Git (по желанию) и текстовый редактор.

2. **Создание структуры проекта**:
   - Создайте папки для проекта. Например:
     ```
     /project
       ├── main.go
       ├── config.json
       ├── users.json
       ├── templates/
       │     └── main.html
       └── static/
             └── styles.css
     ```

3. **Настройка конфигурации**:
   - Создайте файл `config.json`:
     ```json
     {
         "server_port": "8080",
         "file_dir": "./files",
         "user_file": "./users.json"
     }
     ```

   - Создайте файл `users.json` с пользователями:
     ```json
     {
         "users": [
             { "username": "admin", "password": "password123" },
             { "username": "user1", "password": "mypassword" }
         ]
     }
     ```

4. **Сборка и запуск**:
   - Перейдите в папку с проектом.
   - Выполните команды:
     ```bash
     go mod init project
     go run main.go
     ```
   - Сервер будет доступен по адресу: [http://localhost:8080](http://localhost:8080).

5. **Использование сайта**:
   - Откройте сайт в браузере.
   - Войдите с использованием одного из логинов из `users.json`.
   - Загрузите файл, удалите его или скачайте при необходимости.
   - Для выхода нажмите кнопку "Logout".

---

#### **Setup and Usage (English)**

1. **Install Required Tools**:
   - Install Go (Golang) from the [official site](https://golang.org/dl/).
   - Ensure Git (optional) and a text editor are installed on your machine.

2. **Create Project Structure**:
   - Set up project folders. For example:
     ```
     /project
       ├── main.go
       ├── config.json
       ├── users.json
       ├── templates/
       │     └── main.html
       └── static/
             └── styles.css
     ```

3. **Configure the Project**:
   - Create `config.json` file:
     ```json
     {
         "server_port": "8080",
         "file_dir": "./files",
         "user_file": "./users.json"
     }
     ```

   - Create `users.json` file with users:
     ```json
     {
         "users": [
             { "username": "admin", "password": "password123" },
             { "username": "user1", "password": "mypassword" }
         ]
     }
     ```

4. **Build and Run**:
   - Navigate to the project folder.
   - Execute the following commands:
     ```bash
     go mod init project
     go run main.go
     ```
   - The server will be accessible at: [http://localhost:8080](http://localhost:8080).

5. **Using the Site**:
   - Open the site in your browser.
   - Log in using one of the accounts from `users.json`.
   - Upload, delete, or download files as needed.
   - To log out, click the "Logout" button.

---

### Принцип работы каждой функции (русский и английский)

---

#### **Функция: Login / Вход**

**Русский**:
- Обрабатывает POST-запросы с формы логина.
- Сравнивает введённый логин и пароль с данными в файле `users.json`.
- При успешной аутентификации создаёт cookie для сохранения сессии.

**English**:
- Handles POST requests from the login form.
- Compares the entered username and password with the data in `users.json`.
- On successful authentication, it creates a cookie to maintain the session.

---

#### **Функция: File Upload / Загрузка файла**

**Русский**:
- Обрабатывает POST-запросы для загрузки файлов.
- Сохраняет загруженный файл в директорию, указанную в `config.json` (`file_dir`).
- Перенаправляет пользователя обратно на главную страницу после загрузки.

**English**:
- Handles POST requests for file uploads.
- Saves the uploaded file to the directory specified in `config.json` (`file_dir`).
- Redirects the user back to the main page after the upload.

---

#### **Функция: File Download / Скачивание файла**

**Русский**:
- Обрабатывает GET-запросы для скачивания файлов.
- Проверяет наличие запрашиваемого файла в директории `file_dir`.
- Возвращает файл с заголовками для скачивания.

**English**:
- Handles GET requests for file downloads.
- Checks if the requested file exists in the `file_dir` directory.
- Returns the file with appropriate headers for downloading.

---

#### **Функция: File Delete / Удаление файла**

**Русский**:
- Обрабатывает POST-запросы для удаления файла.
- Удаляет указанный файл из директории `file_dir`.
- Перенаправляет пользователя обратно на главную страницу после удаления.

**English**:
- Handles POST requests for file deletion.
- Deletes the specified file from the `file_dir` directory.
- Redirects the user back to the main page after deletion.

---

#### **Функция: Logout / Выход**

**Русский**:
- Удаляет cookie, используемую для авторизации.
- Перенаправляет пользователя на страницу входа.

**English**:
- Deletes the cookie used for authorization.
- Redirects the user to the login page.

---