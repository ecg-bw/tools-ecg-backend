package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "embed"

	_ "github.com/lib/pq"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	swagger "tools-ecg-backend/api"
	"tools-ecg-backend/database"
	"tools-ecg-backend/internal/api"
	"tools-ecg-backend/internal/api/handler"
	"tools-ecg-backend/internal/config"
	"tools-ecg-backend/internal/service"
	"tools-ecg-backend/internal/repository"
)

func initDB(config config.DatabaseConfig) (*sql.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable connect_timeout=5", config.Host, config.Port, config.User, config.Password, config.Name)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("sql.Open fehler: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("db.Ping fehler: %w", err)
	}
	return db, nil
}

func main() {
	// 1. Structured Logging initialisieren (JSON für Produktion)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)
	
	config := config.NewConfig()
	err := config.Load()
	if err != nil {
		slog.Error("loading config failed: ", "error", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, err := initDB(config.GetDatabaseConfig())
	if err != nil {
		slog.Error("create database connection failed: ", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// 1. Datenbank-Migrationen idempotent ausführen
	if err := database.RunMigrations(db); err != nil {
		slog.Error("failed to run database migrations: ", "error", err)
	}
	slog.Info("migrations applied successfully")

	// 5. Router aufsetzen
	rootMux := http.NewServeMux()

	// Swagger Docs & UI
	openAPISpec := swagger.GetOpenAPISpec()
	rootMux.HandleFunc("GET /openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/yaml")
		w.Write(openAPISpec)
	})
	rootMux.Handle("/swagger-ui/", httpSwagger.Handler(httpSwagger.URL("/openapi.yaml")))
	rootMux.HandleFunc("GET /swagger-ui", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger-ui/", http.StatusMovedPermanently)
	})

	// Registriere alle OpenAPI v1 Routen
	standplanRepo := repository.NewStandplanRepository(db)
	stanplanService := service.NewStandplanService(standplanRepo)
	apiAdapter := handler.NewApiAdapter(stanplanService)

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
		slog.Info("server started http://loacalhost:" + port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server failed to start: ", "error", err)
		}
	}()

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("shutdown started...")

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("server shutdown failed: ", "error", err)
	}
	log.Println("Server ordnungsgemäß beendet.")
	slog.Info("shutdown completed")
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