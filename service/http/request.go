package http

import (
	"encoding/json"
	"fmt"
	"github.com/spiral/roadrunner"
	"github.com/spiral/roadrunner/service/http/attributes"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
)

const (
	defaultMaxMemory = 32 << 20 // 32 MB
	contentNone      = iota + 900
	contentStream
	contentMultipart
	contentFormData
)

// Request maps net/http requests to PSR7 compatible structure and managed state of temporary uploaded files.
type Request struct {
	// RemoteAddr contains ip address of client, make sure to check X-Real-Ip and X-Forwarded-For for real client address.
	RemoteAddr string `json:"remoteAddr"`

	// Protocol includes HTTP protocol version.
	Protocol string `json:"protocol"`

	// Method contains name of HTTP method used for the request.
	Method string `json:"method"`

	// URI contains full request URI with scheme and query.
	URI string `json:"uri"`

	// Header contains list of request headers.
	Header http.Header `json:"headers"`

	// Cookies contains list of request cookies.
	Cookies map[string]string `json:"cookies"`

	// RawQuery contains non parsed query string (to be parsed on php end).
	RawQuery string `json:"rawQuery"`

	// Parsed indicates that request body has been parsed on RR end.
	Parsed bool `json:"parsed"`

	// Uploads contains list of uploaded files, their names, sized and associations with temporary files.
	Uploads *Uploads `json:"uploads"`

	// Attributes can be set by chained mdwr to safely pass value from Golang to PHP. See: GetAttribute, SetAttribute functions.
	Attributes map[string]interface{} `json:"attributes"`

	// request body can be parsedData or []byte
	body interface{}
}

func fetchIP(pair string) string {
	if !strings.ContainsRune(pair, ':') {
		return pair
	}

	addr, _, _ := net.SplitHostPort(pair)
	return addr
}

// NewRequest creates new PSR7 compatible request using net/http request.
func NewRequest(r *http.Request, cfg *UploadsConfig) (req *Request, err error) {
	req = &Request{
		RemoteAddr: fetchIP(r.RemoteAddr),
		Protocol:   r.Proto,
		Method:     r.Method,
		URI:        uri(r),
		Header:     r.Header,
		Cookies:    make(map[string]string),
		RawQuery:   r.URL.RawQuery,
		Attributes: attributes.All(r),
	}

	for _, c := range r.Cookies() {
		if v, err := url.QueryUnescape(c.Value); err == nil {
			req.Cookies[c.Name] = v
		}
	}

	switch req.contentType() {
	case contentNone:
		return req, nil

	case contentStream:
		req.body, err = ioutil.ReadAll(r.Body)
		return req, err

	case contentMultipart:
		if err = r.ParseMultipartForm(defaultMaxMemory); err != nil {
			return nil, err
		}

		req.Uploads = parseUploads(r, cfg)
		fallthrough
	case contentFormData:
		if err = r.ParseForm(); err != nil {
			return nil, err
		}

		req.body = parseData(r)
	}

	req.Parsed = true
	return req, nil
}

// Open moves all uploaded files to temporary directory so it can be given to php later.
func (r *Request) Open() {
	if r.Uploads == nil {
		return
	}

	r.Uploads.Open()
}

// Close clears all temp file uploads
func (r *Request) Close() {
	if r.Uploads == nil {
		return
	}

	r.Uploads.Clear()
}

// Payload request marshaled RoadRunner payload based on PSR7 data. values encode method is JSON. Make sure to open
// files prior to calling this method.
func (r *Request) Payload() (p *roadrunner.Payload, err error) {
	p = &roadrunner.Payload{}

	if p.Context, err = json.Marshal(r); err != nil {
		return nil, err
	}

	if r.Parsed {
		if p.Body, err = json.Marshal(r.body); err != nil {
			return nil, err
		}
	} else if r.body != nil {
		p.Body = r.body.([]byte)
	}

	return p, nil
}

// contentType returns the payload content type.
func (r *Request) contentType() int {
	if r.Method == "GET" || r.Method == "HEAD" || r.Method == "OPTIONS" {
		return contentNone
	}

	ct := r.Header.Get("content-type")
	if strings.Contains(ct, "application/x-www-form-urlencoded") {
		return contentFormData
	}

	if strings.Contains(ct, "multipart/form-data") {
		return contentMultipart
	}

	return contentStream
}

// uri fetches full uri from request in a form of string (including https scheme if TLS connection is enabled).
func uri(r *http.Request) string {
	if r.TLS != nil {
		return fmt.Sprintf("https://%s%s", r.Host, r.URL.String())
	}

	return fmt.Sprintf("http://%s%s", r.Host, r.URL.String())
}
