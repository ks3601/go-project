package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/lib/pq"
)

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func main() {
	http.HandleFunc("/entry", signInPage)
	http.HandleFunc("/signInFunc", signIn)
	http.HandleFunc("/signUp", registration)
	http.HandleFunc("/registration", registrationPage)

	http.HandleFunc("/services", servicesPage)
	http.HandleFunc("/schedule", schedulePage)
	http.HandleFunc("/contacts", contactsPage)
	http.HandleFunc("/checkAuth", checkAuth)
	http.HandleFunc("/", homePage)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	http.ListenAndServe(":8080", nil)
}
func signIn(w http.ResponseWriter, r *http.Request) {
	var creds Credentials
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil || creds.Username == "" {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	var username = creds.Username
	var password = creds.Password
	fmt.Println("Username:", username)
	fmt.Println("Password:", password)

	// Подключение к базе данных
	connStr := "user=postgres password=1111 dbname=postgres sslmode=disable host=localhost port=5433"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Не удалось подключиться к базе данных: %v", err)
	}
	defer db.Close()

	// Проверка существования пользователя
	var userExists int
	err = db.QueryRow("SELECT COUNT(*) FROM users WHERE username = $1 AND password = $2", username, password).Scan(&userExists)
	if err != nil {
		log.Fatal(err)
	}

	if userExists == 0 {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Успешный вход
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    username,
		Expires:  time.Now().Add(15 * time.Minute),
		HttpOnly: true,
		Secure:   false,
		Path:     "/",
	})
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": username})
}

// Завершаем транзакцию

// закрытие базы данных перед завершением функции

func checkAuth(w http.ResponseWriter, r *http.Request) {
	// получение конкретного куки файла с токеном аутентификации
	cookie, err := r.Cookie("auth_token")
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	username := cookie.Value
	fmt.Print(username)
	// добавление куки файла в ответ функции
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",                     // Имя cookie, которое будет использоваться клиентом.
		Value:    username,                         // Значение cookie, в данном случае токен аутентификации.
		Expires:  time.Now().Add(15 * time.Minute), // Дата и время, когда cookie станет недействительной.
		HttpOnly: true,                             // Cookie доступна только серверу.
		Secure:   false,                            // Используйте true, если работаете через HTTPS, чтобы передача cookie была защищенной.
		Path:     "/",                              // Указывает область действия cookie, доступно для всех путей на сервере.
	})
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": username})
}
func registration(w http.ResponseWriter, r *http.Request) {
	var creds Credentials
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil || creds.Username == "" {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	// подключение в базе данных
	connStr := "user=postgres password=1111 dbname=postgres sslmode=disable host=localhost port=5433"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Не удалось подключиться к базе данных при регистрации: %v", err)
	}
	defer db.Close()
	// закрытие базы данных перед завершением функции
	var existingAmount bool
	var username = creds.Username
	var password = creds.Password
	fmt.Println(username)
	fmt.Println("Пароль", password)
	// начинаем транзакцию
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	// Сначала проверим, существует ли такой вид дохода для пользователя
	err = tx.QueryRow("SELECT udi.username FROM users AS udi WHERE udi.username = $1", username).Scan(&existingAmount)
	if err != nil && err != sql.ErrNoRows {
		log.Fatal(err)
	}

	if err == sql.ErrNoRows {
		// Если нет такого вида дохода, вставляем новый
		_, err = tx.Exec("INSERT INTO users (username,password) VALUES ($1, $2)", username, password)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		// Если такой доход уже есть, обновляем его
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Завершаем транзакцию
	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}
	// добавление куки файла в ответ функции
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",                     // Имя cookie, которое будет использоваться клиентом.
		Value:    username,                         // Значение cookie, в данном случае токен аутентификации.
		Expires:  time.Now().Add(15 * time.Minute), // Дата и время, когда cookie станет недействительной.
		HttpOnly: true,                             // Cookie доступна только серверу.
		Secure:   false,                            // Используйте true, если работаете через HTTPS, чтобы передача cookie была защищенной.
		Path:     "/",                              // Указывает область действия cookie, доступно для всех путей на сервере.
	})

	w.Header().Set("Content-Type", "application/json")
	// отправка данных в сайт через ResponseWriting
	json.NewEncoder(w).Encode(map[string]string{"response": username})
}

func signInPage(w http.ResponseWriter, r *http.Request) {
	fmt.Println("sdsds")
	http.ServeFile(w, r, "./entry.html")
}
func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Println("sdsdsssss")
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, "./index.html")
}

func servicesPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "services.html")
}

func schedulePage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "schedule.html")
}

func contactsPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "contacts.html")
}

func registrationPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "registration.html")
}
