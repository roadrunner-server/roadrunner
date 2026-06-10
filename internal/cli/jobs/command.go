package jobs

import (
	"strings"

	internalRpc "github.com/roadrunner-server/roadrunner/v2025/internal/rpc"

	"github.com/roadrunner-server/api-go/v6/jobs/v2/jobsV2connect"
	"github.com/roadrunner-server/errors"
	"github.com/spf13/cobra"
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
		RunE: func(cmd *cobra.Command, args []string) error {
			const op = errors.Op("jobs_command")

			if cfgFile == nil {
				return errors.E(op, errors.Str("no configuration file provided"))
			}

			baseURL, httpClient, err := internalRpc.NewClient(*cfgFile, *override)
			if err != nil {
				return err
			}

			client := jobsV2connect.NewJobsServiceClient(httpClient, baseURL)
			ctx := cmd.Context()

			switch {
			case pausePipes:
				if len(args) == 0 {
					return errors.Str("incorrect command usage, should be: rr jobs --pause pipe1,pipe2")
				}
				split := strings.Split(strings.Trim(args[0], " "), ",")

				return pause(ctx, client, split, silent)
			case destroyPipes:
				if len(args) == 0 {
					return errors.Str("incorrect command usage, should be: rr jobs --destroy pipe1,pipe2")
				}
				split := strings.Split(strings.Trim(args[0], " "), ",")

				return destroy(ctx, client, split, silent)
			case resumePipes:
				if len(args) == 0 {
					return errors.Str("incorrect command usage, should be: rr jobs --resume pipe1,pipe2")
				}
				split := strings.Split(strings.Trim(args[0], " "), ",")

				return resume(ctx, client, split, silent)
			case listPipes:
				return list(ctx, client)
			default:
				return errors.Str("command should be in form of: `rr jobs --<destroy/resume/pause> pipe1,pipe2`")
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
