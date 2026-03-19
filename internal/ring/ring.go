package ring

import (
	"hash/crc32"
	"slices"
	"strconv"
	"sync"
)

type Ring struct {
	virtualNodes int
	positions    []uint32
	nodeMap      map[uint32]string
	nodes        map[string]struct{}
	mu           sync.RWMutex
}

func NewRing(virtualNodes int) *Ring {
	return &Ring{
		virtualNodes: virtualNodes,
		positions:    []uint32{},
		nodeMap:      make(map[uint32]string),
		nodes:        make(map[string]struct{}),
	}
}

func (r *Ring) AddNode(node string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.nodes[node]; exists {
		return
	}

	for i := 0; i < r.virtualNodes; i++ {
		position := hash(node + "#" + strconv.Itoa(i))
		r.positions = append(r.positions, position)
		r.nodeMap[position] = node
	}
	slices.Sort(r.positions)
	r.nodes[node] = struct{}{}
}

func (r *Ring) GetNode(key string) string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.positions) == 0 {
		return ""
	}
	position := hash(key)
	for _, pos := range r.positions {
		if position <= pos {
			return r.nodeMap[pos]
		}
	}
	return r.nodeMap[r.positions[0]]
}

func (r *Ring) RemoveNode(node string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.nodes[node]; !exists {
		return
	}

	filtered := r.positions[:0]
	for _, pos := range r.positions {
		if r.nodeMap[pos] == node {
			delete(r.nodeMap, pos)
			continue
		}
		filtered = append(filtered, pos)
	}
	r.positions = filtered
	delete(r.nodes, node)
}

func (r *Ring) GetOwner(key string) string {
	return r.GetNode(key)
}

func hash(s string) uint32 {
	// convert string to bytes and compute CRC32 hash
	b := []byte(s)
	return crc32.ChecksumIEEE(b)
}
