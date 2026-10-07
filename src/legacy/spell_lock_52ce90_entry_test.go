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

type lockCastLegacyServer52CE90 struct {
	Server
	srv *server.Server
}

func (s *lockCastLegacyServer52CE90) S() *server.Server { return s.srv }

const lockCastChild52CE90 = "OPENNOX_TEST_LOCK_NATIVE_WIDTH"

// Isolate the original six-int public dispatcher's truncated third-object
// coordinate load. The native route must query the real empty spatial index
// without substituting a selected door or changing any selector argument.
func TestLockCastNativeEntry52CE90(t *testing.T) {
	if os.Getenv(lockCastChild52CE90) != "1" {
		cmd := exec.Command(os.Args[0], "-test.run=^TestLockCastNativeEntry52CE90$", "-test.count=1", "-test.v")
		cmd.Env = append(os.Environ(), lockCastChild52CE90+"=1")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("isolated Lock native entry: %v\n%s", err, out)
		}
		t.Logf("%s", out)
		return
	}
	second, freeSecond := alloc.New(server.Object{})
	defer freeSecond()
	caster, freeCaster := alloc.New(server.Object{})
	defer freeCaster()
	aim, freeAim := alloc.New(server.Object{})
	defer freeAim()
	*caster = server.Object{PosVec: types.Ptf(300, 500)}
	*aim = server.Object{PosVec: types.Ptf(330, 520)}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, ptr := range []unsafe.Pointer{unsafe.Pointer(second), unsafe.Pointer(caster), unsafe.Pointer(aim)} {
			if uintptr(ptr) <= math.MaxUint32 {
				t.Fatalf("Lock pointer=%p, want actual C allocation above 4 GiB", ptr)
			}
		}
	}
	s := new(server.Server)
	s.Map.Init()
	defer s.Map.Free()
	previousServer := GetServer
	t.Cleanup(func() { GetServer = previousServer })
	GetServer = func() Server { return &lockCastLegacyServer52CE90{srv: s} }
	beforeSecond, beforeCaster, beforeAim := *second, *caster, *aim
	t.Logf("native second=%p caster=%p aim=%p", second, caster, aim)
	if got := Nox_xxx_castLock_52CE90(spell.SPELL_LOCK, second, caster, aim, nil, math.MinInt32); got != 0 ||
		*second != beforeSecond || *caster != beforeCaster || *aim != beforeAim {
		t.Fatalf("Lock empty native map result=%d or selector records changed", got)
	}
	t.Log("real empty native spatial index returned no door; selector records unchanged")
}
