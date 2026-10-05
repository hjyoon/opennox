package legacy

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

// Exercise the production wrapper and actual C missile query, native wall and
// obstacle services, native tile query, and server RNG. A successful probe,
// destination, HP result, or screen image is never supplied by a test hook.
func TestMonsterMainDodge547210ActualLegacyBinding(t *testing.T) {
	for _, mode := range []string{"incoming", "outgoing", "own-missile", "not-indexed", "noncoop", "definition-disabled"} {
		t.Run(mode, func(t *testing.T) {
			srv, unit, update, missile := monsterMainBlockLegacyFixture547210(t)
			definition, freeDefinition := alloc.New(server.MonsterDef{})
			t.Cleanup(freeDefinition)
			definition.StatusFlags92, definition.RunMultiplier96 = object.MonStatusCanDodge, 1
			update.MonsterDef = definition
			unit.SpeedCur = 10
			noxflags.SetGame(noxflags.GameModeCoop)
			want := mode == "incoming"
			switch mode {
			case "outgoing":
				missile.VelVec.X = 4
			case "own-missile":
				missile.ObjOwner = unit
			case "noncoop":
				noxflags.UnsetGame(noxflags.GameModeCoop)
			case "definition-disabled":
				definition.StatusFlags92 = 0
			}
			if mode != "not-indexed" {
				srv.Map.AddObjectToIndex(missile)
			}
			probe := Nox_xxx_monsterTestBlockShield_533E70(unit)
			if (probe != 0) != (mode != "outgoing" && mode != "own-missile" && mode != "not-indexed") {
				t.Fatalf("actual missile query=%d for %s", probe, mode)
			}
			if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(definition)) <= math.MaxUint32 {
				t.Fatalf("definition=%p, want C-owned native address above 4 GiB", definition)
			}
			before := *update
			t.Logf("actual MainAI/C missile/tile services: unit=%p update=%p definition=%p missile=%p probe=%d", unit, update, definition, missile, probe)
			Nox_xxx_monsterMainAIFn_547210(unit)
			if !want {
				if *update != before || srv.AI.StackChanged || unit.HealthData.Cur != 100 {
					t.Fatal("unqualified native missile or gate scheduled dodge")
				}
				return
			}
			head := update.AIStackHead()
			destination := head.ArgPos(0)
			if update.AIStackInd != 2 || update.AIStack[0].Type() != ai.ACTION_FIGHT ||
				update.AIStack[1].Type() != ai.DEPENDENCY_TIME || update.AIStack[1].ArgU32(0) != 1431 ||
				head.Type() != ai.ACTION_DODGE || destination.X != 300 || math.Abs(float64(destination.Y-300)) != 15 ||
				head.ArgU32(2) != 0 || !srv.AI.StackChanged || unit.HealthData.Cur != 100 {
				t.Fatalf("actual dodge missing: stack=%+v destination=%v changed=%t HP=%d", update.GetAIStack(), destination, srv.AI.StackChanged, unit.HealthData.Cur)
			}
			// The independently restored live action must consume that stack;
			// this verifies velocity scheduling, not actual world movement.
			if !srv.MonsterActionDodge544640(unit) || unit.VelVec.X != 0 ||
				math.Abs(float64(unit.VelVec.Y)) < 9.9 || float64(unit.VelVec.Y)*float64(destination.Y-300) <= 0 {
				t.Fatalf("real dodge action did not use native destination: velocity=%v", unit.VelVec)
			}
		})
	}
}
