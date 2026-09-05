package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

type spellBookInsertLegacyServer4FE340 struct {
	Server
	unit     *server.Object
	sequence *int32
	spells   [5]int32
	args     [3]int32
	result   int32
}

func (s *spellBookInsertLegacyServer4FE340) SpellBookInsert4FE340(
	unit *server.Object,
	sequence *int32,
	count, delay, targetMode int32,
) int32 {
	s.unit = unit
	s.sequence = sequence
	if sequence != nil {
		copy(s.spells[:], unsafe.Slice(sequence, len(s.spells)))
	}
	s.args = [3]int32{count, delay, targetMode}
	return s.result
}

func TestSpellBookInsertExport4FE340PreservesNativePointersAndSignedDwords(t *testing.T) {
	fake := &spellBookInsertLegacyServer4FE340{result: math.MinInt32 + 0x13579}
	oldGetServer := GetServer
	GetServer = func() Server { return fake }
	t.Cleanup(func() { GetServer = oldGetServer })

	unit := new(server.Object)
	sequence := [5]int32{math.MinInt32, -1, 0, 0x76543210, math.MaxInt32}
	var pin runtime.Pinner
	pin.Pin(unit)
	pin.Pin(&sequence[0])
	defer pin.Unpin()
	if unsafe.Sizeof(uintptr(0)) == 8 {
		if uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 {
			t.Fatalf("unit pointer = %p, want native address above 4 GiB", unit)
		}
		if uintptr(unsafe.Pointer(&sequence[0])) <= math.MaxUint32 {
			t.Fatalf("sequence pointer = %p, want native address above 4 GiB", &sequence[0])
		}
	}

	wantArgs := [3]int32{math.MinInt32, -0x1234567, math.MaxInt32}
	if got := spellBookInsertExportCall4FE340(
		unit,
		&sequence[0],
		wantArgs[0],
		wantArgs[1],
		wantArgs[2],
	); got != fake.result {
		t.Fatalf("export result = %d, want %d", got, fake.result)
	}
	if fake.unit != unit || fake.sequence != &sequence[0] {
		t.Fatalf("export pointers = %p/%p, want %p/%p", fake.unit, fake.sequence, unit, &sequence[0])
	}
	if fake.spells != sequence || fake.args != wantArgs {
		t.Fatalf("export scalars = %v/%v, want %v/%v", fake.spells, fake.args, sequence, wantArgs)
	}
	runtime.KeepAlive(unit)
	runtime.KeepAlive(sequence)
}
