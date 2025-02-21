package shell

import (
	"log"
	"runtime"

	"github.com/murfffi/getaduck/internal/download"
)

func Run() {
	err := download.Do(download.Spec{
		Type:    download.BinTypeDynLib,
		Version: "latest",
		OS:      runtime.GOOS,
		Arch:    runtime.GOARCH,
	})
	if err != nil {
		log.Fatalf("download failed: %v", err)
	}
}
