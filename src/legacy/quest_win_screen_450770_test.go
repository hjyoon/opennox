package legacy

import (
	"encoding/binary"
	"fmt"
	"math"
	"reflect"
	"testing"
)

type questWinModelRow450770 struct {
	player int
	fields [4]uint16
	score  uint32
}

type questWinModelState450770 struct {
	rows         [6]questWinModelRow450770
	total, stage uint32
	width        int32
}

type questWinFixture450770 struct {
	packet   [90]byte
	state    questWinModelState450770
	events   []string
	states   []questWinModelState450770
	failAt   int
	font     int
	widths   [4]int32
	measures int
	sorted   int
	h        questWinHooks450770[*[90]byte, int, int, int, string]
}

func newQuestWinFixture450770(t *testing.T) *questWinFixture450770 {
	t.Helper()
	f := &questWinFixture450770{font: 7, widths: [4]int32{40, 90, 60, 80}}
	put := func(off int, v uint16) { binary.LittleEndian.PutUint16(f.packet[off:], v) }
	put(2, 0xffff)
	put(4, 0x8001)
	for i, id := range []uint16{1, 0, 3, 0, 5, 0} {
		off := 6 + i*14
		put(off, id)
		for j := 1; j < 5; j++ {
			put(off+2*j, uint16(10*i+j))
		}
		binary.LittleEndian.PutUint32(f.packet[off+10:], uint32(100+i))
		f.state.rows[i] = questWinModelRow450770{99, [4]uint16{1, 2, 3, 4}, math.MaxUint32}
	}
	f.h = questWinHooks450770[*[90]byte, int, int, int, string]{
		clearRows:  func() { f.record("clear"); f.state.rows = [6]questWinModelRow450770{} },
		storeTotal: func(v uint32) { f.record(fmt.Sprint("total:", v)); f.state.total = v },
		storeStage: func(v uint32) { f.record(fmt.Sprint("stage:", v)); f.state.stage = v },
		read16: func(p *[90]byte, off uintptr) uint16 {
			f.record(fmt.Sprint("read16:", off))
			return binary.LittleEndian.Uint16(p[off:])
		},
		read32: func(p *[90]byte, off uintptr) uint32 {
			f.record(fmt.Sprint("read32:", off))
			return binary.LittleEndian.Uint32(p[off:])
		},
		lookup: func(id uint16) int {
			f.record(fmt.Sprint("lookup:", id))
			if id == 3 {
				return 0 // A failed lookup must still count as a participant ID.
			}
			return int(id)
		},
		storePlayer: func(i, p int) { f.record(fmt.Sprintf("player:%d:%d", i, p)); f.state.rows[i].player = p },
		store16: func(i int, field questWinField450770, v uint16) {
			f.record(fmt.Sprintf("field:%d:%d:%d", i, field, v))
			f.state.rows[i].fields[field] = v
		},
		storeScore: func(i int, v uint32) { f.record(fmt.Sprintf("score:%d:%d", i, v)); f.state.rows[i].score = v },
		sort:       func(count int) { f.record(fmt.Sprint("sort:", count)); f.sorted = count },
		loadWidth:  func() int32 { f.record("load-width"); return f.state.width },
		storeWidth: func(v int32) { f.record(fmt.Sprint("width:", v)); f.state.width = v },
		child: func(id int32) int {
			f.record(fmt.Sprint("child:", id))
			return 11
		},
		loadText: func(key, source string, line int32) string {
			f.record(fmt.Sprint("text:", line))
			want := map[int32]string{1656: "GUIBrief.c:GeneratorsDestroyed", 1660: "GUIBrief.c:Kills", 1664: "GUIBrief.c:numSecretsFound", 1668: "GUIBrief.c:TotalScore"}
			if key != want[line] || source != `C:\NoxPost\src\client\Gui\GUIBrief.c` {
				t.Fatalf("lookup=%q %q %d", key, source, line)
			}
			return key
		},
		font: func(win int) int {
			f.record("font")
			if win != 11 {
				t.Fatalf("font window=%d", win)
			}
			return f.font
		},
		measure: func(font int, text string) int32 {
			f.record("measure")
			if font != f.font || text == "" {
				t.Fatalf("measure=%d %q", font, text)
			}
			width := f.widths[f.measures]
			f.measures++
			return width
		},
		lock: func(screen, mode int32, flags int8) int32 {
			f.record("lock")
			if screen != 254 || mode != 1 || flags != 1 {
				t.Fatalf("lock=%d/%d/%d", screen, mode, flags)
			}
			return math.MinInt32 + 1
		},
	}
	return f
}

func (f *questWinFixture450770) record(event string) {
	f.events = append(f.events, event)
	f.states = append(f.states, f.state)
	if f.failAt == len(f.events) {
		panic("injected helper failure")
	}
}

func TestQuestWin450770CompleteHelperOrderAndSparseRows(t *testing.T) {
	f := newQuestWinFixture450770(t)
	before := f.packet
	if got := questWinScreen450770(&f.packet, f.h); got != math.MinInt32+1 {
		t.Fatalf("return=%d", got)
	}
	want := []string{"clear", "total:0", "read16:2", "total:65535", "read16:4", "stage:32769"}
	for i, id := range []int{1, 0, 3, 0, 5, 0} {
		off := 6 + i*14
		want = append(want, fmt.Sprint("read16:", off))
		if id == 0 {
			if f.state.rows[i] != (questWinModelRow450770{}) {
				t.Fatalf("empty row %d=%+v", i, f.state.rows[i])
			}
			continue
		}
		player := id
		if id == 3 {
			player = 0
		}
		want = append(want, fmt.Sprint("lookup:", id), fmt.Sprintf("player:%d:%d", i, player))
		for field, wireField := range []int{4, 1, 2, 3} {
			v := 10*i + wireField
			want = append(want, fmt.Sprint("read16:", off+2*wireField), fmt.Sprintf("field:%d:%d:%d", i, field, v))
		}
		want = append(want, fmt.Sprint("read32:", off+10), fmt.Sprintf("score:%d:%d", i, 100+i))
		if got := f.state.rows[i]; got != (questWinModelRow450770{player, [4]uint16{uint16(10*i + 4), uint16(10*i + 1), uint16(10*i + 2), uint16(10*i + 3)}, uint32(100 + i)}) {
			t.Fatalf("row %d=%+v", i, got)
		}
	}
	want = append(want, "sort:3", "load-width", "child:1010",
		"text:1656", "font", "measure", "load-width", "width:40",
		"text:1660", "font", "measure", "load-width", "width:90",
		"text:1664", "font", "measure", "load-width",
		"text:1668", "font", "measure", "load-width", "width:85", "lock")
	if !reflect.DeepEqual(f.events, want) || f.sorted != 3 || f.state.total != 0xffff || f.state.stage != 0x8001 || f.state.width != 85 || f.packet != before {
		t.Fatalf("state=%+v sorted=%d events=%v want=%v", f.state, f.sorted, f.events, want)
	}
}

func TestQuestWin450770LiveFieldsFontAndSignedWidth(t *testing.T) {
	for _, cached := range []int32{1, 85, 86, math.MinInt32, -1} {
		f := newQuestWinFixture450770(t)
		f.state.width = cached
		questWinScreen450770(&f.packet, f.h)
		if f.state.width != cached || f.measures != 0 {
			t.Fatalf("cached=%d measured=%d final=%d", cached, f.measures, f.state.width)
		}
	}
	for _, widths := range [][4]int32{{0, -1, math.MinInt32, -8}, {84, 85, 85, 85}, {86, 40, 40, 40}, {2, 3, 4, math.MaxInt32}} {
		f := newQuestWinFixture450770(t)
		f.widths = widths
		questWinScreen450770(&f.packet, f.h)
		want := int32(0)
		for _, v := range widths {
			want = max(want, v)
		}
		want = min(want, 85)
		if f.state.width != want {
			t.Fatalf("widths=%v final=%d want=%d", widths, f.state.width, want)
		}
	}
	f := newQuestWinFixture450770(t)
	lookup := f.h.lookup
	f.h.lookup = func(id uint16) int {
		p := lookup(id)
		if id == 1 {
			binary.LittleEndian.PutUint16(f.packet[14:], 0x8002) // first row kills, after lookup
		}
		return p
	}
	load := f.h.loadText
	f.h.loadText = func(key, source string, line int32) string { f.font++; return load(key, source, line) }
	measure := f.h.measure
	f.h.measure = func(font int, text string) int32 {
		width := measure(font, text)
		if f.measures == 4 {
			f.state.width = -5 // The final signed comparison must reload this.
		}
		return width
	}
	questWinScreen450770(&f.packet, f.h)
	if f.state.rows[0].fields[0] != 0x8002 || f.font != 11 || f.state.width != 80 {
		t.Fatalf("live mutation: state=%+v font=%d", f.state, f.font)
	}
}

func TestQuestWin450770EveryFaultPrefixAndPartialStores(t *testing.T) {
	base := newQuestWinFixture450770(t)
	questWinScreen450770(&base.packet, base.h)
	for stop := 1; stop <= len(base.events); stop++ {
		f := newQuestWinFixture450770(t)
		f.failAt = stop
		var caught any
		func() {
			defer func() { caught = recover() }()
			questWinScreen450770(&f.packet, f.h)
		}()
		if caught == nil || !reflect.DeepEqual(f.events, base.events[:stop]) || f.state != base.states[stop-1] {
			t.Fatalf("fault %d (%s): panic=%v events=%v state=%+v want=%+v", stop, base.events[stop-1], caught, f.events, f.state, base.states[stop-1])
		}
	}
}
