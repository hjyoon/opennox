package server

import (
	"fmt"
	"reflect"
	"testing"
)

type mimicCollideTestObject4E83D0 struct {
	name     string
	flags    uint32
	classLow uint8
	posXBits uint32
	posYBits uint32
}

type mimicCollideTestAction4E83D0 struct {
	name string
	args [4]uint32
}

type mimicCollideTestPoint4E83D0 struct {
	id uint32
}

func TestMimicCollide4E83D0SuccessOrderLiveReadsAndExactReturn(t *testing.T) {
	mimic := &mimicCollideTestObject4E83D0{name: "mimic"}
	other := &mimicCollideTestObject4E83D0{
		name:     "other",
		classLow: 0x84,
		posXBits: 0x3f800000,
		posYBits: 0x40000000,
	}
	underAttack := &mimicCollideTestAction4E83D0{name: "under", args: [4]uint32{9, 10, 11, 12}}
	fight := &mimicCollideTestAction4E83D0{name: "fight", args: [4]uint32{13, 14, 15, 16}}
	point := &mimicCollideTestPoint4E83D0{id: 0x10203040}
	result := &mimicCollideTestAction4E83D0{name: "result"}
	var events []string
	pushes := 0
	frames := []uint32{0x01020304, 0xfefdfcfb}

	got := mimicCollide4E83D0(mimic, other, point, mimicCollideHooks4E83D0[
		*mimicCollideTestObject4E83D0,
		*mimicCollideTestAction4E83D0,
		*mimicCollideTestPoint4E83D0,
		*mimicCollideTestAction4E83D0,
	]{
		flags: func(got *mimicCollideTestObject4E83D0) uint32 {
			events = append(events, "flags")
			if got != other {
				t.Fatalf("flags object = %p, want %p", got, other)
			}
			return got.flags
		},
		classLow: func(got *mimicCollideTestObject4E83D0) uint8 {
			events = append(events, "class")
			return got.classLow
		},
		isEnemy: func(gotMimic, gotOther *mimicCollideTestObject4E83D0) int32 {
			events = append(events, "enemy")
			if gotMimic != mimic || gotOther != other {
				t.Fatalf("enemy args = (%p, %p), want (%p, %p)", gotMimic, gotOther, mimic, other)
			}
			return -7
		},
		actionScheduled: func(got *mimicCollideTestObject4E83D0, action uint32) int32 {
			events = append(events, fmt.Sprintf("scheduled:%d", action))
			if got != mimic {
				t.Fatalf("scheduled object = %p, want %p", got, mimic)
			}
			return 0
		},
		pushAction: func(got *mimicCollideTestObject4E83D0, action uint32) *mimicCollideTestAction4E83D0 {
			events = append(events, fmt.Sprintf("push:%d", action))
			if got != mimic {
				t.Fatalf("push object = %p, want %p", got, mimic)
			}
			pushes++
			if pushes == 1 {
				return underAttack
			}
			other.posXBits = 0x7fa12345
			return fight
		},
		frame: func() uint32 {
			events = append(events, "frame")
			value := frames[0]
			frames = frames[1:]
			return value
		},
		storeActionArg: func(action *mimicCollideTestAction4E83D0, index int, value uint32) {
			events = append(events, fmt.Sprintf("store:%s:%d:%08x", action.name, index, value))
			action.args[index] = value
		},
		posXBits: func(got *mimicCollideTestObject4E83D0) uint32 {
			events = append(events, "pos-x")
			got.posYBits = 0x80000000
			return got.posXBits
		},
		posYBits: func(got *mimicCollideTestObject4E83D0) uint32 {
			events = append(events, "pos-y")
			return got.posYBits
		},
		monsterCollide: func(gotMimic, gotOther *mimicCollideTestObject4E83D0, gotPoint *mimicCollideTestPoint4E83D0) *mimicCollideTestAction4E83D0 {
			events = append(events, "collide")
			if gotMimic != mimic || gotOther != other || gotPoint != point {
				t.Fatalf("collide args = (%p, %p, %p), want (%p, %p, %p)", gotMimic, gotOther, gotPoint, mimic, other, point)
			}
			return result
		},
	})

	if got != result {
		t.Fatalf("result = %p, want exact %p", got, result)
	}
	wantEvents := []string{
		"flags", "class", "enemy", "scheduled:15",
		"push:43", "frame", "store:under:0:01020304",
		"push:15", "pos-x", "store:fight:0:7fa12345",
		"pos-y", "store:fight:1:80000000", "frame",
		"store:fight:2:fefdfcfb", "collide",
	}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %v, want %v", events, wantEvents)
	}
	if want := [4]uint32{0x01020304, 10, 11, 12}; underAttack.args != want {
		t.Fatalf("under-attack args = %#v, want %#v", underAttack.args, want)
	}
	if want := [4]uint32{0x7fa12345, 0x80000000, 0xfefdfcfb, 16}; fight.args != want {
		t.Fatalf("fight args = %#v, want %#v", fight.args, want)
	}
	if len(frames) != 0 {
		t.Fatalf("unused frame results = %v", frames)
	}
}

func TestMimicCollide4E83D0GuardsStillForwardCollision(t *testing.T) {
	tests := []struct {
		name       string
		other      *mimicCollideTestObject4E83D0
		enemy      int32
		scheduled  int32
		wantEvents []string
	}{
		{name: "nil other", other: nil, wantEvents: []string{"collide"}},
		{name: "dead", other: &mimicCollideTestObject4E83D0{flags: 0xffff8000, classLow: 6}, wantEvents: []string{"flags", "collide"}},
		{name: "non-unit", other: &mimicCollideTestObject4E83D0{classLow: 0xf8}, wantEvents: []string{"flags", "class", "collide"}},
		{name: "not enemy exact zero", other: &mimicCollideTestObject4E83D0{classLow: 2}, enemy: 0, wantEvents: []string{"flags", "class", "enemy", "collide"}},
		{name: "fight already scheduled any nonzero", other: &mimicCollideTestObject4E83D0{classLow: 4}, enemy: -9, scheduled: -3, wantEvents: []string{"flags", "class", "enemy", "scheduled", "collide"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mimic := &mimicCollideTestObject4E83D0{name: "mimic"}
			point := &mimicCollideTestPoint4E83D0{id: 7}
			var events []string
			got := mimicCollide4E83D0(mimic, tc.other, point, mimicCollideHooks4E83D0[
				*mimicCollideTestObject4E83D0,
				*mimicCollideTestAction4E83D0,
				*mimicCollideTestPoint4E83D0,
				int32,
			]{
				flags: func(obj *mimicCollideTestObject4E83D0) uint32 {
					events = append(events, "flags")
					return obj.flags
				},
				classLow: func(obj *mimicCollideTestObject4E83D0) uint8 {
					events = append(events, "class")
					return obj.classLow
				},
				isEnemy: func(*mimicCollideTestObject4E83D0, *mimicCollideTestObject4E83D0) int32 {
					events = append(events, "enemy")
					return tc.enemy
				},
				actionScheduled: func(*mimicCollideTestObject4E83D0, uint32) int32 {
					events = append(events, "scheduled")
					return tc.scheduled
				},
				pushAction: func(*mimicCollideTestObject4E83D0, uint32) *mimicCollideTestAction4E83D0 {
					t.Fatal("pushAction called across a guard")
					return nil
				},
				frame: func() uint32 {
					t.Fatal("frame called across a guard")
					return 0
				},
				storeActionArg: func(*mimicCollideTestAction4E83D0, int, uint32) {
					t.Fatal("storeActionArg called across a guard")
				},
				posXBits: func(*mimicCollideTestObject4E83D0) uint32 {
					t.Fatal("posXBits called across a guard")
					return 0
				},
				posYBits: func(*mimicCollideTestObject4E83D0) uint32 {
					t.Fatal("posYBits called across a guard")
					return 0
				},
				monsterCollide: func(gotMimic, gotOther *mimicCollideTestObject4E83D0, gotPoint *mimicCollideTestPoint4E83D0) int32 {
					events = append(events, "collide")
					if gotMimic != mimic || gotOther != tc.other || gotPoint != point {
						t.Fatal("collision arguments changed")
					}
					return -0x1020304
				},
			})
			if got != -0x1020304 {
				t.Fatalf("result = %d, want %d", got, -0x1020304)
			}
			if !reflect.DeepEqual(events, tc.wantEvents) {
				t.Fatalf("events = %v, want %v", events, tc.wantEvents)
			}
		})
	}
}

func TestMimicCollide4E83D0PushFailuresAreIndependent(t *testing.T) {
	tests := []struct {
		name       string
		first      *mimicCollideTestAction4E83D0
		second     *mimicCollideTestAction4E83D0
		wantEvents []string
	}{
		{
			name:   "under attack push fails",
			second: &mimicCollideTestAction4E83D0{name: "fight"},
			wantEvents: []string{
				"push:43", "push:15", "pos-x", "store:fight:0", "pos-y", "store:fight:1", "frame", "store:fight:2", "collide",
			},
		},
		{
			name:  "fight push fails",
			first: &mimicCollideTestAction4E83D0{name: "under"},
			wantEvents: []string{
				"push:43", "frame", "store:under:0", "push:15", "collide",
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mimic := &mimicCollideTestObject4E83D0{}
			other := &mimicCollideTestObject4E83D0{classLow: 6, posXBits: 1, posYBits: 2}
			var events []string
			pushes := 0
			got := mimicCollide4E83D0(mimic, other, 99, mimicCollideHooks4E83D0[
				*mimicCollideTestObject4E83D0,
				*mimicCollideTestAction4E83D0,
				int,
				int,
			]{
				flags:           func(*mimicCollideTestObject4E83D0) uint32 { return 0 },
				classLow:        func(*mimicCollideTestObject4E83D0) uint8 { return 6 },
				isEnemy:         func(*mimicCollideTestObject4E83D0, *mimicCollideTestObject4E83D0) int32 { return 1 },
				actionScheduled: func(*mimicCollideTestObject4E83D0, uint32) int32 { return 0 },
				pushAction: func(_ *mimicCollideTestObject4E83D0, actionValue uint32) *mimicCollideTestAction4E83D0 {
					pushes++
					action := tc.first
					if pushes == 2 {
						action = tc.second
					}
					events = append(events, fmt.Sprintf("push:%d", actionValue))
					return action
				},
				frame: func() uint32 {
					events = append(events, "frame")
					return 3
				},
				storeActionArg: func(action *mimicCollideTestAction4E83D0, index int, value uint32) {
					events = append(events, fmt.Sprintf("store:%s:%d", action.name, index))
					action.args[index] = value
				},
				posXBits: func(*mimicCollideTestObject4E83D0) uint32 {
					events = append(events, "pos-x")
					return 1
				},
				posYBits: func(*mimicCollideTestObject4E83D0) uint32 {
					events = append(events, "pos-y")
					return 2
				},
				monsterCollide: func(*mimicCollideTestObject4E83D0, *mimicCollideTestObject4E83D0, int) int {
					events = append(events, "collide")
					return 77
				},
			})
			if got != 77 {
				t.Fatalf("result = %d, want 77", got)
			}
			if !reflect.DeepEqual(events, tc.wantEvents) {
				t.Fatalf("events = %v, want %v", events, tc.wantEvents)
			}
		})
	}
}

func TestMimicCollide4E83D0DoesNotGuardNilMimic(t *testing.T) {
	other := &mimicCollideTestObject4E83D0{classLow: 2}
	point := &mimicCollideTestPoint4E83D0{id: 0xa5}
	enemyCalled := false
	got := mimicCollide4E83D0(
		(*mimicCollideTestObject4E83D0)(nil),
		other,
		point,
		mimicCollideHooks4E83D0[
			*mimicCollideTestObject4E83D0,
			*mimicCollideTestAction4E83D0,
			*mimicCollideTestPoint4E83D0,
			uint32,
		]{
			flags:    func(*mimicCollideTestObject4E83D0) uint32 { return 0 },
			classLow: func(*mimicCollideTestObject4E83D0) uint8 { return 2 },
			isEnemy: func(mimic, gotOther *mimicCollideTestObject4E83D0) int32 {
				enemyCalled = true
				if mimic != nil || gotOther != other {
					t.Fatal("nil Mimic argument was changed")
				}
				return 0
			},
			actionScheduled: func(*mimicCollideTestObject4E83D0, uint32) int32 {
				t.Fatal("scheduled called after zero enemy result")
				return 0
			},
			pushAction: func(*mimicCollideTestObject4E83D0, uint32) *mimicCollideTestAction4E83D0 {
				t.Fatal("push called after zero enemy result")
				return nil
			},
			frame: func() uint32 {
				t.Fatal("frame called after zero enemy result")
				return 0
			},
			storeActionArg: func(*mimicCollideTestAction4E83D0, int, uint32) {
				t.Fatal("action argument stored after zero enemy result")
			},
			posXBits: func(*mimicCollideTestObject4E83D0) uint32 {
				t.Fatal("X read after zero enemy result")
				return 0
			},
			posYBits: func(*mimicCollideTestObject4E83D0) uint32 {
				t.Fatal("Y read after zero enemy result")
				return 0
			},
			monsterCollide: func(mimic, gotOther *mimicCollideTestObject4E83D0, gotPoint *mimicCollideTestPoint4E83D0) uint32 {
				if mimic != nil || gotOther != other || gotPoint != point {
					t.Fatal("final nil Mimic arguments changed")
				}
				return 0xfedcba98
			},
		},
	)
	if !enemyCalled {
		t.Fatal("enemy callback was not called with nil Mimic")
	}
	if got != 0xfedcba98 {
		t.Fatalf("result = %#x, want %#x", got, uint32(0xfedcba98))
	}
}
