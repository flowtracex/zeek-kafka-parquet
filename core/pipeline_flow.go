package core

import (
	"context"
	"fmt"
)

// PipelineFlow routes enriched events by log_type to downstream processors
// Each log_type has its own channel for parallel processing
type PipelineFlow struct {
	routes map[string]chan *EnrichedEvent
}

// NewPipelineFlow creates a new pipeline flow router
func NewPipelineFlow(logTypes []string, bufferSize int) *PipelineFlow {
	routes := make(map[string]chan *EnrichedEvent)
	for _, logType := range logTypes {
		routes[logType] = make(chan *EnrichedEvent, bufferSize)
	}

	return &PipelineFlow{
		routes: routes,
	}
}

// Route sends an event to the appropriate channel based on log_type
func (pf *PipelineFlow) Route(event *EnrichedEvent) error {
	ch, ok := pf.routes[event.LogType]
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
func (pf *PipelineFlow) GetRoute(logType string) <-chan *EnrichedEvent {
	return pf.routes[logType]
}

// Start begins routing events from input channel
func (pf *PipelineFlow) Start(ctx context.Context, input <-chan *EnrichedEvent) {
	logger := GetLogger()
	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-input:
			if !ok {
				return
			}
			if err := pf.Route(event); err != nil {
				// Log error but continue processing
				logger.Error("pipeline", "routing error", fmt.Sprintf("log_type=%s error=%v", event.LogType, err))
			}
		}
	}
}

