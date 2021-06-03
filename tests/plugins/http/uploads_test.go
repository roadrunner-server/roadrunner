package http

import (
	"bytes"
	"context"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"testing"
	"time"

	j "github.com/json-iterator/go"
	poolImpl "github.com/spiral/roadrunner/v2/pkg/pool"
	"github.com/spiral/roadrunner/v2/pkg/transport/pipe"
	handler "github.com/spiral/roadrunner/v2/pkg/worker_handler"
	"github.com/spiral/roadrunner/v2/plugins/http/config"
	"github.com/stretchr/testify/assert"
)

var json = j.ConfigCompatibleWithStandardLibrary

const testFile = "uploads_test.go"

func TestHandler_Upload_File(t *testing.T) {
	pool, err := poolImpl.Initialize(context.Background(),
		func() *exec.Cmd { return exec.Command("php", "../../http/client.php", "upload", "pipes") },
		pipe.NewPipeFactory(),
		poolImpl.Config{
			NumWorkers:      1,
			AllocateTimeout: time.Second * 1000,
			DestroyTimeout:  time.Second * 1000,
		})
	if err != nil {
		t.Fatal(err)
	}

	h, err := handler.NewHandler(1024, 500, config.Uploads{
		Dir:    os.TempDir(),
		Forbid: []string{},
	}, nil, pool)
	assert.NoError(t, err)

	hs := &http.Server{Addr: ":8021", Handler: h}
	defer func() {
		errS := hs.Shutdown(context.Background())
		if errS != nil {
			t.Errorf("error during the shutdown: error %v", errS)
		}
	}()

	go func() {
		errL := hs.ListenAndServe()
		if errL != nil && errL != http.ErrServerClosed {
			t.Errorf("error listening the interface: error %v", errL)
		}
	}()
	time.Sleep(time.Millisecond * 10)

	var mb bytes.Buffer
	w := multipart.NewWriter(&mb)

	f := mustOpen(testFile)
	defer func() {
		errC := f.Close()
		if errC != nil {
			t.Errorf("failed to close a file: error %v", errC)
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
		errC := r.Body.Close()
		if errC != nil {
			t.Errorf("error closing the Body: error %v", errC)
		}
	}()

	b, err := ioutil.ReadAll(r.Body)
	assert.NoError(t, err)

	assert.NoError(t, err)
	assert.Equal(t, 200, r.StatusCode)

	fs := fileString(testFile, 0, "application/octet-stream")

	assert.Equal(t, `{"upload":`+fs+`}`, string(b))
}

func TestHandler_Upload_NestedFile(t *testing.T) {
	pool, err := poolImpl.Initialize(context.Background(),
		func() *exec.Cmd { return exec.Command("php", "../../http/client.php", "upload", "pipes") },
		pipe.NewPipeFactory(),
		poolImpl.Config{
			NumWorkers:      1,
			AllocateTimeout: time.Second * 1000,
			DestroyTimeout:  time.Second * 1000,
		})
	if err != nil {
		t.Fatal(err)
	}

	h, err := handler.NewHandler(1024, 500, config.Uploads{
		Dir:    os.TempDir(),
		Forbid: []string{},
	}, nil, pool)
	assert.NoError(t, err)

	hs := &http.Server{Addr: ":8021", Handler: h}
	defer func() {
		errS := hs.Shutdown(context.Background())
		if errS != nil {
			t.Errorf("error during the shutdown: error %v", errS)
		}
	}()

	go func() {
		errL := hs.ListenAndServe()
		if errL != nil && errL != http.ErrServerClosed {
			t.Errorf("error listening the interface: error %v", errL)
		}
	}()
	time.Sleep(time.Millisecond * 10)

	var mb bytes.Buffer
	w := multipart.NewWriter(&mb)

	f := mustOpen(testFile)
	defer func() {
		errC := hs.Close()
		if errC != nil {
			t.Errorf("failed to close a file: error %v", errC)
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
		errC := r.Body.Close()
		if errC != nil {
			t.Errorf("error closing the Body: error %v", errC)
		}
	}()

	b, err := ioutil.ReadAll(r.Body)
	assert.NoError(t, err)

	assert.NoError(t, err)
	assert.Equal(t, 200, r.StatusCode)

	fs := fileString(testFile, 0, "application/octet-stream")

	assert.Equal(t, `{"upload":{"x":{"y":{"z":[`+fs+`]}}}}`, string(b))
}

func TestHandler_Upload_File_NoTmpDir(t *testing.T) {
	pool, err := poolImpl.Initialize(context.Background(),
		func() *exec.Cmd { return exec.Command("php", "../../http/client.php", "upload", "pipes") },
		pipe.NewPipeFactory(),
		poolImpl.Config{
			NumWorkers:      1,
			AllocateTimeout: time.Second * 1000,
			DestroyTimeout:  time.Second * 1000,
		})
	if err != nil {
		t.Fatal(err)
	}

	h, err := handler.NewHandler(1024, 500, config.Uploads{
		Dir:    "-------",
		Forbid: []string{},
	}, nil, pool)
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
	time.Sleep(time.Millisecond * 10)

	var mb bytes.Buffer
	w := multipart.NewWriter(&mb)

	f := mustOpen(testFile)
	defer func() {
		errC := hs.Close()
		if errC != nil {
			t.Errorf("failed to close a file: error %v", errC)
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
		errC := r.Body.Close()
		if errC != nil {
			t.Errorf("error closing the Body: error %v", errC)
		}
	}()

	b, err := ioutil.ReadAll(r.Body)
	assert.NoError(t, err)

	assert.NoError(t, err)
	assert.Equal(t, 200, r.StatusCode)

	fs := fileString(testFile, 6, "application/octet-stream")

	assert.Equal(t, `{"upload":`+fs+`}`, string(b))
}

func TestHandler_Upload_File_Forbids(t *testing.T) {
	pool, err := poolImpl.Initialize(context.Background(),
		func() *exec.Cmd { return exec.Command("php", "../../http/client.php", "upload", "pipes") },
		pipe.NewPipeFactory(),
		poolImpl.Config{
			NumWorkers:      1,
			AllocateTimeout: time.Second * 1000,
			DestroyTimeout:  time.Second * 1000,
		})
	if err != nil {
		t.Fatal(err)
	}

	h, err := handler.NewHandler(1024, 500, config.Uploads{
		Dir:    os.TempDir(),
		Forbid: []string{".go"},
	}, nil, pool)
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
	time.Sleep(time.Millisecond * 10)

	var mb bytes.Buffer
	w := multipart.NewWriter(&mb)

	f := mustOpen(testFile)
	defer func() {
		errC := hs.Close()
		if errC != nil {
			t.Errorf("failed to close a file: error %v", errC)
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
		errC := r.Body.Close()
		if errC != nil {
			t.Errorf("error closing the Body: error %v", errC)
		}
	}()

	b, err := ioutil.ReadAll(r.Body)
	assert.NoError(t, err)

	assert.NoError(t, err)
	assert.Equal(t, 200, r.StatusCode)

	fs := fileString(testFile, 8, "application/octet-stream")

	assert.Equal(t, `{"upload":`+fs+`}`, string(b))
}

func Test_FileExists(t *testing.T) {
	assert.True(t, exists(testFile))
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
	Name   string `json:"name"`
	Size   int64  `json:"size"`
	Mime   string `json:"mime"`
	Error  int    `json:"error"`
	Sha512 string `json:"sha512,omitempty"`
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

	h := sha512.New()
	_, err = io.Copy(h, ff)
	if err != nil {
		fmt.Println(fmt.Errorf("error copying the file, error: %v", err))
	}

	v := &fInfo{
		Name:   s.Name(),
		Size:   s.Size(),
		Error:  errNo,
		Mime:   mime,
		Sha512: hex.EncodeToString(h.Sum(nil)),
	}

	if errNo != 0 {
		v.Sha512 = ""
		v.Size = 0
	}

	r, err := json.Marshal(v)
	if err != nil {
		fmt.Println(fmt.Errorf("error marshaling fInfo, error: %v", err))
	}
	return string(r)
}

// exists if file exists.
func exists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}
