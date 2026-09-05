package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

type spellAcceptLegacyServer4FD400 struct {
	Server
	spellID               spell.ID
	second, third, fourth *server.Object
	arg                   *server.SpellAcceptArg
	level                 int32
	result                int32
}

func (s *spellAcceptLegacyServer4FD400) SpellAccept4FD400(
	spellID spell.ID,
	second, third, fourth *server.Object,
	arg *server.SpellAcceptArg,
	level int32,
) int32 {
	s.spellID = spellID
	s.second, s.third, s.fourth = second, third, fourth
	s.arg = arg
	s.level = level
	return s.result
}

func TestSpellAcceptExport4FD400PreservesNativePointersAndSignedDwords(t *testing.T) {
	fake := &spellAcceptLegacyServer4FD400{result: math.MinInt32 + 0x13579}
	oldGetServer := GetServer
	GetServer = func() Server { return fake }
	t.Cleanup(func() { GetServer = oldGetServer })

	second := new(server.Object)
	third := new(server.Object)
	fourth := new(server.Object)
	target := new(server.Object)
	arg := &server.SpellAcceptArg{
		Obj: target,
		Pos: types.Pointf{X: -123.5, Y: 456.25},
	}

	var pin runtime.Pinner
	pin.Pin(second)
	pin.Pin(third)
	pin.Pin(fourth)
	pin.Pin(target)
	pin.Pin(arg)
	defer pin.Unpin()

	if got, want := spellAcceptArgCSize4FD400(), unsafe.Sizeof(*arg); got != want {
		t.Fatalf("C/Go SpellAcceptArg sizes = %d/%d", got, want)
	}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for name, ptr := range map[string]uintptr{
			"second": uintptr(unsafe.Pointer(second)),
			"third":  uintptr(unsafe.Pointer(third)),
			"fourth": uintptr(unsafe.Pointer(fourth)),
			"target": uintptr(unsafe.Pointer(target)),
			"arg":    uintptr(unsafe.Pointer(arg)),
		} {
			if ptr <= math.MaxUint32 {
				t.Fatalf("%s pointer = %#x, want native address above 4 GiB", name, ptr)
			}
		}
	}

	const spellID = int32(math.MinInt32 + 0x2468)
	const level = int32(math.MaxInt32)
	if got := spellAcceptExportCall4FD400(spellID, second, third, fourth, arg, level); got != fake.result {
		t.Fatalf("export result = %d, want %d", got, fake.result)
	}
	if fake.spellID != spell.ID(spellID) || fake.second != second || fake.third != third || fake.fourth != fourth || fake.arg != arg || fake.level != level {
		t.Fatalf("export call = %d/%p/%p/%p/%p/%d", fake.spellID, fake.second, fake.third, fake.fourth, fake.arg, fake.level)
	}
	if fake.arg.Obj != target || fake.arg.Pos != (types.Pointf{X: -123.5, Y: 456.25}) {
		t.Fatalf("export arg = %p/%v, want %p/(-123.5,456.25)", fake.arg.Obj, fake.arg.Pos, target)
	}

	fake.result = math.MaxInt32
	if got := spellAcceptExportCall4FD400(math.MaxInt32, nil, nil, nil, nil, math.MinInt32); got != math.MaxInt32 {
		t.Fatalf("nil-pointer export result = %d, want %d", got, int32(math.MaxInt32))
	}
	if fake.spellID != spell.ID(math.MaxInt32) || fake.second != nil || fake.third != nil || fake.fourth != nil || fake.arg != nil || fake.level != math.MinInt32 {
		t.Fatalf("nil-pointer export call = %d/%p/%p/%p/%p/%d", fake.spellID, fake.second, fake.third, fake.fourth, fake.arg, fake.level)
	}

	runtime.KeepAlive(second)
	runtime.KeepAlive(third)
	runtime.KeepAlive(fourth)
	runtime.KeepAlive(target)
	runtime.KeepAlive(arg)
}
