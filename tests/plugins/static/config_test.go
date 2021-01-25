package static

import (
	"testing"

	"github.com/spiral/roadrunner/v2/plugins/static"
	"github.com/stretchr/testify/assert"
)

func TestConfig_Forbids(t *testing.T) {
	cfg := static.Config{Static: &struct {
		Dir      string
		Forbid   []string
		Always   []string
		Request  map[string]string
		Response map[string]string
	}{Dir: "", Forbid: []string{".php"}, Always: nil, Request: nil, Response: nil}}

	assert.True(t, cfg.AlwaysForbid("index.php"))
	assert.True(t, cfg.AlwaysForbid("index.PHP"))
	assert.True(t, cfg.AlwaysForbid("phpadmin/index.bak.php"))
	assert.False(t, cfg.AlwaysForbid("index.html"))
}

func TestConfig_Valid(t *testing.T) {
	assert.NoError(t, (&static.Config{Static: &struct {
		Dir      string
		Forbid   []string
		Always   []string
		Request  map[string]string
		Response map[string]string
	}{Dir: "./"}}).Valid())

	assert.Error(t, (&static.Config{Static: &struct {
		Dir      string
		Forbid   []string
		Always   []string
		Request  map[string]string
		Response map[string]string
	}{Dir: "./http.go"}}).Valid())

	assert.Error(t, (&static.Config{Static: &struct {
		Dir      string
		Forbid   []string
		Always   []string
		Request  map[string]string
		Response map[string]string
	}{Dir: "./dir/"}}).Valid())
}
