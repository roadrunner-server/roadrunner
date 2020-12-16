package tests

import (
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/spiral/endure"
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/mocks"
	"github.com/spiral/roadrunner/v2/plugins/config"
	httpPlugin "github.com/spiral/roadrunner/v2/plugins/http"
	"github.com/spiral/roadrunner/v2/plugins/logger"
	"github.com/spiral/roadrunner/v2/plugins/reload"
	"github.com/spiral/roadrunner/v2/plugins/resetter"
	"github.com/spiral/roadrunner/v2/plugins/server"
	"github.com/stretchr/testify/assert"
)

const testDir string = "unit_tests"
const hugeNumberOfFiles uint = 5000

func TestReloadInit(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.DebugLevel))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "configs/.rr-reload.yaml",
		Prefix: "rr",
	}

	// try to remove, skip error
	assert.NoError(t, freeResources(testDir))
	err = os.Mkdir(testDir, 0755)
	assert.NoError(t, err)

	defer func() {
		assert.NoError(t, freeResources(testDir))
	}()

	controller := gomock.NewController(t)
	mockLogger := mocks.NewMockLogger(controller)

	mockLogger.EXPECT().Debug("file was created", "path", gomock.Any(), "name", "file.txt", "size", gomock.Any()).Times(2)
	mockLogger.EXPECT().Info("HTTP plugin got restart request. Restarting...").Times(1)
	mockLogger.EXPECT().Info("HTTP workers Pool successfully restarted").Times(1)
	mockLogger.EXPECT().Info("HTTP listeners successfully re-added").Times(1)
	mockLogger.EXPECT().Info("HTTP plugin successfully restarted").Times(1)

	err = cont.RegisterAll(
		cfg,
		mockLogger,
		&server.Plugin{},
		&httpPlugin.Plugin{},
		&reload.Plugin{},
		&resetter.Plugin{},
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

	tt := time.NewTimer(time.Second * 10)

	go func() {
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

	t.Run("ReloadTestInit", reloadTestInit)

	wg.Wait()
}

func reloadTestInit(t *testing.T) {
	err := ioutil.WriteFile(filepath.Join(testDir, "file.txt"), //nolint:gosec
		[]byte{}, 0755)
	assert.NoError(t, err)
}

func TestReloadHugeNumberOfFiles(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.DebugLevel))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "configs/.rr-reload.yaml",
		Prefix: "rr",
	}

	// try to remove, skip error
	assert.NoError(t, freeResources(testDir))
	err = os.Mkdir(testDir, 0755)
	assert.NoError(t, err)

	defer func() {
		assert.NoError(t, freeResources(testDir))
	}()

	// controller := gomock.NewController(t)
	// mockLogger := mocks.NewMockLogger(controller)
	//
	// mockLogger.EXPECT().Debug("file was created", "path", gomock.Any(), "name", "file.txt", "size", gomock.Any()).Times(2)
	// mockLogger.EXPECT().Info("Resetting http plugin").Times(1)

	err = cont.RegisterAll(
		cfg,
		// mockLogger,
		&logger.ZapLogger{},
		&server.Plugin{},
		&httpPlugin.Plugin{},
		&reload.Plugin{},
		&resetter.Plugin{},
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

	tt := time.NewTimer(time.Second * 160)

	go func() {
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

	t.Run("ReloadTestHugeNumberOfFiles", reloadHugeNumberOfFiles)
	ttt := time.Now()
	t.Run("ReloadRandomlyChangeFile", randomlyChangeFile)
	if time.Since(ttt).Seconds() > 140 {
		t.Fatal("spend too much time on reloading")
	}
	t.Run("ReloadHTTPLiveAfterReset", reloadHTTPLiveAfterReset)

	wg.Wait()
}

func reloadHTTPLiveAfterReset(t *testing.T) {
	req, err := http.NewRequest("GET", "http://localhost:22388", nil)
	assert.NoError(t, err)

	r, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)

	b, err := ioutil.ReadAll(r.Body)
	assert.NoError(t, err)

	assert.Equal(t, 200, r.StatusCode)
	assert.Equal(t, "hello world", string(b))

	err = r.Body.Close()
	assert.NoError(t, err)
}

func randomlyChangeFile(t *testing.T) {
	// we know, that directory contains 5000 files (0-4999)
	// let's try to randomly change it
	for i := 0; i < 100; i++ {
		// rand sleep
		rSleep := rand.Int63n(1000) // nolint:gosec
		time.Sleep(time.Millisecond * time.Duration(rSleep))
		rNum := rand.Int63n(int64(hugeNumberOfFiles))                                                                            // nolint:gosec
		err := ioutil.WriteFile(filepath.Join(testDir, "file_"+strconv.Itoa(int(rNum))+".txt"), []byte("Hello, Gophers!"), 0755) // nolint:gosec
		assert.NoError(t, err)
	}
}

func reloadHugeNumberOfFiles(t *testing.T) {
	for i := uint(0); i < hugeNumberOfFiles; i++ {
		assert.NoError(t, makeFile("file_"+strconv.Itoa(int(i))+".txt"))
	}
}

func freeResources(path string) error {
	return os.RemoveAll(path)
}

func makeFile(filename string) error {
	return ioutil.WriteFile(filepath.Join(testDir, filename), []byte{}, 0755) //nolint:gosec
}

func copyDir(src string, dst string) error {
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)

	si, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !si.IsDir() {
		return errors.E(errors.Str("source is not a directory"))
	}

	_, err = os.Stat(dst)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if err == nil {
		return errors.E(errors.Str("destination already exists"))
	}

	err = os.MkdirAll(dst, si.Mode())
	if err != nil {
		return err
	}

	entries, err := ioutil.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			err = copyDir(srcPath, dstPath)
			if err != nil {
				return err
			}
		} else {
			// Skip symlinks.
			if entry.Mode()&os.ModeSymlink != 0 {
				continue
			}

			err = copyFile(srcPath, dstPath)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return errors.E(err)
	}
	defer func() {
		_ = in.Close()
	}()

	out, err := os.Create(dst)
	if err != nil {
		return errors.E(err)
	}
	defer func() {
		_ = out.Close()
	}()

	_, err = io.Copy(out, in)
	if err != nil {
		return errors.E(err)
	}

	err = out.Sync()
	if err != nil {
		return errors.E(err)
	}

	si, err := os.Stat(src)
	if err != nil {
		return errors.E(err)
	}
	err = os.Chmod(dst, si.Mode())
	if err != nil {
		return errors.E(err)
	}
	return nil
}
