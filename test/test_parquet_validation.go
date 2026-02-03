package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"github.com/parquet-go/parquet-go"
	"pipeline/schema"
)

// FieldComparisonResult holds comparison results for a single field
type FieldComparisonResult struct {
	FieldName      string
	RawValue       interface{}
	ParquetValue   interface{}
	RawType        string
	ParquetType    string
	Match          bool
	Issue          string // "missing", "type_mismatch", "multi_value_collapsed", "truncated", etc.
	RawIsArray     bool
	ParquetIsString bool
}

// FileComparisonResult holds results for a single Parquet file
type FileComparisonResult struct {
	FilePath       string
	LogType        string
	TotalRows      int
	FieldsAnalyzed map[string]int // field -> count of rows with this field
	Issues         []FieldComparisonResult
	MissingFields  []string // Fields in raw JSON but not in Parquet
	ExtraFields    []string // Fields in Parquet but not in raw JSON (shouldn't happen for raw fields)
}

// Global results
var allResults []FileComparisonResult
var logTypeStats = make(map[string]map[string]int) // log_type -> issue_type -> count

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run test_parquet_validation.go <output_directory>")
		fmt.Println("Example: go run test_parquet_validation.go output")
		os.Exit(1)
	}

	outputDir := os.Args[1]
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		log.Fatalf("Directory does not exist: %s", outputDir)
	}

	fmt.Println("=" + strings.Repeat("=", 80))
	fmt.Println("PARQUET VALIDATION TEST")
	fmt.Println("=" + strings.Repeat("=", 80))
	fmt.Println()

	// Find all Parquet files
	parquetFiles := findParquetFiles(outputDir)
	fmt.Printf("Found %d Parquet files to analyze\n\n", len(parquetFiles))

	// Process each file
	for _, filePath := range parquetFiles {
		processParquetFile(filePath)
	}

	// Generate report
	generateReport()
}

func findParquetFiles(rootDir string) []string {
	var files []string
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".parquet") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		log.Fatalf("Error walking directory: %v", err)
	}
	return files
}

func processParquetFile(filePath string) {
	// Extract log type from path (e.g., output/dns/... -> dns)
	parts := strings.Split(filePath, string(os.PathSeparator))
	logType := ""
	for i, part := range parts {
		if part == "output" && i+1 < len(parts) {
			logType = parts[i+1]
			break
		}
	}

	if logType == "" {
		log.Printf("Warning: Could not determine log type from path: %s", filePath)
		return
	}

	fmt.Printf("Processing: %s (log_type: %s)\n", filePath, logType)

	// Get schema struct for this log type
	schemaInstance := getSchemaInstanceForLogType(logType)
	if schemaInstance == nil {
		log.Printf("Warning: No schema struct found for log_type: %s", logType)
		return
	}

	result := FileComparisonResult{
		FilePath:       filePath,
		LogType:        logType,
		FieldsAnalyzed: make(map[string]int),
		Issues:         []FieldComparisonResult{},
		MissingFields:  []string{},
		ExtraFields:    []string{},
	}

	// Read Parquet file directly into structs (sample first 100 rows)
	maxRows := 100
	var rows []interface{}
	var err error
	
	// Use ReadFile with a limit
	switch logType {
	case "dns":
		var dnsRows []schema.DNS
		dnsRows, err = parquet.ReadFile[schema.DNS](filePath)
		if err == nil {
			for i := 0; i < len(dnsRows) && i < maxRows; i++ {
				rows = append(rows, dnsRows[i])
			}
		}
	case "conn":
		var connRows []schema.CONN
		connRows, err = parquet.ReadFile[schema.CONN](filePath)
		if err == nil {
			for i := 0; i < len(connRows) && i < maxRows; i++ {
				rows = append(rows, connRows[i])
			}
		}
	case "http":
		var httpRows []schema.HTTP
		httpRows, err = parquet.ReadFile[schema.HTTP](filePath)
		if err == nil {
			for i := 0; i < len(httpRows) && i < maxRows; i++ {
				rows = append(rows, httpRows[i])
			}
		}
	case "ssl":
		var sslRows []schema.SSL
		sslRows, err = parquet.ReadFile[schema.SSL](filePath)
		if err == nil {
			for i := 0; i < len(sslRows) && i < maxRows; i++ {
				rows = append(rows, sslRows[i])
			}
		}
	case "ssh":
		var sshRows []schema.SSH
		sshRows, err = parquet.ReadFile[schema.SSH](filePath)
		if err == nil {
			for i := 0; i < len(sshRows) && i < maxRows; i++ {
				rows = append(rows, sshRows[i])
			}
		}
	case "ftp":
		var ftpRows []schema.FTP
		ftpRows, err = parquet.ReadFile[schema.FTP](filePath)
		if err == nil {
			for i := 0; i < len(ftpRows) && i < maxRows; i++ {
				rows = append(rows, ftpRows[i])
			}
		}
	case "smtp":
		var smtpRows []schema.SMTP
		smtpRows, err = parquet.ReadFile[schema.SMTP](filePath)
		if err == nil {
			for i := 0; i < len(smtpRows) && i < maxRows; i++ {
				rows = append(rows, smtpRows[i])
			}
		}
	case "dhcp":
		var dhcpRows []schema.DHCP
		dhcpRows, err = parquet.ReadFile[schema.DHCP](filePath)
		if err == nil {
			for i := 0; i < len(dhcpRows) && i < maxRows; i++ {
				rows = append(rows, dhcpRows[i])
			}
		}
	case "rdp":
		var rdpRows []schema.RDP
		rdpRows, err = parquet.ReadFile[schema.RDP](filePath)
		if err == nil {
			for i := 0; i < len(rdpRows) && i < maxRows; i++ {
				rows = append(rows, rdpRows[i])
			}
		}
	case "smb_files":
		var smbRows []schema.SMB_FILES
		smbRows, err = parquet.ReadFile[schema.SMB_FILES](filePath)
		if err == nil {
			for i := 0; i < len(smbRows) && i < maxRows; i++ {
				rows = append(rows, smbRows[i])
			}
		}
	case "smb_mapping":
		var smbRows []schema.SMB_MAPPING
		smbRows, err = parquet.ReadFile[schema.SMB_MAPPING](filePath)
		if err == nil {
			for i := 0; i < len(smbRows) && i < maxRows; i++ {
				rows = append(rows, smbRows[i])
			}
		}
	case "dce_rpc":
		var dceRows []schema.DCE_RPC
		dceRows, err = parquet.ReadFile[schema.DCE_RPC](filePath)
		if err == nil {
			for i := 0; i < len(dceRows) && i < maxRows; i++ {
				rows = append(rows, dceRows[i])
			}
		}
	case "kerberos":
		var kerbRows []schema.KERBEROS
		kerbRows, err = parquet.ReadFile[schema.KERBEROS](filePath)
		if err == nil {
			for i := 0; i < len(kerbRows) && i < maxRows; i++ {
				rows = append(rows, kerbRows[i])
			}
		}
	case "ntlm":
		var ntlmRows []schema.NTLM
		ntlmRows, err = parquet.ReadFile[schema.NTLM](filePath)
		if err == nil {
			for i := 0; i < len(ntlmRows) && i < maxRows; i++ {
				rows = append(rows, ntlmRows[i])
			}
		}
	case "sip":
		var sipRows []schema.SIP
		sipRows, err = parquet.ReadFile[schema.SIP](filePath)
		if err == nil {
			for i := 0; i < len(sipRows) && i < maxRows; i++ {
				rows = append(rows, sipRows[i])
			}
		}
	case "snmp":
		var snmpRows []schema.SNMP
		snmpRows, err = parquet.ReadFile[schema.SNMP](filePath)
		if err == nil {
			for i := 0; i < len(snmpRows) && i < maxRows; i++ {
				rows = append(rows, snmpRows[i])
			}
		}
	case "radius":
		var radiusRows []schema.RADIUS
		radiusRows, err = parquet.ReadFile[schema.RADIUS](filePath)
		if err == nil {
			for i := 0; i < len(radiusRows) && i < maxRows; i++ {
				rows = append(rows, radiusRows[i])
			}
		}
	case "tunnel":
		var tunnelRows []schema.TUNNEL
		tunnelRows, err = parquet.ReadFile[schema.TUNNEL](filePath)
		if err == nil {
			for i := 0; i < len(tunnelRows) && i < maxRows; i++ {
				rows = append(rows, tunnelRows[i])
			}
		}
	default:
		log.Printf("Warning: Unsupported log type for reading: %s", logType)
		return
	}
	
	if err != nil {
		log.Printf("Error reading Parquet file %s: %v", filePath, err)
		return
	}

	// Process each row
	for _, row := range rows {
		result.TotalRows++
		
		structValue := reflect.ValueOf(row)

		// Extract raw_log field
		rawLogValue := structValue.FieldByName("RawLog")
		if !rawLogValue.IsValid() || rawLogValue.Kind() != reflect.String {
			log.Printf("Warning: RawLog field not found or not a string")
			continue
		}

		rawLogJSON := rawLogValue.String()
		if rawLogJSON == "" {
			continue
		}

		// Parse raw JSON
		var rawLogData map[string]interface{}
		if err := json.Unmarshal([]byte(rawLogJSON), &rawLogData); err != nil {
			log.Printf("Error parsing raw JSON: %v", err)
			continue
		}

		// Extract the actual log data (might be nested like {"dns": {...}})
		var actualLogData map[string]interface{}
		if logData, ok := rawLogData[logType]; ok {
			if logMap, ok := logData.(map[string]interface{}); ok {
				actualLogData = logMap
			}
		} else {
			// Try direct access
			actualLogData = rawLogData
		}

		// Compare fields
		compareFields(row, actualLogData, &result)
	}

	allResults = append(allResults, result)
}

func getSchemaInstanceForLogType(logType string) interface{} {
	// Map log types to schema structs
	switch logType {
	case "dns":
		return &schema.DNS{}
	case "conn":
		return &schema.CONN{}
	case "http":
		return &schema.HTTP{}
	case "ssl":
		return &schema.SSL{}
	case "ssh":
		return &schema.SSH{}
	case "ftp":
		return &schema.FTP{}
	case "smtp":
		return &schema.SMTP{}
	case "dhcp":
		return &schema.DHCP{}
	case "rdp":
		return &schema.RDP{}
	case "smb_files":
		return &schema.SMB_FILES{}
	case "smb_mapping":
		return &schema.SMB_MAPPING{}
	case "dce_rpc":
		return &schema.DCE_RPC{}
	case "kerberos":
		return &schema.KERBEROS{}
	case "ntlm":
		return &schema.NTLM{}
	case "sip":
		return &schema.SIP{}
	case "snmp":
		return &schema.SNMP{}
	case "radius":
		return &schema.RADIUS{}
	case "tunnel":
		return &schema.TUNNEL{}
	default:
		return nil
	}
}

func compareFields(parquetRow interface{}, rawLogData map[string]interface{}, result *FileComparisonResult) {
	parquetValue := reflect.ValueOf(parquetRow)
	parquetType := parquetValue.Type()

	// Track which raw fields we've seen
	rawFieldsSeen := make(map[string]bool)

	// Iterate through Parquet struct fields
	for i := 0; i < parquetType.NumField(); i++ {
		field := parquetType.Field(i)
		fieldValue := parquetValue.Field(i)

		// Skip RawLog field (we already used it)
		if field.Name == "RawLog" {
			continue
		}

		// Get Parquet tag name
		parquetTag := field.Tag.Get("parquet")
		if parquetTag == "" {
			continue
		}

		// Convert Parquet field name to Zeek field name
		zeekFieldName := parquetTagToZeekField(parquetTag)

		// Check if field exists in raw log
		rawValue, existsInRaw := rawLogData[zeekFieldName]
		rawFieldsSeen[zeekFieldName] = existsInRaw

		// Get Parquet value
		parquetValueInterface := getFieldValue(fieldValue)

		// Compare
		comparison := FieldComparisonResult{
			FieldName:    zeekFieldName,
			ParquetValue: parquetValueInterface,
			ParquetType:  field.Type.String(),
		}

		if !existsInRaw {
			// Field in Parquet but not in raw (might be enrichment or promoted)
			// Skip for now, focus on raw field preservation
			continue
		}

		comparison.RawValue = rawValue
		comparison.RawType = fmt.Sprintf("%T", rawValue)

		// Check for multi-value fields (arrays stored as strings)
		if rawArray, ok := rawValue.([]interface{}); ok {
			comparison.RawIsArray = true
			parquetStr, isString := parquetValueInterface.(string)
			comparison.ParquetIsString = isString

			if isString {
				// Convert array to comma-separated string for comparison
				arrayStr := arrayToString(rawArray)
				comparison.Match = (parquetStr == arrayStr)
				if !comparison.Match {
					comparison.Issue = "multi_value_collapsed"
					result.Issues = append(result.Issues, comparison)
					result.FieldsAnalyzed[zeekFieldName]++
				}
			} else {
				comparison.Issue = "type_mismatch_array"
				result.Issues = append(result.Issues, comparison)
				result.FieldsAnalyzed[zeekFieldName]++
			}
			continue
		}

		// Check for type mismatches
		if !valuesMatch(rawValue, parquetValueInterface) {
			comparison.Match = false
			comparison.Issue = "type_mismatch"
			result.Issues = append(result.Issues, comparison)
			result.FieldsAnalyzed[zeekFieldName]++
			continue
		}

		// Values match
		comparison.Match = true
		result.FieldsAnalyzed[zeekFieldName]++
	}

	// Find fields in raw but not in Parquet (missing fields)
	for rawField := range rawLogData {
		if rawField == "ts" || rawField == "uid" || strings.HasPrefix(rawField, "id.") {
			// These are promoted, skip
			continue
		}
		if !rawFieldsSeen[rawField] {
			// Check if it's a promoted field (would have different name)
			promoted := isPromotedField(rawField, result.LogType)
			if !promoted {
				result.MissingFields = append(result.MissingFields, rawField)
			}
		}
	}
}

func parquetTagToZeekField(parquetTag string) string {
	// Convert Parquet field names back to Zeek field names
	zeekFieldMap := map[string]string{
		"id_orig_p":   "id.orig_p",
		"id_orig_h":   "id.orig_h",
		"id_resp_p":   "id.resp_p",
		"id_resp_h":   "id.resp_h",
		"qtype":       "qtype",
		"qtype_name":  "qtype_name",
		"rcode":       "rcode",
		"rcode_name":  "rcode_name",
		"qclass":      "qclass",
		"qclass_name": "qclass_name",
	}

	if zeekName, ok := zeekFieldMap[parquetTag]; ok {
		return zeekName
	}
	return parquetTag
}

func getFieldValue(fieldValue reflect.Value) interface{} {
	if !fieldValue.IsValid() {
		return nil
	}

	switch fieldValue.Kind() {
	case reflect.String:
		return fieldValue.String()
	case reflect.Int, reflect.Int32, reflect.Int64:
		return fieldValue.Int()
	case reflect.Float32, reflect.Float64:
		return fieldValue.Float()
	case reflect.Bool:
		return fieldValue.Bool()
	default:
		return fieldValue.Interface()
	}
}

func valuesMatch(rawValue, parquetValue interface{}) bool {
	// Handle nil
	if rawValue == nil && parquetValue == nil {
		return true
	}
	if rawValue == nil || parquetValue == nil {
		return false
	}

	// Type conversion for numeric types
	switch rv := rawValue.(type) {
	case float64:
		switch pv := parquetValue.(type) {
		case float64:
			return rv == pv
		case int64:
			return rv == float64(pv)
		case int32:
			return rv == float64(pv)
		case int:
			return rv == float64(pv)
		}
	case int:
		switch pv := parquetValue.(type) {
		case int64:
			return int64(rv) == pv
		case int32:
			return int32(rv) == pv
		case float64:
			return float64(rv) == pv
		}
	case string:
		if pv, ok := parquetValue.(string); ok {
			return rv == pv
		}
	case bool:
		if pv, ok := parquetValue.(bool); ok {
			return rv == pv
		}
	}

	return false
}

func arrayToString(arr []interface{}) string {
	parts := make([]string, len(arr))
	for i, v := range arr {
		parts[i] = fmt.Sprintf("%v", v)
	}
	return strings.Join(parts, ",")
}

func isPromotedField(fieldName, logType string) bool {
	// Simple check - if field is in common promotion list, it's promoted
	promotedFields := map[string]bool{
		"ts":         true,
		"uid":        true,
		"id.orig_h":  true,
		"id.orig_p":  true,
		"id.resp_h":  true,
		"id.resp_p":  true,
		"proto":      true,
		"service":    true,
	}
	return promotedFields[fieldName]
}

func generateReport() {
	fmt.Println()
	fmt.Println("=" + strings.Repeat("=", 80))
	fmt.Println("VALIDATION REPORT")
	fmt.Println("=" + strings.Repeat("=", 80))
	fmt.Println()

	// Summary by log type
	logTypeSummary := make(map[string]map[string]int)
	for _, result := range allResults {
		if logTypeSummary[result.LogType] == nil {
			logTypeSummary[result.LogType] = make(map[string]int)
		}
		logTypeSummary[result.LogType]["files"]++
		logTypeSummary[result.LogType]["rows"] += result.TotalRows
		logTypeSummary[result.LogType]["issues"] += len(result.Issues)
		logTypeSummary[result.LogType]["missing"] += len(result.MissingFields)
	}

	// Print summary
	fmt.Println("SUMMARY BY LOG TYPE")
	fmt.Println(strings.Repeat("-", 80))
	fmt.Printf("%-15s %8s %10s %10s %10s\n", "LOG_TYPE", "FILES", "ROWS", "ISSUES", "MISSING")
	fmt.Println(strings.Repeat("-", 80))
	for logType := range logTypeSummary {
		stats := logTypeSummary[logType]
		fmt.Printf("%-15s %8d %10d %10d %10d\n",
			logType, stats["files"], stats["rows"], stats["issues"], stats["missing"])
	}
	fmt.Println()

	// Detailed issues by type
	issueTypes := make(map[string][]FieldComparisonResult)
	for _, result := range allResults {
		for _, issue := range result.Issues {
			issueTypes[issue.Issue] = append(issueTypes[issue.Issue], issue)
		}
	}

	fmt.Println("ISSUE BREAKDOWN")
	fmt.Println(strings.Repeat("-", 80))
	for issueType, issues := range issueTypes {
		fmt.Printf("\n%s: %d occurrences\n", strings.ToUpper(issueType), len(issues))
		
		// Group by field name
		fieldCounts := make(map[string]int)
		for _, issue := range issues {
			fieldCounts[issue.FieldName]++
		}

		// Sort by count
		type fieldCount struct {
			name  string
			count int
		}
		sortedFields := make([]fieldCount, 0, len(fieldCounts))
		for name, count := range fieldCounts {
			sortedFields = append(sortedFields, fieldCount{name, count})
		}
		sort.Slice(sortedFields, func(i, j int) bool {
			return sortedFields[i].count > sortedFields[j].count
		})

		// Show top 10 fields
		fmt.Printf("  Top affected fields:\n")
		for i, fc := range sortedFields {
			if i >= 10 {
				break
			}
			fmt.Printf("    - %s: %d occurrences\n", fc.name, fc.count)
		}

		// Show sample issue
		if len(issues) > 0 {
			sample := issues[0]
			fmt.Printf("  Sample:\n")
			fmt.Printf("    Field: %s\n", sample.FieldName)
			fmt.Printf("    Raw Type: %s, Parquet Type: %s\n", sample.RawType, sample.ParquetType)
			if sample.RawIsArray {
				fmt.Printf("    Raw is array, Parquet is string: %v\n", sample.ParquetIsString)
			}
			fmt.Printf("    Raw Value: %v\n", truncateString(fmt.Sprintf("%v", sample.RawValue), 100))
			fmt.Printf("    Parquet Value: %v\n", truncateString(fmt.Sprintf("%v", sample.ParquetValue), 100))
		}
	}
	fmt.Println()

	// Missing fields
	fmt.Println("MISSING FIELDS (in raw JSON but not in Parquet schema)")
	fmt.Println(strings.Repeat("-", 80))
	missingByLogType := make(map[string]map[string]int)
	for _, result := range allResults {
		if missingByLogType[result.LogType] == nil {
			missingByLogType[result.LogType] = make(map[string]int)
		}
		for _, field := range result.MissingFields {
			missingByLogType[result.LogType][field]++
		}
	}

	for logType, fields := range missingByLogType {
		if len(fields) > 0 {
			fmt.Printf("\n%s:\n", logType)
			for field, count := range fields {
				fmt.Printf("  - %s (%d occurrences)\n", field, count)
			}
		}
	}
	fmt.Println()

	// Recommendations
	fmt.Println("RECOMMENDATIONS")
	fmt.Println(strings.Repeat("-", 80))
	if len(issueTypes["multi_value_collapsed"]) > 0 {
		fmt.Println("1. MULTI-VALUE FIELDS (Arrays stored as strings):")
		fmt.Println("   - Fields like 'answers' and 'TTLs' in DNS logs contain arrays")
		fmt.Println("   - Currently stored as comma-separated strings")
		fmt.Println("   - Consider: Using Parquet LIST/ARRAY types, or JSON strings")
		fmt.Println()
	}
	if len(issueTypes["type_mismatch"]) > 0 {
		fmt.Println("2. TYPE MISMATCHES:")
		fmt.Println("   - Some fields have type mismatches between raw and Parquet")
		fmt.Println("   - Review type conversions in parquet.go setFieldValue()")
		fmt.Println()
	}
	if len(missingByLogType) > 0 {
		fmt.Println("3. MISSING FIELDS:")
		fmt.Println("   - Some raw fields are not present in Parquet schema")
		fmt.Println("   - These fields are lost during conversion")
		fmt.Println("   - Add missing fields to config/schema.json and regenerate events.go")
		fmt.Println()
	}

	fmt.Println("=" + strings.Repeat("=", 80))
	fmt.Println("Validation complete!")
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

