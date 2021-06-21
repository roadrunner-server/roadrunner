package jobs

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_Map_All(t *testing.T) {
	m := initDispatcher(map[string]*Options{"default": {Pipeline: "default"}})
	assert.Equal(t, "default", m.match(&Job{Job: "default"}).Pipeline)
}

func Test_Map_Miss(t *testing.T) {
	m := initDispatcher(map[string]*Options{"some.*": {Pipeline: "default"}})

	assert.Nil(t, m.match(&Job{Job: "miss"}))
}

func Test_Map_Best(t *testing.T) {
	m := initDispatcher(map[string]*Options{
		"some.*":       {Pipeline: "default"},
		"some.other.*": {Pipeline: "other"},
	})

	assert.Equal(t, "default", m.match(&Job{Job: "some"}).Pipeline)
	assert.Equal(t, "default", m.match(&Job{Job: "some.any"}).Pipeline)
	assert.Equal(t, "other", m.match(&Job{Job: "some.other"}).Pipeline)
	assert.Equal(t, "other", m.match(&Job{Job: "some.other.job"}).Pipeline)
}

func Test_Map_BestUpper(t *testing.T) {
	m := initDispatcher(map[string]*Options{
		"some.*":       {Pipeline: "default"},
		"some.Other.*": {Pipeline: "other"},
	})

	assert.Equal(t, "default", m.match(&Job{Job: "some"}).Pipeline)
	assert.Equal(t, "default", m.match(&Job{Job: "some.any"}).Pipeline)
	assert.Equal(t, "other", m.match(&Job{Job: "some.OTHER"}).Pipeline)
	assert.Equal(t, "other", m.match(&Job{Job: "Some.other.job"}).Pipeline)
}

func Test_Map_BestReversed(t *testing.T) {
	m := initDispatcher(map[string]*Options{
		"some.*":       {Pipeline: "default"},
		"some.other.*": {Pipeline: "other"},
	})

	assert.Equal(t, "other", m.match(&Job{Job: "some.other.job"}).Pipeline)
	assert.Equal(t, "other", m.match(&Job{Job: "some.other"}).Pipeline)
	assert.Equal(t, "default", m.match(&Job{Job: "some.any"}).Pipeline)
	assert.Equal(t, "default", m.match(&Job{Job: "some"}).Pipeline)
}
