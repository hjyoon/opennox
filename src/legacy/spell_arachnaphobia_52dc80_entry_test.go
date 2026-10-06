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

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type arachnaphobiaCastLegacyServer52DC80 struct {
	Server
	srv *server.Server
}

func (s *arachnaphobiaCastLegacyServer52DC80) S() *server.Server { return s.srv }

const arachnaphobiaCastChild52DC80 = "OPENNOX_TEST_ARACHNAPHOBIA_NATIVE_WIDTH"

// The original five-int C entry truncates caster before loading coordinates:
// GAME.EXE reads +0x3c first; the native compiler may read +0x38 first.
// Isolate that fatal regression in a child, then exercise the real public
// native entry and MapTraceRay without supplying a trace/cast result.
func TestArachnaphobiaCastNativeEntry52DC80(t *testing.T) {
	if os.Getenv(arachnaphobiaCastChild52DC80) != "1" {
		cmd := exec.Command(os.Args[0], "-test.run=^TestArachnaphobiaCastNativeEntry52DC80$", "-test.count=1", "-test.v")
		cmd.Env = append(os.Environ(), arachnaphobiaCastChild52DC80+"=1")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("isolated Arachnaphobia native entry: %v\n%s", err, out)
		}
		t.Logf("%s", out)
		return
	}
	second, freeSecond := alloc.New(server.Object{})
	defer freeSecond()
	owner, freeOwner := alloc.New(server.Object{})
	defer freeOwner()
	caster, freeCaster := alloc.New(server.Object{})
	defer freeCaster()
	arg, freeArg := alloc.New(server.SpellAcceptArg{})
	defer freeArg()
	*caster = server.Object{ObjClass: object.ClassMonster, PosVec: types.Ptf(-23, -23)}
	*arg = server.SpellAcceptArg{Obj: second, Pos: types.Ptf(300, 300)}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, pointer := range []unsafe.Pointer{unsafe.Pointer(second), unsafe.Pointer(owner), unsafe.Pointer(caster), unsafe.Pointer(arg)} {
			if uintptr(pointer) <= math.MaxUint32 {
				t.Fatalf("Arachnaphobia pointer=%p, want actual allocation above 4 GiB", pointer)
			}
		}
	}
	cache := memmap.PtrUint32(0x5D4594, 2487812)
	previousCache, previousServer := *cache, GetServer
	t.Cleanup(func() { *cache, GetServer = previousCache, previousServer })
	*cache = 321
	GetServer = func() Server { return &arachnaphobiaCastLegacyServer52DC80{srv: new(server.Server)} }
	beforeSecond, beforeOwner, beforeCaster, beforeArg := *second, *owner, *caster, *arg
	t.Logf("native caster=%p owner=%p argument=%p cache=%p", caster, owner, arg, cache)
	if got := Nox_xxx_spellArachna_52DC80(spell.SPELL_ARACHNAPHOBIA, second, owner, caster, arg, 1); got != 0 ||
		*cache != 321 || *second != beforeSecond || *owner != beforeOwner || *caster != beforeCaster || *arg != beforeArg {
		t.Fatalf("Arachnaphobia blocked entry changed state: result=%d cache=%d", got, *cache)
	}
	t.Log("real out-of-map ray rejected; native records and type cache unchanged")
}
