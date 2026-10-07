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

func TestMarkPublicSelector52CA80KeepsSixArgumentsAndSignedResult(t *testing.T) {
	second, freeSecond := alloc.New(server.Object{})
	caster, freeCaster := alloc.New(server.Object{})
	aim, freeAim := alloc.New(server.Object{})
	arg, freeArg := alloc.New(server.SpellAcceptArg{})
	for _, free := range []func(){freeSecond, freeCaster, freeAim, freeArg} {
		t.Cleanup(free)
	}
	*second, *caster, *aim, *arg = server.Object{}, server.Object{}, server.Object{}, server.SpellAcceptArg{}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, ptr := range []unsafe.Pointer{unsafe.Pointer(second), unsafe.Pointer(caster), unsafe.Pointer(aim), unsafe.Pointer(arg)} {
			if uintptr(ptr) <= math.MaxUint32 {
				t.Fatalf("native Mark dispatch pointer=%p, want >4 GiB", ptr)
			}
		}
	}
	previous := markCastCall52CA80
	t.Cleanup(func() { markCastCall52CA80 = previous })
	for _, level := range []int{1, 2, 3, 4, 5, 0, -3, math.MinInt32, math.MaxInt32} {
		t.Run(fmt.Sprint(level), func(t *testing.T) {
			calls := 0
			markCastCall52CA80 = func(id spell.ID, a2, a3, a4 *server.Object, sa *server.SpellAcceptArg, power int) int {
				calls++
				if id != spell.ID(math.MinInt32) || a2 != second || a3 != caster || a4 != aim || sa != arg || power != level {
					t.Fatal("Mark selector narrowed or reordered its six arguments")
				}
				return math.MinInt32
			}
			if got := Sub_52CA80(spell.ID(math.MinInt32), second, caster, aim, arg, level); got != math.MinInt32 || calls != 1 {
				t.Fatalf("selector result/calls=%d/%d", got, calls)
			}
		})
	}
}

func TestMarkNative52CA80SharesLiveGlyphCacheNotRetiredBlob(t *testing.T) {
	for _, class := range []object.Class{object.ClassMonster, object.ClassPlayer} {
		t.Run(fmt.Sprintf("class-%x", uint32(class)), func(t *testing.T) {
			caster, freeCaster := alloc.New(server.Object{})
			t.Cleanup(freeCaster)
			data, freeData := alloc.New(server.PlayerUpdateData{})
			t.Cleanup(freeData)
			*data = server.PlayerUpdateData{Field39: 0x10203040}
			*caster = server.Object{ObjClass: class, UpdateData: unsafe.Pointer(data), PosVec: types.Ptf(300, 500)}
			cache, retired := Get_dword_5d4594_2487712_ptr(), memmap.PtrUint32(0x5D4594, 2487712)
			if cache == retired {
				t.Fatal("live C-owned Glyph cache must not alias the retired PE32 blob")
			}
			previousCache, previousRetired, previousServer := *cache, *retired, GetServer
			t.Cleanup(func() { *cache, *retired, GetServer = previousCache, previousRetired, previousServer })
			*cache, *retired = 321, 654
			GetServer = func() Server { return &markCastLegacyServer52CA80{srv: new(server.Server)} }
			beforeCaster, beforeData := *caster, *data
			// An ordinary unregistered type lookup really returns nil. For
			// players the original sounds and succeeds without refilling;
			// non-players must not inspect their update data at all.
			if got := Sub_52CA80(spell.SPELL_MARK, nil, caster, nil, nil, math.MinInt32); got != 1 ||
				*caster != beforeCaster || *data != beforeData || *cache != 321 || *retired != 654 {
				t.Fatalf("real missing-type result/cache/blob=%d/%d/%d", got, *cache, *retired)
			}
		})
	}
}

func TestMarkNative52CA80MoveBindingAndPackedRecharge(t *testing.T) {
	for _, immobile := range []bool{false, true} {
		t.Run(fmt.Sprintf("immobile-%t", immobile), func(t *testing.T) {
			caster, freeCaster := alloc.New(server.Object{})
			t.Cleanup(freeCaster)
			aim, freeAim := alloc.New(server.Object{})
			t.Cleanup(freeAim)
			data, freeData := alloc.New(server.PlayerUpdateData{})
			t.Cleanup(freeData)
			*data = server.PlayerUpdateData{Field39: 0x10203040}
			*caster = server.Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(data), PosVec: types.Ptf(300, 500)}
			*aim = server.Object{TypeInd: 321, PosVec: types.Ptf(330, 520)}
			for i := range data.Field29 {
				marker, free := alloc.New(server.Object{})
				t.Cleanup(free)
				*marker = server.Object{Field34: uint32(i + 20), PosVec: types.Ptf(100, 200)}
				data.Field29[i] = marker
			}
			if immobile {
				data.Field29[0].ObjClass = object.ClassImmobile
			}
			cache := Get_dword_5d4594_2487712_ptr()
			previousCache, previousServer, previousMove := *cache, GetServer, Nox_xxx_unitMove_4E7010
			t.Cleanup(func() { *cache, GetServer, Nox_xxx_unitMove_4E7010 = previousCache, previousServer, previousMove })
			*cache = 321
			srv := new(server.Server)
			srv.SetFrame(100)
			GetServer = func() Server { return &markCastLegacyServer52CA80{srv: srv} }
			calls := 0
			Nox_xxx_unitMove_4E7010 = func(marker *server.Object, point types.Pointf) {
				calls++
				if marker != data.Field29[0] || point != aim.PosVec {
					t.Fatal("native movement bridge changed marker/position")
				}
				marker.PosVec = point
			}
			if got := Sub_52CA80(spell.SPELL_MARK, nil, caster, aim, nil, math.MinInt32); got != 1 || data.Field39 != 0x10203003 || data.Field29[0].Field34 != 100 {
				t.Fatalf("native move result/charges/frame=%d/%08x/%d", got, data.Field39, data.Field29[0].Field34)
			}
			wantCalls, wantPoint := 1, aim.PosVec
			if immobile {
				wantCalls, wantPoint = 0, types.Ptf(100, 200)
			}
			if calls != wantCalls || data.Field29[0].PosVec != wantPoint || *cache != 321 {
				t.Fatalf("native immobile gate calls/point/cache=%d/%v/%d", calls, data.Field29[0].PosVec, *cache)
			}
		})
	}
}
