package http

import "net/http"

var http2pushHeaderKey = http.CanonicalHeaderKey("http2-push")
var trailerHeaderKey = http.CanonicalHeaderKey("trailer")
