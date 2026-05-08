package control

import (
	"context"
	"time"

	"github.com/ChronoCoders/sentra/internal/models"
)

type SchedulerStore interface {
	GetUptimePercent(ctx context.Context, serverID string, hours int) (float64, error)
	GetBandwidthSummary(ctx context.Context, serverID string, hours int) (rx, tx int64, err error)
}

type SchedulerAlerter interface {
	Configured() bool
	SendSummary(label string, hours int, servers []models.ServerSummary)
}

type Scheduler struct {
	store   SchedulerStore
	cache   *StatusCache
	alerter SchedulerAlerter
}

func NewScheduler(store SchedulerStore, cache *StatusCache, alerter SchedulerAlerter) *Scheduler {
	return &Scheduler{store: store, cache: cache, alerter: alerter}
}

func (s *Scheduler) Run(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	var lastDaily, lastWeekly time.Time
	for {
		select {
		case <-ctx.Done():
			return
		case t := <-ticker.C:
			utc := t.UTC()
			day := utc.Truncate(24 * time.Hour)
			if utc.Hour() == 8 && utc.Minute() == 0 && day != lastDaily {
				lastDaily = day
				go s.sendSummary(ctx, "Daily", 24)
			}
			if utc.Weekday() == time.Monday && utc.Hour() == 8 && utc.Minute() == 0 && day != lastWeekly {
				lastWeekly = day
				go s.sendSummary(ctx, "Weekly", 168)
			}
		}
	}
}

func (s *Scheduler) sendSummary(ctx context.Context, label string, hours int) {
	if !s.alerter.Configured() {
		return
	}
	statuses := s.cache.GetAllStatuses()
	var summaries []models.ServerSummary
	for _, ev := range statuses {
		up, _ := s.store.GetUptimePercent(ctx, ev.ServerID, hours)
		rx, tx, _ := s.store.GetBandwidthSummary(ctx, ev.ServerID, hours)
		peers := 0
		if ev.Status != nil {
			peers = len(ev.Status.Peers)
		}
		summaries = append(summaries, models.ServerSummary{
			ServerID:    ev.ServerID,
			UptimePC:    up,
			ActivePeers: peers,
			RxBytes:     rx,
			TxBytes:     tx,
		})
	}
	s.alerter.SendSummary(label, hours, summaries)
}
