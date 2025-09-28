package download_test

import (
	"testing"

	"github.com/murfffi/getaduck/download"
	"github.com/stretchr/testify/require"
)

func TestDo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test that downloads from Github in short mode.")
	}
	t.Run("default lib", func(t *testing.T) {
		for _, version := range []string{
			"1.2.2",
			"v1.3.2",
			"v1.4.0",
			"latest",
			"preview",
		} {
			t.Run(version, func(t *testing.T) {
				for _, arch := range []string{
					"amd64",
					"arm64",
				} {
					t.Run(arch, func(t *testing.T) {
						spec := download.DefaultSpec()
						spec.Version = version
						spec.Arch = arch
						spec.Overwrite = true
						res, err := download.Do(spec)
						require.NoError(t, err)
						require.FileExists(t, res.OutputFile)
					})

				}
			})
		}
	})
	// cli is tested e2e in shell/run_test.go . Avoid multiple downloads.
}
