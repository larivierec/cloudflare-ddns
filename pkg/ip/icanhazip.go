package ip

import (
	"io"
	"net/http"
	"strings"
)

type ICanHazIp struct {
	BaseUrl string
}

func (i *ICanHazIp) setup() {
	i.BaseUrl = "https://ipv4.icanhazip.com"
}

func (i *ICanHazIp) GetCurrentIP() (string, error) {
	i.setup()
	response, err := http.Get(i.BaseUrl)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(body)), err
}