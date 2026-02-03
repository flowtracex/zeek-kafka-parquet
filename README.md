# Kafka → Parquet Pipeline

A production-ready pipeline that consumes Zeek network logs from Kafka, normalizes and enriches them, and outputs to Parquet files and/or Kafka topics.

## Features

- **Multi-log-type support**: Processes 18+ Zeek log types (DNS, CONN, HTTP, SSL, SSH, FTP, SMTP, etc.)
- **Dual output**: Writes to Parquet files and/or Kafka topics
- **Three-layer data model**: Raw fields, normalized fields, and enriched fields
- **Configurable flushing**: Size, time, and event count-based flush conditions
- **Asynchronous processing**: Non-blocking writes with background flushing
- **Health monitoring**: Periodic metrics (memory, buffer usage, throughput)

## Quick Start

### Build

```bash
go build -o pipeline main.go
```

### Run

```bash
./pipeline --config config/config.json
```

## Configuration

### config.json

```json
{
  "kafka": {
    "brokers": ["localhost:9092"],
    "topic": "zeek-raw",
    "group_id": "parquet-writer-group",
    "max_poll_records": 1000
  },
  "output": {
    "parquet": {
      "enabled": true,
      "base_path": "./output",
      "file_prefix": "events"
    },
    "kafka": {
      "enabled": true,
      "brokers": ["localhost:9092"],
      "topic": "zeek-normalized",
      "compression": "lz4"
    }
  },
  "log": {
    "path": "./log"
  },
  "write": {
    "flush_buffer_mb": 1,
    "flush_interval_seconds": 60,
    "flush_event_count": 1000
  }
}
```

**Flush Conditions** (OR logic - any one triggers flush):
- `flush_buffer_mb`: Flush when buffer reaches this size (MB)
- `flush_interval_seconds`: Flush every N seconds (0 to disable)
- `flush_event_count`: Flush after N events (0 to disable)

### normalization.json

Defines field promotion and enrichment per log type:

```json
{
  "dns": {
    "source": "zeek_dns",
    "promote": {
      "ts": "event_time",
      "uid": "flow_id",
      "id.orig_h": "src_ip",
      "id.resp_h": "dst_ip",
      "proto": "protocol"
    },
    "static": {
      "event_type": "dns",
      "event_class": "dns"
    },
    "enrich": {
      "time": true,
      "network": true
    }
  }
}
```

### schema.json

Defines raw field structure for each Zeek log type. Used to generate Parquet schemas.

## Architecture

```
Kafka (zeek-raw)
    ↓
Consumer
    ↓
Normalizer (field promotion)
    ↓
Enricher (time + network enrichment)
    ↓
Fan-out
    ├──→ Parquet Writers (per log_type, async flush)
    └──→ Kafka Producer (structured JSON output)
```

## Project Structure

```
.
├── main.go                 # Pipeline entry point
├── pipeline               # Compiled binary
├── config/                # Configuration files
│   ├── config.json        # Main configuration
│   ├── normalization.json # Field promotion rules
│   └── schema.json        # Raw field definitions
├── core/                  # Core pipeline components
│   ├── kafka.go          # Kafka consumer
│   ├── normalize.go      # Field normalization
│   ├── enrich.go         # Data enrichment
│   ├── dispatch.go       # Event routing
│   ├── parquet.go        # Parquet writer
│   ├── kafka_producer.go # Kafka output producer
│   ├── fanout.go         # Output fan-out
│   └── logger.go         # Logging system
├── schema/                # Generated Parquet schemas
│   └── events.go         # Auto-generated from schema.json
├── test/                 # Test scripts
├── log/                  # Runtime logs
└── output/               # Parquet output files
```

## Output Formats

### Parquet Output

Partitioned by log type and time:

```
output/
├── dns/
│   └── year=2026/month=02/day=02/hour=10/
│       └── events_dns_1.parquet
├── conn/
│   └── year=2026/month=02/day=02/hour=10/
│       └── events_conn_1.parquet
└── ...
```

### Kafka Output

Structured JSON format with three layers:

```json
{
  "source": "zeek",
  "log_type": "dns",
  "raw": {
    "ts": 1769868799.213927,
    "uid": "CfqDt31quyW8AAJct6",
    "id.orig_h": "10.128.0.4",
    ...
  },
  "normalized": {
    "event_time": 1769868799213,
    "ingest_time": 1770089817352,
    "flow_id": "CfqDt31quyW8AAJct6",
    "src_ip": "10.128.0.4",
    "dst_ip": "216.239.34.174",
    ...
  },
  "enriched": {
    "src_ip_is_private": true,
    "dst_ip_is_private": false,
    "direction": "outbound",
    "event_year": 2026,
    ...
  }
}
```

## Data Model

### Three-Layer Architecture

1. **Raw**: All fields exactly as they come from Zeek
2. **Normalized**: Field promotion to canonical names (e.g., `id.orig_h` → `src_ip`)
3. **Enriched**: Derived fields (time components, network analysis, direction)

## Development

### Regenerate Schema

After updating `schema.json` or `normalization.json`:

```bash
go run generate_schema.go > schema/events.go
go build -o pipeline main.go
```

### Testing

```bash
go run test/test_parquet_validation.go output
```

## Requirements

- Go 1.21+
- Kafka broker (tested with Kafka 2.8+)
- Zeek logs in Kafka topic (nested JSON format: `{"dns": {...}}` or `{"conn": {...}}`)

## Dependencies

- `github.com/segmentio/kafka-go` - Kafka client
- `github.com/parquet-go/parquet-go` - Parquet library

## Monitoring

Health metrics are logged every 30 seconds to `log/pipeline.log`:
- Memory usage (RSS and heap)
- Buffer usage (bytes and percentage)
- Throughput (events per second)
- Flush counts and error counts

## Logging

Structured logs with format:
```
TIMESTAMP | LEVEL | COMPONENT | MESSAGE | CONTEXT
```

Example:
```
2026-02-03T10:30:00Z | INFO | startup | configuration loaded | kafka_topic=zeek-raw
2026-02-03T10:30:30Z | INFO | health | memory=612MB buffer=38% eps=8200
```
