package http

import (
	"github.com/spiral/roadrunner"
	"github.com/stretchr/testify/assert"
	"net/http"
	"time"
	"bytes"
	"mime/multipart"
	"io/ioutil"
	"testing"
	"os"
	"context"
	"io"
	"encoding/json"
	"crypto/md5"
	"encoding/hex"
)

func TestServer_Upload_File(t *testing.T) {
	st := &Handler{
		cfg: &Config{
			MaxRequest: 1024,
			Uploads: &UploadsConfig{
				Dir:    os.TempDir(),
				Forbid: []string{},
			},
		},
		rr: roadrunner.NewServer(&roadrunner.ServerConfig{
			Command: "php ../../php-src/tests/http/client.php upload pipes",
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

	hs := &http.Server{Addr: ":8021", Handler: st,}
	defer hs.Shutdown(context.Background())

	go func() { hs.ListenAndServe() }()
	time.Sleep(time.Millisecond * 10)

	var mb bytes.Buffer
	w := multipart.NewWriter(&mb)

	f := mustOpen("uploads_test.go")
	defer f.Close()
	fw, err := w.CreateFormFile("upload", f.Name())
	assert.NotNil(t, fw)
	assert.NoError(t, err)
	io.Copy(fw, f)

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

	fs := fileString("uploads_test.go", 0, "application/octet-stream")

	assert.Equal(t, `{"upload":`+fs+`}`, string(b))
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
	MD5   string `json:"md5"`
}

func fileString(f string, err int, mime string) string {
	s, _ := os.Stat(f)

	ff, _ := os.Open(f)
	defer ff.Close()
	h := md5.New()
	io.Copy(h, ff)

	v := &fInfo{
		Name:  s.Name(),
		Size:  s.Size(),
		Error: err,
		Mime:  mime,
		MD5:   hex.EncodeToString(h.Sum(nil)),
	}

	r, _ := json.Marshal(v)
	return string(r)

}
