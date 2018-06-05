package ext

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_JobError_Error(t *testing.T) {
	e := JobError([]byte("error"))
	assert.Equal(t, "error", e.Error())
}
