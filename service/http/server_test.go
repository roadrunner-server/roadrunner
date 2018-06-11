package http

import (
	"testing"
	"github.com/spiral/roadrunner"
	"os"
	"github.com/stretchr/testify/assert"
	"net/http"
	"context"
	"io/ioutil"
)

// get request and return body
func get(url string) (string, *http.Response, error) {
	r, err := http.Get(url)
	if err != nil {
		return "", nil, err
	}
	defer r.Body.Close()

	b, err := ioutil.ReadAll(r.Body)
	return string(b), r, err
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

func TestServer_Headers(t *testing.T) {
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

	req, err := http.NewRequest("GET", "http://localhost:8077?hello=world", nil)
	assert.NoError(t, err)

	req.Header.Add("input", "sample")

	r, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer r.Body.Close()

	b, err := ioutil.ReadAll(r.Body)
	assert.NoError(t, err)

	assert.NoError(t, err)
	assert.Equal(t, 200, r.StatusCode)
	assert.Equal(t, "world", r.Header.Get("Header"))
	assert.Equal(t, "SAMPLE", string(b))
}

func TestServer_Cookies(t *testing.T) {
	st := &Server{
		cfg: &Config{
			MaxRequest: 1024,
			Uploads: &UploadsConfig{
				Dir:    os.TempDir(),
				Forbid: []string{},
			},
		},
		rr: roadrunner.NewServer(&roadrunner.ServerConfig{
			Command: "php ../../php-src/tests/http/client.php cookie pipes",
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

	req, err := http.NewRequest("GET", "http://localhost:8077", nil)
	assert.NoError(t, err)

	req.AddCookie(&http.Cookie{Name: "input", Value: "input-value"})

	r, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer r.Body.Close()

	b, err := ioutil.ReadAll(r.Body)
	assert.NoError(t, err)

	assert.NoError(t, err)
	assert.Equal(t, 200, r.StatusCode)
	assert.Equal(t, "INPUT-VALUE", string(b))

	for _, c := range r.Cookies() {
		assert.Equal(t, "output", c.Name)
		assert.Equal(t, "cookie-output", c.Value)
	}
}
