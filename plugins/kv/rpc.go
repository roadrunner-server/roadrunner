package kv

import (
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/plugins/logger"
)

// Wrapper for the plugin
type RPCServer struct {
	// svc is a plugin implementing Storage interface
	svc Storage
	// Logger
	log logger.Logger
}

// NewRPCServer construct RPC server for the particular plugin
func NewRPCServer(srv Storage, log logger.Logger) *RPCServer {
	return &RPCServer{
		svc: srv,
		log: log,
	}
}

// data Data
func (r *RPCServer) Has(in []string, res *map[string]bool) error {
	const op = errors.Op("rpc server Has")
	ret, err := r.svc.Has(in...)
	if err != nil {
		return errors.E(op, err)
	}

	// update the value in the pointer
	*res = ret
	return nil
}

// in SetData
func (r *RPCServer) Set(in []Item, ok *bool) error {
	const op = errors.Op("rpc server Set")

	err := r.svc.Set(in...)
	if err != nil {
		return errors.E(op, err)
	}

	*ok = true
	return nil
}

// in Data
func (r *RPCServer) MGet(in []string, res *map[string]interface{}) error {
	const op = errors.Op("rpc server MGet")
	ret, err := r.svc.MGet(in...)
	if err != nil {
		return errors.E(op, err)
	}

	// update return value
	*res = ret
	return nil
}

// in Data
func (r *RPCServer) MExpire(in []Item, ok *bool) error {
	const op = errors.Op("rpc server MExpire")

	err := r.svc.MExpire(in...)
	if err != nil {
		return errors.E(op, err)
	}

	*ok = true
	return nil
}

// in Data
func (r *RPCServer) TTL(in []string, res *map[string]interface{}) error {
	const op = errors.Op("rpc server TTL")

	ret, err := r.svc.TTL(in...)
	if err != nil {
		return errors.E(op, err)
	}

	*res = ret
	return nil
}

// in Data
func (r *RPCServer) Delete(in []string, ok *bool) error {
	const op = errors.Op("rpc server Delete")
	err := r.svc.Delete(in...)
	if err != nil {
		return errors.E(op, err)
	}
	*ok = true
	return nil
}

// in string, storages
func (r *RPCServer) Close(storage string, ok *bool) error {
	const op = errors.Op("rpc server Close")
	err := r.svc.Close()
	if err != nil {
		return errors.E(op, err)
	}
	*ok = true

	return nil
}
