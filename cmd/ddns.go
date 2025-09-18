package main

import (
	"fmt"

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
	ddns.Start()
}
