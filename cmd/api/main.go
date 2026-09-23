package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"tools-ecg-backend/database"
	"tools-ecg-backend/handlers"
)

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func initDB() (*sql.DB, error) {
	host := getEnv("DB_HOST", "postgres")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "user_dev")
	pass := getEnv("DB_PASSWORD", "")
	dbname := getEnv("DB_NAME", "db_dev")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable connect_timeout=5",
		host, port, user, pass, dbname,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("sql.Open fehler: %w", err)
	}

	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("db.Ping fehler: %w", err)
	}

	return db, nil
}

func main() {
	db, err := initDB()
	if err != nil {
		log.Fatalf("Datenbankinitialisierung fehlgeschlagen: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Datenbank nicht erreichbar: %v", err)
	}

	// 1. Migrationen beim Start ausführen
	log.Println("Führe Datenbank-Migrationen aus...")
	if err := database.RunMigrations(db); err != nil {
		log.Fatalf("Migration fehlgeschlagen: %v", err)
	}
	log.Println("Migrationen erfolgreich abgeschlossen.")

	// 2. HTTP Routing (Go 1.22+ Standard Mux)
	mux := http.NewServeMux()
	standplanHandler := &handlers.StandplanHandler{DB: db}

	mux.HandleFunc("GET /editions/{year}/standplan", standplanHandler.GetStandplan)
	mux.HandleFunc("POST /shifts/{shiftId}/assignments", standplanHandler.AssignVolunteer)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server läuft auf Port %s...", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}