package http

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"sync"
)

const (
	// There is no error, the file uploaded with success.
	UploadErrorOK = 0

	// No file was uploaded.
	UploadErrorNoFile = 4

	// Missing a temporary folder.
	UploadErrorNoTmpDir = 5

	// Failed to write file to disk.
	UploadErrorCantWrite = 6

	// ForbidUploads file extension.
	UploadErrorExtension = 7
)

// FileUpload represents singular file wrapUpload.
type FileUpload struct {
	// Name contains filename specified by the client.
	Name string `json:"name"`

	// MimeType contains mime-type provided by the client.
	MimeType string `json:"type"`

	// Size of the uploaded file.
	Size int64 `json:"size"`

	// Error indicates file upload error (if any). See http://php.net/manual/en/features.file-upload.errors.php
	Error int `json:"error"`

	// TempFilename points to temporary file location.
	TempFilename string `json:"tmpName"`

	// associated file header
	header *multipart.FileHeader
}

func (f *FileUpload) Open(cfg *Config) error {
	if cfg.Forbidden(f.Name) {
		f.Error = UploadErrorExtension
		return nil
	}

	file, err := f.header.Open()
	if err != nil {
		f.Error = UploadErrorNoFile
		return err
	}
	defer file.Close()

	tmp, err := ioutil.TempFile(cfg.TmpDir, "upload")
	if err != nil {
		// most likely cause of this issue is missing tmp dir
		f.Error = UploadErrorNoTmpDir
		return err
	}

	f.TempFilename = tmp.Name()
	defer tmp.Close()

	if f.Size, err = io.Copy(tmp, file); err != nil {
		f.Error = UploadErrorCantWrite
	}

	return err
}

func wrapUpload(f *multipart.FileHeader) *FileUpload {
	return &FileUpload{
		Name:     f.Filename,
		MimeType: f.Header.Get("Content-Type"),
		Error:    UploadErrorOK,
		header:   f,
	}
}

type fileTree map[string]interface{}

func (d fileTree) push(k string, v []*FileUpload) {
	if len(v) == 0 {
		// skip empty values
		return
	}

	indexes := make([]string, 0)
	for _, index := range strings.Split(k, "[") {
		indexes = append(indexes, strings.Trim(index, "]"))
	}

	if len(indexes) <= maxLevel {
		d.mount(indexes, v)
	}
}

// mount mounts data tree recursively.
func (d fileTree) mount(i []string, v []*FileUpload) {
	if len(v) == 0 {
		return
	}

	if len(i) == 1 {
		// single value context
		d[i[0]] = v[0]
		return
	}

	if len(i) == 2 && i[1] == "" {
		// non associated array of elements
		d[i[0]] = v
		return
	}

	if p, ok := d[i[0]]; ok {
		p.(fileTree).mount(i[1:], v)
	}

	d[i[0]] = make(fileTree)
	d[i[0]].(fileTree).mount(i[1:], v)
}

// tree manages uploaded files tree and temporary files.
type Uploads struct {
	// pre processed data tree for Uploads.
	tree fileTree

	// flat list of all file Uploads.
	list []*FileUpload
}

// MarshalJSON marshal tree tree into JSON.
func (u *Uploads) MarshalJSON() ([]byte, error) {
	return json.Marshal(u.tree)
}

// Open moves all uploaded files to temp directory, return error in case of issue with temp directory. File errors
// will be handled individually. @todo: do we need it?
func (u *Uploads) Open(cfg *Config) error {
	var wg sync.WaitGroup
	for _, f := range u.list {
		wg.Add(1)
		go func(f *FileUpload) {
			defer wg.Done()
			f.Open(cfg)
		}(f)
	}

	wg.Wait()
	return nil
}

// Clear deletes all temporary files.
func (u *Uploads) Clear() {
	for _, f := range u.list {
		if f.TempFilename != "" && exists(f.TempFilename) {
			os.Remove(f.TempFilename)
		}
	}
}

// parse incoming dataTree request into JSON (including multipart form dataTree)
func parseUploads(r *http.Request) (*Uploads, error) {
	u := &Uploads{
		tree: make(fileTree),
		list: make([]*FileUpload, 0),
	}

	for k, v := range r.MultipartForm.File {
		files := make([]*FileUpload, 0, len(v))
		for _, f := range v {
			files = append(files, wrapUpload(f))
		}

		u.list = append(u.list, files...)
		u.tree.push(k, files)
	}

	return u, nil
}

// exists if file exists. by osutils; todo: better?
func exists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}

	if os.IsNotExist(err) {
		return false
	}

	panic(fmt.Errorf("unable to stat path %q; %v", path, err))
}
