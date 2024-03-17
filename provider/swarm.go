package provider

import (
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/pkg/errors"
)

type Swarm struct {
	client *client.Client
}

func NewSwarm(client *client.Client) *Swarm {
	return &Swarm{client: client}
}

func (p *Swarm) Run(ctx context.Context, job CronJob) error {
	service, _, err := p.client.ServiceInspectWithRaw(context.Background(), job.ID, types.ServiceInspectOptions{})
	if err != nil {
		return errors.Wrapf(err, "inspect service %s", job.ID)
	}

	_, err = p.client.ServiceUpdate(context.Background(), service.ID, service.Version, service.Spec, types.ServiceUpdateOptions{})
	if err != nil {
		return errors.Wrapf(err, "update service %s", job.ID)
	}

	return nil
}

func (p *Swarm) Provide(ctx context.Context) ([]CronJob, error) {
	services, err := p.client.ServiceList(context.Background(), types.ServiceListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "docker service list")
	}

	var res []CronJob

servicesFor:
	for _, service := range services {
		for key, value := range service.Spec.Labels {
			if key == containerLabel {
				res = append(res, CronJob{
					ID:       service.ID,
					Schedule: value,
				})
				continue servicesFor
			}
		}
	}

	return res, nil
}
