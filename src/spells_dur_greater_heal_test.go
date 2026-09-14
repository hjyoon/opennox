package opennox

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

func TestSpellsDurationGreaterHealDispatchAvoidsPE32TargetOffset(t *testing.T) {
	sp := &spellsDuration{s: &Server{Server: &server.Server{}}}
	record := &server.DurSpell{}
	record.Pos.X = math.Float32frombits(0x3fdccccc)
	if got := sp.callCreate4FEBA0(legacy.Get_sub_52F220(), record); got != 1 {
		t.Fatalf("nil caster create = %d", got)
	}
	if got := sp.callUpdate4FEEF0(legacy.Get_sub_52F2E0(), record); got != 1 {
		t.Fatalf("nil target update = %d", got)
	}
	target := &server.Object{ObjFlags: object.Flags(0x20)}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(target)) <= 0xffffffff {
		t.Fatal("expected target above 4 GiB")
	}
	record.Target48 = target
	if got := sp.callUpdate4FEEF0(legacy.Get_sub_52F2E0(), record); got != 1 {
		t.Fatalf("destroyed target update = %d", got)
	}
}
