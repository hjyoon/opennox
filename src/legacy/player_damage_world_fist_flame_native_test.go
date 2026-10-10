package legacy

import (
	"fmt"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/things"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

// Real C dispatcher, registered PlayerDamage/DefaultDamage and UnitSetHP.
// Only scalar absorption/carry and the upper game module's hurt-state boundary
// are declared; no damage callback or HP function is replaced.
func TestPlayerDamageWorldFistFlameCRecords4E17B0(t *testing.T) {
	srv := npcReflectLegacyServer4E17B0(t)
	srv.FreeObjectTypes()
	t.Cleanup(srv.FreeObjectTypes)
	if err := srv.Types.ReadObjectType(&things.Thing{Name: "NativeWorldFistFlameDefault", OnDamage: &things.ProcFunc{Name: "DefaultDamage"}}); err != nil {
		t.Fatal(err)
	}
	if !srv.Objs.Init(64) {
		t.Fatal("native object allocator")
	}
	t.Cleanup(srv.Objs.FreeObjects)
	oldState := Nox_xxx_playerSetState_4FA020
	stateCalls := 0
	Nox_xxx_playerSetState_4FA020 = func(v *server.Object, state server.PlayerState) bool {
		if state != server.PlayerState30 || !v.Class().Has(object.ClassPlayer) {
			t.Fatal("world damage hurt-state boundary")
		}
		stateCalls++
		v.UpdateDataPlayer().State = state
		return true
	}
	t.Cleanup(func() { Nox_xxx_playerSetState_4FA020 = oldState })
	for _, flame := range []bool{false, true} {
		for _, target := range []string{"player", "NPC", "monster", "immune-NPC"} {
			if !flame && target == "immune-NPC" {
				continue
			}
			for _, absorption := range []float32{0, 0.5} {
				t.Run(fmt.Sprintf("flame-%t/%s/armor-%g", flame, target, absorption), func(t *testing.T) {
					v, a, w := spellMissileImpactCRecords4E17B0(t, srv, "NPC")
					v.TypeInd, v.ObjClass, v.ObjFlags = 713, object.Class(2621444), object.Flags(16777732)
					v.HealthData.Cur, v.HealthData.Max, v.HealthData.Field2 = 2000, 2000, 2000
					v.Damage = playerDamageMeleeCallbackNative4E17B0()
					a.TypeInd, a.ObjClass, a.ObjSubClass, a.ObjFlags, a.UpdateData = 734, object.Class(4194816), 0, object.Flags(525), nil
					w.TypeInd, w.ObjClass, w.ObjSubClass, w.ObjFlags, w.ObjOwner = 1217, object.Class(1048576), 0, object.Flags(25166348), nil
					var carry, marker, kind *uint32
					if target == "player" {
						u := v.UpdateDataPlayer()
						u.Field57, u.Field21 = math.Float32bits(absorption), math.Float32bits(0.25)
						carry, marker, kind = &u.Field21, &u.Field76, &u.Field75
					} else {
						u, freeU := alloc.New(server.MonsterUpdateData{})
						t.Cleanup(freeU)
						v.ObjClass, v.ObjSubClass, v.UpdateData = object.ClassMonster, 0x11012, unsafe.Pointer(u)
						u.Field518, u.Field1 = math.Float32bits(absorption), math.Float32bits(0.25)
						carry, marker, kind = &u.Field1, &u.Field547, &u.Field546
						if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(v.UpdateData) <= math.MaxUint32 {
							t.Fatal("C update not above 4 GiB")
						}
						if target == "monster" {
							v.ObjSubClass = 0
							v.Damage = srv.Types.ByID("NativeWorldFistFlameDefault").Damage
						}
						if target == "immune-NPC" {
							v.ObjSubClass |= 0x400
						}
					}
					damage, typ, effective := int32(200), object.DamageCrush, int32(200)
					if absorption != 0 && target != "monster" {
						effective = 150
					}
					if flame {
						a.TypeInd, a.ObjClass, a.ObjFlags = 1399, 0, object.Flags(16777796)
						w.TypeInd, w.ObjClass, w.ObjSubClass, w.ObjFlags = 695, object.Class(2621441), 1, object.Flags(553665028)
						damage, typ, effective = 64, object.DamageFlame, 64
						if target == "player" {
							v.ObjFlags = object.Flags(150995460)
						}
						if target == "immune-NPC" {
							effective = 0
						}
					}
					beforeA, beforeW, beforeState := *a, *w, stateCalls
					for hit := range 3 {
						wantHP := uint16(2000 - int32(hit+1)*effective)
						wantMarker := uint32(1)
						// DefaultDamage resets the NPC marker before immunity,
						// retaining the weapon-type word from PlayerDamage.
						if effective == 0 {
							wantMarker = 0
						}
						if !objectDamageDispatchCallNative(v, a, w, damage, typ) || v.HealthData.Cur != wantHP || *carry != math.Float32bits(0.25) || *marker != wantMarker || *kind != uint32(w.TypeInd) || *a != beforeA || *w != beforeW {
							t.Fatalf("actual world C hit=%d HP=%d/%d marker=%d/%d carry=%g flags=%x", hit, v.HealthData.Cur, wantHP, *marker, *kind, math.Float32frombits(*carry), uint32(v.ObjFlags))
						}
						if effective != 0 && (v.Obj130 != w || v.Field131 != uint32(typ) || v.Frame134 != srv.Frame() || v.Pos132 != w.PrevPos || v.Field38 != math.MaxUint32) {
							t.Fatal("actual world C attribution/sync lost")
						}
						if effective == 0 && v.Obj130 != nil {
							t.Fatal("immune NPC attribution written")
						}
					}
					wantStates := 0
					if target == "player" {
						wantStates = 3
						if v.UpdateDataPlayer().State != server.PlayerState30 {
							t.Fatal("hurt state missing")
						}
					}
					if stateCalls-beforeState != wantStates {
						t.Fatalf("hurt-state calls=%d want=%d", stateCalls-beforeState, wantStates)
					}
					t.Logf("C-owned >4 GiB C->registered damage->DefaultDamage->UnitSetHP: target=%s source=%p weapon=%p flame=%t armor=%g HP=2000->%d", target, a, w, flame, absorption, v.HealthData.Cur)
				})
			}
		}
	}
}
