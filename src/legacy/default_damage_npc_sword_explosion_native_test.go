package legacy

import (
	"fmt"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/things"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

// Exact reported stock records enter the real C dispatcher, registered
// SkeletonDamage/DefaultDamage, fire protection and UnitSetHP. All records
// are C-owned high addresses; this is controlled native, not natural GUI combat.
func TestDefaultDamageNPCSwordExplosionNative4E0B30(t *testing.T) {
	srv := npcReflectLegacyServer4E17B0(t)
	srv.Audio.Init(srv)
	t.Cleanup(srv.Audio.Free)
	srv.FreeObjectTypes()
	t.Cleanup(srv.FreeObjectTypes)
	for _, callback := range []string{"DefaultDamage", "SkeletonDamage"} {
		if err := srv.Types.ReadObjectType(&things.Thing{Name: "NativeArmedExplosion" + callback, OnDamage: &things.ProcFunc{Name: callback}}); err != nil {
			t.Fatal(err)
		}
	}
	if !srv.Objs.Init(64) {
		t.Fatal("C-owned object allocator")
	}
	t.Cleanup(srv.Objs.FreeObjects)
	for _, tc := range []struct {
		name            string
		typeID          uint16
		subclass, flags uint32
		callback        string
	}{
		{"Skeleton", 1314, 70210, 17302020, "SkeletonDamage"},
		{"NPC", 1343, 69778, 16777732, "DefaultDamage"},
		{"Troll", 2273, 66050, 151519748, "DefaultDamage"},
	} {
		for _, playerSource := range []bool{false, true} {
			for _, defense := range []string{"none", "invulnerable", "fire-immune"} {
				t.Run(fmt.Sprintf("%s/player-source-%t/%s", tc.name, playerSource, defense), func(t *testing.T) {
					v, a, w := srv.Objs.NewObject(&server.ObjectType{}), srv.Objs.NewObject(&server.ObjectType{}), srv.Objs.NewObject(&server.ObjectType{})
					u, freeU := alloc.New(server.MonsterUpdateData{})
					au, freeAU := alloc.New(server.MonsterUpdateData{})
					pu, freePU := alloc.New(server.PlayerUpdateData{})
					info, freeInfo := alloc.New(server.Player{})
					hp, freeHP := alloc.New(server.HealthData{})
					init, freeInit := alloc.New(server.ModifierInitData{})
					for _, free := range []func(){freeU, freeAU, freePU, freeInfo, freeHP, freeInit} {
						t.Cleanup(free)
					}
					*u = server.MonsterUpdateData{Field547: 99, Field546: 77, Field523_2: 88, Field1: math.Float32bits(0.125)}
					*hp = server.HealthData{Cur: 2000, Max: 2000, Field2: 2000}
					*info = server.Player{PlayerInd: 9}
					*pu = server.PlayerUpdateData{Player: info, State: server.PlayerState13}
					v.TypeInd, v.ObjClass, v.ObjSubClass, v.ObjFlags, v.Material = tc.typeID, object.Class(2097154), object.SubClass(tc.subclass), object.Flags(tc.flags), 0x4000
					v.UpdateData, v.HealthData, v.Damage = unsafe.Pointer(u), hp, srv.Types.ByID("NativeArmedExplosion"+tc.callback).Damage
					a.TypeInd, a.ObjClass, a.ObjSubClass, a.ObjFlags, a.UpdateData = 1343, object.Class(2097154), 69778, object.Flags(16777732), unsafe.Pointer(au)
					if playerSource {
						a.TypeInd, a.ObjClass, a.ObjSubClass, a.UpdateData = 713, object.Class(2621444), 0, unsafe.Pointer(pu)
					}
					a.PrevPos = types.Ptf(32.5, -18.25)
					w.TypeInd, w.ObjClass, w.ObjSubClass, w.ObjFlags, w.InitData = 1174, object.Class(85458944), 256, object.Flags(17301776), unsafe.Pointer(init)
					w.PrevPos = types.Ptf(44, 7)
					// Team-less MONSTER owners are allies in the real enemy
					// service. Declare opposing native team IDs for this hostile
					// fixture; keep all reported class/subclass/flags unchanged.
					v.TeamVal.ID, a.TeamVal.ID = 2, 1
					if !srv.IsEnemyTo(v, a) {
						t.Fatal("hostile fixture rejected by the production enemy service")
					}
					if defense == "invulnerable" {
						v.Buffs |= 1 << server.ENCHANT_INVULNERABLE
					}
					if defense == "fire-immune" {
						v.ObjSubClass |= 0x400
					}
					for _, p := range []unsafe.Pointer{unsafe.Pointer(v), unsafe.Pointer(a), unsafe.Pointer(w), unsafe.Pointer(u), unsafe.Pointer(au), unsafe.Pointer(pu), unsafe.Pointer(info), unsafe.Pointer(hp), unsafe.Pointer(init)} {
						if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(p) <= math.MaxUint32 {
							t.Fatalf("C-owned pointer=%p, want >4 GiB", p)
						}
					}
					beforeA, beforeW, beforePU := *a, *w, *pu
					var steps [3]uint16
					for hit := range 3 {
						perHit := 3
						if defense == "fire-immune" {
							perHit = 1
						}
						want, marker := uint16(2000-perHit*(hit+1)), uint32(1)
						if defense == "invulnerable" {
							want, marker = 2000, 99
						}
						if !objectDamageDispatchCallNative(v, a, w, 3, object.DamageExplosion) || hp.Cur != want || hp.Max != 2000 || hp.Field2 != 2000 || u.Field547 != marker || u.Field1 != math.Float32bits(0.125) || *a != beforeA || *w != beforeW || *pu != beforePU {
							t.Fatalf("actual C EXPLOSION hit=%d HP=%d/%d marker=%d/%d", hit, hp.Cur, want, u.Field547, marker)
						}
						if defense != "invulnerable" {
							if v.Obj130 != w || v.Pos132 != a.PrevPos || v.Field131 != 7 || v.Frame134 != 1400 || v.Field38 != math.MaxUint32 || u.Field546 != 1174 || u.Field523_2 != 88 || !u.StatusFlags.Has(object.MonStatusInjured|object.MonStatusOnFire) || (!playerSource && au.Field130 != 1400) {
								t.Fatal("actual C weapon attribution/position/state lost")
							}
						} else if v.Obj130 != nil || v.Frame134 != 0 || u.Field546 != 77 || u.Field523_2 != 88 || u.StatusFlags != 0 || au.Field130 != 0 {
							t.Fatal("invulnerable target acquired injury metadata")
						}
						steps[hit] = hp.Cur
					}
					t.Logf("C-owned >4 GiB real C->%s->DefaultDamage->UnitSetHP target=%p name=%s target_type=%d target_class=2097154 target_subclass=%d target_flags=%d source=%p source_type=%d source_class=%d source_subclass=%d source_flags=16777732 weapon=%p weapon_type=1174 weapon_class=85458944 weapon_subclass=256 weapon_flags=17301776 damage=3 damage_type=7 defense=%s HP=2000->%d->%d->%d", tc.callback, v, tc.name, tc.typeID, tc.subclass, tc.flags, a, a.TypeInd, a.Class(), a.SubClass(), w, defense, steps[0], steps[1], steps[2])
				})
			}
		}
	}
}
