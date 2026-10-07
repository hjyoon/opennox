package legacy

import (
	"math"
	"os"
	"os/exec"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

const markSlotChild52CBD0 = "OPENNOX_TEST_MARK_SLOTS_NATIVE_WIDTH"

// Exercise the actual public entry with high C-owned objects. A child process
// keeps the pre-port integer-pointer SIGSEGV from killing the parent runner.
func TestMarkSlotCastNativeEntry52CBD0(t *testing.T) {
	if os.Getenv(markSlotChild52CBD0) != "1" {
		cmd := exec.Command(os.Args[0], "-test.run=^TestMarkSlotCastNativeEntry52CBD0$", "-test.count=1", "-test.v")
		cmd.Env = append(os.Environ(), markSlotChild52CBD0+"=1")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("isolated Mark 1..4 native entry: %v\n%s", err, out)
		}
		t.Logf("%s", out)
		return
	}
	second, freeSecond := alloc.New(server.Object{})
	defer freeSecond()
	caster, freeCaster := alloc.New(server.Object{})
	defer freeCaster()
	arg, freeArg := alloc.New(server.SpellAcceptArg{})
	defer freeArg()
	*second = server.Object{PosVec: types.Ptf(330, 520)}
	*caster = server.Object{ObjClass: object.ClassMonster, PosVec: types.Ptf(300, 500)}
	*arg = server.SpellAcceptArg{Obj: second, Pos: types.Ptf(310, 510)}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, ptr := range []unsafe.Pointer{unsafe.Pointer(second), unsafe.Pointer(caster), unsafe.Pointer(arg)} {
			if uintptr(ptr) <= math.MaxUint32 {
				t.Fatalf("Mark 1..4 pointer=%p, want actual C allocation above 4 GiB", ptr)
			}
		}
	}
	cache := Get_dword_5d4594_2487712_ptr()
	previousCache, previousServer := *cache, GetServer
	t.Cleanup(func() { *cache, GetServer = previousCache, previousServer })
	*cache = 0x5678
	GetServer = func() Server { return &markCastLegacyServer52CA80{srv: new(server.Server)} }
	beforeSecond, beforeCaster, beforeArg := *second, *caster, *arg
	t.Logf("native second=%p caster=%p argument=%p", second, caster, arg)
	for id := spell.SPELL_MARK_1; id <= spell.SPELL_MARK_4; id++ {
		if got := Sub_52CBD0(id, second, caster, nil, arg, math.MinInt32); got != 1 ||
			*second != beforeSecond || *caster != beforeCaster || *arg != beforeArg || *cache != 0x5678 {
			t.Fatalf("Mark slot id=%d non-player result=%d or selector records/cache changed", id, got)
		}
	}
	t.Log("four real native non-player entries returned success with all records/cache unchanged")
}
