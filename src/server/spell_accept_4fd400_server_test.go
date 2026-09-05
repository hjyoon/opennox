package server

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/things"

	"github.com/opennox/opennox/v1/common/sound"
)

func TestSpellAcceptNative4FD400Layouts(t *testing.T) {
	wantObjectSize := uintptr(780)
	wantObjectClass := uintptr(8)
	wantArgSize := uintptr(12)
	wantArgPos := uintptr(4)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize = 928
		wantObjectClass = 12
		wantArgSize = 16
		wantArgPos = 8
	}
	for _, check := range []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), wantObjectClass},
		{"SpellAcceptArg size", unsafe.Sizeof(SpellAcceptArg{}), wantArgSize},
		{"SpellAcceptArg.Obj", unsafe.Offsetof(SpellAcceptArg{}.Obj), 0},
		{"SpellAcceptArg.Pos", unsafe.Offsetof(SpellAcceptArg{}.Pos), wantArgPos},
		{"SpellAcceptArg.Obj width", unsafe.Sizeof(SpellAcceptArg{}.Obj), unsafe.Sizeof(uintptr(0))},
	} {
		if check.got != check.want {
			t.Errorf("%s on %s/%s = %d, want %d", check.name, runtime.GOOS, runtime.GOARCH, check.got, check.want)
		}
	}
}

func TestSpellAcceptNative4FD400PreservesPointersAndDwords(t *testing.T) {
	second := new(Object)
	third := new(Object)
	fourth := new(Object)
	arg := new(SpellAcceptArg)

	got := spellAcceptNative4FD400(1, second, third, fourth, arg, math.MaxInt32, spellAcceptNativeDeps4FD400{
		spellHasFlags: func(gotID int32, gotMask uint32) int32 {
			if gotID != 1 || gotMask != spellAcceptTargetFlag4FD400 {
				t.Fatalf("flags args = %d/%#x, want 1/%#x", gotID, gotMask, spellAcceptTargetFlag4FD400)
			}
			return 0
		},
		tickRate: func() uint32 {
			t.Fatal("instant path loaded tick rate")
			return 0
		},
		runtime: SpellAcceptRuntime4FD400{
			CaptureMagic: func(spell.ID, *Object) int32 {
				t.Fatal("nil target reached capture callback")
				return 0
			},
			Audio: func(soundID sound.ID, obj *Object) {
				t.Fatalf("unexpected audio callback %v/%p", soundID, obj)
			},
			Instant: func(gotID spell.ID, gotSecond, gotThird, gotFourth *Object, gotArg *SpellAcceptArg, gotLevel int32) int32 {
				if gotID != 1 || gotSecond != second || gotThird != third || gotFourth != fourth || gotArg != arg || gotLevel != math.MaxInt32 {
					t.Fatalf("instant args = %d/%p/%p/%p/%p/%d", gotID, gotSecond, gotThird, gotFourth, gotArg, gotLevel)
				}
				return math.MinInt32
			},
			Duration: func(spell.ID, *Object, *Object, *Object, *SpellAcceptArg, int32, uint32) int32 {
				t.Fatal("instant path reached duration callback")
				return 0
			},
			PlasmaTime: func() float64 {
				t.Fatal("instant path loaded plasma time")
				return 0
			},
		},
	})
	if got != math.MinInt32 {
		t.Fatalf("result = %d, want verbatim %d", got, int32(math.MinInt32))
	}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for name, ptr := range map[string]uintptr{
			"second": uintptr(unsafe.Pointer(second)),
			"third":  uintptr(unsafe.Pointer(third)),
			"fourth": uintptr(unsafe.Pointer(fourth)),
			"arg":    uintptr(unsafe.Pointer(arg)),
		} {
			if ptr <= math.MaxUint32 {
				t.Fatalf("%s pointer = %#x, want native address above 4 GiB", name, ptr)
			}
		}
	}
	runtime.KeepAlive(second)
	runtime.KeepAlive(third)
	runtime.KeepAlive(fourth)
	runtime.KeepAlive(arg)
}

func TestSpellAcceptServer4FD400UsesDirectIDAndWrappingTimeout(t *testing.T) {
	const id = spell.ID(28)
	s := new(Server)
	s.Spells.byID = map[spell.ID]*SpellDef{
		id: {
			ID:     id,
			Effect: spell.ID(7),
			Def:    things.Spell{Flags: things.SpellFlagUnk8},
		},
	}
	s.SetTickRate(0x80000001)
	second := new(Object)
	third := new(Object)
	fourth := new(Object)
	target := &Object{ObjClass: object.ClassPlayer}
	arg := &SpellAcceptArg{Obj: target}

	captureCalls := 0
	got := s.SpellAccept4FD400(id, second, third, fourth, arg, math.MaxInt32, SpellAcceptRuntime4FD400{
		CaptureMagic: func(gotID spell.ID, gotTarget *Object) int32 {
			captureCalls++
			if gotID != id || gotTarget != target {
				t.Fatalf("capture args = %d/%p, want %d/%p", gotID, gotTarget, id, target)
			}
			return 1
		},
		Audio: func(_ sound.ID, _ *Object) {
			t.Fatal("successful duration emitted audio")
		},
		Instant: func(spell.ID, *Object, *Object, *Object, *SpellAcceptArg, int32) int32 {
			t.Fatal("Firewalk ID was redirected through mutable Effect")
			return 0
		},
		Duration: func(gotID spell.ID, gotSecond, gotThird, gotFourth *Object, gotArg *SpellAcceptArg, gotLevel int32, timeout uint32) int32 {
			if gotID != id || gotSecond != second || gotThird != third || gotFourth != fourth || gotArg != arg || gotLevel != math.MaxInt32 {
				t.Fatalf("duration args = %d/%p/%p/%p/%p/%d", gotID, gotSecond, gotThird, gotFourth, gotArg, gotLevel)
			}
			if timeout != 0x80000003 {
				t.Fatalf("wrapping timeout = %#x, want 0x80000003", timeout)
			}
			return math.MinInt32
		},
		PlasmaTime: func() float64 {
			t.Fatal("Firewalk path loaded plasma time")
			return 0
		},
	})
	if got != math.MinInt32 {
		t.Fatalf("result = %d, want verbatim %d", got, int32(math.MinInt32))
	}
	if captureCalls != 1 {
		t.Fatalf("capture calls = %d, want 1", captureCalls)
	}
	runtime.KeepAlive(second)
	runtime.KeepAlive(third)
	runtime.KeepAlive(fourth)
	runtime.KeepAlive(target)
	runtime.KeepAlive(arg)
}

func TestSpellAcceptServer4FD400UndefinedHoleAndEntryGuards(t *testing.T) {
	s := new(Server)
	failRuntime := SpellAcceptRuntime4FD400{
		CaptureMagic: func(spell.ID, *Object) int32 { t.Fatal("unexpected capture"); return 0 },
		Audio:        func(sound.ID, *Object) { t.Fatal("unexpected audio") },
		Instant: func(spell.ID, *Object, *Object, *Object, *SpellAcceptArg, int32) int32 {
			t.Fatal("unexpected instant callback")
			return 0
		},
		Duration: func(spell.ID, *Object, *Object, *Object, *SpellAcceptArg, int32, uint32) int32 {
			t.Fatal("unexpected duration callback")
			return 0
		},
		PlasmaTime: func() float64 { t.Fatal("unexpected plasma time"); return 0 },
	}
	second := new(Object)
	third := new(Object)
	arg := new(SpellAcceptArg)
	if got := s.SpellAccept4FD400(7, second, third, nil, arg, 0, failRuntime); got != 1 {
		t.Fatalf("undefined selector hole result = %d, want 1", got)
	}
	for _, test := range []struct {
		name          string
		id            spell.ID
		second, third *Object
		arg           *SpellAcceptArg
	}{
		{name: "zero spell", second: second, third: third, arg: arg},
		{name: "nil third", id: 1, second: second, arg: arg},
		{name: "nil second", id: 1, third: third, arg: arg},
		{name: "nil argument", id: 1, second: second, third: third},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := s.SpellAccept4FD400(test.id, test.second, test.third, nil, test.arg, 0, failRuntime); got != 0 {
				t.Fatalf("result = %d, want 0", got)
			}
		})
	}
}
