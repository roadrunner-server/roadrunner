package http

import (
	"testing"
	"github.com/spiral/roadrunner"
	"os"
	"github.com/stretchr/testify/assert"
	"net/http"
	"context"
	"bytes"
	"io"
)

// get request and return body
func get(url string) (string, *http.Response, error) {
	r, err := http.Get(url)
	if err != nil {
		return "", nil, err
	}
	defer r.Body.Close()

	buf := new(bytes.Buffer)
	io.Copy(buf, r.Body)

	return buf.String(), r, nil
}

func TestServer_Echo(t *testing.T) {
	st := &Server{
		cfg: &Config{
			MaxRequest: 1024,
			Uploads: &UploadsConfig{
				Dir:    os.TempDir(),
				Forbid: []string{},
			},
		},
		rr: roadrunner.NewServer(&roadrunner.ServerConfig{
			Command: "php ../../php-src/tests/http/client.php echo pipes",
			Relay:   "pipes",
			Pool: &roadrunner.Config{
				NumWorkers:      1,
				AllocateTimeout: 10000000,
				DestroyTimeout:  10000000,
			},
		}),
	}

	assert.NoError(t, st.rr.Start())
	defer st.rr.Stop()

	hs := &http.Server{Addr: ":8077", Handler: st,}
	defer hs.Shutdown(context.Background())

	go func() { hs.ListenAndServe() }()

	body, r, err := get("http://localhost:8077/?hello=world")
	assert.NoError(t, err)
	assert.Equal(t, 200, r.StatusCode)
	assert.Equal(t, "WORLD", body)
}

func TestServer_EchoHeaders(t *testing.T) {
	st := &Server{
		cfg: &Config{
			MaxRequest: 1024,
			Uploads: &UploadsConfig{
				Dir:    os.TempDir(),
				Forbid: []string{},
			},
		},
		rr: roadrunner.NewServer(&roadrunner.ServerConfig{
			Command: "php ../../php-src/tests/http/client.php header pipes",
			Relay:   "pipes",
			Pool: &roadrunner.Config{
				NumWorkers:      1,
				AllocateTimeout: 10000000,
				DestroyTimeout:  10000000,
			},
		}),
	}

	assert.NoError(t, st.rr.Start())
	defer st.rr.Stop()

	hs := &http.Server{Addr: ":8077", Handler: st,}
	defer hs.Shutdown(context.Background())

	go func() { hs.ListenAndServe() }()

	_, r, err := get("http://localhost:8077/?hello=world")
	assert.NoError(t, err)
	assert.Equal(t, 200, r.StatusCode)
	assert.Equal(t, "world", r.Header.Get("Header"))
}
