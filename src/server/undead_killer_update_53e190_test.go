package server

import (
	"testing"
	"unsafe"
)

func TestUndeadKillerUpdate53E190CancelledOrExpired(t *testing.T) {
	for _, tc := range []struct {
		name    string
		frame   uint32
		created uint32
		flags   uint32
		data    bool
		deleted bool
	}{
		{"live", 170, 100, 0, true, false},
		{"expired", 171, 100, 0, true, true},
		{"cancelled", 100, 100, 1, true, true},
		{"other flag", 100, 100, 2, true, false},
		{"missing data live", 170, 100, 0, false, false},
		{"missing data expired", 171, 100, 0, false, true},
		{"frame wrap live", 20, ^uint32(0) - 49, 0, true, false},
		{"frame wrap expired", 21, ^uint32(0) - 49, 0, true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			obj := &Object{Field34: tc.created}
			if tc.data {
				obj.CollideData = unsafe.Pointer(&UndeadKillerCollideData{Spell: &DurSpell{Flags88: tc.flags}})
			}
			var deleted int
			UndeadKillerUpdate53E190(obj, tc.frame, func(got *Object) {
				if got != obj {
					t.Errorf("deleted %p, want %p", got, obj)
				}
				deleted++
			})
			if (deleted == 1) != tc.deleted {
				t.Fatalf("deleted %d times, want %t", deleted, tc.deleted)
			}
		})
	}
}
