package roadrunner

import (
	"github.com/pkg/errors"
	"github.com/spiral/goridge"
	"github.com/stretchr/testify/assert"
	"testing"
)

type relayMock struct {
	error   bool
	payload string
}

func (r *relayMock) Send(data []byte, flags byte) (err error) {
	if r.error {
		return errors.New("send error")
	}

	return nil
}

func (r *relayMock) Receive() (data []byte, p goridge.Prefix, err error) {
	return []byte(r.payload), goridge.NewPrefix().WithFlag(goridge.PayloadControl), nil
}

func (r *relayMock) Close() error {
	return nil
}

func Test_Protocol_Errors(t *testing.T) {
	err := sendControl(&relayMock{}, make(chan int))
	assert.Error(t, err)
}

func Test_Protocol_FetchPID(t *testing.T) {
	pid, err := fetchPID(&relayMock{error: false, payload: "{\"pid\":100}"})
	assert.NoError(t, err)
	assert.Equal(t, 100, pid)

	_, err = fetchPID(&relayMock{error: true, payload: "{\"pid\":100}"})
	assert.Error(t, err)

	_, err = fetchPID(&relayMock{error: false, payload: "{\"pid:100"})
	assert.Error(t, err)
}
