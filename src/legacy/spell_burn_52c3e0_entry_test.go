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

type burnCastLegacyServer52C3E0 struct {
	Server
	srv *server.Server
}

func (s *burnCastLegacyServer52C3E0) S() *server.Server { return s.srv }

const burnCastChild52C3E0 = "OPENNOX_TEST_BURN_NATIVE_WIDTH"

// The five-int C entry truncates caster before reading its TypeInd at +4.
// Isolate that fatal regression, then exercise the real public entry and
// production MapTraceRay without supplying a trace/cast result.
func TestBurnCastNativeEntry52C3E0(t *testing.T) {
	if os.Getenv(burnCastChild52C3E0) != "1" {
		cmd := exec.Command(os.Args[0], "-test.run=^TestBurnCastNativeEntry52C3E0$", "-test.count=1", "-test.v")
		cmd.Env = append(os.Environ(), burnCastChild52C3E0+"=1")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("isolated Burn native entry: %v\n%s", err, out)
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
				t.Fatalf("Burn pointer=%p, want actual allocation above 4 GiB", pointer)
			}
		}
	}
	glyphCache := Get_dword_5d4594_2487712_ptr()
	flameCache := memmap.PtrUint32(0x5D4594, 2487732)
	previousGlyph, previousFlame, previousServer := *glyphCache, *flameCache, GetServer
	t.Cleanup(func() { *glyphCache, *flameCache, GetServer = previousGlyph, previousFlame, previousServer })
	*glyphCache, *flameCache = 321, 654
	GetServer = func() Server { return &burnCastLegacyServer52C3E0{srv: new(server.Server)} }
	beforeSecond, beforeOwner, beforeCaster, beforeArg := *second, *owner, *caster, *arg
	t.Logf("native caster=%p owner=%p argument=%p glyph-cache=%p flame-cache=%p", caster, owner, arg, glyphCache, flameCache)
	if got := Nox_xxx_castBurn_52C3E0(spell.SPELL_BURN, second, owner, caster, arg, 1); got != 0 ||
		*glyphCache != 321 || *flameCache != 654 || *second != beforeSecond || *owner != beforeOwner || *caster != beforeCaster || *arg != beforeArg {
		t.Fatalf("Burn blocked entry changed state: result=%d glyph=%d flame=%d", got, *glyphCache, *flameCache)
	}
	t.Log("real out-of-map ray rejected; native records and both caches unchanged")
}
