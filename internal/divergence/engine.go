package divergence

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Dubjay/specter/internal/types"
	"github.com/google/uuid"
)

func Analyze(req *http.Request, live, shadow *types.CapturedResponse) DivergenceEvent {
	db, err := DiffBodies(live.Body,shadow.Body)
	if err != nil {
		fmt.Printf("%v",err)
	}
	ds := DiffStatus(live.StatusCode,shadow.StatusCode)
	
	dl:= DiffLatency(live.Latency.Milliseconds(),shadow.Latency.Milliseconds())

	// check if they are identical
	if len(db) == 0 && ds == nil {
		return DivergenceEvent{
			ID: uuid.New().String(),
			Timestamp: time.Now(),
			RequestPath: req.URL.Path,
			Method: req.Method,
			BodyDiff: nil,
			StatusDiff: nil,
			LatencyDiff: dl,
			Diverged: false,
		}
	}
	return  DivergenceEvent{
		ID: uuid.New().String(),
		Timestamp: time.Now(),
		RequestPath: req.URL.Path,
		Method: req.Method,
		BodyDiff: db,
		StatusDiff: ds,
		LatencyDiff: dl,
		Diverged: true,
	}
}