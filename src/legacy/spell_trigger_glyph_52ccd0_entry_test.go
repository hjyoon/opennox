package legacy

import (
	"math"
	"os"
	"os/exec"
	"testing"
	"unsafe"

	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type triggerGlyphLegacyServer52CCD0 struct {
	Server
	srv *server.Server
}

func (s *triggerGlyphLegacyServer52CCD0) S() *server.Server { return s.srv }

const triggerGlyphChild52CCD0 = "OPENNOX_TEST_TRIGGER_GLYPH_NATIVE_WIDTH"

// Exercise the public selector against the real world list and ownership
// service. Isolate the old C callee's truncated-object SIGSEGV from the runner.
func TestTriggerGlyphNativeEntry52CCD0(t *testing.T) {
	if os.Getenv(triggerGlyphChild52CCD0) != "1" {
		cmd := exec.Command(os.Args[0], "-test.run=^TestTriggerGlyphNativeEntry52CCD0$", "-test.count=1", "-test.v")
		cmd.Env = append(os.Environ(), triggerGlyphChild52CCD0+"=1")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("isolated TriggerGlyph native entry: %v\n%s", err, out)
		}
		t.Logf("%s", out)
		return
	}
	worldObject, freeObject := alloc.New(server.Object{})
	defer freeObject()
	caster, freeCaster := alloc.New(server.Object{})
	defer freeCaster()
	arg, freeArg := alloc.New(server.SpellAcceptArg{})
	defer freeArg()
	*worldObject = server.Object{PosVec: types.Ptf(11, 13)}
	*caster = server.Object{PosVec: types.Ptf(100, 200)}
	*arg = server.SpellAcceptArg{Obj: worldObject, Pos: types.Ptf(300, 400)}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, ptr := range []unsafe.Pointer{unsafe.Pointer(worldObject), unsafe.Pointer(caster), unsafe.Pointer(arg)} {
			if uintptr(ptr) <= math.MaxUint32 {
				t.Fatalf("TriggerGlyph pointer=%p, want actual allocation above 4 GiB", ptr)
			}
		}
	}
	srv := new(server.Server)
	srv.Objs.SetObjects(worldObject)
	oldServer := GetServer
	GetServer = func() Server { return &triggerGlyphLegacyServer52CCD0{srv: srv} }
	t.Cleanup(func() { GetServer = oldServer })
	wantObject, wantCaster, wantArg := *worldObject, *caster, *arg
	// Neither this unit nor any ancestor is the caster. The original returns
	// zero without looking up a type, sound or death service.
	if got := Sub_52CCD0(spell.SPELL_TRIGGER_GLYPH, worldObject, caster, worldObject, arg, 1); got != 0 ||
		*worldObject != wantObject || *caster != wantCaster || *arg != wantArg || srv.Objs.First() != worldObject {
		t.Fatalf("TriggerGlyph no-owned-glyph entry changed state: result=%d", got)
	}
	t.Logf("native world=%p caster=%p argument=%p result=0 ownership=unrelated", worldObject, caster, arg)
}
