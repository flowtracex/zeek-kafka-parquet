package core

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

// NormalizedEvent represents a normalized event with common fields
type NormalizedEvent struct {
	EventTime  int64
	IngestTime int64
	LogType    string
	SrcIP      string
	DstIP      string
	SrcPort    int32
	DstPort    int32
	FlowID     string
	RawLog     string

	// Preserve all original Zeek fields as a map
	ZeekFields map[string]interface{}

	// Log-type-specific promoted fields (only if in normalization.json)
	Protocol  string // CONN: promoted from proto
	ConnState string // CONN: promoted from conn_state

	// Per-log enrichment flags (from normalization.json enrich section)
	EnrichTime    bool
	EnrichNetwork bool

	// Static fields from normalization.json
	EventType  string // From static.event_type
	EventClass string // From static.event_class
}

// Normalizer maps Zeek logs to normalized events
type Normalizer struct {
	normalizationRules map[string]NormalizationRule
}

// EnrichConfig defines per-log enrichment flags
type EnrichConfig struct {
	Time    bool `json:"time"`    // Time enrichment flag
	Network bool `json:"network"` // Network enrichment flag
}

// NormalizationRule defines how to promote Zeek fields
type NormalizationRule struct {
	Source  string            `json:"source"`
	Promote map[string]string `json:"promote"` // Field promotion (mapping)
	Static  map[string]string `json:"static"`
	Enrich  *EnrichConfig     `json:"enrich"` // Per-log enrichment flags (optional)
}

// NewNormalizer creates a normalizer from normalization.json
func NewNormalizer(rules map[string]NormalizationRule) *Normalizer {
	return &Normalizer{
		normalizationRules: rules,
	}
}

// Normalize converts a Zeek log to a normalized event
// Preserves ALL Zeek fields in ZeekFields map
func (n *Normalizer) Normalize(zeekLog *ZeekLog) (*NormalizedEvent, error) {
	rule, ok := n.normalizationRules[zeekLog.LogType]
	if !ok {
		return nil, fmt.Errorf("no normalization rule for log_type: %s", zeekLog.LogType)
	}

	event := &NormalizedEvent{
		LogType:    zeekLog.LogType,
		IngestTime: time.Now().UnixNano() / int64(time.Millisecond),
		ZeekFields: make(map[string]interface{}),
		RawLog:     zeekLog.Raw,
		// Set enrichment flags from rule (default to false if not specified)
		EnrichTime:    false,
		EnrichNetwork: false,
		// Set static fields from rule
		EventType:  zeekLog.LogType, // Default to log type
		EventClass: "unknown",        // Default
	}

	// Apply enrichment flags from normalization rule
	if rule.Enrich != nil {
		event.EnrichTime = rule.Enrich.Time
		event.EnrichNetwork = rule.Enrich.Network
	}

	// Apply static fields from normalization rule
	if eventType, ok := rule.Static["event_type"]; ok {
		event.EventType = eventType
	}
	if eventClass, ok := rule.Static["event_class"]; ok {
		event.EventClass = eventClass
	}

	// Copy all original Zeek fields
	for k, v := range zeekLog.Data {
		event.ZeekFields[k] = v
	}

	// Apply field promotion (mapping)
	// normalization.json defines: zeekField -> normField
	// Example: "qtype_name" -> "qtype" means read qtype_name from Zeek, store as qtype
	// Promoted fields replace raw fields (no duplication)
	for zeekField, normField := range rule.Promote {
		value, ok := zeekLog.Data[zeekField]
		if !ok {
			continue
		}

		switch normField {
		case "event_time":
			// Convert Zeek timestamp (float64 seconds) to milliseconds
			if ts, ok := value.(float64); ok {
				event.EventTime = int64(ts * 1000)
			}
		case "src_ip":
			if ip, ok := value.(string); ok {
				event.SrcIP = ip
			}
		case "dst_ip":
			if ip, ok := value.(string); ok {
				event.DstIP = ip
			}
		case "src_port":
			if port, ok := value.(float64); ok {
				event.SrcPort = int32(port)
			} else if port, ok := value.(int); ok {
				event.SrcPort = int32(port)
			}
		case "dst_port":
			if port, ok := value.(float64); ok {
				event.DstPort = int32(port)
			} else if port, ok := value.(int); ok {
				event.DstPort = int32(port)
			}
		case "flow_id":
			if uid, ok := value.(string); ok {
				event.FlowID = uid
			}
		case "protocol":
			if proto, ok := value.(string); ok {
				event.Protocol = proto
			}
		case "conn_state":
			if state, ok := value.(string); ok {
				event.ConnState = state
			}
		}
	}

	return event, nil
}

// LoadNormalizationRules loads rules from JSON
func LoadNormalizationRules(data []byte) (map[string]NormalizationRule, error) {
	var rules map[string]NormalizationRule
	if err := json.Unmarshal(data, &rules); err != nil {
		return nil, fmt.Errorf("failed to parse normalization rules: %w", err)
	}

	// Remove common_enrichment if present
	delete(rules, "common_enrichment")

	return rules, nil
}

// Helper to safely extract string from interface{}
func getString(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

// Helper to safely extract int32 from interface{}
func getInt32(v interface{}) int32 {
	switch val := v.(type) {
	case float64:
		return int32(val)
	case int:
		return int32(val)
	case int32:
		return val
	case string:
		if i, err := strconv.Atoi(val); err == nil {
			return int32(i)
		}
	}
	return 0
}

