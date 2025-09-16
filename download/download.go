// Package download implements downloading DuckDB releases as a library
package download

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/ansel1/merry/v2"
	"golang.org/x/mod/semver"

	"github.com/murfffi/getaduck/internal/sclerr"
)

const (
	LatestVersion      = "latest"
	duckDbReleasesRoot = "https://github.com/duckdb/duckdb/releases"
)

type BinType int

const (
	BinTypeDynLib = BinType(iota)
	BinTypeCli
)

type Spec struct {
	// Type of binary to download (enum)
	Type BinType

	// DuckDB version, defaults to latest
	Version string

	// Target OS defaults to runtime.GOOS
	OS string

	// Target arch defaults to runtime.GOARCH
	Arch string

	// Overwrite forces downloading a file even if there is an existing appropriate in the working directory
	// The definition of "appropriate" will evolve over time - for now, all existing files are accepted
	Overwrite bool
}

func DefaultSpec() Spec {
	return Spec{
		Type:    BinTypeDynLib,
		Version: LatestVersion,
		OS:      runtime.GOOS,
		Arch:    runtime.GOARCH,
	}
}

type Result struct {
	OutputFile string
	// Download may be false if there was an existing appropriate file and Spec.Overwrite was false
	// See Spec.Overwrite for details.
	Downloaded bool
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
	res.Downloaded = true
	path := getGithubURL(spec)
	tmpFile, err := fetchZip(path)
	if err != nil {
		return res, err
	}
	defer func() {
		_ = os.Remove(tmpFile)
	}()
	return res, extractOne(tmpFile, entryName)
}

func existsAppropriate(fileName string) bool {
	fi, err := os.Stat(fileName)
	// the details of the error are not valuable in this context
	// we will try to download and write the file if this fails
	return err == nil && fi.Mode().IsRegular()
}

func getGithubURL(spec Spec) string {
	var archivePrefix string
	switch spec.Type {
	case BinTypeCli:
		archivePrefix = "duckdb_cli"
	case BinTypeDynLib:
		archivePrefix = "libduckdb"
	default:
		panic("unhandled spec type")
	}
	return fmt.Sprintf("%s/download/%s/%s-%s-%s.zip", duckDbReleasesRoot, spec.Version, archivePrefix, spec.OS, spec.Arch)
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

func getLatestVersionPath() (string, error) {
	redirectErr := errors.New("redirect")
	client := http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return redirectErr
		},
	}
	const latestUrl = duckDbReleasesRoot + "/latest"
	resp, err := client.Head(latestUrl)
	if errors.Is(err, redirectErr) {
		location := resp.Header.Get("Location")
		prefix := duckDbReleasesRoot + "/tag/"
		if !strings.HasPrefix(location, prefix) {
			return "", fmt.Errorf("unexpected release redirect location: %s", location)
		}
		return location[len(prefix):], nil
	}
	if err != nil {
		return "", fmt.Errorf("HEAD failed for %s: %w", latestUrl, err)
	}
	_ = resp.Body.Close()
	return "", fmt.Errorf("redirect expected for %s but got code %d", latestUrl, resp.StatusCode)
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
	defer sclerr.CloseQuietly(outFile)
	fileReader, err := file.Open()
	if err != nil {
		return err
	}
	defer sclerr.CloseQuietly(fileReader)
	_, err = io.Copy(outFile, fileReader)
	return err
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

func fetchZip(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", genericDownloadErr(url, err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP error when trying to download %s: %d", url, resp.StatusCode)
	}
	defer sclerr.CloseQuietly(resp.Body)
	tmpZip, err := os.CreateTemp("", "getaduck")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer sclerr.CloseQuietly(tmpZip)
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
