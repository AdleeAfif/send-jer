package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/mux"

	"send-jer/config"
	"send-jer/internal/api"
	"send-jer/internal/api/middleware"
	"send-jer/internal/database"
	"send-jer/internal/workers"
)

func main() {
	config.LoadConfig()

	if err := database.Connect(); err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer database.Disconnect()

	emailWorker := workers.NewEmailWorker()
	cleanupJob := workers.NewCleanupJob()
	cleanupJob.Start()

	authHandler := api.NewAuthHandler()
	userHandler := api.NewUserHandler()
	adminHandler := api.NewAdminHandler()
	webhookHandler := api.NewWebhookHandler(emailWorker)

	r := mux.NewRouter()
	r.Use(middleware.CORS)

	r.HandleFunc("/register", authHandler.Register).Methods("POST", "OPTIONS")
	r.HandleFunc("/login", authHandler.Login).Methods("POST", "OPTIONS")
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy"}`))
	}).Methods("GET")

	jwtRouter := r.PathPrefix("/user").Subrouter()
	jwtRouter.Use(middleware.JWTAuth)
	jwtRouter.HandleFunc("/settings/email-ttl", userHandler.UpdateEmailTTL).Methods("PUT", "OPTIONS")
	jwtRouter.HandleFunc("/email-logs", userHandler.GetEmailLogs).Methods("GET", "OPTIONS")

	adminRouter := r.PathPrefix("/admin").Subrouter()
	adminRouter.Use(middleware.JWTAuth)
	adminRouter.Use(middleware.AdminOnly)
	adminRouter.HandleFunc("/users", adminHandler.ListUsers).Methods("GET", "OPTIONS")
	adminRouter.HandleFunc("/users/{id}", adminHandler.GetUser).Methods("GET", "OPTIONS")
	adminRouter.HandleFunc("/users/{id}", adminHandler.UpdateUser).Methods("PUT", "OPTIONS")
	adminRouter.HandleFunc("/users/{id}", adminHandler.DeleteUser).Methods("DELETE", "OPTIONS")

	webhookRouter := r.PathPrefix("/webhook").Subrouter()
	webhookRouter.Use(middleware.APIKeyAuth)
	webhookRouter.HandleFunc("/send-email", webhookHandler.SendEmail).Methods("POST", "OPTIONS")

	port := config.AppConfig.Port
	addr := fmt.Sprintf(":%s", port)

	log.Printf("Server starting on port %s", port)
	log.Printf("Email worker running with %d goroutines", config.AppConfig.WorkerConcurrency)
	log.Printf("Cleanup job scheduled: %s", config.AppConfig.CleanupCronSchedule)
	log.Println("=== API Endpoints ===")
	log.Println("POST   /register              - Register new account")
	log.Println("POST   /login                 - Login user")
	log.Println("GET    /health                - Health check")
	log.Println("PUT    /user/settings/email-ttl - Update email log TTL (JWT)")
	log.Println("GET    /user/email-logs       - Get email logs (JWT)")
	log.Println("GET    /admin/users           - List users (Admin JWT)")
	log.Println("GET    /admin/users/{id}      - Get user details (Admin JWT)")
	log.Println("PUT    /admin/users/{id}      - Update user role/TTL (Admin JWT)")
	log.Println("DELETE /admin/users/{id}      - Delete user (Admin JWT)")
	log.Println("POST   /webhook/send-email    - Send email (API Key)")
	log.Println("=====================")

	go func() {
		if err := http.ListenAndServe(addr, r); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	emailWorker.Stop()
	cleanupJob.Stop()
	log.Println("Server stopped gracefully")
}
