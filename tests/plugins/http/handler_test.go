package http

import (
	"bytes"
	"context"
	"io/ioutil"
	"mime/multipart"
	"net/url"
	"os/exec"
	"runtime"
	"strings"

	"github.com/spiral/roadrunner/v2/pkg/pool"
	"github.com/spiral/roadrunner/v2/pkg/transport/pipe"
	"github.com/spiral/roadrunner/v2/plugins/http/config"
	handler "github.com/spiral/roadrunner/v2/plugins/http/worker_handler"
	"github.com/stretchr/testify/assert"

	"net/http"
	"os"
	"testing"
	"time"
)

func TestHandler_Echo(t *testing.T) {
	p, err := pool.Initialize(context.Background(),
		func() *exec.Cmd { return exec.Command("php", "../../http/client.php", "echo", "pipes") },
		pipe.NewPipeFactory(),
		pool.Config{
			NumWorkers:      1,
			AllocateTimeout: time.Second * 1000,
			DestroyTimeout:  time.Second * 1000,
		})
	if err != nil {
		t.Fatal(err)
	}

	h, err := handler.NewHandler(1024, config.Uploads{
		Dir:    os.TempDir(),
		Forbid: []string{},
	}, nil, p)
	assert.NoError(t, err)

	hs := &http.Server{Addr: ":8177", Handler: h}
	defer func() {
		errS := hs.Shutdown(context.Background())
		if errS != nil {
			t.Errorf("error during the shutdown: error %v", err)
		}
	}()
	go func(server *http.Server) {
		err = server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			t.Errorf("error listening the interface: error %v", err)
		}
	}(hs)
	time.Sleep(time.Millisecond * 10)

	body, r, err := get("http://localhost:8177/?hello=world")
	assert.NoError(t, err)
	defer func() {
		_ = r.Body.Close()
	}()
	assert.Equal(t, 201, r.StatusCode)
	assert.Equal(t, "WORLD", body)
}

func Test_HandlerErrors(t *testing.T) {
	_, err := handler.NewHandler(1024, config.Uploads{
		Dir:    os.TempDir(),
		Forbid: []string{},
	}, nil, nil)
	assert.Error(t, err)
}

func TestHandler_Headers(t *testing.T) {
	p, err := pool.Initialize(context.Background(),
		func() *exec.Cmd { return exec.Command("php", "../../http/client.php", "header", "pipes") },
		pipe.NewPipeFactory(),
		pool.Config{
			NumWorkers:      1,
			AllocateTimeout: time.Second * 1000,
			DestroyTimeout:  time.Second * 1000,
		})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		p.Destroy(context.Background())
	}()

	h, err := handler.NewHandler(1024, config.Uploads{
		Dir:    os.TempDir(),
		Forbid: []string{},
	}, nil, p)
	assert.NoError(t, err)

	hs := &http.Server{Addr: ":8078", Handler: h}
	defer func() {
		errS := hs.Shutdown(context.Background())
		if errS != nil {
			t.Errorf("error during the shutdown: error %v", err)
		}
	}()

	go func() {
		err = hs.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			t.Errorf("error listening the interface: error %v", err)
		}
	}()
	time.Sleep(time.Millisecond * 100)

	req, err := http.NewRequest("GET", "http://localhost:8078?hello=world", nil)
	assert.NoError(t, err)

	req.Header.Add("input", "sample")

	r, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer func() {
		err = r.Body.Close()
		if err != nil {
			t.Errorf("error during the closing Body: error %v", err)
		}
	}()

	b, err := ioutil.ReadAll(r.Body)
	assert.NoError(t, err)

	assert.NoError(t, err)
	assert.Equal(t, 200, r.StatusCode)
	assert.Equal(t, "world", r.Header.Get("Header"))
	assert.Equal(t, "SAMPLE", string(b))
}

func TestHandler_Empty_User_Agent(t *testing.T) {
	p, err := pool.Initialize(context.Background(),
		func() *exec.Cmd { return exec.Command("php", "../../http/client.php", "user-agent", "pipes") },
		pipe.NewPipeFactory(),
		pool.Config{
			NumWorkers:      1,
			AllocateTimeout: time.Second * 1000,
			DestroyTimeout:  time.Second * 1000,
		})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		p.Destroy(context.Background())
	}()

	h, err := handler.NewHandler(1024, config.Uploads{
		Dir:    os.TempDir(),
		Forbid: []string{},
	}, nil, p)
	assert.NoError(t, err)

	hs := &http.Server{Addr: ":19658", Handler: h}
	defer func() {
		errS := hs.Shutdown(context.Background())
		if errS != nil {
			t.Errorf("error during the shutdown: error %v", err)
		}
	}()

	go func() {
		err = hs.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			t.Errorf("error listening the interface: error %v", err)
		}
	}()
	time.Sleep(time.Millisecond * 10)

	req, err := http.NewRequest("GET", "http://localhost:19658?hello=world", nil)
	assert.NoError(t, err)

	req.Header.Add("user-agent", "")

	r, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer func() {
		err = r.Body.Close()
		if err != nil {
			t.Errorf("error during the closing Body: error %v", err)
		}
	}()

	b, err := ioutil.ReadAll(r.Body)
	assert.NoError(t, err)

	assert.NoError(t, err)
	assert.Equal(t, 200, r.StatusCode)
	assert.Equal(t, "", string(b))
}

func TestHandler_User_Agent(t *testing.T) {
	p, err := pool.Initialize(context.Background(),
		func() *exec.Cmd { return exec.Command("php", "../../http/client.php", "user-agent", "pipes") },
		pipe.NewPipeFactory(),
		pool.Config{
			NumWorkers:      1,
			AllocateTimeout: time.Second * 1000,
			DestroyTimeout:  time.Second * 1000,
		})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		p.Destroy(context.Background())
	}()

	h, err := handler.NewHandler(1024, config.Uploads{
		Dir:    os.TempDir(),
		Forbid: []string{},
	}, nil, p)
	assert.NoError(t, err)

	hs := &http.Server{Addr: ":25688", Handler: h}
	defer func() {
		errS := hs.Shutdown(context.Background())
		if errS != nil {
			t.Errorf("error during the shutdown: error %v", err)
		}
	}()

	go func() {
		err = hs.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			t.Errorf("error listening the interface: error %v", err)
		}
	}()
	time.Sleep(time.Millisecond * 10)

	req, err := http.NewRequest("GET", "http://localhost:25688?hello=world", nil)
	assert.NoError(t, err)

	req.Header.Add("User-Agent", "go-agent")

	r, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer func() {
		err = r.Body.Close()
		if err != nil {
			t.Errorf("error during the closing Body: error %v", err)
		}
	}()

	b, err := ioutil.ReadAll(r.Body)
	assert.NoError(t, err)

	assert.NoError(t, err)
	assert.Equal(t, 200, r.StatusCode)
	assert.Equal(t, "go-agent", string(b))
}

func TestHandler_Cookies(t *testing.T) {
	p, err := pool.Initialize(context.Background(),
		func() *exec.Cmd { return exec.Command("php", "../../http/client.php", "cookie", "pipes") },
		pipe.NewPipeFactory(),
		pool.Config{
			NumWorkers:      1,
			AllocateTimeout: time.Second * 1000,
			DestroyTimeout:  time.Second * 1000,
		})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		p.Destroy(context.Background())
	}()

	h, err := handler.NewHandler(1024, config.Uploads{
		Dir:    os.TempDir(),
		Forbid: []string{},
	}, nil, p)
	assert.NoError(t, err)

	hs := &http.Server{Addr: ":8079", Handler: h}
	defer func() {
		errS := hs.Shutdown(context.Background())
		if errS != nil {
			t.Errorf("error during the shutdown: error %v", err)
		}
	}()

	go func() {
		err = hs.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			t.Errorf("error listening the interface: error %v", err)
		}
	}()
	time.Sleep(time.Millisecond * 10)

	req, err := http.NewRequest("GET", "http://localhost:8079", nil)
	assert.NoError(t, err)

	req.AddCookie(&http.Cookie{Name: "input", Value: "input-value"})

	r, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer func() {
		err = r.Body.Close()
		if err != nil {
			t.Errorf("error during the closing Body: error %v", err)
		}
	}()

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
	p, err := pool.Initialize(context.Background(),
		func() *exec.Cmd { return exec.Command("php", "../../http/client.php", "payload", "pipes") },
		pipe.NewPipeFactory(),
		pool.Config{
			NumWorkers:      1,
			AllocateTimeout: time.Second * 1000,
			DestroyTimeout:  time.Second * 1000,
		})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		p.Destroy(context.Background())
	}()

	h, err := handler.NewHandler(1024, config.Uploads{
		Dir:    os.TempDir(),
		Forbid: []string{},
	}, nil, p)
	assert.NoError(t, err)

	hs := &http.Server{Addr: ":8090", Handler: h}
	defer func() {
		errS := hs.Shutdown(context.Background())
		if errS != nil {
			t.Errorf("error during the shutdown: error %v", err)
		}
	}()

	go func() {
		err = hs.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			t.Errorf("error listening the interface: error %v", err)
		}
	}()
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
	defer func() {
		err = r.Body.Close()
		if err != nil {
			t.Errorf("error during the closing Body: error %v", err)
		}
	}()

	b, err := ioutil.ReadAll(r.Body)
	assert.NoError(t, err)

	assert.NoError(t, err)
	assert.Equal(t, 200, r.StatusCode)
	assert.Equal(t, `{"value":"key"}`, string(b))
}

func TestHandler_JsonPayload_PUT(t *testing.T) {
	p, err := pool.Initialize(context.Background(),
		func() *exec.Cmd { return exec.Command("php", "../../http/client.php", "payload", "pipes") },
		pipe.NewPipeFactory(),
		pool.Config{
			NumWorkers:      1,
			AllocateTimeout: time.Second * 1000,
			DestroyTimeout:  time.Second * 1000,
		})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		p.Destroy(context.Background())
	}()

	h, err := handler.NewHandler(1024, config.Uploads{
		Dir:    os.TempDir(),
		Forbid: []string{},
	}, nil, p)
	assert.NoError(t, err)

	hs := &http.Server{Addr: ":8081", Handler: h}
	defer func() {
		errS := hs.Shutdown(context.Background())
		if errS != nil {
			t.Errorf("error during the shutdown: error %v", err)
		}
	}()

	go func() {
		err = hs.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			t.Errorf("error listening the interface: error %v", err)
		}
	}()
	time.Sleep(time.Millisecond * 10)

	req, err := http.NewRequest("PUT", "http://localhost"+hs.Addr, bytes.NewBufferString(`{"key":"value"}`))
	assert.NoError(t, err)

	req.Header.Add("Content-Type", "application/json")

	r, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer func() {
		err = r.Body.Close()
		if err != nil {
			t.Errorf("error during the closing Body: error %v", err)
		}
	}()

	b, err := ioutil.ReadAll(r.Body)
	assert.NoError(t, err)

	assert.NoError(t, err)
	assert.Equal(t, 200, r.StatusCode)
	assert.Equal(t, `{"value":"key"}`, string(b))
}

func TestHandler_JsonPayload_PATCH(t *testing.T) {
	p, err := pool.Initialize(context.Background(),
		func() *exec.Cmd { return exec.Command("php", "../../http/client.php", "payload", "pipes") },
		pipe.NewPipeFactory(),
		pool.Config{
			NumWorkers:      1,
			AllocateTimeout: time.Second * 1000,
			DestroyTimeout:  time.Second * 1000,
		})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		p.Destroy(context.Background())
	}()

	h, err := handler.NewHandler(1024, config.Uploads{
		Dir:    os.TempDir(),
		Forbid: []string{},
	}, nil, p)
	assert.NoError(t, err)

	hs := &http.Server{Addr: ":8082", Handler: h}
	defer func() {
		errS := hs.Shutdown(context.Background())
		if errS != nil {
			t.Errorf("error during the shutdown: error %v", err)
		}
	}()

	go func() {
		err = hs.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			t.Errorf("error listening the interface: error %v", err)
		}
	}()
	time.Sleep(time.Millisecond * 10)

	req, err := http.NewRequest("PATCH", "http://localhost"+hs.Addr, bytes.NewBufferString(`{"key":"value"}`))
	assert.NoError(t, err)

	req.Header.Add("Content-Type", "application/json")

	r, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer func() {
		err = r.Body.Close()
		if err != nil {
			t.Errorf("error during the closing Body: error %v", err)
		}
	}()

	b, err := ioutil.ReadAll(r.Body)
	assert.NoError(t, err)

	assert.NoError(t, err)
	assert.Equal(t, 200, r.StatusCode)
	assert.Equal(t, `{"value":"key"}`, string(b))
}

func TestHandler_FormData_POST(t *testing.T) {
	p, err := pool.Initialize(context.Background(),
		func() *exec.Cmd { return exec.Command("php", "../../http/client.php", "data", "pipes") },
		pipe.NewPipeFactory(),
		pool.Config{
			NumWorkers:      1,
			AllocateTimeout: time.Second * 1000,
			DestroyTimeout:  time.Second * 1000,
		})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		p.Destroy(context.Background())
	}()

	h, err := handler.NewHandler(1024, config.Uploads{
		Dir:    os.TempDir(),
		Forbid: []string{},
	}, nil, p)
	assert.NoError(t, err)

	hs := &http.Server{Addr: ":8083", Handler: h}
	defer func() {
		errS := hs.Shutdown(context.Background())
		if errS != nil {
			t.Errorf("error during the shutdown: error %v", err)
		}
	}()

	go func() {
		err = hs.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			t.Errorf("error listening the interface: error %v", err)
		}
	}()
	time.Sleep(time.Millisecond * 500)

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
	defer func() {
		err = r.Body.Close()
		if err != nil {
			t.Errorf("error during the closing Body: error %v", err)
		}
	}()

	b, err := ioutil.ReadAll(r.Body)
	assert.NoError(t, err)

	assert.NoError(t, err)
	assert.Equal(t, 200, r.StatusCode)

	// Sorted
	assert.Equal(t, "{\"arr\":{\"c\":{\"p\":\"l\",\"z\":\"\"},\"x\":{\"y\":{\"e\":\"f\",\"z\":\"y\"}}},\"key\":\"value\",\"name\":[\"name1\",\"name2\",\"name3\"]}", string(b))
}

func TestHandler_FormData_POST_Overwrite(t *testing.T) {
	p, err := pool.Initialize(context.Background(),
		func() *exec.Cmd { return exec.Command("php", "../../http/client.php", "data", "pipes") },
		pipe.NewPipeFactory(),
		pool.Config{
			NumWorkers:      1,
			AllocateTimeout: time.Second * 1000,
			DestroyTimeout:  time.Second * 1000,
		})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		p.Destroy(context.Background())
	}()

	h, err := handler.NewHandler(1024, config.Uploads{
		Dir:    os.TempDir(),
		Forbid: []string{},
	}, nil, p)
	assert.NoError(t, err)

	hs := &http.Server{Addr: ":8083", Handler: h}
	defer func() {
		errS := hs.Shutdown(context.Background())
		if errS != nil {
			t.Errorf("error during the shutdown: error %v", err)
		}
	}()

	go func() {
		err = hs.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			t.Errorf("error listening the interface: error %v", err)
		}
	}()
	time.Sleep(time.Millisecond * 10)

	form := url.Values{}

	form.Add("key", "value")
	form.Add("key", "value2")
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
	defer func() {
		err = r.Body.Close()
		if err != nil {
			t.Errorf("error during the closing Body: error %v", err)
		}
	}()

	b, err := ioutil.ReadAll(r.Body)
	assert.NoError(t, err)

	assert.NoError(t, err)
	assert.Equal(t, 200, r.StatusCode)

	assert.Equal(t, `{"arr":{"c":{"p":"l","z":""},"x":{"y":{"e":"f","z":"y"}}},"key":"value2","name":["name1","name2","name3"]}`, string(b))
}

func TestHandler_FormData_POST_Form_UrlEncoded_Charset(t *testing.T) {
	p, err := pool.Initialize(context.Background(),
		func() *exec.Cmd { return exec.Command("php", "../../http/client.php", "data", "pipes") },
		pipe.NewPipeFactory(),
		pool.Config{
			NumWorkers:      1,
			AllocateTimeout: time.Second * 1000,
			DestroyTimeout:  time.Second * 1000,
		})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		p.Destroy(context.Background())
	}()

	h, err := handler.NewHandler(1024, config.Uploads{
		Dir:    os.TempDir(),
		Forbid: []string{},
	}, nil, p)
	assert.NoError(t, err)

	hs := &http.Server{Addr: ":8083", Handler: h}
	defer func() {
		errS := hs.Shutdown(context.Background())
		if errS != nil {
			t.Errorf("error during the shutdown: error %v", err)
		}
	}()

	go func() {
		err = hs.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			t.Errorf("error listening the interface: error %v", err)
		}
	}()
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
	defer func() {
		err = r.Body.Close()
		if err != nil {
			t.Errorf("error during the closing Body: error %v", err)
		}
	}()

	b, err := ioutil.ReadAll(r.Body)
	assert.NoError(t, err)

	assert.NoError(t, err)
	assert.Equal(t, 200, r.StatusCode)

	assert.Equal(t, `{"arr":{"c":{"p":"l","z":""},"x":{"y":{"e":"f","z":"y"}}},"key":"value","name":["name1","name2","name3"]}`, string(b))
}

func TestHandler_FormData_PUT(t *testing.T) {
	p, err := pool.Initialize(context.Background(),
		func() *exec.Cmd { return exec.Command("php", "../../http/client.php", "data", "pipes") },
		pipe.NewPipeFactory(),
		pool.Config{
			NumWorkers:      1,
			AllocateTimeout: time.Second * 1000,
			DestroyTimeout:  time.Second * 1000,
		})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		p.Destroy(context.Background())
	}()

	h, err := handler.NewHandler(1024, config.Uploads{
		Dir:    os.TempDir(),
		Forbid: []string{},
	}, nil, p)
	assert.NoError(t, err)

	hs := &http.Server{Addr: ":17834", Handler: h}
	defer func() {
		errS := hs.Shutdown(context.Background())
		if errS != nil {
			t.Errorf("error during the shutdown: error %v", err)
		}
	}()

	go func() {
		err = hs.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			t.Errorf("error listening the interface: error %v", err)
		}
	}()
	time.Sleep(time.Millisecond * 500)

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
	defer func() {
		err = r.Body.Close()
		if err != nil {
			t.Errorf("error during the closing Body: error %v", err)
		}
	}()

	b, err := ioutil.ReadAll(r.Body)
	assert.NoError(t, err)

	assert.NoError(t, err)
	assert.Equal(t, 200, r.StatusCode)

	assert.Equal(t, `{"arr":{"c":{"p":"l","z":""},"x":{"y":{"e":"f","z":"y"}}},"key":"value","name":["name1","name2","name3"]}`, string(b))
}

func TestHandler_FormData_PATCH(t *testing.T) {
	p, err := pool.Initialize(context.Background(),
		func() *exec.Cmd { return exec.Command("php", "../../http/client.php", "data", "pipes") },
		pipe.NewPipeFactory(),
		pool.Config{
			NumWorkers:      1,
			AllocateTimeout: time.Second * 1000,
			DestroyTimeout:  time.Second * 1000,
		})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		p.Destroy(context.Background())
	}()

	h, err := handler.NewHandler(1024, config.Uploads{
		Dir:    os.TempDir(),
		Forbid: []string{},
	}, nil, p)
	assert.NoError(t, err)

	hs := &http.Server{Addr: ":8085", Handler: h}
	defer func() {
		errS := hs.Shutdown(context.Background())
		if errS != nil {
			t.Errorf("error during the shutdown: error %v", err)
		}
	}()

	go func() {
		err = hs.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			t.Errorf("error listening the interface: error %v", err)
		}
	}()
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
	defer func() {
		err = r.Body.Close()
		if err != nil {
			t.Errorf("error during the closing Body: error %v", err)
		}
	}()

	b, err := ioutil.ReadAll(r.Body)
	assert.NoError(t, err)

	assert.NoError(t, err)
	assert.Equal(t, 200, r.StatusCode)

	assert.Equal(t, "{\"arr\":{\"c\":{\"p\":\"l\",\"z\":\"\"},\"x\":{\"y\":{\"e\":\"f\",\"z\":\"y\"}}},\"key\":\"value\",\"name\":[\"name1\",\"name2\",\"name3\"]}", string(b))
}

func TestHandler_Multipart_POST(t *testing.T) {
	p, err := pool.Initialize(context.Background(),
		func() *exec.Cmd { return exec.Command("php", "../../http/client.php", "data", "pipes") },
		pipe.NewPipeFactory(),
		pool.Config{
			NumWorkers:      1,
			AllocateTimeout: time.Second * 1000,
			DestroyTimeout:  time.Second * 1000,
		})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		p.Destroy(context.Background())
	}()

	h, err := handler.NewHandler(1024, config.Uploads{
		Dir:    os.TempDir(),
		Forbid: []string{},
	}, nil, p)
	assert.NoError(t, err)

	hs := &http.Server{Addr: ":8019", Handler: h}
	defer func() {
		errS := hs.Shutdown(context.Background())
		if errS != nil {
			t.Errorf("error during the shutdown: error %v", err)
		}
	}()

	go func() {
		err = hs.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			t.Errorf("error listening the interface: error %v", err)
		}
	}()
	time.Sleep(time.Millisecond * 10)

	var mb bytes.Buffer
	w := multipart.NewWriter(&mb)
	err = w.WriteField("key", "value")
	if err != nil {
		t.Errorf("error writing the field: error %v", err)
	}

	err = w.WriteField("key", "value")
	if err != nil {
		t.Errorf("error writing the field: error %v", err)
	}

	err = w.WriteField("name[]", "name1")
	if err != nil {
		t.Errorf("error writing the field: error %v", err)
	}

	err = w.WriteField("name[]", "name2")
	if err != nil {
		t.Errorf("error writing the field: error %v", err)
	}

	err = w.WriteField("name[]", "name3")
	if err != nil {
		t.Errorf("error writing the field: error %v", err)
	}

	err = w.WriteField("arr[x][y][z]", "y")
	if err != nil {
		t.Errorf("error writing the field: error %v", err)
	}

	err = w.WriteField("arr[x][y][e]", "f")

	if err != nil {
		t.Errorf("error writing the field: error %v", err)
	}

	err = w.WriteField("arr[c]p", "l")
	if err != nil {
		t.Errorf("error writing the field: error %v", err)
	}

	err = w.WriteField("arr[c]z", "")
	if err != nil {
		t.Errorf("error writing the field: error %v", err)
	}

	err = w.Close()
	if err != nil {
		t.Errorf("error closing the writer: error %v", err)
	}

	req, err := http.NewRequest("POST", "http://localhost"+hs.Addr, &mb)
	assert.NoError(t, err)

	req.Header.Set("Content-Type", w.FormDataContentType())

	r, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer func() {
		err = r.Body.Close()
		if err != nil {
			t.Errorf("error during the closing Body: error %v", err)
		}
	}()

	b, err := ioutil.ReadAll(r.Body)
	assert.NoError(t, err)

	assert.NoError(t, err)
	assert.Equal(t, 200, r.StatusCode)

	assert.Equal(t, "{\"arr\":{\"c\":{\"p\":\"l\",\"z\":\"\"},\"x\":{\"y\":{\"e\":\"f\",\"z\":\"y\"}}},\"key\":\"value\",\"name\":[\"name1\",\"name2\",\"name3\"]}", string(b))
}

func TestHandler_Multipart_PUT(t *testing.T) {
	p, err := pool.Initialize(context.Background(),
		func() *exec.Cmd { return exec.Command("php", "../../http/client.php", "data", "pipes") },
		pipe.NewPipeFactory(),
		pool.Config{
			NumWorkers:      1,
			AllocateTimeout: time.Second * 1000,
			DestroyTimeout:  time.Second * 1000,
		})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		p.Destroy(context.Background())
	}()

	h, err := handler.NewHandler(1024, config.Uploads{
		Dir:    os.TempDir(),
		Forbid: []string{},
	}, nil, p)
	assert.NoError(t, err)

	hs := &http.Server{Addr: ":8020", Handler: h}
	defer func() {
		errS := hs.Shutdown(context.Background())
		if errS != nil {
			t.Errorf("error during the shutdown: error %v", err)
		}
	}()

	go func() {
		err = hs.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			t.Errorf("error listening the interface: error %v", err)
		}
	}()
	time.Sleep(time.Millisecond * 500)

	var mb bytes.Buffer
	w := multipart.NewWriter(&mb)
	err = w.WriteField("key", "value")
	if err != nil {
		t.Errorf("error writing the field: error %v", err)
	}

	err = w.WriteField("key", "value")
	if err != nil {
		t.Errorf("error writing the field: error %v", err)
	}

	err = w.WriteField("name[]", "name1")

	if err != nil {
		t.Errorf("error writing the field: error %v", err)
	}

	err = w.WriteField("name[]", "name2")
	if err != nil {
		t.Errorf("error writing the field: error %v", err)
	}

	err = w.WriteField("name[]", "name3")
	if err != nil {
		t.Errorf("error writing the field: error %v", err)
	}

	err = w.WriteField("arr[x][y][z]", "y")
	if err != nil {
		t.Errorf("error writing the field: error %v", err)
	}

	err = w.WriteField("arr[x][y][e]", "f")
	if err != nil {
		t.Errorf("error writing the field: error %v", err)
	}

	err = w.WriteField("arr[c]p", "l")
	if err != nil {
		t.Errorf("error writing the field: error %v", err)
	}

	err = w.WriteField("arr[c]z", "")
	if err != nil {
		t.Errorf("error writing the field: error %v", err)
	}

	err = w.Close()
	if err != nil {
		t.Errorf("error closing the writer: error %v", err)
	}

	req, err := http.NewRequest("PUT", "http://localhost"+hs.Addr, &mb)
	assert.NoError(t, err)

	req.Header.Set("Content-Type", w.FormDataContentType())

	r, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer func() {
		err = r.Body.Close()
		if err != nil {
			t.Errorf("error during the closing Body: error %v", err)
		}
	}()

	b, err := ioutil.ReadAll(r.Body)
	assert.NoError(t, err)

	assert.NoError(t, err)
	assert.Equal(t, 200, r.StatusCode)

	assert.Equal(t, `{"arr":{"c":{"p":"l","z":""},"x":{"y":{"e":"f","z":"y"}}},"key":"value","name":["name1","name2","name3"]}`, string(b))
}

func TestHandler_Multipart_PATCH(t *testing.T) {
	p, err := pool.Initialize(context.Background(),
		func() *exec.Cmd { return exec.Command("php", "../../http/client.php", "data", "pipes") },
		pipe.NewPipeFactory(),
		pool.Config{
			NumWorkers:      1,
			AllocateTimeout: time.Second * 1000,
			DestroyTimeout:  time.Second * 1000,
		})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		p.Destroy(context.Background())
	}()

	h, err := handler.NewHandler(1024, config.Uploads{
		Dir:    os.TempDir(),
		Forbid: []string{},
	}, nil, p)
	assert.NoError(t, err)

	hs := &http.Server{Addr: ":8021", Handler: h}
	defer func() {
		errS := hs.Shutdown(context.Background())
		if errS != nil {
			t.Errorf("error during the shutdown: error %v", err)
		}
	}()

	go func() {
		err = hs.ListenAndServe()

		if err != nil && err != http.ErrServerClosed {
			t.Errorf("error listening the interface: error %v", err)
		}
	}()
	time.Sleep(time.Millisecond * 500)

	var mb bytes.Buffer
	w := multipart.NewWriter(&mb)
	err = w.WriteField("key", "value")
	if err != nil {
		t.Errorf("error writing the field: error %v", err)
	}

	err = w.WriteField("key", "value")
	if err != nil {
		t.Errorf("error writing the field: error %v", err)
	}

	err = w.WriteField("name[]", "name1")
	if err != nil {
		t.Errorf("error writing the field: error %v", err)
	}

	err = w.WriteField("name[]", "name2")

	if err != nil {
		t.Errorf("error writing the field: error %v", err)
	}

	err = w.WriteField("name[]", "name3")

	if err != nil {
		t.Errorf("error writing the field: error %v", err)
	}

	err = w.WriteField("arr[x][y][z]", "y")
	if err != nil {
		t.Errorf("error writing the field: error %v", err)
	}

	err = w.WriteField("arr[x][y][e]", "f")
	if err != nil {
		t.Errorf("error writing the field: error %v", err)
	}

	err = w.WriteField("arr[c]p", "l")
	if err != nil {
		t.Errorf("error writing the field: error %v", err)
	}

	err = w.WriteField("arr[c]z", "")
	if err != nil {
		t.Errorf("error writing the field: error %v", err)
	}

	err = w.Close()
	if err != nil {
		t.Errorf("error closing the writer: error %v", err)
	}

	req, err := http.NewRequest("PATCH", "http://localhost"+hs.Addr, &mb)
	assert.NoError(t, err)

	req.Header.Set("Content-Type", w.FormDataContentType())

	r, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer func() {
		err = r.Body.Close()
		if err != nil {
			t.Errorf("error during the closing Body: error %v", err)
		}
	}()

	b, err := ioutil.ReadAll(r.Body)
	assert.NoError(t, err)

	assert.NoError(t, err)
	assert.Equal(t, 200, r.StatusCode)

	assert.Equal(t, `{"arr":{"c":{"p":"l","z":""},"x":{"y":{"e":"f","z":"y"}}},"key":"value","name":["name1","name2","name3"]}`, string(b))
}

func TestHandler_Error(t *testing.T) {
	p, err := pool.Initialize(context.Background(),
		func() *exec.Cmd { return exec.Command("php", "../../http/client.php", "error", "pipes") },
		pipe.NewPipeFactory(),
		pool.Config{
			NumWorkers:      1,
			AllocateTimeout: time.Second * 1000,
			DestroyTimeout:  time.Second * 1000,
		})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		p.Destroy(context.Background())
	}()

	h, err := handler.NewHandler(1024, config.Uploads{
		Dir:    os.TempDir(),
		Forbid: []string{},
	}, nil, p)
	assert.NoError(t, err)

	hs := &http.Server{Addr: ":8177", Handler: h}
	defer func() {
		errS := hs.Shutdown(context.Background())
		if errS != nil {
			t.Errorf("error during the shutdown: error %v", err)
		}
	}()

	go func() {
		err = hs.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			t.Errorf("error listening the interface: error %v", err)
		}
	}()
	time.Sleep(time.Millisecond * 10)

	_, r, err := get("http://localhost:8177/?hello=world")
	assert.NoError(t, err)
	defer func() {
		_ = r.Body.Close()
	}()
	assert.Equal(t, 500, r.StatusCode)
}

func TestHandler_Error2(t *testing.T) {
	p, err := pool.Initialize(context.Background(),
		func() *exec.Cmd { return exec.Command("php", "../../http/client.php", "error2", "pipes") },
		pipe.NewPipeFactory(),
		pool.Config{
			NumWorkers:      1,
			AllocateTimeout: time.Second * 1000,
			DestroyTimeout:  time.Second * 1000,
		})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		p.Destroy(context.Background())
	}()

	h, err := handler.NewHandler(1024, config.Uploads{
		Dir:    os.TempDir(),
		Forbid: []string{},
	}, nil, p)
	assert.NoError(t, err)

	hs := &http.Server{Addr: ":8177", Handler: h}
	defer func() {
		errS := hs.Shutdown(context.Background())
		if errS != nil {
			t.Errorf("error during the shutdown: error %v", err)
		}
	}()

	go func() {
		err = hs.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			t.Errorf("error listening the interface: error %v", err)
		}
	}()
	time.Sleep(time.Millisecond * 10)

	_, r, err := get("http://localhost:8177/?hello=world")
	assert.NoError(t, err)
	defer func() {
		_ = r.Body.Close()
	}()
	assert.Equal(t, 500, r.StatusCode)
}

func TestHandler_Error3(t *testing.T) {
	p, err := pool.Initialize(context.Background(),
		func() *exec.Cmd { return exec.Command("php", "../../http/client.php", "pid", "pipes") },
		pipe.NewPipeFactory(),
		pool.Config{
			NumWorkers:      1,
			AllocateTimeout: time.Second * 1000,
			DestroyTimeout:  time.Second * 1000,
		})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		p.Destroy(context.Background())
	}()

	h, err := handler.NewHandler(1, config.Uploads{
		Dir:    os.TempDir(),
		Forbid: []string{},
	}, nil, p)
	assert.NoError(t, err)

	hs := &http.Server{Addr: ":8177", Handler: h}
	defer func() {
		errS := hs.Shutdown(context.Background())
		if errS != nil {
			t.Errorf("error during the shutdown: error %v", err)
		}
	}()

	go func() {
		err = hs.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			t.Errorf("error listening the interface: error %v", err)
		}
	}()
	time.Sleep(time.Millisecond * 10)

	b2 := &bytes.Buffer{}
	for i := 0; i < 1024*1024; i++ {
		b2.Write([]byte("  "))
	}

	req, err := http.NewRequest("POST", "http://localhost"+hs.Addr, b2)
	assert.NoError(t, err)

	r, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer func() {
		err = r.Body.Close()
		if err != nil {
			t.Errorf("error during the closing Body: error %v", err)
		}
	}()

	assert.NoError(t, err)
	assert.Equal(t, 500, r.StatusCode)
}

func TestHandler_ResponseDuration(t *testing.T) {
	p, err := pool.Initialize(context.Background(),
		func() *exec.Cmd { return exec.Command("php", "../../http/client.php", "echo", "pipes") },
		pipe.NewPipeFactory(),
		pool.Config{
			NumWorkers:      1,
			AllocateTimeout: time.Second * 1000,
			DestroyTimeout:  time.Second * 1000,
		})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		p.Destroy(context.Background())
	}()

	h, err := handler.NewHandler(1024, config.Uploads{
		Dir:    os.TempDir(),
		Forbid: []string{},
	}, nil, p)
	assert.NoError(t, err)

	hs := &http.Server{Addr: ":8177", Handler: h}
	defer func() {
		errS := hs.Shutdown(context.Background())
		if errS != nil {
			t.Errorf("error during the shutdown: error %v", err)
		}
	}()

	go func() {
		err = hs.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			t.Errorf("error listening the interface: error %v", err)
		}
	}()
	time.Sleep(time.Millisecond * 10)

	gotresp := make(chan interface{})
	h.AddListener(func(event interface{}) {
		switch t := event.(type) {
		case handler.ResponseEvent:
			if t.Elapsed() > 0 {
				close(gotresp)
			}
		default:
		}
	})

	body, r, err := get("http://localhost:8177/?hello=world")
	assert.NoError(t, err)
	defer func() {
		_ = r.Body.Close()
	}()

	<-gotresp

	assert.Equal(t, 201, r.StatusCode)
	assert.Equal(t, "WORLD", body)
}

func TestHandler_ResponseDurationDelayed(t *testing.T) {
	p, err := pool.Initialize(context.Background(),
		func() *exec.Cmd { return exec.Command("php", "../../http/client.php", "echoDelay", "pipes") },
		pipe.NewPipeFactory(),
		pool.Config{
			NumWorkers:      1,
			AllocateTimeout: time.Second * 1000,
			DestroyTimeout:  time.Second * 1000,
		})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		p.Destroy(context.Background())
	}()

	h, err := handler.NewHandler(1024, config.Uploads{
		Dir:    os.TempDir(),
		Forbid: []string{},
	}, nil, p)
	assert.NoError(t, err)

	hs := &http.Server{Addr: ":8177", Handler: h}
	defer func() {
		errS := hs.Shutdown(context.Background())
		if errS != nil {
			t.Errorf("error during the shutdown: error %v", err)
		}
	}()

	go func() {
		err = hs.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			t.Errorf("error listening the interface: error %v", err)
		}
	}()
	time.Sleep(time.Millisecond * 10)

	gotresp := make(chan interface{})
	h.AddListener(func(event interface{}) {
		switch tp := event.(type) {
		case handler.ResponseEvent:
			if tp.Elapsed() > time.Second {
				close(gotresp)
			}
		default:
		}
	})

	body, r, err := get("http://localhost:8177/?hello=world")
	assert.NoError(t, err)
	defer func() {
		_ = r.Body.Close()
	}()
	<-gotresp

	assert.Equal(t, 201, r.StatusCode)
	assert.Equal(t, "WORLD", body)
}

func TestHandler_ErrorDuration(t *testing.T) {
	p, err := pool.Initialize(context.Background(),
		func() *exec.Cmd { return exec.Command("php", "../../http/client.php", "error", "pipes") },
		pipe.NewPipeFactory(),
		pool.Config{
			NumWorkers:      1,
			AllocateTimeout: time.Second * 1000,
			DestroyTimeout:  time.Second * 1000,
		})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		p.Destroy(context.Background())
	}()

	h, err := handler.NewHandler(1024, config.Uploads{
		Dir:    os.TempDir(),
		Forbid: []string{},
	}, nil, p)
	assert.NoError(t, err)

	hs := &http.Server{Addr: ":8177", Handler: h}
	defer func() {
		errS := hs.Shutdown(context.Background())
		if errS != nil {
			t.Errorf("error during the shutdown: error %v", err)
		}
	}()

	go func() {
		err = hs.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			t.Errorf("error listening the interface: error %v", err)
		}
	}()
	time.Sleep(time.Millisecond * 10)

	goterr := make(chan interface{})
	h.AddListener(func(event interface{}) {
		switch tp := event.(type) {
		case handler.ErrorEvent:
			if tp.Elapsed() > 0 {
				close(goterr)
			}
		default:
		}
	})

	_, r, err := get("http://localhost:8177/?hello=world")
	assert.NoError(t, err)
	defer func() {
		_ = r.Body.Close()
	}()

	<-goterr

	assert.Equal(t, 500, r.StatusCode)
}

func TestHandler_IP(t *testing.T) {
	trusted := []string{
		"10.0.0.0/8",
		"127.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"::1/128",
		"fc00::/7",
		"fe80::/10",
	}

	cidrs, err := config.ParseCIDRs(trusted)
	assert.NoError(t, err)
	assert.NotNil(t, cidrs)

	p, err := pool.Initialize(context.Background(),
		func() *exec.Cmd { return exec.Command("php", "../../http/client.php", "ip", "pipes") },
		pipe.NewPipeFactory(),
		pool.Config{
			NumWorkers:      1,
			AllocateTimeout: time.Second * 1000,
			DestroyTimeout:  time.Second * 1000,
		})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		p.Destroy(context.Background())
	}()

	h, err := handler.NewHandler(1024, config.Uploads{
		Dir:    os.TempDir(),
		Forbid: []string{},
	}, cidrs, p)
	assert.NoError(t, err)

	hs := &http.Server{Addr: "127.0.0.1:8177", Handler: h}
	defer func() {
		errS := hs.Shutdown(context.Background())
		if errS != nil {
			t.Errorf("error during the shutdown: error %v", err)
		}
	}()

	go func() {
		err = hs.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			t.Errorf("error listening the interface: error %v", err)
		}
	}()
	time.Sleep(time.Millisecond * 10)

	body, r, err := get("http://127.0.0.1:8177/")
	assert.NoError(t, err)
	defer func() {
		_ = r.Body.Close()
	}()
	assert.Equal(t, 200, r.StatusCode)
	assert.Equal(t, "127.0.0.1", body)
}

func TestHandler_XRealIP(t *testing.T) {
	trusted := []string{
		"10.0.0.0/8",
		"127.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"::1/128",
		"fc00::/7",
		"fe80::/10",
	}

	cidrs, err := config.ParseCIDRs(trusted)
	assert.NoError(t, err)
	assert.NotNil(t, cidrs)

	p, err := pool.Initialize(context.Background(),
		func() *exec.Cmd { return exec.Command("php", "../../http/client.php", "ip", "pipes") },
		pipe.NewPipeFactory(),
		pool.Config{
			NumWorkers:      1,
			AllocateTimeout: time.Second * 1000,
			DestroyTimeout:  time.Second * 1000,
		})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		p.Destroy(context.Background())
	}()

	h, err := handler.NewHandler(1024, config.Uploads{
		Dir:    os.TempDir(),
		Forbid: []string{},
	}, cidrs, p)
	assert.NoError(t, err)

	hs := &http.Server{Addr: "127.0.0.1:8179", Handler: h}
	defer func() {
		errS := hs.Shutdown(context.Background())
		if errS != nil {
			t.Errorf("error during the shutdown: error %v", err)
		}
	}()

	go func() {
		err = hs.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			t.Errorf("error listening the interface: error %v", err)
		}
	}()
	time.Sleep(time.Millisecond * 10)

	body, r, err := getHeader("http://127.0.0.1:8179/", map[string]string{
		"X-Real-Ip": "200.0.0.1",
	})

	assert.NoError(t, err)
	defer func() {
		_ = r.Body.Close()
	}()
	assert.Equal(t, 200, r.StatusCode)
	assert.Equal(t, "200.0.0.1", body)
}

func TestHandler_XForwardedFor(t *testing.T) {
	trusted := []string{
		"10.0.0.0/8",
		"127.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"100.0.0.0/16",
		"200.0.0.0/16",
		"::1/128",
		"fc00::/7",
		"fe80::/10",
	}

	cidrs, err := config.ParseCIDRs(trusted)
	assert.NoError(t, err)
	assert.NotNil(t, cidrs)

	p, err := pool.Initialize(context.Background(),
		func() *exec.Cmd { return exec.Command("php", "../../http/client.php", "ip", "pipes") },
		pipe.NewPipeFactory(),
		pool.Config{
			NumWorkers:      1,
			AllocateTimeout: time.Second * 1000,
			DestroyTimeout:  time.Second * 1000,
		})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		p.Destroy(context.Background())
	}()

	h, err := handler.NewHandler(1024, config.Uploads{
		Dir:    os.TempDir(),
		Forbid: []string{},
	}, cidrs, p)
	assert.NoError(t, err)

	hs := &http.Server{Addr: "127.0.0.1:8177", Handler: h}
	defer func() {
		errS := hs.Shutdown(context.Background())
		if errS != nil {
			t.Errorf("error during the shutdown: error %v", err)
		}
	}()

	go func() {
		err = hs.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			t.Errorf("error listening the interface: error %v", err)
		}
	}()
	time.Sleep(time.Millisecond * 10)

	body, r, err := getHeader("http://127.0.0.1:8177/", map[string]string{
		"X-Forwarded-For": "100.0.0.1, 200.0.0.1, invalid, 101.0.0.1",
	})

	assert.NoError(t, err)
	assert.Equal(t, 200, r.StatusCode)
	assert.Equal(t, "101.0.0.1", body)
	_ = r.Body.Close()

	body, r, err = getHeader("http://127.0.0.1:8177/", map[string]string{
		"X-Forwarded-For": "100.0.0.1, 200.0.0.1, 101.0.0.1, invalid",
	})

	assert.NoError(t, err)
	_ = r.Body.Close()
	assert.Equal(t, 200, r.StatusCode)
	assert.Equal(t, "101.0.0.1", body)
}

func TestHandler_XForwardedFor_NotTrustedRemoteIp(t *testing.T) {
	trusted := []string{
		"10.0.0.0/8",
	}

	cidrs, err := config.ParseCIDRs(trusted)
	assert.NoError(t, err)
	assert.NotNil(t, cidrs)

	p, err := pool.Initialize(context.Background(),
		func() *exec.Cmd { return exec.Command("php", "../../http/client.php", "ip", "pipes") },
		pipe.NewPipeFactory(),
		pool.Config{
			NumWorkers:      1,
			AllocateTimeout: time.Second * 1000,
			DestroyTimeout:  time.Second * 1000,
		})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		p.Destroy(context.Background())
	}()

	h, err := handler.NewHandler(1024, config.Uploads{
		Dir:    os.TempDir(),
		Forbid: []string{},
	}, cidrs, p)
	assert.NoError(t, err)

	hs := &http.Server{Addr: "127.0.0.1:8177", Handler: h}
	defer func() {
		errS := hs.Shutdown(context.Background())
		if errS != nil {
			t.Errorf("error during the shutdown: error %v", err)
		}
	}()

	go func() {
		err = hs.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			t.Errorf("error listening the interface: error %v", err)
		}
	}()
	time.Sleep(time.Millisecond * 10)

	body, r, err := getHeader("http://127.0.0.1:8177/", map[string]string{
		"X-Forwarded-For": "100.0.0.1, 200.0.0.1, invalid, 101.0.0.1",
	})

	assert.NoError(t, err)
	_ = r.Body.Close()
	assert.Equal(t, 200, r.StatusCode)
	assert.Equal(t, "127.0.0.1", body)
}

func BenchmarkHandler_Listen_Echo(b *testing.B) {
	p, err := pool.Initialize(context.Background(),
		func() *exec.Cmd { return exec.Command("php", "../../http/client.php", "echo", "pipes") },
		pipe.NewPipeFactory(),
		pool.Config{
			NumWorkers:      uint64(runtime.NumCPU()),
			AllocateTimeout: time.Second * 1000,
			DestroyTimeout:  time.Second * 1000,
		})
	if err != nil {
		b.Fatal(err)
	}
	defer func() {
		p.Destroy(context.Background())
	}()

	h, err := handler.NewHandler(1024, config.Uploads{
		Dir:    os.TempDir(),
		Forbid: []string{},
	}, nil, p)
	assert.NoError(b, err)

	hs := &http.Server{Addr: ":8177", Handler: h}
	defer func() {
		errS := hs.Shutdown(context.Background())
		if errS != nil {
			b.Errorf("error during the shutdown: error %v", err)
		}
	}()

	go func() {
		err = hs.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			b.Errorf("error listening the interface: error %v", err)
		}
	}()
	time.Sleep(time.Millisecond * 10)

	b.ResetTimer()
	b.ReportAllocs()
	bb := "WORLD"
	for n := 0; n < b.N; n++ {
		r, err := http.Get("http://localhost:8177/?hello=world")
		if err != nil {
			b.Fail()
		}
		// Response might be nil here
		if r != nil {
			br, err := ioutil.ReadAll(r.Body)
			if err != nil {
				b.Errorf("error reading Body: error %v", err)
			}
			if string(br) != bb {
				b.Fail()
			}
			err = r.Body.Close()
			if err != nil {
				b.Errorf("error closing the Body: error %v", err)
			}
		} else {
			b.Errorf("got nil response")
		}
	}
}
