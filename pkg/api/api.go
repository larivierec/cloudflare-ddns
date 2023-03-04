package api

import (
	"github.com/cloudflare/cloudflare-go"
)

type CloudflareCredentials struct {
	ApiKey          string
	AccountEmail    string
	CloudflareToken string
}

func CloudflareApi(creds CloudflareCredentials) (*cloudflare.API, error) {
	if creds.AccountEmail != "" && creds.ApiKey != "" {
		return cloudflare.New(creds.ApiKey, creds.AccountEmail)
	}
	return cloudflare.NewWithAPIToken(creds.CloudflareToken)
}
