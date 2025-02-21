package shell

import (
	"github.com/murfffi/getaduck/download"
	"log"
)

func Run() {
	err := download.Do(download.DefaultSpec())
	if err != nil {
		log.Fatalf("download failed: %v", err)
	}
}
