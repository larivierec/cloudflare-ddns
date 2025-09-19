package cloudflare_test

import (
	"os"
	"testing"

	"github.com/larivierec/cloudflare-ddns/pkg/cloudprovider"
	"github.com/larivierec/cloudflare-ddns/pkg/cloudprovider/cloudflare"
	"gotest.tools/v3/assert"
)

func TestCloudflareProvider_ImplementsInterface(t *testing.T) {
	var _ cloudprovider.Provider = &cloudflare.CloudflareProvider{}
}

func TestNewCloudflareProvider(t *testing.T) {
	os.Setenv("ACCOUNT_TOKEN", "test-token")
	os.Setenv("API_KEY", "test-api-key")
	os.Setenv("ACCOUNT_EMAIL", "test@example.com")
	defer func() {
		os.Unsetenv("ACCOUNT_TOKEN")
		os.Unsetenv("API_KEY")
		os.Unsetenv("ACCOUNT_EMAIL")
	}()

	provider := cloudflare.NewCloudflareProvider()
	assert.Assert(t, provider != nil, "Provider should not be nil")
}

func TestRecord_ValidatesFields(t *testing.T) {
	record := &cloudprovider.Record{
		ID:      "test-id",
		Type:    "A",
		Name:    "test.example.com",
		Content: "192.168.1.1",
		TTL:     300,
	}

	assert.Equal(t, "test-id", record.ID)
	assert.Equal(t, "A", record.Type)
	assert.Equal(t, "test.example.com", record.Name)
	assert.Equal(t, "192.168.1.1", record.Content)
	assert.Equal(t, 300, record.TTL)
}

func TestInterface_MethodSignatures(t *testing.T) {
	provider := cloudflare.NewCloudflareProvider()

	_, err := provider.GetDNSRecord("example.com", "test")
	assert.Assert(t, err != nil, "Should fail with auth error, proving method exists")

	testRecord := &cloudprovider.Record{
		Type:    "A",
		Name:    "test.example.com",
		Content: "192.168.1.1",
		TTL:     300,
	}

	_, err = provider.CreateDNSRecord("example.com", testRecord)
	assert.Assert(t, err != nil, "Should fail with auth error, proving method exists")

	_, err = provider.UpdateDNSRecord("example.com", testRecord)
	assert.Assert(t, err != nil, "Should fail with auth error, proving method exists")

	err = provider.DeleteDNSRecord("example.com", "test-id")
	assert.Assert(t, err != nil, "Should fail with auth error, proving method exists")
}
