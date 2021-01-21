package http

import (
	"os"
	"testing"

	"github.com/spiral/roadrunner/v2/plugins/http/config"
	"github.com/stretchr/testify/assert"
)

func TestFsConfig_Forbids(t *testing.T) {
	cfg := config.Uploads{Forbid: []string{".php"}}

	assert.True(t, cfg.Forbids("index.php"))
	assert.True(t, cfg.Forbids("index.PHP"))
	assert.True(t, cfg.Forbids("phpadmin/index.bak.php"))
	assert.False(t, cfg.Forbids("index.html"))
}

func TestFsConfig_TmpFallback(t *testing.T) {
	cfg := config.Uploads{Dir: "test"}
	assert.Equal(t, "test", cfg.TmpDir())

	cfg = config.Uploads{Dir: ""}
	assert.Equal(t, os.TempDir(), cfg.TmpDir())
}
