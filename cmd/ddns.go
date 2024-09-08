package main

import (
	"log"
	"net/http"
	"os"

	ddns "github.com/larivierec/cloudflare-ddns/pkg/cmd"
)

func main() {
	mode := os.Getenv("MODE")
	if mode == "serverless" {
		log.Println("Running in serverless mode")
		http.HandleFunc("/", ddns.StartServerless)
		log.Fatal(http.ListenAndServe(":9000", nil))
	} else {
		ddns.Start()
	}
}
