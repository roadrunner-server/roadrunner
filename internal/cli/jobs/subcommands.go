package jobs

import (
	"net/rpc"
	"os"

	jobsv1 "go.buf.build/protocolbuffers/go/roadrunner-server/api/proto/jobs/v1"
)

func pause(client *rpc.Client, pause []string, silent *bool) error {
	pipes := &jobsv1.Pipelines{Pipelines: pause}
	er := &jobsv1.Empty{}

	err := client.Call(pauseRPC, pipes, er)
	if err != nil {
		return err
	}

	if !*silent {
		renderPipelines(os.Stdout, pause).Render()
	}

	return nil
}

func resume(client *rpc.Client, resume []string, silent *bool) error {
	pipes := &jobsv1.Pipelines{Pipelines: resume}
	er := &jobsv1.Empty{}

	err := client.Call(resumeRPC, pipes, er)
	if err != nil {
		return err
	}

	if !*silent {
		renderPipelines(os.Stdout, resume).Render()
	}

	return nil
}

func destroy(client *rpc.Client, destroy []string, silent *bool) error {
	pipes := &jobsv1.Pipelines{Pipelines: destroy}
	resp := &jobsv1.Pipelines{}

	err := client.Call(destroyRPC, pipes, resp)
	if err != nil {
		return err
	}

	if !*silent {
		renderPipelines(os.Stdout, resp.GetPipelines()).Render()
	}

	return nil
}

func list(client *rpc.Client) error {
	resp := &jobsv1.Pipelines{}
	er := &jobsv1.Empty{}

	err := client.Call(listRPC, er, resp)
	if err != nil {
		return err
	}

	renderPipelines(os.Stdout, resp.GetPipelines()).Render()

	return nil
}
