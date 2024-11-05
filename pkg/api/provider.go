package api

type Record struct {
	ID      string
	Type    string
	Name    string
	Content string
	TTL     int
}

type CloudProvider interface {
	ListDNSRecordsFiltered(zoneName string, recordName string) (map[string]string, error)
	UpdateDNSRecord(zone string, record Record) (map[string]string, error)
	FillRecord(generic map[string]string, record *Record)
	// GetDNSRecord(zoneName string, recordName string) (map[string]string, error)
}
