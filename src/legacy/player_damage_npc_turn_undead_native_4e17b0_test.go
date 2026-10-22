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
	"github.com/opennox/opennox/v1/legacy/common/ccall"
	"github.com/opennox/opennox/v1/server"
)

// Deletion is the single declared world-service boundary. Damage dispatch,
// owner resolution, budget stores, SkeletonDamage/DefaultDamage and HP use the
// real production registrations. No spy supplies HP or a damage result.
func TestPlayerDamageNPCTurnUndeadNativeCollisionHP4E17B0(t *testing.T) {
	srv := npcReflectLegacyServer4E17B0(t)
	srv.Audio.Init(srv)
	t.Cleanup(srv.Audio.Free)
	srv.FreeObjectTypes()
	t.Cleanup(srv.FreeObjectTypes)
	if err := srv.Types.ReadObjectType(&things.Thing{Name: "NativeTurnUndeadPlayerDamage", OnDamage: &things.ProcFunc{Name: "PlayerDamage"}}); err != nil {
		t.Fatal(err)
	}
	if !srv.Objs.Init(64) {
		t.Fatal("C-owned object allocator")
	}
	t.Cleanup(srv.Objs.FreeObjects)
	world := &turnUndeadDamageLegacyServer4E0B30{srv: srv}
	GetServer = func() Server { return world }
	collide, size, ok := server.ObjectCollideHandler("UndeadKillerCollide")
	if !ok || collide == nil || size != unsafe.Sizeof(server.UndeadKillerCollideData{}) {
		t.Fatal("native C collision registration")
	}
	for _, owner := range []string{"player", "NPC", "unowned"} {
		for _, defense := range []string{"none", "invulnerable", "no-update", "non-undead"} {
			t.Run(fmt.Sprintf("%s/%s", owner, defense), func(t *testing.T) {
				v, a, killer := srv.Objs.NewObject(&server.ObjectType{}), srv.Objs.NewObject(&server.ObjectType{}), srv.Objs.NewObject(&server.ObjectType{})
				u, freeU := alloc.New(server.MonsterUpdateData{})
				au, freeAU := alloc.New(server.MonsterUpdateData{})
				pu, freePU := alloc.New(server.PlayerUpdateData{})
				info, freeInfo := alloc.New(server.Player{})
				hp, freeHP := alloc.New(server.HealthData{})
				spell, freeSpell := alloc.New(server.DurSpell{})
				data, freeData := alloc.New(server.UndeadKillerCollideData{})
				for _, free := range []func(){freeU, freeAU, freePU, freeInfo, freeHP, freeSpell, freeData} {
					t.Cleanup(free)
				}
				*u = server.MonsterUpdateData{Field547: 99, Field546: 77, Field523_2: 88, Field518: math.Float32bits(.9), Field1: math.Float32bits(0.125)}
				*hp = server.HealthData{Cur: 2000, Max: 2000, Field2: 2000}
				*info = server.Player{PlayerInd: 9}
				*pu = server.PlayerUpdateData{Player: info, State: server.PlayerState13}
				v.TypeInd, v.ObjClass, v.ObjSubClass, v.ObjFlags, v.Material = 1314, object.Class(2097154), 0x11052, object.Flags(17302020), 0x4000
				v.UpdateData, v.HealthData, v.Damage = unsafe.Pointer(u), hp, srv.Types.ByID("NativeTurnUndeadPlayerDamage").Damage
				a.TypeInd, a.ObjClass, a.ObjSubClass, a.ObjFlags, a.UpdateData = 1343, object.Class(2097154), 69778, object.Flags(16777732), unsafe.Pointer(au)
				if owner == "player" {
					a.TypeInd, a.ObjClass, a.ObjSubClass, a.UpdateData = 713, object.Class(2621444), 0, unsafe.Pointer(pu)
				}
				a.PrevPos = types.Ptf(32.5, -18.25)
				// Original LIGHT|CLIENT_PREDICT metadata, no InitData or
				// missile flags. Synthetic ID 777 is attribution only.
				killer.TypeInd, killer.ObjClass, killer.PrevPos, killer.CollideData = 777, object.Class(0x40080000), types.Ptf(44, 7), unsafe.Pointer(data)
				data.Spell = spell
				if owner != "unowned" {
					killer.ObjOwner = a
				}
				v.TeamVal.ID, a.TeamVal.ID = 2, 1
				if owner != "unowned" && !srv.IsEnemyTo(v, a) {
					t.Fatal("production hostility service rejected fixture")
				}
				if owner == "unowned" && killer.FindOwnerChainPlayer() != killer {
					t.Fatal("terminal non-unit owner identity")
				}
				switch defense {
				case "invulnerable":
					v.Buffs |= 1 << server.ENCHANT_INVULNERABLE
				case "no-update":
					v.ObjFlags |= object.FlagNoUpdate
				case "non-undead":
					v.ObjSubClass &^= 0x40
				}
				for _, p := range []unsafe.Pointer{unsafe.Pointer(v), unsafe.Pointer(a), unsafe.Pointer(killer), unsafe.Pointer(u), unsafe.Pointer(au), unsafe.Pointer(pu), unsafe.Pointer(info), unsafe.Pointer(hp), unsafe.Pointer(spell), unsafe.Pointer(data)} {
					if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(p) <= math.MaxUint32 {
						t.Fatalf("C-owned pointer=%p, want >4 GiB", p)
					}
				}
				beforeA, beforeKiller, beforePU := *a, *killer, *pu
				world.deleted = nil
				var steps [3]uint16
				for hit := range 3 {
					spell.Field72 = 7 // Explicit input budget, never an HP/result rewrite.
					ccall.CallVoidPtr3(collide, unsafe.Pointer(killer), unsafe.Pointer(v), nil)
					want, budget, deletions := uint16(2000), int32(0), hit+1
					if defense == "none" {
						want = uint16(2000 - 7*(hit+1))
					}
					if defense == "non-undead" {
						budget, deletions = 7, 0
					}
					if hp.Cur != want || hp.Max != 2000 || hp.Field2 != 2000 || spell.Field72 != budget || len(world.deleted) != deletions || *a != beforeA || *killer != beforeKiller || *pu != beforePU || u.Field1 != math.Float32bits(0.125) {
						t.Fatalf("actual C collision hit=%d HP=%d/%d budget=%d/%d deletes=%d/%d", hit, hp.Cur, want, spell.Field72, budget, len(world.deleted), deletions)
					}
					for _, deleted := range world.deleted {
						if deleted != killer {
							t.Fatal("collision deleted wrong object")
						}
					}
					if defense == "none" {
						marker, kind := uint32(1), uint32(777)
						if owner == "unowned" {
							marker, kind = 2, 6
						}
						if v.Obj130 != killer || v.Pos132 != killer.PrevPos || v.Field131 != 6 || v.Frame134 != 1400 || v.Field38 != math.MaxUint32 || u.Field547 != marker || u.Field546 != kind || u.Field523_2 != 88 || !u.StatusFlags.Has(object.MonStatusInjured|object.MonStatusOnFire) || (owner == "NPC" && au.Field130 != 1400) {
							t.Fatal("native Turn Undead injury/attribution state")
						}
					} else if v.Obj130 != nil || v.Frame134 != 0 || u.Field546 != 77 || au.Field130 != 0 {
						t.Fatal("immune/inactive/ineligible target acquired injury metadata")
					}
					steps[hit] = hp.Cur
				}
				t.Logf("C-owned >4 GiB C->UndeadKillerCollide->registered PlayerDamage->DefaultDamage->UnitSetHP owner=%s defense=%s damage_type=6 budget=7->%d HP=2000->%d->%d->%d; DelayedDelete is an observed world boundary", owner, defense, spell.Field72, steps[0], steps[1], steps[2])
			})
		}
	}
}
