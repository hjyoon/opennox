package opennox

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

func TestSpellsDurationTagUpdateUsesNativeRecord(t *testing.T) {
	target := &server.Object{ObjFlags: object.Flags(0x20)}
	record := &server.DurSpell{Target48: target}
	record.Pos.X = 1.725
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(target)) <= 0xffffffff {
		t.Fatal("expected a target pointer above 4 GiB")
	}
	if got := (&spellsDuration{}).callUpdate4FEEF0(legacy.Get_sub_530250(), record); got != 1 {
		t.Fatalf("tag callback returned %d, want 1", got)
	}
	target.ObjFlags = 0
	if got := (&spellsDuration{}).callUpdate4FEEF0(legacy.Get_sub_530250(), record); got != 0 {
		t.Fatalf("tag callback returned %d, want 0", got)
	}
}

func TestSpellsDurationTagCreateAndDestroyUseNativeRecord(t *testing.T) {
	sp := &spellsDuration{s: &Server{Server: &server.Server{}}}
	record := &server.DurSpell{}
	record.Pos.X = 1.725
	if got := sp.callCreate4FEBA0(legacy.Get_nox_xxx_spellTagCreature_530160(), record); got != 1 {
		t.Fatalf("invalid tag create returned %d, want 1", got)
	}
	sp.callDestroy4FEDA0(legacy.Get_sub_530270(), record)
}
