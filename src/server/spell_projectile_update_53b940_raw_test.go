package server

import "testing"

func TestSpellFlightRawRequiredPointers53B940(t *testing.T) {
	for _, tc := range []struct {
		name    string
		missile *Object
	}{{"nil-object", nil}, {"nil-update", new(Object)}} {
		t.Run(tc.name, func(t *testing.T) {
			// Actual original 0053B940 faults at +748 or at update+4.
			// Only the registered Go dispatcher's existing guard returns early.
			SpellProjectileUpdate53B940(tc.missile, SpellProjectileUpdateRuntime53B940{})
			defer func() {
				if recover() == nil {
					t.Fatal("raw entry accepted an invalid original pointer precondition")
				}
			}()
			SpellProjectileUpdateRaw53B940(tc.missile, SpellProjectileUpdateRuntime53B940{})
		})
	}
}
