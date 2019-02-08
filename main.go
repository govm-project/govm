package main

import (
	"os"

	"github.com/govm-project/govm/cmd"

	log "github.com/sirupsen/logrus"
)

func main() {
	vm := cmd.Init()
	err := vm.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
