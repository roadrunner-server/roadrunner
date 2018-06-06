package roadrunner

import (
	"testing"
	"runtime"
	"github.com/stretchr/testify/assert"
	"os/user"
)

func Test_ResolveUser_Error(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("not supported on " + runtime.GOOS)
	}

	u, err := resolveUser("-1")
	assert.Nil(t, u)
	assert.Error(t, err)

	u, err = resolveUser("random-user")
	assert.Nil(t, u)
	assert.Error(t, err)
}

func Test_ResolveUser(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("not supported on " + runtime.GOOS)
	}

	current, err := user.Current()
	assert.NotNil(t, current)
	assert.NoError(t, err)

	u, err := resolveUser(current.Uid)
	assert.NoError(t, err)
	assert.NotNil(t, u)
	assert.Equal(t, current.Uid, u.Uid)

	u, err = resolveUser(current.Username)
	assert.NoError(t, err)
	assert.NotNil(t, u)
	assert.Equal(t, current.Uid, u.Uid)
}

func Test_ResolveGroup_Error(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("not supported on " + runtime.GOOS)
	}

	g, err := resolveGroup("-1")
	assert.Nil(t, g)
	assert.Error(t, err)

	g, err = resolveGroup("random-group")
	assert.Nil(t, g)
	assert.Error(t, err)
}

func Test_ResolveGroup(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("not supported on " + runtime.GOOS)
	}

	current, err := user.Current()
	assert.NotNil(t, current)
	assert.NoError(t, err)

	g, err := resolveGroup(current.Gid)
	assert.NoError(t, err)
	assert.NotNil(t, g)
	assert.Equal(t, current.Gid, g.Gid)

	g2, err := resolveGroup(g.Name)
	assert.NoError(t, err)
	assert.NotNil(t, g2)
	assert.Equal(t, g2.Gid, g.Gid)
}
