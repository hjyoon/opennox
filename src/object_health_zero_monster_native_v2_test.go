package opennox

import (
	"bytes"
	"fmt"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

// The complete legacy MonsterDie runtime remains bound: real stack reset,
// buff clearing, shadow cleanup, owner/slave cleanup, inventory dispatch and
// owner HP reporting. These inputs have no owner, inventory or active buff16.
// This is not a stock world/death-animation or player-respawn test.
func TestObjectSetHealthZeroNativeMonsterLifecycle(t *testing.T) {
	for _, subclass := range []uint32{0x302, 0x11012, 0x51002} {
		for _, action := range []ai.ActionType{ai.ACTION_IDLE, ai.ACTION_FIGHT} {
			for _, dead := range []bool{false, true} {
				for _, god := range []bool{false, true} {
					for _, requested := range []int{-1, 0} {
						t.Run(fmt.Sprintf("subclass-%x/action-%d/dead-%t/god-%t/value-%d", subclass, action, dead, god, requested), func(t *testing.T) {
							s := scriptNSRoamFlagServer515C80(t)
							noxflags.ResetGame()
							noxflags.SetGame(noxflags.GameHost)
							oldEngine := noxflags.GetEngine()
							noxflags.UnsetEngine(noxflags.EngineGodMode)
							t.Cleanup(func() { noxflags.SetEngine(oldEngine); noxflags.UnsetEngine(^oldEngine) })
							if god {
								noxflags.SetEngine(noxflags.EngineGodMode)
							}
							s.SetFrame(37)
							s.Audio.Init(s.Server)
							t.Cleanup(s.Audio.Free)
							if !s.Objs.Init(2) {
								t.Fatal("cannot initialize the real native object allocator")
							}
							t.Cleanup(s.Objs.FreeObjects)
							unit := s.Objs.NewObject(&server.ObjectType{})
							if unit == nil {
								t.Fatal("cannot allocate a server-associated monster")
							}
							update, freeUpdate := alloc.New(server.MonsterUpdateData{})
							health, freeHealth := alloc.New(server.HealthData{})
							t.Cleanup(freeHealth)
							t.Cleanup(freeUpdate)
							*health = server.HealthData{Cur: 17, Max: 31, Field2: 0x4321, Field16: 0x87654321}
							*update = server.MonsterUpdateData{AIStackInd: 0, Field2: 19, Field67: 23, Field74: 29,
								Field124: 101, Field137: 103}
							update.AIStack[0] = server.AIStackItem{Action: uint32(action), Args: [4]uintptr{101, 102, 103, 104}}
							unit.ObjClass, unit.ObjSubClass = object.ClassMonster, object.SubClass(subclass)
							unit.ObjFlags = object.FlagEnabled | object.FlagMissileHit
							if dead {
								unit.ObjFlags |= object.FlagDead
							}
							unit.TypeInd, unit.Extent, unit.Field38 = 420, 0xf1234567, 0x12345678
							unit.HealthData, unit.UpdateData = health, unsafe.Pointer(update)
							unit.Buffs, unit.BuffsDur[3], unit.BuffsPower[3] = 1<<3, 77, 8
							s.Objs.List = unit
							t.Cleanup(func() { s.Objs.List, s.Objs.DeletedList = nil, nil })
							for _, ptr := range []unsafe.Pointer{unsafe.Pointer(unit), unsafe.Pointer(update), unsafe.Pointer(health)} {
								if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(ptr) <= math.MaxUint32 {
									t.Fatal("monster death fixture is not above 4 GiB")
								}
							}
							initialUnit, initialUpdate, initialHealth := *unit, *update, *health
							reset := func() {
								*unit, *update, *health = initialUnit, initialUpdate, initialHealth
								s.Objs.List, s.Objs.DeletedList = unit, nil
								s.Audio.Reset()
								s.AI.StackChanged = false
							}
							// Actual registered native lethal damage is the independent
							// lifecycle control, including its whole-record side effects.
							legacy.Nox_xxx_unitDamageClear_4EE5E0(unit, int(initialHealth.Cur))
							if health.Cur != 0 || !unit.Flags().Has(object.FlagDead) || s.Objs.DeletedList != nil {
								t.Fatal("native lethal control did not enter monster death")
							}
							if !dead && (update.AIStackInd != 1 || update.AIStack[0].Type() != ai.ACTION_DEAD ||
								update.AIStack[1].Type() != ai.ACTION_DYING || unit.Buffs != 0 ||
								unit.BuffsDur[3] != 0 || unit.BuffsPower[3] != 0 || unit.Flags().Has(object.FlagMissileHit) ||
								uint32(unit.SubClass()) != subclass&^0x100) {
								t.Fatal("actual MonsterDie did not complete its native stack/buff/flag cleanup")
							}
							wantedUnit := scriptNSRoamFlagBytes515C80(unsafe.Pointer(unit), unsafe.Sizeof(*unit))
							wantedUpdate := scriptNSRoamFlagBytes515C80(unsafe.Pointer(update), unsafe.Sizeof(*update))
							wantedHealth, wantedStack := *health, s.AI.StackChanged
							check := func(label string) {
								t.Helper()
								if !bytes.Equal(wantedUnit, scriptNSRoamFlagBytes515C80(unsafe.Pointer(unit), unsafe.Sizeof(*unit))) ||
									!bytes.Equal(wantedUpdate, scriptNSRoamFlagBytes515C80(unsafe.Pointer(update), unsafe.Sizeof(*update))) ||
									*health != wantedHealth || s.AI.StackChanged != wantedStack ||
									s.Objs.List != unit || s.Objs.DeletedList != nil || s.Frame() != 37 ||
									noxflags.HasEngine(noxflags.EngineGodMode) != god {
									t.Errorf("%s differs from actual native monster death: HP=%d dead=%t stack=%v", label, health.Cur, unit.Flags().Has(object.FlagDead), update.GetAIStack())
								}
								if events, _ := luaCompatAudioEvents(t, s); len(events) != 0 {
									t.Errorf("%s emitted an unexpected immediate death sound", label)
								}
							}
							for _, route := range []string{"root-object", "public-sdk", "legacy-object-interface"} {
								reset()
								switch route {
								case "root-object":
									asObjectS(unit).SetHealth(requested)
								case "public-sdk":
									s.noxScriptP().toObj(unit).SetHealth(requested)
								case "legacy-object-interface":
									scrObject{asObjectS(unit)}.SetHealth(requested)
								}
								check(route)
								asObjectS(unit).SetHealth(0)
								check(route + " repeated")
							}
						})
					}
				}
			}
		}
	}
}
