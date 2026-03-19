package divergence

import (
	"log"
	"math/rand"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/Dubjay/specter/internal/store"
	"github.com/Dubjay/specter/internal/types"
	"github.com/google/uuid"
)

type DivergenceEngine struct {
	store                store.Store
	samplingRate         float64
	divergenceOnly       bool
	startedAt            time.Time
	totalRequests        atomic.Uint64
	divergences          atomic.Uint64
	matches              atomic.Uint64
	statusMismatches     atomic.Uint64
	bodyMismatches       atomic.Uint64
	latencyFaster        atomic.Uint64
	latencySimilar       atomic.Uint64
	latency2xSlower      atomic.Uint64
	latency5xSlower      atomic.Uint64
	latency10xSlower     atomic.Uint64
	latency10xPlusSlower atomic.Uint64
	lastEventAtUnixNano  atomic.Int64
}

type StatsSnapshot struct {
	StartedAt        time.Time         `json:"started_at"`
	LastEventAt      *time.Time        `json:"last_event_at,omitempty"`
	TotalRequests    uint64            `json:"total_requests"`
	Divergences      uint64            `json:"divergences"`
	Matches          uint64            `json:"matches"`
	StatusMismatches uint64            `json:"status_mismatches"`
	BodyMismatches   uint64            `json:"body_mismatches"`
	LatencyBuckets   map[string]uint64 `json:"latency_buckets"`
	DivergenceRate   float64           `json:"divergence_rate"`
}

func NewEngine(s store.Store) *DivergenceEngine {
	return &DivergenceEngine{store: s, samplingRate: 1.0, startedAt: time.Now()}
}

func (de *DivergenceEngine) Analyze(req *http.Request, live, shadow *types.CapturedResponse) types.DivergenceEvent {
	de.totalRequests.Add(1)
	de.lastEventAtUnixNano.Store(time.Now().UnixNano())

	db, err := DiffBodies(live.Body, shadow.Body)
	if err != nil {
		log.Printf("specter: failed to diff bodies: %v", err)
	}
	ds := DiffStatus(live.StatusCode, shadow.StatusCode)

	dl := DiffLatency(live.Latency.Milliseconds(), shadow.Latency.Milliseconds())
	de.recordLatencyBucket(dl.BucketLabel())

	if ds != nil {
		de.statusMismatches.Add(1)
	}
	if len(db) > 0 {
		de.bodyMismatches.Add(1)
	}

	// check if they are identical
	if len(db) == 0 && ds == nil {
		de.matches.Add(1)
		result := types.DivergenceEvent{
			ID:          uuid.New().String(),
			Timestamp:   time.Now(),
			RequestPath: req.URL.Path,
			Method:      req.Method,
			BodyDiff:    nil,
			StatusDiff:  nil,
			LatencyDiff: dl,
			Diverged:    false,
		}
		if !de.divergenceOnly {
			de.save(result)
		}
		return result
	}
	de.divergences.Add(1)
	result := types.DivergenceEvent{
		ID:          uuid.New().String(),
		Timestamp:   time.Now(),
		RequestPath: req.URL.Path,
		Method:      req.Method,
		BodyDiff:    db,
		StatusDiff:  ds,
		LatencyDiff: dl,
		Diverged:    true,
	}
	if rand.Float64() > de.samplingRate {
		return result
	}
	de.save(result)
	return result
}

func (de *DivergenceEngine) StatsSnapshot() StatsSnapshot {
	total := de.totalRequests.Load()
	divergences := de.divergences.Load()
	rate := 0.0
	if total > 0 {
		rate = float64(divergences) / float64(total)
	}

	lastEventUnixNano := de.lastEventAtUnixNano.Load()
	var lastEventAt *time.Time
	if lastEventUnixNano > 0 {
		t := time.Unix(0, lastEventUnixNano)
		lastEventAt = &t
	}

	return StatsSnapshot{
		StartedAt:        de.startedAt,
		LastEventAt:      lastEventAt,
		TotalRequests:    total,
		Divergences:      divergences,
		Matches:          de.matches.Load(),
		StatusMismatches: de.statusMismatches.Load(),
		BodyMismatches:   de.bodyMismatches.Load(),
		LatencyBuckets: map[string]uint64{
			"faster":      de.latencyFaster.Load(),
			"similar":     de.latencySimilar.Load(),
			"2x slower":   de.latency2xSlower.Load(),
			"5x slower":   de.latency5xSlower.Load(),
			"10x slower":  de.latency10xSlower.Load(),
			"10x+ slower": de.latency10xPlusSlower.Load(),
		},
		DivergenceRate: rate,
	}
}

func (de *DivergenceEngine) recordLatencyBucket(bucket string) {
	switch bucket {
	case "faster":
		de.latencyFaster.Add(1)
	case "similar":
		de.latencySimilar.Add(1)
	case "2x slower":
		de.latency2xSlower.Add(1)
	case "5x slower":
		de.latency5xSlower.Add(1)
	case "10x slower":
		de.latency10xSlower.Add(1)
	default:
		de.latency10xPlusSlower.Add(1)
	}
}

func (de *DivergenceEngine) save(event types.DivergenceEvent) {
	if de.store == nil {
		return
	}
	if err := de.store.Save(event); err != nil {
		log.Printf("specter: failed to persist divergence event %s: %v", event.ID, err)
	}
}
