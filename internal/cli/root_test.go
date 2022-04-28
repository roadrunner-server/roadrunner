package cli_test

import (
	"os"
	"path"
	"testing"

	"github.com/roadrunner-server/roadrunner/v2/internal/cli"
	"github.com/stretchr/testify/require"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestCommandSubcommands(t *testing.T) {
	cmd := cli.NewCommand("unit test")

	cases := []struct {
		giveName string
	}{
		{giveName: "workers"},
		{giveName: "reset"},
		{giveName: "serve"},
	}

	// get all existing subcommands and put into the map
	subcommands := make(map[string]*cobra.Command)
	for _, sub := range cmd.Commands() {
		subcommands[sub.Name()] = sub
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.giveName, func(t *testing.T) {
			if _, exists := subcommands[tt.giveName]; !exists {
				assert.Failf(t, "command not found", "command [%s] was not found", tt.giveName)
			}
		})
	}
}

func TestCommandFlags(t *testing.T) {
	cmd := cli.NewCommand("unit test")

	cases := []struct {
		giveName      string
		wantShorthand string
		wantDefault   string
	}{
		{giveName: "config", wantShorthand: "c", wantDefault: ".rr.yaml"},
		{giveName: "WorkDir", wantShorthand: "w", wantDefault: ""},
		{giveName: "dotenv", wantShorthand: "", wantDefault: ""},
		{giveName: "debug", wantShorthand: "d", wantDefault: "false"},
		{giveName: "override", wantShorthand: "o", wantDefault: "[]"},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.giveName, func(t *testing.T) {
			flag := cmd.Flag(tt.giveName)

			if flag == nil {
				assert.Failf(t, "flag not found", "flag [%s] was not found", tt.giveName)

				return
			}

			assert.Equal(t, tt.wantShorthand, flag.Shorthand)
			assert.Equal(t, tt.wantDefault, flag.DefValue)
		})
	}
}

func TestCommandSimpleExecuting(t *testing.T) {
	cmd := cli.NewCommand("unit test")
	cmd.SetArgs([]string{"-c", "./../../.rr.yaml"})

	var executed bool

	if cmd.Run == nil { // override "Run" property for test (if it was not set)
		cmd.Run = func(cmd *cobra.Command, args []string) {
			executed = true
		}
	}

	assert.NoError(t, cmd.Execute())
	assert.True(t, executed)
}

func TestCommandNoEnvFileError(t *testing.T) {
	cmd := cli.NewCommand("unit test")
	cmd.SetArgs([]string{"-c", "./../../.rr.yaml", "--dotenv", "foo/bar"})

	var executed bool

	if cmd.Run == nil { // override "Run" property for test (if it was not set)
		cmd.Run = func(cmd *cobra.Command, args []string) {
			executed = true
		}
	}

	assert.Error(t, cmd.Execute())
	assert.False(t, executed)
}

func TestCommandNoEnvFileNoError(t *testing.T) {
	tmp := os.TempDir()

	cmd := cli.NewCommand("unit test")
	cmd.SetArgs([]string{"-c", path.Join(tmp, ".rr.yaml"), "--dotenv", path.Join(tmp, ".env")})

	var executed bool

	f, err := os.Create(path.Join(tmp, ".env"))
	require.NoError(t, err)
	f2, err := os.Create(path.Join(tmp, ".rr.yaml"))
	require.NoError(t, err)

	defer func() {
		_ = f.Close()
		_ = f2.Close()
	}()

	if cmd.Run == nil { // override "Run" property for test (if it was not set)
		cmd.Run = func(cmd *cobra.Command, args []string) {
			executed = true
		}
	}

	assert.NoError(t, cmd.Execute())
	assert.True(t, executed)

	t.Cleanup(func() {
		_ = os.RemoveAll(path.Join(tmp, ".env"))
		_ = os.RemoveAll(path.Join(tmp, ".rr.yaml"))
	})
}

func TestCommandWorkingDir(t *testing.T) {
	tmp := os.TempDir()

	cmd := cli.NewCommand("serve")
	cmd.SetArgs([]string{"-w", tmp})

	var executed bool

	var wd string

	f2, err := os.Create(path.Join(tmp, ".rr.yaml"))
	require.NoError(t, err)

	if cmd.Run == nil { // override "Run" property for test (if it was not set)
		cmd.Run = func(cmd *cobra.Command, args []string) {
			executed = true
			wd, _ = os.Getwd()
		}
	}

	assert.NoError(t, cmd.Execute())
	assert.True(t, executed)
	assert.Equal(t, "/tmp", wd)

	t.Cleanup(func() {
		_ = f2.Close()
		_ = os.RemoveAll(path.Join(tmp, ".rr.yaml"))
	})
}
