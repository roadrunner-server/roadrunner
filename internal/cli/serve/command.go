package serve

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/roadrunner-server/endure/v2"
	"github.com/roadrunner-server/roadrunner/v2024/container"
	"github.com/roadrunner-server/roadrunner/v2024/internal/meta"
	"github.com/roadrunner-server/roadrunner/v2024/internal/sdnotify"
	"gopkg.in/yaml.v3"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"

	configImpl "github.com/roadrunner-server/config/v5"
	"github.com/roadrunner-server/errors"
	"github.com/spf13/cobra"
)

// envVarPattern matches ${var} or $var or ${var:-default} patterns
var envVarPattern = regexp.MustCompile(`\$\{([^{}:\-]+)(?::-([^{}]+))?\}|\$([A-Za-z0-9_]+)`)

// expandEnvVars replaces environment variables in the input string
func expandEnvVars(input string) string {
	return envVarPattern.ReplaceAllStringFunc(input, func(match string) string {
		// Case 1: ${VAR:-default}
		if strings.Contains(match, ":-") {
			parts := strings.Split(match[2:len(match)-1], ":-")
			value := os.Getenv(parts[0])
			if value != "" {
				return value
			}
			return parts[1]
		}

		// Case 2: ${VAR} or $VAR
		varName := match
		if strings.HasPrefix(match, "${") {
			varName = match[2 : len(match)-1]
		} else {
			varName = match[1:]
		}

		if value := os.Getenv(varName); value != "" {
			return value
		}

		// Return original if not found
		return match
	})
}

// processConfig reads the config file, processes environment variables, and returns a path to the processed config
func processConfig(cfgFile string) (string, error) {
	content, err := os.ReadFile(cfgFile)
	if err != nil {
		return "", err
	}

	// Check if envfile is specified
	var cfg map[string]interface{}
	if err := yaml.Unmarshal(content, &cfg); err != nil {
		return "", err
	}

	envFile, ok := cfg["envfile"].(string)
	if !ok || envFile == "" {
		return "", nil
	}

	// Load env file if specified
	if !filepath.IsAbs(envFile) {
		envFile = filepath.Join(filepath.Dir(cfgFile), envFile)
	}

	err = godotenv.Load(envFile)
	if err != nil {
		return "", err
	}

	// Perform environment variable substitution
	expandedContent := expandEnvVars(string(content))

	// Create temporary file with processed content
	tmpFile, err := os.CreateTemp("", "rr-processed-*.yaml")
	if err != nil {
		return "", err
	}
	defer func() {
		_ = tmpFile.Close()
	}()

	if err = os.WriteFile(tmpFile.Name(), []byte(expandedContent), 0644); err != nil {
		return "", err
	}

	return tmpFile.Name(), nil
}

func NewCommand(override *[]string, cfgFile *string, silent *bool, experimental *bool) *cobra.Command { //nolint:funlen

	return &cobra.Command{
		Use:   "serve",
		Short: "Start RoadRunner server",
		RunE: func(*cobra.Command, []string) error {
			const op = errors.Op("handle_serve_command")
			// just to be safe
			if cfgFile == nil {
				return errors.E(op, errors.Str("no configuration file provided"))
			}

			// Process config and get temporary file path
			tempFile, err := processConfig(*cfgFile)
			if err != nil {
				return errors.E(op, err)
			}
			if len(tempFile) > 0 {
				*cfgFile = tempFile
				defer func() {
					_ = os.Remove(tempFile)
				}()
			}

			// create endure container config
			containerCfg, err := container.NewConfig(*cfgFile)
			if err != nil {
				return errors.E(op, err)
			}

			cfg := &configImpl.Plugin{
				Path:                 *cfgFile,
				Timeout:              containerCfg.GracePeriod,
				Flags:                *override,
				Version:              meta.Version(),
				ExperimentalFeatures: *experimental,
			}

			endureOptions := []endure.Options{
				endure.GracefulShutdownTimeout(containerCfg.GracePeriod),
			}

			if containerCfg.PrintGraph {
				endureOptions = append(endureOptions, endure.Visualize())
			}

			// create endure container
			ll, err := container.ParseLogLevel(containerCfg.LogLevel)
			if err != nil {
				if !*silent {
					fmt.Println(fmt.Errorf("[WARN] Failed to parse log level, using default (error): %w", err))
				}
			}
			cont := endure.New(ll, endureOptions...)

			// register plugins
			err = cont.RegisterAll(append(container.Plugins(), cfg)...)
			if err != nil {
				return errors.E(op, err)
			}

			// init container and all services
			err = cont.Init()
			if err != nil {
				return errors.E(op, err)
			}

			// start serving the graph
			errCh, err := cont.Serve()
			if err != nil {
				return errors.E(op, err)
			}

			oss, stop := make(chan os.Signal, 5), make(chan struct{}, 1) //nolint:gomnd
			signal.Notify(oss, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGABRT)

			go func() {
				// first catch - stop the container
				<-oss
				// send signal to stop execution
				stop <- struct{}{}

				// notify about stopping
				_, _ = sdnotify.SdNotify(sdnotify.Stopping)

				// after first hit we are waiting for the second catch - exit from the process
				<-oss
				fmt.Println("exit forced")
				os.Exit(1)
			}()

			if !*silent {
				fmt.Printf("[INFO] RoadRunner server started; version: %s, buildtime: %s\n", meta.Version(), meta.BuildTime())
			}

			// at this moment, we're almost sure that the container is running (almost- because we don't know if the plugins won't report an error on the next step)
			notified, err := sdnotify.SdNotify(sdnotify.Ready)
			if err != nil {
				if !*silent {
					fmt.Printf("[WARN] sdnotify: %s\n", err)
				}
			}

			if !*silent {
				if notified {
					fmt.Println("[INFO] sdnotify: notified")
					stopCh := make(chan struct{}, 1)
					if containerCfg.WatchdogSec > 0 {
						fmt.Printf("[INFO] sdnotify: watchdog enabled, timeout: %d seconds\n", containerCfg.WatchdogSec)
						sdnotify.StartWatchdog(containerCfg.WatchdogSec, stopCh)
					}

					// if notified -> notify about stop
					defer func() {
						stopCh <- struct{}{}
					}()
				} else {
					fmt.Println("[INFO] sdnotify: not notified")
				}
			}

			for {
				select {
				case e := <-errCh:
					return fmt.Errorf("error: %w\nplugin: %s", e.Error, e.VertexID)
				case <-stop: // stop the container after first signal
					fmt.Printf("stop signal received, grace timeout is: %0.f seconds\n", containerCfg.GracePeriod.Seconds())

					if err = cont.Stop(); err != nil {
						return fmt.Errorf("error: %w", err)
					}

					return nil
				}
			}
		},
	}
}
