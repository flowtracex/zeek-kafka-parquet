package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

// Schema definitions
type SchemaDef struct {
	LogType string                 `json:"log_type"`
	Fields  map[string]string     `json:"fields"`
}

type SchemaFile struct {
	Schemas map[string]SchemaDef `json:"-"`
}

// Normalization definitions
type NormalizationRule struct {
	Source  string            `json:"source"`
	Promote map[string]string `json:"promote"`
	Static  map[string]string `json:"static"`
	Enrich  *EnrichConfig     `json:"enrich"`
}

type EnrichConfig struct {
	Time    bool `json:"time"`
	Network bool `json:"network"`
}

type NormalizationFile struct {
	Rules map[string]NormalizationRule `json:"-"`
}

func main() {
	// Read schema.json
	schemaData, err := os.ReadFile("config/schema.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading config/schema.json: %v\n", err)
		os.Exit(1)
	}

	var schemaRaw map[string]interface{}
	if err := json.Unmarshal(schemaData, &schemaRaw); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing config/schema.json: %v\n", err)
		os.Exit(1)
	}

	// Read normalization.json
	normData, err := os.ReadFile("config/normalization.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading config/normalization.json: %v\n", err)
		os.Exit(1)
	}

	var normRules map[string]NormalizationRule
	if err := json.Unmarshal(normData, &normRules); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing config/normalization.json: %v\n", err)
		os.Exit(1)
	}

	// Generate events.go
	output := generateEventsGo(schemaRaw, normRules)

	// Write to file
	if err := os.WriteFile("schema/events.go", []byte(output), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing schema/events.go: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Generated schema/events.go successfully")
}

func generateEventsGo(schemaRaw map[string]interface{}, normRules map[string]NormalizationRule) string {
	var buf strings.Builder

	// Header
	buf.WriteString("package schema\n\n")
	buf.WriteString("import (\n")
	buf.WriteString("\t\"time\"\n")
	buf.WriteString(")\n\n")

	// Generate struct for each log type in normalization.json
	logTypes := make([]string, 0, len(normRules))
	for logType := range normRules {
		logTypes = append(logTypes, logType)
	}
	sort.Strings(logTypes)

	for _, logType := range logTypes {
		rule := normRules[logType]
		sourceKey := rule.Source // e.g., "zeek_dns", "zeek_conn"

		// Find schema for this source
		schemaDef, ok := schemaRaw[sourceKey]
		if !ok {
			continue // Skip if schema not found
		}

		schemaMap, ok := schemaDef.(map[string]interface{})
		if !ok {
			continue
		}

		fieldsMap, ok := schemaMap["fields"].(map[string]interface{})
		if !ok {
			continue
		}

		// Get promoted field names (to exclude from raw)
		// Key is the normalized field name (e.g., "service", "direction")
		promotedFields := make(map[string]bool)
		for _, normField := range rule.Promote {
			promotedFields[normField] = true
		}

		// Get raw field names that are promoted (to exclude from raw section)
		promotedRawFields := make(map[string]bool)
		for rawField := range rule.Promote {
			promotedRawFields[rawField] = true
		}

		// Generate struct
		structName := strings.ToUpper(logType)
		buf.WriteString(fmt.Sprintf("// %s represents a normalized %s event following the three-layer model:\n", structName, logType))
		buf.WriteString("// 1) Conditional enrichment (controlled per log type via normalization.json enrich flags)\n")
		buf.WriteString("// 2) Normalized/promoted fields (from normalization.json mapping)\n")
		buf.WriteString("// 3) Unmapped raw fields (from schema.json, NOT in normalization.json)\n")
		buf.WriteString("// 4) raw_log (complete original JSON)\n")
		buf.WriteString(fmt.Sprintf("type %s struct {\n", structName))

		// Enrichment fields (conditional)
		hasTimeEnrich := rule.Enrich != nil && rule.Enrich.Time
		hasNetworkEnrich := rule.Enrich != nil && rule.Enrich.Network

		buf.WriteString("\t// =========================\n")
		buf.WriteString("\t// CONDITIONAL ENRICHMENT (Layer 3)\n")
		buf.WriteString("\t// Applied based on normalization.json enrich flags (time, network)\n")
		buf.WriteString("\t// Fields may be zero/empty if enrichment is disabled for this log type\n")
		buf.WriteString("\t// =========================\n")

		if hasTimeEnrich {
			buf.WriteString("\tIngestTime       int64  `parquet:\"ingest_time\"`\n")
			buf.WriteString("\tEventYear        int32  `parquet:\"event_year\"`\n")
			buf.WriteString("\tEventMonth       int32  `parquet:\"event_month\"`\n")
			buf.WriteString("\tEventDay         int32  `parquet:\"event_day\"`\n")
			buf.WriteString("\tEventHour        int32  `parquet:\"event_hour\"`\n")
			buf.WriteString("\tEventWeekday     int32  `parquet:\"event_weekday\"`\n")
		}

		// Static fields (always present)
		buf.WriteString("\tEventType        string `parquet:\"event_type\"`\n")
		buf.WriteString("\tEventClass       string `parquet:\"event_class\"`\n")

		if hasNetworkEnrich {
			buf.WriteString("\tSrcIPIsPrivate   bool   `parquet:\"src_ip_is_private\"`\n")
			buf.WriteString("\tDstIPIsPrivate   bool   `parquet:\"dst_ip_is_private\"`\n")
			// Only add Direction if it's not already promoted
			if !promotedFields["direction"] {
				buf.WriteString("\tDirection        string `parquet:\"direction\"`\n")
			}
			// Only add Service enrichment if it's not already promoted
			if !promotedFields["service"] {
				buf.WriteString("\tService          string `parquet:\"service\"`\n")
			}
		}

		// Promoted fields
		buf.WriteString("\n\t// =========================\n")
		buf.WriteString("\t// NORMALIZED (PROMOTED) FIELDS (Layer 2)\n")
		buf.WriteString("\t// From normalization.json mapping - these replace raw fields\n")
		buf.WriteString("\t// DO NOT include raw versions\n")
		buf.WriteString("\t// =========================\n")

		// Sort promoted fields for consistent output
		promotedPairs := make([]struct {
			rawField  string
			normField string
		}, 0, len(rule.Promote))
		for rawField, normField := range rule.Promote {
			promotedPairs = append(promotedPairs, struct {
				rawField  string
				normField string
			}{rawField, normField})
		}
		sort.Slice(promotedPairs, func(i, j int) bool {
			return promotedPairs[i].normField < promotedPairs[j].normField
		})

		for _, pair := range promotedPairs {
			goFieldName := toGoFieldName(pair.normField)
			parquetName := pair.normField
			goType := getGoTypeForPromotedField(pair.normField, fieldsMap[pair.rawField])
			comment := fmt.Sprintf("// Promoted from: %s", pair.rawField)
			buf.WriteString(fmt.Sprintf("\t%s %s `parquet:\"%s\"` %s\n", goFieldName, goType, parquetName, comment))
		}

		// Raw fields (unmapped)
		buf.WriteString("\n\t// =========================\n")
		buf.WriteString("\t// RAW FIELDS (UNMAPPED) (Layer 1)\n")
		buf.WriteString("\t// From schema.json, but NOT in normalization.json mapping\n")
		buf.WriteString("\t// These are raw fields that were NOT promoted\n")
		buf.WriteString("\t// =========================\n")

		// Get all raw fields, exclude promoted ones
		rawFieldNames := make([]string, 0)
		for fieldName := range fieldsMap {
			if !promotedRawFields[fieldName] {
				rawFieldNames = append(rawFieldNames, fieldName)
			}
		}
		sort.Strings(rawFieldNames)

		// Enrichment field names that might conflict with raw fields
		enrichmentGoNames := map[string]bool{
			"Service":  true,
			"Direction": true,
		}

		for _, fieldName := range rawFieldNames {
			fieldType, ok := fieldsMap[fieldName].(string)
			if !ok {
				continue
			}
			goFieldName := toGoFieldName(fieldName)
			parquetName := sanitizeParquetName(fieldName)
			
			// If raw field name conflicts with enrichment field, rename it
			if enrichmentGoNames[goFieldName] {
				goFieldName = "Zeek" + goFieldName // e.g., "Service" -> "ZeekService"
			}
			
			goType := getGoTypeForRawField(fieldName, fieldType)
			buf.WriteString(fmt.Sprintf("\t%s %s `parquet:\"%s\"` // Not mapped\n", goFieldName, goType, parquetName))
		}

		// Raw log
		buf.WriteString("\n\t// =========================\n")
		buf.WriteString("\t// FULL RAW LOG (complete original JSON)\n")
		buf.WriteString("\t// =========================\n")
		buf.WriteString("\tRawLog string `parquet:\"raw_log\"`\n")

		buf.WriteString("}\n\n")
	}

	// ToParquetTime helper
	buf.WriteString("// ToParquetTime converts Go time to milliseconds since epoch\n")
	buf.WriteString("func ToParquetTime(t time.Time) int64 {\n")
	buf.WriteString("\treturn t.UnixNano() / int64(time.Millisecond)\n")
	buf.WriteString("}\n")

	return buf.String()
}

func toGoFieldName(fieldName string) string {
	// Convert field names like "event_time" -> "EventTime", "id.orig_h" -> "IdOrigH"
	// Special handling for "IP" and "ID" at the END of compound names: "src_ip" -> "SrcIP", "flow_id" -> "FlowID"
	// But "id" at the start should be "Id": "id.orig_p" -> "IdOrigP"
	normalized := strings.ReplaceAll(fieldName, ".", "_")
	parts := strings.Split(normalized, "_")
	var result strings.Builder
	for i, part := range parts {
		if len(part) > 0 {
			upperPart := strings.ToUpper(part)
			// Handle "IP" and "ID" specially - uppercase when at the END of compound names
			if (upperPart == "IP" || upperPart == "ID") && i == len(parts)-1 && len(parts) > 1 {
				// Last part and it's IP/ID in a compound name
				result.WriteString(upperPart)
			} else {
				// Normal conversion: "id" -> "Id", "orig" -> "Orig", "src" -> "Src"
				result.WriteString(strings.ToUpper(part[:1]))
				if len(part) > 1 {
					result.WriteString(part[1:])
				}
			}
		}
	}
	return result.String()
}

func sanitizeParquetName(fieldName string) string {
	// Convert field names like "id.orig_h" -> "id_orig_h" for Parquet
	return strings.ReplaceAll(fieldName, ".", "_")
}

func getGoType(zeekType string) string {
	switch zeekType {
	case "float":
		return "float64"
	case "int":
		return "int32" // Most int fields are int32
	case "string":
		return "string"
	case "bool":
		return "bool"
	default:
		return "string" // Default to string for unknown types
	}
}

// getGoTypeForRawField returns the appropriate Go type for raw fields
// Some fields like orig_bytes, resp_bytes should be int64
func getGoTypeForRawField(fieldName, zeekType string) string {
	// Fields that should be int64
	int64Fields := map[string]bool{
		"orig_bytes":    true,
		"resp_bytes":    true,
		"missed_bytes":  true,
		"orig_pkts":     true,
		"orig_ip_bytes": true,
		"resp_pkts":     true,
		"resp_ip_bytes": true,
	}
	
	if int64Fields[fieldName] && zeekType == "int" {
		return "int64"
	}
	
	return getGoType(zeekType)
}

func getGoTypeForPromotedField(normField string, rawFieldType interface{}) string {
	// Special handling for promoted fields
	switch normField {
	case "event_time":
		return "int64" // Always int64 for timestamps
	case "src_port", "dst_port":
		return "int32"
	case "src_ip", "dst_ip", "flow_id", "protocol", "conn_state", "service":
		return "string"
	default:
		// Try to infer from raw field type
		if rawFieldType != nil {
			if typeStr, ok := rawFieldType.(string); ok {
				switch typeStr {
				case "float":
					return "int64" // event_time is converted from float to int64
				case "int":
					return "int32"
				default:
					return "string"
				}
			}
		}
		return "string"
	}
}

