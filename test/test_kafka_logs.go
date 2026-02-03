package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

func main() {
	const (
		broker = "localhost:9092"
		topic  = "zeek-logs"
	)

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{broker},
		Topic:   topic,
		GroupID: fmt.Sprintf("test-reader-%d", time.Now().Unix()),
	})
	defer reader.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	stats := struct {
		DNSMessages  int
		CONNMessages int
		DNSBytes     int64
		CONNBytes    int64
		OtherMessages int
		Errors       int
		TotalMessages int
	}{}

	startTime := time.Now()
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	fmt.Printf("🔍 Testing Kafka topic: %s\n", topic)
	fmt.Printf("⏱️  Running for 1 minute...\n\n")

	// Start a goroutine to print periodic updates
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				elapsed := time.Since(startTime).Seconds()
				fmt.Printf("[%.0fs] DNS: %d msgs (%.2f KB) | CONN: %d msgs (%.2f KB) | Other: %d | Total: %d\n",
					elapsed,
					stats.DNSMessages, float64(stats.DNSBytes)/1024,
					stats.CONNMessages, float64(stats.CONNBytes)/1024,
					stats.OtherMessages, stats.TotalMessages)
			}
		}
	}()

	// Consume messages
	for {
		select {
		case <-ctx.Done():
			goto done
		default:
			msg, err := reader.FetchMessage(ctx)
			if err != nil {
				if err == context.DeadlineExceeded {
					goto done
				}
				stats.Errors++
				if stats.Errors <= 5 {
					log.Printf("Error fetching message: %v", err)
				}
				continue
			}

			stats.TotalMessages++
			msgSize := int64(len(msg.Value))

			// Try to parse and identify log type
			var data map[string]interface{}
			if err := json.Unmarshal(msg.Value, &data); err != nil {
				stats.OtherMessages++
				continue
			}

			// Extract log_type
			logType, ok := data["log_type"].(string)
			if !ok {
				// Try _path field (Zeek sometimes uses this)
				if path, ok := data["_path"].(string); ok {
					logType = path
				} else {
					stats.OtherMessages++
					// Show first unknown message for debugging
					if stats.OtherMessages == 1 {
						preview := string(msg.Value)
						if len(preview) > 200 {
							preview = preview[:200] + "..."
						}
						fmt.Printf("⚠️  Unknown message format (first 200 chars): %s\n", preview)
					}
					continue
				}
			}

			// Categorize by log type
			switch logType {
			case "dns":
				stats.DNSMessages++
				stats.DNSBytes += msgSize
			case "conn":
				stats.CONNMessages++
				stats.CONNBytes += msgSize
			default:
				stats.OtherMessages++
				if stats.OtherMessages <= 3 {
					fmt.Printf("⚠️  Unknown log_type: %s\n", logType)
				}
			}

			// Commit message
			if err := reader.CommitMessages(ctx, msg); err != nil {
				stats.Errors++
			}
		}
	}

done:
	elapsed := time.Since(startTime)
	fmt.Printf("\n" + repeat("=", 60) + "\n")
	fmt.Printf("📊 FINAL STATISTICS (after %.1f seconds)\n", elapsed.Seconds())
	fmt.Printf(repeat("=", 60) + "\n")
	fmt.Printf("DNS Logs:\n")
	fmt.Printf("  Messages: %d\n", stats.DNSMessages)
	fmt.Printf("  Total Size: %.2f KB (%.2f MB)\n", float64(stats.DNSBytes)/1024, float64(stats.DNSBytes)/(1024*1024))
	if stats.DNSMessages > 0 {
		fmt.Printf("  Avg Size: %.2f bytes/message\n", float64(stats.DNSBytes)/float64(stats.DNSMessages))
	}
	fmt.Printf("\nCONN Logs:\n")
	fmt.Printf("  Messages: %d\n", stats.CONNMessages)
	fmt.Printf("  Total Size: %.2f KB (%.2f MB)\n", float64(stats.CONNBytes)/1024, float64(stats.CONNBytes)/(1024*1024))
	if stats.CONNMessages > 0 {
		fmt.Printf("  Avg Size: %.2f bytes/message\n", float64(stats.CONNBytes)/float64(stats.CONNMessages))
	}
	fmt.Printf("\nOther Messages: %d\n", stats.OtherMessages)
	fmt.Printf("Total Messages: %d\n", stats.TotalMessages)
	fmt.Printf("Errors: %d\n", stats.Errors)

	// Summary
	fmt.Printf("\n" + repeat("=", 60) + "\n")
	if stats.DNSMessages > 0 && stats.CONNMessages > 0 {
		fmt.Printf("✅ SUCCESS: Both DNS and CONN logs are being received!\n")
	} else if stats.DNSMessages > 0 {
		fmt.Printf("⚠️  WARNING: Only DNS logs received (no CONN logs)\n")
	} else if stats.CONNMessages > 0 {
		fmt.Printf("⚠️  WARNING: Only CONN logs received (no DNS logs)\n")
	} else if stats.TotalMessages > 0 {
		fmt.Printf("⚠️  WARNING: Messages received but no DNS/CONN logs detected\n")
	} else {
		fmt.Printf("❌ ERROR: No messages received from Kafka!\n")
		fmt.Printf("   Check:\n")
		fmt.Printf("   - Kafka is running: localhost:9092\n")
		fmt.Printf("   - Topic exists: %s\n", topic)
		fmt.Printf("   - Zeek is sending logs to Kafka\n")
	}
	fmt.Printf(repeat("=", 60) + "\n")
}

// Helper function to repeat strings
func repeat(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}

