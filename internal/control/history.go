package control

import (
	"context"
	"sync"
	"time"

	"github.com/ChronoCoders/sentra/internal/models"
	"github.com/rs/zerolog/log"
)

type HistoryStore interface {
	InsertHistory(ctx context.Context, serverID string, status *models.Status) error
	InsertPeerHistory(ctx context.Context, serverID, publicKey string, rxBytes, txBytes int64) error
}

type HistoryRecorder struct {
	bus         *EventBus
	store       HistoryStore
	mu          sync.Mutex
	lastPeerRec map[string]time.Time
}

func NewHistoryRecorder(bus *EventBus, store HistoryStore) *HistoryRecorder {
	return &HistoryRecorder{
		bus:         bus,
		store:       store,
		lastPeerRec: make(map[string]time.Time),
	}
}

func (h *HistoryRecorder) Run(ctx context.Context) {
	ch := h.bus.Subscribe()
	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-ch:
			if !ok {
				return
			}
			if event.Status != nil {
				if err := h.store.InsertHistory(ctx, event.ServerID, event.Status); err != nil {
					log.Error().Err(err).Str("server", event.ServerID).Msg("failed to record history")
				}
				h.recordPeerHistory(ctx, event)
			}
		}
	}
}

func (h *HistoryRecorder) recordPeerHistory(ctx context.Context, event models.StatusEvent) {
	now := time.Now()
	for _, peer := range event.Status.Peers {
		key := event.ServerID + "|" + peer.PublicKey
		h.mu.Lock()
		last, ok := h.lastPeerRec[key]
		if ok && now.Sub(last) < 55*time.Second {
			h.mu.Unlock()
			continue
		}
		h.lastPeerRec[key] = now
		h.mu.Unlock()
		if err := h.store.InsertPeerHistory(ctx, event.ServerID, peer.PublicKey, peer.ReceiveBytes, peer.TransmitBytes); err != nil {
			log.Error().Err(err).Str("server", event.ServerID).Str("peer", peer.PublicKey[:8]).Msg("failed to record peer history")
		}
	}
}
