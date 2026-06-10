package rpc

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	rpcPlugin "github.com/roadrunner-server/rpc/v6"
	"github.com/spf13/viper"
)

const (
	rpcKey string = "rpc.listen"
	// default envs
	envDefault = ":-"
)

// NewClient creates a client ONLY for internal usage (communication between our application and the RR side).
// It returns the base URL of the RoadRunner Connect-RPC plane together with an *http.Client configured for it.
func NewClient(cfg string, flags []string) (string, *http.Client, error) {
	v := viper.New()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.SetConfigFile(cfg)

	err := v.ReadInConfig()
	if err != nil {
		return "", nil, err
	}

	// override config Flags
	if len(flags) > 0 {
		for _, f := range flags {
			key, val, errP := parseFlag(f)
			if errP != nil {
				return "", nil, errP
			}

			v.Set(key, val)
		}
	}

	ver := v.Get(versionKey)
	if ver == nil {
		return "", nil, fmt.Errorf("rr configuration file should contain a version e.g: version: 3")
	}

	if _, ok := ver.(string); !ok {
		return "", nil, fmt.Errorf("version should be a string: `version: \"3\"`, actual type is: %T", ver)
	}

	err = handleInclude(ver.(string), v)
	if err != nil {
		return "", nil, fmt.Errorf("failed to handle includes: %w", err)
	}

	// rpc.listen might be set by the -o flags or env variable
	if !v.IsSet(rpcPlugin.PluginName) {
		return "", nil, errors.New("rpc service not specified in the configuration. Tip: add\n rpc:\n\r listen: rr_rpc_address")
	}

	return HTTPClient(v.GetString(rpcKey))
}

// HTTPClient maps an rpc DSN (tcp://host:port, unix://file.sock) to the base
// URL of the Connect-RPC plane plus an *http.Client able to reach it. The
// endpoint is probed eagerly so an unreachable server fails fast, matching the
// previous eager-dial behavior.
func HTTPClient(addr string) (string, *http.Client, error) {
	scheme, rest, ok := strings.Cut(addr, "://")
	if !ok {
		return "", nil, errors.New("invalid socket DSN (tcp://:6001, unix://file.sock)")
	}

	probe := net.Dialer{Timeout: time.Second * 5}
	conn, err := probe.DialContext(context.Background(), scheme, rest)
	if err != nil {
		return "", nil, err
	}
	_ = conn.Close()

	switch scheme {
	case "tcp":
		return "http://" + rest, http.DefaultClient, nil
	case "unix":
		transport := &http.Transport{
			DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
				var d net.Dialer
				return d.DialContext(ctx, "unix", rest)
			},
		}

		return "http://localhost", &http.Client{Transport: transport}, nil
	default:
		return "", nil, fmt.Errorf("unsupported rpc DSN scheme: %q", scheme)
	}
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

// ExpandVal replaces ${var} or $var in the string based on the mapping function.
// For example, os.ExpandEnv(s) is equivalent to os.Expand(s, os.Getenv).
func ExpandVal(s string, mapping func(string) string) string {
	var buf []byte
	// ${} is all ASCII, so bytes are fine for this operation.
	i := 0
	for j := 0; j < len(s); j++ {
		if s[j] == '$' && j+1 < len(s) {
			if buf == nil {
				buf = make([]byte, 0, 2*len(s))
			}
			buf = append(buf, s[i:j]...)
			name, w := getShellName(s[j+1:])
			if name == "" && w > 0 { //nolint:revive
				// Encountered invalid syntax; eat the
				// characters.
			} else if name == "" {
				// Valid syntax, but $ was not followed by a
				// name. Leave the dollar character untouched.
				buf = append(buf, s[j])
				// parse default syntax
			} else if idx := strings.Index(s, envDefault); idx != -1 {
				// ${key:=default} or ${key:-val}
				substr := strings.Split(name, envDefault)
				if len(substr) != 2 {
					return ""
				}

				key := substr[0]
				defaultVal := substr[1]

				res := mapping(key)
				if res == "" {
					res = defaultVal
				}

				buf = append(buf, res...)
			} else {
				buf = append(buf, mapping(name)...)
			}
			j += w
			i = j + 1
		}
	}
	if buf == nil {
		return s
	}
	return string(buf) + s[i:]
}

// getShellName returns the name that begins the string and the number of bytes
// consumed to extract it. If the name is enclosed in {}, it's part of a ${}
// expansion, and two more bytes are needed than the length of the name.
func getShellName(s string) (string, int) {
	switch {
	case s[0] == '{':
		if len(s) > 2 && isShellSpecialVar(s[1]) && s[2] == '}' {
			return s[1:2], 3
		}
		// Scan to closing brace
		for i := 1; i < len(s); i++ {
			if s[i] == '}' {
				if i == 1 {
					return "", 2 // Bad syntax; eat "${}"
				}
				return s[1:i], i + 1
			}
		}
		return "", 1 // Bad syntax; eat "${"
	case isShellSpecialVar(s[0]):
		return s[0:1], 1
	}
	// Scan alphanumerics.
	var i int
	for i = 0; i < len(s) && isAlphaNum(s[i]); i++ { //nolint:revive

	}
	return s[:i], i
}

func expandEnvViper(v *viper.Viper) {
	for _, key := range v.AllKeys() {
		val := v.Get(key)
		switch t := val.(type) {
		case string:
			// for string expand it
			v.Set(key, parseEnvDefault(t))
		case []any:
			// for slice -> check if it's a slice of strings
			strArr := make([]string, 0, len(t))
			for i := range t {
				if valStr, ok := t[i].(string); ok {
					strArr = append(strArr, parseEnvDefault(valStr))
					continue
				}

				v.Set(key, val)
			}

			// we should set the whole array
			if len(strArr) > 0 {
				v.Set(key, strArr)
			}
		default:
			v.Set(key, val)
		}
	}
}

// isShellSpecialVar reports whether the character identifies a special
// shell variable such as $*.
func isShellSpecialVar(c uint8) bool {
	switch c {
	case '*', '#', '$', '@', '!', '?', '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return true
	}
	return false
}

// isAlphaNum reports whether the byte is an ASCII letter, number, or underscore.
func isAlphaNum(c uint8) bool {
	return c == '_' || '0' <= c && c <= '9' || 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z'
}

func parseEnvDefault(val string) string {
	// tcp://127.0.0.1:${RPC_PORT:-36643}
	// for envs like this, part would be tcp://127.0.0.1:
	return ExpandVal(val, os.Getenv)
}
