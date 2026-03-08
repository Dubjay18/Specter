package store

import (
	"github.com/Dubjay/specter/internal/types"
)

type Store interface {
    Save(event types.DivergenceEvent) error
    List(limit int) ([]types.DivergenceEvent, error)
    Close() error
  }

