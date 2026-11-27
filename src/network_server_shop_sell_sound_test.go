package opennox

import (
	"bytes"
	"encoding/binary"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/server"
)

func TestShopSellProductionOriginalSound510BE0(t *testing.T) {
	for _, queued := range []bool{false, true} {
		name := "during-audio"
		if queued {
			name = "queued-before-audio"
		}
		t.Run(name, func(t *testing.T) {
			s, unit, _, _ := basicSenderProductionFixture5187E0(t)
			*unit = server.Object{ObjClass: object.ClassPlayer, NetCode: 0xa5b61234, PosVec: types.Ptf(300, 320)}
			// Both descriptors are enabled so choosing the wrong effect cannot
			// hide behind an unavailable sound in this binding regression.
			loadShopTestSounds(t, s, "TreasurePickup", "InventoryPickup")
			if !queued {
				s.Audio.Reset()
			}
			s.shopSellRuntime5109C0().PlaySellSound(unit)
			var events []server.AudioEvent
			s.Audio.EachEvent(func(event *server.AudioEvent) { events = append(events, *event) })
			if queued {
				if len(events) != 0 {
					t.Fatal("queued sale event arrived before the audio phase")
				}
				s.Audio.Reset()
				s.Audio.EachEvent(func(event *server.AudioEvent) { events = append(events, *event) })
			}
			if len(events) != 1 {
				t.Fatalf("sale audio events = %d, want one", len(events))
			}
			// Both original 00510BE0 and 00510D10 pass literal sound 307,
			// audience kind2 and the seller's complete net-code dword.
			if event := events[0]; event.Sound != sound.ID(307) {
				t.Fatalf("sale sound = %d (%s), want original TreasurePickup 307", event.Sound, event.Sound)
			} else if event.Obj != unit || event.Pos != unit.PosVec || event.Kind != 2 || event.Code != unit.NetCode {
				t.Fatalf("sale audience = object %p position %v kind %d code %#x", event.Obj, event.Pos, event.Kind, event.Code)
			}
		})
	}
}

func loadShopTestSounds(t *testing.T, s *Server, ids ...string) {
	t.Helper()
	s.Audio.Init(s.Server)
	t.Cleanup(s.Audio.Free)
	wire := binary.LittleEndian.AppendUint32(nil, uint32(len(ids)))
	for _, id := range ids {
		wire = append(wire, byte(len(id)))
		wire = append(wire, id...)
		wire = append(wire, 8, 0, 0, 40, 0, 0, 0, 0, 0, 1, 's', 0)
	}
	file := luaCompatAudioFile{bytes.NewReader(wire)}
	if !s.Audio.ReadAUD(file) || file.Len() != 0 {
		t.Fatal("real audio descriptor parser")
	}
}

func TestShopNativeProductionSoundReturnsImmediately(t *testing.T) {
	oldFlags := noxflags.GetGame()
	noxflags.ResetGame()
	t.Cleanup(func() { noxflags.ResetGame(); noxflags.SetGame(oldFlags) })
	s, unit, _, _ := basicSenderProductionFixture5187E0(t)
	*unit = server.Object{ObjClass: object.ClassPlayer, NetCode: 0xa5b61234, PosVec: types.Ptf(300, 320)}
	loadShopTestSounds(t, s, "TreasurePickup", "InventoryPickup", "ShopRepairItem", "NoCanDo")

	if got := s.Audio.EventObjNative501960(sound.SoundInventoryPickup, unit, 2, unit.NetCode); got != 0 {
		t.Fatalf("initial sound returned %#x", got)
	}
	head := func() *server.AudioEvent {
		var first *server.AudioEvent
		s.Audio.EachEvent(func(event *server.AudioEvent) {
			if first == nil {
				first = event
			}
		})
		return first
	}
	for _, tc := range []struct {
		sound sound.ID
		play  func(*server.Object) uintptr
		kind  int
		code  uint32
	}{
		{sound.SoundTreasurePickup, s.Game4ShopSellRuntime().PlaySellSound, 2, unit.NetCode},
		{sound.SoundShopRepairItem, s.Game4ShopRepairRuntime().PlayRepairSound, 2, unit.NetCode},
		{sound.SoundNoCanDo, s.Game4ShopRepairRuntime().PlayRejectSound, 0, 0},
		{sound.SoundNoCanDo, s.Game4ShopBuyRuntime().PlayRejectSound, 0, 0},
	} {
		previous := head()
		if got := tc.play(unit); previous == nil || got != uintptr(unsafe.Pointer(previous)) {
			t.Fatalf("sound %d returned %#x, want live head %p", tc.sound, got, previous)
		}
		event := head()
		if event == previous || event.Sound != tc.sound || event.Obj != unit || event.Pos != unit.PosVec || event.Kind != tc.kind || event.Code != tc.code {
			t.Fatalf("sound %d inserted %#v", tc.sound, event)
		}
	}
}
