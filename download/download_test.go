package download_test

import (
	"testing"

	"github.com/murfffi/getaduck/download"
	"github.com/stretchr/testify/require"
)

func TestDo(t *testing.T) {
	if !testing.Short() {
		t.Skip("skipping test that downloads from Github in short mode.")
	}
	t.Run("default lib", func(t *testing.T) {
		res, err := download.Do(download.DefaultSpec())
		require.NoError(t, err)
		require.FileExists(t, res.OutputFile)
	})
	// cli is tested e2e in shell/run_test.go . Avoid multiple downloads.
}
