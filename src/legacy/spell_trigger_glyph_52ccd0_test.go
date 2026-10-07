package legacy

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/spell"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestTriggerGlyphPublicSelector52CCD0KeepsPointersAndSignedResult(t *testing.T) {
	second, freeSecond := alloc.New(server.Object{})
	caster, freeCaster := alloc.New(server.Object{})
	fourth, freeFourth := alloc.New(server.Object{})
	arg, freeArg := alloc.New(server.SpellAcceptArg{})
	for _, free := range []func(){freeSecond, freeCaster, freeFourth, freeArg} {
		t.Cleanup(free)
	}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, ptr := range []unsafe.Pointer{unsafe.Pointer(second), unsafe.Pointer(caster), unsafe.Pointer(fourth), unsafe.Pointer(arg)} {
			if uintptr(ptr) <= math.MaxUint32 {
				t.Fatalf("native pointer=%p, want above 4 GiB", ptr)
			}
		}
	}
	old := triggerGlyphCastCall52CCD0
	t.Cleanup(func() { triggerGlyphCastCall52CCD0 = old })
	calls := 0
	triggerGlyphCastCall52CCD0 = func(id spell.ID, a2, a3, a4 *server.Object, sa *server.SpellAcceptArg, level int) int {
		calls++
		if id != spell.ID(math.MinInt32) || a2 != second || a3 != caster || a4 != fourth || sa != arg || level != math.MinInt32 {
			t.Fatal("public TriggerGlyph selector narrowed or reordered arguments")
		}
		return math.MinInt32
	}
	if got := Sub_52CCD0(spell.ID(math.MinInt32), second, caster, fourth, arg, math.MinInt32); got != math.MinInt32 || calls != 1 {
		t.Fatalf("signed selector result/calls=%d/%d", got, calls)
	}
}
