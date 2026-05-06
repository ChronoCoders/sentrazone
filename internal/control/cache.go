package control

import (
	"context"
	"sync"
	"time"

	"github.com/ChronoCoders/sentra/internal/models"
)

type StatusBroadcaster interface {
	Broadcast(event models.StatusEvent)
}

type OfflineAlerter interface {
	ServerWentOffline(serverID string)
	ServerCameOnline(serverID string)
}

type StatusCache struct {
	mu          sync.RWMutex
	statuses    map[string]*models.Status
	lastSeen    map[string]time.Time
	bus         *EventBus
	broadcaster StatusBroadcaster
	alerter     OfflineAlerter
}

func NewStatusCache(bus *EventBus, broadcaster StatusBroadcaster, alerter OfflineAlerter) *StatusCache {
	c := &StatusCache{
		bus:         bus,
		broadcaster: broadcaster,
		alerter:     alerter,
		statuses:    make(map[string]*models.Status),
		lastSeen:    make(map[string]time.Time),
	}
	go c.listen()
	go c.watchOffline()
	return c
}

func (c *StatusCache) listen() {
	ch := c.bus.Subscribe()
	for event := range ch {
		c.mu.Lock()
		c.statuses[event.ServerID] = event.Status
		c.lastSeen[event.ServerID] = time.Now()
		c.mu.Unlock()

		if c.alerter != nil {
			c.alerter.ServerCameOnline(event.ServerID)
		}
		if c.broadcaster != nil {
			c.broadcaster.Broadcast(event)
		}
	}
}

func (c *StatusCache) watchOffline() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		now := time.Now()
		c.mu.RLock()
		for sid, t := range c.lastSeen {
			if now.Sub(t) > 60*time.Second && c.alerter != nil {
				go c.alerter.ServerWentOffline(sid)
			}
		}
		c.mu.RUnlock()
	}
}

func (c *StatusCache) GetStatus(ctx context.Context, serverID string) (*models.Status, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.statuses[serverID], nil
}

func (c *StatusCache) GetAllStatuses() []models.StatusEvent {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var events []models.StatusEvent
	for id, status := range c.statuses {
		t := c.lastSeen[id]
		if t.IsZero() {
			t = time.Now()
		}
		events = append(events, models.StatusEvent{
			ServerID: id,
			Status:   status,
			Time:     t,
		})
	}
	return events
}

func (c *StatusCache) ListPeers(ctx context.Context, serverID string) ([]models.Peer, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if status, ok := c.statuses[serverID]; ok {
		return status.Peers, nil
	}
	return nil, nil
}
