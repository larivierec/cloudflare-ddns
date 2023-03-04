package ipify

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestIpify(t *testing.T) {
	_, err := GetCurrentIP()
	assert.NilError(t, err)
}
