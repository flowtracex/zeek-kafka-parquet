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

	// DNS-specific fields
	DNSQuery string
	QType    string
	RCode    int32

	// CONN-specific fields
	Protocol  string
	ConnState string
}

// Normalizer maps Zeek logs to normalized events
type Normalizer struct {
	normalizationRules map[string]NormalizationRule
}

// NormalizationRule defines how to map Zeek fields
type NormalizationRule struct {
	Source string            `json:"source"`
	Map    map[string]string `json:"map"`
	Static map[string]string `json:"static"`
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
	}

	// Copy all original Zeek fields
	for k, v := range zeekLog.Data {
		event.ZeekFields[k] = v
	}

	// Apply field mappings
	for zeekField, normField := range rule.Map {
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
		case "dns_query":
			if query, ok := value.(string); ok {
				event.DNSQuery = query
			}
		case "qtype":
			if qtype, ok := value.(string); ok {
				event.QType = qtype
			}
		case "rcode":
			if rcode, ok := value.(float64); ok {
				event.RCode = int32(rcode)
			} else if rcode, ok := value.(int); ok {
				event.RCode = int32(rcode)
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

