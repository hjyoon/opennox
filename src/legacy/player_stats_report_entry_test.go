package legacy

import (
	"bytes"
	_ "embed"
	"encoding/hex"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
	"github.com/opennox/opennox/v1/server"
)

//go:embed testdata/stock_stats_and_item_report_original_v1.json.gz
var statsAndItemReportOriginal []byte

//go:embed testdata/stock_item_durability_original_v1.json.gz
var itemDurabilityReportOriginal []byte

type stockStatsReportPacket struct {
	Data                        string
	Recipient, Sequence, Remove int32
	Related                     uint32
}

type stockStatsAndItemReportCase struct {
	Input struct {
		Name, Entry            string
		Recipient              int32
		UnitClass              uint32 `json:"unit_class"`
		SenderResult           int64  `json:"sender_result"`
		Code, Extent           uint32
		CurrentHP              uint16 `json:"current_hp"`
		MaxHP                  uint16 `json:"max_hp"`
		MaxMana                uint16 `json:"max_mana"`
		Capacity               uint16
		Speed, Strength, Armor uint32
		Level                  uint8
		Delta                  int16
		NullHealth             bool `json:"null_health"`
		NullUpdate             bool `json:"null_update"`
		NullPlayer             bool `json:"null_player"`
	}
	Observed struct {
		Word          uint32
		ReturnsHealth bool `json:"returns_health"`
		Packet        *stockStatsReportPacket
	}
}

func stockReportQueue(t *testing.T, packet *stockStatsReportPacket, expectedResult uintptr) func() {
	t.Helper()
	preserveImportantPacketState(t)
	setImportantNativeListC(nil, nil)
	pool := alloc.NewClassHigh("stock-stats-report", importantPacketSizeC(), 1)
	setImportantAllocClassC(pool.UPtr())
	setImportantRecipientMaskC(1<<6 | 1<<31 | 1)
	gameFrameHook = func() uint32 { return 0x12345678 }
	if packet != nil && expectedResult == 0 && pool.NewObject() == nil {
		t.Fatal("could not occupy the packet buffer")
	}
	return func() {
		t.Helper()
		first, last := importantNativeListC()
		if packet == nil || expectedResult == 0 {
			if first != nil || last != nil {
				t.Fatal("a skipped or failed report appended a packet")
			}
			return
		}
		want, err := hex.DecodeString(packet.Data)
		if err != nil || first == nil || first != last {
			t.Fatalf("expected one queued packet %x, ends=(%p,%p), decode=%v", want, first, last, err)
		}
		got := (*importantPacketLegacy)(first)
		if got.Recipient != int8(packet.Recipient) || got.SequenceEnabled != byte(packet.Sequence) ||
			got.RemoveIfDisconnected != uint32(packet.Remove) || importantPacketRelatedObjectC(first) != nil ||
			packet.Related != 0 || int(got.PayloadSize) != len(want) || !bytes.Equal(got.Payload[:got.PayloadSize], want) {
			t.Fatalf("queue packet=%x to %d sequence=%d remove=%d, want %+v", got.Payload[:got.PayloadSize],
				got.Recipient, got.SequenceEnabled, got.RemoveIfDisconnected, packet)
		}
	}
}

func TestStatsAndItemReportFrozenOriginalC(t *testing.T) {
	rows := tradeOfferOriginalRows[stockStatsAndItemReportCase](t, statsAndItemReportOriginal,
		"104c9efd67a9fbbe6ddc6d47937a6278635c625c1af92f273a8e14288170169f", 114)
	handles.InitHigh()
	t.Cleanup(handles.Release)
	unit, freeUnit := alloc.NewHigh(server.Object{})
	t.Cleanup(freeUnit)
	health, freeHealth := alloc.NewHigh(server.HealthData{})
	t.Cleanup(freeHealth)
	update, freeUpdate := alloc.NewHigh(server.PlayerUpdateData{})
	t.Cleanup(freeUpdate)
	info, freeInfo := alloc.NewHigh(server.Player{})
	t.Cleanup(freeInfo)
	if cgoABIPointerSize == 8 {
		for _, ptr := range []unsafe.Pointer{unsafe.Pointer(unit), unsafe.Pointer(health), unsafe.Pointer(update), unsafe.Pointer(info)} {
			if uintptr(ptr) <= 0xffffffff {
				t.Fatalf("report fixture pointer %p must be above 4 GiB", ptr)
			}
		}
	}
	entries := map[string]int{"health-delta": 0, "item-health": 1, "player-stats": 2, "armor-value": 3}
	for _, row := range rows {
		input, result := row.Input, row.Observed
		// The real queue accepts valid recipients and produces zero or one.
		if result.Packet != nil && (input.SenderResult < 0 || input.SenderResult > 1 ||
			input.Recipient < 0 || input.Recipient >= 32 && input.Recipient != 255) {
			continue
		}
		t.Run(input.Name, func(t *testing.T) {
			entry, known := entries[input.Entry]
			if !known {
				t.Fatalf("unrecognized stock entry %q", input.Entry)
			}
			*unit = server.Object{ObjClass: object.Class(input.UnitClass), NetCode: input.Code, Extent: input.Extent,
				CarryCapacity: input.Capacity}
			*health = server.HealthData{Cur: input.CurrentHP, Max: input.MaxHP}
			*update = server.PlayerUpdateData{ManaMax: input.MaxMana}
			info.Info().SetField2235(input.Speed)
			info.Info().SetField2239(input.Strength)
			info.Level = input.Level ^ 0xff
			if !input.NullHealth {
				unit.HealthData = health
			}
			if !input.NullUpdate {
				unit.UpdateData = unsafe.Pointer(update)
			}
			if !input.NullPlayer {
				update.Player = info
			}
			before, healthBefore := *unit, *health
			a, b := uint32(input.Level), uint32(0)
			switch entry {
			case 0:
				a, b = input.Code, uint32(input.Delta)
			case 3:
				a = input.Armor
			}
			want := uintptr(result.Word)
			if result.ReturnsHealth {
				want = uintptr(unsafe.Pointer(health))
			}
			checkQueue := stockReportQueue(t, result.Packet, want)
			if got := callStatsReportEntry(entry, input.Recipient, unit, a, b); got != want {
				t.Fatalf("C ABI return=%#x, want %#x", got, want)
			}
			if *unit != before || *health != healthBefore {
				t.Fatal("report modified item or health data")
			}
			checkQueue()
		})
	}
}

type stockDurabilityReportCase struct {
	Input struct {
		Name                   string
		Before, After, Maximum uint16
		Cooperative            bool
		NullHealth             bool `json:"null_health"`
		Recipient              int32
		SenderResult           int32 `json:"sender_result"`
	}
	Observed struct {
		Word          uint32
		ReturnsHealth bool `json:"returns_health"`
		Events        []struct {
			Kind, Data string
			Recipient  int32
		}
	}
}

func TestItemDurabilityReportFrozenOriginalC(t *testing.T) {
	rows := tradeOfferOriginalRows[stockDurabilityReportCase](t, itemDurabilityReportOriginal,
		"d09322370a1419298acf5368e9d9def01b310d660588092ea11ea8144d48e1b8", 36)
	handles.InitHigh()
	t.Cleanup(handles.Release)
	health, freeHealth := alloc.NewHigh(server.HealthData{})
	t.Cleanup(freeHealth)
	item, freeItem := alloc.NewHigh(server.Object{})
	*item = server.Object{ObjClass: 0x01001000, NetCode: 0x27ab, Extent: 0x3179}
	t.Cleanup(freeItem)
	originalFlags := noxflags.GetGame()
	t.Cleanup(func() {
		noxflags.ResetGame()
		noxflags.SetGame(originalFlags)
	})
	for _, row := range rows {
		input, result := row.Input, row.Observed
		if input.SenderResult < 0 {
			continue
		}
		t.Run(input.Name, func(t *testing.T) {
			noxflags.ResetGame()
			if input.Cooperative {
				noxflags.SetGame(noxflags.GameModeCoop)
			}
			*health = server.HealthData{Cur: input.After, Max: input.Maximum}
			item.HealthData = health
			if input.NullHealth {
				item.HealthData = nil
			}
			var packet *stockStatsReportPacket
			for _, event := range result.Events {
				if event.Kind == "packet" {
					if packet != nil {
						t.Fatal("unexpected second stock durability packet")
					}
					packet = &stockStatsReportPacket{Data: event.Data, Recipient: event.Recipient, Sequence: 1}
				}
			}
			want := uintptr(result.Word)
			if result.ReturnsHealth {
				want = uintptr(unsafe.Pointer(health))
			}
			checkQueue := stockReportQueue(t, packet, want)
			if got := callStatsReportEntry(4, input.Recipient, item, uint32(input.Before), uint32(input.After)); got != want {
				t.Fatalf("durability result=%#x want %#x", got, want)
			}
			checkQueue()
		})
	}
}

func TestItemHealthBucketFrozenOriginalC(t *testing.T) {
	for _, tc := range []struct {
		current, maximum uint32
		bucket           uintptr
	}{
		{0, 0, 4}, {100, 100, 0}, {101, 100, 1}, {50, 100, 1},
		{49, 100, 2}, {25, 100, 2}, {24, 100, 3},
		{32768, 65535, 1}, {32767, 65535, 2}, {16383, 65535, 3},
	} {
		if got := callStatsReportEntry(5, 0, nil, tc.current, tc.maximum); got != tc.bucket {
			t.Errorf("health %d/%d returned bucket %d, want %d", tc.current, tc.maximum, got, tc.bucket)
		}
	}
}
