package store

import (
	"sort"
	"sync"
)

type MetaInfo struct {
	OldShard int
	NewShard int
	Changed  bool
}

type Store struct {
	mu     sync.RWMutex
	shards map[int]map[string]struct{}
}

func New() *Store {
	return &Store{shards: make(map[int]map[string]struct{})}
}

func (s *Store) ReportIDs(ids []string, shard int) (newItems, removals []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	cur := s.shards[shard]
	if cur == nil {
		cur = make(map[string]struct{})
		s.shards[shard] = cur
	}
	next := make(map[string]struct{})
	for _, id := range ids {
		next[id] = struct{}{}
	}
	for id := range next {
		if _, ok := cur[id]; !ok {
			newItems = append(newItems, id)
		}
	}
	for id := range cur {
		if _, ok := next[id]; !ok {
			removals = append(removals, id)
		}
	}
	sort.Strings(newItems)
	sort.Strings(removals)
	s.shards[shard] = next
	return newItems, removals
}

func (s *Store) All() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	seen := make(map[string]struct{})
	for _, m := range s.shards {
		for id := range m {
			seen[id] = struct{}{}
		}
	}
	out := make([]string, 0, len(seen))
	for id := range seen {
		out = append(out, id)
	}
	sort.Strings(out)
	return out
}

func (s *Store) Diff(newMapping map[string]int) (map[string]MetaInfo, []string) {
	s.mu.RLock()
	oldMapping := make(map[string]int)
	for shard, m := range s.shards {
		for id := range m {
			oldMapping[id] = shard
		}
	}
	s.mu.RUnlock()
	info := make(map[string]MetaInfo)
	var changed []string
	allIDs := make(map[string]struct{})
	for id := range oldMapping {
		allIDs[id] = struct{}{}
	}
	for id := range newMapping {
		allIDs[id] = struct{}{}
	}
	for id := range allIDs {
		oldShard, hasOld := oldMapping[id]
		newShard, hasNew := newMapping[id]
		meta := MetaInfo{NewShard: newShard}
		if hasOld {
			meta.OldShard = oldShard
		}
		meta.Changed = !hasOld || !hasNew || oldShard != newShard
		info[id] = meta
		if meta.Changed && hasOld && hasNew && oldShard != newShard {
			changed = append(changed, id)
		}
	}
	sort.Strings(changed)
	return info, changed
}

func (s *Store) Duplicates() map[string][]int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	idToShards := make(map[string][]int)
	for shard, m := range s.shards {
		for id := range m {
			idToShards[id] = append(idToShards[id], shard)
		}
	}
	out := make(map[string][]int)
	for id, shards := range idToShards {
		if len(shards) > 1 {
			sort.Ints(shards)
			out[id] = shards
		}
	}
	return out
}
