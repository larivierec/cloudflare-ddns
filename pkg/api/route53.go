package api

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
)

type Route53Provider struct {
	svc *route53.Client
}

func NewRoute53Provider() (*Route53Provider, error) {
	// Load the default AWS configuration
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS SDK config: %v", err)
	}

	// Create a new Route53 client
	return &Route53Provider{
		svc: route53.NewFromConfig(cfg),
	}, nil
}

func (p *Route53Provider) ListDNSRecordsFiltered(zoneId string, recordName string) (map[string]string, error) {
	zone, err := p.getHostedZone(zoneId)
	if err != nil {
		return nil, err
	}

	records, err := p.svc.ListResourceRecordSets(context.TODO(), &route53.ListResourceRecordSetsInput{
		HostedZoneId:    zone.Id,
		StartRecordType: types.RRTypeA,
		StartRecordName: &recordName,
	})

	if err != nil {
		return nil, err
	}

	for _, record := range records.ResourceRecordSets {
		if strings.EqualFold(strings.Trim(*record.Name, "."), recordName) {
			return p.convertToGenericMap(record), nil
		}
	}

	return nil, fmt.Errorf("record %s not found in hosted zone %s", recordName, *zone.Name)
}

func (p *Route53Provider) UpdateDNSRecord(zoneId string, rec Record) (map[string]string, error) {
	zone, err := p.getHostedZone(zoneId)
	if err != nil {
		return nil, err
	}

	changeBatch := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &types.ChangeBatch{
			Changes: []types.Change{
				{
					Action: types.ChangeActionUpsert,
					ResourceRecordSet: &types.ResourceRecordSet{
						Name: aws.String(rec.Name),
						Type: types.RRType(rec.Type),
						TTL:  aws.Int64(int64(rec.TTL)),
						ResourceRecords: []types.ResourceRecord{
							{Value: aws.String(rec.Content)},
						},
					},
				},
			},
		},
	}

	resp, err := p.svc.ChangeResourceRecordSets(context.TODO(), &route53.ChangeResourceRecordSetsInput{
		HostedZoneId: aws.String(*zone.Id),
		ChangeBatch:  changeBatch.ChangeBatch,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to update DNS record: %v", err)
	}

	if resp.ChangeInfo.Status == types.ChangeStatusPending {
		log.Printf("record %s updated successfully to %s\n", rec.Name, rec.Content)
	}
	return map[string]string{
		"type":    rec.Type,
		"name":    rec.Name,
		"content": rec.Content,
	}, nil
}

func (p *Route53Provider) GetDNSRecord(zoneName, recordName string) (map[string]string, error) {
	hostedZone, err := p.getHostedZone(zoneName)
	if err != nil {
		return nil, err
	}

	result, err := p.svc.ListResourceRecordSets(context.TODO(), &route53.ListResourceRecordSetsInput{
		HostedZoneId:    aws.String(*hostedZone.Id),
		StartRecordName: aws.String(recordName),
		StartRecordType: types.RRTypeA, // Assuming A record here
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list DNS record: %v", err)
	}

	// Find the relevant record in the response
	for _, recordSet := range result.ResourceRecordSets {
		if aws.ToString(recordSet.Name) == recordName {
			return map[string]string{
				"content": aws.ToString(recordSet.ResourceRecords[0].Value),
				"ttl":     fmt.Sprintf("%d", aws.ToInt64(recordSet.TTL)),
			}, nil
		}
	}

	return nil, fmt.Errorf("DNS record %s not found", recordName)
}

func (p *Route53Provider) getHostedZone(zone string) (*types.HostedZone, error) {
	result, err := p.svc.GetHostedZone(context.TODO(), &route53.GetHostedZoneInput{
		Id: aws.String(zone),
	})

	if err != nil {
		log.Printf("hosted zone id %s not found, trying by name\n", zone)
		listZones, err := p.svc.ListHostedZonesByName(context.TODO(), &route53.ListHostedZonesByNameInput{
			DNSName: aws.String(zone),
		})

		if err != nil {
			return nil, fmt.Errorf("hosted zone name %s not found", zone)
		}

		for _, awsZone := range listZones.HostedZones {
			if strings.EqualFold(*awsZone.Name, zone) {
				return &awsZone, nil
			}
		}
	}

	return result.HostedZone, nil
}

func (c *Route53Provider) convertToGenericMap(record types.ResourceRecordSet) map[string]string {
	// Generic map to standardize fields
	genericRecord := map[string]string{
		// "id":      fmt.Sprintf("%v", record.["id"]),
		"type":    fmt.Sprintf("%v", record.Type),
		"name":    fmt.Sprintf("%v", *record.Name),
		"content": fmt.Sprintf("%v", *record.ResourceRecords[0].Value),
	}

	return genericRecord
}

func (c *Route53Provider) FillRecord(generic map[string]string, record *Record) {
	record.Content = generic["content"]
	record.Name = generic["name"]
	record.Type = generic["type"]
}
