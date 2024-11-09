package ipprovider_test

import (
	"testing"

	"github.com/larivierec/cloudflare-ddns/pkg/ipprovider"
	"github.com/larivierec/cloudflare-ddns/pkg/metrics"
	"gotest.tools/v3/assert"
)

func TestICanHaz(t *testing.T) {
	_, err := ipprovider.GetCurrentIP(&ipprovider.ICanHazIp{}, metrics.IncrementProvider)
	assert.NilError(t, err)
}
