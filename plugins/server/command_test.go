package server

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServerCommandChecker(t *testing.T) {
	s := &Plugin{}
	cmd1 := "php ../../tests/client.php"
	assert.NoError(t, s.scanCommand(strings.Split(cmd1, " ")))
}
