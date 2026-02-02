package core

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/segmentio/kafka-go"
)

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// KafkaConfig holds Kafka connection settings
type KafkaConfig struct {
	Brokers       []string `json:"brokers"`
	Topic         string   `json:"topic"`
	GroupID       string   `json:"group_id"`
	MaxPollRecords int    `json:"max_poll_records"`
}

// ZeekLog represents a raw Zeek log from Kafka
type ZeekLog struct {
	LogType string                 `json:"log_type"`
	Data    map[string]interface{} `json:"data"`
	Raw     string                 `json:"raw"` // Original JSON string
}

// Consumer reads from Kafka and emits parsed Zeek logs
// Runs in a single goroutine with no IO operations
type Consumer struct {
	reader *kafka.Reader
	output chan *ZeekLog
	errors chan error
}

// NewConsumer creates a new Kafka consumer
func NewConsumer(cfg KafkaConfig) (*Consumer, error) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  cfg.Brokers,
		Topic:    cfg.Topic,
		GroupID:  cfg.GroupID,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
		// Start from earliest if this is a new consumer group
		// (Kafka will use stored offset if group already exists)
	})

	return &Consumer{
		reader: reader,
		output: make(chan *ZeekLog, 1000), // Buffered channel
		errors: make(chan error, 100),
	}, nil
}

// Start begins consuming from Kafka
// This runs in a single goroutine and does NO IO operations
func (c *Consumer) Start(ctx context.Context) {
	defer close(c.output)
	defer close(c.errors)
	defer c.reader.Close()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			// Use ReadMessage (auto-commits) like the working script
			msg, err := c.reader.ReadMessage(ctx)
			if err != nil {
				c.errors <- fmt.Errorf("kafka read error: %w", err)
				continue
			}

			// Parse Zeek log - handle nested format: {"dns": {...}} or {"conn": {...}}
			var rawData map[string]interface{}
			if err := json.Unmarshal(msg.Value, &rawData); err != nil {
				c.errors <- fmt.Errorf("parse error: %w (first 100 chars: %s)", err, string(msg.Value[:min(100, len(msg.Value))]))
				// ReadMessage auto-commits, so no need to commit manually
				continue
			}

			// Extract log_type from top-level key (e.g., "dns", "conn")
			// Zeek sends nested format: {"dns": {...}} or {"conn": {...}}
			var zeekLog ZeekLog
			var logData map[string]interface{}
			
			// Try to find log type as a top-level key
			for key, value := range rawData {
				// Check if this key looks like a log type and value is an object
				if dataMap, ok := value.(map[string]interface{}); ok {
					zeekLog.LogType = key
					logData = dataMap
					break
				}
			}

			// If not found in nested format, try flat format
			if zeekLog.LogType == "" {
				// Try flat format with log_type field
				if logType, ok := rawData["log_type"].(string); ok {
					zeekLog.LogType = logType
					logData = rawData
				} else if logType, ok := rawData["_path"].(string); ok {
					// Zeek sometimes uses _path field
					zeekLog.LogType = logType
					logData = rawData
				} else {
					// Unknown format
					msgPreview := string(msg.Value[:min(200, len(msg.Value))])
					c.errors <- fmt.Errorf("missing log_type in message (preview: %s)", msgPreview)
					continue
				}
			}

			// Store the actual log data
			zeekLog.Data = logData

			// Store raw JSON string
			zeekLog.Raw = string(msg.Value)

			// Emit parsed event (non-blocking if channel has capacity)
			// ReadMessage auto-commits, so no need to commit manually
			select {
			case c.output <- &zeekLog:
				// Message successfully emitted
			case <-ctx.Done():
				return
			default:
				// Channel full - this should not happen with proper buffering
				log.Printf("WARNING: output channel full, dropping message")
			}
		}
	}
}

// Output returns the channel of parsed Zeek logs
func (c *Consumer) Output() <-chan *ZeekLog {
	return c.output
}

// Errors returns the channel of errors
func (c *Consumer) Errors() <-chan error {
	return c.errors
}

