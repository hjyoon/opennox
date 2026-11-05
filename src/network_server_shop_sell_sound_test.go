package opennox

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

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
			s.Audio.Init(s.Server)
			t.Cleanup(s.Audio.Free)
			// Both descriptors are enabled so choosing the wrong effect cannot
			// hide behind an unavailable sound in this binding regression.
			wire := binary.LittleEndian.AppendUint32(nil, 2)
			for _, id := range []string{"TreasurePickup", "InventoryPickup"} {
				wire = append(wire, byte(len(id)))
				wire = append(wire, id...)
				wire = append(wire, 8, 0, 0, 40, 0, 0, 0, 0, 0, 1, 's', 0)
			}
			file := luaCompatAudioFile{bytes.NewReader(wire)}
			if !s.Audio.ReadAUD(file) || file.Len() != 0 {
				t.Fatal("real audio descriptor parser")
			}
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
