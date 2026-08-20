package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type soloMonsterKillRewardWorld4EE500 struct {
	flag           int32
	attribution    map[string]string
	parent         map[string]string
	classLow       map[string]uint8
	owner          map[string]string
	experience     map[string]float32
	monitored      int32
	awarded        float64
	message        string
	events         []string
	faultAt        int
	afterEvent     map[string]func()
	givePlayer     string
	giveExperience float32
}

func newSoloMonsterKillRewardWorld4EE500() *soloMonsterKillRewardWorld4EE500 {
	return &soloMonsterKillRewardWorld4EE500{
		flag:        1,
		attribution: make(map[string]string),
		parent:      make(map[string]string),
		classLow:    make(map[string]uint8),
		owner:       make(map[string]string),
		experience:  make(map[string]float32),
		monitored:   1,
		awarded:     5.75,
		message:     "localized-gainpoints",
		afterEvent:  make(map[string]func()),
	}
}

func (w *soloMonsterKillRewardWorld4EE500) record(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
	if after := w.afterEvent[event]; after != nil {
		after()
	}
}

func (w *soloMonsterKillRewardWorld4EE500) hooks() soloMonsterKillRewardHooks4EE500[string, string] {
	return soloMonsterKillRewardHooks4EE500[string, string]{
		gameFlag: func(flag uint32) int32 {
			value := w.flag
			w.record(fmt.Sprintf("game-flag:%08x=%d", flag, value))
			return value
		},
		loadAttribution: func(killed string) string {
			value := w.attribution[killed]
			w.record("load-attribution:" + killed + "=" + value)
			return value
		},
		findParent: func(attribution string) string {
			value := w.parent[attribution]
			w.record("find-parent:" + attribution + "=" + value)
			return value
		},
		loadClassLow: func(obj string) uint8 {
			value := w.classLow[obj]
			w.record(fmt.Sprintf("load-class:%s=%02x", obj, value))
			if obj == "" {
				panic("nil-object-class")
			}
			return value
		},
		isMonitored: func(player, monster string) int32 {
			value := w.monitored
			w.record(fmt.Sprintf("is-monitored:%s:%s=%d", player, monster, value))
			return value
		},
		loadOwner: func(obj string) string {
			value := w.owner[obj]
			w.record("load-owner:" + obj + "=" + value)
			if obj == "" {
				panic("nil-object-owner")
			}
			return value
		},
		loadExperience: func(killed string) float32 {
			value := w.experience[killed]
			w.record(fmt.Sprintf("load-experience:%s=%08x", killed, math.Float32bits(value)))
			return value
		},
		giveXP: func(player string, experience float32) float64 {
			w.givePlayer = player
			w.giveExperience = experience
			value := w.awarded
			w.record(fmt.Sprintf("give-xp:%s:%08x=%016x", player, math.Float32bits(experience), math.Float64bits(value)))
			return value
		},
		loadString: func(key, path string, line int) string {
			message := w.message
			w.record(fmt.Sprintf("load-string:%s:%s:%d=%s", key, path, line, message))
			return message
		},
		sendLineMessage: func(player, message string, points uint32) {
			w.record(fmt.Sprintf("send-line:%s:%s:%08x", player, message, points))
		},
	}
}

func configureSoloMonsterKillRewardHappy4EE500(w *soloMonsterKillRewardWorld4EE500) {
	w.attribution["killed"] = "monster"
	w.parent["monster"] = "player"
	w.classLow["player"] = soloMonsterKillRewardPlayerBit4EE500
	w.classLow["monster"] = soloMonsterKillRewardMonsterBit4EE500
	w.owner["monster"] = "player"
	w.experience["killed"] = 12.5
}

func TestSoloMonsterKillReward4EE500EarlyGatesAndParentFault(t *testing.T) {
	w := newSoloMonsterKillRewardWorld4EE500()
	soloMonsterKillReward4EE500("", w.hooks())
	if len(w.events) != 0 {
		t.Fatalf("nil events = %v, want none", w.events)
	}

	w = newSoloMonsterKillRewardWorld4EE500()
	w.flag = 0
	soloMonsterKillReward4EE500("killed", w.hooks())
	if want := []string{"game-flag:00000800=0"}; !reflect.DeepEqual(w.events, want) {
		t.Fatalf("flag events = %v, want %v", w.events, want)
	}

	w = newSoloMonsterKillRewardWorld4EE500()
	soloMonsterKillReward4EE500("killed", w.hooks())
	if want := []string{"game-flag:00000800=1", "load-attribution:killed="}; !reflect.DeepEqual(w.events, want) {
		t.Fatalf("attribution events = %v, want %v", w.events, want)
	}

	w = newSoloMonsterKillRewardWorld4EE500()
	w.attribution["killed"] = "orphan"
	defer func() {
		if got := recover(); got != "nil-object-class" {
			t.Fatalf("panic = %v, want nil-object-class", got)
		}
		want := []string{
			"game-flag:00000800=1",
			"load-attribution:killed=orphan",
			"find-parent:orphan=",
			"load-class:=00",
		}
		if !reflect.DeepEqual(w.events, want) {
			t.Fatalf("parent fault events = %v, want %v", w.events, want)
		}
	}()
	soloMonsterKillReward4EE500("killed", w.hooks())
}

func TestSoloMonsterKillReward4EE500RejectsNonPlayerParent(t *testing.T) {
	w := newSoloMonsterKillRewardWorld4EE500()
	w.attribution["killed"] = "owner"
	w.parent["owner"] = "terminal"
	w.classLow["terminal"] = soloMonsterKillRewardMonsterBit4EE500
	soloMonsterKillReward4EE500("killed", w.hooks())
	want := []string{
		"game-flag:00000800=1",
		"load-attribution:killed=owner",
		"find-parent:owner=terminal",
		"load-class:terminal=02",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
}

func TestSoloMonsterKillReward4EE500DirectPlayerSkipsChain(t *testing.T) {
	w := newSoloMonsterKillRewardWorld4EE500()
	w.attribution["killed"] = "player"
	w.parent["player"] = "player"
	w.classLow["player"] = soloMonsterKillRewardPlayerBit4EE500
	w.experience["killed"] = 3.25
	w.awarded = 0
	soloMonsterKillReward4EE500("killed", w.hooks())
	want := []string{
		"game-flag:00000800=1",
		"load-attribution:killed=player",
		"find-parent:player=player",
		"load-class:player=04",
		"load-experience:killed=40500000",
		"give-xp:player:40500000=0000000000000000",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
}

func TestSoloMonsterKillReward4EE500NoMonsterChainIsAllowed(t *testing.T) {
	w := newSoloMonsterKillRewardWorld4EE500()
	w.attribution["killed"] = "first"
	w.parent["first"] = "player"
	w.classLow["player"] = soloMonsterKillRewardPlayerBit4EE500
	w.owner["first"] = "second"
	w.owner["second"] = "player"
	w.experience["killed"] = 1
	soloMonsterKillReward4EE500("killed", w.hooks())
	want := []string{
		"game-flag:00000800=1",
		"load-attribution:killed=first",
		"find-parent:first=player",
		"load-class:player=04",
		"load-class:first=00",
		"load-owner:first=second",
		"load-class:second=00",
		"load-owner:second=player",
		"load-experience:killed=3f800000",
		"give-xp:player:3f800000=4017000000000000",
		"load-string:gainpoints:C:\\NoxPost\\src\\Server\\Object\\health.c:172=localized-gainpoints",
		"send-line:player:localized-gainpoints:00000005",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
}

func TestSoloMonsterKillReward4EE500NilOwnerFaultsOnNextClassRead(t *testing.T) {
	w := newSoloMonsterKillRewardWorld4EE500()
	w.attribution["killed"] = "first"
	w.parent["first"] = "player"
	w.classLow["player"] = soloMonsterKillRewardPlayerBit4EE500
	defer func() {
		if got := recover(); got != "nil-object-class" {
			t.Fatalf("panic = %v, want nil-object-class", got)
		}
		want := []string{
			"game-flag:00000800=1",
			"load-attribution:killed=first",
			"find-parent:first=player",
			"load-class:player=04",
			"load-class:first=00",
			"load-owner:first=",
			"load-class:=00",
		}
		if !reflect.DeepEqual(w.events, want) {
			t.Fatalf("events = %v, want %v", w.events, want)
		}
	}()
	soloMonsterKillReward4EE500("killed", w.hooks())
}

func TestSoloMonsterKillReward4EE500FirstMonsterLoadsLiveOwnerThenStops(t *testing.T) {
	for _, monitored := range []int32{0, 1, -1} {
		t.Run(fmt.Sprintf("monitored-%d", monitored), func(t *testing.T) {
			w := newSoloMonsterKillRewardWorld4EE500()
			configureSoloMonsterKillRewardHappy4EE500(w)
			w.monitored = monitored
			w.classLow["second-monster"] = soloMonsterKillRewardMonsterBit4EE500
			event := fmt.Sprintf("is-monitored:player:monster=%d", monitored)
			w.afterEvent[event] = func() {
				w.owner["monster"] = "second-monster"
			}
			soloMonsterKillReward4EE500("killed", w.hooks())

			prefix := []string{
				"game-flag:00000800=1",
				"load-attribution:killed=monster",
				"find-parent:monster=player",
				"load-class:player=04",
				"load-class:monster=02",
				event,
				"load-owner:monster=second-monster",
			}
			if len(w.events) < len(prefix) || !reflect.DeepEqual(w.events[:len(prefix)], prefix) {
				t.Fatalf("events prefix = %v, want %v", w.events, prefix)
			}
			for _, got := range w.events {
				if got == "load-class:second-monster=02" || got == "load-owner:second-monster=" {
					t.Fatalf("second Monster was inspected: %v", w.events)
				}
			}
			if monitored == 0 && len(w.events) != len(prefix) {
				t.Fatalf("rejected events = %v, want %v", w.events, prefix)
			}
			if monitored != 0 && w.givePlayer != "player" {
				t.Fatalf("give player = %q, want player", w.givePlayer)
			}
		})
	}
}

func TestSoloMonsterKillReward4EE500CachesExperienceAndPlayer(t *testing.T) {
	w := newSoloMonsterKillRewardWorld4EE500()
	configureSoloMonsterKillRewardHappy4EE500(w)
	w.afterEvent["load-experience:killed=41480000"] = func() {
		w.experience["killed"] = 99
		w.parent["monster"] = "replacement-player"
	}
	soloMonsterKillReward4EE500("killed", w.hooks())
	if w.givePlayer != "player" || math.Float32bits(w.giveExperience) != 0x41480000 {
		t.Fatalf("give args = (%q,%08x), want (player,41480000)", w.givePlayer, math.Float32bits(w.giveExperience))
	}
}

func TestSoloMonsterKillReward4EE500PositiveGateAndConversion(t *testing.T) {
	for _, tc := range []struct {
		name       string
		awarded    float64
		wantSend   bool
		wantPoints uint32
	}{
		{name: "negative", awarded: -1},
		{name: "negative-zero", awarded: math.Copysign(0, -1)},
		{name: "positive-zero", awarded: 0},
		{name: "nan", awarded: math.NaN()},
		{name: "negative-infinity", awarded: math.Inf(-1)},
		{name: "fraction", awarded: 3.875, wantSend: true, wantPoints: 3},
		{name: "low-word", awarded: float64(uint64(1)<<32) + 7.75, wantSend: true, wantPoints: 7},
		{name: "positive-infinity", awarded: math.Inf(1), wantSend: true, wantPoints: 0},
		{name: "qword-overflow", awarded: 0x1p63, wantSend: true, wantPoints: 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := newSoloMonsterKillRewardWorld4EE500()
			configureSoloMonsterKillRewardHappy4EE500(w)
			w.awarded = tc.awarded
			soloMonsterKillReward4EE500("killed", w.hooks())
			var sends []string
			for _, event := range w.events {
				if len(event) >= len("send-line:") && event[:len("send-line:")] == "send-line:" {
					sends = append(sends, event)
				}
			}
			if !tc.wantSend {
				if len(sends) != 0 {
					t.Fatalf("send events = %v, want none", sends)
				}
				return
			}
			want := fmt.Sprintf("send-line:player:localized-gainpoints:%08x", tc.wantPoints)
			if !reflect.DeepEqual(sends, []string{want}) {
				t.Fatalf("send events = %v, want [%s]", sends, want)
			}
		})
	}
}

func TestSoloMonsterKillRewardPoints4EE500X87QwordEdges(t *testing.T) {
	belowPositiveLimit := math.Nextafter(0x1p63, 0)
	belowNegativeLimit := math.Nextafter(-0x1p63, math.Inf(-1))
	for _, tc := range []struct {
		name  string
		value float64
		want  uint32
	}{
		{name: "positive-fraction", value: 9.99, want: 9},
		{name: "negative-fraction", value: -9.99, want: 0xfffffff7},
		{name: "positive-low-word", value: float64(uint64(1)<<32) + 3, want: 3},
		{name: "negative-limit", value: -0x1p63, want: 0},
		{name: "last-positive-qword", value: belowPositiveLimit, want: uint32(int64(belowPositiveLimit))},
		{name: "positive-overflow", value: 0x1p63, want: 0},
		{name: "negative-overflow", value: belowNegativeLimit, want: 0},
		{name: "positive-infinity", value: math.Inf(1), want: 0},
		{name: "negative-infinity", value: math.Inf(-1), want: 0},
		{name: "nan", value: math.NaN(), want: 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := soloMonsterKillRewardPoints4EE500(tc.value); got != tc.want {
				t.Fatalf("points(%016x) = %08x, want %08x", math.Float64bits(tc.value), got, tc.want)
			}
		})
	}
}

func TestSoloMonsterKillReward4EE500AllHappyPathFaultPrefixes(t *testing.T) {
	baseline := newSoloMonsterKillRewardWorld4EE500()
	configureSoloMonsterKillRewardHappy4EE500(baseline)
	soloMonsterKillReward4EE500("killed", baseline.hooks())

	for faultAt := 1; faultAt <= len(baseline.events); faultAt++ {
		t.Run(fmt.Sprintf("fault-%02d", faultAt), func(t *testing.T) {
			w := newSoloMonsterKillRewardWorld4EE500()
			configureSoloMonsterKillRewardHappy4EE500(w)
			w.faultAt = faultAt
			defer func() {
				if got, want := recover(), baseline.events[faultAt-1]; got != want {
					t.Fatalf("panic = %v, want %q", got, want)
				}
				if want := baseline.events[:faultAt]; !reflect.DeepEqual(w.events, want) {
					t.Fatalf("events = %v, want %v", w.events, want)
				}
			}()
			soloMonsterKillReward4EE500("killed", w.hooks())
		})
	}
}
