package roadrunner

import (
	"github.com/stretchr/testify/assert"
	"runtime"
	"testing"
)

func Test_ServerConfig_PipeFactory(t *testing.T) {
	cfg := &ServerConfig{Relay: "pipes"}
	f, err := cfg.makeFactory()

	assert.NoError(t, err)
	assert.IsType(t, &PipeFactory{}, f)

	cfg = &ServerConfig{Relay: "pipe"}
	f, err = cfg.makeFactory()
	assert.NoError(t, err)
	assert.NotNil(t, f)
	defer f.Close()

	assert.NoError(t, err)
	assert.IsType(t, &PipeFactory{}, f)
}

func Test_ServerConfig_SocketFactory(t *testing.T) {
	cfg := &ServerConfig{Relay: "tcp://:9111"}
	f, err := cfg.makeFactory()
	assert.NoError(t, err)
	assert.NotNil(t, f)
	defer f.Close()

	assert.NoError(t, err)
	assert.IsType(t, &SocketFactory{}, f)
	assert.Equal(t, "tcp", f.(*SocketFactory).ls.Addr().Network())
	assert.Equal(t, "[::]:9111", f.(*SocketFactory).ls.Addr().String())

	cfg = &ServerConfig{Relay: "tcp://localhost:9112"}
	f, err = cfg.makeFactory()
	assert.NoError(t, err)
	assert.NotNil(t, f)
	defer f.Close()

	assert.NoError(t, err)
	assert.IsType(t, &SocketFactory{}, f)
	assert.Equal(t, "tcp", f.(*SocketFactory).ls.Addr().Network())
	assert.Equal(t, "127.0.0.1:9112", f.(*SocketFactory).ls.Addr().String())
}

func Test_ServerConfig_UnixSocketFactory(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("not supported on " + runtime.GOOS)
	}

	cfg := &ServerConfig{Relay: "unix://unix.sock"}
	f, err := cfg.makeFactory()
	defer f.Close()

	assert.NoError(t, err)
	assert.IsType(t, &SocketFactory{}, f)
	assert.Equal(t, "unix", f.(*SocketFactory).ls.Addr().Network())
	assert.Equal(t, "unix.sock", f.(*SocketFactory).ls.Addr().String())
}

func Test_ServerConfig_ErrorFactory(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("not supported on " + runtime.GOOS)
	}

	cfg := &ServerConfig{Relay: "uni:unix.sock"}
	f, err := cfg.makeFactory()
	assert.Nil(t, f)
	assert.Error(t, err)
	assert.Equal(t, "invalid relay DSN (pipes, tcp://:6001, unix://rr.sock)", err.Error())
}

func Test_ServerConfig_ErrorMethod(t *testing.T) {
	cfg := &ServerConfig{Relay: "xinu://unix.sock"}

	f, err := cfg.makeFactory()
	assert.Nil(t, f)
	assert.Error(t, err)
}
