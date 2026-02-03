package core

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/compress"
)

// KafkaOutputEvent represents the structured output format for Kafka
type KafkaOutputEvent struct {
	Source  string                 `json:"source"`
	LogType string                 `json:"log_type"`
	Raw     map[string]interface{} `json:"raw"`
	Normalized struct {
		EventTime  int64  `json:"event_time"`
		IngestTime int64  `json:"ingest_time"`
		FlowID     string `json:"flow_id"`
		SrcIP      string `json:"src_ip"`
		DstIP      string `json:"dst_ip"`
		SrcPort    int32  `json:"src_port"`
		DstPort    int32  `json:"dst_port"`
		Protocol   string `json:"protocol"`
		Service    string `json:"service"`
		EventType  string `json:"event_type"`
		EventClass string `json:"event_class"`
	} `json:"normalized"`
	Enriched struct {
		SrcIPIsPrivate bool   `json:"src_ip_is_private"`
		DstIPIsPrivate bool   `json:"dst_ip_is_private"`
		Direction      string `json:"direction"`
		EventYear      int32  `json:"event_year"`
		EventMonth     int32  `json:"event_month"`
		EventDay       int32  `json:"event_day"`
		EventHour      int32  `json:"event_hour"`
		EventWeekday   int32  `json:"event_weekday"`
		EnrichTime     bool   `json:"enrich_time"`
		EnrichNetwork  bool   `json:"enrich_network"`
	} `json:"enriched"`
}

// transformToKafkaOutput converts an EnrichedEvent to KafkaOutputEvent format
func transformToKafkaOutput(event *EnrichedEvent) *KafkaOutputEvent {
	output := &KafkaOutputEvent{
		Source:  "zeek",
		LogType: event.LogType,
		Raw:     make(map[string]interface{}),
	}

	// Copy raw Zeek fields (one level only, no nesting)
	if event.ZeekFields != nil {
		for k, v := range event.ZeekFields {
			output.Raw[k] = v
		}
	}

	// Populate normalized fields
	output.Normalized.EventTime = event.EventTime
	output.Normalized.IngestTime = event.IngestTime
	output.Normalized.FlowID = event.FlowID
	output.Normalized.SrcIP = event.SrcIP
	output.Normalized.DstIP = event.DstIP
	output.Normalized.SrcPort = event.SrcPort
	output.Normalized.DstPort = event.DstPort
	output.Normalized.Protocol = event.Protocol
	
	// Use raw "service" field if available, otherwise use enriched Service
	if rawService, ok := event.ZeekFields["service"]; ok {
		if serviceStr, ok := rawService.(string); ok {
			output.Normalized.Service = serviceStr
		}
	} else {
		output.Normalized.Service = event.Service
	}
	
	output.Normalized.EventType = event.EventType
	output.Normalized.EventClass = event.EventClass

	// Populate enriched fields
	output.Enriched.SrcIPIsPrivate = event.SrcIPIsPrivate
	output.Enriched.DstIPIsPrivate = event.DstIPIsPrivate
	output.Enriched.Direction = event.Direction
	output.Enriched.EventYear = event.EventYear
	output.Enriched.EventMonth = event.EventMonth
	output.Enriched.EventDay = event.EventDay
	output.Enriched.EventHour = event.EventHour
	output.Enriched.EventWeekday = event.EventWeekday
	output.Enriched.EnrichTime = event.EnrichTime
	output.Enriched.EnrichNetwork = event.EnrichNetwork

	return output
}

// KafkaProducerConfig holds Kafka producer configuration
type KafkaProducerConfig struct {
	Brokers     []string
	Topic       string
	Compression string // "none", "gzip", "snappy", "lz4", "zstd"
}

// KafkaProducer writes enriched events to Kafka
type KafkaProducer struct {
	config    KafkaProducerConfig
	writer    *kafka.Writer
	input     <-chan *EnrichedEvent
	errorChan chan error
	mu        sync.Mutex
	sentCount int64 // Total events sent (for health monitoring)
	errorCount int64 // Total errors (for health monitoring)
}

// NewKafkaProducer creates a new Kafka producer
func NewKafkaProducer(cfg KafkaProducerConfig, input <-chan *EnrichedEvent) (*KafkaProducer, error) {
	// Configure compression
	var compression compress.Compression
	switch cfg.Compression {
	case "gzip":
		compression = compress.Gzip
	case "snappy":
		compression = compress.Snappy
	case "lz4":
		compression = compress.Lz4
	case "zstd":
		compression = compress.Zstd
	default:
		compression = compress.None // No compression
	}

	writer := &kafka.Writer{
		Addr:         kafka.TCP(cfg.Brokers...),
		Topic:        cfg.Topic,
		Balancer:     &kafka.LeastBytes{},
		Compression:  compression,
		BatchSize:    100,  // Batch messages for efficiency
		BatchTimeout: 10e6, // 10ms
		Async:        false, // Synchronous writes for reliability
	}

	return &KafkaProducer{
		config:    cfg,
		writer:    writer,
		input:     input,
		errorChan: make(chan error, 100),
	}, nil
}

// Start begins the producer goroutine
func (kp *KafkaProducer) Start(ctx context.Context) error {
	logger := GetLogger()
	logger.Info("startup", "kafka producer started",
		fmt.Sprintf("topic=%s brokers=%v compression=%s", kp.config.Topic, kp.config.Brokers, kp.config.Compression))

	for {
		select {
		case <-ctx.Done():
			// Close writer on shutdown
			if err := kp.writer.Close(); err != nil {
				logger.Error("kafka", "producer close failed", fmt.Sprintf("error=%v", err))
			}
			return nil

		case event, ok := <-kp.input:
			if !ok {
				// Channel closed
				if err := kp.writer.Close(); err != nil {
					logger.Error("kafka", "producer close failed", fmt.Sprintf("error=%v", err))
				}
				return nil
			}

			// Transform to Kafka output format
			kafkaEvent := transformToKafkaOutput(event)

			// Serialize to JSON
			jsonData, err := json.Marshal(kafkaEvent)
			if err != nil {
				atomic.AddInt64(&kp.errorCount, 1)
				kp.errorChan <- fmt.Errorf("serialize event failed: %w", err)
				continue
			}

			// Write to Kafka
			msg := kafka.Message{
				Key:   []byte(event.LogType), // Use log_type as key for partitioning
				Value: jsonData,
			}

			if err := kp.writer.WriteMessages(ctx, msg); err != nil {
				atomic.AddInt64(&kp.errorCount, 1)
				kp.errorChan <- fmt.Errorf("kafka write failed: %w", err)
				continue
			}

			atomic.AddInt64(&kp.sentCount, 1)
		}
	}
}

// Errors returns the error channel
func (kp *KafkaProducer) Errors() <-chan error {
	return kp.errorChan
}

// GetMetrics returns producer metrics (thread-safe)
func (kp *KafkaProducer) GetMetrics() (sentCount, errorCount int64) {
	return atomic.LoadInt64(&kp.sentCount), atomic.LoadInt64(&kp.errorCount)
}

