package provider

import (
	"github.com/larivierec/cloudflare-ddns/pkg/api"
)

func GetCurrentIP(provider api.Interface) (string, error) {
	return provider.GetCurrentIP()
}
