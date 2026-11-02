package legacy

import (
	"bytes"
	"github.com/opennox/opennox/v1/server"
	"os"
	"os/exec"
	"strconv"
	"testing"
	"unsafe"
)

func TestMonsterRoamWaypointLifecycleNativeLayout545790(t *testing.T) {
	TestMonsterDodgeablesNativeLayout547C50(t)
	unit, update, item := server.Object{}, server.MonsterUpdateData{}, server.AIStackItem{}
	prefix := unsafe.Offsetof(unit.ScriptPickup) + unsafe.Sizeof(unit.ScriptPickup)
	want := [8]uintptr{unsafe.Offsetof(unit.UpdateData), prefix, unsafe.Offsetof(update.Waypoints), unsafe.Offsetof(update.Field91), unsafe.Offsetof(update.AIStackInd), unsafe.Offsetof(update.AIStack), unsafe.Sizeof(item), unsafe.Offsetof(item.Args)}
	if got := roamLifecycleOriginalLayout545790(); got != want {
		t.Fatal("independent C/Go object, MUD and native stack layouts", got, want)
	}
}

func TestMonsterRoamWaypointLifecycleActualCBefore545790(t *testing.T) {
	if os.Getenv("OPENNOX_ROAM_LIFECYCLE_C_BEFORE") != "1" {
		return
	}
	if raw := os.Getenv("OPENNOX_ROAM_LIFECYCLE_C_CHILD"); raw != "" {
		entry, err := strconv.Atoi(raw)
		if err != nil || entry < 0 || entry > 1 {
			t.Fatal("invalid child", raw)
		}
		TestMonsterRoamWaypointLifecycleNativeLayout545790(t)
		TestMonsterRetreatMasterNativeLayout5456D0(t)
		s, unit, update, owner := monsterMainBlockLegacyFixture547210(t)
		update.Waypoints[0], update.Field91 = (*server.Waypoint)(unsafe.Pointer(owner)), 19
		s.MonsterActionRoamStart545790(unit)
		if update.Waypoints[0] != nil || update.Field91 != 0 {
			t.Fatal("ordinary native start control")
		}
		update.AIStackInd, update.AIStack[0].Args[0] = 0, uintptr(unsafe.Pointer(owner))
		s.MonsterActionRoamCancel5457C0(unit)
		if update.AIStack[0].Args[0] != 0 {
			t.Fatal("ordinary native cancel control")
		}
		update.SoundSet122 = unsafe.Pointer(owner)
		records := []struct {
			ptr  unsafe.Pointer
			size uintptr
		}{
			{unsafe.Pointer(unit), unsafe.Sizeof(*unit)},
			{unsafe.Pointer(update), unsafe.Sizeof(*update)},
			{unsafe.Pointer(owner), unsafe.Sizeof(*owner)},
			{unsafe.Pointer(unit.HealthData), unsafe.Sizeof(*unit.HealthData)},
		}
		before := make([][]byte, len(records))
		for i, record := range records {
			before[i] = bytes.Clone(unsafe.Slice((*byte)(record.ptr), int(record.size)))
		}
		set, action := dodgeOriginalControls547C50(unit)
		if set != uintptr(unsafe.Pointer(owner)) || action != update.AIStack[0].Action || retreatPredicatesOriginalC545520(0, unit) != 1 {
			t.Fatal("actual typed sound/action/predicate controls")
		}
		for i, record := range records {
			if !bytes.Equal(before[i], unsafe.Slice((*byte)(record.ptr), int(record.size))) {
				t.Fatal("complete initial control record changed", i)
			}
		}
		if unsafe.Sizeof(uintptr(0)) != 8 || uintptr(unsafe.Pointer(unit)) <= 0xffffffff || uintptr(unsafe.Pointer(update)) <= 0xffffffff || uintptr(unsafe.Pointer(owner)) <= 0xffffffff {
			t.Fatal("high native control pointers required")
		}
		if roamLifecycleOriginalABI545790(entry) != 32 {
			t.Fatal("unchanged original legacy-int ABI required")
		}
		t.Logf("ROAM_LIFECYCLE_VALID_NATIVE_CONTROLS entry=%d unit=%p update=%p owner=%p", entry, unit, update, owner)
		roamLifecycleOriginalC545790(entry, unit)
		t.Fatal("old C entry unexpectedly returned")
	}
	for entry := 0; entry < 2; entry++ {
		t.Run(strconv.Itoa(entry), func(t *testing.T) {
			cmd := exec.Command(os.Args[0], "-test.run=^TestMonsterRoamWaypointLifecycleActualCBefore545790$", "-test.v")
			cmd.Env = append(os.Environ(), "OPENNOX_ROAM_LIFECYCLE_C_CHILD="+strconv.Itoa(entry))
			out, err := cmd.CombinedOutput()
			t.Logf("%s", out)
			if err == nil || !bytes.Contains(out, []byte("ROAM_LIFECYCLE_VALID_NATIVE_CONTROLS")) || !bytes.Contains(out, []byte("SIGSEGV")) || !bytes.Contains(out, []byte("_Cfunc_roam_lifecycle_original_545790")) {
				t.Fatal("actual old C must fault after independent positive controls", err)
			}
		})
	}
}
