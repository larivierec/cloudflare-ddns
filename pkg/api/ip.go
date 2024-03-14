package api

type Interface interface {
	GetCurrentIP() (string, error)
	GetProviderName() string
}
