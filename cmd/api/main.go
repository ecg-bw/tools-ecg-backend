package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "embed"

	_ "github.com/lib/pq"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	"tools-ecg-backend/api"
	"tools-ecg-backend/database"
	"tools-ecg-backend/handler"
	"tools-ecg-backend/repository"
	"tools-ecg-backend/service"
)

//go:embed openapi.yaml
var openAPISpec []byte

func initDB() (*sql.DB, error) {
	host := getEnv("DB_HOST", "postgres")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	pass := getEnv("DB_PASSWORD", "postgres")
	dbname := getEnv("DB_NAME", "tools_ecg")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable connect_timeout=5", host, port, user, pass, dbname)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("sql.Open fehler: %w", err)
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("db.Ping fehler: %w", err)
	}
	return db, nil
}

func main() {
	db, err := initDB()
	if err != nil {
		log.Fatalf("Datenbankverbindung fehlgeschlagen: %v", err)
	}
	defer db.Close()

	// 1. Datenbank-Migrationen idempotent ausführen
	if err := database.RunMigrations(db); err != nil {
		log.Fatalf("Fehler beim Ausführen der DB-Migrationen: %v", err)
	}
	log.Println("DB-Migrationen erfolgreich angewendet.")

	// 2. Repositories initialisieren
	editionRepo := repository.NewEditionRepository(db)
	shiftRepo := repository.NewShiftRepository(db)
	volunteerRepo := repository.NewVolunteerRepository(db)
	roleWaffleRepo := repository.NewRoleWaffleRepository(db)

	// 3. Services initialisieren (SOLID: Dependency Injection)
	editionSvc := service.NewEditionService(editionRepo)
	shiftSvc := service.NewShiftService(shiftRepo)
	volunteerSvc := service.NewVolunteerService(volunteerRepo)
	roleWaffleSvc := service.NewRoleWaffleService(roleWaffleRepo, roleWaffleRepo)

	// 4. API Adapter für generiertes ServerInterface
	apiAdapter := handler.NewApiAdapter(editionSvc, shiftSvc, volunteerSvc, roleWaffleSvc)

	// 5. Router aufsetzen
	rootMux := http.NewServeMux()

	// Swagger Docs & UI
	rootMux.HandleFunc("GET /openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/yaml")
		w.Write(openAPISpec)
	})
	rootMux.Handle("/swagger-ui/", httpSwagger.Handler(httpSwagger.URL("/openapi.yaml")))
	rootMux.HandleFunc("GET /swagger-ui", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger-ui/", http.StatusMovedPermanently)
	})

	// Optionaler Bootstrap Shortcut (Admin-Tooling)
	rootMux.HandleFunc("POST /v1/editions/{editionId}/bootstrap", func(w http.ResponseWriter, r *http.Request) {
		editionIDStr := r.PathValue("editionId")
		var id int
		if _, err := fmt.Sscanf(editionIDStr, "%d", &id); err != nil {
			http.Error(w, "ungültige editionId", http.StatusBadRequest)
			return
		}
		var req struct {
			StartDate string `json:"start_date"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)
		if err := editionSvc.BootstrapEdition(r.Context(), id, req.StartDate); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"message":"Standplan erfolgreich initialisiert"}`))
	})

	// Registriere alle OpenAPI v1 Routen
	apiHandler := api.HandlerWithOptions(apiAdapter, api.StdHTTPServerOptions{
		BaseURL:    "/v1",
		BaseRouter: rootMux,
	})

	port := getEnv("PORT", "8080")
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      corsMiddleware(apiHandler),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Server läuft auf http://localhost:%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server-Fehler: %v", err)
		}
	}()

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Herunterfahren eingeleitet...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Fehler beim Shutdown: %v", err)
	}
	log.Println("Server ordnungsgemäß beendet.")
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}