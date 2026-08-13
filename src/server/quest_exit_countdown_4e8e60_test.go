package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type questExitTestPlayer4E8E60 struct {
	name  string
	state uint32
}

type questExitTestUpdate4E8E60 struct {
	name   string
	player *questExitTestPlayer4E8E60
	exit   *questExitTestObject4E8E60
}

type questExitTestObject4E8E60 struct {
	name   string
	update *questExitTestUpdate4E8E60
	next   *questExitTestObject4E8E60
}

type questExitTestState4E8E60 struct {
	events         []string
	balance        float64
	timerActive    int32
	remaining      int32
	first          *questExitTestObject4E8E60
	stopResult     int32
	countdown      int32
	sendResult     int32
	mutateExitLoad func(*questExitTestUpdate4E8E60)
}

func (s *questExitTestState4E8E60) event(format string, args ...any) {
	s.events = append(s.events, fmt.Sprintf(format, args...))
}

func (s *questExitTestState4E8E60) hooks() questExitCountdownHooks4E8E60[
	*questExitTestObject4E8E60,
	*questExitTestUpdate4E8E60,
	*questExitTestPlayer4E8E60,
] {
	return questExitCountdownHooks4E8E60[
		*questExitTestObject4E8E60,
		*questExitTestUpdate4E8E60,
		*questExitTestPlayer4E8E60,
	]{
		balanceFloat: func(key string) float64 {
			s.event("balance:%s", key)
			return s.balance
		},
		floatToInt: func(value float32) int32 {
			s.event("round:%08x", math.Float32bits(value))
			return questExitRound4E8E60(value)
		},
		timerActive: func() int32 {
			s.event("timer-active")
			return s.timerActive
		},
		timerRemainingMillis: func() int32 {
			s.event("timer-remaining")
			return s.remaining
		},
		firstUnit: func() *questExitTestObject4E8E60 {
			s.event("first")
			return s.first
		},
		nextUnit: func(obj *questExitTestObject4E8E60) *questExitTestObject4E8E60 {
			s.event("next:%s", obj.name)
			return obj.next
		},
		loadUpdateData: func(obj *questExitTestObject4E8E60) *questExitTestUpdate4E8E60 {
			s.event("update:%s", obj.name)
			return obj.update
		},
		loadPlayer: func(update *questExitTestUpdate4E8E60) *questExitTestPlayer4E8E60 {
			s.event("player:%s", update.name)
			return update.player
		},
		loadQuestState: func(player *questExitTestPlayer4E8E60) uint32 {
			s.event("state:%s", player.name)
			return player.state
		},
		loadQuestExit: func(update *questExitTestUpdate4E8E60) *questExitTestObject4E8E60 {
			s.event("exit:%s", update.name)
			if s.mutateExitLoad != nil {
				s.mutateExitLoad(update)
			}
			return update.exit
		},
		stopTimer: func(value int32) int32 {
			s.event("stop:%d", value)
			return s.stopResult
		},
		countdownStarted: func() int32 {
			s.event("countdown")
			return s.countdown
		},
		startCountdown: func(seconds int32, id string) {
			s.event("start:%d:%s", seconds, id)
		},
		sendGauntlet: func(recipient int32) int32 {
			s.event("send:%d", recipient)
			return s.sendResult
		},
	}
}

func questExitTestUnit4E8E60(name string, state uint32, ready bool) *questExitTestObject4E8E60 {
	player := &questExitTestPlayer4E8E60{name: name, state: state}
	update := &questExitTestUpdate4E8E60{name: name, player: player}
	unit := &questExitTestObject4E8E60{name: name, update: update}
	if ready {
		update.exit = &questExitTestObject4E8E60{name: name + "-exit"}
	}
	return unit
}

func TestQuestExitCountdown4E8E60EmptyStopsAfterBalanceAndTimerRead(t *testing.T) {
	state := &questExitTestState4E8E60{balance: 12.5, stopResult: math.MinInt32 + 17}
	got := questExitCountdown4E8E60(state.hooks())
	if got != state.stopResult {
		t.Fatalf("result = %d, want %d", got, state.stopResult)
	}
	want := []string{
		"balance:QuestExitTimerStart", "round:41480000", "timer-active", "first", "stop:0",
	}
	if !reflect.DeepEqual(state.events, want) {
		t.Fatalf("events = %q, want %q", state.events, want)
	}
}

func TestQuestExitCountdown4E8E60ActiveTimerUsesSignedTruncation(t *testing.T) {
	unit := questExitTestUnit4E8E60("inactive", 0, true)
	state := &questExitTestState4E8E60{
		balance: 7.5, timerActive: -1, remaining: -1999, first: unit, stopResult: -73,
	}
	if got := questExitCountdown4E8E60(state.hooks()); got != -73 {
		t.Fatalf("result = %d, want -73", got)
	}
	want := []string{
		"balance:QuestExitTimerStart", "round:40f00000", "timer-active", "timer-remaining", "first",
		"update:inactive", "player:inactive", "state:inactive", "next:inactive", "stop:0",
	}
	if !reflect.DeepEqual(state.events, want) {
		t.Fatalf("events = %q, want %q", state.events, want)
	}
}

func TestQuestExitCountdown4E8E60ActiveTimerStartsTruncatedNegativeSecond(t *testing.T) {
	unit := questExitTestUnit4E8E60("player", 1, false)
	state := &questExitTestState4E8E60{
		balance: 7.5, timerActive: 2, remaining: -1999, first: unit, sendResult: 81,
	}
	if got := questExitCountdown4E8E60(state.hooks()); got != 81 {
		t.Fatalf("result = %d, want 81", got)
	}
	wantTail := []string{
		"round:80000000", "countdown", "start:-1:objcoll.c:ExitCountdown", "send:255",
	}
	if got := state.events[len(state.events)-len(wantTail):]; !reflect.DeepEqual(got, wantTail) {
		t.Fatalf("event tail = %q, want %q", got, wantTail)
	}
}

func TestQuestExitCountdown4E8E60ExactStateRatioAndLiveSuccessor(t *testing.T) {
	a := questExitTestUnit4E8E60("a", 1, true)
	b := questExitTestUnit4E8E60("b", 1, true)
	c := questExitTestUnit4E8E60("c", 1, false)
	d := questExitTestUnit4E8E60("d", 2, true)
	a.next, b.next, c.next = b, c, d
	state := &questExitTestState4E8E60{
		balance: 10, first: a, sendResult: math.MinInt32 + 29,
		mutateExitLoad: func(update *questExitTestUpdate4E8E60) {
			if update == a.update {
				a.next = c
			}
		},
	}
	got := questExitCountdown4E8E60(state.hooks())
	if got != state.sendResult {
		t.Fatalf("result = %d, want %d", got, state.sendResult)
	}
	want := []string{
		"balance:QuestExitTimerStart", "round:41200000", "timer-active", "first",
		"update:a", "player:a", "state:a", "exit:a", "next:a",
		"update:c", "player:c", "state:c", "exit:c", "next:c",
		"update:d", "player:d", "state:d", "next:d",
		"round:40a00000", "start:5:objcoll.c:ExitCountdown", "send:255",
	}
	if !reflect.DeepEqual(state.events, want) {
		t.Fatalf("events = %q, want %q", state.events, want)
	}
}

func TestQuestExitCountdown4E8E60ZeroReadyReturnsLiveCountdownResult(t *testing.T) {
	unit := questExitTestUnit4E8E60("player", 1, false)
	state := &questExitTestState4E8E60{balance: 9, first: unit, countdown: -7, sendResult: 91}
	if got := questExitCountdown4E8E60(state.hooks()); got != -7 {
		t.Fatalf("result = %d, want -7", got)
	}
	wantTail := []string{"round:00000000", "countdown"}
	if got := state.events[len(state.events)-len(wantTail):]; !reflect.DeepEqual(got, wantTail) {
		t.Fatalf("event tail = %q, want %q", got, wantTail)
	}
}

func TestQuestExitCountdown4E8E60ZeroReadyStartsFullTimerWhenIdle(t *testing.T) {
	unit := questExitTestUnit4E8E60("player", 1, false)
	state := &questExitTestState4E8E60{balance: 9, first: unit, sendResult: 27}
	if got := questExitCountdown4E8E60(state.hooks()); got != 27 {
		t.Fatalf("result = %d, want 27", got)
	}
	wantTail := []string{
		"round:00000000", "countdown", "start:9:objcoll.c:ExitCountdown", "send:255",
	}
	if got := state.events[len(state.events)-len(wantTail):]; !reflect.DeepEqual(got, wantTail) {
		t.Fatalf("event tail = %q, want %q", got, wantTail)
	}
}

func TestQuestExitCountdown4E8E60RoundMatchesX87FISTP(t *testing.T) {
	tests := []struct {
		value float32
		want  int32
	}{
		{0.5, 0}, {1.5, 2}, {2.5, 2}, {-1.5, -2}, {-2.5, -2},
		{math.Float32frombits(0x4effffff), 2147483520},
		{math.Float32frombits(0x4f000000), math.MinInt32},
		{math.Float32frombits(0xcf000000), math.MinInt32},
		{float32(math.Inf(1)), math.MinInt32},
		{float32(math.Inf(-1)), math.MinInt32},
		{float32(math.NaN()), math.MinInt32},
	}
	for _, tc := range tests {
		if got := questExitRound4E8E60(tc.value); got != tc.want {
			t.Errorf("round(%08x) = %d, want %d", math.Float32bits(tc.value), got, tc.want)
		}
	}
}

func TestQuestExitCountdown4E8E60InvalidBalanceKeepsWrappingBranch(t *testing.T) {
	unit := questExitTestUnit4E8E60("player", 1, true)
	state := &questExitTestState4E8E60{
		balance: math.Inf(1), first: unit, countdown: math.MaxInt32, sendResult: 33,
	}
	if got := questExitCountdown4E8E60(state.hooks()); got != math.MaxInt32 {
		t.Fatalf("result = %d, want %d", got, int32(math.MaxInt32))
	}
	wantTail := []string{"round:cf000000", "countdown"}
	if got := state.events[len(state.events)-len(wantTail):]; !reflect.DeepEqual(got, wantTail) {
		t.Fatalf("event tail = %q, want %q", got, wantTail)
	}
}
