package legacy

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestMonsterRoamWaypointLifecycleNativeProduction545790(t *testing.T) {
	const child = "OPENNOX_ROAM_LIFECYCLE_PRODUCTION_CHILD"
	mode := os.Getenv(child)
	if mode == "" {
		for entry := 0; entry < 2; entry++ {
			for _, index := range []int{0, 11, 23} {
				for selected := 0; selected < 2; selected++ {
					mode := fmt.Sprintf("%d:%d:%d", entry, index, selected)
					t.Run(mode, func(t *testing.T) {
						cmd := exec.Command(os.Args[0], "-test.v", "-test.run=^TestMonsterRoamWaypointLifecycleNativeProduction545790$")
						cmd.Env = append(os.Environ(), child+"="+mode)
						out, err := cmd.CombinedOutput()
						t.Logf("%s", out)
						if err != nil || !bytes.Contains(out, []byte("ROAM_LIFECYCLE_REAL_PRODUCTION_PASS")) {
							t.Fatal("actual C and registered ordinary ROAM start/cancel", err)
						}
					})
				}
			}
		}
		return
	}
	parts := strings.Split(mode, ":")
	if len(parts) != 3 {
		t.Fatal("production mode")
	}
	entry, e1 := strconv.Atoi(parts[0])
	index, e2 := strconv.Atoi(parts[1])
	selected, e3 := strconv.Atoi(parts[2])
	if e1 != nil || e2 != nil || e3 != nil || entry < 0 || entry > 1 || (index != 0 && index != 11 && index != 23) || selected < 0 || selected > 1 {
		t.Fatal("production mode values")
	}
	if roamLifecycleOriginalABI545790(entry) != 64 {
		t.Fatal("typed native public ABI required")
	}
	s, unit, first, owner := monsterMainBlockLegacyFixture547210(t)
	second, freeSecond := alloc.New(server.MonsterUpdateData{})
	defer freeSecond()
	waypoints, freeWaypoints := alloc.New([16]server.Waypoint{})
	defer freeWaypoints()
	*second = *first
	updates := [2]*server.MonsterUpdateData{first, second}
	for n, update := range updates {
		update.Field91 = 0x12340000 + uint32(n)
		update.AIStackInd = int8(index)
		for i := range update.Waypoints {
			update.Waypoints[i] = &waypoints[i]
		}
		for i := range update.AIStack {
			update.AIStack[i].Action = uint32(ai.ACTION_ROAM)
			update.AIStack[i].Field5 = 0xa5000000 + uint32(n*256+i)
			for j := range update.AIStack[i].Args {
				update.AIStack[i].Args[j] = uintptr(unsafe.Pointer(owner)) + uintptr(i*8+j)
			}
		}
	}
	unit.UpdateData = unsafe.Pointer(updates[selected])
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, ptr := range []unsafe.Pointer{unsafe.Pointer(unit), unsafe.Pointer(first), unsafe.Pointer(second), unsafe.Pointer(owner), unsafe.Pointer(waypoints)} {
			if uintptr(ptr) <= 0xffffffff {
				t.Fatal("native production pointers must exceed4GiB")
			}
		}
	}
	beforeUnit, beforeOwner, beforeHealth := *unit, *owner, *unit.HealthData
	beforeUpdates, beforeWaypoints := [2]server.MonsterUpdateData{*first, *second}, *waypoints
	wantUpdates := beforeUpdates
	if entry == 0 {
		wantUpdates[selected].Waypoints[0] = nil
		wantUpdates[selected].Field91 = 0
	} else {
		wantUpdates[selected].AIStack[index].Args[0] = 0
	}
	logic, other, frame, changed := s.Rand.Logic.Index(), s.Rand.Other.Index(), s.Frame(), s.AI.StackChanged
	check := func() {
		t.Helper()
		pairs := []struct {
			got, want unsafe.Pointer
			size      uintptr
		}{
			{unsafe.Pointer(unit), unsafe.Pointer(&beforeUnit), unsafe.Sizeof(beforeUnit)},
			{unsafe.Pointer(owner), unsafe.Pointer(&beforeOwner), unsafe.Sizeof(beforeOwner)},
			{unsafe.Pointer(unit.HealthData), unsafe.Pointer(&beforeHealth), unsafe.Sizeof(beforeHealth)},
			{unsafe.Pointer(first), unsafe.Pointer(&wantUpdates[0]), unsafe.Sizeof(wantUpdates[0])},
			{unsafe.Pointer(second), unsafe.Pointer(&wantUpdates[1]), unsafe.Sizeof(wantUpdates[1])},
			{unsafe.Pointer(waypoints), unsafe.Pointer(&beforeWaypoints), unsafe.Sizeof(beforeWaypoints)},
		}
		for i, pair := range pairs {
			if !bytes.Equal(unsafe.Slice((*byte)(pair.got), int(pair.size)), unsafe.Slice((*byte)(pair.want), int(pair.size))) {
				t.Fatal("whole native production record changed outside original stores", i)
			}
		}
		if s.Frame() != frame || s.Rand.Logic.Index() != logic || s.Rand.Other.Index() != other || s.AI.StackChanged != changed {
			t.Fatal("leaf operation changed frame, RNG or stack-change state")
		}
	}
	if result := roamLifecycleOriginalC545790(entry, unit); result != unsafe.Pointer(updates[selected]) {
		t.Fatal("complete native update-pointer return", result, updates[selected])
	}
	check()
	*first, *second = beforeUpdates[0], beforeUpdates[1]
	if entry == 0 {
		(cgoAIAction{typ: ai.ACTION_ROAM}).Start(unit)
	} else {
		(cgoAIAction{typ: ai.ACTION_ROAM}).Cancel(unit)
	}
	check()
	t.Logf("ROAM_LIFECYCLE_REAL_PRODUCTION_PASS mode=%s unit=%p update=%p whole-records=true native-return=true ordinary-registered-action=true", mode, unit, updates[selected])
}
