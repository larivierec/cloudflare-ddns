package provider

import (
	"github.com/larivierec/cloudflare-ddns/pkg/api"
)

func GetProviderName(provider api.Interface) string {
	return provider.GetProviderName()
}

func GetCurrentIP(provider api.Interface) (string, error) {
	return provider.GetCurrentIP()
}
