package legacy

import (
	"fmt"
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

type telekinesisCastLegacyServer52D330 struct {
	Server
	srv *server.Server
}

func (s *telekinesisCastLegacyServer52D330) S() *server.Server { return s.srv }

// A subprocess preserves the actual red six-int SIGSEGV without terminating
// unrelated tests. The green public entry must forward every native pointer.
func TestTelekinesisCastEntryNativePointers52D330(t *testing.T) {
	const child = "OPENNOX_TEST_TELEKINESIS_NATIVE_ENTRY_52D330"
	if os.Getenv(child) != "1" {
		binary, err := os.Executable()
		if err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command(binary, "-test.run=^TestTelekinesisCastEntryNativePointers52D330$", "-test.v")
		cmd.Env = append(os.Environ(), child+"=1")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("native public entry subprocess: %v\n%s", err, out)
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
	*second = server.Object{ObjClass: object.ClassPlayer}
	*arg = server.SpellAcceptArg{Obj: second, Pos: types.Ptf(12, 34)}
	for _, p := range []unsafe.Pointer{unsafe.Pointer(second), unsafe.Pointer(owner), unsafe.Pointer(caster), unsafe.Pointer(arg)} {
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(p) <= math.MaxUint32 {
			t.Fatalf("native pointer below 4 GiB: %p", p)
		}
	}
	t.Logf("native C-owned target=%p owner=%p caster=%p acceptance=%p", second, owner, caster, arg)
	before := *arg
	previous := telekinesisCastCall52D330
	t.Cleanup(func() { telekinesisCastCall52D330 = previous })
	for _, power := range []int{1, 2, 3, 4, 5, 0, -129, 128, math.MinInt32, math.MaxInt32} {
		calls := 0
		telekinesisCastCall52D330 = func(id spell.ID, gotSecond, gotOwner, gotCaster *server.Object, gotArg *server.SpellAcceptArg, gotPower int) int {
			calls++
			if id != spell.SPELL_TELEKINESIS || gotSecond != second || gotOwner != owner || gotCaster != caster || gotArg != arg || gotPower != power {
				t.Fatalf("native forward=%s/%p/%p/%p/%p/%d", id, gotSecond, gotOwner, gotCaster, gotArg, gotPower)
			}
			return -17
		}
		if got := Nox_xxx_castTelekinesis_52D330(spell.SPELL_TELEKINESIS, second, owner, caster, arg, power); got != -17 || calls != 1 || *arg != before {
			t.Fatalf("native result=%d calls=%d acceptance=%+v", got, calls, *arg)
		}
	}
}

func TestTelekinesisCastNativeMissingStockType52D330(t *testing.T) {
	for _, class := range []object.Class{0, object.ClassMonster, object.ClassPlayer, object.ClassPlayer | object.ClassImmobile} {
		t.Run(fmt.Sprintf("class%08x", class), func(t *testing.T) {
			srv, target, update := monsterLookAtFixture5125A0(t)
			target.ObjClass = class
			arg, freeArg := alloc.New(server.SpellAcceptArg{})
			t.Cleanup(freeArg)
			*arg = server.SpellAcceptArg{Obj: target, Pos: types.Ptf(44, 55)}
			previousServer := GetServer
			t.Cleanup(func() { GetServer = previousServer })
			GetServer = func() Server { return &telekinesisCastLegacyServer52D330{srv: srv} }
			beforeTarget, beforeUpdate, beforeArg := *target, *update, *arg
			// Empty real type registry: no supplied allocation result, no
			// placement/buff/cancel/audio should run on allocation failure.
			if got := castTelekinesisNative52D330(spell.SPELL_TELEKINESIS, nil, nil, nil, arg, 3); got != 1 || *target != beforeTarget || *update != beforeUpdate || *arg != beforeArg {
				t.Fatalf("native missing-type result=%d target/acceptance changed", got)
			}
		})
	}
}
