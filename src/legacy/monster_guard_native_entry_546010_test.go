package legacy

import (
	"bytes"
	"os"
	"os/exec"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
	"github.com/opennox/opennox/v1/server"
)

func TestMonsterGuardNextPublicCNative546010(t *testing.T) {
	const child = "NOX_MONSTER_GUARD_NATIVE_CHILD"
	if os.Getenv(child) == "" {
		bin, err := os.Executable()
		if err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command(bin, "-test.run=^TestMonsterGuardNextPublicCNative546010$", "-test.v")
		cmd.Env = append(os.Environ(), child+"=shopkeeper")
		out, err := cmd.CombinedOutput()
		t.Logf("%s", out)
		if err != nil {
			t.Fatalf("actual native public Guard path: %v", err)
		}
		return
	}
	s, unit, update, missile := monsterMainBlockLegacyFixture547210(t)
	s.SetFrame(1401)
	unit.NetCode, unit.Buffs = 0, 1<<29
	unit.PrevPos = unit.PosVec
	update.Aggression, update.CurrentEnemy = 0.5, nil
	update.AIStackInd = 0
	update.AIStack[0].Action = uint32(ai.ACTION_GUARD)
	update.AIStack[0].SetArgs(unit.PosVec, uint32(0))
	beforeUnit, beforeUpdate, beforeHealth, beforeMissile := *unit, *update, *unit.HealthData, *missile
	s.AI.StackChanged = false
	action, ok := server.GetAIAction(ai.ACTION_GUARD).(cgoAIAction)
	if !ok {
		t.Fatal("ordinary Guard dispatcher")
	}
	action.Update(unit)
	if *unit != beforeUnit || *update != beforeUpdate || s.AI.StackChanged {
		t.Fatal("ordinary shopkeeper Guard changed its records")
	}
	t.Logf("GUARD CONTROL COMPLETED: unit=%p update=%p native-noop=true", unit, update)
	ccall.CallVoidPtr(monsterGuardPublicC546010(), unit.CObj())
	for _, pair := range []struct {
		got, want unsafe.Pointer
		size      uintptr
	}{
		{unit.CObj(), unsafe.Pointer(&beforeUnit), unsafe.Sizeof(beforeUnit)},
		{unsafe.Pointer(update), unsafe.Pointer(&beforeUpdate), unsafe.Sizeof(beforeUpdate)},
		{unsafe.Pointer(unit.HealthData), unsafe.Pointer(&beforeHealth), unsafe.Sizeof(beforeHealth)},
		{missile.CObj(), unsafe.Pointer(&beforeMissile), unsafe.Sizeof(beforeMissile)},
	} {
		if !bytes.Equal(unsafe.Slice((*byte)(pair.got), int(pair.size)), unsafe.Slice((*byte)(pair.want), int(pair.size))) {
			t.Fatal("whole native Guard records differ")
		}
	}
	if s.AI.StackChanged {
		t.Fatal("native shopkeeper Guard changed stack")
	}
	t.Log("GUARD NATIVE C PASS: void-update-contract=true full-records=true")
}
