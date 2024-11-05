package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

var cloudflareAPIUrl = "https://api.cloudflare.com/client/v4"

type CloudflareProvider struct {
	creds CloudflareCredentials
}

type CloudflareCredentials struct {
	ApiKey          string
	AccountEmail    string
	CloudflareToken string
}

type Zone struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func NewCloudflareProvider(creds *CloudflareCredentials) (*CloudflareProvider, error) {
	return &CloudflareProvider{creds: *creds}, nil
}

func (c *CloudflareProvider) ListDNSRecordsFiltered(zoneName string, recordName string) (map[string]string, error) {
	zoneID, err := c.getZoneID(zoneName)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/zones/%s/dns_records?type=A&name=%s", cloudflareAPIUrl, zoneID, recordName)
	req, _ := http.NewRequest("GET", url, nil)
	c.setHeaders(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error retrieving DNS records: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to list DNS records: %s", string(body))
	}

	var result struct {
		Result []map[string]interface{} `json:"result"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	for _, record := range result.Result {
		if strings.EqualFold(record["name"].(string), recordName) && strings.EqualFold(record["type"].(string), "A") {
			return c.convertToGenericMap(record), nil
		}
	}

	return nil, fmt.Errorf("record %s not found", recordName)
}

func (c *CloudflareProvider) UpdateDNSRecord(zone string, rec Record) (map[string]string, error) {
	zoneID, err := c.getZoneID(zone)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("%s/zones/%s/dns_records/%s", cloudflareAPIUrl, zoneID, rec.ID)
	data := map[string]interface{}{
		"type":    rec.Type,
		"name":    rec.Name,
		"content": rec.Content,
		"proxied": true,
	}

	jsonData, _ := json.Marshal(data)
	req, _ := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	c.setHeaders(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error updating DNS record: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to update DNS record: %s", string(body))
	}

	var result struct {
		Result map[string]interface{} `json:"result"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	log.Printf("record %s updated successfully\n", result.Result["name"].(string))
	return c.convertToGenericMap(result.Result), nil
}

func (c *CloudflareProvider) setHeaders(req *http.Request) {
	if c.creds.CloudflareToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.creds.CloudflareToken)
	} else {
		req.Header.Set("X-Auth-Email", c.creds.AccountEmail)
		req.Header.Set("X-Auth-Key", c.creds.ApiKey)
	}
	req.Header.Set("Content-Type", "application/json")
}

func (c *CloudflareProvider) getZoneID(zoneName string) (string, error) {
	url := fmt.Sprintf("%s/zones?name=%s", cloudflareAPIUrl, zoneName)
	req, _ := http.NewRequest("GET", url, nil)
	c.setHeaders(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error retrieving zone ID: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to list zones: %s", string(body))
	}

	var result struct {
		Result []Zone `json:"result"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	if len(result.Result) == 0 {
		return "", fmt.Errorf("zone %s not found", zoneName)
	}

	return result.Result[0].ID, nil
}

func (c *CloudflareProvider) convertToGenericMap(record map[string]interface{}) map[string]string {
	// Generic map to standardize fields
	genericRecord := map[string]string{
		"id":      fmt.Sprintf("%v", record["id"]),
		"type":    fmt.Sprintf("%v", record["type"]),
		"name":    fmt.Sprintf("%v", record["name"]),
		"content": fmt.Sprintf("%v", record["content"]),
	}

	if proxied, ok := record["proxied"].(bool); ok {
		genericRecord["proxied"] = fmt.Sprintf("%v", proxied)
	}
	return genericRecord
}

func (c *CloudflareProvider) FillRecord(generic map[string]string, record *Record) {
	record.Content = generic["content"]
	record.ID = generic["id"]
	record.Name = generic["name"]
	record.Type = generic["type"]
}
