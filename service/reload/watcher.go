package reload

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var ErrorSkip = errors.New("file is skipped")
var NoWalkerConfig = errors.New("should add at least one walker config, when reload is set to true")

// SimpleHook is used to filter by simple criteria, CONTAINS
type SimpleHook func(filename string, pattern []string) error

// An Event describes an event that is received when files or directory
// changes occur. It includes the os.FileInfo of the changed file or
// directory and the type of event that's occurred and the full path of the file.
type Event struct {
	path string
	info os.FileInfo

	service string // type of service, http, grpc, etc...
}

type WatcherConfig struct {
	// service name
	serviceName string

	// recursive or just add by singe directory
	recursive bool

	// directories used per-service
	directories []string

	// simple hook, just CONTAINS
	filterHooks func(filename string, pattern []string) error

	// path to file with files
	files map[string]os.FileInfo

	// ignored directories, used map for O(1) amortized get
	ignored map[string]struct{}

	// filePatterns to ignore
	filePatterns []string
}

type Watcher struct {
	// main event channel
	Event chan Event
	close chan struct{}

	//=============================
	mu *sync.Mutex

	// indicates is walker started or not
	started bool

	// config for each service
	// need pointer here to assign files
	watcherConfigs map[string]WatcherConfig
}

// Options is used to set Watcher Options
type Options func(*Watcher)

// NewWatcher returns new instance of File Watcher
func NewWatcher(configs []WatcherConfig, options ...Options) (*Watcher, error) {
	w := &Watcher{
		Event: make(chan Event),
		mu:    &sync.Mutex{},

		close: make(chan struct{}),

		//workingDir:     workDir,
		watcherConfigs: make(map[string]WatcherConfig),
	}

	// add watcherConfigs by service names
	for _, v := range configs {
		w.watcherConfigs[v.serviceName] = v
	}

	// apply options
	for _, option := range options {
		option(w)
	}
	err := w.initFs()
	if err != nil {
		return nil, err
	}

	return w, nil
}

// initFs makes initial map with files
func (w *Watcher) initFs() error {
	for srvName, config := range w.watcherConfigs {
		fileList, err := w.retrieveFileList(srvName, config)
		if err != nil {
			return err
		}
		// workaround. in golang you can't assign to map in struct field
		tmp := w.watcherConfigs[srvName]
		tmp.files = fileList
		w.watcherConfigs[srvName] = tmp
	}
	return nil
}

// ConvertIgnored is used to convert slice to map with ignored files
func ConvertIgnored(ignored []string) (map[string]struct{}, error) {
	if len(ignored) == 0 {
		return nil, nil
	}

	ign := make(map[string]struct{}, len(ignored))
	for i := 0; i < len(ignored); i++ {
		abs, err := filepath.Abs(ignored[i])
		if err != nil {
			return nil, err
		}
		ign[abs] = struct{}{}
	}

	return ign, nil

}

// GetAllFiles returns all files initialized for particular company
func (w *Watcher) GetAllFiles(serviceName string) []os.FileInfo {
	var ret []os.FileInfo

	for _, v := range w.watcherConfigs[serviceName].files {
		ret = append(ret, v)
	}

	return ret
}

// https://en.wikipedia.org/wiki/Inotify
// SetMaxFileEvents sets max file notify events for Watcher
// In case of file watch errors, this value can be increased system-wide
// For linux: set --> fs.inotify.max_user_watches = 600000 (under /etc/<choose_name_here>.conf)
// Add apply: sudo sysctl -p --system
//func SetMaxFileEvents(events int) Options {
//	return func(watcher *Watcher) {
//		watcher.maxFileWatchEvents = events
//	}
//
//}

// pass map from outside
func (w *Watcher) retrieveFilesSingle(serviceName, path string) (map[string]os.FileInfo, error) {
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

	fileInfoList, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	// recursive calls are slow in compare to goto
	// so, we will add files with goto pattern
outer:
	for i := 0; i < len(fileInfoList); i++ {
		// if file in ignored --> continue
		if _, ignored := w.watcherConfigs[serviceName].ignored[path]; ignored {
			continue
		}

		// if filename does not contain pattern --> ignore that file
		if w.watcherConfigs[serviceName].filePatterns != nil && w.watcherConfigs[serviceName].filterHooks != nil {
			err = w.watcherConfigs[serviceName].filterHooks(fileInfoList[i].Name(), w.watcherConfigs[serviceName].filePatterns)
			if err == ErrorSkip {
				continue outer
			}
		}

		filesList[fileInfoList[i].Name()] = fileInfoList[i]
	}

	return filesList, nil
}

func (w *Watcher) StartPolling(duration time.Duration) error {
	w.mu.Lock()
	if w.started {
		w.mu.Unlock()
		return errors.New("already started")
	}

	w.started = true
	w.mu.Unlock()

	return w.waitEvent(duration)
}

// this is blocking operation
func (w *Watcher) waitEvent(d time.Duration) error {
	ticker := time.NewTicker(d)
	for {
		select {
		case <-w.close:
			ticker.Stop()
			// just exit
			// no matter for the pollEvents
			return nil
		case <-ticker.C:
			// this is not very effective way
			// because we have to wait on Lock
			// better is to listen files in parallel, but, since that would be used in debug... TODO
			for serviceName, config := range w.watcherConfigs {
				go func(sn string, c WatcherConfig) {
					fileList, _ := w.retrieveFileList(sn, c)
					w.pollEvents(c.serviceName, fileList)
				}(serviceName, config)
			}
		}
	}

}

// retrieveFileList get file list for service
func (w *Watcher) retrieveFileList(serviceName string, config WatcherConfig) (map[string]os.FileInfo, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	fileList := make(map[string]os.FileInfo)
	if config.recursive {
		// walk through directories recursively
		for _, dir := range config.directories {
			// full path is workdir/relative_path
			fullPath, err := filepath.Abs(dir)
			if err != nil {
				return nil, err
			}
			list, err := w.retrieveFilesRecursive(serviceName, fullPath)
			if err != nil {
				return nil, err
			}

			for k, v := range list {
				fileList[k] = v
			}
		}
		return fileList, nil
	}

	for _, dir := range config.directories {
		// full path is workdir/relative_path
		fullPath, err := filepath.Abs(dir)
		if err != nil {
			return nil, err
		}

		// list is pathToFiles with files
		list, err := w.retrieveFilesSingle(serviceName, fullPath)
		if err != nil {
			return nil, err
		}

		for pathToFile, file := range list {
			fileList[pathToFile] = file
		}
	}

	return fileList, nil
}

func (w *Watcher) retrieveFilesRecursive(serviceName, root string) (map[string]os.FileInfo, error) {
	fileList := make(map[string]os.FileInfo)

	return fileList, filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// If path is ignored and it's a directory, skip the directory. If it's
		// ignored and it's a single file, skip the file.
		_, ignored := w.watcherConfigs[serviceName].ignored[path]
		if ignored {
			if info.IsDir() {
				// if it's dir, ignore whole
				return filepath.SkipDir
			}
			return nil
		}

		// if filename does not contain pattern --> ignore that file
		err = w.watcherConfigs[serviceName].filterHooks(info.Name(), w.watcherConfigs[serviceName].filePatterns)
		if err == ErrorSkip {
			return nil
		}

		// Add the path and it's info to the file list.
		fileList[path] = info
		return nil
	})
}

func (w *Watcher) pollEvents(serviceName string, files map[string]os.FileInfo) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Store create and remove events for use to check for rename events.
	creates := make(map[string]os.FileInfo)
	removes := make(map[string]os.FileInfo)

	// Check for removed files.
	for pth, info := range w.watcherConfigs[serviceName].files {
		if _, found := files[pth]; !found {
			removes[pth] = info
		}
	}

	// Check for created files, writes and chmods.
	for pth, info := range files {
		if info.IsDir() {
			continue
		}
		oldInfo, found := w.watcherConfigs[serviceName].files[pth]
		if !found {
			// A file was created.
			creates[pth] = info
			continue
		}
		if oldInfo.ModTime() != info.ModTime() {
			w.watcherConfigs[serviceName].files[pth] = info
			w.Event <- Event{
				path:    pth,
				info:    info,
				service: serviceName,
			}
		}
		if oldInfo.Mode() != info.Mode() {
			w.watcherConfigs[serviceName].files[pth] = info
			w.Event <- Event{
				path:    pth,
				info:    info,
				service: serviceName,
			}
		}
	}

	//Check for renames and moves.
	for path1, info1 := range removes {
		for path2, info2 := range creates {
			if sameFile(info1, info2) {
				e := Event{
					path:    path2,
					info:    info2,
					service: serviceName,
				}

				// remove initial path
				delete(w.watcherConfigs[serviceName].files, path1)
				// update with new
				w.watcherConfigs[serviceName].files[path2] = info2


				w.Event <- e
			}
		}
	}

	//Send all the remaining create and remove events.
	for pth, info := range creates {
		w.watcherConfigs[serviceName].files[pth] = info
		w.Event <- Event{
			path:    pth,
			info:    info,
			service: serviceName,
		}
	}
	for pth, info := range removes {
		delete(w.watcherConfigs[serviceName].files, pth)
		w.Event <- Event{
			path:    pth,
			info:    info,
			service: serviceName,
		}
	}
}

func (w *Watcher) Stop() {
	w.close <- struct{}{}
}
