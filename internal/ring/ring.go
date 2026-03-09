package ring

import (
	"hash/crc32"
	"slices"
	"strconv"
)


type Ring struct {
    virtualNodes int
    positions    []uint32
    nodeMap      map[uint32]string
}

func NewRing(virtualNodes int) *Ring {
	return &Ring{
		virtualNodes: virtualNodes,
		positions:    []uint32{},
		nodeMap:      make(map[uint32]string),
	}
}

func (r *Ring) AddNode(node string) {
	for i := 0; i < r.virtualNodes; i++ {
		position := hash(node + "#" + strconv.Itoa(i))
		r.positions = append(r.positions, position)
		r.nodeMap[position] = node
	}
	slices.Sort(r.positions)
}

func (r *Ring) GetNode(key string) string {
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
	for i := 0; i < r.virtualNodes; i++ {
		position := hash(node + "#" + strconv.Itoa(i))
		delete(r.nodeMap, position)
		for j, pos := range r.positions {
			if pos == position {
				r.positions = append(r.positions[:j], r.positions[j+1:]...)
				break
			}
		}
	}
}

func (r *Ring) GetOwner(key string) string {
	return r.GetNode(key)
}

func hash(s string) uint32 {
	// convert string to bytes and compute CRC32 hash
	b := []byte(s)
	return crc32.ChecksumIEEE(b)
}