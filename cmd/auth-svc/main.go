package main

import (
	"os"

	"github.com/netbill/auth-svc/internal/build/cli"
)

func main() {
	cli.Run(os.Args)
}
