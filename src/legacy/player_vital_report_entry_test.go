package legacy

import (
	"bytes"
	_ "embed"
	"encoding/hex"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/player"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
	"github.com/opennox/opennox/v1/server"
)

//go:embed testdata/vital_report_entry_original_v2.json.gz
var vitalReportOriginal []byte

type vitalReportCase struct {
	Input struct {
		Entry, Name               string
		Recipient                 int32
		UnitClass                 uint32 `json:"unit_class"`
		PlayerClass               uint8  `json:"player_class"`
		Stamina                   uint8
		Enchantment, Gold, Extent uint32
		NetCode                   uint32 `json:"net_code"`
		Current, Maximum          uint16
		NullUnit                  bool `json:"null_unit"`
		NullUpdate                bool `json:"null_update"`
		NullPlayer                bool `json:"null_player"`
	}
	Observed struct {
		ReturnWord  uint32 `json:"return_word"`
		ReturnsUnit bool   `json:"returns_unit"`
		Packet      *struct {
			Kind, Data        string
			Recipient, Remove int32
			Related           uint32
		}
	}
}

func TestVitalReportEntryFrozenOriginalC(t *testing.T) {
	rows := tradeOfferOriginalRows[vitalReportCase](t, vitalReportOriginal,
		"938e0a54d6fb4a5534c98447c6d22c379f2e38bf96de8d7c5323e871eb0ddbc0", 132)
	handles.InitHigh()
	t.Cleanup(handles.Release)
	unit, freeUnit := alloc.NewHigh(server.Object{})
	t.Cleanup(freeUnit)
	update, freeUpdate := alloc.NewHigh(server.PlayerUpdateData{})
	t.Cleanup(freeUpdate)
	info, freeInfo := alloc.NewHigh(server.Player{})
	t.Cleanup(freeInfo)
	if cgoABIPointerSize == 8 {
		for _, ptr := range []unsafe.Pointer{unsafe.Pointer(unit), unsafe.Pointer(update), unsafe.Pointer(info)} {
			if uintptr(ptr) <= 0xffffffff {
				t.Fatalf("vital report fixture address %p is below 4 GiB", ptr)
			}
		}
	}

	world := &server.Server{}
	host := &playerReportPacketServer{world: world}
	oldServer := GetServer
	t.Cleanup(func() { GetServer = oldServer })
	GetServer = func() Server { return host }

	entries := map[string]int{"stamina": 0, "enchantment": 1, "gold": 2, "total-mana": 3, "mana": 4}
	for _, row := range rows {
		t.Run(row.Input.Name, func(t *testing.T) {
			input, expected := row.Input, row.Observed
			entry, known := entries[input.Entry]
			if !known || input.NullUnit {
				t.Fatalf("unexpected stock input %+v", input)
			}
			*unit = server.Object{ObjClass: object.Class(input.UnitClass), Field110: input.Enchantment,
				NetCode: input.NetCode, Extent: input.Extent}
			*update = server.PlayerUpdateData{Stamina: input.Stamina, ManaCur: input.Current, ManaMax: input.Maximum}
			info.GoldVal = input.Gold
			info.Info().SetPlayerClass(player.Class(input.PlayerClass))
			if !input.NullUpdate {
				unit.UpdateData = unsafe.Pointer(update)
			}
			if !input.NullPlayer {
				update.Player = info
			}
			before := *unit
			updateBytes := unsafe.Slice((*byte)(unsafe.Pointer(update)), int(unsafe.Sizeof(*update)))
			updateBefore := bytes.Clone(updateBytes)
			packet := expected.Packet
			var payload []byte
			if packet != nil {
				var err error
				payload, err = hex.DecodeString(packet.Data)
				if err != nil || packet.Related != 0 || packet.Recipient != input.Recipient {
					t.Fatalf("invalid stock packet %+v (%v)", packet, err)
				}
			}
			calls := 0
			world.NetSendPacketXxx = func(recipient int, data []byte, related *server.Object, remove, sequence int) int {
				calls++
				if entry >= 3 || packet == nil || recipient != int(packet.Recipient) || related != nil ||
					remove != int(packet.Remove) || (sequence != 0) != (packet.Kind == "sequenced") ||
					!bytes.Equal(data, payload) {
					t.Errorf("send %x to %d with %p/%d/%d, want %+v", data, recipient, related, remove, sequence, packet)
				}
				return int(int32(expected.ReturnWord))
			}

			if entry >= 3 {
				preserveImportantPacketState(t)
				setImportantNativeListC(nil, nil)
				pool := alloc.NewClassHigh("vital-report-entry", importantPacketSizeC(), 1)
				setImportantAllocClassC(pool.UPtr())
				setImportantRecipientMaskC(1 << 6)
				gameFrameHook = func() uint32 { return 0x12345678 }
				if packet != nil && expected.ReturnWord == 0 {
					if pool.NewObject() == nil {
						t.Fatal("could not occupy the packet pool")
					}
				} else if packet != nil {
					if packet.Kind != "replace" ||
						sendImportantPacketWrapperC(255, []byte{payload[0], 0xee}, nil, 1, importantPacketSequenceDisabled) != 1 {
						t.Fatal("could not seed a pending broadcast to replace")
					}
				}
			}
			want := uintptr(int32(expected.ReturnWord))
			if expected.ReturnsUnit {
				want = uintptr(unsafe.Pointer(unit))
			}
			if got := callVitalReportEntry(entry, input.Recipient, unit); got != want {
				t.Fatalf("native return %#x, want %#x", got, want)
			}
			if *unit != before || !bytes.Equal(updateBytes, updateBefore) ||
				info.GoldVal != input.Gold || byte(info.PlayerClass()) != input.PlayerClass {
				t.Fatal("report modified the source object, update or player")
			}
			if entry < 3 {
				if (packet != nil && calls != 1) || (packet == nil && calls != 0) {
					t.Fatalf("got %d sends for %+v", calls, packet)
				}
			} else {
				first, last := importantNativeListC()
				if packet == nil || expected.ReturnWord == 0 {
					if first != nil || last != nil {
						t.Fatal("a skipped or failed mana report queued a packet")
					}
					return
				}
				if first == nil || first != last || importantPacketNextC(first) != nil {
					t.Fatal("mana report did not replace the pending broadcast")
				}
				got := (*importantPacketLegacy)(first)
				if got.Recipient != int8(packet.Recipient) || got.RemoveIfDisconnected != uint32(packet.Remove) ||
					got.SequenceEnabled != 0 || importantPacketRelatedObjectC(first) != nil ||
					int(got.PayloadSize) != len(payload) || !bytes.Equal(got.Payload[:got.PayloadSize], payload) {
					t.Fatalf("queued mana packet %x for %d, want %+v", got.Payload[:got.PayloadSize], got.Recipient, packet)
				}
			}
		})
	}
}

func TestVitalReportEntryGuardsMissingRecords(t *testing.T) {
	world := &server.Server{}
	host := &playerReportPacketServer{world: world}
	oldServer := GetServer
	t.Cleanup(func() { GetServer = oldServer })
	GetServer = func() Server { return host }
	unit, release := alloc.NewHigh(server.Object{ObjClass: object.ClassPlayer})
	t.Cleanup(release)
	for entry := range 5 {
		if got := callVitalReportEntry(entry, 6, nil); got != 0 {
			t.Fatalf("entry %d returned %#x for a missing unit", entry, got)
		}
		if entry != 1 && callVitalReportEntry(entry, 6, unit) != uintptr(unsafe.Pointer(unit)) {
			t.Fatalf("entry %d lost the address of the player with no update record", entry)
		}
	}
}
