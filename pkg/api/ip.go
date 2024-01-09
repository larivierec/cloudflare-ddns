package api

type Interface interface {
	GetCurrentIP() (string, error)
}
