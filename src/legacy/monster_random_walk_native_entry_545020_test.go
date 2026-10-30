package legacy

import (
	"bytes"
	"os"
	"os/exec"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/prand"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
	"github.com/opennox/opennox/v1/server"
)

// The original update dispatcher overwrites EAX immediately after its callback.
// Root and Confused are observed as real void AI updates; the direction helper
// has a separate defined0..255 public int, which is checked here.
func TestMonsterRandomWalkNextPublicCNative545020(t *testing.T) {
	const child = "NOX_MONSTER_RANDOM_WALK_NATIVE_CHILD"
	if os.Getenv(child) == "" {
		for _, mode := range []string{"root", "direction", "confused"} {
			t.Run(mode, func(t *testing.T) {
				bin, err := os.Executable()
				if err != nil {
					t.Fatal(err)
				}
				cmd := exec.Command(bin, "-test.run=^TestMonsterRandomWalkNextPublicCNative545020$", "-test.v")
				cmd.Env = append(os.Environ(), child+"="+mode)
				out, err := cmd.CombinedOutput()
				t.Logf("%s", out)
				if err != nil {
					t.Fatalf("actual native public RandomWalk/direction/Confused path: %v", err)
				}
			})
		}
		return
	}
	mode := os.Getenv(child)
	s, unit, update, missile := monsterMainBlockLegacyFixture547210(t)
	def, free := alloc.New(server.MonsterDef{})
	t.Cleanup(free)
	*def = server.MonsterDef{RunMultiplier96: 1.75, MoveSndFrameA100: 3, MoveSndFrameB104: 7}
	unit.ObjSubClass, unit.Direction1, unit.Direction2, unit.SpeedCur = object.SubClass(0x400), 255, 13, 2.5
	update.MonsterDef, update.SoundSet122, update.AIStackInd = def, nil, 0
	update.AIStack[0].Action = uint32(ai.ACTION_RANDOM_WALK)
	seed := 0
	if mode == "confused" {
		update.AIStack[0].Action = uint32(ai.ACTION_CONFUSED)
		for ; seed < 1000; seed++ {
			if prand.New(seed).IntClamp(0, 100) >= 15 {
				break
			}
		}
		if seed == 1000 {
			t.Fatal("cannot select natural RandomWalk branch")
		}
	}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, p := range []unsafe.Pointer{unit.CObj(), unsafe.Pointer(update), unsafe.Pointer(def)} {
			if uintptr(p) <= 0xffffffff {
				t.Fatal("actual native unit/update/definition must exceed4GiB")
			}
		}
	}
	beforeUnit, beforeUpdate, beforeDef, beforeMissile := *unit, *update, *def, *missile
	s.Rand.Logic = prand.New(seed)
	a, ok := server.GetAIAction(update.AIStack[0].Type()).(cgoAIAction)
	if !ok {
		t.Fatal("ordinary AI dispatcher")
	}
	a.Update(unit)
	wantUnit, wantUpdate, wantDef := *unit, *update, *def
	wantDirection := int(unit.Direction1)
	t.Logf("RANDOM WALK CONTROL COMPLETED: mode=%s unit=%p update=%p definition=%p direction=%d", mode, unit, update, def, wantDirection)
	*unit, *update, *def = beforeUnit, beforeUpdate, beforeDef
	s.Rand.Logic = prand.New(seed)
	if mode == "direction" {
		got := ccall.CallIntPtr(monsterRandomDirectionPublicC545090(), unit.CObj())
		if got != wantDirection {
			t.Fatalf("original helper public int=%d want=%d", got, wantDirection)
		}
		wantUnit, wantUpdate, wantDef = beforeUnit, beforeUpdate, beforeDef
	} else if mode == "root" {
		ccall.CallVoidPtr(monsterRandomWalkPublicC545020(), unit.CObj())
	} else if mode == "confused" {
		ccall.CallVoidPtr(monsterConfusedPublicC545140(), unit.CObj())
	} else {
		t.Fatal("unknown frozen native mode")
	}
	raw := func(p unsafe.Pointer, n uintptr) []byte { return unsafe.Slice((*byte)(p), int(n)) }
	for _, v := range []struct {
		got, want unsafe.Pointer
		n         uintptr
	}{
		{unit.CObj(), unsafe.Pointer(&wantUnit), unsafe.Sizeof(wantUnit)},
		{unsafe.Pointer(update), unsafe.Pointer(&wantUpdate), unsafe.Sizeof(wantUpdate)},
		{unsafe.Pointer(def), unsafe.Pointer(&wantDef), unsafe.Sizeof(wantDef)},
		{missile.CObj(), unsafe.Pointer(&beforeMissile), unsafe.Sizeof(beforeMissile)},
	} {
		if !bytes.Equal(raw(v.got, v.n), raw(v.want, v.n)) {
			t.Fatal("whole native unit/update/definition/missile records differ")
		}
	}
	t.Logf("RANDOM WALK NATIVE C PASS: mode=%s int-helper-and-void-update-contract=true full-records=true", mode)
}
