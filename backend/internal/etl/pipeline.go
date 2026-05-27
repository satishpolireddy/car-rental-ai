package etl

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/satishpolireddy/car-rental-ai/config"
	"github.com/satishpolireddy/car-rental-ai/internal/models"
	"github.com/satishpolireddy/car-rental-ai/internal/repository"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
)

// Pipeline is a concurrent ETL processor that ingests raw car inventory data,
// transforms it, and loads it into SQL Server.
// Uses worker pools for scalable batch processing.
type Pipeline struct {
	db          *gorm.DB
	carRepo     *repository.CarRepository
	cfg         config.ETLConfig
	stopCh      chan struct{}
	wg          sync.WaitGroup
}

// RawCarRecord simulates data coming from an external fleet management system or CSV.
type RawCarRecord struct {
	ExternalID   string  `json:"external_id"`
	Make         string  `json:"make"`
	Model        string  `json:"model"`
	Year         int     `json:"year"`
	Category     string  `json:"category"`
	DailyRate    float64 `json:"daily_rate"`
	Location     string  `json:"location"`
	Seats        int     `json:"seats"`
	Transmission string  `json:"transmission"`
	FuelType     string  `json:"fuel_type"`
	Features     string  `json:"features"`
	ImageURL     string  `json:"image_url"`
}

func NewPipeline(db *gorm.DB, carRepo *repository.CarRepository, cfg config.ETLConfig) *Pipeline {
	return &Pipeline{
		db:      db,
		carRepo: carRepo,
		cfg:     cfg,
		stopCh:  make(chan struct{}),
	}
}

// Start runs the ETL pipeline on a schedule until Stop() is called.
func (p *Pipeline) Start(ctx context.Context) {
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		ticker := time.NewTicker(p.cfg.PollInterval)
		defer ticker.Stop()

		log.Info("ETL pipeline started")
		for {
			select {
			case <-ticker.C:
				if err := p.Run(ctx); err != nil {
					log.WithError(err).Error("ETL pipeline run failed")
				}
			case <-p.stopCh:
				log.Info("ETL pipeline stopping")
				return
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (p *Pipeline) Stop() {
	close(p.stopCh)
	p.wg.Wait()
}

// Run executes one full ETL cycle: Extract → Transform → Load.
func (p *Pipeline) Run(ctx context.Context) error {
	start := time.Now()
	etlLog := &models.ETLLog{
		PipelineName: "car_inventory",
		Status:       "running",
		StartedAt:    start,
	}
	p.db.Create(etlLog)

	records, err := p.extract(ctx)
	if err != nil {
		p.failLog(etlLog, err)
		return fmt.Errorf("extract: %w", err)
	}
	etlLog.RecordsIn = len(records)

	transformed, err := p.transform(ctx, records)
	if err != nil {
		p.failLog(etlLog, err)
		return fmt.Errorf("transform: %w", err)
	}

	loaded, err := p.load(ctx, transformed)
	if err != nil {
		p.failLog(etlLog, err)
		return fmt.Errorf("load: %w", err)
	}
	etlLog.RecordsOut = loaded

	now := time.Now()
	etlLog.Status = "success"
	etlLog.FinishedAt = &now
	p.db.Save(etlLog)

	log.WithFields(log.Fields{
		"records_in":  etlLog.RecordsIn,
		"records_out": loaded,
		"duration_ms": time.Since(start).Milliseconds(),
	}).Info("ETL pipeline completed")

	return nil
}

// extract simulates pulling raw records from an external source.
// In production, replace with API calls, CSV/S3 reads, or Kafka consumer.
func (p *Pipeline) extract(ctx context.Context) ([]RawCarRecord, error) {
	// Simulated source data — replace with real external source
	return []RawCarRecord{
		{ExternalID: "EXT001", Make: "Toyota", Model: "Camry", Year: 2023, Category: "standard", DailyRate: 65.00, Location: "New York", Seats: 5, Transmission: "automatic", FuelType: "petrol"},
		{ExternalID: "EXT002", Make: "Tesla", Model: "Model 3", Year: 2024, Category: "luxury", DailyRate: 120.00, Location: "Los Angeles", Seats: 5, Transmission: "automatic", FuelType: "electric"},
		{ExternalID: "EXT003", Make: "Ford", Model: "Explorer", Year: 2023, Category: "suv", DailyRate: 90.00, Location: "Chicago", Seats: 7, Transmission: "automatic", FuelType: "petrol"},
		{ExternalID: "EXT004", Make: "Honda", Model: "Civic", Year: 2024, Category: "economy", DailyRate: 45.00, Location: "New York", Seats: 5, Transmission: "automatic", FuelType: "petrol"},
		{ExternalID: "EXT005", Make: "Mercedes", Model: "E-Class", Year: 2024, Category: "luxury", DailyRate: 180.00, Location: "Miami", Seats: 5, Transmission: "automatic", FuelType: "petrol"},
	}, nil
}

// transform validates and normalises raw records using a worker pool.
func (p *Pipeline) transform(ctx context.Context, records []RawCarRecord) ([]models.Car, error) {
	type result struct {
		car models.Car
		err error
	}

	jobs := make(chan RawCarRecord, len(records))
	results := make(chan result, len(records))

	// Spin up worker pool
	g, gctx := errgroup.WithContext(ctx)
	for i := 0; i < p.cfg.Workers; i++ {
		g.Go(func() error {
			for raw := range jobs {
				select {
				case <-gctx.Done():
					return gctx.Err()
				default:
					car, err := transformRecord(raw)
					results <- result{car: car, err: err}
				}
			}
			return nil
		})
	}

	// Send jobs
	for _, r := range records {
		jobs <- r
	}
	close(jobs)

	// Wait and close results
	go func() {
		g.Wait()
		close(results)
	}()

	var cars []models.Car
	for res := range results {
		if res.err != nil {
			log.WithError(res.err).Warn("Skipping invalid record during transform")
			continue
		}
		cars = append(cars, res.car)
	}

	return cars, g.Wait()
}

func transformRecord(raw RawCarRecord) (models.Car, error) {
	if raw.Make == "" || raw.Model == "" || raw.Year == 0 {
		return models.Car{}, fmt.Errorf("invalid record: missing required fields")
	}
	if raw.DailyRate <= 0 {
		return models.Car{}, fmt.Errorf("invalid daily rate for %s %s", raw.Make, raw.Model)
	}
	category := raw.Category
	if category == "" {
		category = "standard"
	}
	seats := raw.Seats
	if seats == 0 {
		seats = 5
	}
	return models.Car{
		Make:         raw.Make,
		Model:        raw.Model,
		Year:         raw.Year,
		Category:     category,
		DailyRate:    raw.DailyRate,
		Available:    true,
		Location:     raw.Location,
		Seats:        seats,
		Transmission: raw.Transmission,
		FuelType:     raw.FuelType,
		Features:     raw.Features,
		ImageURL:     raw.ImageURL,
	}, nil
}

// load persists cars in configurable batches for memory efficiency.
func (p *Pipeline) load(ctx context.Context, cars []models.Car) (int, error) {
	total := 0
	for i := 0; i < len(cars); i += p.cfg.BatchSize {
		end := i + p.cfg.BatchSize
		if end > len(cars) {
			end = len(cars)
		}
		batch := cars[i:end]
		if err := p.carRepo.BulkUpsert(ctx, batch); err != nil {
			return total, fmt.Errorf("batch load at offset %d: %w", i, err)
		}
		total += len(batch)
	}
	return total, nil
}

func (p *Pipeline) failLog(etlLog *models.ETLLog, err error) {
	now := time.Now()
	etlLog.Status = "failed"
	etlLog.ErrorMsg = err.Error()
	etlLog.FinishedAt = &now
	p.db.Save(etlLog)
}
