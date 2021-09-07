package grpc

import (
	"testing"

	json "github.com/json-iterator/go"
	"github.com/stretchr/testify/assert"
)

type jsonCodec struct{}

func (jsonCodec) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (jsonCodec) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

func (jsonCodec) Name() string {
	return "json"
}

func TestCodec_String(t *testing.T) {
	c := codec{jsonCodec{}}

	assert.Equal(t, "raw:json", c.String())

	r := rawMessage{}
	r.Reset()
	r.ProtoMessage()
	assert.Equal(t, "rawMessage", r.String())
}

func TestCodec_Unmarshal_ByPass(t *testing.T) {
	c := codec{jsonCodec{}}

	s := struct {
		Name string
	}{}

	assert.NoError(t, c.Unmarshal([]byte(`{"name":"name"}`), &s))
	assert.Equal(t, "name", s.Name)
}

func TestCodec_Marshal_ByPass(t *testing.T) {
	c := codec{jsonCodec{}}

	s := struct {
		Name string
	}{
		Name: "name",
	}

	d, err := c.Marshal(s)
	assert.NoError(t, err)

	assert.Equal(t, `{"Name":"name"}`, string(d))
}

func TestCodec_Unmarshal_Raw(t *testing.T) {
	c := codec{jsonCodec{}}

	s := rawMessage{}

	assert.NoError(t, c.Unmarshal([]byte(`{"name":"name"}`), &s))
	assert.Equal(t, `{"name":"name"}`, string(s))
}

func TestCodec_Marshal_Raw(t *testing.T) {
	c := codec{jsonCodec{}}

	s := rawMessage(`{"Name":"name"}`)

	d, err := c.Marshal(s)
	assert.NoError(t, err)

	assert.Equal(t, `{"Name":"name"}`, string(d))
}
