package api

import (
	"context"
	"fmt"
	"log"

	"github.com/cloudflare/cloudflare-go"
)

var (
	cfApi *cloudflare.API
)

type CloudflareCredentials struct {
	ApiKey          string
	AccountEmail    string
	CloudflareToken string
}

func InitializeAPI(creds *CloudflareCredentials) error {
	var err error
	if creds.AccountEmail != "" && creds.ApiKey != "" {
		cfApi, err = cloudflare.New(creds.ApiKey, creds.AccountEmail)
	} else {
		cfApi, err = cloudflare.NewWithAPIToken(creds.CloudflareToken)
	}
	return err
}

func ListDNSRecordsFiltered(zoneName string, wantedRecordName string) (cloudflare.DNSRecord, string, error) {
	zoneId, err := cfApi.ZoneIDByName(zoneName)
	if err != nil {
		log.Fatalf("unable to get zone by name %s", zoneName)
	}
	records, _, err := cfApi.ListDNSRecords(context.TODO(), cloudflare.ZoneIdentifier(zoneId), cloudflare.ListDNSRecordsParams{})
	if err != nil {
		log.Fatalf("unable to list dns records for domain %s", wantedRecordName)
	}

	for _, record := range records {
		if record.Name == wantedRecordName && record.Type == "A" {
			return record, zoneId, nil
		}
	}
	return cloudflare.DNSRecord{}, "", fmt.Errorf("record %s not found", wantedRecordName)
}

func UpdateDNSRecord(ipifyResult string, zoneId string, record cloudflare.DNSRecord) error {
	err := cfApi.UpdateDNSRecord(context.TODO(), cloudflare.ZoneIdentifier(zoneId), cloudflare.UpdateDNSRecordParams{
		Type:     record.Type,
		Name:     record.Name,
		ID:       record.ID,
		Proxied:  record.Proxied,
		Priority: record.Priority,
		TTL:      record.TTL,
		Content:  ipifyResult,
	})

	if err != nil {
		log.Println("unable to update dns record")
	}
	log.Printf("record updated successfully\n")
	return err
}
