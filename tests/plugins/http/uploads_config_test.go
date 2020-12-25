package http

import (
	"os"
	"testing"

	httpPlugin "github.com/spiral/roadrunner/v2/pkg/plugins/http"
	"github.com/stretchr/testify/assert"
)

func TestFsConfig_Forbids(t *testing.T) {
	cfg := httpPlugin.UploadsConfig{Forbid: []string{".php"}}

	assert.True(t, cfg.Forbids("index.php"))
	assert.True(t, cfg.Forbids("index.PHP"))
	assert.True(t, cfg.Forbids("phpadmin/index.bak.php"))
	assert.False(t, cfg.Forbids("index.html"))
}

func TestFsConfig_TmpFallback(t *testing.T) {
	cfg := httpPlugin.UploadsConfig{Dir: "test"}
	assert.Equal(t, "test", cfg.TmpDir())

	cfg = httpPlugin.UploadsConfig{Dir: ""}
	assert.Equal(t, os.TempDir(), cfg.TmpDir())
}
