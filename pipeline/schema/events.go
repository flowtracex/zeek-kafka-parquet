package schema

import (
	"time"
)

// DNS represents a normalized DNS event with all Zeek fields preserved
type DNS struct {
	// Common normalized fields
	EventTime        int64  `parquet:"event_time"`
	IngestTime       int64  `parquet:"ingest_time"`
	EventYear        int32  `parquet:"event_year"`
	EventMonth       int32  `parquet:"event_month"`
	EventDay         int32  `parquet:"event_day"`
	EventHour        int32  `parquet:"event_hour"`
	EventWeekday     int32  `parquet:"event_weekday"`
	EventType        string `parquet:"event_type"`
	EventClass       string `parquet:"event_class"`
	SrcIP            string `parquet:"src_ip"`
	DstIP            string `parquet:"dst_ip"`
	SrcIPIsPrivate   bool   `parquet:"src_ip_is_private"`
	DstIPIsPrivate   bool   `parquet:"dst_ip_is_private"`
	Direction        string `parquet:"direction"`
	Service          string `parquet:"service"`
	FlowID           string `parquet:"flow_id"`

	// DNS-specific fields
	DNSQuery string `parquet:"dns_query"`
	QType    string `parquet:"qtype"`
	RCode    int32  `parquet:"rcode"`

	// Raw Zeek log (preserved as JSON string)
	RawLog string `parquet:"raw_log"`
}

// CONN represents a normalized CONN event with all Zeek fields preserved
type CONN struct {
	// Common normalized fields
	EventTime        int64  `parquet:"event_time"`
	IngestTime       int64  `parquet:"ingest_time"`
	EventYear        int32  `parquet:"event_year"`
	EventMonth       int32  `parquet:"event_month"`
	EventDay         int32  `parquet:"event_day"`
	EventHour        int32  `parquet:"event_hour"`
	EventWeekday     int32  `parquet:"event_weekday"`
	EventType        string `parquet:"event_type"`
	EventClass       string `parquet:"event_class"`
	SrcIP            string `parquet:"src_ip"`
	DstIP            string `parquet:"dst_ip"`
	SrcPort          int32  `parquet:"src_port"`
	DstPort          int32  `parquet:"dst_port"`
	SrcIPIsPrivate   bool   `parquet:"src_ip_is_private"`
	DstIPIsPrivate   bool   `parquet:"dst_ip_is_private"`
	Direction        string `parquet:"direction"`
	Service          string `parquet:"service"`
	FlowID           string `parquet:"flow_id"`

	// CONN-specific fields
	Protocol  string `parquet:"protocol"`
	ConnState string `parquet:"conn_state"`

	// Raw Zeek log (preserved as JSON string)
	RawLog string `parquet:"raw_log"`
}

// ToParquetTime converts Go time to milliseconds since epoch
func ToParquetTime(t time.Time) int64 {
	return t.UnixNano() / int64(time.Millisecond)
}
