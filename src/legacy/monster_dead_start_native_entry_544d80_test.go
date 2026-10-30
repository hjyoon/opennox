package legacy

import (
	"bytes"
	"fmt"
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

func TestMonsterDeadStartNextPublicCNative544D80(t *testing.T) {
	const child = "NOX_MONSTER_DEAD_START_NATIVE_CHILD"
	if os.Getenv(child) == "" {
		for _, low := range []uint32{0, 128, 255} {
			t.Run(fmt.Sprint(low), func(t *testing.T) {
				bin, err := os.Executable()
				if err != nil {
					t.Fatal(err)
				}
				cmd := exec.Command(bin, "-test.run=^TestMonsterDeadStartNextPublicCNative544D80$", "-test.v")
				cmd.Env = append(os.Environ(), child+"="+fmt.Sprint(low))
				out, err := cmd.CombinedOutput()
				t.Logf("%s", out)
				if err != nil {
					t.Fatalf("actualC dead-start path: %v", err)
				}
			})
		}
		return
	}
	low, err := strconv.ParseUint(os.Getenv(child), 10, 8)
	if err != nil {
		t.Fatal(err)
	}
	s, unit, update, _ := monsterMainBlockLegacyFixture547210(t)
	view := s.TypePredicateRuntime534A10()
	view.Write(1, 100)
	view.Write(2, 200)
	unit.TypeInd = 300
	def, freeDef := alloc.New(server.MonsterDef{})
	defer freeDef()
	*def = server.MonsterDef{}
	update.MonsterDef = def
	unit.Field131 = 0
	unit.ObjSubClass = 0
	unit.ObjFlags = object.Flags(0x12000000 | uint32(low))
	unit.VelVec, unit.ForceVec, unit.Pos24 = types.Ptf(1, 2), types.Ptf(3, 4), types.Ptf(5, 6)
	if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unit.CObj()) <= 0xffffffff || uintptr(unit.UpdateData) <= 0xffffffff || uintptr(unsafe.Pointer(def)) <= 0xffffffff) {
		t.Fatal("native object/update/definition links must exceed4GiB")
	}
	a, ok := server.GetAIAction(ai.ACTION_DEAD).(cgoAIAction)
	if !ok {
		t.Fatal("ordinary dead dispatcher")
	}
	beforeUnit, beforeUpdate := *unit, *update
	a.Start(unit)
	if unit.VelVec != (types.Pointf{}) || unit.ForceVec != (types.Pointf{}) || unit.Pos24 != (types.Pointf{}) || unit.ObjFlags != beforeUnit.ObjFlags|object.FlagShort|object.FlagAllowOverlap {
		t.Fatal("ordinary nonzombie start motion/flags")
	}
	expectedUnit, expectedUpdate, expectedDef := *unit, *update, *def
	t.Logf("DEAD START ORDINARY COMPLETED: low=%d unit=%p update=%p def=%p zero-motion=true flags=%x", low, unit, update, def, uint32(unit.ObjFlags))
	*unit, *update = beforeUnit, beforeUpdate
	got := uint8(ccall.CallIntPtr(monsterDeadStartPublicC544D80(), unit.CObj()))
	rawUnit := func(v *server.Object) []byte { return unsafe.Slice((*byte)(unsafe.Pointer(v)), int(unsafe.Sizeof(*v))) }
	rawUpdate := func(v *server.MonsterUpdateData) []byte {
		return unsafe.Slice((*byte)(unsafe.Pointer(v)), int(unsafe.Sizeof(*v)))
	}
	rawDef := func(v *server.MonsterDef) []byte {
		return unsafe.Slice((*byte)(unsafe.Pointer(v)), int(unsafe.Sizeof(*v)))
	}
	if got != uint8(expectedUnit.ObjFlags) || !bytes.Equal(rawUnit(unit), rawUnit(&expectedUnit)) || !bytes.Equal(rawUpdate(update), rawUpdate(&expectedUpdate)) || !bytes.Equal(rawDef(def), rawDef(&expectedDef)) {
		t.Fatalf("actualC nonzombie dead-start diverged:char=%d want%d", got, uint8(expectedUnit.ObjFlags))
	}
	t.Logf("DEAD START NATIVE C PASS: low=%d char=%d zero-motion=true full-unit-update-def=true", low, got)
}
