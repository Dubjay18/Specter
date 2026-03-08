package divergence

import (
	"log"
	"net/http"
	"time"

	"github.com/Dubjay/specter/internal/store"
	"github.com/Dubjay/specter/internal/types"
	"github.com/google/uuid"
)

type DivergenceEngine struct {
	store store.Store
}

func NewEngine(s store.Store) *DivergenceEngine {
	return &DivergenceEngine{store: s}
}

func (de *DivergenceEngine) Analyze(req *http.Request, live, shadow *types.CapturedResponse) types.DivergenceEvent {
	db, err := DiffBodies(live.Body, shadow.Body)
	if err != nil {
		log.Printf("specter: failed to diff bodies: %v", err)
	}
	ds := DiffStatus(live.StatusCode, shadow.StatusCode)

	dl := DiffLatency(live.Latency.Milliseconds(), shadow.Latency.Milliseconds())

	// check if they are identical
	if len(db) == 0 && ds == nil {
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
		de.save(result)
		return result
	}
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
	de.save(result)
	return result
}

func (de *DivergenceEngine) save(event types.DivergenceEvent) {
	if de.store == nil {
		return
	}
	if err := de.store.Save(event); err != nil {
		log.Printf("specter: failed to persist divergence event %s: %v", event.ID, err)
	}
}
