package core

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/parquet-go/parquet-go"

	"pipeline/schema"
)

// WriterConfig holds writer configuration
type WriterConfig struct {
	BasePath         string
	FilePrefix       string
	FlushBufferMB    int
	FlushIntervalSec int
	FlushEventCount  int
}

// ParquetWriter writes events to Parquet files
// One writer per log_type with async flush
type ParquetWriter struct {
	logType         string
	config          WriterConfig
	input           <-chan *EnrichedEvent
	buffer          []interface{} // Current in-memory buffer
	bufferSize      int64          // Current buffer size in bytes (approximate)
	bufferLimit     int64          // Buffer limit in bytes
	eventCount      int            // Current event count in buffer
	eventCountLimit int            // Event count limit for flushing
	lastFlushTime   time.Time      // Last flush time for time-based flushing
	flushInterval   time.Duration  // Time interval for flushing
	mu              sync.Mutex     // Protects buffer, bufferSize, eventCount, lastFlushTime
	flushWg         sync.WaitGroup // Tracks background flush operations
	fileCounter     int
	basePath        string
	normalizer      *Normalizer    // Reference to normalizer for static fields (needs to be exported or we pass rules)
}

// NewParquetWriter creates a new Parquet writer for a log_type
func NewParquetWriter(logType string, cfg WriterConfig, input <-chan *EnrichedEvent) *ParquetWriter {
	bufferLimit := int64(cfg.FlushBufferMB) * 1024 * 1024 // Convert MB to bytes
	flushInterval := time.Duration(cfg.FlushIntervalSec) * time.Second

	return &ParquetWriter{
		logType:         logType,
		config:          cfg,
		input:           input,
		buffer:          make([]interface{}, 0, 1000),
		bufferLimit:     bufferLimit,
		eventCountLimit: cfg.FlushEventCount,
		flushInterval:  flushInterval,
		lastFlushTime:   time.Now(),
		basePath:        filepath.Join(cfg.BasePath, logType),
	}
}

// Start begins the writer goroutine
// This is the ONLY writer goroutine for this log_type
func (pw *ParquetWriter) Start(ctx context.Context) error {
	// Create output directory
	if err := os.MkdirAll(pw.basePath, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Start time-based flush ticker (if interval > 0)
	var ticker *time.Ticker
	var tickerChan <-chan time.Time
	if pw.flushInterval > 0 {
		ticker = time.NewTicker(pw.flushInterval)
		tickerChan = ticker.C
		defer ticker.Stop()
	}

	// Main writer loop
	for {
		select {
		case <-ctx.Done():
			// Flush remaining buffer before shutdown
			pw.flushBuffer(ctx, true)
			pw.flushWg.Wait() // Wait for all background flushes to complete
			return nil

		case <-tickerChan:
			// Time-based flush trigger
			pw.mu.Lock()
			shouldFlush := len(pw.buffer) > 0
			pw.mu.Unlock()
			if shouldFlush {
				pw.flushBuffer(ctx, false)
			}

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
			pw.eventCount++

			// Check flush conditions (OR logic - any one can trigger)
			shouldFlush := false
			
			// Condition 1: Size-based flush
			if pw.bufferSize >= pw.bufferLimit {
				shouldFlush = true
			}
			
			// Condition 2: Event count-based flush
			if pw.eventCountLimit > 0 && pw.eventCount >= pw.eventCountLimit {
				shouldFlush = true
			}
			
			// Note: Time-based flush is handled by the ticker above
			
			pw.mu.Unlock()

			// CRITICAL: Async flush when any condition is met
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
	oldEventCount := pw.eventCount
	pw.buffer = make([]interface{}, 0, 1000)
	pw.bufferSize = 0
	pw.eventCount = 0
	pw.lastFlushTime = time.Now()
	pw.mu.Unlock()
	
	// Log flush reason for debugging (if multiple conditions met, log all)
	_ = oldEventCount // Use event count for potential future logging

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

	// Get schema and create writer based on log_type using reflection
	parquetSchema, err := pw.getSchemaForLogType()
	if err != nil {
		log.Printf("ERROR [%s]: failed to get schema: %v", pw.logType, err)
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

// getSchemaForLogType returns the Parquet schema for the log type using reflection
func (pw *ParquetWriter) getSchemaForLogType() (*parquet.Schema, error) {
	structType := pw.getStructTypeForLogType()
	if structType == nil {
		return nil, fmt.Errorf("unknown log_type: %s", pw.logType)
	}
	
	// Create a zero value instance to get the schema
	instance := reflect.New(structType).Interface()
	return parquet.SchemaOf(instance), nil
}

// getStructTypeForLogType returns the reflect.Type for the log type's struct
func (pw *ParquetWriter) getStructTypeForLogType() reflect.Type {
	// Convert log_type to struct name (e.g., "dns" -> "DNS", "dce_rpc" -> "DCE_RPC")
	structName := pw.logTypeToStructName(pw.logType)
	return pw.getStructTypeByName(structName)
}

// logTypeToStructName converts log_type to struct name
func (pw *ParquetWriter) logTypeToStructName(logType string) string {
	// Handle special cases and convert to uppercase with underscores
	parts := strings.Split(logType, "_")
	var result strings.Builder
	for _, part := range parts {
		if len(part) > 0 {
			result.WriteString(strings.ToUpper(part[:1]))
			if len(part) > 1 {
				result.WriteString(strings.ToUpper(part[1:]))
			}
		}
	}
	return result.String()
}

// getStructTypeByName gets the struct type by name using reflection
func (pw *ParquetWriter) getStructTypeByName(structName string) reflect.Type {
	// Use a type registry approach - check known types
	schemaTypes := map[string]reflect.Type{
		"DNS":        reflect.TypeOf(schema.DNS{}),
		"CONN":       reflect.TypeOf(schema.CONN{}),
		"HTTP":       reflect.TypeOf(schema.HTTP{}),
		"SSL":        reflect.TypeOf(schema.SSL{}),
		"SSH":        reflect.TypeOf(schema.SSH{}),
		"FTP":        reflect.TypeOf(schema.FTP{}),
		"SMTP":       reflect.TypeOf(schema.SMTP{}),
		"DHCP":       reflect.TypeOf(schema.DHCP{}),
		"RDP":        reflect.TypeOf(schema.RDP{}),
		"SMB_FILES":  reflect.TypeOf(schema.SMB_FILES{}),
		"SMB_MAPPING": reflect.TypeOf(schema.SMB_MAPPING{}),
		"DCE_RPC":    reflect.TypeOf(schema.DCE_RPC{}),
		"KERBEROS":   reflect.TypeOf(schema.KERBEROS{}),
		"NTLM":       reflect.TypeOf(schema.NTLM{}),
		"SIP":        reflect.TypeOf(schema.SIP{}),
		"SNMP":       reflect.TypeOf(schema.SNMP{}),
		"RADIUS":     reflect.TypeOf(schema.RADIUS{}),
		"TUNNEL":     reflect.TypeOf(schema.TUNNEL{}),
	}
	
	if t, ok := schemaTypes[structName]; ok {
		return t
	}
	return nil
}

// convertEvent converts an EnrichedEvent to the appropriate Parquet struct
// Uses reflection to work with any log type
func (pw *ParquetWriter) convertEvent(event *EnrichedEvent) interface{} {
	ingestTime := schema.ToParquetTime(time.Now())
	
	// Get the struct type for this log type
	structType := pw.getStructTypeForLogType()
	if structType == nil {
		log.Printf("WARNING: unknown log_type: %s", pw.logType)
		return nil
	}
	
	// Create a new instance of the struct
	instance := reflect.New(structType).Elem()
	
	// Populate the struct using reflection
	pw.populateStruct(instance, event, ingestTime)
	
	return instance.Addr().Interface()
}

// populateStruct populates a struct instance with data from EnrichedEvent
func (pw *ParquetWriter) populateStruct(structValue reflect.Value, event *EnrichedEvent, ingestTime int64) {
	// Iterate through struct fields and populate them
	structType := structValue.Type()
	for i := 0; i < structValue.NumField(); i++ {
		field := structType.Field(i)
		fieldValue := structValue.Field(i)
		
		// Skip unexported fields
		if !fieldValue.CanSet() {
			continue
		}
		
		// Get the Parquet tag to find the field name
		parquetTag := field.Tag.Get("parquet")
		if parquetTag == "" {
			continue
		}
		
		// Map Parquet field names to values
		switch parquetTag {
		case "ingest_time":
			fieldValue.SetInt(ingestTime)
		case "event_year":
			fieldValue.SetInt(int64(event.EventYear))
		case "event_month":
			fieldValue.SetInt(int64(event.EventMonth))
		case "event_day":
			fieldValue.SetInt(int64(event.EventDay))
		case "event_hour":
			fieldValue.SetInt(int64(event.EventHour))
		case "event_weekday":
			fieldValue.SetInt(int64(event.EventWeekday))
		case "event_type":
			// Get from static fields stored in NormalizedEvent
			fieldValue.SetString(event.EventType)
		case "event_class":
			// Get from static fields stored in NormalizedEvent
			fieldValue.SetString(event.EventClass)
		case "src_ip_is_private":
			fieldValue.SetBool(event.SrcIPIsPrivate)
		case "dst_ip_is_private":
			fieldValue.SetBool(event.DstIPIsPrivate)
		case "direction":
			fieldValue.SetString(event.Direction)
		case "service":
			fieldValue.SetString(event.Service)
		case "event_time":
			fieldValue.SetInt(event.EventTime)
		case "flow_id":
			fieldValue.SetString(event.FlowID)
		case "src_ip":
			fieldValue.SetString(event.SrcIP)
		case "dst_ip":
			fieldValue.SetString(event.DstIP)
		case "src_port":
			fieldValue.SetInt(int64(event.SrcPort))
		case "dst_port":
			fieldValue.SetInt(int64(event.DstPort))
		case "protocol":
			fieldValue.SetString(event.Protocol)
		case "raw_log":
			fieldValue.SetString(event.RawLog)
		default:
			// Try to get from ZeekFields using the Parquet field name
			// Convert parquet name back to Zeek field name if needed
			zeekFieldName := pw.parquetNameToZeekField(parquetTag)
			if zeekFieldName != "" {
				// Try to set the value based on field type
				if v, ok := event.ZeekFields[zeekFieldName]; ok {
					pw.setFieldValue(fieldValue, v)
				}
			}
		}
	}
}

// parquetNameToZeekField converts Parquet field name to Zeek field name
func (pw *ParquetWriter) parquetNameToZeekField(parquetName string) string {
	// Convert Parquet field names back to Zeek field names
	// Most match, but some need conversion (e.g., id_orig_p -> id.orig_p)
	zeekFieldMap := map[string]string{
		"id_orig_p": "id.orig_p",
		"id_orig_h": "id.orig_h",
		"id_resp_p": "id.resp_p",
		"id_resp_h": "id.resp_h",
	}
	
	if zeekName, ok := zeekFieldMap[parquetName]; ok {
		return zeekName
	}
	return parquetName
}

// setFieldValue sets a field value based on its type
func (pw *ParquetWriter) setFieldValue(fieldValue reflect.Value, value interface{}) {
	switch fieldValue.Kind() {
	case reflect.String:
		// Handle arrays/slices by converting to JSON string
		if arr, ok := value.([]interface{}); ok {
			jsonBytes, err := json.Marshal(arr)
			if err == nil {
				fieldValue.SetString(string(jsonBytes))
			} else {
				// Fallback: convert to comma-separated string
				parts := make([]string, len(arr))
				for i, v := range arr {
					parts[i] = fmt.Sprintf("%v", v)
				}
				fieldValue.SetString(strings.Join(parts, ","))
			}
		} else if s, ok := value.(string); ok {
			fieldValue.SetString(s)
		}
	case reflect.Int32:
		switch v := value.(type) {
		case float64:
			fieldValue.SetInt(int64(v))
		case int:
			fieldValue.SetInt(int64(v))
		case int32:
			fieldValue.SetInt(int64(v))
		}
	case reflect.Int64:
		switch v := value.(type) {
		case float64:
			fieldValue.SetInt(int64(v))
		case int:
			fieldValue.SetInt(int64(v))
		case int64:
			fieldValue.SetInt(v)
		}
	case reflect.Float64:
		if f, ok := value.(float64); ok {
			fieldValue.SetFloat(f)
		}
	case reflect.Bool:
		if b, ok := value.(bool); ok {
			fieldValue.SetBool(b)
		}
	}
}
