package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/spell"

	"github.com/opennox/opennox/v1/server"
)

type spellGrantLegacyServer4FB550 struct {
	Server
	unit   *server.Object
	args   [4]int32
	result int32
}

func (s *spellGrantLegacyServer4FB550) SpellGrantToPlayer4FB550(
	unit *server.Object,
	spellID, notify, shop, override int32,
) int32 {
	s.unit = unit
	s.args = [4]int32{spellID, notify, shop, override}
	return s.result
}

func TestSpellGrantExport4FB550PreservesNativePointerAndScalars(t *testing.T) {
	fake := &spellGrantLegacyServer4FB550{result: -0x2345678}
	oldGetServer := GetServer
	GetServer = func() Server { return fake }
	t.Cleanup(func() { GetServer = oldGetServer })

	unit := new(server.Object)
	var pin runtime.Pinner
	pin.Pin(unit)
	defer pin.Unpin()
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 {
		t.Fatalf("unit pointer = %p, want native address above 4 GiB", unit)
	}

	wantArgs := [4]int32{
		0x76543210,
		-0x1234567,
		0x10203040,
		math.MinInt32 + 0x1234,
	}
	if got := spellGrantExportCall4FB550(
		unit,
		wantArgs[0],
		wantArgs[1],
		wantArgs[2],
		wantArgs[3],
	); got != fake.result {
		t.Fatalf("export result = %d, want %d", got, fake.result)
	}
	if fake.unit != unit || fake.args != wantArgs {
		t.Fatalf("export call = %p/%v, want %p/%v", fake.unit, fake.args, unit, wantArgs)
	}

	fake.result = 0x1234567
	wantArgs = [4]int32{10, -2, 3, -4}
	if got := Nox_xxx_spellGrantToPlayer_4FB550(unit, spell.ID(wantArgs[0]), int(wantArgs[1]), int(wantArgs[2]), int(wantArgs[3])); got != int(fake.result) {
		t.Fatalf("Go wrapper result = %d, want %d", got, fake.result)
	}
	if fake.unit != unit || fake.args != wantArgs {
		t.Fatalf("Go wrapper call = %p/%v, want %p/%v", fake.unit, fake.args, unit, wantArgs)
	}
	runtime.KeepAlive(unit)
}
