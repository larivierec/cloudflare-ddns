package api

import (
	"testing"

	"gotest.tools/v3/assert"
)

const (
	zone   = "74f896578568875d67af7c4fb1a0442d"
	token  = "RZihSwZ7Vc3uv20CNWquj5AHYdJ5iBMwxrsKok6u"
	domain = "garb.dev"
)

func setup() {
	InitializeAPI(&CloudflareCredentials{
		CloudflareToken: token,
	})
}

func TestListRecords(t *testing.T) {
	setup()
	_, zoneId, err := ListDNSRecordsFiltered(domain, domain)
	assert.NilError(t, err)
	assert.Equal(t, zoneId, zone)
}

func TestUpdateRecord(t *testing.T) {
	setup()
	record, _, _ := ListDNSRecordsFiltered(domain, domain)
	err := UpdateDNSRecord("192.0.2.1", zone, record)
	assert.NilError(t, err)
}
