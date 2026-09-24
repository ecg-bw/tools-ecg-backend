package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"

	"tools-ecg-backend/handler"
	"tools-ecg-backend/repository"
	"tools-ecg-backend/service"
)

func initDB() (*sql.DB, error) {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

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
		log.Fatalf("Fehler beim Öffnen der DB: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Datenbank nicht erreichbar: %v", err)
	}
	log.Println("Erfolgreich mit PostgreSQL verbunden.")

	// Repositories
	editionRepo := repository.NewEditionRepository(db)
	shiftRepo := repository.NewShiftRepository(db)
	volunteerRepo := repository.NewVolunteerRepository(db)

	// Services
	editionService := service.NewEditionService(editionRepo)
	shiftService := service.NewShiftService(shiftRepo)
	volunteerService := service.NewVolunteerService(volunteerRepo)

	// Handlers
	editionHandler := handler.NewEditionHandler(editionService)
	shiftHandler := handler.NewShiftHandler(shiftService)
	volunteerHandler := handler.NewVolunteerHandler(volunteerService)

	// Mux Router (Go 1.22+)
	mux := http.NewServeMux()

	// 1. Standplan & Bootstrap
	mux.HandleFunc("GET /v1/editions/{year}/standplan", editionHandler.GetStandplan)
	mux.HandleFunc("POST /v1/editions/{editionId}/bootstrap", editionHandler.Bootstrap)

	// 2. Schichten verwalten
	mux.HandleFunc("GET /v1/days/{dayId}/shifts", shiftHandler.GetShiftsByDay)
	mux.HandleFunc("POST /v1/shifts", shiftHandler.CreateShift)
	mux.HandleFunc("PUT /v1/shifts/{shiftId}", shiftHandler.UpdateShift)

	// 3. Belegungen (Assignments)
	mux.HandleFunc("POST /v1/shifts/{shiftId}/assignments", shiftHandler.AssignVolunteer)
	mux.HandleFunc("DELETE /v1/shifts/{shiftId}/assignments/{assignmentId}", shiftHandler.RemoveAssignment)

	// 4. Helfer & Persönlicher Plan
	mux.HandleFunc("GET /v1/volunteers", volunteerHandler.List)
	mux.HandleFunc("POST /v1/volunteers", volunteerHandler.Create)
	mux.HandleFunc("GET /v1/volunteers/{volunteerId}/shifts", volunteerHandler.GetVolunteerShifts)

	// Einfacher CORS-Wrapper
	corsHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		mux.ServeHTTP(w, r)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Server läuft auf http://localhost:%s\n", port)
	if err := http.ListenAndServe(":"+port, corsHandler); err != nil {
		log.Fatalf("Server-Fehler: %v", err)
	}
}