package provider

import (
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/pkg/errors"
)

type DockerProvider struct {
	client *client.Client
}

func NewDockerProvider(client *client.Client) *DockerProvider {
	return &DockerProvider{client: client}
}

func (p *DockerProvider) Provide(ctx context.Context) ([]CronJob, error) {
	containers, err := p.client.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "docker container list")
	}

	var jobs []CronJob

	for _, c := range containers {
		if schedule, ok := c.Labels[containerLabel]; ok {
			jobs = append(jobs, CronJob{
				ID:       c.ID,
				Schedule: schedule,
			})
		}
	}

	return jobs, nil
}

func (p *DockerProvider) Run(ctx context.Context, job CronJob) error {
	if err := p.client.ContainerRestart(ctx, job.ID, container.StopOptions{}); err != nil {
		return errors.Wrapf(err, "restart container %s", job.ID)
	}

	return nil
}
