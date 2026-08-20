package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type directExperienceTestPlayer4EF3A0 struct {
	name  string
	token uint32
}

type directExperienceTestUpdate4EF3A0 struct {
	name   string
	player *directExperienceTestPlayer4EF3A0
}

type directExperienceTestObject4EF3A0 struct {
	name       string
	experience float32
	update     *directExperienceTestUpdate4EF3A0
}

type directExperienceTestWorld4EF3A0 struct {
	unit  *directExperienceTestObject4EF3A0
	award float32

	events         []string
	after          map[string]func()
	awardLoads     int
	protectedToken uint32
	protectedAward float32
	linePoints     uint32
}

func directExperienceObjectName4EF3A0(unit *directExperienceTestObject4EF3A0) string {
	if unit == nil {
		return "nil"
	}
	return unit.name
}

func directExperienceUpdateName4EF3A0(update *directExperienceTestUpdate4EF3A0) string {
	if update == nil {
		return "nil"
	}
	return update.name
}

func directExperiencePlayerName4EF3A0(player *directExperienceTestPlayer4EF3A0) string {
	if player == nil {
		return "nil"
	}
	return player.name
}

func (w *directExperienceTestWorld4EF3A0) record(event string) {
	w.events = append(w.events, event)
	if after := w.after[event]; after != nil {
		delete(w.after, event)
		after()
	}
}

func (w *directExperienceTestWorld4EF3A0) hooks() directExperienceGrantHooks4EF3A0[
	*directExperienceTestObject4EF3A0,
	*directExperienceTestUpdate4EF3A0,
	*directExperienceTestPlayer4EF3A0,
	string,
] {
	return directExperienceGrantHooks4EF3A0[
		*directExperienceTestObject4EF3A0,
		*directExperienceTestUpdate4EF3A0,
		*directExperienceTestPlayer4EF3A0,
		string,
	]{
		loadAwardArg: func() float32 {
			w.awardLoads++
			value := w.award
			w.record(fmt.Sprintf("award:%d=%08x", w.awardLoads, math.Float32bits(value)))
			return value
		},
		loadUnitArg: func() *directExperienceTestObject4EF3A0 {
			unit := w.unit
			w.record("arg:" + directExperienceObjectName4EF3A0(unit))
			return unit
		},
		loadExperience: func(unit *directExperienceTestObject4EF3A0) float32 {
			name := directExperienceObjectName4EF3A0(unit)
			if unit == nil {
				event := "experience:" + name
				w.record(event)
				panic(event)
			}
			value := unit.experience
			w.record(fmt.Sprintf("experience:%s=%08x", name, math.Float32bits(value)))
			return value
		},
		loadUpdateData: func(unit *directExperienceTestObject4EF3A0) *directExperienceTestUpdate4EF3A0 {
			name := directExperienceObjectName4EF3A0(unit)
			if unit == nil {
				event := "update:" + name
				w.record(event)
				panic(event)
			}
			update := unit.update
			w.record("update:" + name + "=" + directExperienceUpdateName4EF3A0(update))
			return update
		},
		storeExperience: func(unit *directExperienceTestObject4EF3A0, value float32) {
			name := directExperienceObjectName4EF3A0(unit)
			event := fmt.Sprintf("store:%s=%08x", name, math.Float32bits(value))
			w.record(event)
			if unit == nil {
				panic(event)
			}
			unit.experience = value
		},
		loadPlayer: func(update *directExperienceTestUpdate4EF3A0) *directExperienceTestPlayer4EF3A0 {
			name := directExperienceUpdateName4EF3A0(update)
			if update == nil {
				event := "player:" + name
				w.record(event)
				panic(event)
			}
			player := update.player
			w.record("player:" + name + "=" + directExperiencePlayerName4EF3A0(player))
			return player
		},
		loadExperienceToken: func(player *directExperienceTestPlayer4EF3A0) uint32 {
			name := directExperiencePlayerName4EF3A0(player)
			if player == nil {
				event := "token:" + name
				w.record(event)
				panic(event)
			}
			token := player.token
			w.record(fmt.Sprintf("token:%s=%08x", name, token))
			return token
		},
		protectExperience: func(token uint32, award float32) {
			w.record(fmt.Sprintf("protect:%08x:%08x", token, math.Float32bits(award)))
			w.protectedToken, w.protectedAward = token, award
		},
		reportExperience: func(unit *directExperienceTestObject4EF3A0) {
			w.record("report:" + directExperienceObjectName4EF3A0(unit))
		},
		loadString: func(key, path string, line int) string {
			w.record(fmt.Sprintf("string:%s:%s:%d", key, path, line))
			return "gain-message"
		},
		sendLineMessage: func(unit *directExperienceTestObject4EF3A0, message string, points uint32) {
			w.linePoints = points
			w.record(fmt.Sprintf("line:%s:%s:%08x", directExperienceObjectName4EF3A0(unit), message, points))
		},
		syncLevel: func(unit *directExperienceTestObject4EF3A0) {
			w.record("sync:" + directExperienceObjectName4EF3A0(unit))
		},
	}
}

func newDirectExperienceTestWorld4EF3A0() *directExperienceTestWorld4EF3A0 {
	player := &directExperienceTestPlayer4EF3A0{name: "player", token: 0x12345678}
	update := &directExperienceTestUpdate4EF3A0{name: "update", player: player}
	return &directExperienceTestWorld4EF3A0{
		unit:  &directExperienceTestObject4EF3A0{name: "unit", experience: 10, update: update},
		award: 2.5,
		after: make(map[string]func()),
	}
}

func directExperienceMustPanic4EF3A0(t *testing.T, run func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Fatal("call did not panic")
		}
	}()
	run()
}

func TestDirectExperienceGrant4EF3A0OrderCachingAndReloads(t *testing.T) {
	w := newDirectExperienceTestWorld4EF3A0()
	replacementPlayer := &directExperienceTestPlayer4EF3A0{name: "replacement-player", token: 0xaabbccdd}
	replacementUpdate := &directExperienceTestUpdate4EF3A0{name: "replacement-update", player: replacementPlayer}
	lateUpdate := &directExperienceTestUpdate4EF3A0{name: "late-update"}

	w.after["award:1=40200000"] = func() { w.award = 3.5 }
	w.after["award:2=40600000"] = func() { w.unit.experience = 20 }
	w.after["experience:unit=41a00000"] = func() { w.unit.update = replacementUpdate }
	w.after["update:unit=replacement-update"] = func() { w.unit.update = lateUpdate }
	w.after["token:replacement-player=aabbccdd"] = func() { replacementPlayer.token = 9 }
	w.after["protect:aabbccdd:40600000"] = func() { w.unit.experience = 999 }
	w.after["report:unit"] = func() { w.award = 4.75 }

	directExperienceGrant4EF3A0(w.hooks())
	want := []string{
		"award:1=40200000",
		"arg:unit",
		"award:2=40600000",
		"experience:unit=41a00000",
		"update:unit=replacement-update",
		"store:unit=41b40000",
		"player:replacement-update=replacement-player",
		"token:replacement-player=aabbccdd",
		"protect:aabbccdd:40600000",
		"report:unit",
		"award:3=40980000",
		`string:health.c:gainpoints:C:\NoxPost\src\Server\GameMech\explevel.c:381`,
		"line:unit:gain-message:00000004",
		"sync:unit",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want %q", w.events, want)
	}
	if w.unit.update != lateUpdate || w.unit.experience != 999 {
		t.Fatalf("callback mutations lost: update=%p experience=%v", w.unit.update, w.unit.experience)
	}
	if w.protectedToken != 0xaabbccdd || math.Float32bits(w.protectedAward) != 0x40600000 || w.linePoints != 4 {
		t.Fatalf("protected/line = %08x/%08x/%08x", w.protectedToken, math.Float32bits(w.protectedAward), w.linePoints)
	}
}

func TestDirectExperienceGrant4EF3A0AlwaysReportsMessagesAndSyncs(t *testing.T) {
	tests := []struct {
		name       string
		award      float32
		wantPoints uint32
	}{
		{name: "positive zero", award: 0, wantPoints: 0},
		{name: "negative", award: -1.75, wantPoints: 0xffffffff},
		{name: "nan", award: math.Float32frombits(0x7fc12345), wantPoints: 0},
		{name: "positive infinity", award: float32(math.Inf(1)), wantPoints: 0},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := newDirectExperienceTestWorld4EF3A0()
			w.award = test.award
			directExperienceGrant4EF3A0(w.hooks())
			if w.awardLoads != 3 || w.linePoints != test.wantPoints {
				t.Fatalf("award loads/points = %d/%08x, want 3/%08x", w.awardLoads, w.linePoints, test.wantPoints)
			}
			wantTail := []string{
				fmt.Sprintf("line:unit:gain-message:%08x", test.wantPoints),
				"sync:unit",
			}
			if !reflect.DeepEqual(w.events[len(w.events)-2:], wantTail) {
				t.Fatalf("tail = %q, want %q", w.events[len(w.events)-2:], wantTail)
			}
		})
	}
}

func TestDirectExperienceGrant4EF3A0AdditionSpillsOnce(t *testing.T) {
	tests := []struct {
		name       string
		experience float32
		award      float32
		wantBits   uint32
	}{
		{name: "ordinary", experience: 10, award: 2.5, wantBits: 0x41480000},
		{name: "overflow", experience: math.MaxFloat32, award: math.MaxFloat32, wantBits: 0x7f800000},
		{name: "negative zeros", experience: math.Float32frombits(0x80000000), award: math.Float32frombits(0x80000000), wantBits: 0x80000000},
		{name: "cancellation", experience: -2.5, award: 2.5, wantBits: 0},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := newDirectExperienceTestWorld4EF3A0()
			w.unit.experience, w.award = test.experience, test.award
			directExperienceGrant4EF3A0(w.hooks())
			if got := math.Float32bits(w.unit.experience); got != test.wantBits {
				t.Fatalf("stored experience = %08x, want %08x", got, test.wantBits)
			}
		})
	}
}

func TestDirectExperienceTruncLow4EF3A0(t *testing.T) {
	tests := []struct {
		name  string
		value float64
		want  uint32
	}{
		{name: "positive", value: 3.9, want: 3},
		{name: "negative", value: -3.9, want: 0xfffffffd},
		{name: "positive low dword", value: 0x1_0000_0003, want: 3},
		{name: "negative low dword", value: -0x1_0000_0003, want: 0xfffffffd},
		{name: "largest float32 below signed max", value: float64(math.Float32frombits(0x5effffff)), want: 0},
		{name: "signed minimum", value: -0x1p63, want: 0},
		{name: "positive overflow", value: 0x1p63, want: 0},
		{name: "negative overflow", value: math.Nextafter(-0x1p63, math.Inf(-1)), want: 0},
		{name: "nan", value: math.NaN(), want: 0},
		{name: "positive infinity", value: math.Inf(1), want: 0},
		{name: "negative infinity", value: math.Inf(-1), want: 0},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := directExperienceTruncLow4EF3A0(test.value); got != test.want {
				t.Fatalf("result = %08x, want %08x", got, test.want)
			}
		})
	}
}

func TestDirectExperienceGrant4EF3A0FaultPrefixes(t *testing.T) {
	t.Run("nil unit faults after two award loads", func(t *testing.T) {
		w := newDirectExperienceTestWorld4EF3A0()
		w.unit = nil
		directExperienceMustPanic4EF3A0(t, func() { directExperienceGrant4EF3A0(w.hooks()) })
		want := []string{"award:1=40200000", "arg:nil", "award:2=40200000", "experience:nil"}
		if !reflect.DeepEqual(w.events, want) {
			t.Fatalf("events = %q, want %q", w.events, want)
		}
	})

	t.Run("nil update faults after experience store", func(t *testing.T) {
		w := newDirectExperienceTestWorld4EF3A0()
		w.unit.update = nil
		directExperienceMustPanic4EF3A0(t, func() { directExperienceGrant4EF3A0(w.hooks()) })
		if math.Float32bits(w.unit.experience) != 0x41480000 || w.events[len(w.events)-1] != "player:nil" {
			t.Fatalf("experience/events = %08x/%q", math.Float32bits(w.unit.experience), w.events)
		}
	})

	t.Run("nil player faults at token after store", func(t *testing.T) {
		w := newDirectExperienceTestWorld4EF3A0()
		w.unit.update.player = nil
		directExperienceMustPanic4EF3A0(t, func() { directExperienceGrant4EF3A0(w.hooks()) })
		if math.Float32bits(w.unit.experience) != 0x41480000 || w.events[len(w.events)-1] != "token:nil" {
			t.Fatalf("experience/events = %08x/%q", math.Float32bits(w.unit.experience), w.events)
		}
	})

	t.Run("protect fault prevents report", func(t *testing.T) {
		w := newDirectExperienceTestWorld4EF3A0()
		w.after["protect:12345678:40200000"] = func() { panic("protect") }
		directExperienceMustPanic4EF3A0(t, func() { directExperienceGrant4EF3A0(w.hooks()) })
		if w.awardLoads != 2 || w.events[len(w.events)-1] != "protect:12345678:40200000" {
			t.Fatalf("loads/events = %d/%q", w.awardLoads, w.events)
		}
	})

	t.Run("report fault prevents message reload", func(t *testing.T) {
		w := newDirectExperienceTestWorld4EF3A0()
		w.after["report:unit"] = func() { panic("report") }
		directExperienceMustPanic4EF3A0(t, func() { directExperienceGrant4EF3A0(w.hooks()) })
		if w.awardLoads != 2 || w.events[len(w.events)-1] != "report:unit" {
			t.Fatalf("loads/events = %d/%q", w.awardLoads, w.events)
		}
	})

	t.Run("line fault prevents level synchronization", func(t *testing.T) {
		w := newDirectExperienceTestWorld4EF3A0()
		w.after["line:unit:gain-message:00000002"] = func() { panic("line") }
		directExperienceMustPanic4EF3A0(t, func() { directExperienceGrant4EF3A0(w.hooks()) })
		if w.events[len(w.events)-1] != "line:unit:gain-message:00000002" {
			t.Fatalf("events = %q", w.events)
		}
	})
}
