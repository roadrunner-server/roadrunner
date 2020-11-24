package tests

import (
	"bytes"
	"crypto/tls"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/spiral/endure"
	"github.com/spiral/roadrunner/v2/plugins/config"
	httpPlugin "github.com/spiral/roadrunner/v2/plugins/http"
	"github.com/spiral/roadrunner/v2/plugins/logger"
	"github.com/yookoala/gofast"

	rpcPlugin "github.com/spiral/roadrunner/v2/plugins/rpc"
	"github.com/spiral/roadrunner/v2/plugins/server"
	"github.com/stretchr/testify/assert"
)

var sslClient = &http.Client{
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	},
}

func TestHTTPInit(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.DebugLevel), endure.Visualize(endure.StdOut, ""))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "configs/.rr-http.yaml",
		Prefix: "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&rpcPlugin.Plugin{},
		&logger.ZapLogger{},
		&server.Plugin{},
		&httpPlugin.Plugin{},
	)
	assert.NoError(t, err)

	err = cont.Init()
	if err != nil {
		t.Fatal(err)
	}

	ch, err := cont.Serve()
	assert.NoError(t, err)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		tt := time.NewTimer(time.Second * 2)
		defer wg.Done()
		for {
			select {
			case e := <-ch:
				assert.Fail(t, "error", e.Error.Error())
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
			case <-sig:
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			case <-tt.C:
				// timeout
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			}
		}
	}()

	wg.Wait()
}

func TestSSL(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.DebugLevel), endure.Visualize(endure.StdOut, ""))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "configs/.rr-ssl.yaml",
		Prefix: "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&rpcPlugin.Plugin{},
		&logger.ZapLogger{},
		&server.Plugin{},
		&httpPlugin.Plugin{},
	)
	assert.NoError(t, err)

	err = cont.Init()
	if err != nil {
		t.Fatal(err)
	}

	ch, err := cont.Serve()
	assert.NoError(t, err)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		tt := time.NewTimer(time.Second * 5)
		defer wg.Done()
		for {
			select {
			case e := <-ch:
				assert.Fail(t, "error", e.Error.Error())
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
			case <-sig:
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			case <-tt.C:
				// timeout
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			}
		}
	}()

	t.Run("SSLEcho", sslEcho)
	t.Run("SSLNoRedirect", sslNoRedirect)
	t.Run("fCGIecho", fcgiEcho)
	wg.Wait()
}

func sslNoRedirect(t *testing.T) {
	req, err := http.NewRequest("GET", "http://localhost:8085?hello=world", nil)
	assert.NoError(t, err)

	r, err := sslClient.Do(req)
	assert.NoError(t, err)

	assert.Nil(t, r.TLS)

	b, err := ioutil.ReadAll(r.Body)
	assert.NoError(t, err)

	assert.NoError(t, err)
	assert.Equal(t, 201, r.StatusCode)
	assert.Equal(t, "WORLD", string(b))

	err2 := r.Body.Close()
	if err2 != nil {
		t.Errorf("fail to close the Body: error %v", err2)
	}
}

func sslEcho(t *testing.T) {
	req, err := http.NewRequest("GET", "https://localhost:8893?hello=world", nil)
	assert.NoError(t, err)

	r, err := sslClient.Do(req)
	assert.NoError(t, err)

	b, err := ioutil.ReadAll(r.Body)
	assert.NoError(t, err)

	assert.NoError(t, err)
	assert.Equal(t, 201, r.StatusCode)
	assert.Equal(t, "WORLD", string(b))

	err2 := r.Body.Close()
	if err2 != nil {
		t.Errorf("fail to close the Body: error %v", err2)
	}
}

func fcgiEcho(t *testing.T) {
	fcgiConnFactory := gofast.SimpleConnFactory("tcp", "0.0.0.0:6920")

	fcgiHandler := gofast.NewHandler(
		gofast.BasicParamsMap(gofast.BasicSession),
		gofast.SimpleClientFactory(fcgiConnFactory, 0),
	)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "http://site.local/?hello=world", nil)
	fcgiHandler.ServeHTTP(w, req)

	body, err := ioutil.ReadAll(w.Result().Body)

	assert.NoError(t, err)
	assert.Equal(t, 201, w.Result().StatusCode)
	assert.Equal(t, "WORLD", string(body))
}

func TestSSLRedirect(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.DebugLevel), endure.Visualize(endure.StdOut, ""))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "configs/.rr-ssl-redirect.yaml",
		Prefix: "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&rpcPlugin.Plugin{},
		&logger.ZapLogger{},
		&server.Plugin{},
		&httpPlugin.Plugin{},
	)
	assert.NoError(t, err)

	err = cont.Init()
	if err != nil {
		t.Fatal(err)
	}

	ch, err := cont.Serve()
	assert.NoError(t, err)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		tt := time.NewTimer(time.Second * 5)
		defer wg.Done()
		for {
			select {
			case e := <-ch:
				assert.Fail(t, "error", e.Error.Error())
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
			case <-sig:
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			case <-tt.C:
				// timeout
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			}
		}
	}()

	t.Run("SSLRedirect", sslRedirect)
	wg.Wait()
}

func sslRedirect(t *testing.T) {
	req, err := http.NewRequest("GET", "http://localhost:8087?hello=world", nil)
	assert.NoError(t, err)

	r, err := sslClient.Do(req)
	assert.NoError(t, err)
	assert.NotNil(t, r.TLS)

	b, err := ioutil.ReadAll(r.Body)
	assert.NoError(t, err)

	assert.NoError(t, err)
	assert.Equal(t, 201, r.StatusCode)
	assert.Equal(t, "WORLD", string(b))

	err2 := r.Body.Close()
	if err2 != nil {
		t.Errorf("fail to close the Body: error %v", err2)
	}
}

func TestSSLPushPipes(t *testing.T) {
	time.Sleep(time.Second)
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.DebugLevel), endure.Visualize(endure.StdOut, ""))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "configs/.rr-ssl-push.yaml",
		Prefix: "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&rpcPlugin.Plugin{},
		&logger.ZapLogger{},
		&server.Plugin{},
		&httpPlugin.Plugin{},
	)
	assert.NoError(t, err)

	err = cont.Init()
	if err != nil {
		t.Fatal(err)
	}

	ch, err := cont.Serve()
	assert.NoError(t, err)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		tt := time.NewTimer(time.Second * 5)
		defer wg.Done()
		for {
			select {
			case e := <-ch:
				assert.Fail(t, "error", e.Error.Error())
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
			case <-sig:
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			case <-tt.C:
				// timeout
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			}
		}
	}()

	t.Run("SSLPush", sslPush)
	wg.Wait()
}

func sslPush(t *testing.T) {
	req, err := http.NewRequest("GET", "https://localhost:8894?hello=world", nil)
	assert.NoError(t, err)

	r, err := sslClient.Do(req)
	assert.NoError(t, err)

	assert.NotNil(t, r.TLS)

	b, err := ioutil.ReadAll(r.Body)
	assert.NoError(t, err)

	assert.Equal(t, "", r.Header.Get("Http2-Push"))

	assert.NoError(t, err)
	assert.Equal(t, 201, r.StatusCode)
	assert.Equal(t, "WORLD", string(b))

	err2 := r.Body.Close()
	if err2 != nil {
		t.Errorf("fail to close the Body: error %v", err2)
	}
}

func TestFastCGI_RequestUri(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.DebugLevel), endure.Visualize(endure.StdOut, ""))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "configs/.rr-fcgi-reqUri.yaml",
		Prefix: "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&rpcPlugin.Plugin{},
		&logger.ZapLogger{},
		&server.Plugin{},
		&httpPlugin.Plugin{},
	)
	assert.NoError(t, err)

	err = cont.Init()
	if err != nil {
		t.Fatal(err)
	}

	ch, err := cont.Serve()
	assert.NoError(t, err)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		tt := time.NewTimer(time.Second * 5)
		defer wg.Done()
		for {
			select {
			case e := <-ch:
				assert.Fail(t, "error", e.Error.Error())
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
			case <-sig:
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			case <-tt.C:
				// timeout
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			}
		}
	}()

	t.Run("FastCGIServiceRequestUri", fcgiReqUri)
	wg.Wait()
}

func fcgiReqUri(t *testing.T) {
	fcgiConnFactory := gofast.SimpleConnFactory("tcp", "0.0.0.0:6921")

	fcgiHandler := gofast.NewHandler(
		gofast.BasicParamsMap(gofast.BasicSession),
		gofast.SimpleClientFactory(fcgiConnFactory, 0),
	)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "http://site.local/hello-world", nil)
	fcgiHandler.ServeHTTP(w, req)

	body, err := ioutil.ReadAll(w.Result().Body)
	assert.NoError(t, err)
	assert.Equal(t, 200, w.Result().StatusCode)
	assert.Equal(t, "http://site.local/hello-world", string(body))
}

func TestH2CUpgrade(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.DebugLevel), endure.Visualize(endure.StdOut, ""))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "configs/.rr-h2c.yaml",
		Prefix: "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&rpcPlugin.Plugin{},
		&logger.ZapLogger{},
		&server.Plugin{},
		&httpPlugin.Plugin{},
	)
	assert.NoError(t, err)

	err = cont.Init()
	if err != nil {
		t.Fatal(err)
	}

	ch, err := cont.Serve()
	assert.NoError(t, err)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		tt := time.NewTimer(time.Second * 5)
		defer wg.Done()
		for {
			select {
			case e := <-ch:
				assert.Fail(t, "error", e.Error.Error())
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
			case <-sig:
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			case <-tt.C:
				// timeout
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			}
		}
	}()

	t.Run("H2cUpgrade", h2cUpgrade)
	wg.Wait()
}

func h2cUpgrade(t *testing.T) {
	req, err := http.NewRequest("PRI", "http://localhost:8083?hello=world", nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Add("Upgrade", "h2c")
	req.Header.Add("Connection", "HTTP2-Settings")
	req.Header.Add("HTTP2-Settings", "")

	r, err2 := http.DefaultClient.Do(req)
	if err2 != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "101 Switching Protocols", r.Status)

	err3 := r.Body.Close()
	if err3 != nil {
		t.Fatal(err)
	}
}

func TestH2C(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.DebugLevel), endure.Visualize(endure.StdOut, ""))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "configs/.rr-h2c.yaml",
		Prefix: "rr",
	}

	err = cont.RegisterAll(
		cfg,
		&rpcPlugin.Plugin{},
		&logger.ZapLogger{},
		&server.Plugin{},
		&httpPlugin.Plugin{},
	)
	assert.NoError(t, err)

	err = cont.Init()
	if err != nil {
		t.Fatal(err)
	}

	ch, err := cont.Serve()
	assert.NoError(t, err)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		tt := time.NewTimer(time.Second * 5)
		defer wg.Done()
		for {
			select {
			case e := <-ch:
				assert.Fail(t, "error", e.Error.Error())
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
			case <-sig:
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			case <-tt.C:
				// timeout
				err = cont.Stop()
				if err != nil {
					assert.FailNow(t, "error", err.Error())
				}
				return
			}
		}
	}()

	t.Run("H2c", h2c)
	wg.Wait()
}

func h2c(t *testing.T) {
	req, err := http.NewRequest("PRI", "http://localhost:8083?hello=world", nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Add("Connection", "HTTP2-Settings")
	req.Header.Add("HTTP2-Settings", "")

	r, err2 := http.DefaultClient.Do(req)
	if err2 != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "201 Created", r.Status)

	err3 := r.Body.Close()
	if err3 != nil {
		t.Fatal(err)
	}
}

func get(url string) (string, *http.Response, error) {
	r, err := http.Get(url)
	if err != nil {
		return "", nil, err
	}
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return "", nil, err
	}
	defer func() {
		_ = r.Body.Close()
	}()
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

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return "", nil, err
	}

	err = r.Body.Close()
	if err != nil {
		return "", nil, err
	}
	return string(b), r, err
}
