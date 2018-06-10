package http

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestFsConfig_Forbids(t *testing.T) {
	cfg := FsConfig{Forbid: []string{".php"}}

	assert.True(t, cfg.Forbids("index.php"))
	assert.True(t, cfg.Forbids("index.PHP"))
	assert.True(t, cfg.Forbids("phpadmin/index.bak.php"))
	assert.False(t, cfg.Forbids("index.html"))
}
