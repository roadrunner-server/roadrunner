package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSSL_Valid1(t *testing.T) {
	conf := &SSL{
		Address:  "",
		Redirect: false,
		Key:      "",
		Cert:     "",
		RootCA:   "",
		host:     "",
		Port:     0,
	}

	err := conf.Valid()
	assert.Error(t, err)
}

func TestSSL_Valid2(t *testing.T) {
	conf := &SSL{
		Address:  ":hello",
		Redirect: false,
		Key:      "",
		Cert:     "",
		RootCA:   "",
		host:     "",
		Port:     0,
	}

	err := conf.Valid()
	assert.Error(t, err)
}

func TestSSL_Valid3(t *testing.T) {
	conf := &SSL{
		Address:  ":555",
		Redirect: false,
		Key:      "",
		Cert:     "",
		RootCA:   "",
		host:     "",
		Port:     0,
	}

	err := conf.Valid()
	assert.Error(t, err)
}

func TestSSL_Valid4(t *testing.T) {
	conf := &SSL{
		Address:  ":555",
		Redirect: false,
		Key:      "../../../tests/plugins/http/fixtures/server.key",
		Cert:     "../../../tests/plugins/http/fixtures/server.crt",
		RootCA:   "",
		host:     "",
		// private
		Port: 0,
	}

	err := conf.Valid()
	assert.NoError(t, err)
}

func TestSSL_Valid5(t *testing.T) {
	conf := &SSL{
		Address:  "a:b:c",
		Redirect: false,
		Key:      "../../../tests/plugins/http/fixtures/server.key",
		Cert:     "../../../tests/plugins/http/fixtures/server.crt",
		RootCA:   "",
		host:     "",
		// private
		Port: 0,
	}

	err := conf.Valid()
	assert.Error(t, err)
}

func TestSSL_Valid6(t *testing.T) {
	conf := &SSL{
		Address:  ":",
		Redirect: false,
		Key:      "../../../tests/plugins/http/fixtures/server.key",
		Cert:     "../../../tests/plugins/http/fixtures/server.crt",
		RootCA:   "",
		host:     "",
		// private
		Port: 0,
	}

	err := conf.Valid()
	assert.Error(t, err)
}

func TestSSL_Valid7(t *testing.T) {
	conf := &SSL{
		Address:  "localhost:555:1",
		Redirect: false,
		Key:      "../../../tests/plugins/http/fixtures/server.key",
		Cert:     "../../../tests/plugins/http/fixtures/server.crt",
		RootCA:   "",
		host:     "",
		// private
		Port: 0,
	}

	err := conf.Valid()
	assert.Error(t, err)
}
