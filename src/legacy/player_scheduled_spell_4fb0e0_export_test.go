package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/spell"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/server"
)

type playerScheduledSpellLegacyServer4FB0E0 struct {
	Server
	srv *server.Server
}

func (s *playerScheduledSpellLegacyServer4FB0E0) S() *server.Server {
	return s.srv
}

func TestPlayerScheduledSpellExports4FB0E0PreserveNativePointers(t *testing.T) {
	oldFlags := noxflags.GetGame()
	noxflags.ResetGame()
	t.Cleanup(func() {
		noxflags.ResetGame()
		noxflags.SetGame(oldFlags)
	})

	srv := new(server.Server)
	oldGetServer := GetServer
	GetServer = func() Server { return &playerScheduledSpellLegacyServer4FB0E0{srv: srv} }
	t.Cleanup(func() { GetServer = oldGetServer })

	oldCast := Nox_xxx_castSpellByUser_4FDD20
	t.Cleanup(func() { Nox_xxx_castSpellByUser_4FDD20 = oldCast })

	player := &server.Player{PlayerInd: 3}
	update := &server.PlayerUpdateData{
		TrapSpells:    [5]uint32{1, 2, 3, 4, 5},
		TrapSpellsCnt: 0xabcdef01,
		Field55:       -123,
		Field56:       456,
		Player:        player,
	}
	unit := &server.Object{UpdateData: unsafe.Pointer(update)}
	target := &server.Object{}

	var castUnit *server.Object
	var castTarget *server.Object
	var castID int
	var castArgAddress uintptr
	Nox_xxx_castSpellByUser_4FDD20 = func(id int, gotUnit *server.Object, rawArg unsafe.Pointer) int {
		arg := (*server.SpellAcceptArg)(rawArg)
		castID = id
		castUnit = gotUnit
		castTarget = arg.Obj
		castArgAddress = uintptr(rawArg)
		if arg.Pos.X != -123 || arg.Pos.Y != 456 {
			t.Fatalf("cast position = %v, want (-123,456)", arg.Pos)
		}
		return 1
	}

	var pin runtime.Pinner
	pin.Pin(player)
	pin.Pin(update)
	pin.Pin(unit)
	pin.Pin(target)
	defer pin.Unpin()

	if got := playerScheduledSpellExportCall4FB0E0(unit, target); got != 1 {
		t.Fatalf("FIFO export result = %d, want 1", got)
	}
	if castID != int(spell.ID(1)) || castUnit != unit || castTarget != target {
		t.Fatalf("FIFO cast = %d/%p/%p, want 1/%p/%p", castID, castUnit, castTarget, unit, target)
	}
	if got, want := update.TrapSpells, ([5]uint32{0, 2, 3, 4, 5}); got != want {
		t.Fatalf("FIFO spells = %v, want %v", got, want)
	}
	if got := update.TrapSpellsCnt; got != 0xabcdef00 {
		t.Fatalf("FIFO count = %#x, want 0xabcdef00", got)
	}

	update.TrapSpells = [5]uint32{6, 7, 8, 9, 10}
	update.TrapSpellsCnt = 0x10203002
	if got := playerScheduledSpellQueueExportCall4FB1D0(unit, target); got != 1 {
		t.Fatalf("LIFO export result = %d, want 1", got)
	}
	if castID != 7 || castUnit != unit || castTarget != target {
		t.Fatalf("LIFO cast = %d/%p/%p, want 7/%p/%p", castID, castUnit, castTarget, unit, target)
	}
	if got, want := update.TrapSpells, ([5]uint32{6, 7, 8, 9, 10}); got != want {
		t.Fatalf("LIFO spells = %v, want %v", got, want)
	}
	if got := update.TrapSpellsCnt; got != 0x10203001 {
		t.Fatalf("LIFO count = %#x, want 0x10203001", got)
	}

	if unsafe.Sizeof(uintptr(0)) == 8 {
		for name, ptr := range map[string]uintptr{
			"unit": uintptr(unsafe.Pointer(unit)), "target": uintptr(unsafe.Pointer(target)),
			"update": uintptr(unsafe.Pointer(update)), "player": uintptr(unsafe.Pointer(player)),
			"spell-arg": castArgAddress,
		} {
			if ptr <= math.MaxUint32 {
				t.Fatalf("%s pointer = %#x, want native address above 4 GiB", name, ptr)
			}
		}
	}
	runtime.KeepAlive(player)
	runtime.KeepAlive(update)
	runtime.KeepAlive(unit)
	runtime.KeepAlive(target)
}
