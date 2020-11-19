package tests

import (
	"os"
	"testing"

	"github.com/spiral/roadrunner/v2/plugins/http"
	"github.com/stretchr/testify/assert"
)

func TestFsConfig_Forbids(t *testing.T) {
	cfg := http.UploadsConfig{Forbid: []string{".php"}}

	assert.True(t, cfg.Forbids("index.php"))
	assert.True(t, cfg.Forbids("index.PHP"))
	assert.True(t, cfg.Forbids("phpadmin/index.bak.php"))
	assert.False(t, cfg.Forbids("index.html"))
}

func TestFsConfig_TmpFallback(t *testing.T) {
	cfg := http.UploadsConfig{Dir: "test"}
	assert.Equal(t, "test", cfg.TmpDir())

	cfg = http.UploadsConfig{Dir: ""}
	assert.Equal(t, os.TempDir(), cfg.TmpDir())
}
