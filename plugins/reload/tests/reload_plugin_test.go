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

	"github.com/spiral/endure"
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/plugins/config"
	httpPlugin "github.com/spiral/roadrunner/v2/plugins/http"
	"github.com/spiral/roadrunner/v2/plugins/logger"
	"github.com/spiral/roadrunner/v2/plugins/reload"
	"github.com/spiral/roadrunner/v2/plugins/resetter"
	"github.com/spiral/roadrunner/v2/plugins/server"
	"github.com/stretchr/testify/assert"
)

const testDir string = "unit_tests"
const testCopyToDir string = "unit_tests_copied"
const hugeNumberOfFiles uint = 5000

func TestReloadInit(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
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

	err = cont.RegisterAll(
		cfg,
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
	reloadHTTPLiveAfterReset(t, "22388")

	wg.Wait()
}

func reloadTestInit(t *testing.T) {
	err := ioutil.WriteFile(filepath.Join(testDir, "file.txt"), //nolint:gosec
		[]byte{}, 0755)
	assert.NoError(t, err)
}

func TestReloadHugeNumberOfFiles(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "configs/.rr-reload.yaml",
		Prefix: "rr",
	}

	// try to remove, skip error
	assert.NoError(t, freeResources(testDir))
	assert.NoError(t, freeResources(testCopyToDir))
	err = os.Mkdir(testDir, 0755)
	assert.NoError(t, err)
	err = os.Mkdir(testCopyToDir, 0755)
	assert.NoError(t, err)

	defer func() {
		assert.NoError(t, freeResources(testDir))
		assert.NoError(t, freeResources(testCopyToDir))
	}()

	err = cont.RegisterAll(
		cfg,
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

	tt := time.NewTimer(time.Second * 100)

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
	if time.Since(ttt).Seconds() > 80 {
		t.Fatal("spend too much time on reloading")
	}
	reloadHTTPLiveAfterReset(t, "22388")

	wg.Wait()
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

// Should be events only about creating files with txt ext
func TestReloadFilterFileExt(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "configs/.rr-reload-2.yaml",
		Prefix: "rr",
	}

	// try to remove, skip error
	assert.NoError(t, freeResources(testDir))
	err = os.Mkdir(testDir, 0755)
	assert.NoError(t, err)

	defer func() {
		assert.NoError(t, freeResources(testDir))
	}()

	err = cont.RegisterAll(
		cfg,
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

	tt := time.NewTimer(time.Second * 40)

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

	t.Run("ReloadMakeFiles", reloadMakeFiles)
	ttt := time.Now()
	t.Run("ReloadFilteredExt", reloadFilteredExt)
	if time.Since(ttt).Seconds() > 20 {
		t.Fatal("spend too much time on reloading")
	}

	reloadHTTPLiveAfterReset(t, "27388")

	wg.Wait()
}

func reloadMakeFiles(t *testing.T) {
	for i := uint(0); i < 100; i++ {
		assert.NoError(t, makeFile("file_"+strconv.Itoa(int(i))+".txt"))
	}
	for i := uint(0); i < 100; i++ {
		assert.NoError(t, makeFile("file_"+strconv.Itoa(int(i))+".abc"))
	}
	for i := uint(0); i < 100; i++ {
		assert.NoError(t, makeFile("file_"+strconv.Itoa(int(i))+".def"))
	}
}

func reloadFilteredExt(t *testing.T) {
	// change files with abc extension
	for i := 0; i < 10; i++ {
		// rand sleep
		rSleep := rand.Int63n(1000) // nolint:gosec
		time.Sleep(time.Millisecond * time.Duration(rSleep))
		rNum := rand.Int63n(int64(hugeNumberOfFiles))                                                                            // nolint:gosec
		err := ioutil.WriteFile(filepath.Join(testDir, "file_"+strconv.Itoa(int(rNum))+".abc"), []byte("Hello, Gophers!"), 0755) // nolint:gosec
		assert.NoError(t, err)
	}

	// change files with def extension
	for i := 0; i < 10; i++ {
		// rand sleep
		rSleep := rand.Int63n(1000) // nolint:gosec
		time.Sleep(time.Millisecond * time.Duration(rSleep))
		rNum := rand.Int63n(int64(hugeNumberOfFiles))                                                                            // nolint:gosec
		err := ioutil.WriteFile(filepath.Join(testDir, "file_"+strconv.Itoa(int(rNum))+".def"), []byte("Hello, Gophers!"), 0755) // nolint:gosec
		assert.NoError(t, err)
	}
}

// Should be events only about creating files with txt ext
func TestReloadCopy3k(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "configs/.rr-reload-3.yaml",
		Prefix: "rr",
	}

	// try to remove, skip error
	assert.NoError(t, freeResources(testDir))
	assert.NoError(t, freeResources(testCopyToDir))
	assert.NoError(t, freeResources("dir1"))
	err = os.Mkdir(testDir, 0755)
	assert.NoError(t, err)
	err = os.Mkdir(testCopyToDir, 0755)
	assert.NoError(t, err)
	err = os.Mkdir("dir1", 0755)
	assert.NoError(t, err)

	defer func() {
		assert.NoError(t, freeResources(testDir))
		assert.NoError(t, freeResources(testCopyToDir))
		assert.NoError(t, freeResources("dir1"))
	}()

	err = cont.RegisterAll(
		cfg,
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

	tt := time.NewTimer(time.Second * 220)

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

	// Scenario
	// 1
	// Create 3k files with txt, abc, def extensions
	// Copy files to the unit_tests_copy dir
	// 2
	// Delete both dirs, recreate
	// Create 3k files with txt, abc, def extensions
	// Move files to the unit_tests_copy dir
	// 3
	// Recursive

	t.Run("ReloadMake3kFiles", reloadMake3kFiles)
	ttt := time.Now()
	t.Run("ReloadCopyFiles", reloadCopyFiles)
	if time.Since(ttt).Seconds() > 120 {
		t.Fatal("spend too much time on copy")
	}

	t.Run("ReloadRecursiveDirsSupport", copyFilesRecursive)
	t.Run("RandomChangesInRecursiveDirs", randomChangesInRecursiveDirs)
	t.Run("RemoveFilesSupport", removeFilesSupport)
	t.Run("ReloadMoveSupport", reloadMoveSupport)

	reloadHTTPLiveAfterReset(t, "37388")

	wg.Wait()
}

func reloadMoveSupport(t *testing.T) {
	t.Run("MoveSupportCopy", copyFilesRecursive)
	// move some files
	for i := 0; i < 50; i++ {
		// rand sleep
		rSleep := rand.Int63n(1000) // nolint:gosec
		time.Sleep(time.Millisecond * time.Duration(rSleep))
		rNum := rand.Int63n(int64(200)) // nolint:gosec
		rDir := rand.Int63n(9)          // nolint:gosec
		rExt := rand.Int63n(3)          // nolint:gosec

		ext := []string{
			".txt",
			".abc",
			".def",
		}

		// change files with def extension
		dirs := []string{
			"dir1",
			"dir1/dir2",
			"dir1/dir2/dir3",
			"dir1/dir2/dir3/dir4",
			"dir1/dir2/dir3/dir4/dir5",
			"dir1/dir2/dir3/dir4/dir5/dir6",
			"dir1/dir2/dir3/dir4/dir5/dir6/dir7",
			"dir1/dir2/dir3/dir4/dir5/dir6/dir7/dir8",
			"dir1/dir2/dir3/dir4/dir5/dir6/dir7/dir8/dir9",
			"dir1/dir2/dir3/dir4/dir5/dir6/dir7/dir8/dir9/dir10",
		}

		// move file
		err := os.Rename(filepath.Join(dirs[rDir], "file_"+strconv.Itoa(int(rNum))+ext[rExt]), filepath.Join(dirs[rDir+1], "file_"+strconv.Itoa(int(rNum))+ext[rExt]))
		assert.NoError(t, err)
	}
}

func removeFilesSupport(t *testing.T) {
	// remove some files
	for i := 0; i < 50; i++ {
		// rand sleep
		rSleep := rand.Int63n(1000) // nolint:gosec
		time.Sleep(time.Millisecond * time.Duration(rSleep))
		rNum := rand.Int63n(int64(200)) // nolint:gosec
		rDir := rand.Int63n(10)         // nolint:gosec
		rExt := rand.Int63n(3)          // nolint:gosec

		ext := []string{
			".txt",
			".abc",
			".def",
		}

		// change files with def extension
		dirs := []string{
			"dir1",
			"dir1/dir2",
			"dir1/dir2/dir3",
			"dir1/dir2/dir3/dir4",
			"dir1/dir2/dir3/dir4/dir5",
			"dir1/dir2/dir3/dir4/dir5/dir6",
			"dir1/dir2/dir3/dir4/dir5/dir6/dir7",
			"dir1/dir2/dir3/dir4/dir5/dir6/dir7/dir8",
			"dir1/dir2/dir3/dir4/dir5/dir6/dir7/dir8/dir9",
			"dir1/dir2/dir3/dir4/dir5/dir6/dir7/dir8/dir9/dir10",
		}
		err := os.Remove(filepath.Join(dirs[rDir], "file_"+strconv.Itoa(int(rNum))+ext[rExt]))
		assert.NoError(t, err)
	}
}

func randomChangesInRecursiveDirs(t *testing.T) {
	// change files with def extension
	dirs := []string{
		"dir1",
		"dir1/dir2",
		"dir1/dir2/dir3",
		"dir1/dir2/dir3/dir4",
		"dir1/dir2/dir3/dir4/dir5",
		"dir1/dir2/dir3/dir4/dir5/dir6",
		"dir1/dir2/dir3/dir4/dir5/dir6/dir7",
		"dir1/dir2/dir3/dir4/dir5/dir6/dir7/dir8",
		"dir1/dir2/dir3/dir4/dir5/dir6/dir7/dir8/dir9",
		"dir1/dir2/dir3/dir4/dir5/dir6/dir7/dir8/dir9/dir10",
	}

	ext := []string{
		".txt",
		".abc",
		".def",
	}

	filenames := []string{
		"file_", // should be update
		"foo_",  // should be created
		"bar_",  // should be created
	}
	for i := 0; i < 50; i++ {
		// rand sleep
		rSleep := rand.Int63n(1000) // nolint:gosec
		time.Sleep(time.Millisecond * time.Duration(rSleep))
		rNum := rand.Int63n(int64(200)) // nolint:gosec
		rDir := rand.Int63n(10)         // nolint:gosec
		rExt := rand.Int63n(3)          // nolint:gosec
		rName := rand.Int63n(3)         // nolint:gosec

		err := ioutil.WriteFile(filepath.Join(dirs[rDir], filenames[rName]+strconv.Itoa(int(rNum))+ext[rExt]), []byte("Hello, Gophers!"), 0755) // nolint:gosec
		assert.NoError(t, err)
	}
}

func copyFilesRecursive(t *testing.T) {
	err := copyDir(testDir, "dir1")
	assert.NoError(t, err)
	err = copyDir(testDir, "dir1/dir2")
	assert.NoError(t, err)
	err = copyDir(testDir, "dir1/dir2/dir3")
	assert.NoError(t, err)
	err = copyDir(testDir, "dir1/dir2/dir3/dir4")
	assert.NoError(t, err)
	err = copyDir(testDir, "dir1/dir2/dir3/dir4/dir5")
	assert.NoError(t, err)
	err = copyDir(testDir, "dir1/dir2/dir3/dir4/dir5/dir6")
	assert.NoError(t, err)
	err = copyDir(testDir, "dir1/dir2/dir3/dir4/dir5/dir6/dir7")
	assert.NoError(t, err)
	err = copyDir(testDir, "dir1/dir2/dir3/dir4/dir5/dir6/dir7/dir8")
	assert.NoError(t, err)
	err = copyDir(testDir, "dir1/dir2/dir3/dir4/dir5/dir6/dir7/dir8/dir9")
	assert.NoError(t, err)
	err = copyDir(testDir, "dir1/dir2/dir3/dir4/dir5/dir6/dir7/dir8/dir9/dir10")
	assert.NoError(t, err)
}

func reloadCopyFiles(t *testing.T) {
	err := copyDir(testDir, testCopyToDir)
	assert.NoError(t, err)

	assert.NoError(t, freeResources(testDir))
	assert.NoError(t, freeResources(testCopyToDir))
	err = os.Mkdir(testDir, 0755)
	assert.NoError(t, err)
	err = os.Mkdir(testCopyToDir, 0755)
	assert.NoError(t, err)

	// recreate files
	for i := uint(0); i < 200; i++ {
		assert.NoError(t, makeFile("file_"+strconv.Itoa(int(i))+".txt"))
	}
	for i := uint(0); i < 200; i++ {
		assert.NoError(t, makeFile("file_"+strconv.Itoa(int(i))+".abc"))
	}
	for i := uint(0); i < 200; i++ {
		assert.NoError(t, makeFile("file_"+strconv.Itoa(int(i))+".def"))
	}

	err = copyDir(testDir, testCopyToDir)
	assert.NoError(t, err)
}

func reloadMake3kFiles(t *testing.T) {
	for i := uint(0); i < 100; i++ {
		assert.NoError(t, makeFile("file_"+strconv.Itoa(int(i))+".txt"))
	}
	for i := uint(0); i < 100; i++ {
		assert.NoError(t, makeFile("file_"+strconv.Itoa(int(i))+".abc"))
	}
	for i := uint(0); i < 100; i++ {
		assert.NoError(t, makeFile("file_"+strconv.Itoa(int(i))+".def"))
	}
}

func TestReloadNoRecursion(t *testing.T) {
	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	assert.NoError(t, err)

	cfg := &config.Viper{
		Path:   "configs/.rr-reload-4.yaml",
		Prefix: "rr",
	}

	// try to remove, skip error
	assert.NoError(t, freeResources(testDir))
	assert.NoError(t, freeResources(testCopyToDir))
	assert.NoError(t, freeResources("dir1"))
	err = os.Mkdir(testDir, 0755)
	assert.NoError(t, err)

	err = os.Mkdir("dir1", 0755)
	assert.NoError(t, err)

	err = os.Mkdir(testCopyToDir, 0755)
	assert.NoError(t, err)

	defer func() {
		assert.NoError(t, freeResources(testDir))
		assert.NoError(t, freeResources(testCopyToDir))
		assert.NoError(t, freeResources("dir1"))
	}()

	err = cont.RegisterAll(
		cfg,
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

	tt := time.NewTimer(time.Second * 30)

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
	t.Run("ReloadMakeFiles", reloadMakeFiles) // make files in the testDir
	t.Run("ReloadCopyFilesRecursive", reloadCopyFiles)

	reloadHTTPLiveAfterReset(t, "22766")

	wg.Wait()
}

// ========================================================================

func reloadHTTPLiveAfterReset(t *testing.T, port string) {
	req, err := http.NewRequest("GET", "http://localhost:"+port, nil)
	assert.NoError(t, err)

	r, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	b, err := ioutil.ReadAll(r.Body)
	assert.NoError(t, err)

	assert.Equal(t, 200, r.StatusCode)
	assert.Equal(t, "hello world", string(b))

	err = r.Body.Close()
	assert.NoError(t, err)
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
