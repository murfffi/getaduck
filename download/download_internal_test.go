package download

import (
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFetchZipWithClient(t *testing.T) {
	t.Run("empty body", func(t *testing.T) {
		_, err := fetchZipWithClient("", true, clientStub{
			response: &http.Response{
				StatusCode: 200,
			},
		})
		require.Error(t, err)
	})
	t.Run("empty etag", func(t *testing.T) {
		name, err := fetchZipWithClient("", true, clientStub{
			response: &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader("")),
			},
		})
		require.NoError(t, err)
		require.NoError(t, os.Remove(name))
		require.NotContains(t, name, "etag")
	})
}

type clientStub struct {
	response *http.Response
	err      error
}

func (c clientStub) Get(_ string) (resp *http.Response, err error) {
	return c.response, c.err
}
