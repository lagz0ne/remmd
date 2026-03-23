package main

import (
	"os"

	"github.com/lagz0ne/remmd/internal/cli"
)

func main() {
	if err := cli.NewRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
