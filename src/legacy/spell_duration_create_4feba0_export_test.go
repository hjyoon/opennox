package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

type spellDurationCreateLegacyServer4FEBA0 struct {
	Server
	spellID               int32
	second, third, fourth *server.Object
	arg                   *server.SpellAcceptArg
	level                 int32
	create                unsafe.Pointer
	update                unsafe.Pointer
	destroy               unsafe.Pointer
	duration              int32
	result                int32
}

func (s *spellDurationCreateLegacyServer4FEBA0) SpellDurationCreate4FEBA0(
	spellID int32,
	second, third, fourth *server.Object,
	arg *server.SpellAcceptArg,
	level int32,
	create, update, destroy unsafe.Pointer,
	duration int32,
) int32 {
	s.spellID = spellID
	s.second, s.third, s.fourth = second, third, fourth
	s.arg = arg
	s.level = level
	s.create, s.update, s.destroy = create, update, destroy
	s.duration = duration
	return s.result
}

func TestSpellDurationCreateExport4FEBA0PreservesNativePointersAndSignedDwords(t *testing.T) {
	fake := &spellDurationCreateLegacyServer4FEBA0{result: math.MinInt32 + 0x13579}
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
	create := new(byte)
	update := new(byte)
	destroy := new(byte)

	var pin runtime.Pinner
	pin.Pin(second)
	pin.Pin(third)
	pin.Pin(fourth)
	pin.Pin(target)
	pin.Pin(arg)
	pin.Pin(create)
	pin.Pin(update)
	pin.Pin(destroy)
	defer pin.Unpin()

	if got, want := spellDurationCreateArgCSize4FEBA0(), unsafe.Sizeof(*arg); got != want {
		t.Fatalf("C/Go SpellAcceptArg sizes = %d/%d", got, want)
	}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for name, ptr := range map[string]uintptr{
			"second":  uintptr(unsafe.Pointer(second)),
			"third":   uintptr(unsafe.Pointer(third)),
			"fourth":  uintptr(unsafe.Pointer(fourth)),
			"target":  uintptr(unsafe.Pointer(target)),
			"arg":     uintptr(unsafe.Pointer(arg)),
			"create":  uintptr(unsafe.Pointer(create)),
			"update":  uintptr(unsafe.Pointer(update)),
			"destroy": uintptr(unsafe.Pointer(destroy)),
		} {
			if ptr <= math.MaxUint32 {
				t.Fatalf("%s pointer = %#x, want native address above 4 GiB", name, ptr)
			}
		}
	}

	const spellID = int32(math.MinInt32 + 0x2468)
	const level = int32(math.MaxInt32)
	const duration = int32(math.MinInt32)
	if got := spellDurationCreateExportCall4FEBA0(
		spellID,
		second,
		third,
		fourth,
		arg,
		level,
		unsafe.Pointer(create),
		unsafe.Pointer(update),
		unsafe.Pointer(destroy),
		duration,
	); got != fake.result {
		t.Fatalf("export result = %d, want %d", got, fake.result)
	}
	if fake.spellID != spellID || fake.second != second || fake.third != third || fake.fourth != fourth || fake.arg != arg || fake.level != level {
		t.Fatalf("export object call = %d/%p/%p/%p/%p/%d", fake.spellID, fake.second, fake.third, fake.fourth, fake.arg, fake.level)
	}
	if fake.create != unsafe.Pointer(create) || fake.update != unsafe.Pointer(update) || fake.destroy != unsafe.Pointer(destroy) || fake.duration != duration {
		t.Fatalf("export callback call = %p/%p/%p/%d", fake.create, fake.update, fake.destroy, fake.duration)
	}
	if fake.arg.Obj != target || fake.arg.Pos != (types.Pointf{X: -123.5, Y: 456.25}) {
		t.Fatalf("export arg = %p/%v, want %p/(-123.5,456.25)", fake.arg.Obj, fake.arg.Pos, target)
	}

	fake.result = math.MaxInt32
	if got := spellDurationCreateExportCall4FEBA0(
		math.MaxInt32,
		nil,
		nil,
		nil,
		nil,
		math.MinInt32,
		nil,
		nil,
		nil,
		math.MaxInt32,
	); got != math.MaxInt32 {
		t.Fatalf("nil-pointer export result = %d, want %d", got, int32(math.MaxInt32))
	}
	if fake.spellID != math.MaxInt32 || fake.second != nil || fake.third != nil || fake.fourth != nil || fake.arg != nil || fake.level != math.MinInt32 || fake.create != nil || fake.update != nil || fake.destroy != nil || fake.duration != math.MaxInt32 {
		t.Fatalf("nil-pointer export call = %d/%p/%p/%p/%p/%d/%p/%p/%p/%d", fake.spellID, fake.second, fake.third, fake.fourth, fake.arg, fake.level, fake.create, fake.update, fake.destroy, fake.duration)
	}

	runtime.KeepAlive(second)
	runtime.KeepAlive(third)
	runtime.KeepAlive(fourth)
	runtime.KeepAlive(target)
	runtime.KeepAlive(arg)
	runtime.KeepAlive(create)
	runtime.KeepAlive(update)
	runtime.KeepAlive(destroy)
}
