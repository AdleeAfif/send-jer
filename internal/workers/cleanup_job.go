package workers

import (
	"log"
	"time"

	"github.com/robfig/cron/v3"

	"send-jer/config"
	"send-jer/internal/services"
)

// CleanupJob handles cleanup of old email logs
type CleanupJob struct {
	userService  *services.UserService
	emailService *services.EmailService
	cron         *cron.Cron
}

// NewCleanupJob creates a new CleanupJob
func NewCleanupJob() *CleanupJob {
	return &CleanupJob{
		userService:  services.NewUserService(),
		emailService: services.NewEmailService(),
		cron:         cron.New(),
	}
}

// Start starts the cleanup cron job
func (j *CleanupJob) Start() {
	schedule := config.AppConfig.CleanupCronSchedule
	
	_, err := j.cron.AddFunc(schedule, j.cleanup)
	if err != nil {
		log.Printf("Failed to schedule cleanup job: %s\n", err.Error())
		return
	}
	
	j.cron.Start()
	log.Printf("Cleanup job scheduled with schedule: %s\n", schedule)
	
	// Run once on startup
	go j.cleanup()
}

// cleanup performs the actual cleanup
func (j *CleanupJob) cleanup() {
	log.Println("Starting email log cleanup...")
	startTime := time.Now()
	
	// Get all users with TTL > 0
	users, err := j.userService.GetAllUsersWithTTL()
	if err != nil {
		log.Printf("Failed to get users for cleanup: %s\n", err.Error())
		return
	}
	
	totalDeleted := int64(0)
	
	for _, user := range users {
		deleted, err := j.emailService.CleanupOldLogs(user.ID, user.EmailLogTTLDays)
		if err != nil {
			log.Printf("Failed to cleanup logs for user %s: %s\n", user.Email, err.Error())
			continue
		}
		
		if deleted > 0 {
			log.Printf("Deleted %d old logs for user %s (TTL: %d days)\n", deleted, user.Email, user.EmailLogTTLDays)
			totalDeleted += deleted
		}
	}
	
	duration := time.Since(startTime)
	log.Printf("Cleanup completed: Deleted %d logs from %d users in %v\n", totalDeleted, len(users), duration)
}

// Stop stops the cleanup job
func (j *CleanupJob) Stop() {
	j.cron.Stop()
	log.Println("Cleanup job stopped")
}
