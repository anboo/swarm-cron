package provider

import (
	"context"
)

const containerLabel = "cronjob_schedule"

type CronJob struct {
	ID       string
	Schedule string
}

type Provider interface {
	Provide(ctx context.Context) ([]CronJob, error)
	Run(ctx context.Context, job CronJob) error
}
