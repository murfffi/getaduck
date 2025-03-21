package main

import (
	"flag"
	"log"
	"os"

	"github.com/murfffi/getaduck/shell"
)

func main() {
	err := shell.RunArgs(os.Args, flag.ExitOnError)
	if err != nil {
		log.Fatal(err)
	}
}
