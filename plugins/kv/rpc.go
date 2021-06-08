package kv

import (
	"github.com/spiral/errors"
	kvv1 "github.com/spiral/roadrunner/v2/pkg/proto/kv/v1beta"
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

// Has accept []*payload.Payload proto payload with Storage and Item
func (r *rpc) Has(in *kvv1.Payload, res *map[string]bool) error {
	const op = errors.Op("rpc_has")

	if in.Storage == "" {
		return errors.E(op, errors.Str("no storage provided"))
	}

	keys := make([]string, 0, len(in.GetItems()))

	for i := 0; i < len(in.GetItems()); i++ {
		keys = append(keys, in.Items[i].Key)
	}

	if st, ok := r.storages[in.Storage]; ok {
		ret, err := st.Has(keys...)
		if err != nil {
			return err
		}

		// update the value in the pointer
		// save the result
		*res = ret
		return nil
	}

	return errors.E(op, errors.Errorf("no such storage: %s", in.GetStorage()))
}

// Set accept proto payload with Storage and Item
func (r *rpc) Set(in *kvv1.Payload, ok *bool) error {
	const op = errors.Op("rpc_set")

	if st, exists := r.storages[in.GetStorage()]; exists {
		err := st.Set(in.GetItems()...)
		if err != nil {
			return err
		}

		// save the result
		*ok = true
		return nil
	}

	*ok = false
	return errors.E(op, errors.Errorf("no such storage: %s", in.GetStorage()))
}

// MGet accept proto payload with Storage and Item
func (r *rpc) MGet(in *kvv1.Payload, res *map[string]interface{}) error {
	const op = errors.Op("rpc_mget")

	keys := make([]string, 0, len(in.GetItems()))

	for i := 0; i < len(in.GetItems()); i++ {
		keys = append(keys, in.Items[i].Key)
	}

	if st, exists := r.storages[in.GetStorage()]; exists {
		ret, err := st.MGet(keys...)
		if err != nil {
			return err
		}

		// save the result
		*res = ret
		return nil
	}

	return errors.E(op, errors.Errorf("no such storage: %s", in.GetStorage()))
}

// MExpire accept proto payload with Storage and Item
func (r *rpc) MExpire(in *kvv1.Payload, ok *bool) error {
	const op = errors.Op("rpc_mexpire")

	if st, exists := r.storages[in.GetStorage()]; exists {
		err := st.MExpire(in.GetItems()...)
		if err != nil {
			return errors.E(op, err)
		}

		// save the result
		*ok = true
		return nil
	}

	*ok = false
	return errors.E(op, errors.Errorf("no such storage: %s", in.GetStorage()))
}

// TTL accept proto payload with Storage and Item
func (r *rpc) TTL(in *kvv1.Payload, res *map[string]interface{}) error {
	const op = errors.Op("rpc_ttl")
	keys := make([]string, 0, len(in.GetItems()))

	for i := 0; i < len(in.GetItems()); i++ {
		keys = append(keys, in.Items[i].Key)
	}

	if st, exists := r.storages[in.GetStorage()]; exists {
		ret, err := st.TTL(keys...)
		if err != nil {
			return err
		}

		// save the result
		*res = ret
		return nil
	}

	return errors.E(op, errors.Errorf("no such storage: %s", in.GetStorage()))
}

// Delete accept proto payload with Storage and Item
func (r *rpc) Delete(in *kvv1.Payload, ok *bool) error {
	const op = errors.Op("rcp_delete")

	keys := make([]string, 0, len(in.GetItems()))

	for i := 0; i < len(in.GetItems()); i++ {
		keys = append(keys, in.Items[i].Key)
	}
	if st, exists := r.storages[in.GetStorage()]; exists {
		err := st.Delete(keys...)
		if err != nil {
			return errors.E(op, err)
		}

		// save the result
		*ok = true
		return nil
	}

	*ok = false
	return errors.E(op, errors.Errorf("no such storage: %s", in.GetStorage()))
}
