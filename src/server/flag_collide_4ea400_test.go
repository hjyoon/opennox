package server

import (
	"reflect"
	"testing"
)

type flagCollideTestObject4EA400 struct {
	flags   uint32
	typeInd uint16
	class   uint8
}

func defaultFlagCollideHooks4EA400() flagCollideHooks4EA400[*flagCollideTestObject4EA400, *uint32] {
	return flagCollideHooks4EA400[*flagCollideTestObject4EA400, *uint32]{
		loadFlags:         func(obj *flagCollideTestObject4EA400) uint32 { return obj.flags },
		hasGameFlag:       func(uint32) int32 { return 0 },
		loadGameBallCache: func() uint32 { return 0 },
		lookupGameBall:    func(string) uint32 { return 0 },
		storeGameBall:     func(uint32) {},
		loadTypeInd:       func(obj *flagCollideTestObject4EA400) uint16 { return obj.typeInd },
		loadClassLow:      func(obj *flagCollideTestObject4EA400) uint8 { return obj.class },
		pickupCTF:         func(*flagCollideTestObject4EA400, *flagCollideTestObject4EA400, *uint32) {},
		pickupGameBall:    func(*flagCollideTestObject4EA400, *flagCollideTestObject4EA400, *uint32) {},
	}
}

func TestFlagCollide4EA400NilTargetReturnsBeforeSource(t *testing.T) {
	hooks := defaultFlagCollideHooks4EA400()
	hooks.loadFlags = func(*flagCollideTestObject4EA400) uint32 {
		t.Fatal("nil target loaded flags")
		return 0
	}
	flagCollide4EA400(nil, nil, (*uint32)(nil), hooks)
}

func TestFlagCollide4EA400DeadFlagStopsBeforeModes(t *testing.T) {
	target := &flagCollideTestObject4EA400{flags: 0xffff8000}
	hooks := defaultFlagCollideHooks4EA400()
	hooks.hasGameFlag = func(uint32) int32 {
		t.Fatal("dead target checked a game mode")
		return 0
	}
	flagCollide4EA400(nil, target, (*uint32)(nil), hooks)
}

func TestFlagCollide4EA400CTFHasPriorityAndUsesLiveClass(t *testing.T) {
	source := &flagCollideTestObject4EA400{}
	target := &flagCollideTestObject4EA400{}
	collision := uint32(0xdeadbeef)
	events := make([]string, 0, 4)
	hooks := defaultFlagCollideHooks4EA400()
	hooks.loadFlags = func(obj *flagCollideTestObject4EA400) uint32 {
		events = append(events, "flags")
		target.class = 0
		return obj.flags
	}
	hooks.hasGameFlag = func(mask uint32) int32 {
		events = append(events, "mode32")
		if mask != 0x20 {
			t.Fatalf("mode = %#x", mask)
		}
		target.class = 4
		return -1
	}
	hooks.loadClassLow = func(obj *flagCollideTestObject4EA400) uint8 {
		events = append(events, "class")
		return obj.class
	}
	hooks.pickupCTF = func(gotSource, gotTarget *flagCollideTestObject4EA400, gotCollision *uint32) {
		events = append(events, "ctf")
		if gotSource != source || gotTarget != target || gotCollision != &collision {
			t.Fatal("CTF did not receive the original arguments")
		}
	}
	hooks.loadGameBallCache = func() uint32 { t.Fatal("CTF read FlagBall cache"); return 0 }
	flagCollide4EA400(source, target, &collision, hooks)
	want := []string{"flags", "mode32", "class", "ctf"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestFlagCollide4EA400CTFNonPlayerNeverChecksFlagBall(t *testing.T) {
	target := &flagCollideTestObject4EA400{}
	hooks := defaultFlagCollideHooks4EA400()
	hooks.hasGameFlag = func(mask uint32) int32 {
		if mask != 0x20 {
			t.Fatalf("CTF fallthrough checked mode %#x", mask)
		}
		return 1
	}
	hooks.pickupCTF = func(*flagCollideTestObject4EA400, *flagCollideTestObject4EA400, *uint32) {
		t.Fatal("non-Player entered CTF pickup")
	}
	flagCollide4EA400(nil, target, nil, hooks)
}

func TestFlagCollide4EA400FlagBallCacheLookupPrecedesLiveType(t *testing.T) {
	source := &flagCollideTestObject4EA400{}
	target := &flagCollideTestObject4EA400{typeInd: 1}
	collision := uint32(7)
	events := make([]string, 0, 8)
	hooks := defaultFlagCollideHooks4EA400()
	hooks.loadFlags = func(obj *flagCollideTestObject4EA400) uint32 {
		events = append(events, "flags")
		return obj.flags
	}
	hooks.hasGameFlag = func(mask uint32) int32 {
		events = append(events, map[uint32]string{0x20: "mode32", 0x40: "mode64"}[mask])
		if mask == 0x40 {
			return 1
		}
		return 0
	}
	hooks.loadGameBallCache = func() uint32 {
		events = append(events, "cache")
		return 0
	}
	hooks.lookupGameBall = func(name string) uint32 {
		events = append(events, "lookup")
		if name != "GameBall" {
			t.Fatalf("lookup = %q", name)
		}
		target.typeInd = 0x4321
		return 0x4321
	}
	hooks.storeGameBall = func(ind uint32) {
		events = append(events, "store")
		if ind != 0x4321 {
			t.Fatalf("stored = %#x", ind)
		}
	}
	hooks.loadTypeInd = func(obj *flagCollideTestObject4EA400) uint16 {
		events = append(events, "type")
		return obj.typeInd
	}
	hooks.loadClassLow = func(*flagCollideTestObject4EA400) uint8 {
		t.Fatal("matching type loaded class")
		return 0
	}
	hooks.pickupGameBall = func(gotSource, gotTarget *flagCollideTestObject4EA400, gotCollision *uint32) {
		events = append(events, "ball")
		if gotSource != source || gotTarget != target || gotCollision != &collision {
			t.Fatal("FlagBall did not receive the original arguments")
		}
	}
	flagCollide4EA400(source, target, &collision, hooks)
	want := []string{"flags", "mode32", "mode64", "cache", "lookup", "store", "type", "ball"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestFlagCollide4EA400FlagBallMismatchChecksLivePlayerClass(t *testing.T) {
	target := &flagCollideTestObject4EA400{typeInd: 7}
	hooks := defaultFlagCollideHooks4EA400()
	hooks.hasGameFlag = func(mask uint32) int32 {
		if mask == 0x40 {
			return 1
		}
		return 0
	}
	hooks.loadGameBallCache = func() uint32 { return 8 }
	loads := 0
	hooks.loadClassLow = func(obj *flagCollideTestObject4EA400) uint8 {
		loads++
		obj.class = 4
		return obj.class
	}
	called := false
	hooks.pickupGameBall = func(*flagCollideTestObject4EA400, *flagCollideTestObject4EA400, *uint32) {
		called = true
	}
	flagCollide4EA400(nil, target, nil, hooks)
	if loads != 1 || !called {
		t.Fatalf("class loads/called = %d/%t", loads, called)
	}
}

func TestFlagCollide4EA400ZeroLookupMatchesZeroType(t *testing.T) {
	target := &flagCollideTestObject4EA400{}
	hooks := defaultFlagCollideHooks4EA400()
	hooks.hasGameFlag = func(mask uint32) int32 {
		if mask == 0x40 {
			return -1
		}
		return 0
	}
	hooks.loadClassLow = func(*flagCollideTestObject4EA400) uint8 {
		t.Fatal("zero TypeInd matching zero lookup loaded class")
		return 0
	}
	called := false
	hooks.pickupGameBall = func(*flagCollideTestObject4EA400, *flagCollideTestObject4EA400, *uint32) {
		called = true
	}
	flagCollide4EA400(nil, target, nil, hooks)
	if !called {
		t.Fatal("zero cached lookup did not match zero TypeInd")
	}
}
