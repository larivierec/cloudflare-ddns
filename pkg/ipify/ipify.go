package ipify

import (
	"encoding/json"
	"io"
	"net/http"
)

type IpInfo struct {
	Ip string `json:"ip"`
}

const baseUrl = "https://api64.ipify.org?format=json"

func GetCurrentIP() (IpInfo, error) {
	ipInfo := IpInfo{}
	response, err := http.Get(baseUrl)
	if err != nil {
		return ipInfo, err
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return ipInfo, err
	}

	err = json.Unmarshal(body, &ipInfo)
	return ipInfo, err
}
