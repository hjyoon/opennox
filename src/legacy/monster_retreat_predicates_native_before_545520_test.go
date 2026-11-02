package legacy

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"os/exec"
	"strconv"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func TestMonsterRetreatPredicatesNativeLayout545520(t *testing.T) {
	obj, update, health := server.Object{}, server.MonsterUpdateData{}, server.HealthData{}
	want := [6]uintptr{unsafe.Offsetof(obj.HealthData), unsafe.Offsetof(obj.UpdateData), unsafe.Offsetof(obj.Buffs), unsafe.Offsetof(health.Max), unsafe.Offsetof(update.ResumeLevel), unsafe.Offsetof(update.StatusFlags)}
	if got := retreatPredicatesOriginalLayout545520(); got != want {
		t.Fatal("independent native C/Go layout", got, want)
	}
}

func TestMonsterRetreatPredicatesNativeCBefore545520(t *testing.T) {
	if os.Getenv("OPENNOX_RETREAT_PREDICATES_C_BEFORE") != "1" {
		return
	}
	if os.Getenv("OPENNOX_RETREAT_PREDICATES_C_CHILD") == "" {
		for entry := 0; entry < 3; entry++ {
			t.Run(strconv.Itoa(entry), func(t *testing.T) {
				cmd := exec.Command(os.Args[0], "-test.run=^TestMonsterRetreatPredicatesNativeCBefore545520$", "-test.v")
				cmd.Env = append(os.Environ(), "OPENNOX_RETREAT_PREDICATES_C_CHILD="+strconv.Itoa(entry))
				out, err := cmd.CombinedOutput()
				t.Logf("%s", out)
				marker := []byte(fmt.Sprintf("RETREAT_PREDICATE_VALID_NATIVE_CONTROLS entry=%d", entry))
				if err == nil || !bytes.Contains(out, marker) || !bytes.Contains(out, []byte("SIGSEGV")) || !bytes.Contains(out, []byte("_Cfunc_retreat_original_public_545520")) {
					t.Fatal("legacy-int public C must fault only after independent native controls", err)
				}
			})
		}
		return
	}
	entry, err := strconv.Atoi(os.Getenv("OPENNOX_RETREAT_PREDICATES_C_CHILD"))
	if err != nil || entry < 0 || entry > 2 {
		t.Fatal("invalid child selector")
	}
	TestMonsterRetreatPredicatesNativeLayout545520(t)
	TestMonsterDodgeablesNativeLayout547C50(t)
	w := newDodgeWorld547C50(t, dodgeRow547C50{})
	w.Unit.HealthData = (*server.HealthData)(unsafe.Pointer(&w.Heap[24000]))
	w.Unit.HealthData.Cur, w.Unit.HealthData.Max = 25, 100
	w.Data.ResumeLevel = 0.5
	w.Data.StatusFlags = 0x20
	w.Unit.Buffs = 0x20000000
	w.Data.AIStackInd = 0
	w.Data.AIStack[0].Action = 0
	w.Data.SoundSet122 = unsafe.Pointer(w.Pushes[0])
	for _, ptr := range []unsafe.Pointer{unsafe.Pointer(w.Unit), unsafe.Pointer(w.Unit.HealthData), unsafe.Pointer(w.Data)} {
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(ptr) <= math.MaxUint32 {
			t.Fatal("native control must exceed4GiB")
		}
	}
	before := bytes.Clone(w.Heap)
	set, action := dodgeOriginalControls547C50(w.Unit)
	if set != uintptr(w.Data.SoundSet122) || action != 0 || unitBuffTestExportCall4FF350(w.Unit, 29) != 1 || !bytes.Equal(before, w.Heap) {
		t.Fatal("typed native sound/action/buff positive controls or whole allocation guard")
	}
	if retreatPredicatesOriginalABI545520(entry) != 32 {
		t.Fatal("unchanged legacy-int public ABI required")
	}
	t.Logf("RETREAT_PREDICATE_VALID_NATIVE_CONTROLS entry=%d unit=%p health=%p update=%p", entry, w.Unit, w.Unit.HealthData, w.Data)
	retreatPredicatesOriginalC545520(entry, w.Unit)
	t.Fatal("old C unexpectedly returned")
}
