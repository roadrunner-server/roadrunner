package cli

import (
	"log"
	"net/http/pprof"
	"net/rpc"
	"os"
	"path/filepath"

	"github.com/spiral/errors"
	goridgeRpc "github.com/spiral/goridge/v3/pkg/rpc"
	rpcPlugin "github.com/spiral/roadrunner/v2/plugins/rpc"

	"github.com/spiral/roadrunner/v2/plugins/config"

	"net/http"

	"github.com/spf13/cobra"
	endure "github.com/spiral/endure/pkg/container"
)

var (
	// WorkDir is working directory
	WorkDir string
	// CfgFile is path to the .rr.yaml
	CfgFile string
	// Debug mode
	Debug bool
	// Container is the pointer to the Endure container
	Container *endure.Endure
	cfg       *config.Viper
	root      = &cobra.Command{
		Use:           "rr",
		SilenceErrors: true,
		SilenceUsage:  true,
		Version:       Version,
	}
)

func Execute() {
	if err := root.Execute(); err != nil {
		// exit with error, fatal invoke os.Exit(1)
		log.Fatal(err)
	}
}

func init() {
	root.PersistentFlags().StringVarP(&CfgFile, "config", "c", ".rr.yaml", "config file (default is .rr.yaml)")
	root.PersistentFlags().StringVarP(&WorkDir, "WorkDir", "w", "", "work directory")
	root.PersistentFlags().BoolVarP(&Debug, "debug", "d", false, "debug mode")
	cobra.OnInitialize(func() {
		if CfgFile != "" {
			if absPath, err := filepath.Abs(CfgFile); err == nil {
				CfgFile = absPath

				// force working absPath related to config file
				if err := os.Chdir(filepath.Dir(absPath)); err != nil {
					panic(err)
				}
			}
		}

		if WorkDir != "" {
			if err := os.Chdir(WorkDir); err != nil {
				panic(err)
			}
		}

		cfg = &config.Viper{}
		cfg.Path = CfgFile
		cfg.Prefix = "rr"

		// register config
		err := Container.Register(cfg)
		if err != nil {
			panic(err)
		}

		// if debug mode is on - run debug server
		if Debug {
			go runDebugServer()
		}
	})
}

// RPCClient is using to make a requests to the ./rr reset, ./rr workers
func RPCClient() (*rpc.Client, error) {
	rpcConfig := &rpcPlugin.Config{}

	err := cfg.Init()
	if err != nil {
		return nil, err
	}

	if !cfg.Has(rpcPlugin.PluginName) {
		return nil, errors.E("rpc service disabled")
	}

	err = cfg.UnmarshalKey(rpcPlugin.PluginName, rpcConfig)
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

// debug server
func runDebugServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	srv := http.Server{
		Addr:    ":6061",
		Handler: mux,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
