package main

import (
	"log"
	"time"

	"swarm-cron/provider"

	"github.com/docker/docker/client"
	"github.com/robfig/cron/v3"
)

func main() {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatalf("Ошибка при создании Docker клиента: %v", err)
	}

	scheduler := NewScheduler(
		NewParser(cron.NewParser(cron.Minute|cron.Hour|cron.Dom|cron.Month|cron.Dow)),
		provider.NewSwarm(cli),
	)

	scheduler.Start()

	time.Sleep(1 * time.Hour)
}
