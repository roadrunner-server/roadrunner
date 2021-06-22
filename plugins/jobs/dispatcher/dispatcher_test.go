package dispatcher

import (
	"testing"

	"github.com/spiral/roadrunner/v2/plugins/jobs/structs"
	"github.com/stretchr/testify/assert"
)

func Test_Map_All(t *testing.T) {
	m := initDispatcher(map[string]*structs.Options{"default": {Pipeline: "default"}})
	assert.Equal(t, "default", m.Match(&structs.Job{Job: "default"}).Pipeline)
}

func Test_Map_Miss(t *testing.T) {
	m := initDispatcher(map[string]*structs.Options{"some.*": {Pipeline: "default"}})

	assert.Nil(t, m.Match(&structs.Job{Job: "miss"}))
}

func Test_Map_Best(t *testing.T) {
	m := initDispatcher(map[string]*structs.Options{
		"some.*":       {Pipeline: "default"},
		"some.other.*": {Pipeline: "other"},
	})

	assert.Equal(t, "default", m.Match(&structs.Job{Job: "some"}).Pipeline)
	assert.Equal(t, "default", m.Match(&structs.Job{Job: "some.any"}).Pipeline)
	assert.Equal(t, "other", m.Match(&structs.Job{Job: "some.other"}).Pipeline)
	assert.Equal(t, "other", m.Match(&structs.Job{Job: "some.other.job"}).Pipeline)
}

func Test_Map_BestUpper(t *testing.T) {
	m := initDispatcher(map[string]*structs.Options{
		"some.*":       {Pipeline: "default"},
		"some.Other.*": {Pipeline: "other"},
	})

	assert.Equal(t, "default", m.Match(&structs.Job{Job: "some"}).Pipeline)
	assert.Equal(t, "default", m.Match(&structs.Job{Job: "some.any"}).Pipeline)
	assert.Equal(t, "other", m.Match(&structs.Job{Job: "some.OTHER"}).Pipeline)
	assert.Equal(t, "other", m.Match(&structs.Job{Job: "Some.other.job"}).Pipeline)
}

func Test_Map_BestReversed(t *testing.T) {
	m := initDispatcher(map[string]*structs.Options{
		"some.*":       {Pipeline: "default"},
		"some.other.*": {Pipeline: "other"},
	})

	assert.Equal(t, "other", m.Match(&structs.Job{Job: "some.other.job"}).Pipeline)
	assert.Equal(t, "other", m.Match(&structs.Job{Job: "some.other"}).Pipeline)
	assert.Equal(t, "default", m.Match(&structs.Job{Job: "some.any"}).Pipeline)
	assert.Equal(t, "default", m.Match(&structs.Job{Job: "some"}).Pipeline)
}
