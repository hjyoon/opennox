package opennox

import (
	"testing"

	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

func TestShieldDurationDispatchUsesNativeCallbacks(t *testing.T) {
	sp := &spellsDuration{}
	record := &server.DurSpell{}
	if got := sp.callCreate4FEBA0(legacy.Get_nox_xxx_castShield1_52F5A0(), record); got != 1 {
		t.Fatalf("create without target = %d, want 1", got)
	}
	if got := sp.callUpdate4FEEF0(legacy.Get_sub_52F650(), record); got != 1 {
		t.Fatalf("update without target = %d, want 1", got)
	}
	sp.callDestroy4FEDA0(legacy.Get_sub_52F670(), record)

	record.Target48 = &server.Object{ObjClass: 4}
	if got := sp.callUpdate4FEEF0(legacy.Get_sub_52F650(), record); got != 0 {
		t.Fatalf("update with live target = %d, want 0", got)
	}
	record.Target48.ObjFlags = 0x20
	if got := sp.callUpdate4FEEF0(legacy.Get_sub_52F650(), record); got != 1 {
		t.Fatalf("update with removed target = %d, want 1", got)
	}
}
