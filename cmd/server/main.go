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
	// Load configuration
	config.LoadConfig()

	// Connect to MongoDB
	if err := database.Connect(); err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer database.Disconnect()

	// Initialize workers
	emailWorker := workers.NewEmailWorker()
	cleanupJob := workers.NewCleanupJob()
	cleanupJob.Start()

	// Initialize handlers
	authHandler := api.NewAuthHandler()
	userHandler := api.NewUserHandler()
	webhookHandler := api.NewWebhookHandler(emailWorker)

	// Setup router
	r := mux.NewRouter()

	// Apply CORS middleware globally
	r.Use(middleware.CORS)

	// Public routes
	r.HandleFunc("/register", authHandler.Register).Methods("POST", "OPTIONS")
	r.HandleFunc("/login", authHandler.Login).Methods("POST", "OPTIONS")

	// Health check
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy"}`))
	}).Methods("GET")

	// Protected routes (JWT)
	jwtRouter := r.PathPrefix("/user").Subrouter()
	jwtRouter.Use(middleware.JWTAuth)
	jwtRouter.HandleFunc("/settings/email-ttl", userHandler.UpdateEmailTTL).Methods("PUT", "OPTIONS")
	jwtRouter.HandleFunc("/email-logs", userHandler.GetEmailLogs).Methods("GET", "OPTIONS")

	// Webhook routes (API Key)
	webhookRouter := r.PathPrefix("/webhook").Subrouter()
	webhookRouter.Use(middleware.APIKeyAuth)
	webhookRouter.HandleFunc("/send-email", webhookHandler.SendEmail).Methods("POST", "OPTIONS")

	// Start server
	port := config.AppConfig.Port
	addr := fmt.Sprintf(":%s", port)
	
	log.Printf("🚀 Server starting on port %s\n", port)
	log.Printf("📧 Email worker running with %d goroutines\n", config.AppConfig.WorkerConcurrency)
	log.Printf("🧹 Cleanup job scheduled: %s\n", config.AppConfig.CleanupCronSchedule)
	log.Println("\n=== API Endpoints ===")
	log.Println("POST   /register              - Register new user")
	log.Println("POST   /login                 - Login user")
	log.Println("GET    /health                - Health check")
	log.Println("PUT    /user/settings/email-ttl - Update email log TTL (JWT)")
	log.Println("GET    /user/email-logs       - Get email logs (JWT)")
	log.Println("POST   /webhook/send-email    - Send email (API Key)")
	log.Println("=====================\n")

	// Handle graceful shutdown
	go func() {
		if err := http.ListenAndServe(addr, r); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("\n🛑 Shutting down server...")
	
	// Cleanup
	emailWorker.Stop()
	cleanupJob.Stop()
	
	log.Println("✅ Server stopped gracefully")
}
