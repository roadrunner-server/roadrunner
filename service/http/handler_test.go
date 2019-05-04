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

// get request and return body
func getHeader(url string, h map[string]string) (string, *http.Response, error) {
	req, err := http.NewRequest("GET", url, bytes.NewBuffer(nil))
	if err != nil {
		return "", nil, err
	}

	for k, v := range h {
		req.Header.Set(k, v)
	}

	r, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", nil, err
	}
	defer r.Body.Close()

	b, err := ioutil.ReadAll(r.Body)
	return string(b), r, err
}

func TestHandler_Echo(t *testing.T) {
	h := &Handler{
		cfg: &Config{
			MaxRequestSize: 1024,
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

	assert.NoError(t, h.rr.Start())
	defer h.rr.Stop()

	hs := &http.Server{Addr: ":8177", Handler: h}
	defer hs.Shutdown(context.Background())

	go func() { hs.ListenAndServe() }()
	time.Sleep(time.Millisecond * 10)

	body, r, err := get("http://localhost:8177/?hello=world")
	assert.NoError(t, err)
	assert.Equal(t, 201, r.StatusCode)
	assert.Equal(t, "WORLD", body)
}

func Test_HandlerErrors(t *testing.T) {
	h := &Handler{
		cfg: &Config{
			MaxRequestSize: 1024,
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

	h.ServeHTTP(wr, rq)
	assert.Equal(t, 500, wr.Code)
}

func Test_Handler_JSON_error(t *testing.T) {
	h := &Handler{
		cfg: &Config{
			MaxRequestSize: 1024,
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

	h.ServeHTTP(wr, rq)
	assert.Equal(t, 500, wr.Code)
}

func TestHandler_Headers(t *testing.T) {
	h := &Handler{
		cfg: &Config{
			MaxRequestSize: 1024,
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

	assert.NoError(t, h.rr.Start())
	defer h.rr.Stop()

	hs := &http.Server{Addr: ":8078", Handler: h}
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

func TestHandler_Empty_User_Agent(t *testing.T) {
	h := &Handler{
		cfg: &Config{
			MaxRequestSize: 1024,
			Uploads: &UploadsConfig{
				Dir:    os.TempDir(),
				Forbid: []string{},
			},
		},
		rr: roadrunner.NewServer(&roadrunner.ServerConfig{
			Command: "php ../../tests/http/client.php user-agent pipes",
			Relay:   "pipes",
			Pool: &roadrunner.Config{
				NumWorkers:      1,
				AllocateTimeout: 10000000,
				DestroyTimeout:  10000000,
			},
		}),
	}

	assert.NoError(t, h.rr.Start())
	defer h.rr.Stop()

	hs := &http.Server{Addr: ":8088", Handler: h}
	defer hs.Shutdown(context.Background())

	go func() { hs.ListenAndServe() }()
	time.Sleep(time.Millisecond * 10)

	req, err := http.NewRequest("GET", "http://localhost:8088?hello=world", nil)
	assert.NoError(t, err)

	req.Header.Add("user-agent", "")

	r, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer r.Body.Close()

	b, err := ioutil.ReadAll(r.Body)
	assert.NoError(t, err)

	assert.NoError(t, err)
	assert.Equal(t, 200, r.StatusCode)
	assert.Equal(t, "", string(b))
}

func TestHandler_User_Agent(t *testing.T) {
	h := &Handler{
		cfg: &Config{
			MaxRequestSize: 1024,
			Uploads: &UploadsConfig{
				Dir:    os.TempDir(),
				Forbid: []string{},
			},
		},
		rr: roadrunner.NewServer(&roadrunner.ServerConfig{
			Command: "php ../../tests/http/client.php user-agent pipes",
			Relay:   "pipes",
			Pool: &roadrunner.Config{
				NumWorkers:      1,
				AllocateTimeout: 10000000,
				DestroyTimeout:  10000000,
			},
		}),
	}

	assert.NoError(t, h.rr.Start())
	defer h.rr.Stop()

	hs := &http.Server{Addr: ":8088", Handler: h}
	defer hs.Shutdown(context.Background())

	go func() { hs.ListenAndServe() }()
	time.Sleep(time.Millisecond * 10)

	req, err := http.NewRequest("GET", "http://localhost:8088?hello=world", nil)
	assert.NoError(t, err)

	req.Header.Add("User-Agent", "go-agent")

	r, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer r.Body.Close()

	b, err := ioutil.ReadAll(r.Body)
	assert.NoError(t, err)

	assert.NoError(t, err)
	assert.Equal(t, 200, r.StatusCode)
	assert.Equal(t, "go-agent", string(b))
}

func TestHandler_Cookies(t *testing.T) {
	h := &Handler{
		cfg: &Config{
			MaxRequestSize: 1024,
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

	assert.NoError(t, h.rr.Start())
	defer h.rr.Stop()

	hs := &http.Server{Addr: ":8079", Handler: h}
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

func TestHandler_JsonPayload_POST(t *testing.T) {
	h := &Handler{
		cfg: &Config{
			MaxRequestSize: 1024,
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

	assert.NoError(t, h.rr.Start())
	defer h.rr.Stop()

	hs := &http.Server{Addr: ":8090", Handler: h}
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

func TestHandler_JsonPayload_PUT(t *testing.T) {
	h := &Handler{
		cfg: &Config{
			MaxRequestSize: 1024,
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

	assert.NoError(t, h.rr.Start())
	defer h.rr.Stop()

	hs := &http.Server{Addr: ":8081", Handler: h}
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

func TestHandler_JsonPayload_PATCH(t *testing.T) {
	h := &Handler{
		cfg: &Config{
			MaxRequestSize: 1024,
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

	assert.NoError(t, h.rr.Start())
	defer h.rr.Stop()

	hs := &http.Server{Addr: ":8082", Handler: h}
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

func TestHandler_FormData_POST(t *testing.T) {
	h := &Handler{
		cfg: &Config{
			MaxRequestSize: 1024,
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

	assert.NoError(t, h.rr.Start())
	defer h.rr.Stop()

	hs := &http.Server{Addr: ":8083", Handler: h}
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

func TestHandler_FormData_POST_Overwrite(t *testing.T) {
	h := &Handler{
		cfg: &Config{
			MaxRequestSize: 1024,
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

	assert.NoError(t, h.rr.Start())
	defer h.rr.Stop()

	hs := &http.Server{Addr: ":8083", Handler: h}
	defer hs.Shutdown(context.Background())

	go func() { hs.ListenAndServe() }()
	time.Sleep(time.Millisecond * 10)

	form := url.Values{}

	form.Add("key", "value1")
	form.Add("key", "value2")

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

	assert.Equal(t, `{"key":"value2","arr":{"x":{"y":null}}}`, string(b))
}

func TestHandler_FormData_POST_Form_UrlEncoded_Charset(t *testing.T) {
	h := &Handler{
		cfg: &Config{
			MaxRequestSize: 1024,
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

	assert.NoError(t, h.rr.Start())
	defer h.rr.Stop()

	hs := &http.Server{Addr: ":8083", Handler: h}
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

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")

	r, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer r.Body.Close()

	b, err := ioutil.ReadAll(r.Body)
	assert.NoError(t, err)

	assert.NoError(t, err)
	assert.Equal(t, 200, r.StatusCode)

	assert.Equal(t, `{"arr":{"c":{"p":"l","z":""},"x":{"y":{"e":"f","z":"y"}}},"key":"value","name":["name1","name2","name3"]}`, string(b))
}

func TestHandler_FormData_PUT(t *testing.T) {
	h := &Handler{
		cfg: &Config{
			MaxRequestSize: 1024,
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

	assert.NoError(t, h.rr.Start())
	defer h.rr.Stop()

	hs := &http.Server{Addr: ":8084", Handler: h}
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

func TestHandler_FormData_PATCH(t *testing.T) {
	h := &Handler{
		cfg: &Config{
			MaxRequestSize: 1024,
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

	assert.NoError(t, h.rr.Start())
	defer h.rr.Stop()

	hs := &http.Server{Addr: ":8085", Handler: h}
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

func TestHandler_Multipart_POST(t *testing.T) {
	h := &Handler{
		cfg: &Config{
			MaxRequestSize: 1024,
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

	assert.NoError(t, h.rr.Start())
	defer h.rr.Stop()

	hs := &http.Server{Addr: ":8019", Handler: h}
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

func TestHandler_Multipart_PUT(t *testing.T) {
	h := &Handler{
		cfg: &Config{
			MaxRequestSize: 1024,
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

	assert.NoError(t, h.rr.Start())
	defer h.rr.Stop()

	hs := &http.Server{Addr: ":8020", Handler: h}
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

func TestHandler_Multipart_PATCH(t *testing.T) {
	h := &Handler{
		cfg: &Config{
			MaxRequestSize: 1024,
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

	assert.NoError(t, h.rr.Start())
	defer h.rr.Stop()

	hs := &http.Server{Addr: ":8021", Handler: h}
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

func TestHandler_Error(t *testing.T) {
	h := &Handler{
		cfg: &Config{
			MaxRequestSize: 1024,
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

	assert.NoError(t, h.rr.Start())
	defer h.rr.Stop()

	hs := &http.Server{Addr: ":8177", Handler: h}
	defer hs.Shutdown(context.Background())

	go func() { hs.ListenAndServe() }()
	time.Sleep(time.Millisecond * 10)

	_, r, err := get("http://localhost:8177/?hello=world")
	assert.NoError(t, err)
	assert.Equal(t, 500, r.StatusCode)
}

func TestHandler_Error2(t *testing.T) {
	h := &Handler{
		cfg: &Config{
			MaxRequestSize: 1024,
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

	assert.NoError(t, h.rr.Start())
	defer h.rr.Stop()

	hs := &http.Server{Addr: ":8177", Handler: h}
	defer hs.Shutdown(context.Background())

	go func() { hs.ListenAndServe() }()
	time.Sleep(time.Millisecond * 10)

	_, r, err := get("http://localhost:8177/?hello=world")
	assert.NoError(t, err)
	assert.Equal(t, 500, r.StatusCode)
}

func TestHandler_Error3(t *testing.T) {
	h := &Handler{
		cfg: &Config{
			MaxRequestSize: 1,
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

	assert.NoError(t, h.rr.Start())
	defer h.rr.Stop()

	hs := &http.Server{Addr: ":8177", Handler: h}
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

func TestHandler_ResponseDuration(t *testing.T) {
	h := &Handler{
		cfg: &Config{
			MaxRequestSize: 1024,
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

	assert.NoError(t, h.rr.Start())
	defer h.rr.Stop()

	hs := &http.Server{Addr: ":8177", Handler: h}
	defer hs.Shutdown(context.Background())

	go func() { hs.ListenAndServe() }()
	time.Sleep(time.Millisecond * 10)

	gotresp := make(chan interface{})
	h.Listen(func(event int, ctx interface{}) {
		if event == EventResponse {
			c := ctx.(*ResponseEvent)

			if c.Elapsed() > 0 {
				close(gotresp)
			}
		}
	})

	body, r, err := get("http://localhost:8177/?hello=world")
	assert.NoError(t, err)

	<-gotresp

	assert.Equal(t, 201, r.StatusCode)
	assert.Equal(t, "WORLD", body)
}

func TestHandler_ResponseDurationDelayed(t *testing.T) {
	h := &Handler{
		cfg: &Config{
			MaxRequestSize: 1024,
			Uploads: &UploadsConfig{
				Dir:    os.TempDir(),
				Forbid: []string{},
			},
		},
		rr: roadrunner.NewServer(&roadrunner.ServerConfig{
			Command: "php ../../tests/http/client.php echoDelay pipes",
			Relay:   "pipes",
			Pool: &roadrunner.Config{
				NumWorkers:      1,
				AllocateTimeout: 10000000,
				DestroyTimeout:  10000000,
			},
		}),
	}

	assert.NoError(t, h.rr.Start())
	defer h.rr.Stop()

	hs := &http.Server{Addr: ":8177", Handler: h}
	defer hs.Shutdown(context.Background())

	go func() { hs.ListenAndServe() }()
	time.Sleep(time.Millisecond * 10)

	gotresp := make(chan interface{})
	h.Listen(func(event int, ctx interface{}) {
		if event == EventResponse {
			c := ctx.(*ResponseEvent)

			if c.Elapsed() > time.Second {
				close(gotresp)
			}
		}
	})

	body, r, err := get("http://localhost:8177/?hello=world")
	assert.NoError(t, err)

	<-gotresp

	assert.Equal(t, 201, r.StatusCode)
	assert.Equal(t, "WORLD", body)
}

func TestHandler_ErrorDuration(t *testing.T) {
	h := &Handler{
		cfg: &Config{
			MaxRequestSize: 1024,
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

	assert.NoError(t, h.rr.Start())
	defer h.rr.Stop()

	hs := &http.Server{Addr: ":8177", Handler: h}
	defer hs.Shutdown(context.Background())

	go func() { hs.ListenAndServe() }()
	time.Sleep(time.Millisecond * 10)

	goterr := make(chan interface{})
	h.Listen(func(event int, ctx interface{}) {
		if event == EventError {
			c := ctx.(*ErrorEvent)

			if c.Elapsed() > 0 {
				close(goterr)
			}
		}
	})

	_, r, err := get("http://localhost:8177/?hello=world")
	assert.NoError(t, err)

	<-goterr

	assert.Equal(t, 500, r.StatusCode)
}

func TestHandler_IP(t *testing.T) {
	h := &Handler{
		cfg: &Config{
			MaxRequestSize: 1024,
			Uploads: &UploadsConfig{
				Dir:    os.TempDir(),
				Forbid: []string{},
			},
			TrustedSubnets: []string{
				"10.0.0.0/8",
				"127.0.0.0/8",
				"172.16.0.0/12",
				"192.168.0.0/16",
				"::1/128",
				"fc00::/7",
				"fe80::/10",
			},
		},
		rr: roadrunner.NewServer(&roadrunner.ServerConfig{
			Command: "php ../../tests/http/client.php ip pipes",
			Relay:   "pipes",
			Pool: &roadrunner.Config{
				NumWorkers:      1,
				AllocateTimeout: 10000000,
				DestroyTimeout:  10000000,
			},
		}),
	}

	h.cfg.parseCIDRs()

	assert.NoError(t, h.rr.Start())
	defer h.rr.Stop()

	hs := &http.Server{Addr: "127.0.0.1:8177", Handler: h}
	defer hs.Shutdown(context.Background())

	go func() { hs.ListenAndServe() }()
	time.Sleep(time.Millisecond * 10)

	body, r, err := get("http://127.0.0.1:8177/")
	assert.NoError(t, err)
	assert.Equal(t, 200, r.StatusCode)
	assert.Equal(t, "127.0.0.1", body)
}

func TestHandler_XRealIP(t *testing.T) {
	h := &Handler{
		cfg: &Config{
			MaxRequestSize: 1024,
			Uploads: &UploadsConfig{
				Dir:    os.TempDir(),
				Forbid: []string{},
			},
			TrustedSubnets: []string{
				"10.0.0.0/8",
				"127.0.0.0/8",
				"172.16.0.0/12",
				"192.168.0.0/16",
				"::1/128",
				"fc00::/7",
				"fe80::/10",
			},
		},
		rr: roadrunner.NewServer(&roadrunner.ServerConfig{
			Command: "php ../../tests/http/client.php ip pipes",
			Relay:   "pipes",
			Pool: &roadrunner.Config{
				NumWorkers:      1,
				AllocateTimeout: 10000000,
				DestroyTimeout:  10000000,
			},
		}),
	}

	h.cfg.parseCIDRs()

	assert.NoError(t, h.rr.Start())
	defer h.rr.Stop()

	hs := &http.Server{Addr: "127.0.0.1:8177", Handler: h}
	defer hs.Shutdown(context.Background())

	go func() { hs.ListenAndServe() }()
	time.Sleep(time.Millisecond * 10)

	body, r, err := getHeader("http://127.0.0.1:8177/", map[string]string{
		"X-Real-Ip": "200.0.0.1",
	})

	assert.NoError(t, err)
	assert.Equal(t, 200, r.StatusCode)
	assert.Equal(t, "200.0.0.1", body)
}

func TestHandler_XForwardedFor(t *testing.T) {
	h := &Handler{
		cfg: &Config{
			MaxRequestSize: 1024,
			Uploads: &UploadsConfig{
				Dir:    os.TempDir(),
				Forbid: []string{},
			},
			TrustedSubnets: []string{
				"10.0.0.0/8",
				"127.0.0.0/8",
				"172.16.0.0/12",
				"192.168.0.0/16",
				"100.0.0.0/16",
				"200.0.0.0/16",
				"::1/128",
				"fc00::/7",
				"fe80::/10",
			},
		},
		rr: roadrunner.NewServer(&roadrunner.ServerConfig{
			Command: "php ../../tests/http/client.php ip pipes",
			Relay:   "pipes",
			Pool: &roadrunner.Config{
				NumWorkers:      1,
				AllocateTimeout: 10000000,
				DestroyTimeout:  10000000,
			},
		}),
	}

	h.cfg.parseCIDRs()

	assert.NoError(t, h.rr.Start())
	defer h.rr.Stop()

	hs := &http.Server{Addr: "127.0.0.1:8177", Handler: h}
	defer hs.Shutdown(context.Background())

	go func() { hs.ListenAndServe() }()
	time.Sleep(time.Millisecond * 10)

	body, r, err := getHeader("http://127.0.0.1:8177/", map[string]string{
		"X-Forwarded-For": "100.0.0.1, 200.0.0.1, invalid, 101.0.0.1",
	})

	assert.NoError(t, err)
	assert.Equal(t, 200, r.StatusCode)
	assert.Equal(t, "200.0.0.1", body)
}

func BenchmarkHandler_Listen_Echo(b *testing.B) {
	h := &Handler{
		cfg: &Config{
			MaxRequestSize: 1024,
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

	h.rr.Start()
	defer h.rr.Stop()

	hs := &http.Server{Addr: ":8177", Handler: h}
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
