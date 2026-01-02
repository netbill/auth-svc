package main

import (
	"os"

	"github.com/netbill/auth-svc/cmd/cli"
)

func main() {
	if !cli.Run(os.Args) {
		os.Exit(1)
	}
}
