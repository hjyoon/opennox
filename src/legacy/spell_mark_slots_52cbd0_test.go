package legacy

import (
	"fmt"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestMarkSlotPublicSelector52CBD0KeepsSixArguments(t *testing.T) {
	second, freeSecond := alloc.New(server.Object{})
	caster, freeCaster := alloc.New(server.Object{})
	aim, freeAim := alloc.New(server.Object{})
	arg, freeArg := alloc.New(server.SpellAcceptArg{})
	for _, free := range []func(){freeSecond, freeCaster, freeAim, freeArg} {
		t.Cleanup(free)
	}
	*second, *caster, *aim, *arg = server.Object{}, server.Object{}, server.Object{}, server.SpellAcceptArg{}
	previous := markSlotCastCall52CBD0
	t.Cleanup(func() { markSlotCastCall52CBD0 = previous })
	for _, level := range []int{1, 2, 3, 4, 5, 0, -3, math.MinInt32, math.MaxInt32} {
		t.Run(fmt.Sprint(level), func(t *testing.T) {
			calls := 0
			markSlotCastCall52CBD0 = func(id spell.ID, a2, a3, a4 *server.Object, sa *server.SpellAcceptArg, power int) int {
				calls++
				if id != spell.ID(math.MinInt32) || a2 != second || a3 != caster || a4 != aim || sa != arg || power != level {
					t.Fatal("Mark slot selector narrowed/reordered arguments")
				}
				return math.MinInt32
			}
			if got := Sub_52CBD0(spell.ID(math.MinInt32), second, caster, aim, arg, level); got != math.MinInt32 || calls != 1 {
				t.Fatalf("result/calls=%d/%d", got, calls)
			}
		})
	}
}

func TestMarkSlotNative52CBD0SharedCacheAndMoveBinding(t *testing.T) {
	for index := 0; index < 4; index++ {
		for _, mode := range []string{"monster", "missing-type", "move", "immobile"} {
			t.Run(fmt.Sprintf("slot-%d/%s", index, mode), func(t *testing.T) {
				caster, freeCaster := alloc.New(server.Object{})
				aim, freeAim := alloc.New(server.Object{})
				marker, freeMarker := alloc.New(server.Object{})
				data, freeData := alloc.New(server.PlayerUpdateData{})
				for _, free := range []func(){freeCaster, freeAim, freeMarker, freeData} {
					t.Cleanup(free)
				}
				*data = server.PlayerUpdateData{Field39: 0x10203040}
				*caster = server.Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(data), PosVec: types.Ptf(300, 500)}
				*aim = server.Object{TypeInd: 321, PosVec: types.Ptf(330, 520)}
				*marker = server.Object{Field34: 99, PosVec: types.Ptf(100, 200)}
				if unsafe.Sizeof(uintptr(0)) == 8 {
					for _, ptr := range []unsafe.Pointer{unsafe.Pointer(caster), unsafe.Pointer(aim), unsafe.Pointer(marker), unsafe.Pointer(data)} {
						if uintptr(ptr) <= math.MaxUint32 {
							t.Fatalf("Mark slot native pointer=%p want >4 GiB", ptr)
						}
					}
				}
				if mode == "monster" {
					caster.ObjClass, caster.UpdateData = object.ClassMonster, nil
				} else if mode == "move" || mode == "immobile" {
					data.Field29[index] = marker
					if mode == "immobile" {
						marker.ObjClass = object.ClassImmobile
					}
				}
				cache, retired := Get_dword_5d4594_2487712_ptr(), memmap.PtrUint32(0x5D4594, 2487712)
				if cache == retired {
					t.Fatal("Mark slot must use shared live Glyph cache, not retired PE blob")
				}
				previousCache, previousRetired, previousServer, previousMove := *cache, *retired, GetServer, Nox_xxx_unitMove_4E7010
				t.Cleanup(func() {
					*cache, *retired, GetServer, Nox_xxx_unitMove_4E7010 = previousCache, previousRetired, previousServer, previousMove
				})
				*cache, *retired = 321, 654
				srv := new(server.Server)
				srv.SetFrame(100)
				GetServer = func() Server { return &markCastLegacyServer52CA80{srv: srv} }
				calls := 0
				Nox_xxx_unitMove_4E7010 = func(actual *server.Object, point types.Pointf) {
					calls++
					if actual != marker || point != aim.PosVec {
						t.Fatal("Mark slot movement lost native identity or Glyph coordinates")
					}
					actual.PosVec = point
				}
				before := *data
				if got := Sub_52CBD0(spell.ID(46+index), nil, caster, aim, nil, math.MinInt32); got != 1 || *cache != 321 || *retired != 654 {
					t.Fatalf("result/cache/blob=%d/%d/%d", got, *cache, *retired)
				}
				if mode == "monster" || mode == "missing-type" {
					if calls != 0 || *data != before {
						t.Fatal("non-player/missing allocation touched marker or charges")
					}
					return
				}
				shift := uint(index * 8)
				wantCharges := uint32(0x10203040)&^(uint32(0xff)<<shift) | uint32(3)<<shift
				wantCalls, wantPoint := 1, aim.PosVec
				if mode == "immobile" {
					wantCalls, wantPoint = 0, types.Ptf(100, 200)
				}
				if calls != wantCalls || marker.PosVec != wantPoint || marker.Field34 != 100 || data.Field39 != wantCharges {
					t.Fatalf("move/position/stamp/charges=%d/%v/%d/%08x", calls, marker.PosVec, marker.Field34, data.Field39)
				}
			})
		}
	}
}
