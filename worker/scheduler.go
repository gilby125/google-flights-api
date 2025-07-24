package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gilby125/google-flights-api/queue"
	"github.com/gilby125/google-flights-api/db"
	"github.com/robfig/cron/v3"
)

// Cronner defines the interface for cron operations needed by the scheduler.
// This allows mocking the cron dependency in tests.
type Cronner interface {
	Start()
	Stop() context.Context // cron.Stop() returns a context
	AddFunc(spec string, cmd func()) (cron.EntryID, error)
}

// Scheduler schedules jobs to be executed at specific times
type Scheduler struct {
	queue      queue.Queue
	postgresDB db.PostgresDB
	cron       Cronner // Use the interface
	stopChan   chan struct{}
}

// NewScheduler creates a new scheduler instance.
// It accepts a Cronner interface; if nil, a default cron.Cron instance is created.
func NewScheduler(queue queue.Queue, postgresDB db.PostgresDB, cronner Cronner) *Scheduler {
	if cronner == nil {
		cronner = cron.New() // Default to real cron if nil is passed
	}
	return &Scheduler{
		queue:      queue,
		postgresDB: postgresDB,
		cron:       cronner, // Assign the interface
		stopChan:   make(chan struct{}),
	}
}

// Start starts the scheduler
func (s *Scheduler) Start() error {
	log.Println("Starting scheduler")
	s.cron.Start() // Call interface method

	// Load and schedule bulk search jobs from database
	err := s.loadScheduledBulkSearches()
	if err != nil {
		return fmt.Errorf("failed to load scheduled bulk searches: %w", err)
	}

	return nil
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	log.Println("Stopping scheduler")
	close(s.stopChan)
	<-s.cron.Stop().Done() // Call interface method
}

// AddJob adds a job payload to the designated scheduled jobs queue.
// Note: This does *not* schedule based on time within the scheduler itself,
// but relies on workers processing the 'scheduled_jobs' queue.

// AddJob adds a job to the scheduler
func (s *Scheduler) AddJob(payload []byte) error {
	// Validate the cron expression
	// if _, err := cronparser.Parse(schedule); err != nil {
	// 	return fmt.Errorf("invalid cron expression: %w", err)
	// }

	// Create a unique job ID
	jobID := fmt.Sprintf("scheduled_job-%d", time.Now().UnixNano())

	// Create the job
	job := queue.Job{
		ID:      jobID,
		Type:    "scheduled_job", // Define a specific job type for scheduled jobs
		Payload: payload,
	}

	// Serialize the job
	jobBytes, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job: %w", err)
	}

	// Add the job to the queue
	queueName := "scheduled_jobs"
	_, err = s.queue.Enqueue(context.Background(), queueName, jobBytes)
	if err != nil {
		return fmt.Errorf("failed to enqueue job: %w", err)
	}

	log.Printf("Scheduled job %s to queue %s", jobID, queueName)
	return nil
}

// loadScheduledBulkSearches loads and schedules bulk search jobs from the database
func (s *Scheduler) loadScheduledBulkSearches() error {
	ctx := context.Background()
	
	// Get all enabled scheduled jobs
	rows, err := s.postgresDB.ListJobs(ctx)
	if err != nil {
		return fmt.Errorf("failed to list scheduled jobs: %w", err)
	}
	defer rows.Close()

	jobCount := 0
	for rows.Next() {
		var job struct {
			ID             int
			Name           string
			CronExpression string
			Enabled        bool
			LastRun        *time.Time
			CreatedAt      time.Time
			UpdatedAt      time.Time
		}
		
		err := rows.Scan(&job.ID, &job.Name, &job.CronExpression, &job.Enabled, &job.LastRun, &job.CreatedAt, &job.UpdatedAt)
		if err != nil {
			log.Printf("Error scanning scheduled job: %v", err)
			continue
		}

		// Only schedule enabled jobs
		if !job.Enabled {
			log.Printf("Skipping disabled job: %s (ID: %d)", job.Name, job.ID)
			continue
		}

		// Add the job to the cron scheduler
		_, err = s.cron.AddFunc(job.CronExpression, func() {
			s.executeScheduledBulkSearch(job.ID, job.Name)
		})
		
		if err != nil {
			log.Printf("Failed to schedule job %s (ID: %d): %v", job.Name, job.ID, err)
			continue
		}
		
		jobCount++
		log.Printf("Scheduled bulk search job: %s (ID: %d) with cron: %s", job.Name, job.ID, job.CronExpression)
	}
	
	log.Printf("Successfully scheduled %d bulk search jobs", jobCount)
	return nil
}

// executeScheduledBulkSearch executes a scheduled bulk search job
func (s *Scheduler) executeScheduledBulkSearch(jobID int, jobName string) {
	ctx := context.Background()
	log.Printf("Executing scheduled bulk search: %s (ID: %d)", jobName, jobID)
	
	// Get job details from database
	details, err := s.postgresDB.GetJobDetailsByID(ctx, jobID)
	if err != nil {
		log.Printf("Error getting job details for %s (ID: %d): %v", jobName, jobID, err)
		return
	}

	// Calculate dates based on execution time if dynamic dates are enabled
	var departureDateFrom, departureDateTo, returnDateFrom, returnDateTo time.Time
	
	if details.DynamicDates {
		// Use dates relative to current execution time
		now := time.Now()
		
		// Get days from execution (default to 0 if not set)
		daysFromExecution := 0
		if details.DaysFromExecution.Valid {
			daysFromExecution = int(details.DaysFromExecution.Int32)
		}
		
		// Get search window days (default to 1 if not set)
		searchWindowDays := 1
		if details.SearchWindowDays.Valid {
			searchWindowDays = int(details.SearchWindowDays.Int32)
		}
		
		// Calculate departure date range relative to execution time
		departureDateFrom = now.AddDate(0, 0, daysFromExecution)
		departureDateTo = departureDateFrom.AddDate(0, 0, searchWindowDays-1)
		
		log.Printf("Dynamic dates for job %s: searching %d days from now (%s to %s)", 
			jobName, daysFromExecution, departureDateFrom.Format("2006-01-02"), departureDateTo.Format("2006-01-02"))
		
		// For round trips, calculate return dates if trip length is specified
		if details.TripType == "round_trip" && details.TripLength.Valid {
			tripLength := int(details.TripLength.Int32)
			returnDateFrom = departureDateFrom.AddDate(0, 0, tripLength)
			returnDateTo = departureDateTo.AddDate(0, 0, tripLength)
			log.Printf("Round trip return dates: %s to %s (trip length: %d days)", 
				returnDateFrom.Format("2006-01-02"), returnDateTo.Format("2006-01-02"), tripLength)
		}
	} else {
		// Use static dates from job details
		departureDateFrom = details.DepartureDateStart
		departureDateTo = details.DepartureDateEnd
		
		if details.ReturnDateStart.Valid {
			returnDateFrom = details.ReturnDateStart.Time
		}
		if details.ReturnDateEnd.Valid {
			returnDateTo = details.ReturnDateEnd.Time
		}
		
		log.Printf("Static dates for job %s: departure %s to %s", 
			jobName, departureDateFrom.Format("2006-01-02"), departureDateTo.Format("2006-01-02"))
	}
	
	// Get trip length
	tripLength := 0
	if details.TripLength.Valid {
		tripLength = int(details.TripLength.Int32)
	}

	// Create bulk search payload with calculated dates
	bulkSearchPayload := BulkSearchPayload{
		Origins:           []string{details.Origin},
		Destinations:      []string{details.Destination},
		DepartureDateFrom: departureDateFrom,
		DepartureDateTo:   departureDateTo,
		ReturnDateFrom:    returnDateFrom,
		ReturnDateTo:      returnDateTo,
		TripLength:        tripLength,
		Adults:            details.Adults,
		Children:          details.Children,
		InfantsLap:        details.InfantsLap,
		InfantsSeat:       details.InfantsSeat,
		TripType:          details.TripType,
		Class:             details.Class,
		Stops:             details.Stops,
		Currency:          details.Currency,
	}

	// Serialize the payload
	payloadBytes, err := json.Marshal(bulkSearchPayload)
	if err != nil {
		log.Printf("Error serializing bulk search payload for job %s: %v", jobName, err)
		return
	}

	// Enqueue the bulk search job
	bulkJobID := fmt.Sprintf("scheduled_bulk_search-%d-%d", jobID, time.Now().Unix())
	_, err = s.queue.Enqueue(ctx, "bulk_search", payloadBytes)
	if err != nil {
		log.Printf("Error enqueuing bulk search for job %s: %v", jobName, err)
		return
	}

	// Update the last run time
	err = s.postgresDB.UpdateJobLastRun(ctx, jobID)
	if err != nil {
		log.Printf("Error updating last run time for job %s: %v", jobName, err)
		// Don't return - the job was successfully enqueued
	}

	log.Printf("Successfully enqueued scheduled bulk search: %s (bulk job ID: %s)", jobName, bulkJobID)
}

// processFlightSearch processes a flight search job
func (s *Scheduler) processFlightSearch(ctx context.Context, job queue.Job) error {
	// Map the payload to flights API arguments
	// var tripType flights.TripType
	// switch payload.TripType {
	// case "one_way":
	// 	tripType = flights.OneWay
	// case "round_trip":
	// 	tripType = flights.RoundTrip
	// default:
	// 	return fmt.Errorf("invalid trip type: %s", payload.TripType)
	// }

	// var class flights.Class
	// switch payload.Class {
	// case "economy":
	// 	class = flights.Economy
	// case "premium_economy":
	// 	class = flights.PremiumEconomy
	// case "business":
	// 	class = flights.Business
	// case "first":
	// 	class = flights.First
	// default:
	// 	return fmt.Errorf("invalid class: %s", payload.Class)
	// }

	// var stops flights.Stops
	// switch payload.Stops {
	// case "nonstop":
	// 	stops = flights.Nonstop
	// case "one_stop":
	// 	stops = flights.Stop1
	// case "two_stops":
	// 	stops = flights.Stop2
	// case "any":
	// 	stops = flights.AnyStops
	// default:
	// 	return fmt.Errorf("invalid stops: %s", payload.Stops)
	// }

	// // Parse currency
	// cur, err := currency.ParseISO(payload.Currency)
	// if err != nil {
	// 	return fmt.Errorf("invalid currency: %s", payload.Currency)
	// }

	// // Get flight offers
	// offers, priceRange, err := session.GetOffers(
	// 	ctx,
	// 	flights.Args{
	// 		Date:        payload.DepartureDate,
	// 		ReturnDate:  payload.ReturnDate,
	// 		SrcAirports: []string{payload.Origin},
	// 		DstAirports: []string{payload.Destination},
	// 		Options: flights.Options{
	// 			Travelers: flights.Travelers{
	// 				Adults:       payload.Adults,
	// 				Children:     payload.Children,
	// 				InfantOnLap:  payload.InfantsLap,
	// 				InfantInSeat: payload.InfantsSeat,
	// 			},
	// 			Currency: cur,
	// 			Stops:    stops,
	// 			Class:    class,
	// 			TripType: tripType,
	// 			Lang:     language.English,
	// 		},
	// 	},
	// )
	// if err != nil {
	// 	return fmt.Errorf("failed to get flight offers: %w", err)
	// }

	// // Store the results
	// return worker.storeFlightOffers(ctx, payload, offers, priceRange)
	return nil
}
