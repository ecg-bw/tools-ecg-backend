package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
)

type DBStatus struct {
	Status      string    `json:"status"`
	Stage       string    `json:"stage"`
	CurrentUser string    `json:"current_user"`
	CurrentDB   string    `json:"current_db"`
	Time        time.Time `json:"server_time"`
	TestTableOk bool      `json:"test_table_write_ok"`
}

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
	stage := getEnv("STAGE", "unknown")
	port := getEnv("PORT", "8080")

	log.Printf("[INFO] Starte Backend für Stage: %s...", stage)

	db, err := initDB()
	if err != nil {
		log.Printf("[WARN] Datenbankverbindung fehlgeschlagen: %v", err)
	} else {
		log.Println("[INFO] Verbindung zu PostgreSQL erfolgreich hergestellt.")
		defer db.Close()
	}

	// 1. Healthcheck / Ping
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "up",
			"stage":  stage,
		})
	})

	// 2. Ausführlicher DB-Isolationstest
	http.HandleFunc("/api/test-db", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// CORS Header für Frontend-Zugriff erlauben
		w.Header().Set("Access-Control-Allow-Origin", "*")

		if db == nil {
			var connErr error
			db, connErr = initDB()
			if connErr != nil {
				w.WriteHeader(http.StatusServiceUnavailable)
				json.NewEncoder(w).Encode(map[string]string{
					"error":  "Keine Verbindung zur Datenbank möglich",
					"detail": connErr.Error(),
				})
				return
			}
		}

		var currentUser, currentDB string
		var dbTime time.Time

		// Prüfe tatsächlichen User und DB laut Postgres
		err := db.QueryRow("SELECT current_user, current_database(), NOW()").Scan(&currentUser, &currentDB, &dbTime)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Query fehlgeschlagen", "detail": err.Error()})
			return
		}

		// Berechtigungstest: Erstelle eine kleine temporäre Tabelle und schreibe hinein
		writeOk := false
		_, err = db.Exec(`
			CREATE TABLE IF NOT EXISTS _isolation_test (id serial primary key, created_at timestamp);
			INSERT INTO _isolation_test (created_at) VALUES (NOW());
		`)
		if err == nil {
			writeOk = true
		} else {
			log.Printf("[WARN] Schreibtest auf Tabelle fehlgeschlagen: %v", err)
		}

		res := DBStatus{
			Status:      "connected",
			Stage:       stage,
			CurrentUser: currentUser,
			CurrentDB:   currentDB,
			Time:        dbTime,
			TestTableOk: writeOk,
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(res)
	})

	log.Printf("[INFO] Server lauscht auf Port :%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Serverabsturz: %v", err)
	}
}