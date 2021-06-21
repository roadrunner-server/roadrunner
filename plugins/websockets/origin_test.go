package websockets

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_Origin(t *testing.T) {
	cfg := &Config{
		AllowedOrigin: "*",
		Broker:        "any",
	}

	err := cfg.InitDefault()
	assert.NoError(t, err)

	assert.True(t, isOriginAllowed("http://some.some.some.sssome", cfg))
	assert.True(t, isOriginAllowed("http://", cfg))
	assert.True(t, isOriginAllowed("http://google.com", cfg))
	assert.True(t, isOriginAllowed("ws://*", cfg))
	assert.True(t, isOriginAllowed("*", cfg))
	assert.True(t, isOriginAllowed("you are bad programmer", cfg)) // True :(
	assert.True(t, isOriginAllowed("****", cfg))
	assert.True(t, isOriginAllowed("asde!@#!!@#!%", cfg))
	assert.True(t, isOriginAllowed("http://*.domain.com", cfg))
}

func TestConfig_OriginWildCard(t *testing.T) {
	cfg := &Config{
		AllowedOrigin: "https://*my.site.com",
		Broker:        "any",
	}

	err := cfg.InitDefault()
	assert.NoError(t, err)

	assert.True(t, isOriginAllowed("https://my.site.com", cfg))
	assert.False(t, isOriginAllowed("http://", cfg))
	assert.False(t, isOriginAllowed("http://google.com", cfg))
	assert.False(t, isOriginAllowed("ws://*", cfg))
	assert.False(t, isOriginAllowed("*", cfg))
	assert.False(t, isOriginAllowed("you are bad programmer", cfg)) // True :(
	assert.False(t, isOriginAllowed("****", cfg))
	assert.False(t, isOriginAllowed("asde!@#!!@#!%", cfg))
	assert.False(t, isOriginAllowed("http://*.domain.com", cfg))

	assert.False(t, isOriginAllowed("https://*site.com", cfg))
	assert.True(t, isOriginAllowed("https://some.my.site.com", cfg))
}

func TestConfig_OriginWildCard2(t *testing.T) {
	cfg := &Config{
		AllowedOrigin: "https://my.*.com",
		Broker:        "any",
	}

	err := cfg.InitDefault()
	assert.NoError(t, err)

	assert.True(t, isOriginAllowed("https://my.site.com", cfg))
	assert.False(t, isOriginAllowed("http://", cfg))
	assert.False(t, isOriginAllowed("http://google.com", cfg))
	assert.False(t, isOriginAllowed("ws://*", cfg))
	assert.False(t, isOriginAllowed("*", cfg))
	assert.False(t, isOriginAllowed("you are bad programmer", cfg)) // True :(
	assert.False(t, isOriginAllowed("****", cfg))
	assert.False(t, isOriginAllowed("asde!@#!!@#!%", cfg))
	assert.False(t, isOriginAllowed("http://*.domain.com", cfg))

	assert.False(t, isOriginAllowed("https://*site.com", cfg))
	assert.True(t, isOriginAllowed("https://my.bad.com", cfg))
}
