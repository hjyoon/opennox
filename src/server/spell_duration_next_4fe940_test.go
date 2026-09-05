package server

import "testing"

func TestSpellDurationNext4FE940LoadsNextOnceAndPreservesValue(t *testing.T) {
	const record = uint64(0x100000101)
	const original = uint64(0x200000202)
	const replacement = uint64(0x300000303)
	live := original
	loads := 0

	got := SpellDurationNext4FE940(record, SpellDurationNextHooks4FE940[uint64]{
		LoadNext: func(gotRecord uint64) uint64 {
			loads++
			if gotRecord != record {
				t.Fatalf("record = %#x, want %#x", gotRecord, record)
			}
			value := live
			live = replacement
			return value
		},
	})

	if got != original {
		t.Fatalf("result = %#x, want original live next %#x", got, original)
	}
	if loads != 1 {
		t.Fatalf("next loads = %d, want 1", loads)
	}
	if live != replacement {
		t.Fatalf("live next = %#x, want callback replacement %#x", live, replacement)
	}
}

func TestSpellDurationNext4FE940PreservesNil(t *testing.T) {
	record := new(int)
	loads := 0
	got := SpellDurationNext4FE940(record, SpellDurationNextHooks4FE940[*int]{
		LoadNext: func(gotRecord *int) *int {
			loads++
			if gotRecord != record {
				t.Fatalf("record = %p, want %p", gotRecord, record)
			}
			return nil
		},
	})
	if got != nil {
		t.Fatalf("result = %p, want nil", got)
	}
	if loads != 1 {
		t.Fatalf("next loads = %d, want 1", loads)
	}
}

func TestSpellDurationNext4FE940PropagatesLoadFault(t *testing.T) {
	const record = uint64(0x100000101)
	stop := &struct{}{}
	loads := 0
	var recovered any
	func() {
		defer func() { recovered = recover() }()
		SpellDurationNext4FE940(record, SpellDurationNextHooks4FE940[uint64]{
			LoadNext: func(gotRecord uint64) uint64 {
				loads++
				if gotRecord != record {
					t.Fatalf("record = %#x, want %#x", gotRecord, record)
				}
				panic(stop)
			},
		})
	}()

	if recovered != stop {
		t.Fatalf("recovered = %#v, want sentinel", recovered)
	}
	if loads != 1 {
		t.Fatalf("next loads before fault = %d, want 1", loads)
	}
}
