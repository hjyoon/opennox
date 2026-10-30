package legacy

import (
	"bytes"
	"os"
	"os/exec"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
	"github.com/opennox/opennox/v1/server"
)

func TestMonsterDeadUpdateNextPublicCNative544EC0(t *testing.T) {
	const child = "NOX_MONSTER_DEAD_UPDATE_NATIVE_CHILD"
	if os.Getenv(child) == "" {
		for _, mode := range []string{"update", "cleanup"} {
			t.Run(mode, func(t *testing.T) {
				bin, err := os.Executable()
				if err != nil {
					t.Fatal(err)
				}
				cmd := exec.Command(bin, "-test.run=^TestMonsterDeadUpdateNextPublicCNative544EC0$", "-test.v")
				cmd.Env = append(os.Environ(), child+"="+mode)
				out, err := cmd.CombinedOutput()
				t.Logf("%s", out)
				if err != nil {
					t.Fatalf("actual public C DeadUpdate/Cleanup: %v", err)
				}
			})
		}
		return
	}
	s, unit, update, missile := monsterMainBlockLegacyFixture547210(t)
	view := s.TypePredicateRuntime534A10()
	view.Write(1, 100)
	view.Write(2, 200)
	unit.TypeInd = 300
	definition, free := alloc.New(server.MonsterDef{})
	defer free()
	*definition = server.MonsterDef{}
	waypoints, freeWaypoints := alloc.Make([]server.Waypoint{}, 16)
	defer freeWaypoints()
	update.MonsterDef = definition
	update.Field74 = 2
	update.Field282_1 = 3
	update.Field91 = 0x12345678
	update.Field523_2, update.Field523_3 = 0x34, 0xd5
	update.Stamina, update.Field282_2 = 0xa5, 0xa5a5
	update.CurrentEnemy = missile
	unit.Update = unit.CObj()
	for i := range update.Waypoints {
		update.Waypoints[i] = &waypoints[i]
		update.SeenEnemies[i] = missile
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unit.CObj()) <= 0xffffffff || uintptr(unit.UpdateData) <= 0xffffffff || uintptr(unsafe.Pointer(definition)) <= 0xffffffff || uintptr(unsafe.Pointer(&waypoints[0])) <= 0xffffffff) {
		t.Fatal("native links must exceed4GiB")
	}
	a, ok := server.GetAIAction(ai.ACTION_DEAD).(cgoAIAction)
	if !ok {
		t.Fatal("ordinary dead dispatcher")
	}
	beforeUnit, beforeUpdate, beforeDefinition, beforeMissile := *unit, *update, *definition, *missile
	a.Update(unit)
	expectedUnit, expectedUpdate := *unit, *update
	if unit.Update != nil || update.Field74 != 0 || update.Field282_1 != 0 || update.Field91 != 0 || update.Field523_2 != 0 || update.CurrentEnemy != nil {
		t.Fatal("ordinary cleanup did not complete")
	}
	for i := range update.Waypoints {
		if (update.Waypoints[i] == nil) != (i < 2) || (update.SeenEnemies[i] == nil) != (i < 3) {
			t.Fatal("ordinary counted prefixes")
		}
	}
	t.Logf("DEAD UPDATE ORDINARY COMPLETED: mode=%s unit=%p update=%p counted-prefixes=true", os.Getenv(child), unit, update)
	*unit, *update = beforeUnit, beforeUpdate
	if os.Getenv(child) == "cleanup" {
		expectedUnit = beforeUnit
		got := ccall.CallUPtrPtr(monsterDeadCleanupPublicC544F70(), unit.CObj())
		if got != uintptr(unsafe.Pointer(update)) {
			t.Fatalf("native cleanup result=%x want=%p", got, update)
		}
	} else {
		ccall.CallVoidPtr(monsterDeadUpdatePublicC544EC0(), unit.CObj())
	}
	raw := func(v unsafe.Pointer, n uintptr) []byte { return unsafe.Slice((*byte)(v), int(n)) }
	if !bytes.Equal(raw(unit.CObj(), unsafe.Sizeof(*unit)), raw(unsafe.Pointer(&expectedUnit), unsafe.Sizeof(expectedUnit))) || !bytes.Equal(raw(unsafe.Pointer(update), unsafe.Sizeof(*update)), raw(unsafe.Pointer(&expectedUpdate), unsafe.Sizeof(expectedUpdate))) || !bytes.Equal(raw(unsafe.Pointer(definition), unsafe.Sizeof(*definition)), raw(unsafe.Pointer(&beforeDefinition), unsafe.Sizeof(beforeDefinition))) || !bytes.Equal(raw(missile.CObj(), unsafe.Sizeof(*missile)), raw(unsafe.Pointer(&beforeMissile), unsafe.Sizeof(beforeMissile))) {
		t.Fatal("full native records changed beyond original cleanup")
	}
	t.Logf("DEAD UPDATE NATIVE C PASS: mode=%s full-records=true pointer-return=true", os.Getenv(child))
}
