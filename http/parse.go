package http

import (
	"strings"
	"net/http"
	"os"
)

// MaxLevel defines maximum tree depth for incoming request data and files.
const MaxLevel = 127

type dataTree map[string]interface{}
type fileTree map[string]interface{}

// parseData parses incoming request body into data tree.
func parseData(r *http.Request) (dataTree, error) {
	data := make(dataTree)
	for k, v := range r.PostForm {
		data.push(k, v)
	}

	for k, v := range r.MultipartForm.Value {
		data.push(k, v)
	}

	return data, nil
}

// pushes value into data tree.
func (d dataTree) push(k string, v []string) {
	if len(v) == 0 {
		// skip empty values
		return
	}

	indexes := make([]string, 0)
	for _, index := range strings.Split(k, "[") {
		indexes = append(indexes, strings.Trim(index, "]"))
	}

	if len(indexes) <= MaxLevel {
		d.mount(indexes, v)
	}
}

// mount mounts data tree recursively.
func (d dataTree) mount(i []string, v []string) {
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
		p.(dataTree).mount(i[1:], v)
	}

	d[i[0]] = make(dataTree)
	d[i[0]].(dataTree).mount(i[1:], v)
}

// parse incoming dataTree request into JSON (including multipart form dataTree)
func parseUploads(r *http.Request, cfg *UploadsConfig) (*Uploads, error) {
	u := &Uploads{
		cfg:  cfg,
		tree: make(fileTree),
		list: make([]*FileUpload, 0),
	}

	for k, v := range r.MultipartForm.File {
		files := make([]*FileUpload, 0, len(v))
		for _, f := range v {
			files = append(files, NewUpload(f))
		}

		u.list = append(u.list, files...)
		u.tree.push(k, files)
	}

	return u, nil
}

// exists if file exists.
func exists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}

	if os.IsNotExist(err) {
		return false
	}

	return false
}

// pushes new file upload into it's proper place.
func (d fileTree) push(k string, v []*FileUpload) {
	if len(v) == 0 {
		// skip empty values
		return
	}

	indexes := make([]string, 0)
	for _, index := range strings.Split(k, "[") {
		indexes = append(indexes, strings.Trim(index, "]"))
	}

	if len(indexes) <= MaxLevel {
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
