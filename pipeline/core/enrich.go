package core

import (
	"net"
	"time"
)

// EnrichedEvent extends NormalizedEvent with common enrichment fields
type EnrichedEvent struct {
	*NormalizedEvent

	// Time enrichment
	EventYear    int32
	EventMonth   int32
	EventDay     int32
	EventHour    int32
	EventWeekday int32

	// IP enrichment
	SrcIPIsPrivate bool
	DstIPIsPrivate bool

	// Direction and service
	Direction string
	Service   string
}

// Enricher applies common enrichment to normalized events
type Enricher struct{}

// NewEnricher creates a new enricher
func NewEnricher() *Enricher {
	return &Enricher{}
}

// Enrich applies common enrichment to a normalized event
// ONLY common enrichment - no log-type-specific enrichment
func (e *Enricher) Enrich(event *NormalizedEvent) *EnrichedEvent {
	enriched := &EnrichedEvent{
		NormalizedEvent: event,
	}

	// Time enrichment from event_time
	if event.EventTime > 0 {
		t := time.Unix(event.EventTime/1000, (event.EventTime%1000)*int64(time.Millisecond))
		enriched.EventYear = int32(t.Year())
		enriched.EventMonth = int32(t.Month())
		enriched.EventDay = int32(t.Day())
		enriched.EventHour = int32(t.Hour())
		enriched.EventWeekday = int32(t.Weekday())
	}

	// IP enrichment
	enriched.SrcIPIsPrivate = isPrivateIP(event.SrcIP)
	enriched.DstIPIsPrivate = isPrivateIP(event.DstIP)

	// Direction derivation
	enriched.Direction = deriveDirection(enriched.SrcIPIsPrivate, enriched.DstIPIsPrivate)

	// Service mapping from dst_port
	enriched.Service = portToService(event.DstPort)

	return enriched
}

// isPrivateIP checks if an IP is in RFC1918 private ranges
func isPrivateIP(ipStr string) bool {
	if ipStr == "" {
		return false
	}

	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}

	if ip.To4() == nil {
		// IPv6 - not handling private IPv6 for simplicity
		return false
	}

	// RFC1918 private ranges
	_, private24, _ := net.ParseCIDR("10.0.0.0/8")
	_, private20, _ := net.ParseCIDR("172.16.0.0/12")
	_, private16, _ := net.ParseCIDR("192.168.0.0/16")

	return private24.Contains(ip) || private20.Contains(ip) || private16.Contains(ip)
}

// deriveDirection determines traffic direction
func deriveDirection(srcPrivate, dstPrivate bool) string {
	if srcPrivate && !dstPrivate {
		return "outbound" // private → public
	}
	if !srcPrivate && dstPrivate {
		return "inbound" // public → private
	}
	if srcPrivate && dstPrivate {
		return "internal" // private → private
	}
	return "external" // public → public
}

// portToService maps common ports to service names
func portToService(port int32) string {
	serviceMap := map[int32]string{
		53:   "dns",
		80:   "http",
		443:  "https",
		22:   "ssh",
		25:   "smtp",
		587:  "smtp",
		993:  "imaps",
		995:  "pop3s",
		1433: "mssql",
		3306: "mysql",
		5432: "postgresql",
		3389: "rdp",
		5900: "vnc",
		8080: "http-proxy",
		8443: "https-alt",
	}

	if service, ok := serviceMap[port]; ok {
		return service
	}
	return "unknown"
}

