package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type ankhTestObject4EBF40 struct {
	name     string
	classLow uint8
	data     *ankhTestData4EBF40
	update   *ankhTestUpdate4EBF40
	frame    uint32
}

type ankhTestData4EBF40 struct {
	records [ankhHistoryCount4EBF40]ankhTestRecord4EBF40
	next    uint8
}

type ankhTestRecord4EBF40 struct {
	name   string
	class  uint8
	serial string
	frame  uint32
}

type ankhTestUpdate4EBF40 struct {
	name       string
	player     *ankhTestPlayer4EBF40
	extraLives int32
}

type ankhTestPlayer4EBF40 struct {
	name   string
	class  uint8
	serial string
	ankhs  [ankhPlayerSlotCount4EBF40]*ankhTestObject4EBF40
}

type ankhTestState4EBF40 struct {
	events        []string
	fps           uint32
	frames        []uint32
	frame         uint32
	ticks         []uint64
	tick          uint64
	feedbackTicks uint64
	resetName     string
	resetSerial   uint8
	balance       float32
	created       *ankhTestObject4EBF40
	nextReads     []uint8
	onPickup      func()
	onAudio       func(uint32)
	onPointFX     func()
	onMessage     func(string)
}

func ankhObjectName4EBF40(obj *ankhTestObject4EBF40) string {
	if obj == nil {
		return "nil"
	}
	return obj.name
}

func (s *ankhTestState4EBF40) add(format string, args ...any) {
	s.events = append(s.events, fmt.Sprintf(format, args...))
}

func (s *ankhTestState4EBF40) hooks() ankhCollideHooks4EBF40[
	*ankhTestObject4EBF40,
	*ankhTestData4EBF40,
	*ankhTestUpdate4EBF40,
	*ankhTestPlayer4EBF40,
] {
	return ankhCollideHooks4EBF40[
		*ankhTestObject4EBF40,
		*ankhTestData4EBF40,
		*ankhTestUpdate4EBF40,
		*ankhTestPlayer4EBF40,
	]{
		loadSourceInitData: func(obj *ankhTestObject4EBF40) *ankhTestData4EBF40 {
			s.add("data:%s", ankhObjectName4EBF40(obj))
			if obj == nil {
				panic("nil source")
			}
			return obj.data
		},
		loadTargetClassLow: func(obj *ankhTestObject4EBF40) uint8 {
			s.add("class:%s=%#x", ankhObjectName4EBF40(obj), obj.classLow)
			return obj.classLow
		},
		loadTargetUpdate: func(obj *ankhTestObject4EBF40) *ankhTestUpdate4EBF40 {
			s.add("update:%s", ankhObjectName4EBF40(obj))
			if obj.update == nil {
				panic("nil update")
			}
			return obj.update
		},
		loadPlayer: func(update *ankhTestUpdate4EBF40) *ankhTestPlayer4EBF40 {
			s.add("player:%s", update.name)
			if update.player == nil {
				panic("nil player")
			}
			return update.player
		},
		loadQuestAnkh: func(player *ankhTestPlayer4EBF40, index int) *ankhTestObject4EBF40 {
			obj := player.ankhs[index]
			s.add("slot:%s:%d=%s", player.name, index, ankhObjectName4EBF40(obj))
			return obj
		},
		storeQuestAnkh: func(player *ankhTestPlayer4EBF40, index int, obj *ankhTestObject4EBF40) {
			s.add("store-slot:%s:%d=%s", player.name, index, ankhObjectName4EBF40(obj))
			player.ankhs[index] = obj
		},
		loadFPS: func() uint32 {
			s.add("fps:%d", s.fps)
			return s.fps
		},
		loadFrame: func() uint32 {
			value := s.frame
			if len(s.frames) != 0 {
				value = s.frames[0]
				s.frames = s.frames[1:]
			}
			s.add("frame:%d", value)
			return value
		},
		loadRecordFrame: func(data *ankhTestData4EBF40, index int) uint32 {
			value := data.records[index].frame
			s.add("record-frame:%d=%d", index, value)
			return value
		},
		loadResetName: func() string {
			s.add("reset-name:%q", s.resetName)
			return s.resetName
		},
		storeRecordName: func(data *ankhTestData4EBF40, index int, value string) {
			s.add("store-name:%d=%q", index, value)
			data.records[index].name = value
		},
		loadResetSerialFirst: func() uint8 {
			s.add("reset-serial:%d", s.resetSerial)
			return s.resetSerial
		},
		storeRecordSerialFirst: func(data *ankhTestData4EBF40, index int, value uint8) {
			s.add("store-serial-first:%d=%d", index, value)
			serial := data.records[index].serial
			if value == 0 {
				data.records[index].serial = ""
			} else if serial == "" {
				data.records[index].serial = string([]byte{value})
			} else {
				data.records[index].serial = string([]byte{value}) + serial[1:]
			}
		},
		storeRecordClass: func(data *ankhTestData4EBF40, index int, value uint8) {
			s.add("store-class:%d=%d", index, value)
			data.records[index].class = value
		},
		storeRecordFrame: func(data *ankhTestData4EBF40, index int, value uint32) {
			s.add("store-record-frame:%d=%d", index, value)
			data.records[index].frame = value
		},
		loadRecordClass: func(data *ankhTestData4EBF40, index int) uint8 {
			value := data.records[index].class
			s.add("record-class:%d=%d", index, value)
			return value
		},
		loadPlayerClass: func(player *ankhTestPlayer4EBF40) uint8 {
			s.add("player-class:%s=%d", player.name, player.class)
			return player.class
		},
		loadRecordName: func(data *ankhTestData4EBF40, index int) string {
			value := data.records[index].name
			s.add("record-name:%d=%q", index, value)
			return value
		},
		loadPlayerName: func(player *ankhTestPlayer4EBF40) string {
			s.add("player-name:%s=%q", player.name, player.name)
			return player.name
		},
		loadRecordSerial: func(data *ankhTestData4EBF40, index int) string {
			value := data.records[index].serial
			s.add("record-serial:%d=%q", index, value)
			return value
		},
		loadPlayerSerial: func(player *ankhTestPlayer4EBF40) string {
			s.add("player-serial:%s=%q", player.name, player.serial)
			return player.serial
		},
		storeRecordSerial: func(data *ankhTestData4EBF40, index int, value string) {
			s.add("store-serial:%d=%q", index, value)
			data.records[index].serial = value
		},
		ticks: func() uint64 {
			value := s.tick
			if len(s.ticks) != 0 {
				value = s.ticks[0]
				s.ticks = s.ticks[1:]
			}
			s.add("ticks:%d", value)
			return value
		},
		loadFeedbackTicks: func() uint64 {
			s.add("feedback:%d", s.feedbackTicks)
			return s.feedbackTicks
		},
		priorityMessage: func(obj *ankhTestObject4EBF40, message string, value int32) {
			s.add("message:%s:%s:%d", ankhObjectName4EBF40(obj), message, value)
			if s.onMessage != nil {
				s.onMessage(message)
			}
		},
		audio: func(id uint32, obj *ankhTestObject4EBF40, first, second int32) {
			s.add("audio:%d:%s:%d:%d", id, ankhObjectName4EBF40(obj), first, second)
			if s.onAudio != nil {
				s.onAudio(id)
			}
		},
		storeFeedbackTicks: func(value uint64) {
			s.add("store-feedback:%d", value)
			s.feedbackTicks = value
		},
		loadBalance: func(key string) float32 {
			s.add("balance:%s=%g", key, s.balance)
			return s.balance
		},
		floatToInt: func(value float32) int32 {
			result := ankhRoundFloat32ToInt32_4EBF40(value)
			s.add("float-to-int:%g=%d", value, result)
			return result
		},
		loadExtraLives: func(update *ankhTestUpdate4EBF40) int32 {
			s.add("extra:%s=%d", update.name, update.extraLives)
			return update.extraLives
		},
		newObject: func(name string) *ankhTestObject4EBF40 {
			s.add("new:%s=%s", name, ankhObjectName4EBF40(s.created))
			return s.created
		},
		callPickup: func(who, item *ankhTestObject4EBF40, first int32, second uint32) {
			s.add("pickup:%s:%s:%d:%d", ankhObjectName4EBF40(who), ankhObjectName4EBF40(item), first, second)
			if s.onPickup != nil {
				s.onPickup()
			}
		},
		storeSourceFrame: func(obj *ankhTestObject4EBF40, frame uint32) {
			s.add("store-source-frame:%s=%d", ankhObjectName4EBF40(obj), frame)
			obj.frame = frame
		},
		pointFX: func(id uint32, obj *ankhTestObject4EBF40) uint32 {
			s.add("fx:%d:%s", id, ankhObjectName4EBF40(obj))
			if s.onPointFX != nil {
				s.onPointFX()
			}
			return 0xf1234567
		},
		loadNextIndex: func(data *ankhTestData4EBF40) uint8 {
			value := data.next
			if len(s.nextReads) != 0 {
				value = s.nextReads[0]
				s.nextReads = s.nextReads[1:]
			}
			s.add("next=%d", value)
			return value
		},
		storeNextIndex: func(data *ankhTestData4EBF40, value uint8) {
			s.add("store-next=%d", value)
			data.next = value
		},
	}
}

func TestAnkhCollide4EBF40CachesSourceDataBeforeTargetGuards(t *testing.T) {
	source := &ankhTestObject4EBF40{name: "source", data: &ankhTestData4EBF40{}}
	target := &ankhTestObject4EBF40{name: "target", classLow: 0x80}

	t.Run("nil target", func(t *testing.T) {
		state := &ankhTestState4EBF40{}
		ankhCollide4EBF40(source, (*ankhTestObject4EBF40)(nil), struct{ unread uint32 }{0xdeadbeef}, state.hooks())
		if want := []string{"data:source"}; !reflect.DeepEqual(state.events, want) {
			t.Fatalf("events = %#v, want %#v", state.events, want)
		}
	})

	t.Run("non player", func(t *testing.T) {
		state := &ankhTestState4EBF40{}
		ankhCollide4EBF40(source, target, 7, state.hooks())
		if want := []string{"data:source", "class:target=0x80"}; !reflect.DeepEqual(state.events, want) {
			t.Fatalf("events = %#v, want %#v", state.events, want)
		}
	})

	t.Run("nil source faults first", func(t *testing.T) {
		state := &ankhTestState4EBF40{}
		defer func() {
			if recover() == nil {
				t.Fatal("nil source did not fault")
			}
			if want := []string{"data:nil"}; !reflect.DeepEqual(state.events, want) {
				t.Fatalf("events = %#v, want %#v", state.events, want)
			}
		}()
		ankhCollide4EBF40((*ankhTestObject4EBF40)(nil), target, 0, state.hooks())
	})
}

func TestAnkhCollide4EBF40ExactHistoryDuplicatesSlotAndThrottles(t *testing.T) {
	data := &ankhTestData4EBF40{}
	data.records[0] = ankhTestRecord4EBF40{name: "hero", class: 2, serial: "serial", frame: 520}
	source := &ankhTestObject4EBF40{name: "source", data: data}
	player := &ankhTestPlayer4EBF40{name: "hero", class: 2, serial: "serial"}
	player.ankhs[0] = source
	update := &ankhTestUpdate4EBF40{name: "entry-update", player: player}
	target := &ankhTestObject4EBF40{name: "target", classLow: 0x84, update: update}
	state := &ankhTestState4EBF40{
		fps: 2, frame: 1000, ticks: []uint64{2501, 9000}, feedbackTicks: 1000,
	}

	ankhCollide4EBF40(source, target, &struct{ unread uint32 }{0xcafebabe}, state.hooks())

	if player.ankhs[0] != source || player.ankhs[1] != source {
		t.Fatalf("exact history did not preserve the original duplicate insertion: %#v", player.ankhs)
	}
	if state.feedbackTicks != 9000 {
		t.Fatalf("feedback ticks = %d, want 9000", state.feedbackTicks)
	}
	wantTail := []string{
		"ticks:2501", "feedback:1000",
		"message:target:objcoll.c:ExtraLifeAlreadyAwarded:0",
		"audio:925:target:0:0", "ticks:9000", "store-feedback:9000",
	}
	if got := state.events[len(state.events)-len(wantTail):]; !reflect.DeepEqual(got, wantTail) {
		t.Fatalf("feedback tail = %#v, want %#v", got, wantTail)
	}
	var initialSlot0, insertedSlot0, insertedSlot1 int
	for _, event := range state.events {
		switch event {
		case "slot:hero:0=source":
			if initialSlot0 == 0 {
				initialSlot0++
			} else {
				insertedSlot0++
			}
		case "slot:hero:1=nil":
			insertedSlot1++
		}
	}
	if initialSlot0 != 1 || insertedSlot0 != 1 || insertedSlot1 != 1 {
		t.Fatalf("slot scan counts = initial %d, insert slot0 %d, insert slot1 %d", initialSlot0, insertedSlot0, insertedSlot1)
	}
}

func TestAnkhCollide4EBF40HistoryExpirationIsUnsignedStrictAndOrdered(t *testing.T) {
	tests := []struct {
		name        string
		frame       uint32
		recordFrame uint32
		wantReset   bool
	}{
		{name: "exact boundary", frame: 1000, recordFrame: 520},
		{name: "one over", frame: 1000, recordFrame: 519, wantReset: true},
		{name: "unsigned wrap", frame: 100, recordFrame: math.MaxUint32 - 380, wantReset: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			data := &ankhTestData4EBF40{}
			data.records[0] = ankhTestRecord4EBF40{name: "reset", serial: "", frame: tc.recordFrame}
			source := &ankhTestObject4EBF40{name: "source", data: data}
			player := &ankhTestPlayer4EBF40{name: "reset"}
			target := &ankhTestObject4EBF40{
				name: "target", classLow: ankhPlayerClassLow4EBF40,
				update: &ankhTestUpdate4EBF40{name: "update", player: player},
			}
			state := &ankhTestState4EBF40{fps: 2, frame: tc.frame, resetName: "reset"}
			ankhCollide4EBF40(source, target, (*struct{})(nil), state.hooks())

			resetStart := -1
			for i, event := range state.events {
				if event == "reset-name:\"reset\"" {
					resetStart = i
					break
				}
			}
			if (resetStart >= 0) != tc.wantReset {
				t.Fatalf("reset present = %v, want %v; events %#v", resetStart >= 0, tc.wantReset, state.events)
			}
			if tc.wantReset {
				want := []string{
					"reset-name:\"reset\"", "store-name:0=\"reset\"",
					"reset-serial:0", "store-serial-first:0=0",
					"store-class:0=0", "store-record-frame:0=0",
				}
				if got := state.events[resetStart : resetStart+len(want)]; !reflect.DeepEqual(got, want) {
					t.Fatalf("reset order = %#v, want %#v", got, want)
				}
			}
			wantPrefix := []string{
				"data:source", "class:target=0x4", "update:target", "player:update",
				"slot:reset:0=nil", "slot:reset:1=nil", "slot:reset:2=nil", "slot:reset:3=nil", "slot:reset:4=nil",
				"fps:2", fmt.Sprintf("frame:%d", tc.frame), fmt.Sprintf("record-frame:0=%d", tc.recordFrame),
			}
			if got := state.events[:len(wantPrefix)]; !reflect.DeepEqual(got, wantPrefix) {
				t.Fatalf("prefix = %#v, want %#v", got, wantPrefix)
			}
		})
	}
}

func TestAnkhCollide4EBF40StoredSlotScansAllHistoryThenUsesFeedbackBoundary(t *testing.T) {
	tests := []struct {
		name       string
		now        uint64
		last       uint64
		wantNotify bool
	}{
		{name: "exact boundary", now: 2500, last: 1000},
		{name: "one over", now: 2501, last: 1000, wantNotify: true},
		{name: "unsigned wrap", now: 100, last: math.MaxUint64 - 1400, wantNotify: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			data := &ankhTestData4EBF40{}
			source := &ankhTestObject4EBF40{name: "source", data: data}
			player := &ankhTestPlayer4EBF40{name: "hero", class: 7, serial: "serial"}
			player.ankhs[0] = source
			target := &ankhTestObject4EBF40{
				name: "target", classLow: ankhPlayerClassLow4EBF40,
				update: &ankhTestUpdate4EBF40{name: "update", player: player},
			}
			state := &ankhTestState4EBF40{
				fps: 1, ticks: []uint64{tc.now, 9000}, feedbackTicks: tc.last,
			}
			ankhCollide4EBF40(source, target, (*struct{})(nil), state.hooks())

			recordFrameReads := 0
			messages := 0
			for _, event := range state.events {
				if len(event) >= len("record-frame:") && event[:len("record-frame:")] == "record-frame:" {
					recordFrameReads++
				}
				if len(event) >= len("message:") && event[:len("message:")] == "message:" {
					messages++
				}
			}
			if recordFrameReads != ankhHistoryCount4EBF40 {
				t.Fatalf("history frame reads = %d, want %d", recordFrameReads, ankhHistoryCount4EBF40)
			}
			if (messages == 1) != tc.wantNotify {
				t.Fatalf("message count = %d, wantNotify %v", messages, tc.wantNotify)
			}
			for _, event := range state.events {
				if len(event) >= len("balance:") && event[:len("balance:")] == "balance:" {
					t.Fatal("already-stored path unexpectedly loaded balance")
				}
			}
		})
	}
}

func TestAnkhCollide4EBF40AwardUsesCachedPointersAndLiveReloads(t *testing.T) {
	data := &ankhTestData4EBF40{}
	replacementData := &ankhTestData4EBF40{}
	source := &ankhTestObject4EBF40{name: "source", data: data}
	entryPlayer := &ankhTestPlayer4EBF40{name: "entry", class: 7, serial: "entry-serial"}
	livePlayer := &ankhTestPlayer4EBF40{name: "live", class: 9, serial: "live-serial"}
	occupied := &ankhTestObject4EBF40{name: "occupied"}
	livePlayer.ankhs[0] = occupied
	entryUpdate := &ankhTestUpdate4EBF40{name: "entry-update", player: entryPlayer, extraLives: 1}
	replacementUpdate := &ankhTestUpdate4EBF40{name: "replacement-update", player: &ankhTestPlayer4EBF40{name: "wrong"}}
	target := &ankhTestObject4EBF40{name: "target", classLow: 0x84, update: entryUpdate}
	created := &ankhTestObject4EBF40{name: "created"}
	frames := make([]uint32, ankhHistoryCount4EBF40, ankhHistoryCount4EBF40+2)
	frames = append(frames, 111, 222)
	state := &ankhTestState4EBF40{
		fps: 1, frames: frames, balance: 2.5, created: created,
		nextReads: []uint8{1, 2, 3, 4, 63},
	}
	state.onPickup = func() {
		source.data = replacementData
		target.update = replacementUpdate
		entryUpdate.player = livePlayer
	}

	ankhCollide4EBF40(source, target, struct{ unread [3]uint32 }{}, state.hooks())

	if source.frame != 111 {
		t.Fatalf("source frame = %d, want 111", source.frame)
	}
	if livePlayer.ankhs[0] != occupied || livePlayer.ankhs[1] != source {
		t.Fatalf("live player slots = %#v", livePlayer.ankhs)
	}
	if data.records[1].name != "live" || data.records[2].class != 9 ||
		data.records[3].serial != "live-serial" || data.records[4].frame != 222 {
		t.Fatalf("scattered live history writes = %#v", data.records[:5])
	}
	if data.next != 0 {
		t.Fatalf("wrapped next = %d, want 0", data.next)
	}
	if replacementData.next != 0 || replacementData.records[1].name != "" {
		t.Fatal("callback-replaced source InitData was used instead of entry cache")
	}

	sequence := []string{
		"new:AnkhTradable=created", "pickup:target:created:1:0",
		"frame:111", "store-source-frame:source=111",
		"audio:1004:source:0:0", "fx:130:source",
		"message:target:objcoll.c:AwardExtraLife:0",
	}
	start := -1
	for i, event := range state.events {
		if event == sequence[0] {
			start = i
			break
		}
	}
	if start < 0 {
		t.Fatalf("award sequence start not found in %#v", state.events)
	}
	if got := state.events[start : start+len(sequence)]; !reflect.DeepEqual(got, sequence) {
		t.Fatalf("award sequence around allocation = %#v, want %#v", got, sequence)
	}
	for _, event := range state.events {
		if event == "player:replacement-update" {
			t.Fatal("callback-replaced target UpdateData was loaded")
		}
	}
}

func TestAnkhCollide4EBF40NilCreatedObjectStillCompletesAward(t *testing.T) {
	data := &ankhTestData4EBF40{}
	source := &ankhTestObject4EBF40{name: "source", data: data}
	player := &ankhTestPlayer4EBF40{name: "hero", class: 3, serial: "serial"}
	target := &ankhTestObject4EBF40{
		name: "target", classLow: ankhPlayerClassLow4EBF40,
		update: &ankhTestUpdate4EBF40{name: "update", player: player},
	}
	state := &ankhTestState4EBF40{fps: 1, balance: 1, frames: make([]uint32, ankhHistoryCount4EBF40+2)}
	state.onPickup = func() { t.Fatal("nil created object invoked Pickup") }
	ankhCollide4EBF40(source, target, (*struct{})(nil), state.hooks())
	if player.ankhs[0] != source || data.records[0].name != "hero" || data.records[0].serial != "serial" {
		t.Fatalf("nil-allocation award did not finish: slots %#v, record %#v", player.ankhs, data.records[0])
	}
}

func TestAnkhCollide4EBF40MaximumUsesSignedCountAndMessage(t *testing.T) {
	data := &ankhTestData4EBF40{}
	source := &ankhTestObject4EBF40{name: "source", data: data}
	player := &ankhTestPlayer4EBF40{name: "hero", class: 3, serial: "serial"}
	target := &ankhTestObject4EBF40{
		name: "target", classLow: ankhPlayerClassLow4EBF40,
		update: &ankhTestUpdate4EBF40{name: "update", player: player, extraLives: -1},
	}
	state := &ankhTestState4EBF40{
		fps: 1, balance: -1.5, ticks: []uint64{2501, 9000}, feedbackTicks: 1000,
	}
	ankhCollide4EBF40(source, target, (*struct{})(nil), state.hooks())

	if source.frame != 0 || player.ankhs[0] != nil {
		t.Fatal("signed maximum branch unexpectedly awarded")
	}
	want := []string{
		"balance:MaxExtraLives=-1.5", "float-to-int:-1.5=-2", "extra:update=-1",
		"ticks:2501", "feedback:1000",
		"message:target:pickup.c:MaxTradableAnkhsReached:0",
		"audio:925:target:0:0", "ticks:9000", "store-feedback:9000",
	}
	start := -1
	for i, event := range state.events {
		if event == want[0] {
			start = i
			break
		}
	}
	if start < 0 {
		t.Fatalf("maximum sequence start not found in %#v", state.events)
	}
	if got := state.events[start : start+len(want)]; !reflect.DeepEqual(got, want) {
		t.Fatalf("maximum sequence = %#v, want %#v", got, want)
	}
}

func TestAnkhRoundFloat32ToInt324EBF40(t *testing.T) {
	tests := []struct {
		value float32
		want  int32
	}{
		{value: 1.5, want: 2},
		{value: 2.5, want: 2},
		{value: -1.5, want: -2},
		{value: -2.5, want: -2},
		{value: float32(math.NaN()), want: math.MinInt32},
		{value: 2147483648, want: math.MinInt32},
		{value: float32(math.Inf(-1)), want: math.MinInt32},
	}
	for _, tc := range tests {
		if got := ankhRoundFloat32ToInt32_4EBF40(tc.value); got != tc.want {
			t.Errorf("round(%v) = %d, want %d", tc.value, got, tc.want)
		}
	}
}
