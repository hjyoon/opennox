package opennox

import (
	"testing"

	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

func TestSpellsDurationManaBombDispatchUsesNativeRecord(t *testing.T) {
	sp := &spellsDuration{s: &Server{Server: &server.Server{}}}
	record := &server.DurSpell{Field76: 0x3fdccccc}
	// With no caster, the native update cancels before touching world state.
	// The C callback would read a PE32 slot from this wider record instead.
	if got := sp.callUpdate4FEEF0(legacy.Get_nox_xxx_manaBombBoom_5310C0(), record); got != 1 {
		t.Fatalf("orphan update = %d, want 1", got)
	}
	record.Flag20 = 1
	record.Field84 = 1
	runtime := sp.manaBombRuntime530F90()
	sp.callDestroy4FEDA0(legacy.Get_sub_531290(), record)
	if runtime.LoadCharge(record) != nil {
		t.Fatal("destroy did not clear charge sidecar")
	}
	if record.Field76 != 0x3fdccccc {
		t.Fatalf("legacy PE32 slot changed: %#x", record.Field76)
	}
	visual := &server.Object{}
	runtime.StoreCharge(record, visual)
	if runtime.LoadCharge(record) != visual {
		t.Fatal("charge sidecar lost native object")
	}
	sp.Free()
	if runtime.LoadCharge(record) != nil {
		t.Fatal("duration cleanup did not clear charge sidecar")
	}
}
