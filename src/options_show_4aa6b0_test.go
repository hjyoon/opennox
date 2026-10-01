package opennox

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type optionsShowFixture4AA6B0 struct {
	trace         []string
	fault         int
	newRoot, root int32
	newAnim, anim int32
	advanced      int32
	checkboxes    [3]int32
	currents      [3]uint32
	enabled       [3]int32
	flags         map[int32]uint32
	image         int32
}

func newOptionsShowFixture4AA6B0() *optionsShowFixture4AA6B0 {
	return &optionsShowFixture4AA6B0{
		newRoot: 10, newAnim: 20, advanced: 1, fault: -1,
		currents: [3]uint32{0x1234ffff, 0x80000000, 0xffff0123},
		enabled:  [3]int32{1, 0, -1},
		flags:    map[int32]uint32{361: 0xa5a50000, 362: 0x1234567f, 363: 0xffffffff},
	}
}

func (f *optionsShowFixture4AA6B0) call(format string, args ...any) {
	f.trace = append(f.trace, fmt.Sprintf(format, args...))
	if len(f.trace)-1 == f.fault {
		panic("options callback fault")
	}
}

func (f *optionsShowFixture4AA6B0) hooks() optionsShowHooks4AA6B0[int32, int32, int32] {
	return optionsShowHooks4AA6B0[int32, int32, int32]{
		addState:       func(id int32) { f.call("state:%d", id) },
		newWindow:      func(name string) int32 { f.call("new:%s", name); return f.newRoot },
		storeRoot:      func(win int32) { f.call("publish-root:%d", win); f.root = win },
		advanced:       func(win int32) int32 { f.call("advanced:%d", win); return f.advanced },
		root:           func() int32 { f.call("root:%d", f.root); return f.root },
		setProc:        func(win int32) { f.call("proc:%d", win) },
		tabWidth:       func(width int32) { f.call("tab:%d", width) },
		newAnimation:   func(win int32, args [8]int32) int32 { f.call("new-anim:%d:%v", win, args); return f.newAnim },
		storeAnimation: func(anim int32) { f.call("publish-anim:%d", anim); f.anim = anim },
		animation:      func() int32 { f.call("anim:%d", f.anim); return f.anim },
		stateID:        func(anim, id int32) { f.call("anim-state:%d:%d", anim, id) },
		startOut:       func(anim int32) { f.call("start-out:%d", anim) },
		doneOut:        func(anim int32) { f.call("done-out:%d", anim) },
		child:          func(win, id int32) int32 { f.call("child:%d:%d", win, id); return id },
		thumb:          func(win int32) int32 { f.call("thumb:%d", win); return win + 1000 },
		width:          func(win, value int32) { f.call("width:%d:%d", win, value) },
		height:         func(win, value int32) { f.call("height:%d:%d", win, value) },
		loadImage:      func(name string) int32 { f.call("image:%s", name); f.image++; return f.image },
		images:         func(win, en, sel, hl int32) { f.call("images:%d:%d:%d:%d", win, en, sel, hl) },
		event:          func(win, code int32, first, second uint32) { f.call("event:%d:%x:%d:%d", win, code, first, second) },
		current:        func(channel int32) uint32 { f.call("current:%d", channel); return f.currents[channel] },
		storeCheckbox:  func(channel, win int32) { f.call("publish-check:%d:%d", channel, win); f.checkboxes[channel] = win },
		enabled:        func(channel int32) int32 { f.call("enabled:%d", channel); return f.enabled[channel] },
		checkbox: func(channel int32) int32 {
			f.call("check:%d:%d", channel, f.checkboxes[channel])
			return f.checkboxes[channel]
		},
		flags:       func(win int32) uint32 { f.call("flags:%d", win); return f.flags[win] },
		storeFlags:  func(win int32, flags uint32) { f.call("flags-set:%d:%x", win, flags); f.flags[win] = flags },
		returnNull:  func(win int32) { f.call("return-null:%d", win) },
		backText:    func(name string) { f.call("back:%s", name) },
		backEnabled: func(enabled int32) { f.call("back-enabled:%d", enabled) },
		video:       func() { f.call("video") },
	}
}

func optionsShowWantTrace4AA6B0() []string {
	want := []string{"state:300", "new:Options.wnd", "publish-root:10", "advanced:10", "root:10", "proc:10", "tab:15",
		"root:10", "new-anim:10:[0 0 0 -480 0 20 0 -40]", "publish-anim:20", "anim-state:20:300", "anim:20", "start-out:20", "anim:20", "done-out:20"}
	for channel, value := range []uint32{0x1234, 0x8000, 0xffff} {
		id, check := 351+channel, 361+channel
		want = append(want,
			"root:10", fmt.Sprintf("child:10:%d", id), fmt.Sprintf("thumb:%d", id), fmt.Sprintf("width:%d:24", id+1000),
			fmt.Sprintf("thumb:%d", id), fmt.Sprintf("height:%d:20", id+1000),
			"image:OptionsVolumeSliderLit", "image:OptionsVolumeSliderLit", "image:OptionsVolumeSlider",
			fmt.Sprintf("images:%d:%d:%d:%d", id, 3*channel+3, 3*channel+2, 3*channel+1),
			fmt.Sprintf("event:%d:400b:0:16384", id), fmt.Sprintf("current:%d", channel), fmt.Sprintf("event:%d:400a:%d:0", id, value),
			"root:10", fmt.Sprintf("child:10:%d", check), fmt.Sprintf("publish-check:%d:%d", channel, check),
			fmt.Sprintf("enabled:%d", channel), fmt.Sprintf("check:%d:%d", channel, check), fmt.Sprintf("flags:%d", check))
		want = append(want, []string{"flags-set:361:a5a50004", "flags-set:362:1234567b", "flags-set:363:fffffffb"}[channel])
	}
	return append(want, "root:10", "return-null:10", "back:OptsBack.wnd:Back", "back-enabled:0", "video")
}

func TestOptionsShow4AA6B0ExactTraceAndEveryCallbackFault(t *testing.T) {
	want := optionsShowWantTrace4AA6B0()
	f := newOptionsShowFixture4AA6B0()
	if got := optionsShow4AA6B0(f.hooks()); got != 1 || !reflect.DeepEqual(f.trace, want) {
		t.Fatalf("return=%d\ntrace=%v\nwant =%v", got, f.trace, want)
	}
	for fault := range want {
		f := newOptionsShowFixture4AA6B0()
		f.fault = fault
		func() {
			defer func() {
				if got := recover(); got != "options callback fault" {
					t.Fatalf("fault %d recovery=%v", fault, got)
				}
			}()
			optionsShow4AA6B0(f.hooks())
		}()
		if !reflect.DeepEqual(f.trace, want[:fault+1]) {
			t.Fatalf("fault %d trace=%v, want exact prefix=%v", fault, f.trace, want[:fault+1])
		}
	}
}

func TestOptionsShow4AA6B0PublishesFailuresWithoutInventedCleanup(t *testing.T) {
	want := optionsShowWantTrace4AA6B0()
	for _, tc := range []struct {
		name                 string
		root, advanced, anim int32
		prefix               int
	}{
		{"window", 0, 1, 20, 3}, {"advanced", 10, 0, 20, 4}, {"animation", 10, 1, 0, 10},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newOptionsShowFixture4AA6B0()
			f.newRoot, f.advanced, f.newAnim = tc.root, tc.advanced, tc.anim
			if got := optionsShow4AA6B0(f.hooks()); got != 0 || f.root != tc.root || f.anim != 0 {
				t.Fatalf("failed constructor=%d root=%d anim=%d", got, f.root, f.anim)
			}
			expected := append([]string(nil), want[:tc.prefix]...)
			if tc.root == 0 {
				expected[2] = "publish-root:0"
			}
			if tc.anim == 0 {
				expected[9] = "publish-anim:0"
			}
			if !reflect.DeepEqual(f.trace, expected) {
				t.Fatalf("trace=%v, want=%v", f.trace, expected)
			}
		})
	}
	for _, result := range []int32{math.MinInt32, -1, 1, 2, math.MaxInt32} {
		f := newOptionsShowFixture4AA6B0()
		f.advanced = result
		if got := optionsShow4AA6B0(f.hooks()); got != 1 {
			t.Fatalf("nonzero advanced %d returned %d", result, got)
		}
	}
}

func TestOptionsShow4AA6B0LogicalCurrentAndExactOneCheckboxFlag(t *testing.T) {
	for _, value := range []uint32{0, 1, 0xffff, 0x10000, 0x4000ffff, 0x7fffffff, 0x80000000, 0xffff0000, 0xffffffff} {
		for _, enabled := range []int32{math.MinInt32, -2, -1, 0, 1, 2, 255, math.MaxInt32} {
			f := newOptionsShowFixture4AA6B0()
			h := f.hooks()
			for index := range f.currents {
				f.currents[index], f.enabled[index] = value, enabled
			}
			var events []uint32
			h.event = func(_ int32, code int32, first, second uint32) {
				if code == 0x400a {
					events = append(events, first)
					if second != 0 {
						t.Fatal("value event second argument")
					}
				}
			}
			before := make(map[int32]uint32)
			for key, flags := range f.flags {
				before[key] = flags
			}
			if got := optionsShow4AA6B0(h); got != 1 || !reflect.DeepEqual(events, []uint32{value >> 16, value >> 16, value >> 16}) {
				t.Fatalf("current=%08x enable=%d return=%d events=%v", value, enabled, got, events)
			}
			for key, flags := range before {
				want := flags &^ 4
				if enabled == 1 {
					want |= 4
				}
				if f.flags[key] != want {
					t.Fatalf("enable=%d checkbox=%d flags=%08x want=%08x", enabled, key, f.flags[key], want)
				}
			}
		}
	}
}

func TestOptionsShow4AA6B0CachedReturnsAndLiveGlobals(t *testing.T) {
	f := newOptionsShowFixture4AA6B0()
	h := f.hooks()
	h.storeRoot = func(win int32) { f.root = 11 }
	h.advanced = func(win int32) int32 {
		if win != 10 {
			t.Fatal("advanced must use returned root")
		}
		f.root = 12
		return 1
	}
	h.setProc = func(win int32) {
		if win != 12 {
			t.Fatal("procedure must reload root")
		}
		f.root = 13
	}
	h.newAnimation = func(win int32, _ [8]int32) int32 {
		if win != 13 {
			t.Fatal("animation must reload root")
		}
		f.root = 14
		return 20
	}
	h.storeAnimation = func(anim int32) { f.anim = 21 }
	h.stateID = func(anim, id int32) {
		if anim != 20 || id != 300 {
			t.Fatal("state ID must use returned animation")
		}
		f.anim = 22
	}
	h.startOut = func(anim int32) {
		if anim != 22 {
			t.Fatal("start callback must reload animation")
		}
		f.anim = 23
	}
	h.doneOut = func(anim int32) {
		if anim != 23 {
			t.Fatal("done callback must reload animation")
		}
	}
	thumbCalls := 0
	h.thumb = func(win int32) int32 { thumbCalls++; return win + int32(1000*thumbCalls) }
	h.width = func(win, value int32) {
		if win != 351+int32(thumbCalls/2)+int32(1000*thumbCalls) || value != 24 {
			t.Fatalf("width thumb=%d", win)
		}
	}
	h.height = func(win, value int32) {
		if win != 350+int32(thumbCalls/2)+int32(1000*thumbCalls) || value != 20 {
			t.Fatalf("height thumb=%d", win)
		}
	}
	h.child = func(win, id int32) int32 {
		if win != f.root {
			t.Fatal("child lookup did not reload root")
		}
		f.root++
		return id
	}
	h.storeCheckbox = func(channel, win int32) { f.checkboxes[channel] = win }
	h.enabled = func(channel int32) int32 {
		f.checkboxes[channel] = 400 + channel
		f.flags[400+channel] = 0xabcdef01
		return 1
	}
	h.returnNull = func(win int32) {
		if win != 20 {
			t.Fatalf("tail root=%d, want after six live child lookups", win)
		}
	}
	if got := optionsShow4AA6B0(h); got != 1 || thumbCalls != 6 {
		t.Fatalf("return=%d thumb reads=%d", got, thumbCalls)
	}
	for channel := int32(0); channel < 3; channel++ {
		if f.flags[400+channel] != 0xabcdef05 {
			t.Fatal("flag write missed replacement checkbox")
		}
		if f.flags[361+channel] != []uint32{0xa5a50000, 0x1234567f, 0xffffffff}[channel] {
			t.Fatal("cached checkbox was changed")
		}
	}
}

func TestOptionsShow4AA6B0MissingThumbStillFaultsBeforeImageLoads(t *testing.T) {
	f := newOptionsShowFixture4AA6B0()
	h := f.hooks()
	h.thumb = func(win int32) int32 { return 0 }
	h.width = func(win, width int32) {
		if win == 0 {
			panic("missing native thumb")
		}
	}
	defer func() {
		if got := recover(); got != "missing native thumb" || f.image != 0 {
			t.Fatalf("missing thumb fault=%v images=%d", got, f.image)
		}
	}()
	optionsShow4AA6B0(h)
}
