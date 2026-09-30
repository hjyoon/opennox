package legacy

import (
	"bytes"
	"encoding/binary"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/client/gui"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestQuestWin450770CEntryFullPacketAndSignedReturn(t *testing.T) {
	packet, free := alloc.Make([]byte{0xf0, 0x0c}, 90)
	t.Cleanup(free)
	if unsafe.Sizeof(uintptr(0)) > 4 && uintptr(unsafe.Pointer(&packet[0])) <= math.MaxUint32 {
		t.Fatalf("packet=%p, want above 4 GiB", &packet[0])
	}
	old := questWinCall450770
	t.Cleanup(func() { questWinCall450770 = old })
	for _, ptr := range []*byte{&packet[0], nil} {
		questWinCall450770 = func(p *byte) int32 {
			if p != ptr {
				t.Fatalf("packet=%p want=%p", p, ptr)
			}
			return math.MinInt32 + 1
		}
		if got := questWinCEntry450770(ptr); got != math.MinInt32+1 {
			t.Fatalf("return=%d", got)
		}
	}
}

func TestQuestWin450770CEntryNativeStorageSparseSlotsAndLiveFont(t *testing.T) {
	rows, width := questWinRows450770(), questWinWidthSlot450770()
	total, stage := memmap.PtrUint32(0x5D4594, 832356), memmap.PtrUint32(0x5D4594, 831228)
	oldRows, oldWidth, oldTotal, oldStage, oldCall := *rows, *width, *total, *stage, questWinCall450770
	t.Cleanup(func() {
		*rows, *width, *total, *stage, questWinCall450770 = oldRows, oldWidth, oldTotal, oldStage, oldCall
	})
	win, freeWin := alloc.New(gui.Window{})
	font, freeFont := alloc.Malloc(8)
	text, freeText := alloc.Make([]uint16{'x', 0}, 2)
	t.Cleanup(freeWin)
	t.Cleanup(freeFont)
	t.Cleanup(freeText)
	win.DrawData().SetFont(font)
	var players [6]*server.Player
	for i := range players {
		p, free := alloc.New(server.Player{})
		t.Cleanup(free)
		players[i] = p
		p.NameFinal[0] = uint16('A' + i)
	}
	for _, ptr := range []unsafe.Pointer{unsafe.Pointer(win), font, unsafe.Pointer(&text[0]), unsafe.Pointer(players[0])} {
		if unsafe.Sizeof(uintptr(0)) > 4 && uintptr(ptr) <= math.MaxUint32 {
			t.Fatalf("native pointer=%p, want above 4 GiB", ptr)
		}
	}
	for _, ids := range [][6]uint16{{}, {1, 2, 3, 4, 5, 6}, {1, 0, 3, 0, 5, 0}} {
		packet, free := alloc.Make([]byte{0xf0, 0x0c}, 104)
		t.Cleanup(free)
		binary.LittleEndian.PutUint16(packet[2:], 0xffff)
		binary.LittleEndian.PutUint16(packet[4:], 0x8001)
		var want [6]questWinRow450770
		count := 0
		for i, id := range ids {
			off := 6 + i*14
			binary.LittleEndian.PutUint16(packet[off:], id)
			for field := 1; field < 5; field++ {
				binary.LittleEndian.PutUint16(packet[off+2*field:], uint16(0x8000+i*10+field))
			}
			binary.LittleEndian.PutUint32(packet[off+10:], math.MaxUint32-uint32(i))
			if id != 0 {
				p := players[id-1]
				if id == 3 {
					p = nil // Nonzero wire ID still counts when lookup returns nil.
				}
				want[i] = questWinRow450770{p, uint16(0x8000 + i*10 + 4), uint16(0x8000 + i*10 + 1), uint16(0x8000 + i*10 + 2), uint16(0x8000 + i*10 + 3), math.MaxUint32 - uint32(i)}
				count++
			}
		}
		binary.LittleEndian.PutUint16(packet[90:], 7) // A seventh wire row must not be read.
		before := bytes.Clone(packet)
		for _, cached := range []uint32{0, math.MaxUint32, 100} {
			*width, *total, *stage = cached, 7, 8
			h := questWinNativeHooks450770()
			lookups, measures := 0, 0
			h.lookup = func(id uint16) *server.Player {
				lookups++
				if id == 3 {
					return nil
				}
				return players[id-1]
			}
			h.child = func(id int32) *gui.Window {
				if id != 1010 {
					t.Fatalf("child=%d", id)
				}
				return win
			}
			h.loadText = func(key, source string, line int32) *uint16 {
				if source != questWinSource450770 || line != int32(1656+4*measures) {
					t.Fatalf("text=%q %q %d", key, source, line)
				}
				// Font is read from the actual C-allocated native window only
				// after this helper. Passing nil results is not a skip branch.
				if measures%2 == 0 {
					win.DrawData().SetFont(nil)
					return nil
				}
				win.DrawData().SetFont(font)
				return &text[0]
			}
			h.measure = func(gotFont unsafe.Pointer, gotText *uint16) int32 {
				wantFont, wantText := font, &text[0]
				if measures%2 == 0 {
					wantFont, wantText = nil, nil
				}
				if gotFont != wantFont || gotText != wantText {
					t.Fatalf("font=%p/%p text=%p/%p", gotFont, wantFont, gotText, wantText)
				}
				measures++
				return 100
			}
			h.lock = func(screen, mode int32, flags int8) int32 {
				if screen != 254 || mode != 1 || flags != 1 || *total != 0xffff || *stage != 0x8001 {
					t.Fatalf("lock=%d/%d/%d total=%#x stage=%#x", screen, mode, flags, *total, *stage)
				}
				return math.MinInt32 + 1
			}
			questWinCall450770 = func(packet *byte) int32 { return questWinScreen450770(packet, h) }
			got := questWinCEntry450770(&packet[0])
			wantRows := want
			// Independent expected prefix ordering, with no ties in fixtures.
			for i := 0; i < count; i++ {
				for j := i + 1; j < count; j++ {
					if wantRows[j].Score > wantRows[i].Score {
						wantRows[i], wantRows[j] = wantRows[j], wantRows[i]
					}
				}
			}
			wantMeasures, wantWidth := 0, cached
			if cached == 0 {
				wantMeasures, wantWidth = 4, 85
			}
			if got != math.MinInt32+1 || *rows != wantRows || lookups != count || measures != wantMeasures || *width != wantWidth || !bytes.Equal(packet, before) {
				t.Fatalf("IDs=%v cached=%#x return=%d rows=%+v want=%+v lookups=%d measures=%d width=%d", ids, cached, got, *rows, wantRows, lookups, measures, *width)
			}
		}
	}
}
