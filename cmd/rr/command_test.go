package main

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Main(t *testing.T) {
	os.Args = []string{"rr", "--help"}
	exitFn = func(code int) { assert.Equal(t, 0, code) }

	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	main()
	_ = w.Close()
	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, r)
	require.NoError(t, err)

	assert.Contains(t, buf.String(), "Usage:")
	assert.Contains(t, buf.String(), "Available Commands:")
	assert.Contains(t, buf.String(), "Flags:")
}

func Test_MainWithoutCommands(t *testing.T) {
	os.Args = []string{"rr"}
	exitFn = func(code int) { assert.Equal(t, 0, code) }

	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	main()
	_ = w.Close()
	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, r)
	require.NoError(t, err)

	assert.Contains(t, buf.String(), "Usage:")
	assert.Contains(t, buf.String(), "Available Commands:")
	assert.Contains(t, buf.String(), "Flags:")
}

func Test_MainUnknownSubcommand(t *testing.T) {
	os.Args = []string{"", "foobar"}
	exitFn = func(code int) { assert.Equal(t, 1, code) }

	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stderr = w

	main()
	_ = w.Close()
	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, r)
	require.NoError(t, err)

	assert.Contains(t, buf.String(), "unknown command")
	assert.Contains(t, buf.String(), "foobar")
}
