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

func selfWeaponElectricNativeAlloc[T comparable](t *testing.T, value T) *T {
	t.Helper()
	p, free := alloc.New(value)
	*p = value
	t.Cleanup(free)
	return p
}

// Actual C dispatch, registered PlayerDamage, electric scale, DefaultDamage
// and UnitSetHP. No damage or HP callback is replaced. The original raw8,
// scale1, carry.75 context forwards8 and retains.75 on every repeated hit.
func TestPlayerDamageSelfWeaponElectricNative4E17B0ProductionHP(t *testing.T) {
	srv := npcReflectLegacyServer4E17B0(t)
	srv.Audio.Init(srv)
	t.Cleanup(srv.Audio.Free)
	srv.FreeObjectTypes()
	t.Cleanup(srv.FreeObjectTypes)
	if err := srv.Types.ReadObjectType(&things.Thing{Name: "NativeSelfWeaponElectric", OnDamage: &things.ProcFunc{Name: "PlayerDamage"}}); err != nil {
		t.Fatal(err)
	}
	oldState := Nox_xxx_playerSetState_4FA020
	Nox_xxx_playerSetState_4FA020 = func(*server.Object, server.PlayerState) bool {
		t.Fatal("below-threshold electric hit requested hurt state")
		return false
	}
	t.Cleanup(func() { Nox_xxx_playerSetState_4FA020 = oldState })
	for _, playerTarget := range []bool{false, true} {
		for _, playerSource := range []bool{false, true} {
			for _, typ := range []object.DamageType{object.DamageElectric, object.DamageAirborneElectric} {
				t.Run(fmt.Sprintf("player-%t/source-player-%t/type-%d", playerTarget, playerSource, typ), func(t *testing.T) {
					makeUnit := func(player bool, index int) *server.Object {
						hp := selfWeaponElectricNativeAlloc(t, server.HealthData{Cur: 200, Max: 200, Field2: 200})
						u := selfWeaponElectricNativeAlloc(t, server.Object{TypeInd: 1343, ObjClass: object.ClassMonster, ObjSubClass: 0x11012, HealthData: hp, Material: 0x4000})
						if player {
							info := selfWeaponElectricNativeAlloc(t, server.Player{PlayerUnit: u, PlayerInd: byte(index)})
							update := selfWeaponElectricNativeAlloc(t, server.PlayerUpdateData{Player: info, State: server.PlayerState13})
							u.TypeInd, u.ObjClass, u.ObjSubClass, u.UpdateData = 713, object.ClassPlayer, 0, unsafe.Pointer(update)
						} else {
							u.UpdateData = unsafe.Pointer(selfWeaponElectricNativeAlloc(t, server.MonsterUpdateData{}))
						}
						return u
					}
					target, source := makeUnit(playerTarget, 8), makeUnit(playerSource, 9)
					target.TeamVal.ID, source.TeamVal.ID = 2, 1
					source.PrevPos = types.Ptf(12.5, -40.25)
					target.Damage = srv.Types.ByID("NativeSelfWeaponElectric").Damage
					var carry, marker, kind *uint32
					if playerTarget {
						u := target.UpdateDataPlayer()
						u.Field21, u.Field76, u.Field75, u.Field40_1 = math.Float32bits(.75), 99, 77, 0xabcd
						carry, marker, kind = &u.Field21, &u.Field76, &u.Field75
					} else {
						u := target.UpdateDataMonster()
						u.Field1, u.Field547, u.Field546 = math.Float32bits(.75), 99, 77
						carry, marker, kind = &u.Field1, &u.Field547, &u.Field546
					}
					if !srv.IsEnemyTo(target, source) {
						t.Fatal("native fixture must be hostile")
					}
					if unsafe.Sizeof(uintptr(0)) > 4 && (uintptr(target.CObj()) <= math.MaxUint32 || uintptr(source.CObj()) <= math.MaxUint32 || uintptr(target.UpdateData) <= math.MaxUint32) {
						t.Fatal("expected real C-owned pointers above 4 GiB")
					}
					for hit := 0; hit < 3; hit++ {
						ok := objectDamageDispatchCallNative(target, source, source, 8, typ)
						if !ok || target.HealthData.Cur != uint16(200-8*(hit+1)) || *carry != math.Float32bits(.75) || *marker != 2 || *kind != uint32(typ) || target.Obj130 != source || target.Pos132 != source.PrevPos || target.Field131 != uint32(typ) || target.Frame134 != 1400 || target.Field38 != math.MaxUint32 {
							t.Fatalf("real C hit=%d result=%t HP=%d carry=%08x marker=%d/%d", hit, ok, target.HealthData.Cur, *carry, *marker, *kind)
						}
					}
					if playerTarget && target.UpdateDataPlayer().Field40_1 != 0xabcd {
						t.Fatal("electric low WORD store overwrote adjacent state")
					}
				})
			}
		}
	}
}
