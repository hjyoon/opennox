package legacy

import (
	"bytes"
	"os"
	"os/exec"
	"strconv"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
	"github.com/opennox/opennox/v1/server"
)

func TestMonsterSearchEdibleNativeLayout544A00(t *testing.T) {
	TestMonsterDodgeablesNativeLayout547C50(t)
	unit, update, item := server.Object{}, server.MonsterUpdateData{}, server.AIStackItem{}
	prefix := unsafe.Offsetof(unit.ScriptPickup) + unsafe.Sizeof(unit.ScriptPickup)
	want := [11]uintptr{unsafe.Offsetof(unit.ObjClass), unsafe.Offsetof(unit.ObjSubClass), unsafe.Offsetof(unit.PosVec), unsafe.Offsetof(unit.Poison540), unsafe.Offsetof(unit.UpdateData), prefix, unsafe.Offsetof(update.AIStackInd), unsafe.Offsetof(update.AIStack), unsafe.Sizeof(item), unsafe.Offsetof(item.Args), unsafe.Offsetof(item.Field5)}
	if got := edibleOriginalLayout544A00(); got != want {
		t.Fatal("independent native object/MUD/stack layouts", got, want)
	}
}

func TestMonsterSearchEdibleActualCBefore544A00(t *testing.T) {
	if os.Getenv("OPENNOX_EDIBLES_C_BEFORE") != "1" {
		return
	}
	if raw := os.Getenv("OPENNOX_EDIBLES_C_CHILD"); raw != "" {
		entry, err := strconv.Atoi(raw)
		if err != nil || entry < 0 || entry > 2 {
			t.Fatal("invalid child", raw)
		}
		TestMonsterSearchEdibleNativeLayout544A00(t)
		s, unit, update, owner := monsterMainBlockLegacyFixture547210(t)
		food, freeFood := alloc.New(server.Object{})
		t.Cleanup(freeFood)
		*food = server.Object{ObjClass: object.ClassFood, ObjFlags: object.FlagActive | object.FlagEnabled, PosVec: types.Ptf(320, 300), NewPos: types.Ptf(320, 300)}
		food.Shape.Kind = server.ShapeKindCenter
		s.Map.AddObjectToIndex(food)
		t.Cleanup(func() { s.Map.RemoveObjectFromIndex(food) })
		if got := s.MonsterSearchEdible544A00(unit, 250); got != food {
			t.Fatalf("real native map/interact search=%p want%p", got, food)
		}
		unit.HealthData.Cur, unit.HealthData.Max = 10, 100
		update.ResumeLevel, update.CurrentEnemy = 0.5, nil
		update.AIStackInd, update.AIStack[0].Action = 0, uint32(ai.ACTION_RETREAT)
		s.MonsterActionRetreat545440(unit)
		wantActions := []ai.ActionType{ai.DEPENDENCY_NOT_HEALTHY, ai.DEPENDENCY_NO_VISIBLE_ENEMY, ai.DEPENDENCY_OBJECT_AT_VISIBLE_LOCATION, ai.ACTION_PICKUP_OBJECT, ai.ACTION_MOVE_TO}
		if update.AIStackInd != 5 {
			t.Fatal("ordinary real edible/Push composition head", update.AIStackInd)
		}
		for i, action := range wantActions {
			if update.AIStack[i+1].Type() != action {
				t.Fatal("ordinary real edible/Push order", i, update.AIStack[i+1].Type(), action)
			}
		}
		if update.AIStack[3].ArgObj(2) != food || update.AIStack[4].ArgObj(0) != food || update.AIStack[5].ArgObj(2) != food {
			t.Fatal("ordinary native food identity truncated")
		}
		update.SoundSet122 = unsafe.Pointer(owner)
		records := []struct {
			ptr  unsafe.Pointer
			size uintptr
		}{{unsafe.Pointer(unit), unsafe.Sizeof(*unit)}, {unsafe.Pointer(update), unsafe.Sizeof(*update)}, {unsafe.Pointer(owner), unsafe.Sizeof(*owner)}, {unsafe.Pointer(food), unsafe.Sizeof(*food)}, {unsafe.Pointer(unit.HealthData), unsafe.Sizeof(*unit.HealthData)}}
		before := make([][]byte, len(records))
		for i, r := range records {
			before[i] = bytes.Clone(unsafe.Slice((*byte)(r.ptr), int(r.size)))
		}
		set, action := dodgeOriginalControls547C50(unit)
		if set != uintptr(unsafe.Pointer(owner)) || action != uint32(ai.ACTION_MOVE_TO) || retreatPredicatesOriginalC545520(0, unit) != 0 || ccall.CallIntPtr(statePredicateOriginalPublicC534710(0x5347f0), unit.CObj()) != 0 {
			t.Fatal("typed sound/action/health/poison positive controls")
		}
		for i, r := range records {
			if !bytes.Equal(before[i], unsafe.Slice((*byte)(r.ptr), int(r.size))) {
				t.Fatal("whole initial native control record changed", i)
			}
		}
		if unsafe.Sizeof(uintptr(0)) != 8 {
			t.Fatal("high native-pointer before qualification requires64bit")
		}
		for _, r := range records {
			if uintptr(r.ptr) <= 0xffffffff {
				t.Fatal("control pointers must exceed4GiB")
			}
		}
		if edibleOriginalABI544A00(entry) != 32 {
			t.Fatal("unchanged legacy ABI required")
		}
		t.Logf("EDIBLES_REAL_NATIVE_CONTROLS entry=%d unit=%p update=%p food=%p owner=%p ordinary-search-and-five-Push=true full-records=true", entry, unit, update, food, owner)
		edibleOriginalC544A00(entry, unit, food, 250)
		t.Fatal("unported public C unexpectedly returned")
	}
	for entry := 0; entry < 3; entry++ {
		t.Run(strconv.Itoa(entry), func(t *testing.T) {
			cmd := exec.Command(os.Args[0], "-test.v", "-test.run=^TestMonsterSearchEdibleActualCBefore544A00$")
			cmd.Env = append(os.Environ(), "OPENNOX_EDIBLES_C_CHILD="+strconv.Itoa(entry))
			out, err := cmd.CombinedOutput()
			t.Logf("%s", out)
			if err == nil || !bytes.Contains(out, []byte("EDIBLES_REAL_NATIVE_CONTROLS")) || !bytes.Contains(out, []byte("SIGSEGV")) || !bytes.Contains(out, []byte("_Cfunc_edible_original_public_544a00")) {
				t.Fatal("actual C must fault only after independent valid native controls", err)
			}
		})
	}
}
