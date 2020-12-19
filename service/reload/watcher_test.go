package reload

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var testServiceName = "test"

// scenario
// Create walker instance, init with default config, check that Watcher found all files from config
func Test_Correct_Watcher_Init(t *testing.T) {
	tempDir, err := ioutil.TempDir(".", "")
	defer func() {
		err = freeResources(tempDir)
		if err != nil {
			t.Fatal(err)
		}
	}()
	if err != nil {
		t.Fatal(err)
	}
	err = ioutil.WriteFile(filepath.Join(tempDir, "file.txt"),
		[]byte{}, 0755)
	if err != nil {
		t.Fatal(err)
	}

	wc := WatcherConfig{
		ServiceName:  testServiceName,
		Recursive:    false,
		Directories:  []string{tempDir},
		FilterHooks:  nil,
		Files:        make(map[string]os.FileInfo),
		Ignored:      nil,
		FilePatterns: nil,
	}

	w, err := NewWatcher([]WatcherConfig{wc})
	if err != nil {
		t.Fatal(err)
	}

	if len(w.GetAllFiles(testServiceName)) != 2 {
		t.Fatal("incorrect directories len")
	}
}

// scenario
// create 3 files, create walker instance
// Start poll events
// change file and see, if event had come to handler
func Test_Get_FileEvent(t *testing.T) {
	tempDir, err := ioutil.TempDir(".", "")
	c := make(chan struct{})
	defer func(name string) {
		err = freeResources(name)
		if err != nil {
			c <- struct{}{}
			t.Fatal(err)
		}
		c <- struct{}{}
	}(tempDir)

	if err != nil {
		t.Fatal(err)
	}
	err = ioutil.WriteFile(filepath.Join(tempDir, "file1.txt"),
		[]byte{}, 0755)
	if err != nil {
		t.Fatal(err)
	}

	err = ioutil.WriteFile(filepath.Join(tempDir, "file2.txt"),
		[]byte{}, 0755)
	if err != nil {
		t.Fatal(err)
	}

	err = ioutil.WriteFile(filepath.Join(tempDir, "file3.txt"),
		[]byte{}, 0755)
	if err != nil {
		t.Fatal(err)
	}

	wc := WatcherConfig{
		ServiceName:  testServiceName,
		Recursive:    false,
		Directories:  []string{tempDir},
		FilterHooks:  nil,
		Files:        make(map[string]os.FileInfo),
		Ignored:      nil,
		FilePatterns: []string{"aaa", "txt"},
	}

	w, err := NewWatcher([]WatcherConfig{wc})
	if err != nil {
		t.Fatal(err)
	}

	// should be 3 files and directory
	if len(w.GetAllFiles(testServiceName)) != 4 {
		t.Fatal("incorrect directories len")
	}

	go limitTime(time.Second*10, t.Name(), c)

	go func() {
		go func() {
			time.Sleep(time.Second)
			err2 := ioutil.WriteFile(filepath.Join(tempDir, "file2.txt"),
				[]byte{1, 1, 1}, 0755)
			if err2 != nil {
				panic(err2)
			}
			runtime.Goexit()
		}()

		go func() {
			for e := range w.Event {
				if e.Path != "file2.txt" {
					panic("didn't handle event when write file2")
				}
				w.Stop()
			}
		}()
	}()

	err = w.StartPolling(time.Second)
	if err != nil {
		t.Fatal(err)
	}
}

// scenario
// create 3 files with different extensions, create walker instance
// Start poll events
// change file with txt extension, and see, if event had not come to handler because it was filtered
func Test_FileExtensionFilter(t *testing.T) {
	tempDir, err := ioutil.TempDir(".", "")
	c := make(chan struct{})
	defer func(name string) {
		err = freeResources(name)
		if err != nil {
			c <- struct{}{}
			t.Fatal(err)
		}
		c <- struct{}{}
	}(tempDir)

	if err != nil {
		t.Fatal(err)
	}
	err = ioutil.WriteFile(filepath.Join(tempDir, "file1.aaa"),
		[]byte{}, 0755)
	if err != nil {
		t.Fatal(err)
	}

	err = ioutil.WriteFile(filepath.Join(tempDir, "file2.bbb"),
		[]byte{}, 0755)
	if err != nil {
		t.Fatal(err)
	}

	err = ioutil.WriteFile(filepath.Join(tempDir, "file3.txt"),
		[]byte{}, 0755)
	if err != nil {
		t.Fatal(err)
	}
	wc := WatcherConfig{
		ServiceName: testServiceName,
		Recursive:   false,
		Directories: []string{tempDir},
		FilterHooks: func(filename string, patterns []string) error {
			for i := 0; i < len(patterns); i++ {
				if strings.Contains(filename, patterns[i]) {
					return nil
				}
			}
			return ErrorSkip
		},
		Files:        make(map[string]os.FileInfo),
		Ignored:      nil,
		FilePatterns: []string{"aaa", "bbb"},
	}

	w, err := NewWatcher([]WatcherConfig{wc})
	if err != nil {
		t.Fatal(err)
	}

	dirLen := len(w.GetAllFiles(testServiceName))
	// should be 2 files (one filtered) and directory
	if dirLen != 3 {
		t.Fatalf("incorrect directories len, len is: %d", dirLen)
	}

	go limitTime(time.Second*10, t.Name(), c)

	go func() {
		go func() {
			time.Sleep(time.Second)
			err := ioutil.WriteFile(filepath.Join(tempDir, "file3.txt"),
				[]byte{1, 1, 1}, 0755)
			assert.NoError(t, err)
		}()

		go func() {
			for e := range w.Event {
				fmt.Println(e.Info.Name())
				panic("handled event from filtered file")
			}
		}()
		time.Sleep(time.Second)
		w.Stop()
	}()

	err = w.StartPolling(time.Second)
	assert.NoError(t, err)
}

// nested
// scenario
// create dir and nested dir
// make files with aaa, bbb and txt extensions, filter txt
// change not filtered file, handle event
func Test_Recursive_Support(t *testing.T) {
	tempDir, err := ioutil.TempDir(".", "")
	defer func() {
		err = freeResources(tempDir)
		assert.NoError(t, err)
	}()

	nestedDir, err := ioutil.TempDir(tempDir, "nested")
	assert.NoError(t, err)

	err = ioutil.WriteFile(filepath.Join(tempDir, "file1.aaa"),
		[]byte{}, 0755)
	assert.NoError(t, err)

	err = ioutil.WriteFile(filepath.Join(tempDir, "file2.bbb"),
		[]byte{}, 0755)
	assert.NoError(t, err)

	err = ioutil.WriteFile(filepath.Join(nestedDir, "file3.txt"),
		[]byte{}, 0755)
	assert.NoError(t, err)
	err = ioutil.WriteFile(filepath.Join(nestedDir, "file4.aaa"),
		[]byte{}, 0755)
	assert.NoError(t, err)

	wc := WatcherConfig{
		ServiceName: testServiceName,
		Recursive:   true,
		Directories: []string{tempDir},
		FilterHooks: func(filename string, patterns []string) error {
			for i := 0; i < len(patterns); i++ {
				if strings.Contains(filename, patterns[i]) {
					return nil
				}
			}
			return ErrorSkip
		},
		Files:        make(map[string]os.FileInfo),
		Ignored:      nil,
		FilePatterns: []string{"aaa", "bbb"},
	}

	w, err := NewWatcher([]WatcherConfig{wc})
	assert.NoError(t, err)

	dirLen := len(w.GetAllFiles(testServiceName))
	// should be 3 files (2 from root dir, and 1 from nested), filtered txt
	if dirLen != 3 {
		t.Fatalf("incorrect directories len, len is: %d", dirLen)
	}

	go func() {
		// time sleep is used here because StartPolling is blocking operation
		time.Sleep(time.Second * 5)
		// change file in nested directory
		err = ioutil.WriteFile(filepath.Join(nestedDir, "file4.aaa"),
			[]byte{1, 1, 1}, 0755)
		assert.NoError(t, err)
		go func() {
			time.Sleep(time.Second)
			for e := range w.Event {
				if e.Info.Name() != "file4.aaa" {
					panic("wrong handled event from watcher in nested dir")
				}
				w.Stop()
			}
		}()
	}()

	err = w.StartPolling(time.Second)
	assert.NoError(t, err)
}

func Test_Wrong_Dir(t *testing.T) {
	// no such file or directory
	wrongDir := "askdjfhaksdlfksdf"

	wc := WatcherConfig{
		ServiceName: testServiceName,
		Recursive:   true,
		Directories: []string{wrongDir},
		FilterHooks: func(filename string, patterns []string) error {
			for i := 0; i < len(patterns); i++ {
				if strings.Contains(filename, patterns[i]) {
					return nil
				}
			}
			return ErrorSkip
		},
		Files:        make(map[string]os.FileInfo),
		Ignored:      nil,
		FilePatterns: []string{"aaa", "bbb"},
	}

	_, err := NewWatcher([]WatcherConfig{wc})
	assert.Error(t, err)
}

func Test_Filter_Directory(t *testing.T) {
	tempDir, err := ioutil.TempDir(".", "")
	c := make(chan struct{})
	defer func(name string) {
		err = freeResources(name)
		if err != nil {
			c <- struct{}{}
			t.Fatal(err)
		}
		c <- struct{}{}
	}(tempDir)

	go limitTime(time.Second*10, t.Name(), c)

	nestedDir, err := ioutil.TempDir(tempDir, "nested")
	assert.NoError(t, err)

	err = ioutil.WriteFile(filepath.Join(tempDir, "file1.aaa"),
		[]byte{}, 0755)
	assert.NoError(t, err)

	err = ioutil.WriteFile(filepath.Join(tempDir, "file2.bbb"),
		[]byte{}, 0755)
	assert.NoError(t, err)

	err = ioutil.WriteFile(filepath.Join(nestedDir, "file3.txt"),
		[]byte{}, 0755)
	assert.NoError(t, err)
	err = ioutil.WriteFile(filepath.Join(nestedDir, "file4.aaa"),
		[]byte{}, 0755)
	assert.NoError(t, err)

	ignored, err := ConvertIgnored([]string{nestedDir})
	if err != nil {
		t.Fatal(err)
	}
	wc := WatcherConfig{
		ServiceName: testServiceName,
		Recursive:   true,
		Directories: []string{tempDir},
		FilterHooks: func(filename string, patterns []string) error {
			for i := 0; i < len(patterns); i++ {
				if strings.Contains(filename, patterns[i]) {
					return nil
				}
			}
			return ErrorSkip
		},
		Files:        make(map[string]os.FileInfo),
		Ignored:      ignored,
		FilePatterns: []string{"aaa", "bbb", "txt"},
	}

	w, err := NewWatcher([]WatcherConfig{wc})
	if err != nil {
		t.Fatal(err)
	}

	dirLen := len(w.GetAllFiles(testServiceName))
	// should be 2 files (2 from root dir), filtered other
	if dirLen != 2 {
		t.Fatalf("incorrect directories len, len is: %d", dirLen)
	}

	go func() {
		go func() {
			err := ioutil.WriteFile(filepath.Join(nestedDir, "file4.aaa"),
				[]byte{1, 1, 1}, 0755)
			assert.NoError(t, err)
		}()

		go func() {
			time.Sleep(time.Second)
			for e := range w.Event {
				fmt.Println("file: " + e.Info.Name())
				panic("handled event from watcher in nested dir")
			}
		}()

		// time sleep is used here because StartPolling is blocking operation
		time.Sleep(time.Second * 5)
		w.Stop()

	}()

	err = w.StartPolling(time.Second)
	if err != nil {
		t.Fatal(err)
	}
}

// copy files from nested dir to not ignored
// should fire an event
func Test_Copy_Directory(t *testing.T) {
	tempDir, err := ioutil.TempDir(".", "")
	c := make(chan struct{})
	defer func() {
		err = freeResources(tempDir)
		if err != nil {
			c <- struct{}{}
			t.Fatal(err)
		}
		c <- struct{}{}
	}()

	go limitTime(time.Second*20, t.Name(), c)

	nestedDir, err := ioutil.TempDir(tempDir, "nested")
	assert.NoError(t, err)

	err = ioutil.WriteFile(filepath.Join(tempDir, "file1.aaa"),
		[]byte{}, 0755)
	assert.NoError(t, err)

	err = ioutil.WriteFile(filepath.Join(tempDir, "file2.bbb"),
		[]byte{}, 0755)
	assert.NoError(t, err)

	err = ioutil.WriteFile(filepath.Join(nestedDir, "file3.txt"),
		[]byte{}, 0755)
	assert.NoError(t, err)
	err = ioutil.WriteFile(filepath.Join(nestedDir, "file4.aaa"),
		[]byte{}, 0755)
	assert.NoError(t, err)

	ignored, err := ConvertIgnored([]string{nestedDir})
	assert.NoError(t, err)

	wc := WatcherConfig{
		ServiceName: testServiceName,
		Recursive:   true,
		Directories: []string{tempDir},
		FilterHooks: func(filename string, patterns []string) error {
			for i := 0; i < len(patterns); i++ {
				if strings.Contains(filename, patterns[i]) {
					return nil
				}
			}
			return ErrorSkip
		},
		Files:        make(map[string]os.FileInfo),
		Ignored:      ignored,
		FilePatterns: []string{"aaa", "bbb", "txt"},
	}

	w, err := NewWatcher([]WatcherConfig{wc})
	assert.NoError(t, err)

	dirLen := len(w.GetAllFiles(testServiceName))
	// should be 2 files (2 from root dir), filtered other
	if dirLen != 2 {
		t.Fatalf("incorrect directories len, len is: %d", dirLen)
	}

	go func() {
		go func() {
			time.Sleep(time.Second)
			err := copyDir(nestedDir, filepath.Join(tempDir, "copyTo"))
			assert.NoError(t, err)
		}()

		go func() {
			for range w.Event {
				// here should be event, otherwise we won't stop
				w.Stop()
			}
		}()
	}()

	err = w.StartPolling(time.Second)
	if err != nil {
		t.Fatal(err)
	}
}

func limitTime(d time.Duration, name string, free chan struct{}) {
	go func() {
		ticket := time.NewTicker(d)
		for {
			select {
			case <-ticket.C:
				ticket.Stop()
				panic("timeout exceed, test: " + name)
			case <-free:
				ticket.Stop()
				return
			}
		}
	}()
}

func copyFile(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer func() {
		_ = in.Close()
	}()

	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		if e := out.Close(); e != nil {
			err = e
		}
	}()

	_, err = io.Copy(out, in)
	if err != nil {
		return
	}

	err = out.Sync()
	if err != nil {
		return
	}

	si, err := os.Stat(src)
	if err != nil {
		return
	}
	err = os.Chmod(dst, si.Mode())
	if err != nil {
		return
	}

	return
}

func copyDir(src string, dst string) (err error) {
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)

	si, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !si.IsDir() {
		return fmt.Errorf("source is not a directory")
	}

	_, err = os.Stat(dst)
	if err != nil && !os.IsNotExist(err) {
		return
	}
	if err == nil {
		return fmt.Errorf("destination already exists")
	}

	err = os.MkdirAll(dst, si.Mode())
	if err != nil {
		return
	}

	entries, err := ioutil.ReadDir(src)
	if err != nil {
		return
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			err = copyDir(srcPath, dstPath)
			if err != nil {
				return
			}
		} else {
			// Skip symlinks.
			if entry.Mode()&os.ModeSymlink != 0 {
				continue
			}

			err = copyFile(srcPath, dstPath)
			if err != nil {
				return
			}
		}
	}

	return
}

func freeResources(path string) error {
	return os.RemoveAll(path)
}
