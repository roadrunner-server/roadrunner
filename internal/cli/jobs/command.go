package jobs

import (
	"strings"

	internalRpc "github.com/roadrunner-server/roadrunner/v2/internal/rpc"

	"github.com/roadrunner-server/errors"
	"github.com/spf13/cobra"
)

const (
	listRPC    string = "jobs.List"
	pauseRPC   string = "jobs.Pause"
	destroyRPC string = "jobs.Destroy"
	resumeRPC  string = "jobs.Resume"
)

// NewCommand creates `jobs` command.
func NewCommand(cfgFile *string, override *[]string, silent *bool) *cobra.Command {
	var (
		pausePipes   bool
		destroyPipes bool
		resumePipes  bool
		listPipes    bool
	)

	cmd := &cobra.Command{
		Use:   "jobs",
		Short: "Jobs pipelines manipulation",
		RunE: func(_ *cobra.Command, args []string) error {
			const op = errors.Op("jobs_command")

			if cfgFile == nil {
				return errors.E(op, errors.Str("no configuration file provided"))
			}

			// for the commands other than list, args[1] should contain list of pipelines to pause/resume/destroy
			if !listPipes && len(args[0]) == 0 {
				return errors.Str("pause/resume/destroy commands should have list of the pipelines as second arg")
			}

			client, err := internalRpc.NewClient(*cfgFile, *override)
			if err != nil {
				return err
			}

			defer func() { _ = client.Close() }()

			switch {
			case pausePipes:
				split := strings.Split(strings.Trim(args[0], " "), ",")

				return pause(client, split, silent)
			case destroyPipes:
				split := strings.Split(strings.Trim(args[0], " "), ",")

				return destroy(client, split, silent)
			case resumePipes:
				split := strings.Split(strings.Trim(args[0], " "), ",")

				return resume(client, split, silent)
			case listPipes:
				return list(client)
			default:
				return errors.Str("command should be in form of: `rr jobs pause pipe1,pipe2,etc`")
			}
		},
	}

	// commands
	cmd.Flags().BoolVar(&pausePipes, "pause", false, "pause pipelines")
	cmd.Flags().BoolVar(&destroyPipes, "destroy", false, "destroy pipelines")
	cmd.Flags().BoolVar(&resumePipes, "resume", false, "resume pipelines")
	cmd.Flags().BoolVar(&listPipes, "list", false, "list pipelines")

	return cmd
}
