package cli

import (
	"log"
	"net/rpc"
	"os"
	"path/filepath"

	"github.com/spiral/errors"
	goridgeRpc "github.com/spiral/goridge/v3/pkg/rpc"
	"github.com/spiral/roadrunner-plugins/logger"
	rpcPlugin "github.com/spiral/roadrunner-plugins/rpc"

	"github.com/spiral/roadrunner-plugins/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/spf13/cobra"
	"github.com/spiral/endure"
)

var (
	WorkDir   string
	CfgFile   string
	Container *endure.Endure
	Logger    *zap.Logger
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

	// todo: properly handle debug level
	Logger = initLogger()
	//endureLogger := logger.NewZapAdapter(Logger)

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

		// todo: config is global, not only for serve
		cfg = &config.Viper{}
		cfg.Path = CfgFile
		cfg.Prefix = "rr"

		err := Container.Register(cfg)
		if err != nil {
			panic(err)
		}

		err = Container.Register(&logger.ZapLogger{})
		if err != nil {
			panic(err)
		}
	})
}

// todo: improve
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

func initLogger() *zap.Logger {
	// todo: we do not need it
	cfg := zap.Config{
		Level:    zap.NewAtomicLevelAt(zap.ErrorLevel),
		Encoding: "console",
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey:    "message",
			LevelKey:      "level",
			TimeKey:       "time",
			CallerKey:     "caller",
			NameKey:       "name",
			StacktraceKey: "stack",
			EncodeLevel:   zapcore.CapitalLevelEncoder,
			EncodeTime:    zapcore.ISO8601TimeEncoder,
			EncodeCaller:  zapcore.ShortCallerEncoder,
		},
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}

	l, err := cfg.Build(zap.AddCaller())
	if err != nil {
		panic(err)
	}

	return l
}
