package reload

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/plugins/logger"
)

// SimpleHook is used to filter by simple criteria, CONTAINS
type SimpleHook func(filename string, pattern []string) error

// An Event describes an event that is received when files or directory
// changes occur. It includes the os.FileInfo of the changed file or
// directory and the type of event that's occurred and the full path of the file.
type Event struct {
	Path string
	Info os.FileInfo

	service string // type of service, http, grpc, etc...
}

type WatcherConfig struct {
	// service name
	ServiceName string

	// Recursive or just add by singe directory
	Recursive bool

	// Directories used per-service
	Directories []string

	// simple hook, just CONTAINS
	FilterHooks func(filename string, pattern []string) error

	// path to file with Files
	Files map[string]os.FileInfo

	// Ignored Directories, used map for O(1) amortized get
	Ignored map[string]struct{}

	// FilePatterns to ignore
	FilePatterns []string
}

type Watcher struct {
	// main event channel
	Event chan Event
	close chan struct{}

	// =============================
	mu *sync.Mutex

	// indicates is walker started or not
	started bool

	// config for each service
	// need pointer here to assign files
	watcherConfigs map[string]WatcherConfig

	// logger
	log logger.Logger
}

// Options is used to set Watcher Options
type Options func(*Watcher)

// NewWatcher returns new instance of File Watcher
func NewWatcher(configs []WatcherConfig, log logger.Logger, options ...Options) (*Watcher, error) {
	w := &Watcher{
		Event: make(chan Event),
		mu:    &sync.Mutex{},

		log: log,

		close: make(chan struct{}),

		//workingDir:     workDir,
		watcherConfigs: make(map[string]WatcherConfig),
	}

	// add watcherConfigs by service names
	for _, v := range configs {
		w.watcherConfigs[v.ServiceName] = v
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
	const op = errors.Op("watcher_init_fs")
	for srvName, config := range w.watcherConfigs {
		fileList, err := w.retrieveFileList(srvName, config)
		if err != nil {
			return errors.E(op, err)
		}
		// workaround. in golang you can't assign to map in struct field
		tmp := w.watcherConfigs[srvName]
		tmp.Files = fileList
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

// https://en.wikipedia.org/wiki/Inotify
// SetMaxFileEvents sets max file notify events for Watcher
// In case of file watch errors, this value can be increased system-wide
// For linux: set --> fs.inotify.max_user_watches = 600000 (under /etc/<choose_name_here>.conf)
// Add apply: sudo sysctl -p --system
// func SetMaxFileEvents(events int) Options {
//	return func(watcher *Watcher) {
//		watcher.maxFileWatchEvents = events
//	}
//
// }

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
		if _, ignored := w.watcherConfigs[serviceName].Ignored[path]; ignored {
			continue
		}

		// if filename does not contain pattern --> ignore that file
		if w.watcherConfigs[serviceName].FilePatterns != nil && w.watcherConfigs[serviceName].FilterHooks != nil {
			err = w.watcherConfigs[serviceName].FilterHooks(fileInfoList[i].Name(), w.watcherConfigs[serviceName].FilePatterns)
			if errors.Is(errors.SkipFile, err) {
				continue outer
			}
		}

		filesList[fileInfoList[i].Name()] = fileInfoList[i]
	}

	return filesList, nil
}

func (w *Watcher) StartPolling(duration time.Duration) error {
	w.mu.Lock()
	const op = errors.Op("watcher_start_polling")
	if w.started {
		w.mu.Unlock()
		return errors.E(op, errors.Str("already started"))
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
			// better is to listen files in parallel, but, since that would be used in debug...
			for serviceName := range w.watcherConfigs {
				fileList, _ := w.retrieveFileList(serviceName, w.watcherConfigs[serviceName])
				w.pollEvents(w.watcherConfigs[serviceName].ServiceName, fileList)
			}
		}
	}
}

// retrieveFileList get file list for service
func (w *Watcher) retrieveFileList(serviceName string, config WatcherConfig) (map[string]os.FileInfo, error) {
	fileList := make(map[string]os.FileInfo)
	if config.Recursive {
		// walk through directories recursively
		for i := 0; i < len(config.Directories); i++ {
			// full path is workdir/relative_path
			fullPath, err := filepath.Abs(config.Directories[i])
			if err != nil {
				return nil, err
			}
			list, err := w.retrieveFilesRecursive(serviceName, fullPath)
			if err != nil {
				return nil, err
			}

			for k := range list {
				fileList[k] = list[k]
			}
		}
		return fileList, nil
	}

	for i := 0; i < len(config.Directories); i++ {
		// full path is workdir/relative_path
		fullPath, err := filepath.Abs(config.Directories[i])
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
		const op = errors.Op("retrieve files recursive")
		if err != nil {
			return errors.E(op, err)
		}

		// If path is ignored and it's a directory, skip the directory. If it's
		// ignored and it's a single file, skip the file.
		_, ignored := w.watcherConfigs[serviceName].Ignored[path]
		if ignored {
			if info.IsDir() {
				// if it's dir, ignore whole
				return filepath.SkipDir
			}
			return nil
		}

		// if filename does not contain pattern --> ignore that file
		err = w.watcherConfigs[serviceName].FilterHooks(info.Name(), w.watcherConfigs[serviceName].FilePatterns)
		if errors.Is(errors.SkipFile, err) {
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

	// InsertMany create and remove events for use to check for rename events.
	creates := make(map[string]os.FileInfo)
	removes := make(map[string]os.FileInfo)

	// Check for removed files.
	for pth := range w.watcherConfigs[serviceName].Files {
		if _, found := files[pth]; !found {
			removes[pth] = w.watcherConfigs[serviceName].Files[pth]
			w.log.Debug("file added to the list of removed files", "path", pth, "name", w.watcherConfigs[serviceName].Files[pth].Name(), "size", w.watcherConfigs[serviceName].Files[pth].Size())
		}
	}

	// Check for created files, writes and chmods.
	for pth := range files {
		if files[pth].IsDir() {
			continue
		}
		oldInfo, found := w.watcherConfigs[serviceName].Files[pth]
		if !found {
			// A file was created.
			creates[pth] = files[pth]
			w.log.Debug("file was created", "path", pth, "name", files[pth].Name(), "size", files[pth].Size())
			continue
		}

		if oldInfo.ModTime() != files[pth].ModTime() || oldInfo.Mode() != files[pth].Mode() {
			w.watcherConfigs[serviceName].Files[pth] = files[pth]
			w.log.Debug("file was updated", "path", pth, "name", files[pth].Name(), "size", files[pth].Size())
			w.Event <- Event{
				Path:    pth,
				Info:    files[pth],
				service: serviceName,
			}
		}
	}

	// Send all the remaining create and remove events.
	for pth := range creates {
		// add file to the plugin watch files
		w.watcherConfigs[serviceName].Files[pth] = creates[pth]
		w.log.Debug("file was added to watcher", "path", pth, "name", creates[pth].Name(), "size", creates[pth].Size())

		w.Event <- Event{
			Path:    pth,
			Info:    creates[pth],
			service: serviceName,
		}
	}

	for pth := range removes {
		// delete path from the config
		delete(w.watcherConfigs[serviceName].Files, pth)
		w.log.Debug("file was removed from watcher", "path", pth, "name", removes[pth].Name(), "size", removes[pth].Size())

		w.Event <- Event{
			Path:    pth,
			Info:    removes[pth],
			service: serviceName,
		}
	}
}

func (w *Watcher) Stop() {
	w.close <- struct{}{}
}
