package worker

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/gilby125/google-flights-api/db"
	"github.com/gilby125/google-flights-api/queue"
	"github.com/robfig/cron/v3"
)

// Scheduler manages scheduled jobs
type Scheduler struct {
	queue      queue.Queue
	postgresDB *db.PostgresDB
	cron       *cron.Cron
	mutex      sync.Mutex
	jobs       map[int]cron.EntryID
}

// NewScheduler creates a new scheduler
func NewScheduler(queue queue.Queue, postgresDB *db.PostgresDB) *Scheduler {
	return &Scheduler{
		queue:      queue,
		postgresDB: postgresDB,
		cron:       cron.New(),
		jobs:       make(map[int]cron.EntryID),
	}
}

// Start starts the scheduler
func (s *Scheduler) Start() error {
	// Load all enabled jobs from the database
	rows, err := s.postgresDB.GetDB().Query(
		"SELECT id, name, friendly_schedule FROM scheduled_jobs WHERE enabled = true",
	)
	if err != nil {
		return fmt.Errorf("failed to load scheduled jobs: %w", err)
	}
	defer rows.Close()

	// Schedule each job
	for rows.Next() {
		var id int
		var name, friendlySchedule string
		if err := rows.Scan(&id, &name, &friendlySchedule); err != nil {
			return fmt.Errorf("failed to scan job row: %w", err)
		}

		if err := s.scheduleJobWithFriendlySchedule(id, friendlySchedule); err != nil {
			log.Printf("Warning: Failed to schedule job %d (%s): %v", id, name, err)
			continue
		}

		log.Printf("Scheduled job %d (%s) with schedule: %s", id, name, friendlySchedule)
	}

	// Start the cron scheduler
	s.cron.Start()
	log.Println("Scheduler started")

	return nil
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	ctx := s.cron.Stop()
	<-ctx.Done()
	log.Println("Scheduler stopped")
}

// scheduleJob schedules a job with the given ID and friendly schedule format
func (s *Scheduler) scheduleJobWithFriendlySchedule(jobID int, friendlySchedule string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// If the job is already scheduled, remove it first
	if entryID, exists := s.jobs[jobID]; exists {
		s.cron.Remove(entryID)
		delete(s.jobs, jobID)
	}

	cronExpr, err := s.parseFriendlySchedule(friendlySchedule)
	if err != nil {
		return fmt.Errorf("failed to parse friendly schedule '%s': %w", friendlySchedule, err)
	}

	// Log the cron expression
	log.Printf("Attempting to schedule job %d with cron expression: %s", jobID, cronExpr)

	// Schedule the job
	entryID, err := s.cron.AddFunc(cronExpr, func() {
		s.executeJob(jobID)
	})
	if err != nil {
		return fmt.Errorf("failed to add cron job: %w", err)
	}

	// Store the entry ID
	s.jobs[jobID] = entryID

	return nil
}

func (s *Scheduler) parseFriendlySchedule(friendlySchedule string) (string, error) {
	log.Printf("Parsing friendly schedule: %s", friendlySchedule)
	if friendlySchedule == "" {
		return "", fmt.Errorf("friendly schedule cannot be empty")
	}

	var n int
	var err error
	var hour, min, sec int

	// every N minutes
	var intervalMinutes int
	n, err = fmt.Sscanf(friendlySchedule, "every %d minutes", &intervalMinutes)
	if err == nil && n == 1 {
		if intervalMinutes > 0 {
			cronExpr := fmt.Sprintf("0 */%d * * *", intervalMinutes)
			log.Printf("Parsed cron expression: %s", cronExpr)
			return cronExpr, nil
		}
		return "", fmt.Errorf("invalid interval for minutes schedule")
	}

	// every N hours
	var intervalHours int
	n, err = fmt.Sscanf(friendlySchedule, "every %d hours", &intervalHours)
	if err == nil && n == 1 {
		if intervalHours > 0 {
			cronExpr := fmt.Sprintf("0 0 */%d * *", intervalHours)
			log.Printf("Parsed cron expression: %s", cronExpr)
			return cronExpr, nil
		}
		return "", fmt.Errorf("invalid interval for hours schedule")
	}

	// daily at HH:mm:ss
	n, err = fmt.Sscanf(friendlySchedule, "daily at %d:%d:%d", &hour, &min, &sec)
	if err == nil && n == 3 {
		if hour >= 0 && hour <= 23 && min >= 0 && min <= 59 && sec >= 0 && sec <= 59 {
			cronExpr := fmt.Sprintf("%d %d %d * *", min, hour, "*", "*")
			log.Printf("Parsed cron expression: %s", cronExpr)
			return cronExpr, nil
		}
		return "", fmt.Errorf("invalid time format for daily schedule")
	}

	// daily at HH:mm
	n, err = fmt.Sscanf(friendlySchedule, "daily at %d:%d", &hour, &min)
	if err == nil && n == 2 {
		if hour >= 0 && hour <= 23 && min >= 0 && min <= 59 {
			cronExpr := fmt.Sprintf("0 %d %d * * *", min, hour)
			log.Printf("Parsed cron expression: %s", cronExpr)
			return cronExpr, nil
		}
		return "", fmt.Errorf("invalid time format for daily schedule")
	}

	// weekly on days at HH:mm:ss
	var daysStr string
	n, err = fmt.Sscanf(friendlySchedule, "weekly on %s at %d:%d:%d", &daysStr, &hour, &min, &sec)
	if err == nil && n == 4 {
		if hour >= 0 && hour <= 23 && min >= 0 && min <= 59 && sec >= 0 && sec <= 59 {
			days := strings.ToLower(daysStr)
			daysOfWeek := []string{}
			for _, day := range strings.Split(days, ",") {
				switch strings.TrimSpace(day) {
				case "monday", "mon":
					daysOfWeek = append(daysOfWeek, "mon")
				case "tuesday", "tue":
					daysOfWeek = append(daysOfWeek, "tue")
				case "wednesday", "wed":
					daysOfWeek = append(daysOfWeek, "wed")
				case "thursday", "thu":
					daysOfWeek = append(daysOfWeek, "thu")
				case "friday", "fri":
					daysOfWeek = append(daysOfWeek, "fri")
				case "saturday", "sat":
					daysOfWeek = append(daysOfWeek, "sat")
				case "sunday", "sun":
					daysOfWeek = append(daysOfWeek, "sun")
				default:
					return "", fmt.Errorf("invalid day of week: %s", day)
				}
			}
			cronDays := strings.Join(daysOfWeek, ",")
			cronExpr := fmt.Sprintf("%d %d %d * %s", min, hour, "*", cronDays)
			log.Printf("Parsed cron expression: %s", cronExpr)
			return cronExpr, nil
		}
		return "", fmt.Errorf("invalid time format for weekly schedule")
	}

	// weekly on days at HH:mm
	n, err = fmt.Sscanf(friendlySchedule, "weekly on %s at %d:%d", &daysStr, &hour, &min)
	if err == nil && n == 3 {
		if hour >= 0 && hour <= 23 && min >= 0 && min <= 59 {
			days := strings.ToLower(daysStr)
			daysOfWeek := []string{}
			for _, day := range strings.Split(days, ",") {
				switch strings.TrimSpace(day) {
				case "monday", "mon":
					daysOfWeek = append(daysOfWeek, "mon")
				case "tuesday", "tue":
					daysOfWeek = append(daysOfWeek, "tue")
				case "wednesday", "wed":
					daysOfWeek = append(daysOfWeek, "wed")
				case "thursday", "thu":
					daysOfWeek = append(daysOfWeek, "thu")
				case "friday", "fri":
					daysOfWeek = append(daysOfWeek, "fri")
				case "saturday", "sat":
					daysOfWeek = append(daysOfWeek, "sat")
				case "sunday", "sun":
					daysOfWeek = append(daysOfWeek, "sun")
				default:
					return "", fmt.Errorf("invalid day of week: %s", day)
				}
			}
			cronDays := strings.Join(daysOfWeek, ",")
			cronExpr := fmt.Sprintf("0 %d %d * * %s", min, hour, cronDays)
			log.Printf("Parsed cron expression: %s", cronExpr)
			return cronExpr, nil
		}
		return "", fmt.Errorf("invalid time format for weekly schedule")
	}
	if err != nil && n != 0 {
		return "", fmt.Errorf("failed to parse friendly schedule: %w", err)
	}

	// Add more formats here (e.g., monthly, yearly) as needed

	return "", fmt.Errorf("unrecognized friendly schedule format")
}

// executeJob executes a job with the given ID
func (s *Scheduler) executeJob(jobID int) {
	log.Printf("Executing job %d", jobID)

	// Update the last run time
	_, err := s.postgresDB.GetDB().Exec(
		"UPDATE scheduled_jobs SET last_run = NOW() WHERE id = $1",
		jobID,
	)
	if err != nil {
		log.Printf("Error updating last run time for job %d: %v", jobID, err)
	}

	// Get the job details
	var job struct {
		ID              int
		Name            string
		Origin          string
		Destination     string
		DateStart       time.Time
		DateEnd         time.Time
		ReturnDateStart sql.NullTime
		ReturnDateEnd   sql.NullTime
		TripLength      sql.NullInt32
		Adults          int
		Children        int
		InfantsLap      int
		InfantsSeat     int
		TripType        string
		Class           string
		Stops           string
		Currency        string
	}

	err = s.postgresDB.GetDB().QueryRow(
		`SELECT j.id, j.name, d.origin, d.destination, 
			d.departure_date_start, d.departure_date_end, 
			d.return_date_start, d.return_date_end, 
			d.trip_length, d.adults, d.children, d.infants_lap, d.infants_seat, 
			d.trip_type, d.class, d.stops, 'USD' as currency
		FROM scheduled_jobs j
		JOIN job_details d ON j.id = d.job_id
		WHERE j.id = $1`,
		jobID,
	).Scan(
		&job.ID, &job.Name, &job.Origin, &job.Destination,
		&job.DateStart, &job.DateEnd,
		&job.ReturnDateStart, &job.ReturnDateEnd,
		&job.TripLength, &job.Adults, &job.Children, &job.InfantsLap, &job.InfantsSeat,
		&job.TripType, &job.Class, &job.Stops, &job.Currency,
	)
	if err != nil {
		log.Printf("Error getting job details for job %d: %v", jobID, err)
		return
	}

	// Create a bulk search payload
	payload := BulkSearchPayload{
		Origins:           []string{job.Origin},
		Destinations:      []string{job.Destination},
		DepartureDateFrom: job.DateStart,
		DepartureDateTo:   job.DateEnd,
		Adults:            job.Adults,
		Children:          job.Children,
		InfantsLap:        job.InfantsLap,
		InfantsSeat:       job.InfantsSeat,
		TripType:          job.TripType,
		Class:             job.Class,
		Stops:             job.Stops,
		Currency:          job.Currency,
	}

	// Add return date information if available
	if job.ReturnDateStart.Valid && job.ReturnDateEnd.Valid {
		payload.ReturnDateFrom = job.ReturnDateStart.Time
		payload.ReturnDateTo = job.ReturnDateEnd.Time
	}

	// Add trip length if available
	if job.TripLength.Valid {
		payload.TripLength = int(job.TripLength.Int32)
	}

	// Enqueue the job
	log.Printf("Job %d: Origin=%s, Destination=%s", jobID, job.Origin, job.Destination)
	jobIDStr := fmt.Sprintf("%d", jobID)
	_, err = s.queue.Enqueue(context.Background(), "bulk_search", payload)
	if err != nil {
		log.Printf("Error enqueueing bulk search job: %v", err)
		return
	}
	log.Printf("Enqueued bulk search job with ID: %s", jobIDStr)
}

// AddJob adds a new job to the scheduler
func (s *Scheduler) AddJob(jobID int, friendlySchedule string) error {
	return s.scheduleJobWithFriendlySchedule(jobID, friendlySchedule)
}

// RemoveJob removes a job from the scheduler
func (s *Scheduler) RemoveJob(jobID int) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if entryID, exists := s.jobs[jobID]; exists {
		s.cron.Remove(entryID)
		delete(s.jobs, jobID)
		log.Printf("Removed job %d from scheduler", jobID)
	}
}

// UpdateJob updates an existing job in the scheduler
func (s *Scheduler) UpdateJob(jobID int, friendlySchedule string) error {
	return s.scheduleJobWithFriendlySchedule(jobID, friendlySchedule)
}
