package server

import "testing"

func TestSpellsDurationInitAllocatesNativeRecords(t *testing.T) {
	var spells SpellsDuration
	if !spells.Init() {
		t.Fatal("Init returned false")
	}
	defer spells.Free()

	first := spells.NewRaw()
	second := spells.NewRaw()
	if first == nil || second == nil {
		t.Fatalf("allocated records = (%p, %p), want two non-nil records", first, second)
	}
	if first.ID != 1 || second.ID != 2 {
		t.Fatalf("allocated IDs = (%d, %d), want (1, 2)", first.ID, second.ID)
	}
}

func TestSpellsDurationFreeClearsListAfterAllocator(t *testing.T) {
	var spells SpellsDuration
	if !spells.Init() {
		t.Fatal("Init returned false")
	}
	record := spells.NewRaw()
	if record == nil {
		t.Fatal("NewRaw returned nil")
	}
	spells.List = record
	allocator := spells.alloc.Class
	lastID := spells.lastID

	spells.Free()

	if spells.List != nil {
		t.Fatalf("list = %p, want nil", spells.List)
	}
	if spells.alloc.Class != allocator {
		t.Fatalf("allocator global changed from %p to %p", allocator, spells.alloc.Class)
	}
	if spells.lastID != lastID {
		t.Fatalf("last ID = %d, want preserved %d", spells.lastID, lastID)
	}
}
