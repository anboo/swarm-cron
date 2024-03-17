package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"sort"
	"sync"
	"time"

	"swarm-cron/provider"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
)

type Parser struct {
	cache SyncMap[string, cron.Schedule]
	wrap  cron.Parser
}

// cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
func NewParser(wrap cron.Parser) *Parser {
	return &Parser{
		cache: SyncMap[string, cron.Schedule]{},
		wrap:  wrap,
	}
}

func (p *Parser) Parse(schedule string) (cron.Schedule, error) {
	parsed, exists := p.cache.Load(schedule)
	if exists {
		return parsed, nil
	}

	parsed, err := p.wrap.Parse(schedule)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", schedule, err)
	}
	p.cache.Store(schedule, parsed)

	return parsed, nil
}

const (
	watchingInterval = 10 * time.Second
)

type timer struct {
	cronJobs []provider.CronJob
	ctx      context.Context
	cancel   func()
}

type Scheduler struct {
	id       string
	parser   *Parser
	provider provider.Provider
	wg       sync.WaitGroup
	mu       sync.RWMutex
	state    string
	timers   map[cron.Schedule]timer
}

func NewScheduler(parser *Parser, p provider.Provider) *Scheduler {
	return &Scheduler{
		id:       uuid.New().String(),
		provider: p,
		parser:   parser,
		wg:       sync.WaitGroup{},
		mu:       sync.RWMutex{},
		timers:   make(map[cron.Schedule]timer),
	}
}

func (s *Scheduler) Start() {
	go func() {
		for {
			slog.Debug("watching")
			s.watching()
			<-time.After(watchingInterval)
		}
	}()

	s.startScheduler()
}

func (s *Scheduler) Stop() {
	s.wg.Wait()
}

func (s *Scheduler) startScheduler() {
	go func() {

	}()
}

func (s *Scheduler) watching() {
	jobs, err := s.provider.Provide(context.Background())
	if err != nil {
		slog.Error("provider cron jobs error:", "err", err.Error())
		return
	}

	state := s.calculateChecksum(jobs)
	if s.state == state {
		slog.Debug("state checked", state)
		return
	}
	slog.Info("state changed", "from", s.state, "to", state)

	s.restartTimersForState(state, jobs)
}

func (s *Scheduler) restartTimersForState(state string, jobs []provider.CronJob) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.state = state

	for schedule, t := range s.timers {
		slog.Debug("cancel timer", schedule)
		t.cancel()
	}

	for _, job := range jobs {
		cronSchedule, err := s.parser.Parse(job.Schedule)
		if err != nil {
			slog.Error("cannot parse schedule", cronSchedule, err.Error())
		}

		t, exists := s.timers[cronSchedule]
		if !exists {
			ctx, f := context.WithCancel(context.Background())

			t = timer{
				cronJobs: make([]provider.CronJob, 0),
				ctx:      ctx,
				cancel:   f,
			}

			slog.Info("create timer", "id", job.ID, "schedule", job.Schedule)
		}
		t.cronJobs = append(t.cronJobs, job)

		s.timers[cronSchedule] = t
	}

	for cronSchedule, t := range s.timers {
		cronSchedule := cronSchedule
		t := t

		go func() {
			for {
				now := time.Now()
				next := cronSchedule.Next(now)

				select {
				case <-time.After(next.Sub(now)):
					s.wg.Add(len(t.cronJobs))
					for _, job := range t.cronJobs {
						err := s.provider.Run(context.Background(), job)
						if err != nil {
							slog.Error("run cron job for", job.ID, "error", err.Error())
						}
						s.wg.Done()
					}
				case <-t.ctx.Done():
				}
			}
		}()
	}
}

func (s *Scheduler) calculateChecksum(jobs []provider.CronJob) string {
	sort.SliceStable(jobs, func(i, j int) bool {
		return jobs[i].Schedule > jobs[j].Schedule
	})

	sort.SliceStable(jobs, func(i, j int) bool {
		return jobs[i].ID > jobs[j].ID
	})

	data, err := json.Marshal(jobs)
	if err != nil {
		log.Fatalf("cannot serialize jobs state: %v", err)
	}
	hash := sha256.Sum256(data)

	return hex.EncodeToString(hash[:])
}
