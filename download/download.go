// Package download implements downloading DuckDB releases as a library
package download

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/ansel1/merry/v2"
	"golang.org/x/mod/semver"

	"github.com/murfffi/gorich/helperr"
)

const (
	LatestVersion  = "latest"
	PreviewVersion = "preview"
)

type BinType int

const (
	BinTypeDynLib = BinType(iota)
	BinTypeCli
)

// Prefix is found in the beginning of some archive and file names in DuckDB packages
func (typ BinType) Prefix() string {
	var prefix string
	switch typ {
	case BinTypeCli:
		prefix = "duckdb_cli"
	case BinTypeDynLib:
		prefix = "libduckdb"
	default:
		panic("unhandled spec type")
	}
	return prefix
}

// Spec defines the desired DuckDB binary and download options
// Use DefaultSpec() to get a recommended configuration. The zero value is also valid.
type Spec struct {
	// Type of binary to download (enum)
	Type BinType

	// DuckDB version, defaults to latest
	// Supported values are either plain semantic version with optional 'v' prefix - e.g. 1.2.2, v1.3.2,
	// or "latest" - latest release version
	// or "preview" - latest preview version from https://duckdb.org/docs/installation/?version=main
	Version string

	// Target OS, defaults to runtime.GOOS
	OS string

	// Target arch defaults, to runtime.GOARCH
	Arch string

	// CacheDownload enables caching the bundle downloaded from the Internet in the temp directory,
	// if the server supports it by exposing Etag and Content-Length headers.
	// CacheDownload is independent of the Overwrite setting.
	CacheDownload bool

	// Overwrite forces overwriting the final file even if there is an existing appropriate in the working directory
	// The definition of "appropriate" will evolve over time - for now, all existing files are accepted.
	Overwrite bool
}

// DefaultSpec creates a recommended spec for downloading releases
// The zero-value of Spec is also a valid configuration.
// NB: Changes to the default spec are not considered breaking changes and may happen in a
// minor release. They won't happen in patch releases.
func DefaultSpec() Spec {
	return Spec{
		Type:          BinTypeDynLib,
		Version:       LatestVersion,
		CacheDownload: true,
		OS:            runtime.GOOS,
		Arch:          runtime.GOARCH,
	}
}

type Result struct {
	OutputFile string
	// OutputWritten may be false if there was an existing appropriate file and Spec.Overwrite was false
	// See Spec.Overwrite for details.
	OutputWritten bool
}

// Do downloads a DuckDB release
func Do(spec Spec) (Result, error) {
	res := Result{}
	spec, err := normalizeSpec(spec)
	if err != nil {
		return res, err
	}
	entryName := getEntryName(spec)
	res.OutputFile = entryName
	if !spec.Overwrite && existsAppropriate(entryName) {
		return res, nil
	}
	res.OutputWritten = true
	path := getZipDownloadUrl(spec)
	tmpFile, err := fetchZip(path, spec.CacheDownload)
	if err != nil {
		return res, err
	}
	if !spec.CacheDownload {
		defer func() {
			_ = os.Remove(tmpFile)
		}()
	}
	return res, processZip(spec, entryName, tmpFile)
}

func processZip(spec Spec, entryName string, zipFile string) error {
	if spec.Version != PreviewVersion {
		return extractOne(zipFile, entryName)
	}
	return processPreviewZip(spec, entryName, zipFile)
}

func getZipDownloadUrl(spec Spec) string {
	if spec.Version == PreviewVersion {
		return getPreviewZipUrl(spec)
	}
	return getGithubURL(spec)
}

func existsAppropriate(fileName string) bool {
	fi, err := os.Stat(fileName)
	// the details of the error are not valuable in this context
	// we will try to download and write the file if this fails
	return err == nil && fi.Mode().IsRegular()
}

func normalizeSpec(spec Spec) (Spec, error) {
	spec.Arch = strings.ToLower(spec.Arch)
	spec.OS = strings.ToLower(spec.OS)
	spec.Version = strings.ToLower(spec.Version)

	var err error
	if spec.Version == LatestVersion {
		spec.Version, err = getLatestVersionPath()
		if err != nil {
			return spec, err
		}
	}

	if spec.OS == "darwin" {
		spec.OS = "osx"
	}

	if spec.OS == "osx" {
		spec.Arch = "universal"
	}

	if !semver.IsValid(spec.Version) && semver.IsValid("v"+spec.Version) {
		spec.Version = "v" + spec.Version
	}

	if spec.Arch == "arm64" && semver.IsValid(spec.Version) && semver.Compare(spec.Version, "v1.3.0") < 0 {
		spec.Arch = "aarch64"
	}
	return spec, err
}

func extractOne(zipFile string, name string) error {
	zipReader, err := zip.OpenReader(zipFile)
	if err != nil {
		return merry.Wrap(fmt.Errorf("could not open zip %s: %w", zipFile, err))
	}
	defer func() {
		err = zipReader.Close()
	}()
	err = fmt.Errorf("did not find expected file %s in %+v", name, getNames(zipReader.File))
	for _, file := range zipReader.File {
		if file.Name == name {
			err = extractFile(file)
		}
	}
	return err
}

func getNames(files []*zip.File) []string {
	names := make([]string, len(files))
	for i, f := range files {
		names[i] = f.Name
	}
	return names
}

func extractFile(file *zip.File) error {
	outFile, err := os.OpenFile(file.Name, os.O_RDWR|os.O_CREATE|os.O_TRUNC, file.Mode())
	if err != nil {
		return err
	}
	defer helperr.CloseQuietly(outFile)
	fileReader, err := file.Open()
	if err != nil {
		return err
	}
	defer helperr.CloseQuietly(fileReader)
	_, err = io.Copy(outFile, fileReader)
	if err != nil {
		return err
	}
	return outFile.Close()
}

func getEntryName(spec Spec) string {
	switch spec.Type {
	case BinTypeDynLib:
		return getDynLibName(spec.OS)
	case BinTypeCli:
		return getCliName(spec.OS)
	default:
		panic("unhandled spec type")
	}
}

func getDynLibName(targetOS string) string {
	switch targetOS {
	case "windows":
		return "duckdb.dll"
	case "osx": // uses name from normalizeSpec
		return "libduckdb.dylib"
	default:
		return "libduckdb.so"
	}
}

func getCliName(targetOS string) string {
	name := "duckdb"
	if targetOS == "windows" {
		name += ".exe"
	}
	return name
}

func fetchZip(url string, useEtag bool) (string, error) {
	// It *may* be more efficient (for whom?) to issue a HEAD request first for the ETag and Content-Length.
	// We can't use If-None-Match because we don't know in advance which cached file is for which spec.
	// We could encode the entire spec in the cached file name but the complexity would not be worth it.
	resp, err := http.Get(url)
	if err != nil {
		return "", genericDownloadErr(url, err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP error when trying to download %s: %d", url, resp.StatusCode)
	}
	etagHeader := resp.Header.Get("ETag")
	contentLength := resp.ContentLength
	defer helperr.CloseQuietly(resp.Body)
	var tmpZip *os.File
	if !useEtag && etagHeader != "" {
		tmpZip, err = os.CreateTemp("", "getaduck")
	} else {
		fileName := fmt.Sprintf("getaduck.zip.etag_%s", etagHeader)
		fileName = filepath.Join(os.TempDir(), fileName)
		if info, statErr := os.Stat(fileName); statErr == nil {
			if info.Size() == contentLength {
				return fileName, nil
			}
		}

		tmpZip, err = os.Create(fileName)
	}
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer helperr.CloseQuietly(tmpZip)
	_, err = io.Copy(tmpZip, resp.Body)
	if err != nil {
		return "", genericDownloadErr(url, err)
	}
	err = tmpZip.Close()
	if err != nil {
		return "", fmt.Errorf("failed to close temp file: %w", err)
	}

	return tmpZip.Name(), nil
}

func genericDownloadErr(url string, err error) error {
	return fmt.Errorf("failed to download %s: %w", url, err)
}
