package worker

import (
	"os/exec"
	"testing"

	"github.com/spiral/roadrunner/v2/internal"
	"github.com/stretchr/testify/assert"
)

func Test_NotStarted_String(t *testing.T) {
	cmd := exec.Command("php", "tests/client.php", "echo", "pipes")

	w, _ := InitBaseWorker(cmd)
	assert.Contains(t, w.String(), "php tests/client.php echo pipes")
	assert.Contains(t, w.String(), "inactive")
	assert.Contains(t, w.String(), "numExecs: 0")
}

func Test_NotStarted_Exec(t *testing.T) {
	cmd := exec.Command("php", "tests/client.php", "echo", "pipes")

	w, _ := InitBaseWorker(cmd)

	syncWorker, err := From(w)
	if err != nil {
		t.Fatal(err)
	}

	res, err := syncWorker.Exec(internal.Payload{Body: []byte("hello")})

	assert.Error(t, err)
	assert.Nil(t, res.Body)
	assert.Nil(t, res.Context)

	assert.Contains(t, err.Error(), "Process is not ready (inactive)")
}