package ip

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/larivierec/cloudflare-ddns/pkg/metrics"
)

const ipifyString = "ipify"

type Ipify struct {
	BaseUrl string
}

type IpInfo struct {
	Ip string `json:"ip"`
}

func (i *Ipify) setup() {
	i.BaseUrl = "https://api64.ipify.org?format=json"
}

func (i *Ipify) GetProviderName() string {
	return ipifyString
}

func (i *Ipify) GetCurrentIP() (string, error) {
	i.setup()
	ipInfo := IpInfo{}
	response, err := http.Get(i.BaseUrl)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	metrics.IncrementProvider(i)
	err = json.Unmarshal(body, &ipInfo)
	return ipInfo.Ip, err
}
