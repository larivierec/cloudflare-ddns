package ipprovider

type IncrementFunc func(provider string)

func GetProviderName(provider Provider) string {
	return provider.GetProviderName()
}

func GetCurrentIP(provider Provider, incrementFunc IncrementFunc) (string, error) {
	if incrementFunc != nil {
		incrementFunc(provider.GetProviderName())
	}
	return provider.GetCurrentIP()
}
