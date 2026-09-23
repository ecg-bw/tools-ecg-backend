package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
	"tools-ecg-backend/database"
	"tools-ecg-backend/handlers"
)

func main() {
	dbConnStr := os.Getenv("DATABASE_URL")
	if dbConnStr == "" {
		dbConnStr = "postgres://postgres:postgres@localhost:5432/standplan?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbConnStr)
	if err != nil {
		log.Fatalf("Datenbankverbindung fehlgeschlagen: %v", err)
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