package cloudflare

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/larivierec/cloudflare-ddns/pkg/cloudprovider"
)

var cloudflareAPIUrl = "https://api.cloudflare.com/client/v4"

type CloudflareProvider struct {
	config Configuration
}

type Configuration struct {
	ApiKey          string
	AccountEmail    string
	CloudflareToken string
}

type Zone struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func NewCloudflareProvider() *CloudflareProvider {
	config := Configuration{
		ApiKey:          os.Getenv("API_KEY"),
		AccountEmail:    os.Getenv("ACCOUNT_EMAIL"),
		CloudflareToken: os.Getenv("ACCOUNT_TOKEN"),
	}
	return &CloudflareProvider{config: config}
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

	return nil, fmt.Errorf("[ListDNSRecordsFiltered] record %s not found", recordName)
}

func (c *CloudflareProvider) UpdateDNSRecord(zone string, rec *cloudprovider.Record) (*cloudprovider.Record, error) {
	zoneID, err := c.getZoneID(zone)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("%s/zones/%s/dns_records/%s", cloudflareAPIUrl, zoneID, rec.ID)
	data := map[string]interface{}{
		"type":    rec.Type,
		"name":    rec.Name,
		"content": rec.Content,
		"ttl":     rec.TTL,
		"proxied": rec.Proxied,
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

	log.Printf("[UpdateDNSRecord] record %s updated successfully\n", result.Result["name"].(string))
	return c.mapToRecord(result.Result), nil
}

func (c *CloudflareProvider) GetDNSRecord(zone string, recordName string) (*cloudprovider.Record, error) {
	zoneID, err := c.getZoneID(zone)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("%s/zones/%s/dns_records?name=%s", cloudflareAPIUrl, zoneID, recordName)
	req, _ := http.NewRequest("GET", url, nil)
	c.setHeaders(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error retrieving DNS record: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get DNS record: %s", string(body))
	}

	var result struct {
		Result []map[string]interface{} `json:"result"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	for _, record := range result.Result {
		if strings.EqualFold(record["name"].(string), recordName) {
			return c.mapToRecord(record), nil
		}
	}

	return nil, fmt.Errorf("[GetDNSRecord] record %s not found", recordName)
}

func (c *CloudflareProvider) CreateDNSRecord(zone string, rec *cloudprovider.Record) (*cloudprovider.Record, error) {
	zoneID, err := c.getZoneID(zone)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("%s/zones/%s/dns_records", cloudflareAPIUrl, zoneID)
	data := map[string]interface{}{
		"type":    rec.Type,
		"name":    rec.Name,
		"content": rec.Content,
		"ttl":     rec.TTL,
		"proxied": rec.Proxied,
	}

	jsonData, _ := json.Marshal(data)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	c.setHeaders(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error creating DNS record: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to create DNS record: %s", string(body))
	}

	var result struct {
		Result map[string]interface{} `json:"result"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	log.Printf("DNS record %s created successfully", result.Result["name"].(string))
	return c.mapToRecord(result.Result), nil
}

func (c *CloudflareProvider) DeleteDNSRecord(zone, recordID string) error {
	zoneID, err := c.getZoneID(zone)
	if err != nil {
		return err
	}
	url := fmt.Sprintf("%s/zones/%s/dns_records/%s", cloudflareAPIUrl, zoneID, recordID)
	req, _ := http.NewRequest("DELETE", url, nil)
	c.setHeaders(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("error deleting DNS record: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete DNS record: %s", string(body))
	}

	log.Printf("DNS record %s deleted successfully", recordID)
	return nil
}

func (c *CloudflareProvider) setHeaders(req *http.Request) {
	if c.config.CloudflareToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.config.CloudflareToken)
	} else {
		req.Header.Set("X-Auth-Email", c.config.AccountEmail)
		req.Header.Set("X-Auth-Key", c.config.ApiKey)
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
	if record != nil {
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
	return nil
}

func (c *CloudflareProvider) mapToRecord(data map[string]interface{}) *cloudprovider.Record {
	record := &cloudprovider.Record{
		ID:      fmt.Sprintf("%v", data["id"]),
		Type:    fmt.Sprintf("%v", data["type"]),
		Name:    fmt.Sprintf("%v", data["name"]),
		Content: fmt.Sprintf("%v", data["content"]),
	}

	if ttl, ok := data["ttl"].(float64); ok {
		record.TTL = int(ttl)
	}
	
	if proxied, ok := data["proxied"].(bool); ok {
		record.Proxied = proxied
	}

	return record
}
