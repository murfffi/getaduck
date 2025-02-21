package download

import (
	"archive/zip"
	"errors"
	"fmt"
	"github.com/murfffi/getaduck/internal/sclerr"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/ansel1/merry/v2"
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

func getPath(spec Spec) string {
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

type Spec struct {
	Type    BinType
	Version string
	OS      string
	Arch    string
}

func Do(spec Spec) error {
	var err error
	if spec.Version == LatestVersion {
		spec.Version, err = getLatestVersionPath()
		if err != nil {
			return err
		}
	}
	path := getPath(spec)
	tmpFile, err := fetchZip(path)
	if err != nil {
		return err
	}
	defer func() {
		_ = os.Remove(tmpFile)
	}()
	return extractOne(tmpFile, getEntryName(spec))
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

func DefaultSpec() Spec {
	return Spec{
		Type:    BinTypeDynLib,
		Version: LatestVersion,
		OS:      runtime.GOOS,
		Arch:    runtime.GOARCH,
	}
}

func extractOne(zipFile string, name string) error {
	zipReader, err := zip.OpenReader(zipFile)
	if err != nil {
		return merry.Wrap(fmt.Errorf("could not open zip %s: %w", zipFile, err))
	}
	defer func() {
		err = zipReader.Close()
	}()
	for _, file := range zipReader.File {
		if file.Name == name {
			err = extractFile(file)
		}
	}
	return err
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

func fetchZip(path string) (string, error) {
	resp, err := http.Get(path)
	if err != nil {
		return "", fmt.Errorf("failed to download %s: %w", path, err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download %s: %d", path, resp.StatusCode)
	}
	defer sclerr.CloseQuietly(resp.Body)
	tmpZip, err := os.CreateTemp("", "getaduck")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer sclerr.CloseQuietly(tmpZip)
	_, err = io.Copy(tmpZip, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to download %s: %w", path, err)
	}
	err = tmpZip.Close()
	if err != nil {
		return "", fmt.Errorf("failed to close temp file: %w", err)
	}

	return tmpZip.Name(), nil
}
