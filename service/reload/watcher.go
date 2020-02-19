package reload

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
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

	Type string // type of event, http, grpc, etc...
}

type Watcher struct {
	Event  chan Event
	errors chan error
	close  chan struct{}
	Closed chan struct{}

	mu *sync.Mutex
	//wg *sync.WaitGroup

	filterHooks []FilterFileHookFunc

	started bool // indicates is walker started or not

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
		Event: make(chan Event),
		mu:    &sync.Mutex{},
		//wg:         &sync.WaitGroup{},
		Closed:     make(chan struct{}),
		close:      make(chan struct{}),
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
	//fileList := make(map[string]os.FileInfo, 10)
	fileList, err := w.retrieveSingleDirectoryContent(name)
	if err != nil {
		return err
	}

	for k, v := range fileList {
		w.files[k] = v
	}

	return nil

}

func (w *Watcher) AddRecursive(name string) error {
	name, err := filepath.Abs(name)
	if err != nil {
		return err
	}

	filesList := make(map[string]os.FileInfo, 10)

	err = w.retrieveFilesRecursive(name, filesList)
	if err != nil {
		return err
	}

	for k, v := range filesList {
		w.files[k] = v
	}

	return nil
}

// pass map from outside
func (w *Watcher) retrieveSingleDirectoryContent(path string) (map[string]os.FileInfo, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	filesList := make(map[string]os.FileInfo, 10)

	filesList[path] = stat

	// if it's not a dir, return
	if !stat.IsDir() {
		return filesList, nil
	}

	//err = filepath.Walk(name, func(path string, info os.FileInfo, err error) error {
	//	if info.IsDir() {
	//		return nil
	//	}
	//
	//	fileList[path] = info
	//
	//	return nil
	//})
	//
	//if err != nil {
	//	return err
	//}

	fileInfoList, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	// recursive calls are slow in compare to goto
	// so, we will add files with goto pattern

outer:
	for i := 0; i < len(fileInfoList); i++ {
		var path string
		// BCE check elimination
		// https://go101.org/article/bounds-check-elimination.html
		if len(fileInfoList) != 0 && len(fileInfoList) >= i {
			path = filepath.Join(path, fileInfoList[i].Name())
		} else {
			return nil, errors.New("file info list len")
		}

		// if file in ignored --> continue
		if _, ignored := w.ignored[path]; ignored {
			continue
		}

		for _, fh := range w.filterHooks {
			err := fh(fileInfoList[i], path)
			if err != nil {
				// if err is not nil, move to the start of the cycle since the path not match the hook
				continue outer
			}
		}

		filesList[path] = fileInfoList[i]

	}

	return filesList, nil
}

func (w *Watcher) StartPolling(duration time.Duration) error {
	if duration < time.Second {
		return errors.New("too short duration, please use at least 1 second")
	}

	w.mu.Lock()
	if w.started {
		w.mu.Unlock()
		return errors.New("already started")
	}

	w.started = true
	w.mu.Unlock()

	//w.wg.Done()

	return w.waitEvent(duration)
}

// this is blocking operation
func (w *Watcher) waitEvent(d time.Duration) error {
	for {
		// done lets the inner polling cycle loop know when the
		// current cycle's method has finished executing.
		//done := make(chan struct{})

		// Any events that are found are first piped to evt before
		// being sent to the main Event channel.
		//evt := make(chan Event)

		// Retrieve the file list for all watched file's and dirs.
		//fileList := w.files

		// cancel can be used to cancel the current event polling function.
		cancel := make(chan struct{})

		// Look for events.
		//go func() {
		//	w.pollEvents(w.files, evt, cancel)
		//	done <- struct{}{}
		//}()

		// numEvents holds the number of events for the current cycle.
		//numEvents := 0

		ticker := time.NewTicker(d)
		for {
			select {
			case <-w.close:
				close(cancel)
				close(w.Closed)
				return nil
			case <-ticker.C:
				//fileList := make(map[string]os.FileInfo, 100)
				//w.mu.Lock()
				fileList, _ := w.retrieveFileList(w.workingDir, false)
				w.pollEvents(fileList, cancel)
				//w.mu.Unlock()
			default:

			}
		}

		ticker.Stop()
		//inner:
		//	for {
		//		select {
		//		case <-w.close:
		//			close(cancel)
		//			close(w.Closed)
		//			return nil
		//		case event := <-evt:
		//			//if len(w.operations) > 0 { // Filter Ops.
		//			//	_, found := w.operations[event.Op]
		//			//	if !found {
		//			//		continue
		//			//	}
		//			//}
		//			//numEvents++
		//			//if w.maxFileWatchEvents > 0 && numEvents > w.maxFileWatchEvents {
		//			//	close(cancel)
		//			//	break inner
		//			//}
		//			w.Event <- event
		//		case <-done: // Current cycle is finished.
		//			break inner
		//		}
		//	}

		//// Update the file's list.
		//w.mu.Lock()
		//w.files = fileList
		//w.mu.Unlock()

		//time.Sleep(d)
		//sleepLoop:
		//	for {
		//		select {
		//		case <-w.close:
		//			close(cancel)
		//			return nil
		//		case <-time.After(d):
		//			break sleepLoop
		//		}
		//	} //end Sleep for
	}
}

func (w *Watcher) retrieveFileList(path string, recursive bool) (map[string]os.FileInfo, error) {

	//fileList := make(map[string]os.FileInfo)

	//list := make(map[string]os.FileInfo, 100)
	//var err error

	if recursive {
		//fileList, err := w.retrieveFilesRecursive(path)
		//if err != nil {
		//if os.IsNotExist(err) {
		//	w.mu.Unlock()
		//	// todo path error
		//	_, ok := err.(*os.PathError)
		//	if ok {
		//		w.RemoveRecursive(path)
		//	}
		//	w.mu.Lock()
		//} else {
		//	w.errors <- err
		//}
		//}

		//for k, v := range fileList {
		//	fileList[k] = v
		//}
		//return fileList, nil
		return nil, nil
	} else {
		fileList, err := w.retrieveSingleDirectoryContent(path)
		if err != nil {
			//if os.IsNotExist(err) {
			//	w.mu.Unlock()
			//	_, ok := err.(*os.PathError)
			//	if ok {
			//		w.RemoveRecursive(path)
			//	}
			//	w.mu.Lock()
			//} else {
			//	w.errors <- err
			//}
		}

		for k, v := range fileList {
			fileList[k] = v
		}
		return fileList, nil
	}
	// Add the file's to the file list.

	//return nil
}

// RemoveRecursive removes either a single file or a directory recursively from
// the file's list.
func (w *Watcher) RemoveRecursive(name string) (err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	name, err = filepath.Abs(name)
	if err != nil {
		return err
	}

	// If name is a single file, remove it and return.
	info, found := w.files[name]
	if !found {
		return nil // Doesn't exist, just return.
	}
	if !info.IsDir() {
		delete(w.files, name)
		return nil
	}

	// If it's a directory, delete all of it's contents recursively
	// from w.files.
	for path := range w.files {
		if strings.HasPrefix(path, name) {
			delete(w.files, path)
		}
	}
	return nil
}

func (w *Watcher) retrieveFilesRecursive(name string, fileList map[string]os.FileInfo) error {
	return filepath.Walk(name, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		for _, f := range w.filterHooks {
			err := f(info, path)
			if err == ErrorSkip {
				return nil
			}
			if err != nil {
				return err
			}
		}

		// If path is ignored and it's a directory, skip the directory. If it's
		// ignored and it's a single file, skip the file.
		_, ignored := w.ignored[path]

		if ignored {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		// Add the path and it's info to the file list.
		fileList[path] = info
		return nil
	})
}

// Wait blocks until the watcher is started.
//func (w *Watcher) Wait() {
//	w.wg.Wait()
//}

func (w *Watcher) pollEvents(files map[string]os.FileInfo, cancel chan struct{}) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Store create and remove events for use to check for rename events.
	creates := make(map[string]os.FileInfo)
	removes := make(map[string]os.FileInfo)

	// Check for removed files.
	for path, info := range w.files {
		if _, found := files[path]; !found {
			removes[path] = info
		}
	}

	// Check for created files, writes and chmods.
	for path, info := range files {
		oldInfo, found := w.files[path]
		if !found {
			// A file was created.
			creates[path] = info
			continue
		}
		if oldInfo.ModTime() != info.ModTime() {
			w.files[path] = info
			select {
			case <-cancel:
				return
			case w.Event <- Event{Write, path, path, info, "http"}:
			}
		}
		if oldInfo.Mode() != info.Mode() {
			select {
			case <-cancel:
				return
			case w.Event <- Event{Chmod, path, path, info, "http"}:
			}
		}
	}

	// Check for renames and moves.
	for path1, info1 := range removes {
		for path2, info2 := range creates {
			if sameFile(info1, info2) {
				e := Event{
					Op:       Move,
					Path:     path2,
					OldPath:  path1,
					FileInfo: info1,
				}
				// If they are from the same directory, it's a rename
				// instead of a move event.
				if filepath.Dir(path1) == filepath.Dir(path2) {
					e.Op = Rename
				}

				delete(removes, path1)
				delete(creates, path2)

				select {
				case <-cancel:
					return
				case w.Event <- e:
				}
			}
		}
	}

	// Send all the remaining create and remove events.
	for path, info := range creates {
		select {
		case <-cancel:
			return
		case w.Event <- Event{Create, path, "", info, "http"}:
		}
	}
	for path, info := range removes {
		select {
		case <-cancel:
			return
		case w.Event <- Event{Remove, path, path, info, "http"}:
		}
	}
}
