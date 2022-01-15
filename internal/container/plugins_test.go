package container_test

import (
	"reflect"
	"testing"

	"github.com/spiral/roadrunner-binary/v2/internal/container"
)

func TestPlugins(t *testing.T) {
	for _, p := range container.Plugins() {
		if p == nil {
			t.Error("plugin cannot be nil")
		}

		if pk := reflect.TypeOf(p).Kind(); pk != reflect.Ptr && pk != reflect.Struct {
			t.Errorf("plugin %v must be a structure or pointer to the structure", p)
		}
	}
}
