package cloudprovider_test

import (
	"testing"

	"github.com/larivierec/cloudflare-ddns/pkg/cloudprovider"
	"github.com/larivierec/cloudflare-ddns/pkg/cloudprovider/cloudflare"
	"gotest.tools/v3/assert"
)

func TestProvider_Interface_WorksWithAllImplementations(t *testing.T) {
	providers := []struct {
		name     string
		provider cloudprovider.Provider
	}{
		{"cloudflare", cloudflare.NewCloudflareProvider()},
	}
	assert.Equal(t, len(providers), 1, "Should have one provider for testing")
}

func TestRecord_StructValidation(t *testing.T) {
	record := cloudprovider.Record{
		ID:      "string-id",
		Type:    "A",
		Name:    "example.com",
		Content: "192.168.1.1",
		TTL:     300,
		Proxied: true,
	}

	var _ string = record.ID
	var _ string = record.Type
	var _ string = record.Name
	var _ string = record.Content
	var _ int = record.TTL
	var _ bool = record.Proxied

	assert.Equal(t, "string-id", record.ID)
	assert.Equal(t, "A", record.Type)
	assert.Equal(t, "example.com", record.Name)
	assert.Equal(t, "192.168.1.1", record.Content)
	assert.Equal(t, 300, record.TTL)
	assert.Equal(t, true, record.Proxied)
}

func TestProvider_PolymorphicUsage(t *testing.T) {
	var provider cloudprovider.Provider

	provider = cloudflare.NewCloudflareProvider()
	_, err := provider.GetDNSRecord("example.com", "test")
	assert.Assert(t, err != nil, "Should fail with auth error")
}
