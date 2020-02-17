package reload

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"
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
	ops                map[Op]struct{} // Op filtering.
	files              map[string]string //files by service, http, grpc, etc..
	ignored            map[string]string //ignored files or directories
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

// Add
// name will be
func (w *Watcher) Add(name string) error {
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
	err = w.addDirectoryContent(" ", fileList)
	if err != nil {
		return err
	}


}

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




}

func (w *Watcher) search(map[string]os.FileInfo) error {

}

































