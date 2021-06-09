package server

import (
	"strings"
	"testing"

	"github.com/spiral/errors"
	"github.com/stretchr/testify/assert"
)

func TestServerCommandChecker(t *testing.T) {
	s := &Plugin{}
	cmd1 := "php ../../tests/client.php"
	assert.NoError(t, s.scanCommand(strings.Split(cmd1, " ")))

	cmd2 := "C:/../../abcdef/client.php"
	assert.Error(t, s.scanCommand(strings.Split(cmd2, " ")))

	cmd3 := "sh ./script.sh"
	err := s.scanCommand(strings.Split(cmd3, " "))
	assert.Error(t, err)
	if !errors.Is(errors.FileNotFound, err) {
		t.Fatal("should be of filenotfound type")
	}

	cmd4 := "php ../../tests/client.php --option1 --option2"
	err = s.scanCommand(strings.Split(cmd4, " "))
	assert.NoError(t, err)

	cmd5 := "php ../../tests/cluent.php --option1 --option2"
	err = s.scanCommand(strings.Split(cmd5, " "))
	assert.Error(t, err)
	if !errors.Is(errors.FileNotFound, err) {
		t.Fatal("should be of filenotfound type")
	}

	cmd6 := "php 0/../../tests/cluent.php --option1 --option2"
	err = s.scanCommand(strings.Split(cmd6, " "))
	assert.Error(t, err)
	if errors.Is(errors.FileNotFound, err) {
		t.Fatal("should be of filenotfound type")
	}
}
