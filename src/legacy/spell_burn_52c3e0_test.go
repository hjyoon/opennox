package legacy

import (
	"fmt"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestBurnCastGoEntryForwardsNativePointers52C3E0(t *testing.T) {
	second, freeSecond := alloc.New(server.Object{})
	t.Cleanup(freeSecond)
	owner, freeOwner := alloc.New(server.Object{})
	t.Cleanup(freeOwner)
	caster, freeCaster := alloc.New(server.Object{})
	t.Cleanup(freeCaster)
	arg, freeArg := alloc.New(server.SpellAcceptArg{})
	t.Cleanup(freeArg)
	*arg = server.SpellAcceptArg{Obj: second, Pos: types.Ptf(12, 34)}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, ptr := range []unsafe.Pointer{unsafe.Pointer(second), unsafe.Pointer(owner), unsafe.Pointer(caster), unsafe.Pointer(arg)} {
			if uintptr(ptr) <= math.MaxUint32 {
				t.Fatalf("Burn dispatch pointer=%p, want >4 GiB", ptr)
			}
		}
	}
	beforeSecond, beforeOwner, beforeCaster, beforeArg := *second, *owner, *caster, *arg
	previous := burnCastCall52C3E0
	t.Cleanup(func() { burnCastCall52C3E0 = previous })
	for _, level := range []int{1, 2, 3, 4, 5, 0, -3, math.MinInt32, math.MaxInt32} {
		t.Run(fmt.Sprint(level), func(t *testing.T) {
			calls := 0
			burnCastCall52C3E0 = func(id spell.ID, gotSecond, gotOwner, gotCaster *server.Object, gotArg *server.SpellAcceptArg, gotLevel int) int {
				calls++
				if id != spell.SPELL_BURN || gotSecond != second || gotOwner != owner || gotCaster != caster || gotArg != arg || gotLevel != level {
					t.Fatalf("native dispatch=%s/%p/%p/%p/%p/%d", id, gotSecond, gotOwner, gotCaster, gotArg, gotLevel)
				}
				return -17
			}
			if got := Nox_xxx_castBurn_52C3E0(spell.SPELL_BURN, second, owner, caster, arg, level); got != -17 || calls != 1 ||
				*second != beforeSecond || *owner != beforeOwner || *caster != beforeCaster || *arg != beforeArg {
				t.Fatalf("result/calls/argument=%d/%d/%+v", got, calls, *arg)
			}
		})
	}
}

// Use real MapTraceRay/type allocation/audio services, not a supplied result.
// A Glyph at an out-of-map position must bypass trace using the C-owned cache;
// a non-Glyph must trace the initialized empty wall map. Missing MediumFlame
// allocation still returns success and calls EventPos in the original.
func TestBurnCastNativeMissingFlameAndSharedGlyphCache52C3E0(t *testing.T) {
	for _, isGlyph := range []bool{false, true} {
		for _, level := range []int{1, 2, 3, 4, 5, 0, -3, math.MinInt32, math.MaxInt32} {
			t.Run(fmt.Sprintf("glyph-%t/level-%d", isGlyph, level), func(t *testing.T) {
				srv, caster, update := monsterLookAtFixture5125A0(t)
				srv.Map.Init()
				t.Cleanup(srv.Map.Free)
				if srv.Walls.Init() == 0 {
					t.Fatal("cannot initialize native wall map")
				}
				t.Cleanup(srv.Walls.Free)
				caster.TypeInd, caster.PosVec = 12, types.Ptf(300, 300)
				if isGlyph {
					caster.TypeInd, caster.PosVec = 321, types.Ptf(-23, -23)
				}
				owner, freeOwner := alloc.New(server.Object{})
				t.Cleanup(freeOwner)
				arg, freeArg := alloc.New(server.SpellAcceptArg{})
				t.Cleanup(freeArg)
				*arg = server.SpellAcceptArg{Obj: owner, Pos: types.Ptf(330, 300)}
				glyphCache := Get_dword_5d4594_2487712_ptr()
				retiredGlyph := memmap.PtrUint32(0x5D4594, 2487712)
				flameCache := memmap.PtrUint32(0x5D4594, 2487732)
				if glyphCache == retiredGlyph {
					t.Fatal("C-owned Glyph cache incorrectly aliases the retired PE32 blob")
				}
				previousGlyph, previousRetired, previousFlame, previousServer := *glyphCache, *retiredGlyph, *flameCache, GetServer
				t.Cleanup(func() {
					*glyphCache, *retiredGlyph, *flameCache, GetServer = previousGlyph, previousRetired, previousFlame, previousServer
				})
				*glyphCache, *retiredGlyph, *flameCache = 321, 123, 0xffffffff
				GetServer = func() Server { return &burnCastLegacyServer52C3E0{srv: srv} }
				beforeOwner, beforeCaster, beforeUpdate, beforeArg := *owner, *caster, *update, *arg
				if got := Nox_xxx_castBurn_52C3E0(spell.SPELL_BURN, nil, owner, caster, arg, level); got != 1 ||
					*glyphCache != 321 || *retiredGlyph != 123 || *flameCache != 0xffffffff ||
					*owner != beforeOwner || *caster != beforeCaster || *update != beforeUpdate || *arg != beforeArg {
					t.Fatalf("native missing-flame result/glyph/blob/flame=%d/%08x/%08x/%08x", got, *glyphCache, *retiredGlyph, *flameCache)
				}
			})
		}
	}
}

func TestBurnCastNativeNilGuards52C3E0(t *testing.T) {
	for _, missing := range []string{"argument", "caster", "both"} {
		t.Run(missing, func(t *testing.T) {
			caster, freeCaster := alloc.New(server.Object{})
			t.Cleanup(freeCaster)
			arg, freeArg := alloc.New(server.SpellAcceptArg{})
			t.Cleanup(freeArg)
			switch missing {
			case "argument":
				arg = nil
			case "caster":
				caster = nil
			case "both":
				caster, arg = nil, nil
			}
			glyphCache := Get_dword_5d4594_2487712_ptr()
			flameCache := memmap.PtrUint32(0x5D4594, 2487732)
			previousGlyph, previousFlame, previousServer := *glyphCache, *flameCache, GetServer
			t.Cleanup(func() { *glyphCache, *flameCache, GetServer = previousGlyph, previousFlame, previousServer })
			*glyphCache, *flameCache = 321, 654
			GetServer = func() Server { return &burnCastLegacyServer52C3E0{srv: new(server.Server)} }
			if got := Nox_xxx_castBurn_52C3E0(spell.SPELL_BURN, nil, nil, caster, arg, 1); got != 0 || *glyphCache != 321 || *flameCache != 654 {
				t.Fatalf("guard result/cache=%d/%d/%d", got, *glyphCache, *flameCache)
			}
		})
	}
}
