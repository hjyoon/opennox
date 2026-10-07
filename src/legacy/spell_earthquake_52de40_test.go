package legacy

import (
	"math"
	"os"
	"os/exec"
	"testing"
	"unsafe"

	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

// Isolate the original six-int SIGSEGV while exercising the real public
// selector with actual C-owned native-width addresses, not forged pointers.
func TestEarthquakeCastEntryNativePointers52DE40(t *testing.T) {
	const child = "OPENNOX_TEST_EARTHQUAKE_NATIVE_ENTRY_52DE40"
	if os.Getenv(child) != "1" {
		binary, err := os.Executable()
		if err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command(binary, "-test.run=^TestEarthquakeCastEntryNativePointers52DE40$", "-test.v")
		cmd.Env = append(os.Environ(), child+"=1")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("native public Earthquake entry subprocess: %v\n%s", err, out)
		}
		t.Logf("%s", out)
		return
	}
	srv, caster, _ := monsterLookAtFixture5125A0(t)
	previousServer := GetServer
	t.Cleanup(func() { GetServer = previousServer })
	GetServer = func() Server { return &telekinesisCastLegacyServer52D330{srv: srv} }
	second, freeSecond := alloc.New(server.Object{})
	defer freeSecond()
	owner, freeOwner := alloc.New(server.Object{})
	defer freeOwner()
	arg, freeArg := alloc.New(server.SpellAcceptArg{})
	defer freeArg()
	*arg = server.SpellAcceptArg{Obj: second, Pos: types.Ptf(12, 34)}
	for _, p := range []unsafe.Pointer{unsafe.Pointer(second), unsafe.Pointer(owner), unsafe.Pointer(caster), unsafe.Pointer(arg)} {
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(p) <= math.MaxUint32 {
			t.Fatalf("native pointer below 4 GiB: %p", p)
		}
	}
	t.Logf("native C-owned second=%p owner=%p caster=%p acceptance=%p", second, owner, caster, arg)
	before := *arg
	previous := earthquakeCastCall52DE40
	t.Cleanup(func() { earthquakeCastCall52DE40 = previous })
	powers := []int{1, 2, 3, 4, 5, 0, -129, 128, math.MinInt32, math.MaxInt32}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, power := range []int64{0x100000003, -0x100000003, math.MinInt64, math.MaxInt64} {
			powers = append(powers, int(power))
		}
	}
	for _, power := range powers {
		calls := 0
		earthquakeCastCall52DE40 = func(id spell.ID, gotSecond, gotOwner, gotCaster *server.Object, gotArg *server.SpellAcceptArg, gotPower int) int {
			calls++
			if id != spell.SPELL_EARTHQUAKE || gotSecond != second || gotOwner != owner || gotCaster != caster || gotArg != arg || gotPower != power {
				t.Fatalf("native forward=%s/%p/%p/%p/%p/%d", id, gotSecond, gotOwner, gotCaster, gotArg, gotPower)
			}
			return -17
		}
		if got := Nox_xxx_castEquake_52DE40(spell.SPELL_EARTHQUAKE, second, owner, caster, arg, power); got != -17 || calls != 1 || *arg != before {
			t.Fatalf("native result=%d calls=%d acceptance=%+v", got, calls, *arg)
		}
	}
}

func TestEarthquakeCastNativeBridgeDWORDLevel52DE40(t *testing.T) {
	srv, caster, _ := monsterLookAtFixture5125A0(t)
	srv.Map.Init()
	t.Cleanup(srv.Map.Free)
	previousServer := GetServer
	t.Cleanup(func() { GetServer = previousServer })
	GetServer = func() Server { return &telekinesisCastLegacyServer52D330{srv: srv} }
	level := memmap.PtrUint32(0x5D4594, 2487700)
	previousLevel := *level
	t.Cleanup(func() { *level = previousLevel })
	before := *caster
	powers := []int{1, 2, 3, 4, 5, 0, -129, 128, math.MinInt32, math.MaxInt32}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, power := range []int64{0x100000003, -0x100000003, math.MinInt64, math.MaxInt64} {
			powers = append(powers, int(power))
		}
	}
	for _, power := range powers {
		// Real bridge, real memmap DWORD and real map/sound/quake services.
		// Empty balance and no indexed units do not avoid reading caster.Pos.
		if got := Nox_xxx_castEquake_52DE40(spell.SPELL_EARTHQUAKE, nil, nil, caster, nil, power); got != 1 || *level != uint32(power) || *caster != before {
			t.Fatalf("native bridge result/cache/source=%d/%#x/%p power=%d", got, *level, caster, power)
		}
	}
}
