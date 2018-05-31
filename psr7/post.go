package psr7

import "strings"

type postData map[string]interface{}

func (d postData) push(k string, v []string) {
	if len(v) == 0 {
		// doing nothing
		return
	}

	chunks := make([]string, 0)
	for _, chunk := range strings.Split(k, "[") {
		chunks = append(chunks, strings.Trim(chunk, "]"))
	}

	d.pushChunk(chunks, v)
}

func (d postData) pushChunk(k []string, v []string) {
	if len(v) == 0 || v[0] == "" {
		return
	}

	head := k[0]
	tail := k[1:]
	if len(k) == 1 {
		d[head] = v[0]
		return
	}

	// unnamed array
	if len(tail) == 1 && tail[0] == "" {
		d[head] = v
		return
	}

	if p, ok := d[head]; !ok {
		d[head] = make(postData)
		d[head].(postData).pushChunk(tail, v)
	} else {
		p.(postData).pushChunk(tail, v)
	}
}
