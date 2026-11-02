package legacy

import (
	"bytes"
	"os"
	"os/exec"
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/server"
)

func TestMonsterRetreatMasterNativeLayout5456D0(t *testing.T) {
	unit, update, item := server.Object{}, server.MonsterUpdateData{}, server.AIStackItem{}
	// The shared C object ends at ScriptPickup; Go appends two EXT handles.
	// Keep the existing independent EXT-layout guard as well as every field.
	TestMonsterDodgeablesNativeLayout547C50(t)
	prefix := unsafe.Offsetof(unit.ScriptPickup) + unsafe.Sizeof(unit.ScriptPickup)
	want := [11]uintptr{unsafe.Offsetof(unit.ObjOwner), unsafe.Offsetof(unit.UpdateData), unsafe.Offsetof(unit.PosVec), unsafe.Offsetof(unit.HealthData), prefix, unsafe.Offsetof(update.Field329), unsafe.Offsetof(update.ResumeLevel), unsafe.Offsetof(update.StatusFlags), unsafe.Sizeof(item), unsafe.Offsetof(item.Args), unsafe.Offsetof(item.Field5)}
	if got := retreatMasterOriginalLayout5456D0(); got != want {
		t.Fatal("independent native C/Go object/MUD/action layout", got, want)
	}
}

func TestMonsterRetreatMasterNativeCBefore5456D0(t *testing.T) {
	if os.Getenv("OPENNOX_RETREAT_MASTER_C_BEFORE") != "1" {
		return
	}
	if os.Getenv("OPENNOX_RETREAT_MASTER_C_CHILD") != "1" {
		cmd := exec.Command(os.Args[0], "-test.run=^TestMonsterRetreatMasterNativeCBefore5456D0$", "-test.v")
		cmd.Env = append(os.Environ(), "OPENNOX_RETREAT_MASTER_C_CHILD=1")
		out, err := cmd.CombinedOutput()
		t.Logf("%s", out)
		if err == nil || !bytes.Contains(out, []byte("RETREAT_MASTER_VALID_REAL_NATIVE_CONTROLS")) || !bytes.Contains(out, []byte("SIGSEGV")) || !bytes.Contains(out, []byte("_Cfunc_retreat_master_original_public_5456d0")) {
			t.Fatal("legacy-int public entry must fault only after real native positive controls", err)
		}
		return
	}
	TestMonsterRetreatMasterNativeLayout5456D0(t)
	TestMonsterRetreatPredicatesNativeLayout545520(t)
	s, unit, update, owner := monsterMainBlockLegacyFixture547210(t)
	unit.PosVec, owner.PosVec = types.Ptf(0, 0), types.Ptf(200, 0)
	unit.ObjOwner = owner
	unit.HealthData.Cur, unit.HealthData.Max = 10, 100
	unit.Buffs = 0x20000000
	update.ResumeLevel, update.Field329, update.StatusFlags = 0.5, 20, 0
	update.AIStackInd, update.AIStack[0].Action = 0, uint32(ai.ACTION_RETREAT_TO_MASTER)
	update.SoundSet122 = unsafe.Pointer(owner)
	if !s.MonsterActionRetreatToMaster5456D0(unit) || update.AIStackInd != 2 || update.AIStack[1].Type() != ai.DEPENDENCY_OBJECT_FARTHER_THAN || update.AIStack[2].Type() != ai.ACTION_MOVE_TO {
		t.Fatal("actual ordinary native AI/Push composition did not schedule original actions")
	}
	if update.AIStack[1].ArgObj(2) != owner || update.AIStack[2].ArgObj(2) != owner || unit.HealthData.Cur != 10 || unit.HealthData.Max != 100 {
		t.Fatal("actual native owner identity or health changed")
	}
	pointers := []struct {
		ptr  unsafe.Pointer
		size uintptr
	}{
		{unsafe.Pointer(unit), unsafe.Sizeof(*unit)},
		{unsafe.Pointer(update), unsafe.Sizeof(*update)},
		{unsafe.Pointer(unit.HealthData), unsafe.Sizeof(*unit.HealthData)},
		{unsafe.Pointer(owner), unsafe.Sizeof(*owner)},
	}
	before := make([][]byte, len(pointers))
	for i, item := range pointers {
		before[i] = bytes.Clone(unsafe.Slice((*byte)(item.ptr), int(item.size)))
	}
	set, action := dodgeOriginalControls547C50(unit)
	if set != uintptr(unsafe.Pointer(owner)) || action != uint32(ai.ACTION_MOVE_TO) || unitBuffTestExportCall4FF350(unit, 29) != 1 || retreatPredicatesOriginalC545520(0, unit) != 0 || retreatPredicatesOriginalC545520(1, unit) != 1 || retreatPredicatesOriginalC545520(2, unit) != 0 {
		t.Fatal("actual typed sound/action/buff and all three native predicate positive controls")
	}
	for i, item := range pointers {
		if !bytes.Equal(before[i], unsafe.Slice((*byte)(item.ptr), int(item.size))) {
			t.Fatal("complete original native control record changed", i)
		}
	}
	if retreatMasterOriginalABI5456D0() != 32 {
		t.Fatal("unchanged legacy-int ABI required")
	}
	t.Logf("RETREAT_MASTER_VALID_REAL_NATIVE_CONTROLS unit=%p owner=%p update=%p health=%p nativeOwnerActions=2", unit, owner, update, unit.HealthData)
	retreatMasterOriginalC5456D0(unit)
	t.Fatal("old public C unexpectedly returned")
}
