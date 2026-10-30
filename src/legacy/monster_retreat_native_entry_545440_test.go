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

// The original AI dispatcher ignores update EAX. Observe the actual public
// C entry as a void update, retaining complete native records after the
// ordinary dispatcher control, with real native stack-pop services.
func TestMonsterRetreatNextPublicCNative545440(t *testing.T) {
	const child = "NOX_MONSTER_RETREAT_NATIVE_CHILD"
	if os.Getenv(child) == "" {
		bin, err := os.Executable()
		if err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command(bin, "-test.run=^TestMonsterRetreatNextPublicCNative545440$", "-test.v")
		cmd.Env = append(os.Environ(), child+"=pop")
		out, err := cmd.CombinedOutput()
		t.Logf("%s", out)
		if err != nil {
			t.Fatalf("actual native public Retreat path: %v", err)
		}
		return
	}
	s, unit, update, missile := monsterMainBlockLegacyFixture547210(t)
	update.AIStackInd, update.ResumeLevel = 1, 0.5
	update.AIStack[0] = server.AIStackItem{Action: uint32(ai.ACTION_IDLE)}
	update.AIStack[1] = server.AIStackItem{Action: uint32(ai.ACTION_RETREAT), Args: [4]uintptr{71, 72, 73, 74}}
	update.StatusFlags, update.CurrentEnemy = 0, nil
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, ptr := range []unsafe.Pointer{unit.CObj(), unsafe.Pointer(update), unsafe.Pointer(unit.HealthData)} {
			if uintptr(ptr) <= 0xffffffff {
				t.Fatal("actual native unit/update/health must exceed4GiB")
			}
		}
	}
	beforeUnit, beforeUpdate, beforeHealth, beforeMissile := *unit, *update, *unit.HealthData, *missile
	s.AI.StackChanged = false
	action, ok := server.GetAIAction(ai.ACTION_RETREAT).(cgoAIAction)
	if !ok {
		t.Fatal("ordinary Retreat dispatcher")
	}
	action.Update(unit)
	if update.AIStackInd != 0 || !s.AI.StackChanged {
		t.Fatal("healthy ordinary Retreat did not pop its real native stack")
	}
	wantUnit, wantUpdate, wantHealth := *unit, *update, *unit.HealthData
	t.Logf("RETREAT CONTROL COMPLETED: unit=%p update=%p health=%p real-stack-pop=true", unit, update, unit.HealthData)
	*unit, *update, *unit.HealthData = beforeUnit, beforeUpdate, beforeHealth
	s.AI.StackChanged = false
	ccall.CallVoidPtr(monsterRetreatPublicC545440(), unit.CObj())
	if !s.AI.StackChanged {
		t.Fatal("native C update did not report the real stack change")
	}
	for _, pair := range []struct {
		got, want unsafe.Pointer
		size      uintptr
	}{
		{unit.CObj(), unsafe.Pointer(&wantUnit), unsafe.Sizeof(wantUnit)},
		{unsafe.Pointer(update), unsafe.Pointer(&wantUpdate), unsafe.Sizeof(wantUpdate)},
		{unsafe.Pointer(unit.HealthData), unsafe.Pointer(&wantHealth), unsafe.Sizeof(wantHealth)},
		{missile.CObj(), unsafe.Pointer(&beforeMissile), unsafe.Sizeof(beforeMissile)},
	} {
		if !bytes.Equal(unsafe.Slice((*byte)(pair.got), int(pair.size)), unsafe.Slice((*byte)(pair.want), int(pair.size))) {
			t.Fatal("whole native unit/update/health/missile records differ")
		}
	}
	t.Log("RETREAT NATIVE C PASS: void-update-contract=true real-stack-pop=true full-records=true")
}
