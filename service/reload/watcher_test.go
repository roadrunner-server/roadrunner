package reload

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
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
		serviceName:  testServiceName,
		recursive:    false,
		directories:  []string{tempDir},
		filterHooks:  nil,
		files:        make(map[string]os.FileInfo),
		ignored:      nil,
		filePatterns: nil,
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
	defer func() {
		err = freeResources(tempDir)
		if err != nil {
			t.Fatal(err)
		}
	}()

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
		serviceName:  testServiceName,
		recursive:    false,
		directories:  []string{tempDir},
		filterHooks:  nil,
		files:        make(map[string]os.FileInfo),
		ignored:      nil,
		filePatterns: nil,
	}

	w, err := NewWatcher([]WatcherConfig{wc})
	if err != nil {
		t.Fatal(err)
	}

	// should be 3 files and directory
	if len(w.GetAllFiles(testServiceName)) != 4 {
		t.Fatal("incorrect directories len")
	}

	go func() {
		// time sleep is used here because StartPolling is blocking operation
		time.Sleep(time.Second * 5)
		err = ioutil.WriteFile(filepath.Join(tempDir, "file2.txt"),
			[]byte{1, 1, 1}, 0755)
		if err != nil {
			panic(err)
		}
		go func() {
			for e := range w.Event {
				if e.path != "file2.txt" {
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
	defer func() {
		err = freeResources(tempDir)
		if err != nil {
			t.Fatal(err)
		}
	}()

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
		serviceName: testServiceName,
		recursive:   false,
		directories: []string{tempDir},
		filterHooks: func(filename string, patterns []string) error {
			for i := 0; i < len(patterns); i++ {
				if strings.Contains(filename, patterns[i]) {
					return nil
				}
			}
			return ErrorSkip
		},
		files:        make(map[string]os.FileInfo),
		ignored:      nil,
		filePatterns: []string{"aaa", "bbb"},
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

	go func() {
		err = ioutil.WriteFile(filepath.Join(tempDir, "file3.txt"),
			[]byte{1, 1, 1}, 0755)
		if err != nil {
			panic(err)
		}
		go func() {
			for e := range w.Event {
				fmt.Println(e.info.Name())
				panic("handled event from filtered file")
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

// nested
// scenario
// create dir and nested dir
// make files with aaa, bbb and txt extensions, filter txt
// change not filtered file, handle event
func Test_Recursive_Support(t *testing.T) {
	tempDir, err := ioutil.TempDir(".", "")
	defer func() {
		err = freeResources(tempDir)
		if err != nil {
			t.Fatal(err)
		}
	}()

	nestedDir, err := ioutil.TempDir(tempDir, "/nested")
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

	err = ioutil.WriteFile(filepath.Join(nestedDir, "file3.txt"),
		[]byte{}, 0755)
	if err != nil {
		t.Fatal(err)
	}
	err = ioutil.WriteFile(filepath.Join(nestedDir, "file4.aaa"),
		[]byte{}, 0755)
	if err != nil {
		t.Fatal(err)
	}

	wc := WatcherConfig{
		serviceName: testServiceName,
		recursive:   true,
		directories: []string{tempDir},
		filterHooks: func(filename string, patterns []string) error {
			for i := 0; i < len(patterns); i++ {
				if strings.Contains(filename, patterns[i]) {
					return nil
				}
			}
			return ErrorSkip
		},
		files:        make(map[string]os.FileInfo),
		ignored:      nil,
		filePatterns: []string{"aaa", "bbb"},
	}

	w, err := NewWatcher([]WatcherConfig{wc})
	if err != nil {
		t.Fatal(err)
	}

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
		if err != nil {
			panic(err)
		}
		go func() {
			for e := range w.Event {
				if e.info.Name() != "file4.aaa" {
					panic("wrong handled event from watcher in nested dir")
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

func Test_Wrong_Dir(t *testing.T) {
	// no such file or directory
	wrongDir := "askdjfhaksdlfksdf"

	wc := WatcherConfig{
		serviceName: testServiceName,
		recursive:   true,
		directories: []string{wrongDir},
		filterHooks: func(filename string, patterns []string) error {
			for i := 0; i < len(patterns); i++ {
				if strings.Contains(filename, patterns[i]) {
					return nil
				}
			}
			return ErrorSkip
		},
		files:        make(map[string]os.FileInfo),
		ignored:      nil,
		filePatterns: []string{"aaa", "bbb"},
	}

	_, err := NewWatcher([]WatcherConfig{wc})
	if err == nil {
		t.Fatal(err)
	}
}

func Test_Filter_Directory(t *testing.T) {
	tempDir, err := ioutil.TempDir(".", "")
	defer func() {
		err = freeResources(tempDir)
		if err != nil {
			t.Fatal(err)
		}
	}()

	nestedDir, err := ioutil.TempDir(tempDir, "/nested")
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

	err = ioutil.WriteFile(filepath.Join(nestedDir, "file3.txt"),
		[]byte{}, 0755)
	if err != nil {
		t.Fatal(err)
	}
	err = ioutil.WriteFile(filepath.Join(nestedDir, "file4.aaa"),
		[]byte{}, 0755)
	if err != nil {
		t.Fatal(err)
	}

	ignored, err := ConvertIgnored([]string{nestedDir})
	if err != nil {
		t.Fatal(err)
	}
	wc := WatcherConfig{
		serviceName: testServiceName,
		recursive:   true,
		directories: []string{tempDir},
		filterHooks: func(filename string, patterns []string) error {
			for i := 0; i < len(patterns); i++ {
				if strings.Contains(filename, patterns[i]) {
					return nil
				}
			}
			return ErrorSkip
		},
		files:        make(map[string]os.FileInfo),
		ignored:      ignored,
		filePatterns: []string{"aaa", "bbb", "txt"},
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
		// time sleep is used here because StartPolling is blocking operation
		time.Sleep(time.Second * 5)
		// change file in nested directory
		err = ioutil.WriteFile(filepath.Join(nestedDir, "file4.aaa"),
			[]byte{1, 1, 1}, 0755)
		if err != nil {
			panic(err)
		}
		go func() {
			for range w.Event {
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

func freeResources(path string) error {
	return os.RemoveAll(path)
}
