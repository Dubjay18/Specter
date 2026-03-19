package store

import (
	"sort"

	"github.com/Dubjay/specter/internal/types"
	"github.com/dgraph-io/badger/v4"
)

type store struct {
	db *badger.DB
}

func NewStore(path string) (Store, error) {
	opts := badger.DefaultOptions(path)
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}
	return &store{db: db}, nil
}

func (s *store) Save(event types.DivergenceEvent) error {
	return s.db.Update(func(txn *badger.Txn) error {
		key := []byte(event.ID)
		value, err := event.Marshal()
		if err != nil {
			return err
		}
		return txn.Set(key, value)
	})
}

func (s *store) List(limit int) ([]types.DivergenceEvent, error) {
	var events []types.DivergenceEvent
	err := s.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = true
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			err := item.Value(func(val []byte) error {
				var event types.DivergenceEvent
				if err := event.Unmarshal(val); err != nil {
					return err
				}
				events = append(events, event)
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(events, func(i, j int) bool {
		return events[i].Timestamp.After(events[j].Timestamp)
	})

	if limit > 0 && len(events) > limit {
		events = events[:limit]
	}

	return events, err
}

func (s *store) Close() error {
	return s.db.Close()
}
