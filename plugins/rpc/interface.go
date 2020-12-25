package rpc

// RPCer declares the ability to create set of public RPC methods.
type RPCer interface {
	// Provides RPC methods for the given service.
	RPC() interface{}
}
