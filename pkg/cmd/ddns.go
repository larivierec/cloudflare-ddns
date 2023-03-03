package ddns

import (
	"context"
	"log"
	"os"

	"github.com/cloudflare/cloudflare-go"
	"github.com/lariviere.c/cloudflare-ddns/pkg/api"
	"github.com/lariviere.c/cloudflare-ddns/pkg/ipify"
)

var (
	cachedIpInfo ipify.IpInfo
	creds        api.CloudflareCredentials
	apiObj       *cloudflare.API
)

func Start() {
	creds.ApiKey = os.Getenv("API_KEY")
	creds.AccountEmail = os.Getenv("ACCOUNT_EMAIL")
	creds.CloudflareToken = os.Getenv("ACCOUNT_TOKEN")
	zoneName := os.Args[2]
	recordName := os.Args[3]

	apiObj, err := api.CloudflareApi(creds)

	if err != nil {
		log.Fatalf("Unable to initialize cloudflare api")
	}

	if cachedIpInfo.Ip == "" {
		zoneId, err := apiObj.ZoneIDByName(zoneName)
		if err != nil {
			log.Fatalf("Unable to get zone by name %s", zoneName)
		}
		records, _, err := apiObj.ListDNSRecords(context.TODO(), cloudflare.ZoneIdentifier(zoneId), cloudflare.ListDNSRecordsParams{})
		if err != nil {
			log.Fatalf("Unable to list dns records for domain %s", recordName)
		}
		for _, record := range records {
			if record.Name == recordName && record.Type == "A" {
				cachedIpInfo.Ip = record.Content
				break
			}
		}
	}

	ipifyResult, err := ipify.GetCurrentIP()

	if err != nil {
		log.Fatalf("Unable to get ipify ip, aborting.")
	}

	if cachedIpInfo.Ip != ipifyResult.Ip {
		cachedIpInfo.Ip = ipifyResult.Ip
		// update record
	}

	log.Println("IPs are the same, quitting")
}
