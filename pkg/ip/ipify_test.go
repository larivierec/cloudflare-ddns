package ip

import (
	"testing"

	"github.com/larivierec/cloudflare-ddns/pkg/provider"
	"gotest.tools/v3/assert"
)

func TestIpify(t *testing.T) {
	_, err := provider.GetCurrentIP(&Ipify{})
	assert.NilError(t, err)
}
