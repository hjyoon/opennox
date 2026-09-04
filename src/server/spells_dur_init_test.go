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
