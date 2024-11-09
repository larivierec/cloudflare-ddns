package ipprovider

type Provider interface {
	GetCurrentIP() (string, error)
	GetProviderName() string
}
