package opennox

import (
	"bytes"
	"encoding/binary"
	"net/netip"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/player"
	"github.com/opennox/libs/strman"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestGameExPacketBoundsR270(t *testing.T) {
	s := server.New(nil, nil, strman.New())
	defer s.Close()
	previous, previousFlags := noxServer, *legacy.Get_gameex_flags_ptr()
	previousGame := noxflags.GetGame()
	noxServer = &Server{Server: s}
	noxflags.ResetGame()
	noxflags.SetGame(noxflags.GameHost)
	defer func() {
		noxServer = previous
		*legacy.Get_gameex_flags_ptr() = previousFlags
		noxflags.ResetGame()
		noxflags.SetGame(previousGame)
	}()
	units, freeUnits := alloc.Make([]server.Object{}, 4)
	defer freeUnits()
	updates, freeUpdates := alloc.Make([]server.PlayerUpdateData{}, 4)
	defer freeUpdates()
	p := s.Players.NewRaw(0x1234)
	if p == nil {
		t.Fatal("private player slot unavailable")
	}
	for i := range units {
		units[i].NetCode = uint32(0xf1234000) + uint32(i)
		units[i].ObjClass = object.ClassPlayer
		units[i].UpdateData = unsafe.Pointer(&updates[i])
		if i+1 < len(units) {
			units[i].ObjNext = &units[i+1]
		}
	}
	units[0].ObjClass = object.ClassMonster
	updates[0].Player = p
	units[1].UpdateData = nil
	updates[3].Player = p
	s.Objs.Pending = &units[0]
	defer func() { s.Objs.Pending = nil }()
	if unsafe.Sizeof(uintptr(0)) > 4 && (uintptr(units[0].CObj()) <= 0xffffffff || uintptr(p.C()) <= 0xffffffff) {
		t.Fatal("private object/player pointers must be native-width")
	}
	packet := func(op uint16, code uint32, direction byte) []byte {
		data := make([]byte, 9)
		binary.LittleEndian.PutUint16(data, 0xf13a)
		binary.LittleEndian.PutUint16(data[2:], op)
		binary.LittleEndian.PutUint32(data[4:], code)
		data[8] = direction
		return data
	}
	check := func(t *testing.T, data []byte) {
		t.Helper()
		before := bytes.Clone(data)
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("extension % x panicked: %v", data, r)
				}
			}()
			MixRecvFromReplacer(nil, data, netip.MustParseAddrPort("127.0.0.1:12345"))
		}()
		if !bytes.Equal(data, before) {
			t.Error("extension handler rewrote the received packet")
		}
	}
	t.Run("short_header", func(t *testing.T) {
		*legacy.Get_gameex_flags_ptr() = 0x3f
		data := packet(0, units[3].NetCode, 0)
		for size := 0; size < 4; size++ {
			check(t, data[:size])
		}
	})
	t.Run("short_scroll", func(t *testing.T) {
		*legacy.Get_gameex_flags_ptr() = 0x3f
		data := packet(0, units[3].NetCode, 0)
		for size := 4; size < 9; size++ {
			check(t, data[:size])
		}
	})
	t.Run("short_trap", func(t *testing.T) {
		*legacy.Get_gameex_flags_ptr() = 0x3f
		data := packet(9, units[3].NetCode, 0)
		for size := 4; size < 8; size++ {
			check(t, data[:size])
		}
	})
	t.Run("invalid_player_reference", func(t *testing.T) {
		*legacy.Get_gameex_flags_ptr() = 0x3f
		for _, code := range []uint32{0, 0xffffffff, units[0].NetCode, units[1].NetCode, units[2].NetCode} {
			for _, op := range []uint16{0, 9} {
				for _, direction := range []byte{0, 0x10, 0xff} {
					check(t, packet(op, code, direction))
				}
			}
		}
	})
	t.Run("complete_empty_inventory_commands", func(t *testing.T) {
		unitBytes := unsafe.Slice((*byte)(unsafe.Pointer(&units[0])), len(units)*int(unsafe.Sizeof(units[0])))
		updateBytes := unsafe.Slice((*byte)(unsafe.Pointer(&updates[0])), len(updates)*int(unsafe.Sizeof(updates[0])))
		beforeUnit, beforeUpdate := bytes.Clone(unitBytes), bytes.Clone(updateBytes)
		for _, enabled := range []uint32{0, 8} {
			*legacy.Get_gameex_flags_ptr() = enabled
			for _, class := range []player.Class{player.Warrior, player.Wizard, player.Conjurer} {
				p.Info().SetPlayerClass(class)
				for direction := 0; direction < 256; direction++ {
					check(t, packet(0, units[3].NetCode, byte(direction)))
					check(t, packet(9, units[3].NetCode, byte(direction))[:8])
				}
			}
		}
		if !bytes.Equal(unitBytes, beforeUnit) || !bytes.Equal(updateBytes, beforeUpdate) {
			t.Fatal("complete empty-inventory commands changed native object/update data")
		}
	})
}
