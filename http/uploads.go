package http

import (
	"mime/multipart"
	"encoding/json"
	"log"
	"strings"
	"net/http"
	"io/ioutil"
	"io"
	"sync"
)

// FileUpload represents singular file wrapUpload.
type FileUpload struct {
	// Name contains filename specified by the client.
	Name string `json:"name"`

	// MimeType contains mime-type provided by the client.
	MimeType string `json:"mimetype"`

	// Size of the uploaded file.
	Size int64 `json:"size"`

	// Error indicates file upload error (if any). See http://php.net/manual/en/features.file-upload.errors.php
	Error int

	// TempFilename points to temporary file location.
	TempFilename string `json:"tempFilename"`

	// associated file header
	header *multipart.FileHeader
}

func (f *FileUpload) Open(tmpDir string) error {
	file, err := f.header.Open()
	if err != nil {
		return err
	}

	defer file.Close()

	tmp, err := ioutil.TempFile(tmpDir, "upload")
	if err != nil {
		return err
	}

	f.TempFilename = tmp.Name()
	defer tmp.Close()

	f.Size, err = io.Copy(tmp, file)
	return err
}

func wrapUpload(f *multipart.FileHeader) *FileUpload {
	log.Print(f.Header)
	return &FileUpload{
		Name:     f.Filename,
		MimeType: f.Header.Get("Content-Type"),
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

// OpenUploads moves all uploaded files to temp directory, return error in case of issue with temp directory. File errors
// will be handled individually. @todo: do we need it?
func (u *Uploads) OpenUploads(tmpDir string) error {
	var wg sync.WaitGroup
	for _, f := range u.list {
		wg.Add(1)
		go func(f *FileUpload) {
			defer wg.Done()
			f.Open(tmpDir)
		}(f)
	}

	wg.Wait()
	log.Print(u.list)
	return nil
}

// Clear deletes all temporary files.
func (u *Uploads) Clear() {

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
