package reload

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"syscall"
)

// Config is a Reload configuration point.
//type Config struct {
// Enable or disable Reload extension, default disable.
//Enabled bool
//
// Watch is general pattern of files to watch. It will be applied to every directory in project
//Watch []string
//
// Services is set of services which would be reloaded in case of FS changes
//Services map[string]ServiceConfig
//}
//
//type ServiceConfig struct {
// Watch is per-service specific files to watch
//Watch []string
// Dirs is per-service specific dirs which will be combined with Watch
//Dirs  []string
// Ignore is set of files which would not be watched
//Ignore []string
//}

// An Op is a type that is used to describe what type
// of event has occurred during the watching process.
type Op uint32

// Ops
const (
	Create Op = iota
	Write
	Remove
	Rename
	Chmod
	Move
)

var ops = map[Op]string{
	Create: "CREATE",
	Write:  "WRITE",
	Remove: "REMOVE",
	Rename: "RENAME",
	Chmod:  "CHMOD",
	Move:   "MOVE",
}

var ErrorSkip = errors.New("file is skipped")

// FilterFileHookFunc is a function that is called to filter files during listings.
// If a file is ok to be listed, nil is returned otherwise ErrSkip is returned.
type FilterFileHookFunc func(info os.FileInfo, fullPath string) error

// RegexFilterHook is a function that accepts or rejects a file
// for listing based on whether it's filename or full path matches
// a regular expression.
func RegexFilterHook(r *regexp.Regexp, useFullPath bool) FilterFileHookFunc {
	return func(info os.FileInfo, fullPath string) error {
		str := info.Name()

		if useFullPath {
			str = fullPath
		}

		// Match
		if r.MatchString(str) {
			return nil
		}

		// No match.
		return ErrorSkip
	}
}

func SetFileHooks(fileHook ...FilterFileHookFunc) Options {
	return func(watcher *Watcher) {
		watcher.filterHooks = fileHook
	}
}

// An Event describes an event that is received when files or directory
// changes occur. It includes the os.FileInfo of the changed file or
// directory and the type of event that's occurred and the full path of the file.
type Event struct {
	Op
	Path    string
	OldPath string
	os.FileInfo
}

type Watcher struct {
	Event  chan Event
	errors chan error
	wg     *sync.WaitGroup

	filterHooks []FilterFileHookFunc

	workingDir         string
	maxFileWatchEvents int
	operations         map[Op]struct{}        // Op filtering.
	files              map[string]os.FileInfo //files by service, http, grpc, etc..
	ignored            map[string]string      //ignored files or directories
}

// Options is used to set Watcher Options
type Options func(*Watcher)

func NewWatcher(options ...Options) (*Watcher, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	w := &Watcher{
		workingDir: dir,
		operations: make(map[Op]struct{}),
		files:      make(map[string]os.FileInfo),
		ignored:    make(map[string]string),
	}

	for _, option := range options {
		option(w)
	}

	// dir --> /home/valery/Projects/opensource/roadrunner
	return w, nil
}

// https://en.wikipedia.org/wiki/Inotify
// SetMaxFileEvents sets max file notify events for Watcher
// In case of file watch errors, this value can be increased system-wide
// For linux: set --> fs.inotify.max_user_watches = 600000 (under /etc/<choose_name_here>.conf)
// Add apply: sudo sysctl -p --system
func SetMaxFileEvents(events int) Options {
	return func(watcher *Watcher) {
		watcher.maxFileWatchEvents = events
	}

}

// SetDefaultRootPath is used to set own root path for adding files
func SetDefaultRootPath(path string) Options {
	return func(watcher *Watcher) {
		watcher.workingDir = path
	}
}

// Add
// name will be current working dir
func (w *Watcher) AddSingle(name string) error {
	name, err := filepath.Abs(name)
	if err != nil {

	}

	// Ignored files
	// map is to have O(1) when search for file
	_, ignored := w.ignored[name]
	if ignored {
		return nil
	}

	// small optimization for smallvector
	fileList := make(map[string]os.FileInfo, 10)
	err = w.addDirectoryContent(name, fileList)
	if err != nil {
		return err
	}

	for k, v := range fileList {
		w.files[k] = v
	}

	return nil

}

// pass map from outside
func (w *Watcher) addDirectoryContent(name string, filelist map[string]os.FileInfo) error {
	fileInfo, err := os.Stat(name)
	if err != nil {
		return err
	}

	filelist[name] = fileInfo

	// if it's not a dir, return
	if !fileInfo.IsDir() {
		return nil
	}

	fileInfoList, err := ioutil.ReadDir(name)
	if err != nil {
		return err
	}

	// recursive calls are slow in compare to goto
	// so, we will add files with goto pattern

outer:
	for i := 0; i < len(fileInfoList); i++ {
		var path string
		// BCE check elimination
		// https://go101.org/article/bounds-check-elimination.html
		if len(fileInfoList) != 0 && len(fileInfoList) >= i {
			path = filepath.Join(name, fileInfoList[i].Name())
		} else {
			return errors.New("file info list len")
		}

		// if file in ignored --> continue
		if _, ignored := w.ignored[name]; ignored {
			continue
		}

		for _, fh := range w.filterHooks {
			err := fh(fileInfo, path)
			if err != nil {
				// if err is not nil, move to the start of the cycle since the path not match the hook
				continue outer
			}
		}

		filelist[path] = fileInfo

	}

	return nil
}
