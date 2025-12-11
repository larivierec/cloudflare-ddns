package cloudprovider

type Record struct {
	ID      string
	Type    string
	Name    string
	Content string
	TTL     int
	Proxied bool
}

type Provider interface {
	GetDNSRecord(zone, recordName string) (*Record, error)
	CreateDNSRecord(zone string, record *Record) (*Record, error)
	UpdateDNSRecord(zone string, record *Record) (*Record, error)
	DeleteDNSRecord(zone, recordID string) error
}
