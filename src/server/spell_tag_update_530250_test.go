package server

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func TestSpellTagUpdate530250(t *testing.T) {
	if unsafe.Sizeof(uintptr(0)) == 8 {
		if got := unsafe.Offsetof(DurSpell{}.Pos); got != 48 {
			t.Fatalf("Pos offset = %d, want 48", got)
		}
		if got := unsafe.Offsetof(DurSpell{}.Target48); got != 72 {
			t.Fatalf("Target48 offset = %d, want 72", got)
		}
	} else {
		if got := unsafe.Offsetof(DurSpell{}.Pos); got != 28 {
			t.Fatalf("Pos offset = %d, want 28", got)
		}
		if got := unsafe.Offsetof(DurSpell{}.Target48); got != 48 {
			t.Fatalf("Target48 offset = %d, want 48", got)
		}
	}
	if got := SpellTagUpdate530250(&DurSpell{}); got != 1 {
		t.Fatalf("nil target: got %d, want 1", got)
	}
	for _, tc := range []struct {
		name  string
		flags object.Flags
		want  int32
	}{
		{"clear", 0, 0},
		{"bit 4", 0x10, 0},
		{"bit 5", 0x20, 1},
		{"bit 6", 0x40, 0},
		{"upper bits", 0xffff0000, 0},
		{"upper and bit 5", 0xffff0020, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			target := &Object{ObjFlags: tc.flags}
			record := &DurSpell{Target48: target}
			record.Pos.X = 1.725
			if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(target)) <= 0xffffffff {
				t.Fatal("expected a target pointer above 4 GiB")
			}
			if got := SpellTagUpdate530250(record); got != tc.want {
				t.Fatalf("got %d, want %d", got, tc.want)
			}
			if record.Target48 != target {
				t.Fatal("target identity changed")
			}
		})
	}
}
