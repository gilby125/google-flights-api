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

	// Example: Schedule a job to run every minute
	// This is just an example, replace with actual scheduled jobs from the database
	_, err := s.cron.AddFunc("@every 1m", func() { // Call interface method
		log.Println("Running scheduled job")
		// Fetch pending searches from the database
		// Enqueue them to the queue
	})
	if err != nil {
		return fmt.Errorf("failed to add example cron function: %w", err)
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
