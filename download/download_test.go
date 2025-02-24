package download_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/murfffi/getaduck/download"
)

func TestDo(t *testing.T) {
	outFileName, err := download.Do(download.DefaultSpec())
	require.NoError(t, err)
	require.FileExists(t, outFileName)

}
