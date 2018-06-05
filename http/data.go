package http

import (
	"net/http"
	"strings"
)

const maxLevel = 127

type dataTree map[string]interface{}

// parsePost parses incoming request body into data tree.
func parsePost(r *http.Request) (dataTree, error) {
	data := make(dataTree)

	for k, v := range r.PostForm {
		data.push(k, v)
	}

	for k, v := range r.MultipartForm.Value {
		data.push(k, v)
	}

	return data, nil
}

func (d dataTree) push(k string, v []string) {
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
