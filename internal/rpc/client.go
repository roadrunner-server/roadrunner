// Package rpc contains wrapper around RPC client ONLY for internal usage.
package rpc

import (
	"net/rpc"

	"github.com/roadrunner-server/errors"
	goridgeRpc "github.com/roadrunner-server/goridge/v3/pkg/rpc"
	rpcPlugin "github.com/roadrunner-server/rpc/v2"
	"github.com/spf13/viper"
)

// NewClient creates client ONLY for internal usage (communication between our application with RR side).
// Client will be connected to the RPC.
func NewClient(cfg string) (*rpc.Client, error) {
	v := viper.New()
	v.SetConfigFile(cfg)

	err := v.ReadInConfig()
	if err != nil {
		return nil, err
	}

	if !v.IsSet(rpcPlugin.PluginName) {
		return nil, errors.E("rpc service disabled")
	}

	rpcConfig := &rpcPlugin.Config{}

	err = v.UnmarshalKey(rpcPlugin.PluginName, rpcConfig)
	if err != nil {
		return nil, err
	}

	rpcConfig.InitDefaults()

	conn, err := rpcConfig.Dialer()
	if err != nil {
		return nil, err
	}

	return rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn)), nil
}
