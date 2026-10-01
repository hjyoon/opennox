package legacy

import (
	"math"
	"reflect"
	"testing"
)

type questTimerNode49B6E0 struct {
	index uint8
}

type questTimerFixture49B6E0 struct {
	root, child, player *questTimerNode49B6E0
	hidden              int32
	fps, frame, start   uint32
	title               *int
	result              int32
	trace               []string
	step                func(string)
	copied, formatted   bool
	formattedTitle      *int
	seconds             int32
	queriedRoot         *questTimerNode49B6E0
	updatedChild        *questTimerNode49B6E0
}

func newQuestTimerFixture49B6E0() *questTimerFixture49B6E0 {
	return &questTimerFixture49B6E0{
		root: &questTimerNode49B6E0{}, child: &questTimerNode49B6E0{},
		player: &questTimerNode49B6E0{}, title: new(int),
		fps: 30, frame: 100, start: 100, result: math.MinInt32 + 1,
	}
}

func (f *questTimerFixture49B6E0) visit(name string) {
	f.trace = append(f.trace, name)
	if f.step != nil {
		f.step(name)
	}
}

func (f *questTimerFixture49B6E0) hooks(t *testing.T) questGameOverTimerHooks49B6E0[*questTimerNode49B6E0, *questTimerNode49B6E0, *int] {
	t.Helper()
	return questGameOverTimerHooks49B6E0[*questTimerNode49B6E0, *questTimerNode49B6E0, *int]{
		root: func() *questTimerNode49B6E0 { f.visit("root"); return f.root },
		hidden: func(win *questTimerNode49B6E0) int32 {
			if win != f.root {
				t.Fatal("visibility query did not use entry root")
			}
			f.visit("hidden")
			return f.hidden
		},
		fps:    func() uint32 { f.visit("fps"); return f.fps },
		frame:  func() uint32 { f.visit("frame"); return f.frame },
		start:  func() uint32 { f.visit("start"); return f.start },
		player: func() *questTimerNode49B6E0 { f.visit("player"); return f.player },
		index:  func(p *questTimerNode49B6E0) uint8 { f.visit("index"); return p.index },
		copyEmpty: func() {
			f.visit("copy")
			f.copied = true
		},
		loadText: func(key, source string, line int32) *int {
			f.visit("lookup")
			if key != "Rules.c:Time" || source != `C:\NoxPost\src\client\Gui\GUIGGOvr.c` || line != 265 {
				t.Fatalf("lookup=%q/%q/%d", key, source, line)
			}
			return f.title
		},
		format: func(title *int, seconds int32) {
			f.visit("format")
			f.formatted, f.formattedTitle, f.seconds = true, title, seconds
		},
		child: func(win *questTimerNode49B6E0, id int32) *questTimerNode49B6E0 {
			f.visit("child")
			if id != 10712 {
				t.Fatalf("child id=%d", id)
			}
			f.queriedRoot = win
			return f.child
		},
		setText: func(win *questTimerNode49B6E0) int32 {
			f.visit("set")
			f.updatedChild = win
			return f.result
		},
	}
}

func TestQuestGameOverTimer49B6E0GateAndByteDomain(t *testing.T) {
	f := newQuestTimerFixture49B6E0()
	f.root = nil
	if got := questGameOverTimer49B6E0(f.hooks(t)); got != 0 || !reflect.DeepEqual(f.trace, []string{"root"}) {
		t.Fatalf("nil root=%d trace=%v", got, f.trace)
	}
	for _, hidden := range []int32{1, 256, -1, math.MinInt32, math.MaxInt32} {
		f := newQuestTimerFixture49B6E0()
		f.hidden = hidden
		if got := questGameOverTimer49B6E0(f.hooks(t)); got != hidden || !reflect.DeepEqual(f.trace, []string{"root", "hidden"}) {
			t.Fatalf("hidden=%d return=%d trace=%v", hidden, got, f.trace)
		}
	}
	for index := 0; index < 256; index++ {
		f := newQuestTimerFixture49B6E0()
		f.player.index = uint8(index)
		if got := questGameOverTimer49B6E0(f.hooks(t)); got != f.result {
			t.Fatalf("index=%d result=%d", index, got)
		}
		want := []string{"root", "hidden", "fps", "frame", "start", "player", "index"}
		if index == 31 {
			want = append(want, "copy")
		} else {
			want = append(want, "lookup", "format")
		}
		want = append(want, "root", "child", "set")
		if !reflect.DeepEqual(f.trace, want) || f.copied != (index == 31) || f.formatted != (index != 31) || f.queriedRoot != f.root || f.updatedChild != f.child {
			t.Fatalf("index=%d trace=%v copy=%v format=%v", index, f.trace, f.copied, f.formatted)
		}
	}
	f = newQuestTimerFixture49B6E0()
	f.player, f.title, f.child = nil, nil, nil
	questGameOverTimer49B6E0(f.hooks(t))
	want := []string{"root", "hidden", "fps", "frame", "start", "player", "lookup", "format", "root", "child", "set"}
	if !reflect.DeepEqual(f.trace, want) || !f.formatted || f.formattedTitle != nil || f.updatedChild != nil {
		t.Fatalf("nil player/title/child trace=%v title=%p child=%p", f.trace, f.formattedTitle, f.updatedChild)
	}
}

func TestQuestGameOverTimer49B6E0DWORDArithmetic(t *testing.T) {
	cases := [][3]uint32{
		{30, 100, 100}, {30, 101, 100}, {30, 1000, 100}, {30, 1001, 100},
		{1, 0, math.MaxInt32 - 30}, {1, 0, math.MaxInt32 - 29},
		{0x80000000, 0, 0}, {math.MaxUint32, 0, 0},
		{math.MaxUint32, math.MaxUint32, 30}, {30, math.MaxUint32, math.MaxUint32},
	}
	// Independent wide integer model, including wrap and the sign-bit clamp.
	seed := uint32(0x49b6e0)
	for i := 0; i < 2048; i++ {
		var row [3]uint32
		for j := range row {
			seed = seed*1664525 + 1013904223
			row[j] = seed
		}
		if row[0] == 0 {
			row[0] = 1
		}
		cases = append(cases, row)
	}
	for _, row := range cases {
		f := newQuestTimerFixture49B6E0()
		f.fps, f.frame, f.start = row[0], row[1], row[2]
		questGameOverTimer49B6E0(f.hooks(t))
		wide := (uint64(row[0])*30 + uint64(row[2]) - uint64(row[1])) & math.MaxUint32
		if wide&0x80000000 != 0 {
			wide = 0
		}
		want := int32(wide / uint64(row[0]))
		if f.seconds != want || f.formattedTitle != f.title {
			t.Fatalf("fps/frame/start=%x seconds=%d want=%d", row, f.seconds, want)
		}
	}
}

func TestQuestGameOverTimer49B6E0CachedAndLiveBindings(t *testing.T) {
	for _, host := range []bool{false, true} {
		f := newQuestTimerFixture49B6E0()
		entryPlayer := f.player
		if host {
			entryPlayer.index = 31
		}
		replacementRoot := &questTimerNode49B6E0{}
		f.step = func(name string) {
			switch name {
			case "frame":
				f.fps, f.start = 7, 201 // FPS cached before, timestamp read after.
			case "index":
				f.player = &questTimerNode49B6E0{index: 31 - entryPlayer.index}
			case "copy", "format":
				f.root = replacementRoot
			}
		}
		if got := questGameOverTimer49B6E0(f.hooks(t)); got != f.result || f.queriedRoot != replacementRoot || f.copied != host {
			t.Fatalf("host=%v result=%d root=%p copy=%v", host, got, f.queriedRoot, f.copied)
		}
		if !host && f.seconds != 33 {
			t.Fatalf("FPS was reloaded or timestamp cached early: seconds=%d", f.seconds)
		}
		fpsCalls := 0
		for _, name := range f.trace {
			if name == "fps" {
				fpsCalls++
			}
		}
		if fpsCalls != 1 {
			t.Fatalf("FPS calls=%d", fpsCalls)
		}
	}
}

func TestQuestGameOverTimer49B6E0FaultPrefixes(t *testing.T) {
	for _, host := range []bool{false, true} {
		trace := []string{"root", "hidden", "fps", "frame", "start", "player", "index"}
		if host {
			trace = append(trace, "copy")
		} else {
			trace = append(trace, "lookup", "format")
		}
		trace = append(trace, "root", "child", "set")
		for failAt := range trace {
			f := newQuestTimerFixture49B6E0()
			if host {
				f.player.index = 31
			}
			f.step = func(string) {
				if len(f.trace) == failAt+1 {
					panic("boundary fault")
				}
			}
			var fault any
			func() {
				defer func() { fault = recover() }()
				questGameOverTimer49B6E0(f.hooks(t))
			}()
			if fault != "boundary fault" || !reflect.DeepEqual(f.trace, trace[:failAt+1]) {
				t.Fatalf("host=%v boundary=%d fault=%v trace=%v", host, failAt, fault, f.trace)
			}
		}
	}
	for _, host := range []bool{false, true} {
		f := newQuestTimerFixture49B6E0()
		f.fps = 0
		if host {
			f.player.index = 31
		}
		var fault any
		func() {
			defer func() { fault = recover() }()
			questGameOverTimer49B6E0(f.hooks(t))
		}()
		if host {
			if fault != nil || !f.copied {
				t.Fatalf("host must bypass DIV even at FPS 0: %v", fault)
			}
		} else {
			want := []string{"root", "hidden", "fps", "frame", "start", "player", "index"}
			if fault == nil || !reflect.DeepEqual(f.trace, want) {
				t.Fatalf("zero FPS fault=%v trace=%v", fault, f.trace)
			}
		}
	}
}
