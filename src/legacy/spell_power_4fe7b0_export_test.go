package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/spell"

	"github.com/opennox/opennox/v1/server"
)

type spellPowerLegacyServer4FE7B0 struct {
	Server
	spellID spell.ID
	caster  *server.Object
	result  int32
	calls   int
}

func (s *spellPowerLegacyServer4FE7B0) SpellPower4FE7B0(
	spellID spell.ID,
	caster *server.Object,
) int32 {
	s.spellID = spellID
	s.caster = caster
	s.calls++
	return s.result
}

func TestSpellPowerExport4FE7B0PreservesNativePointerAndSignedDwords(t *testing.T) {
	fake := &spellPowerLegacyServer4FE7B0{result: math.MinInt32}
	oldGetServer := GetServer
	GetServer = func() Server { return fake }
	t.Cleanup(func() { GetServer = oldGetServer })

	caster := new(server.Object)
	var pin runtime.Pinner
	pin.Pin(caster)
	defer pin.Unpin()
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(caster)) <= math.MaxUint32 {
		t.Fatalf("caster pointer = %p, want native address above 4 GiB", caster)
	}

	const spellID = int32(-0x1234567)
	if got := spellPowerExportCall4FE7B0(spellID, caster); got != math.MinInt32 {
		t.Fatalf("CGo export result = %#x, want %#x", got, int32(math.MinInt32))
	}
	if fake.calls != 1 || int32(fake.spellID) != spellID || fake.caster != caster {
		t.Fatalf("export calls/id/caster = %d/%d/%p, want 1/%d/%p", fake.calls, fake.spellID, fake.caster, spellID, caster)
	}

	fake.result = math.MaxInt32
	if got := Nox_xxx_spellGetPower_4FE7B0(spell.ID(spellID), caster); got != math.MaxInt32 {
		t.Fatalf("direct legacy result = %#x, want %#x", got, int32(math.MaxInt32))
	}
	if fake.calls != 2 || int32(fake.spellID) != spellID || fake.caster != caster {
		t.Fatalf("direct calls/id/caster = %d/%d/%p, want 2/%d/%p", fake.calls, fake.spellID, fake.caster, spellID, caster)
	}
	runtime.KeepAlive(caster)
}
