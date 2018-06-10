package static

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestConfig_Forbids(t *testing.T) {
	cfg := Config{Forbid: []string{".php"}}

	assert.True(t, cfg.Forbids("index.php"))
	assert.True(t, cfg.Forbids("index.PHP"))
	assert.True(t, cfg.Forbids("phpadmin/index.bak.php"))
	assert.False(t, cfg.Forbids("index.html"))
}

func TestConfig_Valid(t *testing.T) {
	assert.NoError(t, (&Config{Dir: "./"}).Valid())
	assert.Error(t, (&Config{Dir: "./config.go"}).Valid())
	assert.Error(t, (&Config{Dir: "./dir/"}).Valid())
}
