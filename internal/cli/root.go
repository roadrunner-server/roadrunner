package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/roadrunner-server/roadrunner/v2/internal/cli/reset"
	"github.com/roadrunner-server/roadrunner/v2/internal/cli/serve"
	"github.com/roadrunner-server/roadrunner/v2/internal/cli/workers"
	dbg "github.com/roadrunner-server/roadrunner/v2/internal/debug"
	"github.com/roadrunner-server/roadrunner/v2/internal/meta"

	"github.com/joho/godotenv"
	"github.com/roadrunner-server/config/v2"
	"github.com/spf13/cobra"
)

// NewCommand creates root command.
func NewCommand(cmdName string) *cobra.Command { //nolint:funlen
	const envDotenv = "DOTENV_PATH" // env var name: path to the .env file

	var ( // flag values
		cfgFile  string   // path to the .rr.yaml
		workDir  string   // working directory
		dotenv   string   // path to the .env file
		debug    bool     // debug mode
		override []string // override config values
	)

	var configPlugin = &config.Plugin{} // will be overwritten on pre-run action

	cmd := &cobra.Command{
		Use:           cmdName,
		Short:         "High-performance PHP application server, load-balancer and process manager",
		SilenceErrors: true,
		SilenceUsage:  true,
		Version:       fmt.Sprintf("%s (build time: %s, %s)", meta.Version(), meta.BuildTime(), runtime.Version()),
		PersistentPreRunE: func(*cobra.Command, []string) error {
			if cfgFile != "" {
				if absPath, err := filepath.Abs(cfgFile); err == nil {
					cfgFile = absPath // switch config path to the absolute

					// force working absPath related to config file
					if err = os.Chdir(filepath.Dir(absPath)); err != nil {
						return err
					}
				}
			}

			if workDir != "" {
				if err := os.Chdir(workDir); err != nil {
					return err
				}
			}

			if v, ok := os.LookupEnv(envDotenv); ok { // read path to the dotenv file from environment variable
				dotenv = v
			}

			if dotenv != "" {
				err := godotenv.Load(dotenv)
				if err != nil {
					return err
				}
			}

			cfg := &config.Plugin{Path: cfgFile, Prefix: "rr", Flags: override}
			if err := cfg.Init(); err != nil {
				return err
			}

			if debug {
				srv := dbg.NewServer()
				go func() { _ = srv.Start(":6061") }() // TODO implement graceful server stopping
			}

			// overwrite
			*configPlugin = *cfg

			return nil
		},
	}

	f := cmd.PersistentFlags()

	f.StringVarP(&cfgFile, "config", "c", ".rr.yaml", "config file")
	f.StringVarP(&workDir, "WorkDir", "w", "", "working directory") // TODO change to `workDir`?
	f.StringVarP(&dotenv, "dotenv", "", "", fmt.Sprintf("dotenv file [$%s]", envDotenv))
	f.BoolVarP(&debug, "debug", "d", false, "debug mode")
	f.StringArrayVarP(&override, "override", "o", nil, "override config value (dot.notation=value)")

	cmd.AddCommand(
		workers.NewCommand(configPlugin),
		reset.NewCommand(configPlugin),
		serve.NewCommand(configPlugin),
	)

	return cmd
}
