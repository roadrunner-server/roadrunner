package http

import (
	"bytes"
	"context"
	"github.com/spiral/roadrunner"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"
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
	st := &Handler{
		cfg: &Config{
			MaxRequest: 1024,
			Uploads: &UploadsConfig{
				Dir:    os.TempDir(),
				Forbid: []string{},
			},
		},
		rr: roadrunner.NewServer(&roadrunner.ServerConfig{
			Command: "php ../../tests/http/client.php echo pipes",
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

	hs := &http.Server{Addr: ":8177", Handler: st}
	defer hs.Shutdown(context.Background())

	go func() { hs.ListenAndServe() }()
	time.Sleep(time.Millisecond * 10)

	body, r, err := get("http://localhost:8177/?hello=world")
	assert.NoError(t, err)
	assert.Equal(t, 201, r.StatusCode)
	assert.Equal(t, "WORLD", body)
}

func Test_HandlerErrors(t *testing.T) {
	st := &Handler{
		cfg: &Config{
			MaxRequest: 1024,
			Uploads: &UploadsConfig{
				Dir:    os.TempDir(),
				Forbid: []string{},
			},
		},
		rr: roadrunner.NewServer(&roadrunner.ServerConfig{
			Command: "php ../../tests/http/client.php echo pipes",
			Relay:   "pipes",
			Pool: &roadrunner.Config{
				NumWorkers:      1,
				AllocateTimeout: 10000000,
				DestroyTimeout:  10000000,
			},
		}),
	}

	wr := httptest.NewRecorder()
	rq := httptest.NewRequest("POST", "/", bytes.NewBuffer([]byte("data")))

	st.ServeHTTP(wr, rq)
	assert.Equal(t, 500, wr.Code)
}

func Test_Handler_JSON_error(t *testing.T) {
	st := &Handler{
		cfg: &Config{
			MaxRequest: 1024,
			Uploads: &UploadsConfig{
				Dir:    os.TempDir(),
				Forbid: []string{},
			},
		},
		rr: roadrunner.NewServer(&roadrunner.ServerConfig{
			Command: "php ../../tests/http/client.php echo pipes",
			Relay:   "pipes",
			Pool: &roadrunner.Config{
				NumWorkers:      1,
				AllocateTimeout: 10000000,
				DestroyTimeout:  10000000,
			},
		}),
	}

	wr := httptest.NewRecorder()
	rq := httptest.NewRequest("POST", "/", bytes.NewBuffer([]byte("{sd")))
	rq.Header.Add("Content-Type", "application/json")
	rq.Header.Add("Content-Size", "3")

	st.ServeHTTP(wr, rq)
	assert.Equal(t, 500, wr.Code)
}

func TestServer_Headers(t *testing.T) {
	st := &Handler{
		cfg: &Config{
			MaxRequest: 1024,
			Uploads: &UploadsConfig{
				Dir:    os.TempDir(),
				Forbid: []string{},
			},
		},
		rr: roadrunner.NewServer(&roadrunner.ServerConfig{
			Command: "php ../../tests/http/client.php header pipes",
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

	hs := &http.Server{Addr: ":8078", Handler: st}
	defer hs.Shutdown(context.Background())

	go func() { hs.ListenAndServe() }()
	time.Sleep(time.Millisecond * 10)

	req, err := http.NewRequest("GET", "http://localhost:8078?hello=world", nil)
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
	st := &Handler{
		cfg: &Config{
			MaxRequest: 1024,
			Uploads: &UploadsConfig{
				Dir:    os.TempDir(),
				Forbid: []string{},
			},
		},
		rr: roadrunner.NewServer(&roadrunner.ServerConfig{
			Command: "php ../../tests/http/client.php cookie pipes",
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

	hs := &http.Server{Addr: ":8079", Handler: st}
	defer hs.Shutdown(context.Background())

	go func() { hs.ListenAndServe() }()
	time.Sleep(time.Millisecond * 10)

	req, err := http.NewRequest("GET", "http://localhost:8079", nil)
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

func TestServer_JsonPayload_POST(t *testing.T) {
	st := &Handler{
		cfg: &Config{
			MaxRequest: 1024,
			Uploads: &UploadsConfig{
				Dir:    os.TempDir(),
				Forbid: []string{},
			},
		},
		rr: roadrunner.NewServer(&roadrunner.ServerConfig{
			Command: "php ../../tests/http/client.php payload pipes",
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

	hs := &http.Server{Addr: ":8090", Handler: st}
	defer hs.Shutdown(context.Background())

	go func() { hs.ListenAndServe() }()
	time.Sleep(time.Millisecond * 10)

	req, err := http.NewRequest(
		"POST",
		"http://localhost"+hs.Addr,
		bytes.NewBufferString(`{"key":"value"}`),
	)
	assert.NoError(t, err)

	req.Header.Add("Content-Type", "application/json")

	r, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer r.Body.Close()

	b, err := ioutil.ReadAll(r.Body)
	assert.NoError(t, err)

	assert.NoError(t, err)
	assert.Equal(t, 200, r.StatusCode)
	assert.Equal(t, `{"value":"key"}`, string(b))
}

func TestServer_JsonPayload_PUT(t *testing.T) {
	st := &Handler{
		cfg: &Config{
			MaxRequest: 1024,
			Uploads: &UploadsConfig{
				Dir:    os.TempDir(),
				Forbid: []string{},
			},
		},
		rr: roadrunner.NewServer(&roadrunner.ServerConfig{
			Command: "php ../../tests/http/client.php payload pipes",
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

	hs := &http.Server{Addr: ":8081", Handler: st}
	defer hs.Shutdown(context.Background())

	go func() { hs.ListenAndServe() }()
	time.Sleep(time.Millisecond * 10)

	req, err := http.NewRequest("PUT", "http://localhost"+hs.Addr, bytes.NewBufferString(`{"key":"value"}`))
	assert.NoError(t, err)

	req.Header.Add("Content-Type", "application/json")

	r, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer r.Body.Close()

	b, err := ioutil.ReadAll(r.Body)
	assert.NoError(t, err)

	assert.NoError(t, err)
	assert.Equal(t, 200, r.StatusCode)
	assert.Equal(t, `{"value":"key"}`, string(b))
}

func TestServer_JsonPayload_PATCH(t *testing.T) {
	st := &Handler{
		cfg: &Config{
			MaxRequest: 1024,
			Uploads: &UploadsConfig{
				Dir:    os.TempDir(),
				Forbid: []string{},
			},
		},
		rr: roadrunner.NewServer(&roadrunner.ServerConfig{
			Command: "php ../../tests/http/client.php payload pipes",
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

	hs := &http.Server{Addr: ":8082", Handler: st}
	defer hs.Shutdown(context.Background())

	go func() { hs.ListenAndServe() }()
	time.Sleep(time.Millisecond * 10)

	req, err := http.NewRequest("PATCH", "http://localhost"+hs.Addr, bytes.NewBufferString(`{"key":"value"}`))
	assert.NoError(t, err)

	req.Header.Add("Content-Type", "application/json")

	r, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer r.Body.Close()

	b, err := ioutil.ReadAll(r.Body)
	assert.NoError(t, err)

	assert.NoError(t, err)
	assert.Equal(t, 200, r.StatusCode)
	assert.Equal(t, `{"value":"key"}`, string(b))
}

func TestServer_FormData_POST(t *testing.T) {
	st := &Handler{
		cfg: &Config{
			MaxRequest: 1024,
			Uploads: &UploadsConfig{
				Dir:    os.TempDir(),
				Forbid: []string{},
			},
		},
		rr: roadrunner.NewServer(&roadrunner.ServerConfig{
			Command: "php ../../tests/http/client.php data pipes",
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

	hs := &http.Server{Addr: ":8083", Handler: st}
	defer hs.Shutdown(context.Background())

	go func() { hs.ListenAndServe() }()
	time.Sleep(time.Millisecond * 10)

	form := url.Values{}

	form.Add("key", "value")
	form.Add("name[]", "name1")
	form.Add("name[]", "name2")
	form.Add("name[]", "name3")
	form.Add("arr[x][y][z]", "y")
	form.Add("arr[x][y][e]", "f")
	form.Add("arr[c]p", "l")
	form.Add("arr[c]z", "")

	req, err := http.NewRequest("POST", "http://localhost"+hs.Addr, strings.NewReader(form.Encode()))
	assert.NoError(t, err)

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	r, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer r.Body.Close()

	b, err := ioutil.ReadAll(r.Body)
	assert.NoError(t, err)

	assert.NoError(t, err)
	assert.Equal(t, 200, r.StatusCode)

	assert.Equal(t, `{"arr":{"c":{"p":"l","z":""},"x":{"y":{"e":"f","z":"y"}}},"key":"value","name":["name1","name2","name3"]}`, string(b))
}

func TestServer_FormData_PUT(t *testing.T) {
	st := &Handler{
		cfg: &Config{
			MaxRequest: 1024,
			Uploads: &UploadsConfig{
				Dir:    os.TempDir(),
				Forbid: []string{},
			},
		},
		rr: roadrunner.NewServer(&roadrunner.ServerConfig{
			Command: "php ../../tests/http/client.php data pipes",
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

	hs := &http.Server{Addr: ":8084", Handler: st}
	defer hs.Shutdown(context.Background())

	go func() { hs.ListenAndServe() }()
	time.Sleep(time.Millisecond * 10)

	form := url.Values{}

	form.Add("key", "value")
	form.Add("name[]", "name1")
	form.Add("name[]", "name2")
	form.Add("name[]", "name3")
	form.Add("arr[x][y][z]", "y")
	form.Add("arr[x][y][e]", "f")
	form.Add("arr[c]p", "l")
	form.Add("arr[c]z", "")

	req, err := http.NewRequest("PUT", "http://localhost"+hs.Addr, strings.NewReader(form.Encode()))
	assert.NoError(t, err)

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	r, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer r.Body.Close()

	b, err := ioutil.ReadAll(r.Body)
	assert.NoError(t, err)

	assert.NoError(t, err)
	assert.Equal(t, 200, r.StatusCode)

	assert.Equal(t, `{"arr":{"c":{"p":"l","z":""},"x":{"y":{"e":"f","z":"y"}}},"key":"value","name":["name1","name2","name3"]}`, string(b))
}

func TestServer_FormData_PATCH(t *testing.T) {
	st := &Handler{
		cfg: &Config{
			MaxRequest: 1024,
			Uploads: &UploadsConfig{
				Dir:    os.TempDir(),
				Forbid: []string{},
			},
		},
		rr: roadrunner.NewServer(&roadrunner.ServerConfig{
			Command: "php ../../tests/http/client.php data pipes",
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

	hs := &http.Server{Addr: ":8085", Handler: st}
	defer hs.Shutdown(context.Background())

	go func() { hs.ListenAndServe() }()
	time.Sleep(time.Millisecond * 10)

	form := url.Values{}

	form.Add("key", "value")
	form.Add("name[]", "name1")
	form.Add("name[]", "name2")
	form.Add("name[]", "name3")
	form.Add("arr[x][y][z]", "y")
	form.Add("arr[x][y][e]", "f")
	form.Add("arr[c]p", "l")
	form.Add("arr[c]z", "")

	req, err := http.NewRequest("PATCH", "http://localhost"+hs.Addr, strings.NewReader(form.Encode()))
	assert.NoError(t, err)

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	r, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer r.Body.Close()

	b, err := ioutil.ReadAll(r.Body)
	assert.NoError(t, err)

	assert.NoError(t, err)
	assert.Equal(t, 200, r.StatusCode)

	assert.Equal(t, `{"arr":{"c":{"p":"l","z":""},"x":{"y":{"e":"f","z":"y"}}},"key":"value","name":["name1","name2","name3"]}`, string(b))
}

func TestServer_Multipart_POST(t *testing.T) {
	st := &Handler{
		cfg: &Config{
			MaxRequest: 1024,
			Uploads: &UploadsConfig{
				Dir:    os.TempDir(),
				Forbid: []string{},
			},
		},
		rr: roadrunner.NewServer(&roadrunner.ServerConfig{
			Command: "php ../../tests/http/client.php data pipes",
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

	hs := &http.Server{Addr: ":8019", Handler: st}
	defer hs.Shutdown(context.Background())

	go func() { hs.ListenAndServe() }()
	time.Sleep(time.Millisecond * 10)

	var mb bytes.Buffer
	w := multipart.NewWriter(&mb)
	w.WriteField("key", "value")

	w.WriteField("key", "value")
	w.WriteField("name[]", "name1")
	w.WriteField("name[]", "name2")
	w.WriteField("name[]", "name3")
	w.WriteField("arr[x][y][z]", "y")
	w.WriteField("arr[x][y][e]", "f")
	w.WriteField("arr[c]p", "l")
	w.WriteField("arr[c]z", "")

	w.Close()

	req, err := http.NewRequest("POST", "http://localhost"+hs.Addr, &mb)
	assert.NoError(t, err)

	req.Header.Set("Content-Type", w.FormDataContentType())

	r, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer r.Body.Close()

	b, err := ioutil.ReadAll(r.Body)
	assert.NoError(t, err)

	assert.NoError(t, err)
	assert.Equal(t, 200, r.StatusCode)

	assert.Equal(t, `{"arr":{"c":{"p":"l","z":""},"x":{"y":{"e":"f","z":"y"}}},"key":"value","name":["name1","name2","name3"]}`, string(b))
}

func TestServer_Multipart_PUT(t *testing.T) {
	st := &Handler{
		cfg: &Config{
			MaxRequest: 1024,
			Uploads: &UploadsConfig{
				Dir:    os.TempDir(),
				Forbid: []string{},
			},
		},
		rr: roadrunner.NewServer(&roadrunner.ServerConfig{
			Command: "php ../../tests/http/client.php data pipes",
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

	hs := &http.Server{Addr: ":8020", Handler: st}
	defer hs.Shutdown(context.Background())

	go func() { hs.ListenAndServe() }()
	time.Sleep(time.Millisecond * 10)

	var mb bytes.Buffer
	w := multipart.NewWriter(&mb)
	w.WriteField("key", "value")

	w.WriteField("key", "value")
	w.WriteField("name[]", "name1")
	w.WriteField("name[]", "name2")
	w.WriteField("name[]", "name3")
	w.WriteField("arr[x][y][z]", "y")
	w.WriteField("arr[x][y][e]", "f")
	w.WriteField("arr[c]p", "l")
	w.WriteField("arr[c]z", "")

	w.Close()

	req, err := http.NewRequest("PUT", "http://localhost"+hs.Addr, &mb)
	assert.NoError(t, err)

	req.Header.Set("Content-Type", w.FormDataContentType())

	r, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer r.Body.Close()

	b, err := ioutil.ReadAll(r.Body)
	assert.NoError(t, err)

	assert.NoError(t, err)
	assert.Equal(t, 200, r.StatusCode)

	assert.Equal(t, `{"arr":{"c":{"p":"l","z":""},"x":{"y":{"e":"f","z":"y"}}},"key":"value","name":["name1","name2","name3"]}`, string(b))
}

func TestServer_Multipart_PATCH(t *testing.T) {
	st := &Handler{
		cfg: &Config{
			MaxRequest: 1024,
			Uploads: &UploadsConfig{
				Dir:    os.TempDir(),
				Forbid: []string{},
			},
		},
		rr: roadrunner.NewServer(&roadrunner.ServerConfig{
			Command: "php ../../tests/http/client.php data pipes",
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

	hs := &http.Server{Addr: ":8021", Handler: st}
	defer hs.Shutdown(context.Background())

	go func() { hs.ListenAndServe() }()
	time.Sleep(time.Millisecond * 10)

	var mb bytes.Buffer
	w := multipart.NewWriter(&mb)
	w.WriteField("key", "value")

	w.WriteField("key", "value")
	w.WriteField("name[]", "name1")
	w.WriteField("name[]", "name2")
	w.WriteField("name[]", "name3")
	w.WriteField("arr[x][y][z]", "y")
	w.WriteField("arr[x][y][e]", "f")
	w.WriteField("arr[c]p", "l")
	w.WriteField("arr[c]z", "")

	w.Close()

	req, err := http.NewRequest("PATCH", "http://localhost"+hs.Addr, &mb)
	assert.NoError(t, err)

	req.Header.Set("Content-Type", w.FormDataContentType())

	r, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer r.Body.Close()

	b, err := ioutil.ReadAll(r.Body)
	assert.NoError(t, err)

	assert.NoError(t, err)
	assert.Equal(t, 200, r.StatusCode)

	assert.Equal(t, `{"arr":{"c":{"p":"l","z":""},"x":{"y":{"e":"f","z":"y"}}},"key":"value","name":["name1","name2","name3"]}`, string(b))
}

func TestServer_Error(t *testing.T) {
	st := &Handler{
		cfg: &Config{
			MaxRequest: 1024,
			Uploads: &UploadsConfig{
				Dir:    os.TempDir(),
				Forbid: []string{},
			},
		},
		rr: roadrunner.NewServer(&roadrunner.ServerConfig{
			Command: "php ../../tests/http/client.php error pipes",
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

	hs := &http.Server{Addr: ":8177", Handler: st}
	defer hs.Shutdown(context.Background())

	go func() { hs.ListenAndServe() }()
	time.Sleep(time.Millisecond * 10)

	_, r, err := get("http://localhost:8177/?hello=world")
	assert.NoError(t, err)
	assert.Equal(t, 500, r.StatusCode)
}

func TestServer_Error2(t *testing.T) {
	st := &Handler{
		cfg: &Config{
			MaxRequest: 1024,
			Uploads: &UploadsConfig{
				Dir:    os.TempDir(),
				Forbid: []string{},
			},
		},
		rr: roadrunner.NewServer(&roadrunner.ServerConfig{
			Command: "php ../../tests/http/client.php error2 pipes",
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

	hs := &http.Server{Addr: ":8177", Handler: st}
	defer hs.Shutdown(context.Background())

	go func() { hs.ListenAndServe() }()
	time.Sleep(time.Millisecond * 10)

	_, r, err := get("http://localhost:8177/?hello=world")
	assert.NoError(t, err)
	assert.Equal(t, 500, r.StatusCode)
}

func TestServer_Error3(t *testing.T) {
	st := &Handler{
		cfg: &Config{
			MaxRequest: 1,
			Uploads: &UploadsConfig{
				Dir:    os.TempDir(),
				Forbid: []string{},
			},
		},
		rr: roadrunner.NewServer(&roadrunner.ServerConfig{
			Command: "php ../../tests/http/client.php pid pipes",
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

	hs := &http.Server{Addr: ":8177", Handler: st}
	defer hs.Shutdown(context.Background())

	go func() { hs.ListenAndServe() }()
	time.Sleep(time.Millisecond * 10)

	b2 := &bytes.Buffer{}
	for i := 0; i < 1024*1024; i++ {
		b2.Write([]byte("  "))
	}

	req, err := http.NewRequest("POST", "http://localhost"+hs.Addr, b2)
	assert.NoError(t, err)

	r, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer r.Body.Close()

	assert.NoError(t, err)
	assert.Equal(t, 500, r.StatusCode)
}

func BenchmarkHandler_Listen_Echo(b *testing.B) {
	st := &Handler{
		cfg: &Config{
			MaxRequest: 1024,
			Uploads: &UploadsConfig{
				Dir:    os.TempDir(),
				Forbid: []string{},
			},
		},
		rr: roadrunner.NewServer(&roadrunner.ServerConfig{
			Command: "php ../../tests/http/client.php echo pipes",
			Relay:   "pipes",
			Pool: &roadrunner.Config{
				NumWorkers:      int64(runtime.NumCPU()),
				AllocateTimeout: 10000000,
				DestroyTimeout:  10000000,
			},
		}),
	}

	st.rr.Start()
	defer st.rr.Stop()

	hs := &http.Server{Addr: ":8177", Handler: st}
	defer hs.Shutdown(context.Background())

	go func() { hs.ListenAndServe() }()
	time.Sleep(time.Millisecond * 10)

	bb := "WORLD"
	for n := 0; n < b.N; n++ {
		r, err := http.Get("http://localhost:8177/?hello=world")
		if err != nil {
			b.Fail()
		}
		defer r.Body.Close()

		br, _ := ioutil.ReadAll(r.Body)
		if string(br) != bb {
			b.Fail()
		}
	}
}
