package legacy

import (
	"fmt"
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/server"
)

func npcReflectLegacyServer4E17B0(t *testing.T) *server.Server {
	t.Helper()
	srv := server.New(nil, nil, strman.New())
	t.Cleanup(srv.Close)
	srv.SetFrame(1400)
	oldGetServer := GetServer
	GetServer = func() Server { return &itemDurabilityLegacyServer4E1560{srv: srv} }
	t.Cleanup(func() { GetServer = oldGetServer })
	oldGame, oldEngine, oldGameplay := noxflags.GetGame(), noxflags.GetEngine(), noxflags.GetGamePlay()
	noxflags.UnsetGame(oldGame)
	noxflags.UnsetEngine(oldEngine)
	noxflags.UnsetGamePlay(oldGameplay)
	noxflags.SetGamePlay(noxflags.GameplayFlag1)
	t.Cleanup(func() {
		noxflags.UnsetGame(noxflags.GetGame())
		noxflags.UnsetEngine(noxflags.GetEngine())
		noxflags.UnsetGamePlay(noxflags.GetGamePlay())
		noxflags.SetGame(oldGame)
		noxflags.SetEngine(oldEngine)
		noxflags.SetGamePlay(oldGameplay)
	})
	return srv
}

func npcReflectLegacyObjects4E17B0(t *testing.T, pin *runtime.Pinner) (target, source *server.Object) {
	t.Helper()
	sourcePlayer := &server.Player{}
	sourceUpdate := &server.PlayerUpdateData{Player: sourcePlayer, State: server.PlayerState13}
	npcUpdate := &server.MonsterUpdateData{Field547: 99, Field546: 77, Field523_2: 99}
	source = &server.Object{TypeInd: 71, ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(sourceUpdate), HealthData: &server.HealthData{Cur: 200, Max: 200}, Material: 0x4000}
	target = &server.Object{TypeInd: 72, ObjClass: object.ClassMonster, ObjSubClass: 0x11012, UpdateData: unsafe.Pointer(npcUpdate), HealthData: &server.HealthData{Cur: 200, Max: 200}, Material: 0x4000, Buffs: 1 << server.ENCHANT_REFLECTIVE_SHIELD}
	target.Damage = playerDamageMeleeCallbackNative4E17B0()
	for _, pointer := range []unsafe.Pointer{unsafe.Pointer(source), unsafe.Pointer(target), unsafe.Pointer(sourcePlayer), unsafe.Pointer(sourceUpdate), unsafe.Pointer(npcUpdate), unsafe.Pointer(source.HealthData), unsafe.Pointer(target.HealthData)} {
		pin.Pin(pointer)
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(pointer) <= math.MaxUint32 {
			t.Fatalf("pointer=%p, want native high address", pointer)
		}
	}
	return
}

func TestPlayerDamageNPCReflectNativeCallback4E17B0ElectricHP(t *testing.T) {
	srv := npcReflectLegacyServer4E17B0(t)
	for _, selfWeapon := range []bool{false, true} {
		for _, typ := range []object.DamageType{object.DamageElectric, object.DamageAirborneElectric} {
			for _, front := range []bool{false, true} {
				t.Run(fmt.Sprintf("self-%t/%s/front-%t", selfWeapon, typ, front), func(t *testing.T) {
					var pin runtime.Pinner
					defer pin.Unpin()
					target, source := npcReflectLegacyObjects4E17B0(t, &pin)
					source.PosVec = types.Pointf{X: 20}
					if !front {
						source.PosVec.X = -20
					}
					source.PrevPos = types.Pointf{X: -source.PosVec.X, Y: 7}
					if facing := Nox_server_testTwoPointsAndDirection_4E6E50(target.PosVec, int16(target.Direction1), source.PosVec)&1 != 0; facing != front || !srv.IsEnemyTo(target, source) {
						t.Fatal("native fixture facing/hostility")
					}
					weapon := (*server.Object)(nil)
					if selfWeapon {
						weapon = source
					}
					ud := target.UpdateDataMonster()
					before := *ud
					reflected := front && typ == object.DamageAirborneElectric
					// The real C dispatcher calls the registered production adapter,
					// its production facing helper and DefaultDamage/UnitSetHP.
					result := objectDamageDispatchCallNative(target, source, weapon, 8, typ)
					if reflected {
						before.Field547 = 0
						if result || target.HealthData.Cur != 200 || *ud != before || target.Obj130 != nil || target.Field38 != 0 {
							t.Fatalf("frontal reflection: result=%t HP=%d marker=%d/%d", result, target.HealthData.Cur, ud.Field547, ud.Field546)
						}
					} else if !result || target.HealthData.Cur != 192 || ud.Field547 != 2 || ud.Field546 != uint32(typ) || ud.Field523_2 != 2 || target.Obj130 != source || target.Field131 != uint32(typ) || target.Frame134 != 1400 || target.Field38 != math.MaxUint32 {
						t.Fatalf("rear/type-9 native damage: result=%t HP=%d marker=%d/%d source=%p", result, target.HealthData.Cur, ud.Field547, ud.Field546, target.Obj130)
					}
					if target.Buffs != 1<<server.ENCHANT_REFLECTIVE_SHIELD {
						t.Fatal("native reflection removed the buff")
					}
					t.Logf("C->PlayerDamage NPC: target=%p source=%p weapon=%p type=%d front=%t reflected=%t HP=200->%d", target, source, weapon, typ, front, reflected, target.HealthData.Cur)
					runtime.KeepAlive(source)
					runtime.KeepAlive(target)
				})
			}
		}
	}
}

func TestPlayerDamageNPCReflectNativeCallback4E17B0MissileOwnership(t *testing.T) {
	srv := npcReflectLegacyServer4E17B0(t)
	if !srv.Objs.Init(4) {
		t.Fatal("native projectile allocator initialization failed")
	}
	t.Cleanup(srv.Objs.FreeObjects)
	for _, flags := range []uint32{0, 0x40} {
		for _, typ := range []object.DamageType{object.DamageImpact, object.DamageZapRay} {
			t.Run(fmt.Sprintf("subclass-%x/%s", flags, typ), func(t *testing.T) {
				var pin runtime.Pinner
				defer pin.Unpin()
				target, source := npcReflectLegacyObjects4E17B0(t, &pin)
				sibling, firstOwned := &server.Object{}, &server.Object{}
				// The real owner service consults this projectile's server when
				// checking monitoring; retain the allocator's native server handle.
				missile := srv.Objs.NewObject(&server.ObjectType{})
				missile.TypeInd, missile.ObjClass, missile.ObjSubClass = 73, object.ClassMissile, object.SubClass(flags)
				missile.PosVec, missile.PrevPos = types.Pointf{X: 20}, types.Pointf{X: -20, Y: 3}
				missile.VelVec, missile.Direction1 = types.Pointf{X: -4}, 128
				for _, pointer := range []unsafe.Pointer{unsafe.Pointer(sibling), unsafe.Pointer(firstOwned), unsafe.Pointer(missile)} {
					pin.Pin(pointer)
					if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(pointer) <= math.MaxUint32 {
						t.Fatalf("missile/owner link=%p, want native high address", pointer)
					}
				}
				// Pin every Go-owned link before writing it into the C-owned
				// projectile, and remove those links before its pins are released.
				missile.ObjOwner, missile.Field128 = source, sibling
				defer func() { missile.ObjOwner, missile.Field128 = nil, nil }()
				source.Field129, target.Field129 = missile, firstOwned
				before := *target.UpdateDataMonster()
				// No reflection or owner-list replacement: both production helper
				// implementations must keep the complete native pointer links.
				if objectDamageDispatchCallNative(target, source, missile, 0, typ) {
					t.Fatal("zero-damage frontal missile was not reflected before damage admission")
				}
				if flags&0x40 == 0 {
					if missile.ObjOwner != target || target.Field129 != missile || missile.Field128 != firstOwned || source.Field129 != sibling {
						t.Fatal("native reflected missile lost its transferred owner-list links")
					}
				} else if missile.ObjOwner != source || source.Field129 != missile || missile.Field128 != sibling || target.Field129 != firstOwned {
					t.Fatal("subclass 0x40 changed owner-list links")
				}
				before.Field547 = 0
				if *target.UpdateDataMonster() != before || target.HealthData.Cur != 200 || missile.Direction1 != 0 || missile.NewPos != missile.PrevPos || !(missile.VelVec.X > 0) || missile.VelVec.Y != 0 {
					t.Fatalf("native reflection state: HP=%d direction=%d velocity=%v position=%v", target.HealthData.Cur, missile.Direction1, missile.VelVec, missile.NewPos)
				}
				t.Logf("C->PlayerDamage NPC projectile: missile=%p target=%p owner=%p subclass=%#x velocity=%v", missile, target, missile.ObjOwner, flags, missile.VelVec)
				runtime.KeepAlive(missile)
				runtime.KeepAlive(source)
				runtime.KeepAlive(target)
			})
		}
	}
}
