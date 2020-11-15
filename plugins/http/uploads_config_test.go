package http

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestFsConfig_Forbids(t *testing.T) {
	cfg := UploadsConfig{Forbid: []string{".php"}}

	assert.True(t, cfg.Forbids("index.php"))
	assert.True(t, cfg.Forbids("index.PHP"))
	assert.True(t, cfg.Forbids("phpadmin/index.bak.php"))
	assert.False(t, cfg.Forbids("index.html"))
}

func TestFsConfig_TmpFallback(t *testing.T) {
	cfg := UploadsConfig{Dir: "test"}
	assert.Equal(t, "test", cfg.TmpDir())

	cfg = UploadsConfig{Dir: ""}
	assert.Equal(t, os.TempDir(), cfg.TmpDir())
}
