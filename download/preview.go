package download

import (
	"fmt"
	"os"

	"github.com/ansel1/merry/v2"
)

// Downloading preview releases

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
		return merry.Wrap(fmt.Errorf("failed to extract inner zip '%s' from '%s': %w", innerZip, zipFile, err))
	}
	defer func() {
		_ = os.Remove(innerZip)
	}()
	err = extractOne(innerZip, entryName)
	if err != nil {
		return merry.Wrap(fmt.Errorf("failed to extract entry '%s' from inner zip '%s': %w", entryName, innerZip, err))
	}
	return nil
}

func getInnerZipName(spec Spec) string {
	// duckdb_cli-osx-universal.zip
	// libduckdb-osx-universal.zip
	// duckdb_cli-windows-arm64.zip
	// libduckdb-windows-amd64.zip
	// duckdb_cli-linux-amd64.zip
	// libduckdb-linux-amd64.zip
	// For osx, spec.Arch has been normalized to universal in normalizeSpec
	return fmt.Sprintf("%s-%s-%s.zip", spec.Type.Prefix(), spec.OS, spec.Arch)
}
