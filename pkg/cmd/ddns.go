package ddns

import (
	"context"
	"log"
	"os"

	"github.com/cloudflare/cloudflare-go"
	"github.com/larivierec/cloudflare-ddns/pkg/api"
	"github.com/larivierec/cloudflare-ddns/pkg/ipify"
)

var (
	cachedIpInfo ipify.IpInfo
	creds        api.CloudflareCredentials
)

func Start() {
	creds.ApiKey = os.Getenv("API_KEY")
	creds.AccountEmail = os.Getenv("ACCOUNT_EMAIL")
	creds.CloudflareToken = os.Getenv("ACCOUNT_TOKEN")
	zoneName := os.Args[2]
	recordName := os.Args[3]

	ipifyResult, err := ipify.GetCurrentIP()

	if err != nil {
		log.Fatalf("Unable to get ipify ip, aborting.")
	}

	if cachedIpInfo.Ip != ipifyResult.Ip {
		cachedIpInfo.Ip = ipifyResult.Ip
		var cloudflareRecord cloudflare.DNSRecord
		apiObj, err := api.CloudflareApi(creds)

		if err != nil {
			log.Fatalf("unable to initialize cloudflare api")
		}

		zoneId, err := apiObj.ZoneIDByName(zoneName)
		if err != nil {
			log.Fatalf("unable to get zone by name %s", zoneName)
		}
		records, _, err := apiObj.ListDNSRecords(context.TODO(), cloudflare.ZoneIdentifier(zoneId), cloudflare.ListDNSRecordsParams{})
		if err != nil {
			log.Fatalf("unable to list dns records for domain %s", recordName)
		}

		for _, record := range records {
			if record.Name == recordName && record.Type == "A" {
				cloudflareRecord = record
				break
			}
		}

		if cachedIpInfo.Ip != cloudflareRecord.Content {
			err = apiObj.UpdateDNSRecord(context.TODO(), cloudflare.ZoneIdentifier(zoneId), cloudflare.UpdateDNSRecordParams{
				Type:     cloudflareRecord.Type,
				Name:     cloudflareRecord.Name,
				ID:       cloudflareRecord.ID,
				Proxied:  cloudflareRecord.Proxied,
				Priority: cloudflareRecord.Priority,
				TTL:      cloudflareRecord.TTL,
				Content:  ipifyResult.Ip,
			})

			if err != nil {
				log.Println("unable to update dns record")
			}
			log.Printf("record updated successfully %s\n", err)
		}
	} else {
		log.Println("IPs are the same")
	}
}
