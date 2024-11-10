package cloudprovider

type Record struct {
	ID      string
	Type    string
	Name    string
	Content string
	TTL     int
}

type Provider interface {
	InitializeRecord(zone string, record Record) (map[string]string, error)
	GetDNSRecord(zoneName string, recordName string) (map[string]string, error)
	UpdateDNSRecord(zone string, record Record) (map[string]string, error)
	ListDNSRecordsFiltered(zoneName string, recordName string) (map[string]string, error)
	FillRecord(generic map[string]string, record *Record)
}
