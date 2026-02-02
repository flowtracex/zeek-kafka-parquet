package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"pipeline/core"
)

// Config represents the application configuration
type Config struct {
	Kafka struct {
		Brokers       []string `json:"brokers"`
		Topic         string   `json:"topic"`
		GroupID       string   `json:"group_id"`
		MaxPollRecords int    `json:"max_poll_records"`
	} `json:"kafka"`
	Output struct {
		BasePath   string `json:"base_path"`
		FilePrefix string `json:"file_prefix"`
	} `json:"output"`
	Write struct {
		FlushBufferMB      int `json:"flush_buffer_mb"`
		FlushIntervalSec   int `json:"flush_interval_seconds"`
		FlushEventCount   int `json:"flush_event_count"`
	} `json:"write"`
}

func main() {
	// Load configuration
	configData, err := os.ReadFile("config/config.json")
	if err != nil {
		log.Fatalf("Failed to read config/config.json: %v", err)
	}

	var config Config
	if err := json.Unmarshal(configData, &config); err != nil {
		log.Fatalf("Failed to parse config/config.json: %v", err)
	}

	// Load normalization rules
	normData, err := os.ReadFile("config/normalization.json")
	if err != nil {
		log.Fatalf("Failed to read config/normalization.json: %v", err)
	}

	normalizationRules, err := core.LoadNormalizationRules(normData)
	if err != nil {
		log.Fatalf("Failed to load normalization rules: %v", err)
	}

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Create components
	consumer, err := core.NewConsumer(core.KafkaConfig{
		Brokers:       config.Kafka.Brokers,
		Topic:         config.Kafka.Topic,
		GroupID:       config.Kafka.GroupID,
		MaxPollRecords: config.Kafka.MaxPollRecords,
	})
	if err != nil {
		log.Fatalf("Failed to create consumer: %v", err)
	}

	normalizer := core.NewNormalizer(normalizationRules)
	enricher := core.NewEnricher() // No config needed - uses per-event enrichment flags

	// Get all log types from normalization rules
	logTypes := make([]string, 0, len(normalizationRules))
	for logType := range normalizationRules {
		logTypes = append(logTypes, logType)
	}
	log.Printf("Processing %d log types: %v", len(logTypes), logTypes)
	
	dispatcher := core.NewDispatcher(logTypes, 10000) // Buffer size per route

	// Create Parquet writers for each log type
	writerConfig := core.WriterConfig{
		BasePath:         config.Output.BasePath,
		FilePrefix:       config.Output.FilePrefix,
		FlushBufferMB:    config.Write.FlushBufferMB,
		FlushIntervalSec: config.Write.FlushIntervalSec,
		FlushEventCount:  config.Write.FlushEventCount,
	}

	var wg sync.WaitGroup

	// Start Kafka consumer
	wg.Add(1)
	go func() {
		defer wg.Done()
		consumer.Start(ctx)
	}()

	// Start error handler
	wg.Add(1)
	go func() {
		defer wg.Done()
		errorCount := 0
		for {
			select {
			case <-ctx.Done():
				return
			case err, ok := <-consumer.Errors():
				if !ok {
					return
				}
				errorCount++
				log.Printf("ERROR [%d]: %v", errorCount, err)
				// Log first 10 errors in detail, then summarize
				if errorCount == 10 {
					log.Printf("WARNING: Many errors detected. Check message format and Kafka connectivity.")
				}
			}
		}
	}()

	// Start normalization and enrichment pipeline
	normalizedChan := make(chan *core.NormalizedEvent, 1000)
	var processedCount int64
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(normalizedChan)
		for {
			select {
			case <-ctx.Done():
				return
			case zeekLog, ok := <-consumer.Output():
				if !ok {
					return
				}
				processedCount++
				if processedCount%100 == 0 {
					log.Printf("Processed %d messages", processedCount)
				}
				// Normalize
				normalized, err := normalizer.Normalize(zeekLog)
				if err != nil {
					// Skip log types that don't have normalization rules
					continue
				}
				normalizedChan <- normalized
			}
		}
	}()

	// Start enrichment
	enrichedChan := make(chan *core.EnrichedEvent, 1000)
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(enrichedChan)
		for {
			select {
			case <-ctx.Done():
				return
			case normalized, ok := <-normalizedChan:
				if !ok {
					return
				}
				// Enrich
				enriched := enricher.Enrich(normalized)
				enrichedChan <- enriched
			}
		}
	}()

	// Start dispatcher
	wg.Add(1)
	go func() {
		defer wg.Done()
		dispatcher.Start(ctx, enrichedChan)
	}()

	// Start Parquet writers (one per log type)
	for _, logType := range logTypes {
		logType := logType // Capture for goroutine
		wg.Add(1)
		go func() {
			defer wg.Done()
			writer := core.NewParquetWriter(logType, writerConfig, dispatcher.GetRoute(logType))
			if err := writer.Start(ctx); err != nil {
				log.Printf("ERROR [%s]: writer failed: %v", logType, err)
			}
		}()
	}

	log.Println("Pipeline started. Press Ctrl+C to stop...")
	log.Printf("Consuming from Kafka topic: %s, group: %s", config.Kafka.Topic, config.Kafka.GroupID)
	log.Printf("Output path: %s", config.Output.BasePath)
	log.Printf("Buffer flush size: %d MB", config.Write.FlushBufferMB)

	// Wait for shutdown signal
	<-sigChan
	log.Println("Shutting down...")

	// Cancel context to stop all goroutines
	cancel()

	// Wait for all goroutines to finish
	wg.Wait()

	log.Println("Pipeline stopped.")
}

