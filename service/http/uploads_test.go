package http

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/spiral/roadrunner"
	"github.com/stretchr/testify/assert"
)

func TestHandler_Upload_File(t *testing.T) {
	h := &Handler{
		cfg: &Config{
			MaxRequestSize: 1024,
			Uploads: &UploadsConfig{
				Dir:    os.TempDir(),
				Forbid: []string{},
			},
		},
		rr: roadrunner.NewServer(&roadrunner.ServerConfig{
			Command: "php ../../tests/http/client.php upload pipes",
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
	defer func() {
		err := hs.Shutdown(context.Background())
		if err != nil {
			t.Errorf("error during the shutdown: error %v", err)
		}
	}()

	go func() {
		err := hs.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			t.Errorf("error listening the interface: error %v", err)
		}
	}()
	time.Sleep(time.Millisecond * 10)

	var mb bytes.Buffer
	w := multipart.NewWriter(&mb)

	f := mustOpen("uploads_test.go")
	defer func() {
		err := f.Close()
		if err != nil {
			t.Errorf("failed to close a file: error %v", err)
		}
	}()
	fw, err := w.CreateFormFile("upload", f.Name())
	assert.NotNil(t, fw)
	assert.NoError(t, err)
	_, err = io.Copy(fw, f)
	if err != nil {
		t.Errorf("error copying the file: error %v", err)
	}

	err = w.Close()
	if err != nil {
		t.Errorf("error closing the file: error %v", err)
	}

	req, err := http.NewRequest("POST", "http://localhost"+hs.Addr, &mb)
	assert.NoError(t, err)

	req.Header.Set("Content-Type", w.FormDataContentType())

	r, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer func() {
		err := r.Body.Close()
		if err != nil {
			t.Errorf("error closing the Body: error %v", err)
		}
	}()

	b, err := ioutil.ReadAll(r.Body)
	assert.NoError(t, err)

	assert.NoError(t, err)
	assert.Equal(t, 200, r.StatusCode)

	fs := fileString("uploads_test.go", 0, "application/octet-stream")

	assert.Equal(t, `{"upload":`+fs+`}`, string(b))
}

func TestHandler_Upload_NestedFile(t *testing.T) {
	h := &Handler{
		cfg: &Config{
			MaxRequestSize: 1024,
			Uploads: &UploadsConfig{
				Dir:    os.TempDir(),
				Forbid: []string{},
			},
		},
		rr: roadrunner.NewServer(&roadrunner.ServerConfig{
			Command: "php ../../tests/http/client.php upload pipes",
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
	defer func() {
		err := hs.Shutdown(context.Background())
		if err != nil {
			t.Errorf("error during the shutdown: error %v", err)
		}
	}()

	go func() {
		err := hs.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			t.Errorf("error listening the interface: error %v", err)
		}
	}()
	time.Sleep(time.Millisecond * 10)

	var mb bytes.Buffer
	w := multipart.NewWriter(&mb)

	f := mustOpen("uploads_test.go")
	defer func() {
		err := f.Close()
		if err != nil {
			t.Errorf("failed to close a file: error %v", err)
		}
	}()
	fw, err := w.CreateFormFile("upload[x][y][z][]", f.Name())
	assert.NotNil(t, fw)
	assert.NoError(t, err)
	_, err = io.Copy(fw, f)
	if err != nil {
		t.Errorf("error copying the file: error %v", err)
	}

	err = w.Close()
	if err != nil {
		t.Errorf("error closing the file: error %v", err)
	}

	req, err := http.NewRequest("POST", "http://localhost"+hs.Addr, &mb)
	assert.NoError(t, err)

	req.Header.Set("Content-Type", w.FormDataContentType())

	r, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer func() {
		err := r.Body.Close()
		if err != nil {
			t.Errorf("error closing the Body: error %v", err)
		}
	}()

	b, err := ioutil.ReadAll(r.Body)
	assert.NoError(t, err)

	assert.NoError(t, err)
	assert.Equal(t, 200, r.StatusCode)

	fs := fileString("uploads_test.go", 0, "application/octet-stream")

	assert.Equal(t, `{"upload":{"x":{"y":{"z":[`+fs+`]}}}}`, string(b))
}

func TestHandler_Upload_File_NoTmpDir(t *testing.T) {
	h := &Handler{
		cfg: &Config{
			MaxRequestSize: 1024,
			Uploads: &UploadsConfig{
				Dir:    "-----",
				Forbid: []string{},
			},
		},
		rr: roadrunner.NewServer(&roadrunner.ServerConfig{
			Command: "php ../../tests/http/client.php upload pipes",
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
	defer func() {
		err := hs.Shutdown(context.Background())
		if err != nil {
			t.Errorf("error during the shutdown: error %v", err)
		}
	}()

	go func() {
		err := hs.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			t.Errorf("error listening the interface: error %v", err)
		}
	}()
	time.Sleep(time.Millisecond * 10)

	var mb bytes.Buffer
	w := multipart.NewWriter(&mb)

	f := mustOpen("uploads_test.go")
	defer func() {
		err := f.Close()
		if err != nil {
			t.Errorf("failed to close a file: error %v", err)
		}
	}()
	fw, err := w.CreateFormFile("upload", f.Name())
	assert.NotNil(t, fw)
	assert.NoError(t, err)
	_, err = io.Copy(fw, f)
	if err != nil {
		t.Errorf("error copying the file: error %v", err)
	}

	err = w.Close()
	if err != nil {
		t.Errorf("error closing the file: error %v", err)
	}

	req, err := http.NewRequest("POST", "http://localhost"+hs.Addr, &mb)
	assert.NoError(t, err)

	req.Header.Set("Content-Type", w.FormDataContentType())

	r, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer func() {
		err := r.Body.Close()
		if err != nil {
			t.Errorf("error closing the Body: error %v", err)
		}
	}()

	b, err := ioutil.ReadAll(r.Body)
	assert.NoError(t, err)

	assert.NoError(t, err)
	assert.Equal(t, 200, r.StatusCode)

	fs := fileString("uploads_test.go", 5, "application/octet-stream")

	assert.Equal(t, `{"upload":`+fs+`}`, string(b))
}

func TestHandler_Upload_File_Forbids(t *testing.T) {
	h := &Handler{
		cfg: &Config{
			MaxRequestSize: 1024,
			Uploads: &UploadsConfig{
				Dir:    os.TempDir(),
				Forbid: []string{".go"},
			},
		},
		rr: roadrunner.NewServer(&roadrunner.ServerConfig{
			Command: "php ../../tests/http/client.php upload pipes",
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
	defer func() {
		err := hs.Shutdown(context.Background())
		if err != nil {
			t.Errorf("error during the shutdown: error %v", err)
		}
	}()

	go func() {
		err := hs.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			t.Errorf("error listening the interface: error %v", err)
		}
	}()
	time.Sleep(time.Millisecond * 10)

	var mb bytes.Buffer
	w := multipart.NewWriter(&mb)

	f := mustOpen("uploads_test.go")
	defer func() {
		err := f.Close()
		if err != nil {
			t.Errorf("failed to close a file: error %v", err)
		}
	}()
	fw, err := w.CreateFormFile("upload", f.Name())
	assert.NotNil(t, fw)
	assert.NoError(t, err)
	_, err = io.Copy(fw, f)
	if err != nil {
		t.Errorf("error copying the file: error %v", err)
	}

	err = w.Close()
	if err != nil {
		t.Errorf("error closing the file: error %v", err)
	}

	req, err := http.NewRequest("POST", "http://localhost"+hs.Addr, &mb)
	assert.NoError(t, err)

	req.Header.Set("Content-Type", w.FormDataContentType())

	r, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer func() {
		err := r.Body.Close()
		if err != nil {
			t.Errorf("error closing the Body: error %v", err)
		}
	}()

	b, err := ioutil.ReadAll(r.Body)
	assert.NoError(t, err)

	assert.NoError(t, err)
	assert.Equal(t, 200, r.StatusCode)

	fs := fileString("uploads_test.go", 7, "application/octet-stream")

	assert.Equal(t, `{"upload":`+fs+`}`, string(b))
}

func Test_FileExists(t *testing.T) {
	assert.True(t, exists("uploads_test.go"))
	assert.False(t, exists("uploads_test."))
}

func mustOpen(f string) *os.File {
	r, err := os.Open(f)
	if err != nil {
		panic(err)
	}
	return r
}

type fInfo struct {
	Name  string `json:"name"`
	Size  int64  `json:"size"`
	Mime  string `json:"mime"`
	Error int    `json:"error"`
	MD5   string `json:"md5,omitempty"`
}

func fileString(f string, errNo int, mime string) string {
	s, err := os.Stat(f)
	if err != nil {
		fmt.Println(fmt.Errorf("error stat the file, error: %v", err))
	}

	ff, err := os.Open(f)
	if err != nil {
		fmt.Println(fmt.Errorf("error opening the file, error: %v", err))
	}

	defer func() {
		er := ff.Close()
		if er != nil {
			fmt.Println(fmt.Errorf("error closing the file, error: %v", er))
		}
	}()

	h := md5.New()
	_, err = io.Copy(h, ff)
	if err != nil {
		fmt.Println(fmt.Errorf("error copying the file, error: %v", err))
	}

	v := &fInfo{
		Name:  s.Name(),
		Size:  s.Size(),
		Error: errNo,
		Mime:  mime,
		MD5:   hex.EncodeToString(h.Sum(nil)),
	}

	if errNo != 0 {
		v.MD5 = ""
		v.Size = 0
	}

	r, err := json.Marshal(v)
	if err != nil {
		fmt.Println(fmt.Errorf("error marshalling fInfo, error: %v", err))
	}
	return string(r)

}
