// Package shell implements the getaduck CLI in a way which is embeddable in other programs
package shell

import (
	"log"
	"path/filepath"

	"github.com/murfffi/getaduck/download"
)

// Run executes getaduck
func Run() {
	spec := download.DefaultSpec()
	res, err := download.Do(spec)
	if err != nil {
		log.Fatalf("download failed: %v", err)
	}
	outFileName := res.OutputFile
	absPath, err := filepath.Abs(outFileName)
	if err != nil {
		absPath = outFileName
	}
	log.Print("downloaded: ", absPath)
}
