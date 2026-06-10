package jobs

import (
	"context"
	"os"

	"connectrpc.com/connect"
	jobsV2 "github.com/roadrunner-server/api-go/v6/jobs/v2"
	"github.com/roadrunner-server/api-go/v6/jobs/v2/jobsV2connect"
	"google.golang.org/protobuf/types/known/emptypb"
)

func pause(ctx context.Context, client jobsV2connect.JobsServiceClient, pause []string, silent *bool) error {
	_, err := client.Pause(ctx, connect.NewRequest(&jobsV2.Pipelines{Pipelines: pause}))
	if err != nil {
		return err
	}

	if !*silent {
		_ = renderPipelines(os.Stdout, pause).Render()
	}

	return nil
}

func resume(ctx context.Context, client jobsV2connect.JobsServiceClient, resume []string, silent *bool) error {
	_, err := client.Resume(ctx, connect.NewRequest(&jobsV2.Pipelines{Pipelines: resume}))
	if err != nil {
		return err
	}

	if !*silent {
		_ = renderPipelines(os.Stdout, resume).Render()
	}

	return nil
}

func destroy(ctx context.Context, client jobsV2connect.JobsServiceClient, destroy []string, silent *bool) error {
	resp, err := client.Destroy(ctx, connect.NewRequest(&jobsV2.Pipelines{Pipelines: destroy}))
	if err != nil {
		return err
	}

	if !*silent {
		_ = renderPipelines(os.Stdout, resp.Msg.GetPipelines()).Render()
	}

	return nil
}

func list(ctx context.Context, client jobsV2connect.JobsServiceClient) error {
	resp, err := client.List(ctx, connect.NewRequest(&emptypb.Empty{}))
	if err != nil {
		return err
	}

	_ = renderPipelines(os.Stdout, resp.Msg.GetPipelines()).Render()

	return nil
}
