package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	ddns "github.com/larivierec/cloudflare-ddns/pkg/cmd"
)

const banner = `
cloudflare-ddns
version: %s (%s)

`

var (
	Version = "local"
	Gitsha  = "?"
)

func main() {
	fmt.Printf(banner, Version, Gitsha)
	mode := os.Getenv("MODE")
	if mode == "serverless" {
		log.Println("Running in serverless mode")
		http.HandleFunc("/", ddns.StartServerless)
		log.Fatal(http.ListenAndServe(":9000", nil))
	} else {
		ddns.Start(mode)
	}
}
