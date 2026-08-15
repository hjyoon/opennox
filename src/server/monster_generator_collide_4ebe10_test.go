package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type monsterGeneratorCollideTestState4EBE10 struct {
	events   []string
	classLow uint8
	update   int
	block    int
	result   int64
}

func (s *monsterGeneratorCollideTestState4EBE10) event(format string, args ...any) {
	s.events = append(s.events, fmt.Sprintf(format, args...))
}

func (s *monsterGeneratorCollideTestState4EBE10) hooks() monsterGeneratorCollideHooks4EBE10[int, int, int, int64] {
	return monsterGeneratorCollideHooks4EBE10[int, int, int, int64]{
		loadTargetClassLow: func(target int) uint8 {
			s.event("class:%d=%#x", target, s.classLow)
			return s.classLow
		},
		loadSourceUpdate: func(source int) int {
			s.event("update:%d=%d", source, s.update)
			return s.update
		},
		collisionBlock: func(update int) int {
			s.event("block:%d=%d", update, s.block)
			return s.block
		},
		scriptCallback: func(block, caller, trigger int) int64 {
			s.event("call:%d:%d:%d", block, caller, trigger)
			return s.result
		},
	}
}

func assertMonsterGeneratorCollideEvents4EBE10(t *testing.T, got, want []string) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("events:\n got %#v\nwant %#v", got, want)
	}
}

func TestMonsterGeneratorCollide4EBE10NilTargetReadsNothing(t *testing.T) {
	state := &monsterGeneratorCollideTestState4EBE10{
		classLow: monsterGeneratorPlayerClassLow4EBE10,
		update:   31,
		block:    41,
	}
	monsterGeneratorCollide4EBE10(0, 0, state.hooks())
	assertMonsterGeneratorCollideEvents4EBE10(t, state.events, nil)
}

func TestMonsterGeneratorCollide4EBE10NonPlayerStopsBeforeSource(t *testing.T) {
	state := &monsterGeneratorCollideTestState4EBE10{
		classLow: 0xfb,
		update:   31,
		block:    41,
	}
	monsterGeneratorCollide4EBE10(11, 21, state.hooks())
	assertMonsterGeneratorCollideEvents4EBE10(t, state.events, []string{"class:21=0xfb"})
}

func TestMonsterGeneratorCollide4EBE10OrderArgumentsAndIgnoredReturn(t *testing.T) {
	for _, result := range []int64{math.MinInt64, -1, 0, 1, math.MaxInt64} {
		state := &monsterGeneratorCollideTestState4EBE10{
			classLow: 0x84,
			update:   31,
			block:    41,
			result:   result,
		}
		monsterGeneratorCollide4EBE10(11, 21, state.hooks())
		assertMonsterGeneratorCollideEvents4EBE10(t, state.events, []string{
			"class:21=0x84",
			"update:11=31",
			"block:31=41",
			"call:41:21:11",
		})
	}
}

func TestMonsterGeneratorCollide4EBE10ZeroUpdateStillReachesBlockAndCallback(t *testing.T) {
	state := &monsterGeneratorCollideTestState4EBE10{
		classLow: monsterGeneratorPlayerClassLow4EBE10,
		update:   0,
		block:    72,
	}
	monsterGeneratorCollide4EBE10(11, 21, state.hooks())
	assertMonsterGeneratorCollideEvents4EBE10(t, state.events, []string{
		"class:21=0x4",
		"update:11=0",
		"block:0=72",
		"call:72:21:11",
	})
}

func TestMonsterGeneratorCollide4EBE10SourceFaultFollowsTargetGate(t *testing.T) {
	events := []string{}
	defer func() {
		if recover() == nil {
			t.Fatal("nil source update load returned without panic")
		}
		assertMonsterGeneratorCollideEvents4EBE10(t, events, []string{"class:21", "update:0"})
	}()
	monsterGeneratorCollide4EBE10(0, 21, monsterGeneratorCollideHooks4EBE10[int, int, int, int]{
		loadTargetClassLow: func(target int) uint8 {
			events = append(events, fmt.Sprintf("class:%d", target))
			return monsterGeneratorPlayerClassLow4EBE10
		},
		loadSourceUpdate: func(source int) int {
			events = append(events, fmt.Sprintf("update:%d", source))
			panic("source update-data load")
		},
		collisionBlock: func(int) int {
			t.Fatal("source fault reached block calculation")
			return 0
		},
		scriptCallback: func(int, int, int) int {
			t.Fatal("source fault reached callback")
			return 0
		},
	})
}
