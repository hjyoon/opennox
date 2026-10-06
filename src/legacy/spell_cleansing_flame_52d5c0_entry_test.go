package legacy

import (
	"math"
	"os"
	"os/exec"
	"testing"
	"unsafe"

	"github.com/opennox/libs/spell"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

const cleansingFlameChild52D5C0 = "OPENNOX_TEST_CLEANSING_FLAME_NATIVE_WIDTH"

func TestCleansingFlameCastPublicPointerForwarding52D5C0(t *testing.T) {
	previous := cleansingFlameCastCall52D5C0
	t.Cleanup(func() { cleansingFlameCastCall52D5C0 = previous })
	second, freeSecond := alloc.New(server.Object{})
	defer freeSecond()
	recipient, freeRecipient := alloc.New(server.Object{})
	defer freeRecipient()
	caster, freeCaster := alloc.New(server.Object{})
	defer freeCaster()
	arg, freeArg := alloc.New(server.SpellAcceptArg{})
	defer freeArg()
	for _, id := range []spell.ID{spell.SPELL_CLEANSING_FLAME, spell.SPELL_CLEANSING_MANA_FLAME} {
		for _, level := range []int{-1, 1, 5} {
			calls := 0
			cleansingFlameCastCall52D5C0 = func(gotID spell.ID, gotSecond, gotRecipient, gotCaster *server.Object, gotArg *server.SpellAcceptArg, gotLevel int) int {
				calls++
				if gotID != id || gotSecond != second || gotRecipient != recipient || gotCaster != caster || gotArg != arg || gotLevel != level {
					t.Fatal("public entry changed native pointers, spell or level")
				}
				return -123
			}
			if got := Nox_xxx_spellCastCleansingFlame_52D5C0(id, second, recipient, caster, arg, level); got != -123 || calls != 1 {
				t.Fatalf("public result/calls=%d/%d", got, calls)
			}
		}
	}
}

// The old C body narrows second before reading its owned list at +516.
// Isolate that fatal regression, then use the real public entry and owned
// list guard. No cast, lookup, priority-message or trace result is supplied.
func TestCleansingFlameCastNativeEntry52D5C0(t *testing.T) {
	if os.Getenv(cleansingFlameChild52D5C0) != "1" {
		cmd := exec.Command(os.Args[0], "-test.run=^TestCleansingFlameCastNativeEntry52D5C0$", "-test.count=1", "-test.v")
		cmd.Env = append(os.Environ(), cleansingFlameChild52D5C0+"=1")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("isolated Cleansing Flame native entry: %v\n%s", err, out)
		}
		t.Logf("%s", out)
		return
	}
	second, freeSecond := alloc.New(server.Object{})
	defer freeSecond()
	firstOwned, freeFirst := alloc.New(server.Object{})
	defer freeFirst()
	lastOwned, freeLast := alloc.New(server.Object{})
	defer freeLast()
	arg, freeArg := alloc.New(server.SpellAcceptArg{})
	defer freeArg()
	second.Field129, firstOwned.TypeInd, firstOwned.Field128, lastOwned.TypeInd = firstOwned, 77, lastOwned, 321
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, ptr := range []unsafe.Pointer{unsafe.Pointer(second), unsafe.Pointer(firstOwned), unsafe.Pointer(lastOwned), unsafe.Pointer(arg)} {
			if uintptr(ptr) <= math.MaxUint32 {
				t.Fatalf("pointer=%p, want actual C-owned allocation above 4 GiB", ptr)
			}
		}
	}
	cache := unsafe.Slice(memmap.PtrUint32(0x5D4594, 2487760), 10)
	beforeCache := [10]uint32{}
	copy(beforeCache[:], cache)
	previousServer := GetServer
	t.Cleanup(func() { copy(cache, beforeCache[:]); GetServer = previousServer })
	for i := range cache {
		cache[i] = uint32(321 + i)
	}
	GetServer = func() Server { return &burnCastLegacyServer52C3E0{srv: new(server.Server)} }
	beforeSecond, beforeFirst, beforeLast, beforeArg := *second, *firstOwned, *lastOwned, *arg
	t.Logf("native second=%p owned-first=%p owned-last=%p argument=%p", second, firstOwned, lastOwned, arg)
	// Matching the second owned unit must reject before dereferencing the
	// unused argument or nil caster; the original nil-recipient helper exits.
	if got := Nox_xxx_spellCastCleansingFlame_52D5C0(spell.SPELL_CLEANSING_FLAME, second, nil, nil, arg, 1); got != 0 ||
		*second != beforeSecond || *firstOwned != beforeFirst || *lastOwned != beforeLast || *arg != beforeArg {
		t.Fatalf("owned-list rejection result=%d, native records changed", got)
	}
	for i, value := range cache {
		if value != uint32(321+i) {
			t.Fatalf("cache[%d]=%d changed", i, value)
		}
	}
}
