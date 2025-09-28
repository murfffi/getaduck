package download

import (
	"fmt"
	"os"
)

const (
	PreviewVersion = "preview"
)

func getPreviewZipUrl(spec Spec) string {
	// https://artifacts.duckdb.org/latest/duckdb-binaries-osx.zip
	// https://artifacts.duckdb.org/latest/duckdb-binaries-windows.zip
	// https://artifacts.duckdb.org/latest/duckdb-binaries-linux-amd64.zip
	archSuffix := ""
	if spec.OS == "linux" {
		archSuffix = "-" + spec.Arch
	}
	return fmt.Sprintf("https://artifacts.duckdb.org/latest/duckdb-binaries-%s%s.zip", spec.OS, archSuffix)
}

func processPreviewZip(spec Spec, entryName string, zipFile string) error {
	innerZip := getInnerZipName(spec)
	err := extractOne(zipFile, innerZip)
	if err != nil {
		return err
	}
	defer func() {
		_ = os.Remove(innerZip)
	}()
	return extractOne(innerZip, entryName)
}

func getInnerZipName(spec Spec) string {
	// duckdb_cli-osx-universal.zip
	// libduckdb-osx-universal.zip
	// duckdb_cli-windows-arm64.zip
	// libduckdb-windows-amd64.zip
	// duckdb_cli-linux-amd64.zip
	// libduckdb-linux-amd64.zip
	prefix := ""
	switch spec.Type {
	case BinTypeCli:
		prefix = "duckdb_cli"
	case BinTypeDynLib:
		prefix = "libduckdb"
	}
	// For osx, spec.Arch has been normalized to universal in normalizeSpec
	return fmt.Sprintf("%s-%s-%s.zip", prefix, spec.OS, spec.Arch)
}
