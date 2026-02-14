package main

import (
	"os"

	"github.com/netbill/auth-svc/internal/cli"
)

func main() {
	if !cli.Run(os.Args) {
		os.Exit(1)
	}
}
