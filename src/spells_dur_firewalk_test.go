package opennox

import (
	"math"
	"testing"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

func TestSpellsDurationFirewalkDispatchAvoidsPE32TargetOffset(t *testing.T) {
	sp := &spellsDuration{s: &Server{Server: &server.Server{}}}
	record := &server.DurSpell{}
	record.Pos.X = math.Float32frombits(0x3fdccccc)
	if got := sp.callUpdate4FEEF0(legacy.Get_nox_xxx_firewalkTick_52ED40(), record); got != 1 {
		t.Fatalf("nil target update = %d", got)
	}
	record.Target48 = &server.Object{ObjFlags: object.Flags(0x20)}
	if got := sp.callUpdate4FEEF0(legacy.Get_nox_xxx_firewalkTick_52ED40(), record); got != 1 {
		t.Fatalf("destroyed target update = %d", got)
	}
}
