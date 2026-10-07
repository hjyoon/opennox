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

type curePoisonLegacyServer52CDB0 struct {
	Server
	srv *server.Server
}

func (s *curePoisonLegacyServer52CDB0) S() *server.Server { return s.srv }

const curePoisonChild52CDB0 = "OPENNOX_TEST_CURE_POISON_NATIVE_WIDTH"

// Isolate the old public C entry's first low-DWORD target load from the
// runner. This is separate from the already-Go game spell dispatcher.
func TestCurePoisonNativeEntry52CDB0(t *testing.T) {
	if os.Getenv(curePoisonChild52CDB0) != "1" {
		cmd := exec.Command(os.Args[0], "-test.run=^TestCurePoisonNativeEntry52CDB0$", "-test.count=1", "-test.v")
		cmd.Env = append(os.Environ(), curePoisonChild52CDB0+"=1")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("isolated CurePoison native entry: %v\n%s", err, out)
		}
		t.Logf("%s", out)
		return
	}
	target, freeTarget := alloc.New(server.Object{})
	defer freeTarget()
	health, freeHealth := alloc.New(server.HealthData{})
	defer freeHealth()
	arg, freeArg := alloc.New(server.SpellAcceptArg{})
	defer freeArg()
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, ptr := range []unsafe.Pointer{unsafe.Pointer(target), unsafe.Pointer(health), unsafe.Pointer(arg)} {
			if uintptr(ptr) <= math.MaxUint32 {
				t.Fatalf("CurePoison pointer=%p, want actual allocation above 4 GiB", ptr)
			}
		}
	}
	oldServer := GetServer
	t.Cleanup(func() { GetServer = oldServer })
	GetServer = func() Server { return &curePoisonLegacyServer52CDB0{srv: new(server.Server)} }
	for _, power := range []int32{math.MinInt32, -3, -1, 0, 1, 2, 3, 255, math.MaxInt32} {
		*health = server.HealthData{Field16: 91}
		*target = server.Object{HealthData: health, Poison540: 3, Field542: 1234}
		*arg = server.SpellAcceptArg{Obj: target, Pos: types.Ptf(-1000, 2000)}
		wantArg := *arg
		wantPoison, wantFrame := uint8(0), uint32(0)
		if power < 3 {
			wantPoison, wantFrame = uint8(3)-uint8(power), 91
		}
		if got := Nox_xxx_castCurePoison_52CDB0(spell.SPELL_CURE_POISON, nil, nil, nil, arg, int(power)); got != 1 ||
			*arg != wantArg || target.Poison540 != wantPoison || health.Field16 != wantFrame || target.Field542 != 1234 {
			t.Fatalf("native power=%d result=%d poison=%d frame=%d timer=%d arg=%+v", power, got, target.Poison540, health.Field16, target.Field542, *arg)
		}
	}
	arg.Obj = nil
	if got := Nox_xxx_castCurePoison_52CDB0(spell.SPELL_CURE_POISON, nil, nil, nil, arg, 3); got != 0 {
		t.Fatalf("nil target result=%d, want 0", got)
	}
	t.Logf("native target=%p health=%p argument=%p nine signed powers and nil target passed", target, health, arg)
}
