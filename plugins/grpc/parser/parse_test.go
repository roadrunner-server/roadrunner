package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseFile(t *testing.T) {
	services, err := File("test.proto", "")
	assert.NoError(t, err)
	assert.Len(t, services, 2)

	assert.Equal(t, "app.namespace", services[0].Package)
}

func TestParseFileWithImportsNestedFolder(t *testing.T) {
	services, err := File("./test_nested/test_import.proto", "./test_nested")
	assert.NoError(t, err)
	assert.Len(t, services, 2)

	assert.Equal(t, "app.namespace", services[0].Package)
}

func TestParseFileWithImports(t *testing.T) {
	services, err := File("test_import.proto", ".")
	assert.NoError(t, err)
	assert.Len(t, services, 2)

	assert.Equal(t, "app.namespace", services[0].Package)
}

func TestParseNotFound(t *testing.T) {
	_, err := File("test2.proto", "")
	assert.Error(t, err)
}

func TestParseBytes(t *testing.T) {
	services, err := Bytes([]byte{})
	assert.NoError(t, err)
	assert.Len(t, services, 0)
}

func TestParseString(t *testing.T) {
	services, err := Bytes([]byte(`
syntax = "proto3";
package app.namespace;

// Ping Service.
service PingService {
   // Ping Method.
   rpc Ping (Message) returns (Message) {
   }
}

// Pong service.
service PongService {
   rpc Pong (stream Message) returns (stream Message) {
   }
}

message Message {
   string msg = 1;
   int64 value = 2;
}
`))
	assert.NoError(t, err)
	assert.Len(t, services, 2)

	assert.Equal(t, "app.namespace", services[0].Package)
}
