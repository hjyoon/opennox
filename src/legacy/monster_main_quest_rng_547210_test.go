package legacy

import (
	"fmt"
	"math"
	"testing"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/prand"
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/server"
)

// Read-only Quest20 traces first diverged at frame 996 on a MechanicalGolem
// with stack [GUARD, NOT_UNDER_ATTACK, MOVE_TO], unchanged position since
// frame 980, and Logic index 1496. GAME.EXE 005479FA..00547B57 requires the
// frustration and wait RNG at elapsed > FPS/2; the historical ARM64 no-op
// skipped them. Recreate those branch inputs in isolated C-owned records,
// using the production MainAI wrapper and real Logic RNG, not a map-sequence
// oracle or an injected random result. The fixture is not a live golem world.
func TestMonsterMainProgress547210QuestGolemCapturedLogicRNG(t *testing.T) {
	for _, frame := range []uint32{995, 996, 997} {
		t.Run(fmt.Sprint(frame), func(t *testing.T) {
			srv, unit, update, _ := monsterMainBlockLegacyFixture547210(t)
			noxflags.SetGame(noxflags.GameModeQuest)
			srv.SetTickRate(30)
			srv.SetFrame(frame)
			srv.Rand.Logic = prand.New(1496)
			srv.Rand.Other = prand.New(3905)
			unit.ObjFlags, unit.NetCode, unit.Buffs = object.Flags(0x1180204), 4765, 0
			unit.PosVec = types.Ptf(5075.333, 4837.6665)
			unit.NewPos, unit.PrevPos, unit.SpeedCur = unit.PosVec, unit.PosVec, 10
			*update = server.MonsterUpdateData{
				Aggression: 0.5, StatusFlags: object.MonsterStatus(0x1), AIStackInd: 2,
				Field124: 980, Field125: math.Float32bits(unit.PosVec.X), Field126: math.Float32bits(unit.PosVec.Y),
				Field363: frame + 1, // no earlier inversion cast in this isolated tick
			}
			for i, action := range []ai.ActionType{ai.ACTION_GUARD, ai.DEPENDENCY_NOT_UNDER_ATTACK, ai.ACTION_MOVE_TO} {
				update.AIStack[i].Action = uint32(action)
			}
			beforeUnit, beforeUpdate := *unit, *update
			otherIndex := srv.Rand.Other.Index()
			expected := prand.New(1496)
			chance, duration := expected.Int(0, 100), expected.Int(15, 60)
			if chance < 33 {
				t.Fatalf("captured Logic index no longer selects the original WAIT branch: chance=%d", chance)
			}
			t.Logf("captured MainAI branch: frame=%d unit=%p update=%p Logic=%d chance=%d duration=%d",
				frame, unit, update, srv.Rand.Logic.Index(), chance, duration)
			Nox_xxx_monsterMainAIFn_547210(unit)
			if *unit != beforeUnit || unit.HealthData.Cur != 100 || srv.Rand.Other.Index() != otherIndex {
				t.Fatal("movement bookkeeping changed the unit, health or cosmetic RNG")
			}
			if frame == 995 {
				if *update != beforeUpdate || srv.AI.StackChanged || srv.Rand.Logic.Index() != 1496 {
					t.Fatalf("half-FPS equality consumed RNG or scheduled an action: Logic=%d stack=%+v", srv.Rand.Logic.Index(), update.GetAIStack())
				}
				return
			}
			if update.AIStackInd != 3 || update.AIStackHead().Type() != ai.ACTION_WAIT ||
				update.AIStackHead().Args != [4]uintptr{uintptr(frame + uint32(duration))} ||
				update.StatusFlags != beforeUpdate.StatusFlags|object.MonStatusFrustrated || !srv.AI.StackChanged ||
				update.Field124 != frame || update.Field125 != beforeUpdate.Field125 || update.Field126 != beforeUpdate.Field126 ||
				srv.Rand.Logic.Index() != expected.Index() {
				t.Fatalf("original stalled WAIT/RNG missing: Logic=%d want=%d status=%#x progress=%d stack=%+v",
					srv.Rand.Logic.Index(), expected.Index(), update.StatusFlags, update.Field124, update.GetAIStack())
			}
			for i := 0; i < 3; i++ {
				if update.AIStack[i] != beforeUpdate.AIStack[i] {
					t.Fatalf("WAIT replaced captured underlying action %d", i)
				}
			}
			// A following MainAI tick with WAIT at the head must not repeat the
			// movement timeout or consume a compensating random draw.
			srv.SetFrame(frame + 1)
			srv.AI.StackChanged = false
			waiting := *update
			Nox_xxx_monsterMainAIFn_547210(unit)
			if *update != waiting || *unit != beforeUnit || srv.AI.StackChanged ||
				srv.Rand.Logic.Index() != expected.Index() || srv.Rand.Other.Index() != otherIndex {
				t.Fatal("WAIT tick repeated the stall RNG or changed captured state")
			}
		})
	}
}
