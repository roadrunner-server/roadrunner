package cli

import (
	"log"
	"net/rpc"
	"os"
	"path/filepath"

	"github.com/spiral/errors"
	goridgeRpc "github.com/spiral/goridge/v3/pkg/rpc"
	rpcPlugin "github.com/spiral/roadrunner/v2/plugins/rpc"

	"github.com/spiral/roadrunner/v2/plugins/config"

	"github.com/spf13/cobra"
	"github.com/spiral/endure"
)

var (
	WorkDir   string
	CfgFile   string
	Container *endure.Endure
	cfg       *config.Viper
	root      = &cobra.Command{
		Use:           "rr",
		SilenceErrors: true,
		SilenceUsage:  true,
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
