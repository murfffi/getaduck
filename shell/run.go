// Package shell implements the getaduck CLI in a way which is embeddable in other programs
package shell

import (
	"flag"
	"log"
	"path/filepath"

	"github.com/murfffi/getaduck/download"
	"github.com/murfffi/getaduck/internal/enumflag"
)

// RunArgs executes getaduck CLI
func RunArgs(args []string, onError flag.ErrorHandling) error {
	spec, err := parseSpec(args, onError)
	if err != nil {
		return err
	}

	res, err := download.Do(spec)
	if err != nil {
		return err
	}
	outFileName := res.OutputFile
	absPath, err := filepath.Abs(outFileName)
	if err != nil {
		absPath = outFileName
	}
	if res.OutputWritten {
		log.Print("downloaded: ", absPath)
	} else {
		log.Print("already exists: ", absPath)
	}
	return nil
}

func parseSpec(args []string, onError flag.ErrorHandling) (download.Spec, error) {
	fs := flag.NewFlagSet(args[0], onError)
	spec := download.DefaultSpec()

	// order of args must match download.BinType const order
	binType := enumflag.New("lib", "cli")
	fs.Var(binType, "type", binType.Help("type of binary to download"))
	version := fs.String("version", spec.Version, "DuckDB version")
	binOS := fs.String("os", spec.OS, "target OS")
	binArch := fs.String("arch", spec.Arch, "target architecture")
	binOverwrite := fs.Bool("overwrite", true, "overwrite existing file")
	if err := fs.Parse(args[1:]); err != nil {
		return download.Spec{}, err
	}

	spec.Type = download.BinType(binType.Index())
	spec.Version = *version
	spec.OS = *binOS
	spec.Arch = *binArch
	spec.Overwrite = *binOverwrite
	return spec, nil
}
