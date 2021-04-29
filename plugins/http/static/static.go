package static

import (
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"

	httpConfig "github.com/spiral/roadrunner/v2/plugins/http/config"
)

type ExtensionFilter struct {
	allowed   map[string]struct{}
	forbidden map[string]struct{}
}

func NewExtensionFilter(allow, forbid []string) *ExtensionFilter {
	ef := &ExtensionFilter{
		allowed:   make(map[string]struct{}, len(allow)),
		forbidden: make(map[string]struct{}, len(forbid)),
	}

	for i := 0; i < len(forbid); i++ {
		// skip empty lines
		if forbid[i] == "" {
			continue
		}
		ef.forbidden[forbid[i]] = struct{}{}
	}

	for i := 0; i < len(allow); i++ {
		// skip empty lines
		if allow[i] == "" {
			continue
		}
		ef.allowed[allow[i]] = struct{}{}
	}

	// check if any forbidden items presented in the allowed
	// if presented, delete such items from allowed
	for k := range ef.allowed {
		if _, ok := ef.forbidden[k]; ok {
			delete(ef.allowed, k)
		}
	}

	return ef
}

type FileSystem struct {
	ef *ExtensionFilter
	// embedded
	http.FileSystem
}

// Open wrapper around http.FileSystem Open method, name here is the name of the
func (f FileSystem) Open(name string) (http.File, error) {
	file, err := f.FileSystem.Open(name)
	if err != nil {
		return nil, err
	}

	fstat, err := file.Stat()
	if err != nil {
		return nil, fs.ErrNotExist
	}

	if fstat.IsDir() {
		return nil, fs.ErrPermission
	}

	ext := strings.ToLower(filepath.Ext(fstat.Name()))
	if _, ok := f.ef.forbidden[ext]; ok {
		return nil, fs.ErrPermission
	}

	// if file extension is allowed, append it to the FileInfo slice
	if _, ok := f.ef.allowed[ext]; ok {
		return file, nil
	}

	return nil, fs.ErrNotExist
}

// FS is a constructor for the http.FileSystem
func FS(config *httpConfig.Static) http.FileSystem {
	return FileSystem{NewExtensionFilter(config.Allow, config.Forbid), http.Dir(config.Dir)}
}
