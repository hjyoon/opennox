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

type toxicCloudCastLegacyServer52DB60 struct {
	Server
	srv *server.Server
}

func (s *toxicCloudCastLegacyServer52DB60) S() *server.Server { return s.srv }

const toxicCloudCastChild52DB60 = "OPENNOX_TEST_TOXIC_CLOUD_NATIVE_WIDTH"

// Run the public entry in an isolated process: the old five-int C callee
// truncates caster before loading its position, as in the report.
func TestToxicCloudCastNativeEntry52DB60(t *testing.T) {
	if os.Getenv(toxicCloudCastChild52DB60) != "1" {
		cmd := exec.Command(os.Args[0], "-test.run=^TestToxicCloudCastNativeEntry52DB60$", "-test.count=1", "-test.v")
		cmd.Env = append(os.Environ(), toxicCloudCastChild52DB60+"=1")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("isolated Toxic Cloud native entry: %v\n%s", err, out)
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
				t.Fatalf("Toxic Cloud pointer=%p, want actual allocation above 4 GiB", pointer)
			}
		}
	}
	cache := memmap.PtrUint32(0x5D4594, 2487808)
	previousCache, previousServer := *cache, GetServer
	t.Cleanup(func() { *cache, GetServer = previousCache, previousServer })
	*cache = 321
	GetServer = func() Server { return &toxicCloudCastLegacyServer52DB60{srv: new(server.Server)} }
	beforeSecond, beforeOwner, beforeCaster, beforeArg := *second, *owner, *caster, *arg
	// Real MapTraceRay rejects an out-of-map source before wall allocation,
	// with no supplied trace, allocation, or cast result.
	if got := Nox_xxx_castToxicCloud_52DB60(spell.SPELL_TOXIC_CLOUD, second, owner, caster, arg, 1); got != 0 ||
		*cache != 321 || *second != beforeSecond || *owner != beforeOwner || *caster != beforeCaster || *arg != beforeArg {
		t.Fatalf("Toxic Cloud blocked entry changed state: result=%d cache=%d", got, *cache)
	}
	t.Logf("native caster=%p owner=%p argument=%p cache=%p trace=blocked result=0", caster, owner, arg, cache)
}
