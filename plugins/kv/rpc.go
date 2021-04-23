package kv

import (
	"unsafe"

	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/plugins/kv/payload/generated"
	"github.com/spiral/roadrunner/v2/plugins/logger"
)

// Wrapper for the plugin
type rpc struct {
	// all available storages
	storages map[string]Storage
	// svc is a plugin implementing Storage interface
	srv *Plugin
	// Logger
	log logger.Logger
}

// Has accept []byte flatbuffers payload with Storage and Item
func (r *rpc) Has(in []byte, res *map[string]bool) error {
	const op = errors.Op("rpc_has")
	dataRoot := generated.GetRootAsPayload(in, 0)
	l := dataRoot.ItemsLength()

	keys := make([]string, 0, l)

	tmpItem := &generated.Item{}
	for i := 0; i < l; i++ {
		if !dataRoot.Items(tmpItem, i) {
			continue
		}
		keys = append(keys, strConvert(tmpItem.Key()))
	}

	if st, ok := r.storages[strConvert(dataRoot.Storage())]; ok {
		ret, err := st.Has(keys...)
		if err != nil {
			return err
		}

		// update the value in the pointer
		// save the result
		*res = ret
		return nil
	}

	return errors.E(op, errors.Errorf("no such storage: %s", dataRoot.Storage()))
}

// Set accept []byte flatbuffers payload with Storage and Item
func (r *rpc) Set(in []byte, ok *bool) error {
	const op = errors.Op("rpc_set")

	dataRoot := generated.GetRootAsPayload(in, 0)
	l := dataRoot.ItemsLength()

	items := make([]Item, 0, dataRoot.ItemsLength())
	tmpItem := &generated.Item{}

	for i := 0; i < l; i++ {
		if !dataRoot.Items(tmpItem, i) {
			continue
		}

		itc := Item{
			Key:   string(tmpItem.Key()),
			Value: string(tmpItem.Value()),
			TTL:   string(tmpItem.Timeout()),
		}

		items = append(items, itc)
	}

	if st, exists := r.storages[strConvert(dataRoot.Storage())]; exists {
		err := st.Set(items...)
		if err != nil {
			return err
		}

		// save the result
		*ok = true
		return nil
	}

	*ok = false
	return errors.E(op, errors.Errorf("no such storage: %s", dataRoot.Storage()))
}

// MGet accept []byte flatbuffers payload with Storage and Item
func (r *rpc) MGet(in []byte, res *map[string]interface{}) error {
	const op = errors.Op("rpc_mget")

	dataRoot := generated.GetRootAsPayload(in, 0)
	l := dataRoot.ItemsLength()
	keys := make([]string, 0, l)
	tmpItem := &generated.Item{}

	for i := 0; i < l; i++ {
		if !dataRoot.Items(tmpItem, i) {
			continue
		}
		keys = append(keys, string(tmpItem.Key()))
	}

	if st, exists := r.storages[strConvert(dataRoot.Storage())]; exists {
		ret, err := st.MGet(keys...)
		if err != nil {
			return err
		}

		// save the result
		*res = ret
		return nil
	}

	return errors.E(op, errors.Errorf("no such storage: %s", dataRoot.Storage()))
}

// MExpire accept []byte flatbuffers payload with Storage and Item
func (r *rpc) MExpire(in []byte, ok *bool) error {
	const op = errors.Op("rpc_mexpire")

	dataRoot := generated.GetRootAsPayload(in, 0)
	l := dataRoot.ItemsLength()

	// when unmarshalling the keys, simultaneously, fill up the slice with items
	items := make([]Item, 0, l)
	tmpItem := &generated.Item{}
	for i := 0; i < l; i++ {
		if !dataRoot.Items(tmpItem, i) {
			continue
		}

		itc := Item{
			Key: string(tmpItem.Key()),
			// we set up timeout on the keys, so, value here is redundant
			Value: "",
			TTL:   string(tmpItem.Timeout()),
		}

		items = append(items, itc)
	}

	if st, exists := r.storages[strConvert(dataRoot.Storage())]; exists {
		err := st.MExpire(items...)
		if err != nil {
			return errors.E(op, err)
		}

		// save the result
		*ok = true
		return nil
	}

	*ok = false
	return errors.E(op, errors.Errorf("no such storage: %s", dataRoot.Storage()))
}

// TTL accept []byte flatbuffers payload with Storage and Item
func (r *rpc) TTL(in []byte, res *map[string]interface{}) error {
	const op = errors.Op("rpc_ttl")
	dataRoot := generated.GetRootAsPayload(in, 0)
	l := dataRoot.ItemsLength()
	keys := make([]string, 0, l)
	tmpItem := &generated.Item{}

	for i := 0; i < l; i++ {
		if !dataRoot.Items(tmpItem, i) {
			continue
		}
		keys = append(keys, string(tmpItem.Key()))
	}

	if st, exists := r.storages[strConvert(dataRoot.Storage())]; exists {
		ret, err := st.TTL(keys...)
		if err != nil {
			return err
		}

		// save the result
		*res = ret
		return nil
	}

	return errors.E(op, errors.Errorf("no such storage: %s", dataRoot.Storage()))
}

// Delete accept []byte flatbuffers payload with Storage and Item
func (r *rpc) Delete(in []byte, ok *bool) error {
	const op = errors.Op("rcp_delete")
	dataRoot := generated.GetRootAsPayload(in, 0)
	l := dataRoot.ItemsLength()
	keys := make([]string, 0, l)
	tmpItem := &generated.Item{}

	for i := 0; i < l; i++ {
		if !dataRoot.Items(tmpItem, i) {
			continue
		}
		keys = append(keys, string(tmpItem.Key()))
	}
	if st, exists := r.storages[strConvert(dataRoot.Storage())]; exists {
		err := st.Delete(keys...)
		if err != nil {
			return errors.E(op, err)
		}

		// save the result
		*ok = true
		return nil
	}

	*ok = false
	return errors.E(op, errors.Errorf("no such storage: %s", dataRoot.Storage()))
}

func strConvert(s []byte) string {
	return *(*string)(unsafe.Pointer(&s))
}
