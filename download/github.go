package download

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// Code to download from Github - applies to semver and latest releases

const (
	duckDbReleasesRoot = "https://github.com/duckdb/duckdb/releases"
)

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

func getGithubURL(spec Spec) string {
	return fmt.Sprintf("%s/download/%s/%s-%s-%s.zip", duckDbReleasesRoot, spec.Version, spec.Type.Prefix(), spec.OS, spec.Arch)
}
