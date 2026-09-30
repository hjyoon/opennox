package legacy

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type generatorDrawFixture4BC750 struct {
	trace                                           []string
	stop                                            int
	flags, class, objFlags, timer, net, tick, slave uint32
	count, delay                                    [5]uint8
	kind                                            [5]uint32
	tables                                          [5]int
	indices                                         []int32
	draws                                           []int
	onRandom                                        func()
	onDraw                                          func(int)
}

func newGeneratorDrawFixture4BC750() *generatorDrawFixture4BC750 {
	f := &generatorDrawFixture4BC750{stop: -1, class: 0x123fffff, objFlags: 0xabcdef00, net: 1, tick: 5}
	for i := range f.count {
		f.count[i], f.delay[i], f.kind[i], f.tables[i] = 4, 1, 2, 10+i
	}
	f.count[4], f.delay[4] = 3, 0
	return f
}

func (f *generatorDrawFixture4BC750) step(event string) {
	f.trace = append(f.trace, event)
	if len(f.trace)-1 == f.stop {
		panic("dependency fault")
	}
}

func (f *generatorDrawFixture4BC750) hooks() monsterGeneratorDrawHooks4BC750[int, int, int] {
	return monsterGeneratorDrawHooks4BC750[int, int, int]{
		flags:      func() uint32 { f.step(fmt.Sprintf("flags:%x", f.flags)); return f.flags },
		data:       func() int { f.step("data:9"); return 9 },
		images:     func(d, s int) int { f.step(fmt.Sprintf("images:%d:%d:%d", d, s, f.tables[s])); return f.tables[s] },
		count:      func(d, s int) uint8 { f.step(fmt.Sprintf("count:%d:%d:%d", d, s, f.count[s])); return f.count[s] },
		kind:       func(d, s int) uint32 { f.step(fmt.Sprintf("kind:%d:%d:%d", d, s, f.kind[s])); return f.kind[s] },
		delay:      func(d, s int) uint8 { f.step(fmt.Sprintf("delay:%d:%d:%d", d, s, f.delay[s])); return f.delay[s] },
		delayIndex: func(d, s int) int32 { f.step(fmt.Sprintf("delay-index:%d:%d", d, s)); return 0x12345678 },
		netcode:    func() uint32 { f.step(fmt.Sprintf("net:%d", f.net)); return f.net },
		frame:      func() uint32 { f.step(fmt.Sprintf("frame:%d", f.tick)); return f.tick },
		slave:      func() uint32 { f.step(fmt.Sprintf("slave:%d", f.slave)); return f.slave },
		random: func(min, max int32, source string, line int32) int32 {
			f.step(fmt.Sprintf("random:%d:%d:%s:%d", min, max, source, line))
			if f.onRandom != nil {
				f.onRandom()
			}
			return max
		},
		class:    func() uint32 { f.step(fmt.Sprintf("class:%x", f.class)); return f.class },
		setClass: func(v uint32) { f.step(fmt.Sprintf("set-class:%x", v)); f.class = v },
		objFlags: func() uint32 { f.step(fmt.Sprintf("obj-flags:%x", f.objFlags)); return f.objFlags },
		setFlags: func(v uint32) { f.step(fmt.Sprintf("set-flags:%x", v)); f.objFlags = v },
		timer:    func() uint32 { f.step(fmt.Sprintf("timer:%d", f.timer)); return f.timer },
		setTimer: func(v uint32) { f.step(fmt.Sprintf("set-timer:%d", v)); f.timer = v },
		setState: func(v uint32) { f.step(fmt.Sprintf("set-state:%x", v)); f.flags = v },
		image: func(p int, i int32) int {
			f.step(fmt.Sprintf("image:%d:%d", p, i))
			f.indices = append(f.indices, i)
			return p*100 + int(i)
		},
		draw: func(i int) {
			f.step(fmt.Sprintf("draw:%d", i))
			f.draws = append(f.draws, i)
			if f.onDraw != nil {
				f.onDraw(len(f.draws))
			}
		},
	}
}

func TestMonsterGeneratorDraw4BC750CompleteTrace(t *testing.T) {
	f := newGeneratorDrawFixture4BC750()
	if got := monsterGeneratorDraw4BC750(f.hooks()); got != 1 {
		t.Fatalf("return=%d", got)
	}
	want := []string{
		"flags:0", "data:9", "images:9:0:10", "count:9:0:4", "kind:9:0:2", "delay:9:0:1",
		"net:1", "frame:5", "flags:0", "timer:0", "image:10:3", "draw:1003",
		"flags:0", "net:1", "frame:5", "images:9:4:14", "delay:9:4:0", "count:9:4:3",
		"image:14:0", "draw:1400", "flags:0",
	}
	if !reflect.DeepEqual(f.trace, want) {
		t.Fatalf("trace=%v\nwant=%v", f.trace, want)
	}
}

func TestMonsterGeneratorDraw4BC750KindsAndStatePrecedence(t *testing.T) {
	for _, flags := range []uint32{0, 0x100, 0x200, 0x400, 0x800, 0x300, 0xf00} {
		state := 0
		if flags&0x100 != 0 {
			state = 1
		} else if flags&0x200 != 0 {
			state = 2
		} else if flags&0xc00 != 0 {
			state = 3
		}
		for _, kind := range []uint32{0, 1, 2, 3, 4, 5, 6, math.MaxUint32} {
			t.Run(fmt.Sprintf("%x/%d", flags, kind), func(t *testing.T) {
				f := newGeneratorDrawFixture4BC750()
				f.flags = flags
				f.kind[state] = kind
				f.slave = 11
				if kind == 0 {
					f.timer = 2
				} // address-as-index can be overridden
				got := monsterGeneratorDraw4BC750(f.hooks())
				want := int32(1)
				if kind == 1 || kind == 3 || kind > 5 {
					want = 0
				}
				if got != want {
					t.Fatalf("return=%d want=%d trace=%v", got, want, f.trace)
				}
				if want == 0 {
					if len(f.trace) != 6 || len(f.draws) != 0 {
						t.Fatalf("rejected trace=%v", f.trace)
					}
					return
				}
				if flags&0x800 == 0 && kind == 4 && f.indices[0] != 4 {
					t.Fatalf("inclusive random frame=%v", f.indices)
				}
				if flags&0x800 == 0 && kind == 5 && f.indices[0] != 11 {
					t.Fatalf("slave frame was bounded: %v", f.indices)
				}
				if flags&0x800 != 0 {
					if f.class&0x80000 != 0 || f.objFlags&0x20000000 != 0 || f.objFlags&1 == 0 {
						t.Fatalf("terminal class=%x flags=%x", f.class, f.objFlags)
					}
				}
			})
		}
	}
}

func TestMonsterGeneratorDraw4BC750SignedFrameAndUnsignedWrap(t *testing.T) {
	for _, tc := range []struct {
		net, tick uint32
		delay     uint8
		want      int32
	}{
		{1, math.MaxUint32, 0, 0}, {0, math.MaxInt32, 0, 3}, {0, 0x80000000, 0, math.MinInt32},
		{0, math.MaxUint32, 0, -1}, {0, math.MaxUint32, 1, 3}, {2, 4, 255, 0},
	} {
		f := newGeneratorDrawFixture4BC750()
		f.flags = 0x100
		f.net = tc.net
		f.tick = tc.tick
		f.delay[1] = tc.delay
		f.onDraw = func(int) { f.flags = 0x400 }
		monsterGeneratorDraw4BC750(f.hooks())
		if len(f.indices) != 1 || f.indices[0] != tc.want {
			t.Fatalf("%+v: indices=%v", tc, f.indices)
		}
	}
	f := newGeneratorDrawFixture4BC750()
	f.count[0] = 0
	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("zero divisor was silently skipped")
			}
		}()
		monsterGeneratorDraw4BC750(f.hooks())
	}()
	if len(f.draws) != 0 {
		t.Fatal("zero-divisor path drew")
	}
}

func TestMonsterGeneratorDraw4BC750TimerAndLiveCallbacks(t *testing.T) {
	for _, tc := range []struct {
		delay uint8
		timer uint32
		want  int32
	}{
		{1, 1, 3}, {1, 2, 3}, {1, 9, 3}, {0, 5, 0}, {0, math.MaxUint32, 3}, {255, 1025, 3},
	} {
		f := newGeneratorDrawFixture4BC750()
		f.flags = 0x400
		f.kind[3] = 5
		f.slave = 999
		f.delay[3] = tc.delay
		f.timer = tc.timer
		monsterGeneratorDraw4BC750(f.hooks())
		if f.indices[0] != tc.want || f.timer != tc.timer-1 {
			t.Fatalf("%+v: frame=%d timer=%d", tc, f.indices[0], f.timer)
		}
		if tc.timer == 1 && (f.flags != 0x800 || f.objFlags&1 == 0) {
			t.Fatalf("completion flags=%x/%x", f.flags, f.objFlags)
		}
	}
	f := newGeneratorDrawFixture4BC750()
	f.kind[0] = 4
	f.timer = 1
	f.onRandom = func() { f.delay[0] = 2; f.count[0] = 8; f.tables[0] = 99 }
	monsterGeneratorDraw4BC750(f.hooks())
	if f.indices[0] != 3 || !reflect.DeepEqual(f.draws, []int{1003}) {
		t.Fatalf("cached limit/table or live timer metadata: %v/%v", f.indices, f.draws)
	}
	f = newGeneratorDrawFixture4BC750()
	f.onDraw = func(n int) {
		if n == 1 {
			f.net = 9
			f.tick = 0
			f.tables[4] = 88
			f.delay[4] = 1
			f.count[4] = 3
		}
		if n == 2 {
			f.flags = 0x800
			f.objFlags = 0x12345600
		}
	}
	monsterGeneratorDraw4BC750(f.hooks())
	if !reflect.DeepEqual(f.draws, []int{1003, 8801}) || f.objFlags != 0x12345601 {
		t.Fatalf("live post-draw data=%v flags=%x", f.draws, f.objFlags)
	}
	f = newGeneratorDrawFixture4BC750()
	f.onDraw = func(int) { f.flags = 0x400 }
	monsterGeneratorDraw4BC750(f.hooks())
	if len(f.draws) != 1 {
		t.Fatal("overlay selected before main callback")
	}
}

func TestMonsterGeneratorDraw4BC750EveryDependencyFaultPrefix(t *testing.T) {
	setups := []func(*generatorDrawFixture4BC750){
		func(*generatorDrawFixture4BC750) {},
		func(f *generatorDrawFixture4BC750) { f.flags = 0x800 },
		func(f *generatorDrawFixture4BC750) { f.flags = 0x400; f.timer = 1 },
		func(f *generatorDrawFixture4BC750) { f.kind[0] = 0; f.timer = 2 },
		func(f *generatorDrawFixture4BC750) { f.kind[0] = 4 },
		func(f *generatorDrawFixture4BC750) { f.kind[0] = 5; f.slave = 99 },
		func(f *generatorDrawFixture4BC750) { f.kind[0] = 1 },
	}
	for scenario, setup := range setups {
		baseline := newGeneratorDrawFixture4BC750()
		setup(baseline)
		monsterGeneratorDraw4BC750(baseline.hooks())
		for stop := range baseline.trace {
			f := newGeneratorDrawFixture4BC750()
			setup(f)
			f.stop = stop
			var fault any
			func() { defer func() { fault = recover() }(); monsterGeneratorDraw4BC750(f.hooks()) }()
			if fault != "dependency fault" || !reflect.DeepEqual(f.trace, baseline.trace[:stop+1]) {
				t.Fatalf("scenario=%d stop=%d fault=%v trace=%v", scenario, stop, fault, f.trace)
			}
		}
	}
}
