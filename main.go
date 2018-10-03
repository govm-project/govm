package main

import (
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/govm-project/govm/cli"
)

func main() {
	vm := cli.Init()
	err := vm.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
