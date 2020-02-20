package reload

import (
	"errors"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sync"
	"time"
)

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
var NoWalkerConfig = errors.New("should add at least one walker config, when reload is set to true")

// SimpleHook is used to filter by simple criteria, CONTAINS
type SimpleHook func(filename, pattern string) error

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

type WatcherConfig struct {
	// service name
	serviceName string
	// recursive or just add by singe directory
	recursive bool
	// directories used per-service
	directories []string
	// simple hook, just CONTAINS
	filterHooks SimpleHook
	// path to file with files
	files map[string]os.FileInfo
	//  //ignored files or directories, used map for O(1) amortized get
	ignored map[string]string
}

type Watcher struct {
	// main event channel
	Event chan Event

	errors chan error
	close  chan struct{}
	Closed chan struct{}
	//=============================
	mu *sync.Mutex
	wg *sync.WaitGroup

	// indicates is walker started or not
	started bool
	// working directory, same for all
	workingDir string

	// operation type
	operations map[Op]struct{} // Op filtering.

	// config for each service
	// need pointer here to assign files
	watcherConfigs map[string]WatcherConfig
}

// Options is used to set Watcher Options
type Options func(*Watcher)

func NewWatcher(configs []WatcherConfig, options ...Options) (*Watcher, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	w := &Watcher{
		Event: make(chan Event),
		mu:    &sync.Mutex{},
		wg:    &sync.WaitGroup{},

		Closed: make(chan struct{}),
		close:  make(chan struct{}),

		workingDir:     dir,
		operations:     make(map[Op]struct{}),
		watcherConfigs: make(map[string]WatcherConfig),
	}

	// add watcherConfigs by service names
	for _, v := range configs {
		w.watcherConfigs[v.serviceName] = v
	}

	for _, option := range options {
		option(w)
	}

	if w.watcherConfigs == nil {
		return nil, NoWalkerConfig
	}

	err = w.initFs()
	if err != nil {
		return nil, err
	}

	return w, nil
}

func (w *Watcher) initFs() error {
	for srvName, config := range w.watcherConfigs {
		fileList, err := w.retrieveFileList(srvName, config)
		if err != nil {
			return err
		}
		tmp := w.watcherConfigs[srvName]
		tmp.files = fileList
		w.watcherConfigs[srvName] = tmp
	}
	return nil
}

func (w *Watcher) AddWatcherConfig(config WatcherConfig) {
	w.watcherConfigs[config.serviceName] = config
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

// SetDefaultRootPath is used to set own root path for adding files
func SetDefaultRootPath(path string) Options {
	return func(watcher *Watcher) {
		watcher.workingDir = path
	}
}

// Add
// name will be current working dir
func (w *Watcher) AddSingle(serviceName, relPath string) error {
	absPath, err := filepath.Abs(w.workingDir)
	if err != nil {
		return err
	}

	// full path is workdir/relative_path
	fullPath := path.Join(absPath, relPath)

	// Ignored files
	// map is to have O(1) when search for file
	_, ignored := w.watcherConfigs[serviceName].ignored[fullPath]
	if ignored {
		return nil
	}

	// small optimization for smallvector
	//fileList := make(map[string]os.FileInfo, 10)
	fileList, err := w.retrieveFilesSingle(serviceName, fullPath)
	if err != nil {
		return err
	}

	for fullPth, file := range fileList {
		w.watcherConfigs[serviceName].files[fullPth] = file
	}

	return nil
}

func (w *Watcher) AddRecursive(serviceName string, relPath string) error {
	workDirAbs, err := filepath.Abs(w.workingDir)
	if err != nil {
		return err
	}

	fullPath := path.Join(workDirAbs, relPath)

	filesList, err := w.retrieveFilesRecursive(serviceName, fullPath)
	if err != nil {
		return err
	}

	for pathToFile, file := range filesList {
		w.watcherConfigs[serviceName].files[pathToFile] = file
	}

	return nil
}

func (w *Watcher) AddIgnored(serviceName string, directories []string) error {
	workDirAbs, err := filepath.Abs(w.workingDir)
	if err != nil {
		return err
	}

	// concat wd with relative paths from config
	// todo check path for existance
	for _, v := range directories {
		fullPath := path.Join(workDirAbs, v)
		w.watcherConfigs[serviceName].ignored[fullPath] = fullPath
	}

	return nil
}

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
		var pathToFile string
		// BCE check elimination
		// https://go101.org/article/bounds-check-elimination.html
		if len(fileInfoList) != 0 && len(fileInfoList) >= i {
			pathToFile = filepath.Join(pathToFile, fileInfoList[i].Name())
		} else {
			return nil, errors.New("file info list len")
		}

		// if file in ignored --> continue
		if _, ignored := w.watcherConfigs[serviceName].ignored[path]; ignored {
			continue
		}

		err := w.watcherConfigs[serviceName].filterHooks(fileInfoList[i].Name(), pathToFile)
		if err != nil {
			// if err is not nil, move to the start of the cycle since the pathToFile not match the hook
			continue outer
		}

		filesList[pathToFile] = fileInfoList[i]

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

//func (w *Watcher) updatedFileListForConfig(config WatcherConfig) (map[string]os.FileInfo, error) {
//	if config.recursive {
//		return nil, nil
//	}
//
//	for _, v := range config.directories {
//		files, err := w.retrieveFilesSingle(path.Join(w.workingDir, v))
//		if err != nil {
//			return nil, err
//		}
//
//	}
//
//	return nil, nil
//}

// this is blocking operation
func (w *Watcher) waitEvent(d time.Duration) error {
	for {
		cancel := make(chan struct{})
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

				for serviceName, config := range w.watcherConfigs {
					go func(sn string, c WatcherConfig) {
						fileList, _ := w.retrieveFileList(sn, c)
						w.pollEvents(c.serviceName, fileList, cancel)
					}(serviceName, config)
				}
			default:

			}
		}
	}
}

func (w *Watcher) retrieveFileList(serviceName string, config WatcherConfig) (map[string]os.FileInfo, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	fileList := make(map[string]os.FileInfo)
	if config.recursive {
		// walk through directories recursively
		for _, dir := range config.directories {
			// full path is workdir/relative_path
			fullPath := path.Join(w.workingDir, dir)
			list, err := w.retrieveFilesRecursive(serviceName, fullPath)
			if err != nil {
				return nil, err
			}

			for k, v := range list {
				fileList[k] = v
			}
			return fileList, nil
		}
	}

	for _, dir := range config.directories {
		absPath, err := filepath.Abs(w.workingDir)
		if err != nil {
			return nil, err
		}

		// full path is workdir/relative_path
		fullPath := path.Join(absPath, dir)

		// list is pathToFiles with files
		list, err := w.retrieveFilesSingle(serviceName, fullPath)

		for pathToFile, file := range list {
			fileList[pathToFile] = file
		}
	}

	return fileList, nil

	// Add the file's to the file list.

	//return nil
}

// RemoveRecursive removes either a single file or a directory recursively from
// the file's list.
//func (w *Watcher) RemoveRecursive(name string) (err error) {
//	w.mu.Lock()
//	defer w.mu.Unlock()
//
//	name, err = filepath.Abs(name)
//	if err != nil {
//		return err
//	}
//
//	// If name is a single file, remove it and return.
//	info, found := w.files[name]
//	if !found {
//		return nil // Doesn't exist, just return.
//	}
//	if !info.IsDir() {
//		delete(w.files, name)
//		return nil
//	}
//
//	// If it's a directory, delete all of it's contents recursively
//	// from w.files.
//	for path := range w.files {
//		if strings.HasPrefix(path, name) {
//			delete(w.files, path)
//		}
//	}
//	return nil
//}

func (w *Watcher) retrieveFilesRecursive(serviceName, root string) (map[string]os.FileInfo, error) {
	fileList := make(map[string]os.FileInfo)

	return fileList, filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// filename, pattern TODO
		//err = w.watcherConfigs[serviceName].filterHooks(info.Name(), path)
		//if err == ErrorSkip {
		//	return nil
		//}
		//if err != nil {
		//	return err
		//}

		// If path is ignored and it's a directory, skip the directory. If it's
		// ignored and it's a single file, skip the file.
		_, ignored := w.watcherConfigs[serviceName].ignored[path]

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

func (w *Watcher) pollEvents(serviceName string, files map[string]os.FileInfo, cancel chan struct{}) {
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
			select {
			case <-cancel:
				return
			case w.Event <- Event{Write, pth, pth, info, serviceName}:
			}
		}
		if oldInfo.Mode() != info.Mode() {
			w.watcherConfigs[serviceName].files[pth] = info
			select {
			case <-cancel:
				return
			case w.Event <- Event{Chmod, pth, pth, info, serviceName}:
			}
		}
	}

	// Check for renames and moves.
	//for path1, info1 := range removes {
	//	for path2, info2 := range creates {
	//		if sameFile(info1, info2) {
	//			e := Event{
	//				Op:       Move,
	//				Path:     path2,
	//				OldPath:  path1,
	//				FileInfo: info1,
	//			}
	//			// If they are from the same directory, it's a rename
	//			// instead of a move event.
	//			if filepath.Dir(path1) == filepath.Dir(path2) {
	//				e.Op = Rename
	//			}
	//
	//			delete(removes, path1)
	//			delete(creates, path2)
	//
	//			select {
	//			case <-cancel:
	//				return
	//			case w.Event <- e:
	//			}
	//		}
	//	}
	//}
	//
	////Send all the remaining create and remove events.
	//for pth, info := range creates {
	//	select {
	//	case <-cancel:
	//		return
	//	case w.Event <- Event{Create, pth, pth, info, serviceName}:
	//	}
	//}
	//for pth, info := range removes {
	//	select {
	//	case <-cancel:
	//		return
	//	case w.Event <- Event{Remove, pth, pth, info, serviceName}:
	//	}
	//}
}
