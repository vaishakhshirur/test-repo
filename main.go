package main

import (
	"database/sql"
	"encoding/gob"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	http.HandleFunc("/insecure", insecureHandler)
	http.HandleFunc("/sqlinject", sqlInjectionHandler)
	http.HandleFunc("/cmdinject", commandInjectionHandler)
	http.ListenAndServe(":8080", nil)
}

// 🔐 Hardcoded credentials
var dbUser = "admin"
var dbPass = "password123"
var dbName = "testdb"

func getDBConnection() *sql.DB {
	connStr := fmt.Sprintf("%s:%s@/%s", dbUser, dbPass, dbName)
	db, err := sql.Open("mysql", connStr)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func insecureHandler(w http.ResponseWriter, r *http.Request) {
	// 🔀 Insecure random token
	token := fmt.Sprintf("%x", rand.Int63())
	fmt.Fprintf(w, "Your insecure token is: %s", token)
}

func sqlInjectionHandler(w http.ResponseWriter, r *http.Request) {
	user := r.URL.Query().Get("user")
	db := getDBConnection()
	defer db.Close()

	// ❌ SQL Injection
	query := fmt.Sprintf("SELECT * FROM users WHERE username = '%s'", user)
	rows, err := db.Query(query)
	if err != nil {
		http.Error(w, "DB error", 500)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		rows.Scan(&name)
		fmt.Fprintf(w, "User found: %s", name)
	}
}

func commandInjectionHandler(w http.ResponseWriter, r *http.Request) {
	filename := r.URL.Query().Get("file")

	// ❌ Command Injection
	cmd := exec.Command("ls", filename)
	output, err := cmd.CombinedOutput()
	if err != nil {
		http.Error(w, "Command failed", 500)
		return
	}
	fmt.Fprintf(w, "File list:\n%s", output)
}

func insecureDeserialization() {
	data := os.Getenv("SERIALIZED_INPUT")
	var m map[string]interface{}

	// ❌ Insecure deserialization
	decoder := gob.NewDecoder(os.Stdin)
	err := decoder.Decode(&m)
	if err != nil {
		log.Printf("Deserialization error: %v", err)
	}
}

func logSensitiveInfo() {
	password := "super_secret_password"
	log.Printf("User logged in with password: %s", password) // ❌ Leaky logging
}
