package psr7

import (
	"mime/multipart"
	"strings"
	"github.com/sirupsen/logrus"
)

type fileData map[string]interface{}

type FileUpload struct {
	Name     string `json:"name"`
	MimeType string `json:"mimetype"`
}

func (d fileData) push(k string, v []*multipart.FileHeader) {
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

func (d fileData) pushChunk(k []string, v []*multipart.FileHeader) {
	logrus.Print(v)
	if len(v) == 0 || v[0] == nil {
		return
	}

	head := k[0]
	tail := k[1:]
	if len(k) == 1 {
		d[head] = FileUpload{
			Name:     v[0].Filename,
			MimeType: v[0].Header.Get("Content-Type"),
		}
		return
	}

	// unnamed array
	if len(tail) == 1 && tail[0] == "" {
		d[head] = v
		return
	}

	if p, ok := d[head]; !ok {
		d[head] = make(fileData)
		d[head].(fileData).pushChunk(tail, v)
	} else {
		p.(fileData).pushChunk(tail, v)
	}
}
