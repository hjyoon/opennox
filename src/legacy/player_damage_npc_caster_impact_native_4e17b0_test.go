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

// Synthetic C-owned >4GiB records use the real C dispatcher, registered NPC
// PlayerDamage export, DefaultDamage and UnitDamageClear/UnitSetHP. Neither HP
// nor damage attribution is supplied by a replacement callback. Absorption and
// carry are fixture inputs; actual stock-map Earthquake is a separate test.
func TestPlayerDamageNPCCasterImpactNative4E17B0CRecords(t *testing.T) {
	srv := npcReflectLegacyServer4E17B0(t)
	srv.Audio.Init(srv)
	t.Cleanup(srv.Audio.Free)
	srv.FreeObjectTypes()
	t.Cleanup(srv.FreeObjectTypes)
	if err := srv.Types.ReadObjectType(&things.Thing{Name: "NativeCasterImpactNPC", OnDamage: &things.ProcFunc{Name: "PlayerDamage"}}); err != nil {
		t.Fatal(err)
	}
	if !srv.Objs.Init(200) {
		t.Fatal("NPC caster IMPACT allocator")
	}
	t.Cleanup(srv.Objs.FreeObjects)
	for _, owner := range []string{"player-self", "player-owned-NPC", "imaginary-self", "imaginary-owned-NPC", "world-self", "nil"} {
		for _, absorption := range []float32{0, 0.25, 1} {
			for _, raw := range []int32{3, 8, 19} {
				t.Run(fmt.Sprintf("%s/armor-%g/raw-%d", owner, absorption, raw), func(t *testing.T) {
					v := srv.Objs.NewObject(&server.ObjectType{Damage: srv.Types.ByID("NativeCasterImpactNPC").Damage})
					a, w := srv.Objs.NewObject(&server.ObjectType{}), srv.Objs.NewObject(&server.ObjectType{})
					ud, freeUD := alloc.New(server.MonsterUpdateData{})
					casterUD, freeCasterUD := alloc.New(server.MonsterUpdateData{})
					player, freePlayer := alloc.New(server.Player{})
					playerUD, freePlayerUD := alloc.New(server.PlayerUpdateData{})
					health, freeHealth := alloc.New(server.HealthData{})
					for _, free := range []func(){freeUD, freeCasterUD, freePlayer, freePlayerUD, freeHealth} {
						t.Cleanup(free)
					}
					*ud = server.MonsterUpdateData{Field547: 99, Field546: 77, Field523_2: 88, Field1: math.Float32bits(0.125), Field518: math.Float32bits(absorption)}
					*casterUD = server.MonsterUpdateData{Field523_2: 66, Field1: math.Float32bits(0.375)}
					*player = server.Player{PlayerInd: 9}
					*playerUD = server.PlayerUpdateData{Player: player, State: server.PlayerState13}
					*health = server.HealthData{Cur: 2000, Field2: 2000, Max: 2000}
					v.TypeInd, v.ObjClass, v.ObjSubClass, v.ObjFlags, v.Material = 71, object.ClassMonster, 0x11012, 0, 0x4000
					v.UpdateData, v.HealthData, v.Frame134, v.Pos132 = unsafe.Pointer(ud), health, 77, types.Ptf(31, 47)
					a.TypeInd, a.PrevPos = 713, types.Ptf(101, 103)
					w.TypeInd, w.PrevPos = 1399, types.Ptf(17, 19)
					switch owner {
					case "player-self":
						a.ObjClass, a.UpdateData, w = object.ClassPlayer, unsafe.Pointer(playerUD), a
					case "player-owned-NPC":
						a.ObjClass, a.UpdateData, w.ObjOwner = object.ClassPlayer, unsafe.Pointer(playerUD), a
						w.ObjClass, w.ObjSubClass, w.UpdateData = object.ClassMonster, 0x11012, unsafe.Pointer(casterUD)
					case "imaginary-self":
						w = a
					case "imaginary-owned-NPC":
						w.ObjOwner, w.ObjClass, w.ObjSubClass, w.UpdateData = a, object.ClassMonster, 0x11012, unsafe.Pointer(casterUD)
					case "world-self":
						w.ObjClass, a = object.ClassSimple, w
					case "nil":
						a = nil
					}
					for _, pointer := range []unsafe.Pointer{unsafe.Pointer(v), unsafe.Pointer(w), unsafe.Pointer(ud), unsafe.Pointer(casterUD), unsafe.Pointer(player), unsafe.Pointer(playerUD), unsafe.Pointer(health)} {
						if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(pointer) <= math.MaxUint32 {
							t.Fatalf("NPC caster IMPACT pointer=%p, want >4 GiB", pointer)
						}
					}
					if a != nil && (!srv.IsEnemyTo(v, a) || w.FindOwnerChainPlayer() != a) {
						t.Fatal("native NPC caster hostility/terminal owner fixture")
					}
					beforeCaster, beforeCasterUD, beforePlayer, beforePlayerUD := *w, *casterUD, *player, *playerUD
					beforeSource := server.Object{}
					if a != nil {
						beforeSource = *a
					}
					accumulated := float32((1-float64(absorption))*float64(raw)) + float32(0.125)
					rounded := int32(math.RoundToEven(float64(accumulated)))
					effective := max(rounded, 1)
					if !objectDamageDispatchCallNative(v, a, w, raw, object.DamageImpact) {
						t.Fatal("native NPC caster IMPACT result")
					}
					position := w.PrevPos
					if a == nil {
						position = types.Pointf{}
					}
					marker, markerType := uint32(2), uint32(11)
					if a != nil && a != w {
						marker, markerType = 1, uint32(w.TypeInd)
					}
					if health.Cur != uint16(2000-effective) || health.Field2 != 2000 || health.Max != 2000 || v.Obj130 != w || v.Field131 != 11 || v.Frame134 != 1400 || v.Pos132 != position || v.Field38 != math.MaxUint32 ||
						ud.Field547 != marker || ud.Field546 != markerType || !ud.StatusFlags.Has(object.MonStatusInjured) || ud.StatusFlags.Has(object.MonStatusOnFire) || ud.Field523_2 != 88 ||
						ud.Field1 != math.Float32bits(accumulated-float32(rounded)) || ud.Field518 != math.Float32bits(absorption) {
						t.Fatalf("C->NPC PlayerDamage->DefaultDamage->UnitSetHP: HP=%d want=%d marker=%d/%d carry=%g", health.Cur, 2000-effective, ud.Field547, ud.Field546, math.Float32frombits(ud.Field1))
					}
					if *w != beforeCaster || (a != nil && *a != beforeSource) || *casterUD != beforeCasterUD || *player != beforePlayer || *playerUD != beforePlayerUD {
						t.Fatal("native NPC damage changed the complete caster or terminal owner")
					}
					t.Logf("C->NPC PlayerDamage->DefaultDamage->UnitSetHP: target=%p update=%p source=%p caster=%p owner=%s absorption=%g raw=%d effective=%d HP=2000->%d", v, ud, a, w, owner, absorption, raw, effective, health.Cur)
				})
			}
		}
	}
}
