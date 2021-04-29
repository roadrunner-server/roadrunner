package handler

import "net/http"

var http2pushHeaderKey = http.CanonicalHeaderKey("http2-push")

// TrailerHeaderKey http header key
var TrailerHeaderKey = http.CanonicalHeaderKey("trailer")
