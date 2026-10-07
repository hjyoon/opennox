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

type markCastLegacyServer52CA80 struct {
	Server
	srv *server.Server
}

func (s *markCastLegacyServer52CA80) S() *server.Server { return s.srv }

const markCastChild52CA80 = "OPENNOX_TEST_MARK_NATIVE_WIDTH"

// The original public six-int dispatcher faults on the third object's class
// byte, even for an ordinary non-player that should simply return success.
// Exercise the real entry with C-owned records, isolated from the test runner.
func TestMarkCastNativeEntry52CA80(t *testing.T) {
	if os.Getenv(markCastChild52CA80) != "1" {
		cmd := exec.Command(os.Args[0], "-test.run=^TestMarkCastNativeEntry52CA80$", "-test.count=1", "-test.v")
		cmd.Env = append(os.Environ(), markCastChild52CA80+"=1")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("isolated Mark native entry: %v\n%s", err, out)
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
				t.Fatalf("Mark pointer=%p, want actual C allocation above 4 GiB", ptr)
			}
		}
	}
	cache := Get_dword_5d4594_2487712_ptr()
	previousCache := *cache
	defer func() { *cache = previousCache }()
	*cache = 0x5678
	previousServer := GetServer
	t.Cleanup(func() { GetServer = previousServer })
	GetServer = func() Server { return &markCastLegacyServer52CA80{srv: new(server.Server)} }
	beforeSecond, beforeCaster, beforeArg := *second, *caster, *arg
	t.Logf("native second=%p caster=%p argument=%p", second, caster, arg)
	if got := Sub_52CA80(spell.SPELL_MARK, second, caster, nil, arg, math.MinInt32); got != 1 ||
		*second != beforeSecond || *caster != beforeCaster || *arg != beforeArg || *cache != 0x5678 {
		t.Fatalf("Mark non-player result=%d or selector records/cache changed", got)
	}
	t.Log("real native non-player entry returned success; selector records and live Glyph cache unchanged")
}
