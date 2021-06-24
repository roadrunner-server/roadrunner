package kv

import (
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/plugins/logger"
	kvv1 "github.com/spiral/roadrunner/v2/proto/kv/v1beta"
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

// Has accept []*kvv1.Payload proto payload with Storage and Item
func (r *rpc) Has(in *kvv1.Request, out *kvv1.Response) error {
	const op = errors.Op("rpc_has")

	if in.GetStorage() == "" {
		return errors.E(op, errors.Str("no storage provided"))
	}

	keys := make([]string, 0, len(in.GetItems()))

	for i := 0; i < len(in.GetItems()); i++ {
		keys = append(keys, in.Items[i].Key)
	}

	if st, ok := r.storages[in.GetStorage()]; ok {
		ret, err := st.Has(keys...)
		if err != nil {
			return errors.E(op, err)
		}

		// update the value in the pointer
		// save the result
		out.Items = make([]*kvv1.Item, 0, len(ret))
		for k := range ret {
			out.Items = append(out.Items, &kvv1.Item{
				Key: k,
			})
		}
		return nil
	}

	return errors.E(op, errors.Errorf("no such storage: %s", in.GetStorage()))
}

// Set accept proto payload with Storage and Item
func (r *rpc) Set(in *kvv1.Request, _ *kvv1.Response) error {
	const op = errors.Op("rpc_set")

	if st, exists := r.storages[in.GetStorage()]; exists {
		err := st.Set(in.GetItems()...)
		if err != nil {
			return errors.E(op, err)
		}

		// save the result
		return nil
	}

	return errors.E(op, errors.Errorf("no such storage: %s", in.GetStorage()))
}

// MGet accept proto payload with Storage and Item
func (r *rpc) MGet(in *kvv1.Request, out *kvv1.Response) error {
	const op = errors.Op("rpc_mget")

	keys := make([]string, 0, len(in.GetItems()))

	for i := 0; i < len(in.GetItems()); i++ {
		keys = append(keys, in.Items[i].Key)
	}

	if st, exists := r.storages[in.GetStorage()]; exists {
		ret, err := st.MGet(keys...)
		if err != nil {
			return errors.E(op, err)
		}

		out.Items = make([]*kvv1.Item, 0, len(ret))
		for k := range ret {
			out.Items = append(out.Items, &kvv1.Item{
				Key:   k,
				Value: ret[k],
			})
		}
		return nil
	}

	return errors.E(op, errors.Errorf("no such storage: %s", in.GetStorage()))
}

// MExpire accept proto payload with Storage and Item
func (r *rpc) MExpire(in *kvv1.Request, _ *kvv1.Response) error {
	const op = errors.Op("rpc_mexpire")

	if st, exists := r.storages[in.GetStorage()]; exists {
		err := st.MExpire(in.GetItems()...)
		if err != nil {
			return errors.E(op, err)
		}

		return nil
	}

	return errors.E(op, errors.Errorf("no such storage: %s", in.GetStorage()))
}

// TTL accept proto payload with Storage and Item
func (r *rpc) TTL(in *kvv1.Request, out *kvv1.Response) error {
	const op = errors.Op("rpc_ttl")
	keys := make([]string, 0, len(in.GetItems()))

	for i := 0; i < len(in.GetItems()); i++ {
		keys = append(keys, in.Items[i].Key)
	}

	if st, exists := r.storages[in.GetStorage()]; exists {
		ret, err := st.TTL(keys...)
		if err != nil {
			return errors.E(op, err)
		}

		out.Items = make([]*kvv1.Item, 0, len(ret))
		for k := range ret {
			out.Items = append(out.Items, &kvv1.Item{
				Key:     k,
				Timeout: ret[k],
			})
		}

		return nil
	}

	return errors.E(op, errors.Errorf("no such storage: %s", in.GetStorage()))
}

// Delete accept proto payload with Storage and Item
func (r *rpc) Delete(in *kvv1.Request, _ *kvv1.Response) error {
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

		return nil
	}

	return errors.E(op, errors.Errorf("no such storage: %s", in.GetStorage()))
}

// Clear clean the storage
func (r *rpc) Clear(in *kvv1.Request, _ *kvv1.Response) error {
	const op = errors.Op("rcp_delete")

	if st, exists := r.storages[in.GetStorage()]; exists {
		err := st.Clear()
		if err != nil {
			return errors.E(op, err)
		}

		return nil
	}

	return errors.E(op, errors.Errorf("no such storage: %s", in.GetStorage()))
}
