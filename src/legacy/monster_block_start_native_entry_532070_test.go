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

func TestMonsterBlockStartNextPublicCNative532070(t *testing.T) {
	const child = "NOX_MONSTER_BLOCK_START_NATIVE_CHILD"
	if os.Getenv(child) == "" {
		binary, err := os.Executable()
		if err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command(binary, "-test.run=^TestMonsterBlockStartNextPublicCNative532070$", "-test.v")
		cmd.Env = append(os.Environ(), child+"=1")
		out, err := cmd.CombinedOutput()
		t.Logf("%s", out)
		if err != nil {
			t.Fatalf("actual block-start public C path: %v", err)
		}
		return
	}
	s, unit, update, _ := monsterMainBlockLegacyFixture547210(t)
	s.SetFrame(150)
	update.AIStackInd = 1
	update.AIStack[0].Action = uint32(ai.ACTION_IDLE)
	update.AIStack[1].Action = uint32(ai.ACTION_BLOCK_ATTACK)
	update.AIStack[1].Args[0] = 200
	if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unit.CObj()) <= 0xffffffff || uintptr(unit.UpdateData) <= 0xffffffff) {
		t.Fatal("native C-owned links must exceed4GiB")
	}
	if Nox_xxx_monsterTestBlockShield_533E70(unit) != 0 {
		t.Fatal("unindexed missile must not refresh the deadline")
	}
	a, ok := server.GetAIAction(ai.ACTION_BLOCK_ATTACK).(cgoAIAction)
	if !ok {
		t.Fatal("ordinary block dispatcher")
	}
	beforeUnit, beforeUpdate := *unit, *update
	a.Update(unit)
	if update.AIStackInd != 1 || update.AIStack[1].Args[0] != 200 {
		t.Fatal("ordinary block-start control changed future deadline")
	}
	expectedUnit, expectedUpdate := *unit, *update
	t.Logf("BLOCK START ORDINARY COMPLETED: unit=%p update=%p frame=150 deadline=200 real-shield-probe=0", unit, update)
	*unit, *update = beforeUnit, beforeUpdate
	got := uint8(ccall.CallIntPtr(monsterBlockStartPublicC532070(), unit.CObj()))
	rawUnit := func(v *server.Object) []byte { return unsafe.Slice((*byte)(unsafe.Pointer(v)), int(unsafe.Sizeof(*v))) }
	rawUpdate := func(v *server.MonsterUpdateData) []byte {
		return unsafe.Slice((*byte)(unsafe.Pointer(v)), int(unsafe.Sizeof(*v)))
	}
	if got != 200 || !bytes.Equal(rawUnit(unit), rawUnit(&expectedUnit)) || !bytes.Equal(rawUpdate(update), rawUpdate(&expectedUpdate)) {
		t.Fatalf("actual C block-start differs from ordinary native control: char=%d want200", got)
	}
}
