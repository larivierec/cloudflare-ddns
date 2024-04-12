package api

import (
	"context"
	"fmt"
	"log"

	"github.com/cloudflare/cloudflare-go/v2"
	"github.com/cloudflare/cloudflare-go/v2/dns"
	"github.com/cloudflare/cloudflare-go/v2/option"
	"github.com/cloudflare/cloudflare-go/v2/zones"
)

var (
	client *cloudflare.Client
)

type CloudflareCredentials struct {
	ApiKey          string
	AccountEmail    string
	CloudflareToken string
}

func InitializeAPI(creds *CloudflareCredentials) {
	if creds.AccountEmail != "" && creds.ApiKey != "" {
		client = cloudflare.NewClient(
			option.WithAPIEmail(creds.AccountEmail),
			option.WithAPIKey(creds.ApiKey),
		)
	} else {
		client = cloudflare.NewClient(option.WithAPIToken(creds.CloudflareToken))
	}
}

func ListDNSRecordsFiltered(zoneName string, wantedRecordName string) (dns.Record, string, error) {
	zonePages, err := client.Zones.List(context.TODO(), zones.ZoneListParams{
		Name: cloudflare.F(zoneName),
	})

	if err != nil {
		return dns.Record{}, "", err
	}

	for _, zone := range zonePages.Result {
		recordPages, err := client.DNS.Records.List(context.TODO(), dns.RecordListParams{
			ZoneID: cloudflare.F(zone.ID),
		})

		if err != nil {
			return dns.Record{}, "", err
		}

		for _, record := range recordPages.Result {
			if record.Name == wantedRecordName && record.Type == "A" {
				return record, zone.ID, nil
			}
		}
	}

	return dns.Record{}, "", fmt.Errorf("record %s not found", wantedRecordName)
}

func UpdateDNSRecord(result string, zoneId string, record dns.Record) error {
	newRecord := dns.ARecordParam{
		Type:    cloudflare.F(dns.ARecordTypeA),
		Name:    cloudflare.F(record.Name),
		Proxied: cloudflare.F(record.Proxied),
		Content: cloudflare.F(result),
	}

	resp, err := client.DNS.Records.Update(context.TODO(), record.ID, dns.RecordUpdateParams{
		ZoneID: cloudflare.F(zoneId),
		Record: newRecord,
	})

	if err != nil {
		log.Println("unable to update dns record")
	}
	log.Printf("record %s updated successfully\n", resp.Name)
	return err
}
