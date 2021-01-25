package rpc

import (
	"runtime"
	"testing"

	"github.com/spiral/roadrunner/v2/plugins/rpc"
	"github.com/stretchr/testify/assert"
)

func TestConfig_Listener(t *testing.T) {
	cfg := &rpc.Config{Listen: "tcp://:18001"}

	ln, err := cfg.Listener()
	assert.NoError(t, err)
	assert.NotNil(t, ln)
	defer func() {
		err := ln.Close()
		if err != nil {
			t.Errorf("error closing the listener: error %v", err)
		}
	}()

	assert.Equal(t, "tcp", ln.Addr().Network())
	if runtime.GOOS == "windows" {
		assert.Equal(t, "[::]:18001", ln.Addr().String())
	} else {
		assert.Equal(t, "0.0.0.0:18001", ln.Addr().String())
	}
}

func TestConfig_ListenerUnix(t *testing.T) {
	cfg := &rpc.Config{Listen: "unix://file.sock"}

	ln, err := cfg.Listener()
	assert.NoError(t, err)
	assert.NotNil(t, ln)
	defer func() {
		err := ln.Close()
		if err != nil {
			t.Errorf("error closing the listener: error %v", err)
		}
	}()

	assert.Equal(t, "unix", ln.Addr().Network())
	assert.Equal(t, "file.sock", ln.Addr().String())
}

func Test_Config_Error(t *testing.T) {
	cfg := &rpc.Config{Listen: "uni:unix.sock"}
	ln, err := cfg.Listener()
	assert.Nil(t, ln)
	assert.Error(t, err)
}

func Test_Config_ErrorMethod(t *testing.T) {
	cfg := &rpc.Config{Listen: "xinu://unix.sock"}

	ln, err := cfg.Listener()
	assert.Nil(t, ln)
	assert.Error(t, err)
}

func TestConfig_Dialer(t *testing.T) {
	cfg := &rpc.Config{Listen: "tcp://:18001"}

	ln, _ := cfg.Listener()
	defer func() {
		err := ln.Close()
		if err != nil {
			t.Errorf("error closing the listener: error %v", err)
		}
	}()

	conn, err := cfg.Dialer()
	assert.NoError(t, err)
	assert.NotNil(t, conn)
	defer func() {
		err := conn.Close()
		if err != nil {
			t.Errorf("error closing the connection: error %v", err)
		}
	}()

	assert.Equal(t, "tcp", conn.RemoteAddr().Network())
	assert.Equal(t, "127.0.0.1:18001", conn.RemoteAddr().String())
}

func TestConfig_DialerUnix(t *testing.T) {
	cfg := &rpc.Config{Listen: "unix://file.sock"}

	ln, _ := cfg.Listener()
	defer func() {
		err := ln.Close()
		if err != nil {
			t.Errorf("error closing the listener: error %v", err)
		}
	}()

	conn, err := cfg.Dialer()
	assert.NoError(t, err)
	assert.NotNil(t, conn)
	defer func() {
		err := conn.Close()
		if err != nil {
			t.Errorf("error closing the connection: error %v", err)
		}
	}()

	assert.Equal(t, "unix", conn.RemoteAddr().Network())
	assert.Equal(t, "file.sock", conn.RemoteAddr().String())
}

func Test_Config_DialerError(t *testing.T) {
	cfg := &rpc.Config{Listen: "uni:unix.sock"}
	ln, err := cfg.Dialer()
	assert.Nil(t, ln)
	assert.Error(t, err)
	assert.Equal(t, "invalid socket DSN (tcp://:6001, unix://file.sock)", err.Error())
}

func Test_Config_DialerErrorMethod(t *testing.T) {
	cfg := &rpc.Config{Listen: "xinu://unix.sock"}

	ln, err := cfg.Dialer()
	assert.Nil(t, ln)
	assert.Error(t, err)
}

func Test_Config_Defaults(t *testing.T) {
	c := &rpc.Config{}
	c.InitDefaults()
	assert.Equal(t, "tcp://127.0.0.1:6001", c.Listen)
}
