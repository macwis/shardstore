package store

import (
	"reflect"
	"sync"
	"testing"
)

func TestReportIDs_dedupWithinShard(t *testing.T) {
	s := New()
	newItems, removals := s.ReportIDs([]string{"a", "a", "b", "b"}, 0)
	if !reflect.DeepEqual(newItems, []string{"a", "b"}) {
		t.Errorf("newItems = %v", newItems)
	}
	if len(removals) != 0 {
		t.Errorf("removals = %v", removals)
	}
}

func TestReportIDs_newAndRemovals(t *testing.T) {
	s := New()
	s.ReportIDs([]string{"a", "b", "c"}, 0)
	newItems, removals := s.ReportIDs([]string{"b", "c", "d"}, 0)
	if !reflect.DeepEqual(newItems, []string{"d"}) {
		t.Errorf("newItems = %v", newItems)
	}
	if !reflect.DeepEqual(removals, []string{"a"}) {
		t.Errorf("removals = %v", removals)
	}
}

func TestReportIDs_emptyShardReplacement(t *testing.T) {
	s := New()
	s.ReportIDs([]string{"x"}, 0)
	newItems, removals := s.ReportIDs([]string{}, 0)
	if len(newItems) != 0 {
		t.Errorf("newItems = %v", newItems)
	}
	if !reflect.DeepEqual(removals, []string{"x"}) {
		t.Errorf("removals = %v", removals)
	}
}

func TestAll_deterministicOrder(t *testing.T) {
	s := New()
	s.ReportIDs([]string{"z", "a", "m"}, 0)
	s.ReportIDs([]string{"b", "a"}, 1)
	all := s.All()
	expected := []string{"a", "b", "m", "z"}
	if !reflect.DeepEqual(all, expected) {
		t.Errorf("All() = %v, want %v", all, expected)
	}
}

func TestAll_emptyStore(t *testing.T) {
	s := New()
	all := s.All()
	if len(all) != 0 {
		t.Errorf("All() = %v, want empty", all)
	}
}

func TestAll_sameIDMultipleShards(t *testing.T) {
	s := New()
	s.ReportIDs([]string{"x"}, 0)
	s.ReportIDs([]string{"x"}, 1)
	all := s.All()
	if !reflect.DeepEqual(all, []string{"x"}) {
		t.Errorf("All() = %v", all)
	}
}

func TestDiff_changedShard(t *testing.T) {
	s := New()
	s.ReportIDs([]string{"a", "b"}, 0)
	s.ReportIDs([]string{"c"}, 1)
	info, changed := s.Diff(map[string]int{"a": 0, "b": 1, "c": 1})
	if len(info) != 3 {
		t.Fatalf("len(info) = %d", len(info))
	}
	if info["b"].OldShard != 0 || info["b"].NewShard != 1 || !info["b"].Changed {
		t.Errorf("info[b] = %+v", info["b"])
	}
	if !reflect.DeepEqual(changed, []string{"b"}) {
		t.Errorf("changed = %v", changed)
	}
}

func TestDiff_newIDInMapping(t *testing.T) {
	s := New()
	s.ReportIDs([]string{"a"}, 0)
	info, changed := s.Diff(map[string]int{"a": 0, "b": 1})
	if info["b"].OldShard != 0 || info["b"].NewShard != 1 || !info["b"].Changed {
		t.Errorf("new ID in mapping: info[b] = %+v", info["b"])
	}
	if len(changed) != 0 {
		t.Errorf("new ID not in store should not be in changed: %v", changed)
	}
}

func TestDiff_removedFromMapping(t *testing.T) {
	s := New()
	s.ReportIDs([]string{"a"}, 0)
	info, _ := s.Diff(map[string]int{})
	if meta, ok := info["a"]; !ok || meta.OldShard != 0 || meta.Changed != true {
		t.Errorf("removed ID: info[a] = %+v, ok = %v", info["a"], ok)
	}
}

func TestDuplicates_none(t *testing.T) {
	s := New()
	s.ReportIDs([]string{"a"}, 0)
	s.ReportIDs([]string{"b"}, 1)
	dup := s.Duplicates()
	if len(dup) != 0 {
		t.Errorf("Duplicates() = %v", dup)
	}
}

func TestDuplicates_present(t *testing.T) {
	s := New()
	s.ReportIDs([]string{"a", "dup"}, 0)
	s.ReportIDs([]string{"dup", "b"}, 1)
	s.ReportIDs([]string{"dup"}, 2)
	dup := s.Duplicates()
	want := map[string][]int{"dup": {0, 1, 2}}
	if !reflect.DeepEqual(dup, want) {
		t.Errorf("Duplicates() = %v, want %v", dup, want)
	}
}

func TestConcurrent(t *testing.T) {
	s := New()
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		shard := i
		go func() {
			defer wg.Done()
			s.ReportIDs([]string{"a", "b"}, shard)
			_ = s.All()
		}()
	}
	wg.Wait()
	if len(s.All()) != 2 {
		t.Error("concurrent: expected 2 unique IDs")
	}
}
