package server

import "testing"

func TestSpellDurationFirst4FE930LoadsHeadOnceAndPreservesValue(t *testing.T) {
	const original = uint64(0x100000101)
	const replacement = uint64(0x200000202)
	live := original
	loads := 0

	got := SpellDurationFirst4FE930(SpellDurationFirstHooks4FE930[uint64]{
		LoadHead: func() uint64 {
			loads++
			value := live
			live = replacement
			return value
		},
	})

	if got != original {
		t.Fatalf("result = %#x, want original live head %#x", got, original)
	}
	if loads != 1 {
		t.Fatalf("head loads = %d, want 1", loads)
	}
	if live != replacement {
		t.Fatalf("live head = %#x, want callback replacement %#x", live, replacement)
	}
}

func TestSpellDurationFirst4FE930PreservesNil(t *testing.T) {
	loads := 0
	got := SpellDurationFirst4FE930(SpellDurationFirstHooks4FE930[*int]{
		LoadHead: func() *int {
			loads++
			return nil
		},
	})
	if got != nil {
		t.Fatalf("result = %p, want nil", got)
	}
	if loads != 1 {
		t.Fatalf("head loads = %d, want 1", loads)
	}
}

func TestSpellDurationFirst4FE930PropagatesLoadFault(t *testing.T) {
	stop := &struct{}{}
	loads := 0
	var recovered any
	func() {
		defer func() { recovered = recover() }()
		SpellDurationFirst4FE930(SpellDurationFirstHooks4FE930[uint64]{
			LoadHead: func() uint64 {
				loads++
				panic(stop)
			},
		})
	}()

	if recovered != stop {
		t.Fatalf("recovered = %#v, want sentinel", recovered)
	}
	if loads != 1 {
		t.Fatalf("head loads before fault = %d, want 1", loads)
	}
}
