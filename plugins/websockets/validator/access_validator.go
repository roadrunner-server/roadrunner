package validator

import (
	"net/http"
	"strings"

	json "github.com/json-iterator/go"
	"github.com/spiral/errors"
	handler "github.com/spiral/roadrunner/v2/pkg/worker_handler"
	"github.com/spiral/roadrunner/v2/plugins/http/attributes"
)

type AccessValidatorFn = func(r *http.Request, channels ...string) (*AccessValidator, error)

type AccessValidator struct {
	Header http.Header `json:"headers"`
	Status int         `json:"status"`
	Body   []byte
}

func ServerAccessValidator(r *http.Request) ([]byte, error) {
	const op = errors.Op("server_access_validator")

	err := attributes.Set(r, "ws:joinServer", true)
	if err != nil {
		return nil, errors.E(op, err)
	}

	defer delete(attributes.All(r), "ws:joinServer")

	req := &handler.Request{
		RemoteAddr: handler.FetchIP(r.RemoteAddr),
		Protocol:   r.Proto,
		Method:     r.Method,
		URI:        handler.URI(r),
		Header:     r.Header,
		Cookies:    make(map[string]string),
		RawQuery:   r.URL.RawQuery,
		Attributes: attributes.All(r),
	}

	data, err := json.Marshal(req)
	if err != nil {
		return nil, errors.E(op, err)
	}

	return data, nil
}

func TopicsAccessValidator(r *http.Request, topics ...string) ([]byte, error) {
	const op = errors.Op("topic_access_validator")
	err := attributes.Set(r, "ws:joinTopics", strings.Join(topics, ","))
	if err != nil {
		return nil, errors.E(op, err)
	}

	defer delete(attributes.All(r), "ws:joinTopics")

	req := &handler.Request{
		RemoteAddr: handler.FetchIP(r.RemoteAddr),
		Protocol:   r.Proto,
		Method:     r.Method,
		URI:        handler.URI(r),
		Header:     r.Header,
		Cookies:    make(map[string]string),
		RawQuery:   r.URL.RawQuery,
		Attributes: attributes.All(r),
	}

	data, err := json.Marshal(req)
	if err != nil {
		return nil, errors.E(op, err)
	}

	return data, nil
}
