package main

import (
	"fmt"
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
	ddns.Start(mode)
}
