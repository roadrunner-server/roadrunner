package beanstalk

import (
	"bytes"
	"encoding/gob"
	"github.com/spiral/jobs/v2"
)

func pack(j *jobs.Job) ([]byte, error) {
	b := new(bytes.Buffer)
	err := gob.NewEncoder(b).Encode(j)
	if err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

func unpack(data []byte) (*jobs.Job, error) {
	j := &jobs.Job{}
	err := gob.NewDecoder(bytes.NewBuffer(data)).Decode(j)

	return j, err
}
