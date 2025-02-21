package shell

import (
	"log"
	"path/filepath"

	"github.com/murfffi/getaduck/download"
)

func Run() {
	outFileName, err := download.Do(download.DefaultSpec())
	if err != nil {
		log.Fatalf("download failed: %v", err)
	}
	absPath, err := filepath.Abs(outFileName)
	if err != nil {
		absPath = outFileName
	}
	log.Print("downloaded: ", absPath)
}
