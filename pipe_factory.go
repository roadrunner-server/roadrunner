package roadrunner

import (
	"github.com/spiral/goridge"
	"os/exec"
)

// PipeFactory connects to workers using standard streams (STDIN, STDOUT pipes).
type PipeFactory struct {
}

// NewPipeFactory returns new factory instance and starts listening
func NewPipeFactory() *PipeFactory {
	return &PipeFactory{}
}

// NewWorker creates worker and connects it to appropriate relay or returns error
func (f *PipeFactory) NewWorker(cmd *exec.Cmd) (w *Worker, err error) {
	w, err = NewWorker(cmd)
	if err != nil {
		return nil, err
	}

	in, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	out, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	if err := w.Start(); err != nil {
		return nil, err
	}

	w.attach(goridge.NewPipeRelay(in, out))

	return w, nil
}

// Close closes all open factory descriptors.
func (f *PipeFactory) Close() error {
	return nil
}
