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

func TestMonsterDyingEndNextPublicCNative544CA0(t *testing.T) {
	const child = "NOX_MONSTER_DYING_END_NATIVE_CHILD"
	if os.Getenv(child) == "" {
		for _, mode := range []string{"ordinary", "zombie", "vile"} {
			t.Run(mode, func(t *testing.T) {
				binary, err := os.Executable()
				if err != nil {
					t.Fatal(err)
				}
				cmd := exec.Command(binary, "-test.run=^TestMonsterDyingEndNextPublicCNative544CA0$", "-test.v")
				cmd.Env = append(os.Environ(), child+"="+mode)
				out, err := cmd.CombinedOutput()
				t.Logf("%s", out)
				if err != nil {
					t.Fatalf("actual dying-end public C path: %v", err)
				}
			})
		}
		return
	}
	s, unit, update, _ := monsterMainBlockLegacyFixture547210(t)
	view := s.TypePredicateRuntime534A10()
	view.Write(1, 100)
	view.Write(2, 200)
	unit.TypeInd = 300
	if os.Getenv(child) == "zombie" {
		unit.TypeInd = 100
	}
	if os.Getenv(child) == "vile" {
		unit.TypeInd = 200
	}
	update.StatusFlags = 0
	if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unit.CObj()) <= 0xffffffff || uintptr(unit.UpdateData) <= 0xffffffff) {
		t.Fatal("native C-owned links must exceed4GiB")
	}
	a, ok := server.GetAIAction(ai.ACTION_DYING).(cgoAIAction)
	if !ok {
		t.Fatal("ordinary dying dispatcher")
	}
	beforeUnit, beforeUpdate := *unit, *update
	a.End(unit)
	t.Logf("DYING END ORDINARY COMPLETED: mode=%s unit=%p update=%p type=%d on-fire=false", os.Getenv(child), unit, update, unit.TypeInd)
	rawUnit := func(v *server.Object) []byte { return unsafe.Slice((*byte)(unsafe.Pointer(v)), int(unsafe.Sizeof(*v))) }
	rawUpdate := func(v *server.MonsterUpdateData) []byte {
		return unsafe.Slice((*byte)(unsafe.Pointer(v)), int(unsafe.Sizeof(*v)))
	}
	if !bytes.Equal(rawUnit(unit), rawUnit(&beforeUnit)) || !bytes.Equal(rawUpdate(update), rawUpdate(&beforeUpdate)) {
		t.Fatal("ordinary nonburning control changed native records")
	}
	ccall.CallVoidPtr(monsterDyingEndPublicC544CA0(), unit.CObj())
	if !bytes.Equal(rawUnit(unit), rawUnit(&beforeUnit)) || !bytes.Equal(rawUpdate(update), rawUpdate(&beforeUpdate)) {
		t.Fatal("public C nonburning path changed complete native records")
	}
	t.Logf("DYING END NATIVE C PASS: mode=%s type=%d on-fire=false full-object-update=true", os.Getenv(child), unit.TypeInd)
}
