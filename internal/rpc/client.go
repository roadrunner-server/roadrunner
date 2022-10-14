// Package rpc contains wrapper around RPC client ONLY for internal usage.
// Should be in sync with the RPC plugin
package rpc

import (
	"errors"
	"fmt"
	"net"
	"net/rpc"
	"os"
	"strings"

	goridgeRpc "github.com/roadrunner-server/goridge/v3/pkg/rpc"
	rpcPlugin "github.com/roadrunner-server/rpc/v3"
	"github.com/spf13/viper"
)

const (
	prefix string = "rr"
	rpcKey string = "rpc.listen"
)

// NewClient creates client ONLY for internal usage (communication between our application with RR side).
// Client will be connected to the RPC.
func NewClient(cfg string, flags []string) (*rpc.Client, error) {
	v := viper.New()
	v.AutomaticEnv()
	v.SetEnvPrefix(prefix)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.SetConfigFile(cfg)

	err := v.ReadInConfig()
	if err != nil {
		return nil, err
	}

	// automatically inject ENV variables using ${ENV} pattern
	for _, key := range v.AllKeys() {
		val := v.Get(key)
		if s, ok := val.(string); ok {
			v.Set(key, os.ExpandEnv(s))
		}
	}

	// override config Flags
	if len(flags) > 0 {
		for _, f := range flags {
			key, val, errP := parseFlag(f)
			if errP != nil {
				return nil, errP
			}

			v.Set(key, val)
		}
	}

	// rpc.listen might be set by the -o flags or env variable
	if !v.IsSet(rpcPlugin.PluginName) {
		return nil, errors.New("rpc service not specified in the configuration. Tip: add\n rpc:\n\r listen: rr_rpc_address")
	}

	conn, err := Dialer(v.GetString(rpcKey))
	if err != nil {
		return nil, err
	}

	return rpc.NewClientWithCodec(goridgeRpc.NewClientCodec(conn)), nil
}

// Dialer creates rpc socket Dialer.
func Dialer(addr string) (net.Conn, error) {
	dsn := strings.Split(addr, "://")
	if len(dsn) != 2 {
		return nil, errors.New("invalid socket DSN (tcp://:6001, unix://file.sock)")
	}

	return net.Dial(dsn[0], dsn[1])
}

func parseFlag(flag string) (string, string, error) {
	if !strings.Contains(flag, "=") {
		return "", "", fmt.Errorf("invalid flag `%s`", flag)
	}

	parts := strings.SplitN(strings.TrimLeft(flag, " \"'`"), "=", 2)
	if len(parts) < 2 {
		return "", "", errors.New("usage: -o key=value")
	}

	if parts[0] == "" {
		return "", "", errors.New("key should not be empty")
	}

	if parts[1] == "" {
		return "", "", errors.New("value should not be empty")
	}

	return strings.Trim(parts[0], " \n\t"), parseValue(strings.Trim(parts[1], " \n\t")), nil
}

func parseValue(value string) string {
	escape := []rune(value)[0]

	if escape == '"' || escape == '\'' || escape == '`' {
		value = strings.Trim(value, string(escape))
		value = strings.ReplaceAll(value, fmt.Sprintf("\\%s", string(escape)), string(escape))
	}

	return value
}
