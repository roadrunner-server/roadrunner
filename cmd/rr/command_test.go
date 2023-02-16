package main

import (
	"bytes"
	"io"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_Main(t *testing.T) {
	os.Args = []string{"rr", "--help"}
	exitFn = func(code int) { assert.Equal(t, 0, code) }

	r, w, _ := os.Pipe()
	os.Stdout = w

	main()
	_ = w.Close()
	buf := new(bytes.Buffer)

	_ = r.SetReadDeadline(time.Now().Add(time.Second))
	_, _ = io.Copy(buf, r)

	assert.Contains(t, buf.String(), "Usage:")
	assert.Contains(t, buf.String(), "Available Commands:")
	assert.Contains(t, buf.String(), "Flags:")
}

func Test_MainWithoutCommands(t *testing.T) {
	os.Args = []string{"rr"}
	exitFn = func(code int) { assert.Equal(t, 0, code) }

	r, w, _ := os.Pipe()
	os.Stdout = w

	main()
	buf := new(bytes.Buffer)
	_ = r.SetReadDeadline(time.Now().Add(time.Second))
	_, _ = io.Copy(buf, r)

	assert.Contains(t, buf.String(), "Usage:")
	assert.Contains(t, buf.String(), "Available Commands:")
	assert.Contains(t, buf.String(), "Flags:")
}

func Test_MainUnknownSubcommand(t *testing.T) {
	os.Args = []string{"", "foobar"}
	exitFn = func(code int) { assert.Equal(t, 1, code) }

	r, w, _ := os.Pipe()
	os.Stderr = w

	main()
	_ = w.Close()
	buf := new(bytes.Buffer)

	_ = r.SetReadDeadline(time.Now().Add(time.Second))
	_, _ = io.Copy(buf, r)

	assert.Contains(t, buf.String(), "unknown command")
	assert.Contains(t, buf.String(), "foobar")
}
