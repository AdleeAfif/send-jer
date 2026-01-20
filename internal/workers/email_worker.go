package workers

import (
	"log"
	"sync"

	"send-jer/config"
	"send-jer/internal/models"
	"send-jer/internal/services"
)

// EmailJob represents a job to send an email
type EmailJob struct {
	User    *models.User
	Request models.SendEmailRequest
	LogID   string
}

// EmailWorker handles async email sending
type EmailWorker struct {
	queue        chan EmailJob
	emailService *services.EmailService
	wg           sync.WaitGroup
}

// NewEmailWorker creates a new EmailWorker
func NewEmailWorker() *EmailWorker {
	worker := &EmailWorker{
		queue:        make(chan EmailJob, 100), // Buffer up to 100 jobs
		emailService: services.NewEmailService(),
	}
	
	// Start workers
	worker.Start()
	
	return worker
}

// Start initializes worker goroutines
func (w *EmailWorker) Start() {
	concurrency := config.AppConfig.WorkerConcurrency
	
	for i := 0; i < concurrency; i++ {
		w.wg.Add(1)
		go w.worker(i)
	}
	
	log.Printf("Email worker started with %d goroutines\n", concurrency)
}

// worker processes jobs from the queue
func (w *EmailWorker) worker(id int) {
	defer w.wg.Done()
	
	for job := range w.queue {
		log.Printf("Worker %d: Processing email to %s\n", id, job.Request.To)
		
		// Send email
		err := w.emailService.SendEmail(
			job.User,
			job.Request.To,
			job.Request.Subject,
			job.Request.Body,
			job.Request.IsHTML,
		)
		
		// Update log with result
		status := "success"
		errorMessage := ""
		
		if err != nil {
			status = "failed"
			errorMessage = err.Error()
			log.Printf("Worker %d: Failed to send email to %s: %s\n", id, job.Request.To, err.Error())
		} else {
			log.Printf("Worker %d: Successfully sent email to %s\n", id, job.Request.To)
		}
		
		// Log the email (update existing log if LogID is provided)
		_, logErr := w.emailService.LogEmail(
			job.User.ID,
			job.Request.To,
			job.Request.Subject,
			job.Request.Body,
			status,
			errorMessage,
		)
		
		if logErr != nil {
			log.Printf("Worker %d: Failed to log email: %s\n", id, logErr.Error())
		}
	}
	
	log.Printf("Worker %d: Stopped\n", id)
}

// QueueEmail adds an email to the queue
func (w *EmailWorker) QueueEmail(user *models.User, req models.SendEmailRequest) (string, error) {
	// Create a preliminary log entry
	logEntry, err := w.emailService.LogEmail(
		user.ID,
		req.To,
		req.Subject,
		req.Body,
		"pending",
		"",
	)
	
	if err != nil {
		return "", err
	}
	
	job := EmailJob{
		User:    user,
		Request: req,
		LogID:   logEntry.ID.Hex(),
	}
	
	// Add to queue
	w.queue <- job
	
	return logEntry.ID.Hex(), nil
}

// Stop gracefully stops the worker
func (w *EmailWorker) Stop() {
	close(w.queue)
	w.wg.Wait()
	log.Println("Email worker stopped")
}
