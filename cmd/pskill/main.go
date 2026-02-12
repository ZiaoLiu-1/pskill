package main

import (
	"log"

	"github.com/ZiaoLiu-1/pskill/internal/cli"
)

// Set via ldflags at build time.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cli.SetVersion(version, commit, date)
	if err := cli.Execute(); err != nil {
		log.Fatal(err)
	}
}
