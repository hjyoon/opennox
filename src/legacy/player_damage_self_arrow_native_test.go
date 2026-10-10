package legacy

import (
	"fmt"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestPlayerDamageSelfArrowCRecords4E17B0(t *testing.T) {
	srv := npcReflectLegacyServer4E17B0(t)
	// Package legacy alone does not install the upper-level hurt-state service.
	// These 3-point hits must not call it; HP and damage callbacks remain real.
	oldSetState := Nox_xxx_playerSetState_4FA020
	Nox_xxx_playerSetState_4FA020 = func(*server.Object, server.PlayerState) bool {
		t.Fatal("below-threshold self-arrow requested hurt state")
		return false
	}
	t.Cleanup(func() { Nox_xxx_playerSetState_4FA020 = oldSetState })
	if !srv.Objs.Init(32) {
		t.Fatal("native object allocator")
	}
	t.Cleanup(srv.Objs.FreeObjects)
	for _, owner := range []string{"self", "player", "NPC"} {
		for _, armor := range []float32{0, 0.25} {
			t.Run(fmt.Sprintf("owner-%s/armor-%g", owner, armor), func(t *testing.T) {
				v, _, w := spellMissileImpactCRecords4E17B0(t, srv, owner)
				v.TypeInd, v.ObjClass, v.ObjFlags = 713, object.Class(2621444), object.Flags(16777732)
				w.TypeInd, w.ObjClass, w.ObjSubClass, w.ObjFlags = 529, object.Class(85983233), 16, object.Flags(16794116)
				init, freeInit := alloc.New(server.ModifierInitData{})
				t.Cleanup(freeInit)
				w.InitData = unsafe.Pointer(init)
				if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(w.InitData) <= math.MaxUint32 {
					t.Fatal("C-owned modifier init is not above 4 GiB")
				}
				v.Damage = playerDamageMeleeCallbackNative4E17B0()
				ud := v.UpdateDataPlayer()
				ud.Field57 = math.Float32bits(armor)
				before := *w
				for hit := range 3 {
					wantHP, wantCarry := uint16(60-3*(hit+1)), float32(0)
					if armor != 0 {
						wantHP, wantCarry = []uint16{58, 56, 53}[hit], []float32{0.25, 0.5, -0.25}[hit]
					}
					if !objectDamageDispatchCallNative(v, w, w, 3, object.DamageImpale) || v.HealthData.Cur != wantHP ||
						ud.Field21 != math.Float32bits(wantCarry) || ud.Field76 != 2 || ud.Field75 != 3 ||
						v.Obj130 != w || v.Field131 != 3 || v.Frame134 != srv.Frame() || v.Pos132 != w.PrevPos ||
						v.Field38 != math.MaxUint32 || ud.State != server.PlayerState13 || *w != before {
						t.Fatalf("actual C self-arrow hit=%d HP=%d/%d max=%d marker=%d/%d carry=%g state=%d attribution=%p type=%d frame=%d flags=%x buffs=%x", hit, v.HealthData.Cur, wantHP, v.HealthData.Max, ud.Field76, ud.Field75, math.Float32frombits(ud.Field21), ud.State, v.Obj130, v.Field131, v.Frame134, uint32(v.ObjFlags), v.Buffs)
					}
				}
				t.Logf("C-owned >4 GiB C->registered PlayerDamage->DefaultDamage->UnitSetHP: player=%p self-source=weapon=%p owner=%s armor=%g HP=60->%d", v, w, owner, armor, v.HealthData.Cur)
			})
		}
	}
}
