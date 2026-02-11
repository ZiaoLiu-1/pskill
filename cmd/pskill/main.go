package main

import (
	"log"

	"github.com/ZiaoLiu-1/pskill/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		log.Fatal(err)
	}
}
