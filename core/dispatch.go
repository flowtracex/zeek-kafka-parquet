package core

import (
	"context"
	"fmt"
)

// Dispatcher routes enriched events by log_type
// Each log_type has its own channel
type Dispatcher struct {
	routes map[string]chan *EnrichedEvent
}

// NewDispatcher creates a new dispatcher
func NewDispatcher(logTypes []string, bufferSize int) *Dispatcher {
	routes := make(map[string]chan *EnrichedEvent)
	for _, logType := range logTypes {
		routes[logType] = make(chan *EnrichedEvent, bufferSize)
	}

	return &Dispatcher{
		routes: routes,
	}
}

// Route sends an event to the appropriate channel based on log_type
func (d *Dispatcher) Route(event *EnrichedEvent) error {
	ch, ok := d.routes[event.LogType]
	if !ok {
		return fmt.Errorf("no route for log_type: %s", event.LogType)
	}

	select {
	case ch <- event:
		return nil
	default:
		// Channel full - this should be handled by proper buffering
		return fmt.Errorf("route channel full for log_type: %s", event.LogType)
	}
}

// GetRoute returns the channel for a specific log_type
func (d *Dispatcher) GetRoute(logType string) <-chan *EnrichedEvent {
	return d.routes[logType]
}

// Start begins routing events from input channel
func (d *Dispatcher) Start(ctx context.Context, input <-chan *EnrichedEvent) {
	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-input:
			if !ok {
				return
			}
			if err := d.Route(event); err != nil {
				// Log error but continue processing
				fmt.Printf("dispatch error: %v\n", err)
			}
		}
	}
}

