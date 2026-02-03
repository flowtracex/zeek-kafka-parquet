package core

import (
	"context"
	"sync"
)

// FanOut distributes events to multiple output channels
// Used to send the same event to multiple outputs (e.g., Parquet and Kafka)
type FanOut struct {
	outputs []chan *EnrichedEvent
	mu      sync.RWMutex
}

// GetOutput returns the output channel at index (for accessing specific outputs)
func (fo *FanOut) GetOutput(index int) chan *EnrichedEvent {
	fo.mu.RLock()
	defer fo.mu.RUnlock()
	if index >= 0 && index < len(fo.outputs) {
		return fo.outputs[index]
	}
	return nil
}

// NewFanOut creates a new fan-out component
func NewFanOut() *FanOut {
	return &FanOut{
		outputs: make([]chan *EnrichedEvent, 0),
	}
}

// AddOutput adds an output channel to the fan-out
func (fo *FanOut) AddOutput(output chan *EnrichedEvent) {
	fo.mu.Lock()
	defer fo.mu.Unlock()
	fo.outputs = append(fo.outputs, output)
}

// Start begins fanning out events from input to all outputs
func (fo *FanOut) Start(ctx context.Context, input <-chan *EnrichedEvent) {
	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-input:
			if !ok {
				return
			}

			// Send to all outputs (non-blocking)
			fo.mu.RLock()
			outputs := fo.outputs
			fo.mu.RUnlock()

			for _, output := range outputs {
				select {
				case output <- event:
					// Successfully sent
				case <-ctx.Done():
					return
				default:
					// Channel full - skip this output (should not happen with proper buffering)
					// In production, you might want to log this or use a different strategy
				}
			}
		}
	}
}

