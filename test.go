package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
)

type User struct {
	Name  string
	Email *string
}

func main() {
	http.HandleFunc("/ssrf", ssrfHandler)
	http.HandleFunc("/sql", sqlHandler)
	http.HandleFunc("/nil", nilPointerHandler)
	http.HandleFunc("/file", fileReadHandler)
	http.HandleFunc("/leaky", leakSecretHandler)
	log.Println("Starting on :8080")
	http.ListenAndServe(":8080", nil)
}

// ❌ SSRF: Making a GET request to user-provided URL
func ssrfHandler(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Query().Get("url")
	resp, err := http.Get(url) // 🚨 User controls `url`
	if err != nil {
		http.Error(w, "request failed", 500)
		return
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	w.Write(body)
}

// ❌ SQL injection and hardcoded credentials
func sqlHandler(w http.ResponseWriter, r *http.Request) {
	user := r.URL.Query().Get("user")

	db, err := sql.Open("postgres", "user=admin password=secret dbname=test sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	query := fmt.Sprintf("SELECT * FROM users WHERE username = '%s'", user) // 🚨 SQLi
	rows, err := db.Query(query)
	if err != nil {
		http.Error(w, "query failed", 500)
		return
	}
	defer rows.Close()
	fmt.Fprintf(w, "Queried for user: %s", user)
}

// ❌ Nil pointer dereference
func nilPointerHandler(w http.ResponseWriter, r *http.Request) {
	u := User{Name: "Test"}
	fmt.Fprintf(w, "User Email: %s", *u.Email) // 🚨 Panics if Email is nil
}

// ❌ Path traversal
func fileReadHandler(w http.ResponseWriter, r *http.Request) {
	filename := r.URL.Query().Get("file")
	data, err := ioutil.ReadFile("/tmp/uploads/" + filename) // 🚨 No sanitization
	if err != nil {
		http.Error(w, "file read failed", 500)
		return
	}
	w.Write(data)
}

// ❌ Leaky log
func leakSecretHandler(w http.ResponseWriter, r *http.Request) {
	apiKey := os.Getenv("API_KEY")
	log.Printf("API_KEY is: %s", apiKey) // 🚨 Never log secrets
	fmt.Fprintln(w, "Internal logged API_KEY")
}
