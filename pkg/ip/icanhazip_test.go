package ip

import (
	"testing"

	"github.com/larivierec/cloudflare-ddns/pkg/provider"
	"gotest.tools/v3/assert"
)

func TestICanHaz(t *testing.T) {
	_, err := provider.GetCurrentIP(&ICanHazIp{})
	assert.NilError(t, err)
}
