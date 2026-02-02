package schema

import (
	"time"
)

// CONN represents a normalized conn event following the three-layer model:
// 1) Conditional enrichment (controlled per log type via normalization.json enrich flags)
// 2) Normalized/promoted fields (from normalization.json mapping)
// 3) Unmapped raw fields (from schema.json, NOT in normalization.json)
// 4) raw_log (complete original JSON)
type CONN struct {
	// =========================
	// CONDITIONAL ENRICHMENT (Layer 3)
	// Applied based on normalization.json enrich flags (time, network)
	// Fields may be zero/empty if enrichment is disabled for this log type
	// =========================
	IngestTime       int64  `parquet:"ingest_time"`
	EventYear        int32  `parquet:"event_year"`
	EventMonth       int32  `parquet:"event_month"`
	EventDay         int32  `parquet:"event_day"`
	EventHour        int32  `parquet:"event_hour"`
	EventWeekday     int32  `parquet:"event_weekday"`
	EventType        string `parquet:"event_type"`
	EventClass       string `parquet:"event_class"`
	SrcIPIsPrivate   bool   `parquet:"src_ip_is_private"`
	DstIPIsPrivate   bool   `parquet:"dst_ip_is_private"`
	Direction        string `parquet:"direction"`

	// =========================
	// NORMALIZED (PROMOTED) FIELDS (Layer 2)
	// From normalization.json mapping - these replace raw fields
	// DO NOT include raw versions
	// =========================
	DstIP string `parquet:"dst_ip"` // Promoted from: id.resp_h
	DstPort int32 `parquet:"dst_port"` // Promoted from: id.resp_p
	EventTime int64 `parquet:"event_time"` // Promoted from: ts
	FlowID string `parquet:"flow_id"` // Promoted from: uid
	Protocol string `parquet:"protocol"` // Promoted from: proto
	Service string `parquet:"service"` // Promoted from: service
	SrcIP string `parquet:"src_ip"` // Promoted from: id.orig_h
	SrcPort int32 `parquet:"src_port"` // Promoted from: id.orig_p

	// =========================
	// RAW FIELDS (UNMAPPED) (Layer 1)
	// From schema.json, but NOT in normalization.json mapping
	// These are raw fields that were NOT promoted
	// =========================
	ConnState string `parquet:"conn_state"` // Not mapped
	Duration float64 `parquet:"duration"` // Not mapped
	History string `parquet:"history"` // Not mapped
	IpProto int32 `parquet:"ip_proto"` // Not mapped
	LocalOrig bool `parquet:"local_orig"` // Not mapped
	LocalResp bool `parquet:"local_resp"` // Not mapped
	MissedBytes int64 `parquet:"missed_bytes"` // Not mapped
	OrigBytes int64 `parquet:"orig_bytes"` // Not mapped
	OrigIpBytes int64 `parquet:"orig_ip_bytes"` // Not mapped
	OrigPkts int64 `parquet:"orig_pkts"` // Not mapped
	RespBytes int64 `parquet:"resp_bytes"` // Not mapped
	RespIpBytes int64 `parquet:"resp_ip_bytes"` // Not mapped
	RespPkts int64 `parquet:"resp_pkts"` // Not mapped
	TunnelParents string `parquet:"tunnel_parents"` // Not mapped

	// =========================
	// FULL RAW LOG (complete original JSON)
	// =========================
	RawLog string `parquet:"raw_log"`
}

// DCE_RPC represents a normalized dce_rpc event following the three-layer model:
// 1) Conditional enrichment (controlled per log type via normalization.json enrich flags)
// 2) Normalized/promoted fields (from normalization.json mapping)
// 3) Unmapped raw fields (from schema.json, NOT in normalization.json)
// 4) raw_log (complete original JSON)
type DCE_RPC struct {
	// =========================
	// CONDITIONAL ENRICHMENT (Layer 3)
	// Applied based on normalization.json enrich flags (time, network)
	// Fields may be zero/empty if enrichment is disabled for this log type
	// =========================
	IngestTime       int64  `parquet:"ingest_time"`
	EventYear        int32  `parquet:"event_year"`
	EventMonth       int32  `parquet:"event_month"`
	EventDay         int32  `parquet:"event_day"`
	EventHour        int32  `parquet:"event_hour"`
	EventWeekday     int32  `parquet:"event_weekday"`
	EventType        string `parquet:"event_type"`
	EventClass       string `parquet:"event_class"`
	SrcIPIsPrivate   bool   `parquet:"src_ip_is_private"`
	DstIPIsPrivate   bool   `parquet:"dst_ip_is_private"`
	Direction        string `parquet:"direction"`
	Service          string `parquet:"service"`

	// =========================
	// NORMALIZED (PROMOTED) FIELDS (Layer 2)
	// From normalization.json mapping - these replace raw fields
	// DO NOT include raw versions
	// =========================
	DstIP string `parquet:"dst_ip"` // Promoted from: id.resp_h
	DstPort int32 `parquet:"dst_port"` // Promoted from: id.resp_p
	EventTime int64 `parquet:"event_time"` // Promoted from: ts
	FlowID string `parquet:"flow_id"` // Promoted from: uid
	SrcIP string `parquet:"src_ip"` // Promoted from: id.orig_h
	SrcPort int32 `parquet:"src_port"` // Promoted from: id.orig_p

	// =========================
	// RAW FIELDS (UNMAPPED) (Layer 1)
	// From schema.json, but NOT in normalization.json mapping
	// These are raw fields that were NOT promoted
	// =========================
	Endpoint string `parquet:"endpoint"` // Not mapped
	NamedPipe string `parquet:"named_pipe"` // Not mapped
	Operation string `parquet:"operation"` // Not mapped
	Rtt float64 `parquet:"rtt"` // Not mapped

	// =========================
	// FULL RAW LOG (complete original JSON)
	// =========================
	RawLog string `parquet:"raw_log"`
}

// DHCP represents a normalized dhcp event following the three-layer model:
// 1) Conditional enrichment (controlled per log type via normalization.json enrich flags)
// 2) Normalized/promoted fields (from normalization.json mapping)
// 3) Unmapped raw fields (from schema.json, NOT in normalization.json)
// 4) raw_log (complete original JSON)
type DHCP struct {
	// =========================
	// CONDITIONAL ENRICHMENT (Layer 3)
	// Applied based on normalization.json enrich flags (time, network)
	// Fields may be zero/empty if enrichment is disabled for this log type
	// =========================
	IngestTime       int64  `parquet:"ingest_time"`
	EventYear        int32  `parquet:"event_year"`
	EventMonth       int32  `parquet:"event_month"`
	EventDay         int32  `parquet:"event_day"`
	EventHour        int32  `parquet:"event_hour"`
	EventWeekday     int32  `parquet:"event_weekday"`
	EventType        string `parquet:"event_type"`
	EventClass       string `parquet:"event_class"`

	// =========================
	// NORMALIZED (PROMOTED) FIELDS (Layer 2)
	// From normalization.json mapping - these replace raw fields
	// DO NOT include raw versions
	// =========================
	EventTime int64 `parquet:"event_time"` // Promoted from: ts
	FlowID string `parquet:"flow_id"` // Promoted from: uid

	// =========================
	// RAW FIELDS (UNMAPPED) (Layer 1)
	// From schema.json, but NOT in normalization.json mapping
	// These are raw fields that were NOT promoted
	// =========================
	AssignedAddr string `parquet:"assigned_addr"` // Not mapped
	ClientAddr string `parquet:"client_addr"` // Not mapped
	ClientFqdn string `parquet:"client_fqdn"` // Not mapped
	ClientMessage string `parquet:"client_message"` // Not mapped
	Domain string `parquet:"domain"` // Not mapped
	Duration float64 `parquet:"duration"` // Not mapped
	HostName string `parquet:"host_name"` // Not mapped
	IdOrigH string `parquet:"id_orig_h"` // Not mapped
	IdOrigP int32 `parquet:"id_orig_p"` // Not mapped
	IdRespH string `parquet:"id_resp_h"` // Not mapped
	IdRespP int32 `parquet:"id_resp_p"` // Not mapped
	LeaseTime int32 `parquet:"lease_time"` // Not mapped
	Mac string `parquet:"mac"` // Not mapped
	MsgTypes string `parquet:"msg_types"` // Not mapped
	RequestedAddr string `parquet:"requested_addr"` // Not mapped
	ServerAddr string `parquet:"server_addr"` // Not mapped
	ServerMessage string `parquet:"server_message"` // Not mapped
	Uids string `parquet:"uids"` // Not mapped

	// =========================
	// FULL RAW LOG (complete original JSON)
	// =========================
	RawLog string `parquet:"raw_log"`
}

// DNS represents a normalized dns event following the three-layer model:
// 1) Conditional enrichment (controlled per log type via normalization.json enrich flags)
// 2) Normalized/promoted fields (from normalization.json mapping)
// 3) Unmapped raw fields (from schema.json, NOT in normalization.json)
// 4) raw_log (complete original JSON)
type DNS struct {
	// =========================
	// CONDITIONAL ENRICHMENT (Layer 3)
	// Applied based on normalization.json enrich flags (time, network)
	// Fields may be zero/empty if enrichment is disabled for this log type
	// =========================
	IngestTime       int64  `parquet:"ingest_time"`
	EventYear        int32  `parquet:"event_year"`
	EventMonth       int32  `parquet:"event_month"`
	EventDay         int32  `parquet:"event_day"`
	EventHour        int32  `parquet:"event_hour"`
	EventWeekday     int32  `parquet:"event_weekday"`
	EventType        string `parquet:"event_type"`
	EventClass       string `parquet:"event_class"`
	SrcIPIsPrivate   bool   `parquet:"src_ip_is_private"`
	DstIPIsPrivate   bool   `parquet:"dst_ip_is_private"`
	Direction        string `parquet:"direction"`
	Service          string `parquet:"service"`

	// =========================
	// NORMALIZED (PROMOTED) FIELDS (Layer 2)
	// From normalization.json mapping - these replace raw fields
	// DO NOT include raw versions
	// =========================
	DstIP string `parquet:"dst_ip"` // Promoted from: id.resp_h
	EventTime int64 `parquet:"event_time"` // Promoted from: ts
	FlowID string `parquet:"flow_id"` // Promoted from: uid
	Protocol string `parquet:"protocol"` // Promoted from: proto
	SrcIP string `parquet:"src_ip"` // Promoted from: id.orig_h

	// =========================
	// RAW FIELDS (UNMAPPED) (Layer 1)
	// From schema.json, but NOT in normalization.json mapping
	// These are raw fields that were NOT promoted
	// =========================
	AA bool `parquet:"AA"` // Not mapped
	RA bool `parquet:"RA"` // Not mapped
	RD bool `parquet:"RD"` // Not mapped
	TC bool `parquet:"TC"` // Not mapped
	TTLs string `parquet:"TTLs"` // Not mapped
	Z int32 `parquet:"Z"` // Not mapped
	Answers string `parquet:"answers"` // Not mapped
	IdOrigP int32 `parquet:"id_orig_p"` // Not mapped
	IdRespP int32 `parquet:"id_resp_p"` // Not mapped
	Qclass int32 `parquet:"qclass"` // Not mapped
	QclassName string `parquet:"qclass_name"` // Not mapped
	Qtype int32 `parquet:"qtype"` // Not mapped
	QtypeName string `parquet:"qtype_name"` // Not mapped
	Query string `parquet:"query"` // Not mapped
	Rcode int32 `parquet:"rcode"` // Not mapped
	RcodeName string `parquet:"rcode_name"` // Not mapped
	Rejected bool `parquet:"rejected"` // Not mapped
	TransID int32 `parquet:"trans_id"` // Not mapped

	// =========================
	// FULL RAW LOG (complete original JSON)
	// =========================
	RawLog string `parquet:"raw_log"`
}

// FTP represents a normalized ftp event following the three-layer model:
// 1) Conditional enrichment (controlled per log type via normalization.json enrich flags)
// 2) Normalized/promoted fields (from normalization.json mapping)
// 3) Unmapped raw fields (from schema.json, NOT in normalization.json)
// 4) raw_log (complete original JSON)
type FTP struct {
	// =========================
	// CONDITIONAL ENRICHMENT (Layer 3)
	// Applied based on normalization.json enrich flags (time, network)
	// Fields may be zero/empty if enrichment is disabled for this log type
	// =========================
	IngestTime       int64  `parquet:"ingest_time"`
	EventYear        int32  `parquet:"event_year"`
	EventMonth       int32  `parquet:"event_month"`
	EventDay         int32  `parquet:"event_day"`
	EventHour        int32  `parquet:"event_hour"`
	EventWeekday     int32  `parquet:"event_weekday"`
	EventType        string `parquet:"event_type"`
	EventClass       string `parquet:"event_class"`
	SrcIPIsPrivate   bool   `parquet:"src_ip_is_private"`
	DstIPIsPrivate   bool   `parquet:"dst_ip_is_private"`
	Direction        string `parquet:"direction"`
	Service          string `parquet:"service"`

	// =========================
	// NORMALIZED (PROMOTED) FIELDS (Layer 2)
	// From normalization.json mapping - these replace raw fields
	// DO NOT include raw versions
	// =========================
	DstIP string `parquet:"dst_ip"` // Promoted from: id.resp_h
	DstPort int32 `parquet:"dst_port"` // Promoted from: id.resp_p
	EventTime int64 `parquet:"event_time"` // Promoted from: ts
	FlowID string `parquet:"flow_id"` // Promoted from: uid
	SrcIP string `parquet:"src_ip"` // Promoted from: id.orig_h
	SrcPort int32 `parquet:"src_port"` // Promoted from: id.orig_p

	// =========================
	// RAW FIELDS (UNMAPPED) (Layer 1)
	// From schema.json, but NOT in normalization.json mapping
	// These are raw fields that were NOT promoted
	// =========================
	Arg string `parquet:"arg"` // Not mapped
	Command string `parquet:"command"` // Not mapped
	DataChannelOrigH string `parquet:"data_channel_orig_h"` // Not mapped
	DataChannelPassive bool `parquet:"data_channel_passive"` // Not mapped
	DataChannelRespH string `parquet:"data_channel_resp_h"` // Not mapped
	DataChannelRespP int32 `parquet:"data_channel_resp_p"` // Not mapped
	FileSize int32 `parquet:"file_size"` // Not mapped
	Fuid string `parquet:"fuid"` // Not mapped
	MimeType string `parquet:"mime_type"` // Not mapped
	Password string `parquet:"password"` // Not mapped
	ReplyCode int32 `parquet:"reply_code"` // Not mapped
	ReplyMsg string `parquet:"reply_msg"` // Not mapped
	User string `parquet:"user"` // Not mapped

	// =========================
	// FULL RAW LOG (complete original JSON)
	// =========================
	RawLog string `parquet:"raw_log"`
}

// HTTP represents a normalized http event following the three-layer model:
// 1) Conditional enrichment (controlled per log type via normalization.json enrich flags)
// 2) Normalized/promoted fields (from normalization.json mapping)
// 3) Unmapped raw fields (from schema.json, NOT in normalization.json)
// 4) raw_log (complete original JSON)
type HTTP struct {
	// =========================
	// CONDITIONAL ENRICHMENT (Layer 3)
	// Applied based on normalization.json enrich flags (time, network)
	// Fields may be zero/empty if enrichment is disabled for this log type
	// =========================
	IngestTime       int64  `parquet:"ingest_time"`
	EventYear        int32  `parquet:"event_year"`
	EventMonth       int32  `parquet:"event_month"`
	EventDay         int32  `parquet:"event_day"`
	EventHour        int32  `parquet:"event_hour"`
	EventWeekday     int32  `parquet:"event_weekday"`
	EventType        string `parquet:"event_type"`
	EventClass       string `parquet:"event_class"`
	SrcIPIsPrivate   bool   `parquet:"src_ip_is_private"`
	DstIPIsPrivate   bool   `parquet:"dst_ip_is_private"`
	Direction        string `parquet:"direction"`
	Service          string `parquet:"service"`

	// =========================
	// NORMALIZED (PROMOTED) FIELDS (Layer 2)
	// From normalization.json mapping - these replace raw fields
	// DO NOT include raw versions
	// =========================
	DstIP string `parquet:"dst_ip"` // Promoted from: id.resp_h
	DstPort int32 `parquet:"dst_port"` // Promoted from: id.resp_p
	EventTime int64 `parquet:"event_time"` // Promoted from: ts
	FlowID string `parquet:"flow_id"` // Promoted from: uid
	SrcIP string `parquet:"src_ip"` // Promoted from: id.orig_h
	SrcPort int32 `parquet:"src_port"` // Promoted from: id.orig_p

	// =========================
	// RAW FIELDS (UNMAPPED) (Layer 1)
	// From schema.json, but NOT in normalization.json mapping
	// These are raw fields that were NOT promoted
	// =========================
	Host string `parquet:"host"` // Not mapped
	InfoCode int32 `parquet:"info_code"` // Not mapped
	InfoMsg string `parquet:"info_msg"` // Not mapped
	Method string `parquet:"method"` // Not mapped
	OrigFilenames string `parquet:"orig_filenames"` // Not mapped
	OrigFuids string `parquet:"orig_fuids"` // Not mapped
	OrigMimeTypes string `parquet:"orig_mime_types"` // Not mapped
	Origin string `parquet:"origin"` // Not mapped
	Password string `parquet:"password"` // Not mapped
	Proxied string `parquet:"proxied"` // Not mapped
	Referrer string `parquet:"referrer"` // Not mapped
	RequestBodyLen int32 `parquet:"request_body_len"` // Not mapped
	RespFilenames string `parquet:"resp_filenames"` // Not mapped
	RespFuids string `parquet:"resp_fuids"` // Not mapped
	RespMimeTypes string `parquet:"resp_mime_types"` // Not mapped
	ResponseBodyLen int32 `parquet:"response_body_len"` // Not mapped
	StatusCode int32 `parquet:"status_code"` // Not mapped
	StatusMsg string `parquet:"status_msg"` // Not mapped
	Tags string `parquet:"tags"` // Not mapped
	TransDepth int32 `parquet:"trans_depth"` // Not mapped
	Uri string `parquet:"uri"` // Not mapped
	UserAgent string `parquet:"user_agent"` // Not mapped
	Username string `parquet:"username"` // Not mapped
	Version string `parquet:"version"` // Not mapped

	// =========================
	// FULL RAW LOG (complete original JSON)
	// =========================
	RawLog string `parquet:"raw_log"`
}

// KERBEROS represents a normalized kerberos event following the three-layer model:
// 1) Conditional enrichment (controlled per log type via normalization.json enrich flags)
// 2) Normalized/promoted fields (from normalization.json mapping)
// 3) Unmapped raw fields (from schema.json, NOT in normalization.json)
// 4) raw_log (complete original JSON)
type KERBEROS struct {
	// =========================
	// CONDITIONAL ENRICHMENT (Layer 3)
	// Applied based on normalization.json enrich flags (time, network)
	// Fields may be zero/empty if enrichment is disabled for this log type
	// =========================
	IngestTime       int64  `parquet:"ingest_time"`
	EventYear        int32  `parquet:"event_year"`
	EventMonth       int32  `parquet:"event_month"`
	EventDay         int32  `parquet:"event_day"`
	EventHour        int32  `parquet:"event_hour"`
	EventWeekday     int32  `parquet:"event_weekday"`
	EventType        string `parquet:"event_type"`
	EventClass       string `parquet:"event_class"`
	SrcIPIsPrivate   bool   `parquet:"src_ip_is_private"`
	DstIPIsPrivate   bool   `parquet:"dst_ip_is_private"`
	Direction        string `parquet:"direction"`
	Service          string `parquet:"service"`

	// =========================
	// NORMALIZED (PROMOTED) FIELDS (Layer 2)
	// From normalization.json mapping - these replace raw fields
	// DO NOT include raw versions
	// =========================
	DstIP string `parquet:"dst_ip"` // Promoted from: id.resp_h
	DstPort int32 `parquet:"dst_port"` // Promoted from: id.resp_p
	EventTime int64 `parquet:"event_time"` // Promoted from: ts
	FlowID string `parquet:"flow_id"` // Promoted from: uid
	SrcIP string `parquet:"src_ip"` // Promoted from: id.orig_h
	SrcPort int32 `parquet:"src_port"` // Promoted from: id.orig_p

	// =========================
	// RAW FIELDS (UNMAPPED) (Layer 1)
	// From schema.json, but NOT in normalization.json mapping
	// These are raw fields that were NOT promoted
	// =========================
	Cipher string `parquet:"cipher"` // Not mapped
	Client string `parquet:"client"` // Not mapped
	ClientCertFuid string `parquet:"client_cert_fuid"` // Not mapped
	ClientCertSubject string `parquet:"client_cert_subject"` // Not mapped
	ErrorMsg string `parquet:"error_msg"` // Not mapped
	Forwardable bool `parquet:"forwardable"` // Not mapped
	From float64 `parquet:"from"` // Not mapped
	Renewable bool `parquet:"renewable"` // Not mapped
	RequestType string `parquet:"request_type"` // Not mapped
	ServerCertFuid string `parquet:"server_cert_fuid"` // Not mapped
	ServerCertSubject string `parquet:"server_cert_subject"` // Not mapped
	ZeekService string `parquet:"service"` // Not mapped
	Success bool `parquet:"success"` // Not mapped
	Till float64 `parquet:"till"` // Not mapped

	// =========================
	// FULL RAW LOG (complete original JSON)
	// =========================
	RawLog string `parquet:"raw_log"`
}

// NTLM represents a normalized ntlm event following the three-layer model:
// 1) Conditional enrichment (controlled per log type via normalization.json enrich flags)
// 2) Normalized/promoted fields (from normalization.json mapping)
// 3) Unmapped raw fields (from schema.json, NOT in normalization.json)
// 4) raw_log (complete original JSON)
type NTLM struct {
	// =========================
	// CONDITIONAL ENRICHMENT (Layer 3)
	// Applied based on normalization.json enrich flags (time, network)
	// Fields may be zero/empty if enrichment is disabled for this log type
	// =========================
	IngestTime       int64  `parquet:"ingest_time"`
	EventYear        int32  `parquet:"event_year"`
	EventMonth       int32  `parquet:"event_month"`
	EventDay         int32  `parquet:"event_day"`
	EventHour        int32  `parquet:"event_hour"`
	EventWeekday     int32  `parquet:"event_weekday"`
	EventType        string `parquet:"event_type"`
	EventClass       string `parquet:"event_class"`
	SrcIPIsPrivate   bool   `parquet:"src_ip_is_private"`
	DstIPIsPrivate   bool   `parquet:"dst_ip_is_private"`
	Direction        string `parquet:"direction"`
	Service          string `parquet:"service"`

	// =========================
	// NORMALIZED (PROMOTED) FIELDS (Layer 2)
	// From normalization.json mapping - these replace raw fields
	// DO NOT include raw versions
	// =========================
	DstIP string `parquet:"dst_ip"` // Promoted from: id.resp_h
	EventTime int64 `parquet:"event_time"` // Promoted from: ts
	FlowID string `parquet:"flow_id"` // Promoted from: uid
	SrcIP string `parquet:"src_ip"` // Promoted from: id.orig_h

	// =========================
	// RAW FIELDS (UNMAPPED) (Layer 1)
	// From schema.json, but NOT in normalization.json mapping
	// These are raw fields that were NOT promoted
	// =========================
	Domainname string `parquet:"domainname"` // Not mapped
	Hostname string `parquet:"hostname"` // Not mapped
	IdOrigP int32 `parquet:"id_orig_p"` // Not mapped
	IdRespP int32 `parquet:"id_resp_p"` // Not mapped
	ServerDnsComputerName string `parquet:"server_dns_computer_name"` // Not mapped
	ServerNbComputerName string `parquet:"server_nb_computer_name"` // Not mapped
	ServerTreeName string `parquet:"server_tree_name"` // Not mapped
	Success bool `parquet:"success"` // Not mapped
	Username string `parquet:"username"` // Not mapped

	// =========================
	// FULL RAW LOG (complete original JSON)
	// =========================
	RawLog string `parquet:"raw_log"`
}

// RADIUS represents a normalized radius event following the three-layer model:
// 1) Conditional enrichment (controlled per log type via normalization.json enrich flags)
// 2) Normalized/promoted fields (from normalization.json mapping)
// 3) Unmapped raw fields (from schema.json, NOT in normalization.json)
// 4) raw_log (complete original JSON)
type RADIUS struct {
	// =========================
	// CONDITIONAL ENRICHMENT (Layer 3)
	// Applied based on normalization.json enrich flags (time, network)
	// Fields may be zero/empty if enrichment is disabled for this log type
	// =========================
	IngestTime       int64  `parquet:"ingest_time"`
	EventYear        int32  `parquet:"event_year"`
	EventMonth       int32  `parquet:"event_month"`
	EventDay         int32  `parquet:"event_day"`
	EventHour        int32  `parquet:"event_hour"`
	EventWeekday     int32  `parquet:"event_weekday"`
	EventType        string `parquet:"event_type"`
	EventClass       string `parquet:"event_class"`
	SrcIPIsPrivate   bool   `parquet:"src_ip_is_private"`
	DstIPIsPrivate   bool   `parquet:"dst_ip_is_private"`
	Direction        string `parquet:"direction"`
	Service          string `parquet:"service"`

	// =========================
	// NORMALIZED (PROMOTED) FIELDS (Layer 2)
	// From normalization.json mapping - these replace raw fields
	// DO NOT include raw versions
	// =========================
	DstIP string `parquet:"dst_ip"` // Promoted from: id.resp_h
	DstPort int32 `parquet:"dst_port"` // Promoted from: id.resp_p
	EventTime int64 `parquet:"event_time"` // Promoted from: ts
	FlowID string `parquet:"flow_id"` // Promoted from: uid
	SrcIP string `parquet:"src_ip"` // Promoted from: id.orig_h
	SrcPort int32 `parquet:"src_port"` // Promoted from: id.orig_p

	// =========================
	// RAW FIELDS (UNMAPPED) (Layer 1)
	// From schema.json, but NOT in normalization.json mapping
	// These are raw fields that were NOT promoted
	// =========================
	ConnectInfo string `parquet:"connect_info"` // Not mapped
	FramedAddr string `parquet:"framed_addr"` // Not mapped
	Mac string `parquet:"mac"` // Not mapped
	ReplyMsg string `parquet:"reply_msg"` // Not mapped
	Result string `parquet:"result"` // Not mapped
	Ttl int32 `parquet:"ttl"` // Not mapped
	TunnelClient string `parquet:"tunnel_client"` // Not mapped
	Username string `parquet:"username"` // Not mapped

	// =========================
	// FULL RAW LOG (complete original JSON)
	// =========================
	RawLog string `parquet:"raw_log"`
}

// RDP represents a normalized rdp event following the three-layer model:
// 1) Conditional enrichment (controlled per log type via normalization.json enrich flags)
// 2) Normalized/promoted fields (from normalization.json mapping)
// 3) Unmapped raw fields (from schema.json, NOT in normalization.json)
// 4) raw_log (complete original JSON)
type RDP struct {
	// =========================
	// CONDITIONAL ENRICHMENT (Layer 3)
	// Applied based on normalization.json enrich flags (time, network)
	// Fields may be zero/empty if enrichment is disabled for this log type
	// =========================
	IngestTime       int64  `parquet:"ingest_time"`
	EventYear        int32  `parquet:"event_year"`
	EventMonth       int32  `parquet:"event_month"`
	EventDay         int32  `parquet:"event_day"`
	EventHour        int32  `parquet:"event_hour"`
	EventWeekday     int32  `parquet:"event_weekday"`
	EventType        string `parquet:"event_type"`
	EventClass       string `parquet:"event_class"`
	SrcIPIsPrivate   bool   `parquet:"src_ip_is_private"`
	DstIPIsPrivate   bool   `parquet:"dst_ip_is_private"`
	Direction        string `parquet:"direction"`
	Service          string `parquet:"service"`

	// =========================
	// NORMALIZED (PROMOTED) FIELDS (Layer 2)
	// From normalization.json mapping - these replace raw fields
	// DO NOT include raw versions
	// =========================
	DstIP string `parquet:"dst_ip"` // Promoted from: id.resp_h
	DstPort int32 `parquet:"dst_port"` // Promoted from: id.resp_p
	EventTime int64 `parquet:"event_time"` // Promoted from: ts
	FlowID string `parquet:"flow_id"` // Promoted from: uid
	SrcIP string `parquet:"src_ip"` // Promoted from: id.orig_h
	SrcPort int32 `parquet:"src_port"` // Promoted from: id.orig_p

	// =========================
	// RAW FIELDS (UNMAPPED) (Layer 1)
	// From schema.json, but NOT in normalization.json mapping
	// These are raw fields that were NOT promoted
	// =========================
	CertCount int32 `parquet:"cert_count"` // Not mapped
	CertPermanent bool `parquet:"cert_permanent"` // Not mapped
	CertType string `parquet:"cert_type"` // Not mapped
	ClientBuild string `parquet:"client_build"` // Not mapped
	ClientDigProductID string `parquet:"client_dig_product_id"` // Not mapped
	ClientName string `parquet:"client_name"` // Not mapped
	Cookie string `parquet:"cookie"` // Not mapped
	DesktopHeight int32 `parquet:"desktop_height"` // Not mapped
	DesktopWidth int32 `parquet:"desktop_width"` // Not mapped
	EncryptionLevel string `parquet:"encryption_level"` // Not mapped
	EncryptionMethod string `parquet:"encryption_method"` // Not mapped
	KeyboardLayout string `parquet:"keyboard_layout"` // Not mapped
	RequestedColorDepth string `parquet:"requested_color_depth"` // Not mapped
	Result string `parquet:"result"` // Not mapped
	SecurityProtocol string `parquet:"security_protocol"` // Not mapped

	// =========================
	// FULL RAW LOG (complete original JSON)
	// =========================
	RawLog string `parquet:"raw_log"`
}

// SIP represents a normalized sip event following the three-layer model:
// 1) Conditional enrichment (controlled per log type via normalization.json enrich flags)
// 2) Normalized/promoted fields (from normalization.json mapping)
// 3) Unmapped raw fields (from schema.json, NOT in normalization.json)
// 4) raw_log (complete original JSON)
type SIP struct {
	// =========================
	// CONDITIONAL ENRICHMENT (Layer 3)
	// Applied based on normalization.json enrich flags (time, network)
	// Fields may be zero/empty if enrichment is disabled for this log type
	// =========================
	IngestTime       int64  `parquet:"ingest_time"`
	EventYear        int32  `parquet:"event_year"`
	EventMonth       int32  `parquet:"event_month"`
	EventDay         int32  `parquet:"event_day"`
	EventHour        int32  `parquet:"event_hour"`
	EventWeekday     int32  `parquet:"event_weekday"`
	EventType        string `parquet:"event_type"`
	EventClass       string `parquet:"event_class"`
	SrcIPIsPrivate   bool   `parquet:"src_ip_is_private"`
	DstIPIsPrivate   bool   `parquet:"dst_ip_is_private"`
	Direction        string `parquet:"direction"`
	Service          string `parquet:"service"`

	// =========================
	// NORMALIZED (PROMOTED) FIELDS (Layer 2)
	// From normalization.json mapping - these replace raw fields
	// DO NOT include raw versions
	// =========================
	DstIP string `parquet:"dst_ip"` // Promoted from: id.resp_h
	DstPort int32 `parquet:"dst_port"` // Promoted from: id.resp_p
	EventTime int64 `parquet:"event_time"` // Promoted from: ts
	FlowID string `parquet:"flow_id"` // Promoted from: uid
	SrcIP string `parquet:"src_ip"` // Promoted from: id.orig_h
	SrcPort int32 `parquet:"src_port"` // Promoted from: id.orig_p

	// =========================
	// RAW FIELDS (UNMAPPED) (Layer 1)
	// From schema.json, but NOT in normalization.json mapping
	// These are raw fields that were NOT promoted
	// =========================
	CallID string `parquet:"call_id"` // Not mapped
	ContentType string `parquet:"content_type"` // Not mapped
	Date string `parquet:"date"` // Not mapped
	Method string `parquet:"method"` // Not mapped
	ReplyTo string `parquet:"reply_to"` // Not mapped
	RequestBodyLen int32 `parquet:"request_body_len"` // Not mapped
	RequestFrom string `parquet:"request_from"` // Not mapped
	RequestPath string `parquet:"request_path"` // Not mapped
	RequestTo string `parquet:"request_to"` // Not mapped
	ResponseBodyLen int32 `parquet:"response_body_len"` // Not mapped
	ResponseFrom string `parquet:"response_from"` // Not mapped
	ResponsePath string `parquet:"response_path"` // Not mapped
	ResponseTo string `parquet:"response_to"` // Not mapped
	Seq string `parquet:"seq"` // Not mapped
	StatusCode int32 `parquet:"status_code"` // Not mapped
	StatusMsg string `parquet:"status_msg"` // Not mapped
	Subject string `parquet:"subject"` // Not mapped
	TransDepth int32 `parquet:"trans_depth"` // Not mapped
	Uri string `parquet:"uri"` // Not mapped
	UserAgent string `parquet:"user_agent"` // Not mapped
	Warning string `parquet:"warning"` // Not mapped

	// =========================
	// FULL RAW LOG (complete original JSON)
	// =========================
	RawLog string `parquet:"raw_log"`
}

// SMB_FILES represents a normalized smb_files event following the three-layer model:
// 1) Conditional enrichment (controlled per log type via normalization.json enrich flags)
// 2) Normalized/promoted fields (from normalization.json mapping)
// 3) Unmapped raw fields (from schema.json, NOT in normalization.json)
// 4) raw_log (complete original JSON)
type SMB_FILES struct {
	// =========================
	// CONDITIONAL ENRICHMENT (Layer 3)
	// Applied based on normalization.json enrich flags (time, network)
	// Fields may be zero/empty if enrichment is disabled for this log type
	// =========================
	IngestTime       int64  `parquet:"ingest_time"`
	EventYear        int32  `parquet:"event_year"`
	EventMonth       int32  `parquet:"event_month"`
	EventDay         int32  `parquet:"event_day"`
	EventHour        int32  `parquet:"event_hour"`
	EventWeekday     int32  `parquet:"event_weekday"`
	EventType        string `parquet:"event_type"`
	EventClass       string `parquet:"event_class"`
	SrcIPIsPrivate   bool   `parquet:"src_ip_is_private"`
	DstIPIsPrivate   bool   `parquet:"dst_ip_is_private"`
	Direction        string `parquet:"direction"`
	Service          string `parquet:"service"`

	// =========================
	// NORMALIZED (PROMOTED) FIELDS (Layer 2)
	// From normalization.json mapping - these replace raw fields
	// DO NOT include raw versions
	// =========================
	DstIP string `parquet:"dst_ip"` // Promoted from: id.resp_h
	EventTime int64 `parquet:"event_time"` // Promoted from: ts
	FlowID string `parquet:"flow_id"` // Promoted from: uid
	SrcIP string `parquet:"src_ip"` // Promoted from: id.orig_h

	// =========================
	// RAW FIELDS (UNMAPPED) (Layer 1)
	// From schema.json, but NOT in normalization.json mapping
	// These are raw fields that were NOT promoted
	// =========================
	Action string `parquet:"action"` // Not mapped
	Fuid string `parquet:"fuid"` // Not mapped
	IdOrigP int32 `parquet:"id_orig_p"` // Not mapped
	IdRespP int32 `parquet:"id_resp_p"` // Not mapped
	Name string `parquet:"name"` // Not mapped
	Path string `parquet:"path"` // Not mapped
	PrevName string `parquet:"prev_name"` // Not mapped
	Size int32 `parquet:"size"` // Not mapped
	TimesAccessed float64 `parquet:"times_accessed"` // Not mapped
	TimesChanged float64 `parquet:"times_changed"` // Not mapped
	TimesCreated float64 `parquet:"times_created"` // Not mapped
	TimesModified float64 `parquet:"times_modified"` // Not mapped

	// =========================
	// FULL RAW LOG (complete original JSON)
	// =========================
	RawLog string `parquet:"raw_log"`
}

// SMB_MAPPING represents a normalized smb_mapping event following the three-layer model:
// 1) Conditional enrichment (controlled per log type via normalization.json enrich flags)
// 2) Normalized/promoted fields (from normalization.json mapping)
// 3) Unmapped raw fields (from schema.json, NOT in normalization.json)
// 4) raw_log (complete original JSON)
type SMB_MAPPING struct {
	// =========================
	// CONDITIONAL ENRICHMENT (Layer 3)
	// Applied based on normalization.json enrich flags (time, network)
	// Fields may be zero/empty if enrichment is disabled for this log type
	// =========================
	IngestTime       int64  `parquet:"ingest_time"`
	EventYear        int32  `parquet:"event_year"`
	EventMonth       int32  `parquet:"event_month"`
	EventDay         int32  `parquet:"event_day"`
	EventHour        int32  `parquet:"event_hour"`
	EventWeekday     int32  `parquet:"event_weekday"`
	EventType        string `parquet:"event_type"`
	EventClass       string `parquet:"event_class"`
	SrcIPIsPrivate   bool   `parquet:"src_ip_is_private"`
	DstIPIsPrivate   bool   `parquet:"dst_ip_is_private"`
	Direction        string `parquet:"direction"`
	Service          string `parquet:"service"`

	// =========================
	// NORMALIZED (PROMOTED) FIELDS (Layer 2)
	// From normalization.json mapping - these replace raw fields
	// DO NOT include raw versions
	// =========================
	DstIP string `parquet:"dst_ip"` // Promoted from: id.resp_h
	EventTime int64 `parquet:"event_time"` // Promoted from: ts
	FlowID string `parquet:"flow_id"` // Promoted from: uid
	SrcIP string `parquet:"src_ip"` // Promoted from: id.orig_h

	// =========================
	// RAW FIELDS (UNMAPPED) (Layer 1)
	// From schema.json, but NOT in normalization.json mapping
	// These are raw fields that were NOT promoted
	// =========================
	IdOrigP int32 `parquet:"id_orig_p"` // Not mapped
	IdRespP int32 `parquet:"id_resp_p"` // Not mapped
	NativeFileSystem string `parquet:"native_file_system"` // Not mapped
	Path string `parquet:"path"` // Not mapped
	ZeekService string `parquet:"service"` // Not mapped
	ShareType string `parquet:"share_type"` // Not mapped

	// =========================
	// FULL RAW LOG (complete original JSON)
	// =========================
	RawLog string `parquet:"raw_log"`
}

// SMTP represents a normalized smtp event following the three-layer model:
// 1) Conditional enrichment (controlled per log type via normalization.json enrich flags)
// 2) Normalized/promoted fields (from normalization.json mapping)
// 3) Unmapped raw fields (from schema.json, NOT in normalization.json)
// 4) raw_log (complete original JSON)
type SMTP struct {
	// =========================
	// CONDITIONAL ENRICHMENT (Layer 3)
	// Applied based on normalization.json enrich flags (time, network)
	// Fields may be zero/empty if enrichment is disabled for this log type
	// =========================
	IngestTime       int64  `parquet:"ingest_time"`
	EventYear        int32  `parquet:"event_year"`
	EventMonth       int32  `parquet:"event_month"`
	EventDay         int32  `parquet:"event_day"`
	EventHour        int32  `parquet:"event_hour"`
	EventWeekday     int32  `parquet:"event_weekday"`
	EventType        string `parquet:"event_type"`
	EventClass       string `parquet:"event_class"`
	SrcIPIsPrivate   bool   `parquet:"src_ip_is_private"`
	DstIPIsPrivate   bool   `parquet:"dst_ip_is_private"`
	Direction        string `parquet:"direction"`
	Service          string `parquet:"service"`

	// =========================
	// NORMALIZED (PROMOTED) FIELDS (Layer 2)
	// From normalization.json mapping - these replace raw fields
	// DO NOT include raw versions
	// =========================
	DstIP string `parquet:"dst_ip"` // Promoted from: id.resp_h
	DstPort int32 `parquet:"dst_port"` // Promoted from: id.resp_p
	EventTime int64 `parquet:"event_time"` // Promoted from: ts
	FlowID string `parquet:"flow_id"` // Promoted from: uid
	SrcIP string `parquet:"src_ip"` // Promoted from: id.orig_h
	SrcPort int32 `parquet:"src_port"` // Promoted from: id.orig_p

	// =========================
	// RAW FIELDS (UNMAPPED) (Layer 1)
	// From schema.json, but NOT in normalization.json mapping
	// These are raw fields that were NOT promoted
	// =========================
	Cc string `parquet:"cc"` // Not mapped
	Date string `parquet:"date"` // Not mapped
	FirstReceived string `parquet:"first_received"` // Not mapped
	From string `parquet:"from"` // Not mapped
	Fuids string `parquet:"fuids"` // Not mapped
	Helo string `parquet:"helo"` // Not mapped
	InReplyTo string `parquet:"in_reply_to"` // Not mapped
	IsWebmail bool `parquet:"is_webmail"` // Not mapped
	LastReply string `parquet:"last_reply"` // Not mapped
	Mailfrom string `parquet:"mailfrom"` // Not mapped
	MsgID string `parquet:"msg_id"` // Not mapped
	Path string `parquet:"path"` // Not mapped
	Rcptto string `parquet:"rcptto"` // Not mapped
	ReplyTo string `parquet:"reply_to"` // Not mapped
	SecondReceived string `parquet:"second_received"` // Not mapped
	Subject string `parquet:"subject"` // Not mapped
	Tls bool `parquet:"tls"` // Not mapped
	To string `parquet:"to"` // Not mapped
	TransDepth int32 `parquet:"trans_depth"` // Not mapped
	UserAgent string `parquet:"user_agent"` // Not mapped
	XOriginatingIP string `parquet:"x_originating_ip"` // Not mapped

	// =========================
	// FULL RAW LOG (complete original JSON)
	// =========================
	RawLog string `parquet:"raw_log"`
}

// SNMP represents a normalized snmp event following the three-layer model:
// 1) Conditional enrichment (controlled per log type via normalization.json enrich flags)
// 2) Normalized/promoted fields (from normalization.json mapping)
// 3) Unmapped raw fields (from schema.json, NOT in normalization.json)
// 4) raw_log (complete original JSON)
type SNMP struct {
	// =========================
	// CONDITIONAL ENRICHMENT (Layer 3)
	// Applied based on normalization.json enrich flags (time, network)
	// Fields may be zero/empty if enrichment is disabled for this log type
	// =========================
	IngestTime       int64  `parquet:"ingest_time"`
	EventYear        int32  `parquet:"event_year"`
	EventMonth       int32  `parquet:"event_month"`
	EventDay         int32  `parquet:"event_day"`
	EventHour        int32  `parquet:"event_hour"`
	EventWeekday     int32  `parquet:"event_weekday"`
	EventType        string `parquet:"event_type"`
	EventClass       string `parquet:"event_class"`
	SrcIPIsPrivate   bool   `parquet:"src_ip_is_private"`
	DstIPIsPrivate   bool   `parquet:"dst_ip_is_private"`
	Direction        string `parquet:"direction"`
	Service          string `parquet:"service"`

	// =========================
	// NORMALIZED (PROMOTED) FIELDS (Layer 2)
	// From normalization.json mapping - these replace raw fields
	// DO NOT include raw versions
	// =========================
	DstIP string `parquet:"dst_ip"` // Promoted from: id.resp_h
	DstPort int32 `parquet:"dst_port"` // Promoted from: id.resp_p
	EventTime int64 `parquet:"event_time"` // Promoted from: ts
	FlowID string `parquet:"flow_id"` // Promoted from: uid
	SrcIP string `parquet:"src_ip"` // Promoted from: id.orig_h
	SrcPort int32 `parquet:"src_port"` // Promoted from: id.orig_p

	// =========================
	// RAW FIELDS (UNMAPPED) (Layer 1)
	// From schema.json, but NOT in normalization.json mapping
	// These are raw fields that were NOT promoted
	// =========================
	Community string `parquet:"community"` // Not mapped
	DisplayString string `parquet:"display_string"` // Not mapped
	Duration float64 `parquet:"duration"` // Not mapped
	GetBulkRequests int32 `parquet:"get_bulk_requests"` // Not mapped
	GetRequests int32 `parquet:"get_requests"` // Not mapped
	GetResponses int32 `parquet:"get_responses"` // Not mapped
	SetRequests int32 `parquet:"set_requests"` // Not mapped
	UpSince float64 `parquet:"up_since"` // Not mapped
	Version string `parquet:"version"` // Not mapped

	// =========================
	// FULL RAW LOG (complete original JSON)
	// =========================
	RawLog string `parquet:"raw_log"`
}

// SSH represents a normalized ssh event following the three-layer model:
// 1) Conditional enrichment (controlled per log type via normalization.json enrich flags)
// 2) Normalized/promoted fields (from normalization.json mapping)
// 3) Unmapped raw fields (from schema.json, NOT in normalization.json)
// 4) raw_log (complete original JSON)
type SSH struct {
	// =========================
	// CONDITIONAL ENRICHMENT (Layer 3)
	// Applied based on normalization.json enrich flags (time, network)
	// Fields may be zero/empty if enrichment is disabled for this log type
	// =========================
	IngestTime       int64  `parquet:"ingest_time"`
	EventYear        int32  `parquet:"event_year"`
	EventMonth       int32  `parquet:"event_month"`
	EventDay         int32  `parquet:"event_day"`
	EventHour        int32  `parquet:"event_hour"`
	EventWeekday     int32  `parquet:"event_weekday"`
	EventType        string `parquet:"event_type"`
	EventClass       string `parquet:"event_class"`
	SrcIPIsPrivate   bool   `parquet:"src_ip_is_private"`
	DstIPIsPrivate   bool   `parquet:"dst_ip_is_private"`
	Direction        string `parquet:"direction"`
	Service          string `parquet:"service"`

	// =========================
	// NORMALIZED (PROMOTED) FIELDS (Layer 2)
	// From normalization.json mapping - these replace raw fields
	// DO NOT include raw versions
	// =========================
	DstIP string `parquet:"dst_ip"` // Promoted from: id.resp_h
	DstPort int32 `parquet:"dst_port"` // Promoted from: id.resp_p
	EventTime int64 `parquet:"event_time"` // Promoted from: ts
	FlowID string `parquet:"flow_id"` // Promoted from: uid
	SrcIP string `parquet:"src_ip"` // Promoted from: id.orig_h
	SrcPort int32 `parquet:"src_port"` // Promoted from: id.orig_p

	// =========================
	// RAW FIELDS (UNMAPPED) (Layer 1)
	// From schema.json, but NOT in normalization.json mapping
	// These are raw fields that were NOT promoted
	// =========================
	AuthAttempts int32 `parquet:"auth_attempts"` // Not mapped
	AuthSuccess bool `parquet:"auth_success"` // Not mapped
	CipherAlg string `parquet:"cipher_alg"` // Not mapped
	Client string `parquet:"client"` // Not mapped
	CompressionAlg string `parquet:"compression_alg"` // Not mapped
	ZeekDirection string `parquet:"direction"` // Not mapped
	HostKey string `parquet:"host_key"` // Not mapped
	HostKeyAlg string `parquet:"host_key_alg"` // Not mapped
	KexAlg string `parquet:"kex_alg"` // Not mapped
	MacAlg string `parquet:"mac_alg"` // Not mapped
	Server string `parquet:"server"` // Not mapped
	Version int32 `parquet:"version"` // Not mapped

	// =========================
	// FULL RAW LOG (complete original JSON)
	// =========================
	RawLog string `parquet:"raw_log"`
}

// SSL represents a normalized ssl event following the three-layer model:
// 1) Conditional enrichment (controlled per log type via normalization.json enrich flags)
// 2) Normalized/promoted fields (from normalization.json mapping)
// 3) Unmapped raw fields (from schema.json, NOT in normalization.json)
// 4) raw_log (complete original JSON)
type SSL struct {
	// =========================
	// CONDITIONAL ENRICHMENT (Layer 3)
	// Applied based on normalization.json enrich flags (time, network)
	// Fields may be zero/empty if enrichment is disabled for this log type
	// =========================
	IngestTime       int64  `parquet:"ingest_time"`
	EventYear        int32  `parquet:"event_year"`
	EventMonth       int32  `parquet:"event_month"`
	EventDay         int32  `parquet:"event_day"`
	EventHour        int32  `parquet:"event_hour"`
	EventWeekday     int32  `parquet:"event_weekday"`
	EventType        string `parquet:"event_type"`
	EventClass       string `parquet:"event_class"`
	SrcIPIsPrivate   bool   `parquet:"src_ip_is_private"`
	DstIPIsPrivate   bool   `parquet:"dst_ip_is_private"`
	Direction        string `parquet:"direction"`
	Service          string `parquet:"service"`

	// =========================
	// NORMALIZED (PROMOTED) FIELDS (Layer 2)
	// From normalization.json mapping - these replace raw fields
	// DO NOT include raw versions
	// =========================
	DstIP string `parquet:"dst_ip"` // Promoted from: id.resp_h
	DstPort int32 `parquet:"dst_port"` // Promoted from: id.resp_p
	EventTime int64 `parquet:"event_time"` // Promoted from: ts
	FlowID string `parquet:"flow_id"` // Promoted from: uid
	SrcIP string `parquet:"src_ip"` // Promoted from: id.orig_h
	SrcPort int32 `parquet:"src_port"` // Promoted from: id.orig_p

	// =========================
	// RAW FIELDS (UNMAPPED) (Layer 1)
	// From schema.json, but NOT in normalization.json mapping
	// These are raw fields that were NOT promoted
	// =========================
	CertChainFuids string `parquet:"cert_chain_fuids"` // Not mapped
	Cipher string `parquet:"cipher"` // Not mapped
	ClientCertChainFuids string `parquet:"client_cert_chain_fuids"` // Not mapped
	ClientIssuer string `parquet:"client_issuer"` // Not mapped
	ClientSubject string `parquet:"client_subject"` // Not mapped
	Curve string `parquet:"curve"` // Not mapped
	Established bool `parquet:"established"` // Not mapped
	Issuer string `parquet:"issuer"` // Not mapped
	Ja3 string `parquet:"ja3"` // Not mapped
	Ja3s string `parquet:"ja3s"` // Not mapped
	LastAlert string `parquet:"last_alert"` // Not mapped
	NextProtocol string `parquet:"next_protocol"` // Not mapped
	Resumed bool `parquet:"resumed"` // Not mapped
	ServerName string `parquet:"server_name"` // Not mapped
	SslHistory string `parquet:"ssl_history"` // Not mapped
	Subject string `parquet:"subject"` // Not mapped
	ValidationStatus string `parquet:"validation_status"` // Not mapped
	Version string `parquet:"version"` // Not mapped

	// =========================
	// FULL RAW LOG (complete original JSON)
	// =========================
	RawLog string `parquet:"raw_log"`
}

// TUNNEL represents a normalized tunnel event following the three-layer model:
// 1) Conditional enrichment (controlled per log type via normalization.json enrich flags)
// 2) Normalized/promoted fields (from normalization.json mapping)
// 3) Unmapped raw fields (from schema.json, NOT in normalization.json)
// 4) raw_log (complete original JSON)
type TUNNEL struct {
	// =========================
	// CONDITIONAL ENRICHMENT (Layer 3)
	// Applied based on normalization.json enrich flags (time, network)
	// Fields may be zero/empty if enrichment is disabled for this log type
	// =========================
	IngestTime       int64  `parquet:"ingest_time"`
	EventYear        int32  `parquet:"event_year"`
	EventMonth       int32  `parquet:"event_month"`
	EventDay         int32  `parquet:"event_day"`
	EventHour        int32  `parquet:"event_hour"`
	EventWeekday     int32  `parquet:"event_weekday"`
	EventType        string `parquet:"event_type"`
	EventClass       string `parquet:"event_class"`

	// =========================
	// NORMALIZED (PROMOTED) FIELDS (Layer 2)
	// From normalization.json mapping - these replace raw fields
	// DO NOT include raw versions
	// =========================
	EventTime int64 `parquet:"event_time"` // Promoted from: ts
	FlowID string `parquet:"flow_id"` // Promoted from: uid

	// =========================
	// RAW FIELDS (UNMAPPED) (Layer 1)
	// From schema.json, but NOT in normalization.json mapping
	// These are raw fields that were NOT promoted
	// =========================
	Action string `parquet:"action"` // Not mapped
	IdOrigH string `parquet:"id_orig_h"` // Not mapped
	IdOrigP int32 `parquet:"id_orig_p"` // Not mapped
	IdRespH string `parquet:"id_resp_h"` // Not mapped
	IdRespP int32 `parquet:"id_resp_p"` // Not mapped
	TunnelType string `parquet:"tunnel_type"` // Not mapped

	// =========================
	// FULL RAW LOG (complete original JSON)
	// =========================
	RawLog string `parquet:"raw_log"`
}

// ToParquetTime converts Go time to milliseconds since epoch
func ToParquetTime(t time.Time) int64 {
	return t.UnixNano() / int64(time.Millisecond)
}
