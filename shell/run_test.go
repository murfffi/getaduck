package shell

import (
	"flag"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRunArgs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test that downloads from Github in short mode.")
	}
	t.Run("cli", func(t *testing.T) {
		err := RunArgs([]string{"test", "-type", "cli"}, flag.ContinueOnError)
		require.NoError(t, err)
	})
}

func TestParseSpec(t *testing.T) {
	t.Run("version", func(t *testing.T) {
		spec, err := parseSpec([]string{"test", "--version", "1.1.0"}, flag.ContinueOnError)
		require.NoError(t, err)
		require.Equal(t, "1.1.0", spec.Version)
	})
}
