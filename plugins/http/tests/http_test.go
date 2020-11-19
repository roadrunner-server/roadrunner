package tests

import (
	"bytes"
	"crypto/tls"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/spiral/endure"
	"github.com/spiral/roadrunner/v2/plugins/config"
	httpPlugin "github.com/spiral/roadrunner/v2/plugins/http"
	"github.com/spiral/roadrunner/v2/plugins/logger"

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

	tt := time.NewTimer(time.Second * 1000)
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

	go func() {
		tt := time.NewTimer(time.Second * 1000)
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
}

func sslEcho(t *testing.T) {
	req, err := http.NewRequest("GET", "https://localhost:8888?hello=world", nil)
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

func TestFastCGI(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.DebugLevel), endure.Visualize(endure.StdOut, ""))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "configs/.rr-fcgi.yaml",
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

	go func() {
		tt := time.NewTimer(time.Second * 1000)
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

	t.Run("FastCGIEcho", fcgiEcho)
}

func fcgiEcho(t *testing.T) {
	conn, err := net.Dial("tcp", "0.0.0.0:6920")
	if err != nil {
		t.Fatal(err)
	}

	_, err = conn.Write([]byte("data"))
	if err != nil {
		t.Fatal(err)
	}

	//fcgiConnFactory := gofast.SimpleConnFactory("tcp", "0.0.0.0:6920")
	//
	//fcgiHandler := gofast.NewHandler(
	//	gofast.BasicParamsMap(gofast.BasicSession),
	//	gofast.SimpleClientFactory(fcgiConnFactory, 0),
	//)

	w := httptest.NewRecorder()
	//req := httptest.NewRequest("GET", "http://site.local/?hello=world", nil)
	//err = fcgi.Serve(, req)
	//if err != nil {
	//	t.Fatal(err)
	//}
	//fcgiHandler.ServeHTTP(w, req)

	body, err := ioutil.ReadAll(w.Result().Body)

	assert.NoError(t, err)
	assert.Equal(t, 201, w.Result().StatusCode)
	assert.Equal(t, "WORLD", string(body))
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
