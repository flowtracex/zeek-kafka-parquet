package core

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/parquet-go/parquet-go"

	"pipeline/schema"
)

// WriterConfig holds writer configuration
type WriterConfig struct {
	BasePath      string
	FilePrefix    string
	FlushBufferMB int
}

// ParquetWriter writes events to Parquet files
// One writer per log_type with async flush
type ParquetWriter struct {
	logType      string
	config       WriterConfig
	input        <-chan *EnrichedEvent
	buffer       []interface{} // Current in-memory buffer
	bufferSize   int64          // Current buffer size in bytes (approximate)
	bufferLimit  int64          // Buffer limit in bytes (100MB)
	mu           sync.Mutex      // Protects buffer and bufferSize
	flushWg      sync.WaitGroup  // Tracks background flush operations
	fileCounter  int
	basePath     string
}

// NewParquetWriter creates a new Parquet writer for a log_type
func NewParquetWriter(logType string, cfg WriterConfig, input <-chan *EnrichedEvent) *ParquetWriter {
	bufferLimit := int64(cfg.FlushBufferMB) * 1024 * 1024 // Convert MB to bytes

	return &ParquetWriter{
		logType:     logType,
		config:      cfg,
		input:       input,
		buffer:      make([]interface{}, 0, 1000),
		bufferLimit: bufferLimit,
		basePath:    filepath.Join(cfg.BasePath, logType),
	}
}

// Start begins the writer goroutine
// This is the ONLY writer goroutine for this log_type
func (pw *ParquetWriter) Start(ctx context.Context) error {
	// Create output directory
	if err := os.MkdirAll(pw.basePath, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Main writer loop
	for {
		select {
		case <-ctx.Done():
			// Flush remaining buffer before shutdown
			pw.flushBuffer(ctx, true)
			pw.flushWg.Wait() // Wait for all background flushes to complete
			return nil

		case event, ok := <-pw.input:
			if !ok {
				// Channel closed - flush and exit
				pw.flushBuffer(ctx, true)
				pw.flushWg.Wait()
				return nil
			}

			// Convert event to Parquet struct
			parquetEvent := pw.convertEvent(event)
			if parquetEvent == nil {
				continue
			}

			// Add to buffer
			pw.mu.Lock()
			pw.buffer = append(pw.buffer, parquetEvent)
			// Approximate size: assume ~1KB per event (conservative estimate)
			pw.bufferSize += 1024
			shouldFlush := pw.bufferSize >= pw.bufferLimit
			pw.mu.Unlock()

			// CRITICAL: Async flush when buffer limit reached
			// This does NOT block the Kafka consumer or other writers
			if shouldFlush {
				pw.flushBuffer(ctx, false)
			}
		}
	}
}

// flushBuffer swaps the current buffer and flushes it in a background goroutine
// This ensures the Kafka consumer and other writers are NEVER blocked
func (pw *ParquetWriter) flushBuffer(ctx context.Context, blocking bool) {
	pw.mu.Lock()
	if len(pw.buffer) == 0 {
		pw.mu.Unlock()
		return
	}

	// Swap buffer: create new buffer, keep old one for flushing
	oldBuffer := pw.buffer
	pw.buffer = make([]interface{}, 0, 1000)
	pw.bufferSize = 0
	pw.mu.Unlock()

	// Flush old buffer
	if blocking {
		// Synchronous flush (only on shutdown)
		pw.writeParquetFile(ctx, oldBuffer)
	} else {
		// ASYNCHRONOUS flush in background goroutine
		// This is the key to non-blocking behavior
		pw.flushWg.Add(1)
		go func() {
			defer pw.flushWg.Done()
			pw.writeParquetFile(ctx, oldBuffer)
		}()
	}
}

// writeParquetFile writes a buffer to a Parquet file
// This runs in a background goroutine for async flushes
func (pw *ParquetWriter) writeParquetFile(ctx context.Context, events []interface{}) {
	if len(events) == 0 {
		return
	}

	// Generate file path with hour-based folder
	now := time.Now()
	hourPath := filepath.Join(
		pw.basePath,
		fmt.Sprintf("year=%d", now.Year()),
		fmt.Sprintf("month=%02d", now.Month()),
		fmt.Sprintf("day=%02d", now.Day()),
		fmt.Sprintf("hour=%02d", now.Hour()),
	)

	if err := os.MkdirAll(hourPath, 0755); err != nil {
		log.Printf("ERROR [%s]: failed to create hour directory: %v", pw.logType, err)
		return
	}

	// Generate unique filename
	pw.mu.Lock()
	pw.fileCounter++
	fileNum := pw.fileCounter
	pw.mu.Unlock()

	filename := filepath.Join(
		hourPath,
		fmt.Sprintf("%s_%s_%d.parquet", pw.config.FilePrefix, pw.logType, fileNum),
	)

	// Create Parquet file
	file, err := os.Create(filename)
	if err != nil {
		log.Printf("ERROR [%s]: failed to create file: %v", pw.logType, err)
		return
	}
	defer file.Close()

	// Get schema and create writer based on log_type
	var parquetSchema *parquet.Schema
	switch pw.logType {
	case "dns":
		parquetSchema = parquet.SchemaOf(new(schema.DNS))
	case "conn":
		parquetSchema = parquet.SchemaOf(new(schema.CONN))
	default:
		log.Printf("ERROR [%s]: unknown log_type for Parquet schema", pw.logType)
		return
	}

	// Create Parquet writer
	pwWriter := parquet.NewWriter(file, parquetSchema)
	defer pwWriter.Close()

	// Bulk write - convert all events to rows and write in batch
	rows := make([]parquet.Row, 0, len(events))
	for _, event := range events {
		row := parquetSchema.Deconstruct(nil, event)
		rows = append(rows, row)
	}

	// Write all rows in a single batch (efficient)
	if _, err := pwWriter.WriteRows(rows); err != nil {
		log.Printf("ERROR [%s]: failed to write rows: %v", pw.logType, err)
		return
	}

	// Close writer (flushes data)
	if err := pwWriter.Close(); err != nil {
		log.Printf("ERROR [%s]: failed to close writer: %v", pw.logType, err)
		return
	}

	log.Printf("INFO [%s]: flushed %d events to %s", pw.logType, len(events), filename)
}

// convertEvent converts an EnrichedEvent to the appropriate Parquet struct
func (pw *ParquetWriter) convertEvent(event *EnrichedEvent) interface{} {
	ingestTime := schema.ToParquetTime(time.Now())

	switch pw.logType {
	case "dns":
		return &schema.DNS{
			EventTime:        event.EventTime,
			IngestTime:       ingestTime,
			EventYear:        event.EventYear,
			EventMonth:       event.EventMonth,
			EventDay:         event.EventDay,
			EventHour:        event.EventHour,
			EventWeekday:     event.EventWeekday,
			EventType:        "dns",
			EventClass:       "dns",
			SrcIP:            event.SrcIP,
			DstIP:            event.DstIP,
			SrcIPIsPrivate:   event.SrcIPIsPrivate,
			DstIPIsPrivate:   event.DstIPIsPrivate,
			Direction:        event.Direction,
			Service:          event.Service,
			FlowID:           event.FlowID,
			DNSQuery:         event.DNSQuery,
			QType:            event.QType,
			RCode:            event.RCode,
			RawLog:           event.RawLog,
		}

	case "conn":
		return &schema.CONN{
			EventTime:        event.EventTime,
			IngestTime:       ingestTime,
			EventYear:        event.EventYear,
			EventMonth:       event.EventMonth,
			EventDay:         event.EventDay,
			EventHour:        event.EventHour,
			EventWeekday:     event.EventWeekday,
			EventType:        "network_connection",
			EventClass:       "network",
			SrcIP:            event.SrcIP,
			DstIP:            event.DstIP,
			SrcPort:          event.SrcPort,
			DstPort:          event.DstPort,
			SrcIPIsPrivate:   event.SrcIPIsPrivate,
			DstIPIsPrivate:   event.DstIPIsPrivate,
			Direction:        event.Direction,
			Service:          event.Service,
			FlowID:           event.FlowID,
			Protocol:         event.Protocol,
			ConnState:        event.ConnState,
			RawLog:           event.RawLog,
		}

	default:
		log.Printf("WARNING: unknown log_type: %s", pw.logType)
		return nil
	}
}
