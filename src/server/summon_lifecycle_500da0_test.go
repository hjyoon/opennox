package server

import (
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"
)

func TestSummonPayload500DA0PreservesPackedBytesOnNativeRecord(t *testing.T) {
	record := &DurSpell{
		Field76: ^uintptr(0),
		Field84: 0xabcd0000,
	}
	payload := summonPayload500DA0{
		typeID: 0xbeef,
		position: types.Pointf{
			X: math.Float32frombits(0x81234567),
			Y: math.Float32frombits(0x7f012345),
		},
		direction: 0x9a,
		effectID:  0xcdef,
		complete:  0x72,
	}
	storeSummonPayload500DA0(record, payload)

	if got := summonPayloadTypeID500DA0(record); got != payload.typeID {
		t.Fatalf("type ID = %#x, want %#x", got, payload.typeID)
	}
	gotPosition := summonPayloadPosition500DA0(record)
	if math.Float32bits(gotPosition.X) != math.Float32bits(payload.position.X) ||
		math.Float32bits(gotPosition.Y) != math.Float32bits(payload.position.Y) {
		t.Fatalf("position bits = %#x/%#x, want %#x/%#x",
			math.Float32bits(gotPosition.X), math.Float32bits(gotPosition.Y),
			math.Float32bits(payload.position.X), math.Float32bits(payload.position.Y))
	}
	if got := summonPayloadDirection500DA0(record); got != payload.direction {
		t.Fatalf("direction = %#x, want %#x", got, payload.direction)
	}
	if got := summonPayloadEffectID500DA0(record); got != payload.effectID {
		t.Fatalf("effect ID = %#x, want %#x", got, payload.effectID)
	}
	if got := summonPayloadComplete500DA0(record); got != payload.complete {
		t.Fatalf("complete = %#x, want %#x", got, payload.complete)
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && uint64(record.Field76)>>32 != 0 {
		t.Fatalf("Field76 retained stale native high bits: %#x", record.Field76)
	}
	if got := record.Field84 >> 16; got != 0xabcd {
		t.Fatalf("Field84 high half = %#x, want preserved 0xabcd", got)
	}

	before := record.Field84
	storeSummonPayloadComplete500DA0(record, 1)
	if got := summonPayloadComplete500DA0(record); got != 1 {
		t.Fatalf("stored complete = %d, want 1", got)
	}
	if got, want := record.Field84&^uint32(0xff00), before&^uint32(0xff00); got != want {
		t.Fatalf("complete store changed adjacent bits: got %#x, want %#x", got, want)
	}
}

func TestSummonStart500DA0SuccessOrderPackingAndWrap(t *testing.T) {
	const (
		record = uint64(0x100000101)
		caster = uint64(0x7f031b61e540)
	)
	var events []string
	observe := func(event string) {
		events = append(events, event)
	}
	var storedPayload summonPayload500DA0
	var effectStores []uint16
	var deadline uint32
	var sent struct {
		effectID  uint16
		position  types.Pointf
		direction byte
		typeID    uint16
		duration  uint16
	}
	hooks := summonStartHooks500DA0[uint64, uint64, uint64, uint64]{
		loadSpell: func(got uint64) uint32 {
			observe("spell")
			if got != record {
				t.Fatalf("spell record = %#x", got)
			}
			return 81
		},
		loadCaster: func(got uint64) uint64 {
			observe("caster")
			return caster
		},
		loadObjectFlags: func(got uint64) uint32 {
			observe("flags")
			return 0
		},
		loadRecordFlag: func(got uint64) uint32 {
			observe("record-flag")
			return 0
		},
		loadClassLowByte: func(got uint64) byte {
			observe("class")
			return 0
		},
		place: func(got uint64, position *types.Pointf) int32 {
			observe("place")
			*position = types.Ptf(-12.75, 400.5)
			return 1
		},
		loadDirectionLow: func(got uint64) byte {
			observe("direction")
			if got != caster {
				t.Fatalf("direction caster = %#x", got)
			}
			return 0xa5
		},
		guideName: func(guide int32) string {
			observe("guide-name")
			if guide != 7 {
				t.Fatalf("guide = %d, want 7", guide)
			}
			return "Spider"
		},
		typeID: func(name string) uint16 {
			observe("type-id")
			if name != "Spider" {
				t.Fatalf("type name = %q", name)
			}
			return 0x3456
		},
		loadEffectID: func() uint16 {
			observe("effect-load")
			return 0xfde7
		},
		storeEffectID: func(value uint16) {
			observe("effect-store")
			effectStores = append(effectStores, value)
		},
		storePayload: func(got uint64, payload summonPayload500DA0) {
			observe("payload")
			storedPayload = payload
		},
		guideSize: func(guide int32) int32 {
			observe("guide-size")
			return 0x12340004
		},
		balanceFloat: func(key string, selector int32) float64 {
			observe("balance")
			if key != summonDurationKey500DA0 || selector != 2 {
				t.Fatalf("balance args = %q/%d", key, selector)
			}
			return 19.875
		},
		floatToInt: func(value float32) int32 {
			observe("float-int")
			if value != float32(19.875) {
				t.Fatalf("float conversion input = %v", value)
			}
			return 19
		},
		recordDword: func(got uint64) uint32 {
			t.Fatal("supported guide size used residual record dword")
			return 0
		},
		frame: func() uint32 {
			observe("frame")
			return math.MaxUint32 - 10
		},
		storeDeadline: func(got uint64, value uint32) {
			observe("deadline")
			deadline = value
		},
		sendStartEffect: func(effectID uint16, position types.Pointf, direction byte, typeID, duration uint16) {
			observe("send")
			sent.effectID = effectID
			sent.position = position
			sent.direction = direction
			sent.typeID = typeID
			sent.duration = duration
		},
	}

	if got := summonStart500DA0(record, hooks); got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
	wantEvents := []string{
		"spell", "caster", "flags", "record-flag", "class", "place",
		"caster", "direction", "guide-name", "type-id", "effect-load",
		"effect-store", "effect-store", "payload", "guide-size", "balance",
		"float-int", "frame", "deadline", "send",
	}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %q, want exact oracle order %q", events, wantEvents)
	}
	if want := []uint16{0xfde8, 0}; !reflect.DeepEqual(effectStores, want) {
		t.Fatalf("effect stores = %#v, want %#v", effectStores, want)
	}
	if storedPayload.typeID != 0x3456 ||
		storedPayload.position != (types.Ptf(-12.75, 400.5)) ||
		storedPayload.direction != 0xa5 ||
		storedPayload.effectID != 0xfde7 ||
		storedPayload.complete != 0 {
		t.Fatalf("stored payload = %#v", storedPayload)
	}
	if deadline != 8 {
		t.Fatalf("deadline = %#x, want wrapped 8", deadline)
	}
	if sent.effectID != storedPayload.effectID ||
		sent.position != storedPayload.position ||
		sent.direction != storedPayload.direction ||
		sent.typeID != storedPayload.typeID ||
		sent.duration != 19 {
		t.Fatalf("sent effect = %#v", sent)
	}
}

func TestSummonStart500DA0PlayerGuideFailureUsesCachedUpdate(t *testing.T) {
	const (
		record = uint64(1)
		caster = uint64(0x7f0012345678)
		update = uint64(0x7f0023456789)
		player = uint64(0x7f003456789a)
	)
	var events []string
	hooks := summonStartHooks500DA0[uint64, uint64, uint64, uint64]{
		loadSpell: func(uint64) uint32 {
			events = append(events, "spell")
			return 74 + 11
		},
		loadCaster: func(uint64) uint64 {
			events = append(events, "caster")
			return caster
		},
		loadObjectFlags: func(uint64) uint32 {
			events = append(events, "flags")
			return 0
		},
		loadRecordFlag: func(uint64) uint32 {
			events = append(events, "record-flag")
			return 0
		},
		loadClassLowByte: func(uint64) byte {
			events = append(events, "class")
			return summonPlayerClass500DA0
		},
		loadUpdate: func(uint64) uint64 {
			events = append(events, "update")
			return update
		},
		gameFlags: func(mask uint32) int32 {
			events = append(events, "game-flags")
			return 1
		},
		loadPlayer: func(got uint64) uint64 {
			events = append(events, "player")
			if got != update {
				t.Fatalf("update = %#x", got)
			}
			return player
		},
		loadGuideLevel: func(got uint64, guide int32) uint32 {
			events = append(events, "guide-level")
			return 0
		},
		privateMessage: func(got uint64, message string, value byte) {
			events = append(events, "message")
			if got != caster || message != summonNeedGuideMessage500DA0 || value != 0 {
				t.Fatalf("message args = %#x/%q/%d", got, message, value)
			}
		},
		checkLimit: func(uint64, int32) int32 {
			t.Fatal("guide failure checked summon limit")
			return 0
		},
		place: func(uint64, *types.Pointf) int32 {
			t.Fatal("guide failure attempted placement")
			return 0
		},
	}
	if got := summonStart500DA0(record, hooks); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	want := []string{
		"spell", "caster", "flags", "record-flag", "class", "update",
		"game-flags", "player", "guide-level", "caster", "message",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %q, want %q", events, want)
	}
}

func TestSummonPlacement500F40PlayerClampAndExactTileResult(t *testing.T) {
	const (
		record = uint64(1)
		caster = uint64(0x7f0012345678)
	)
	var destination types.Pointf
	var tracedFrom, tracedTo types.Pointf
	hooks := summonPlacementHooks500F40[uint64, uint64, *types.Pointf, uint64, uint64]{
		loadCaster: func(uint64) uint64 { return caster },
		loadClassLowByte: func(uint64) byte {
			return summonPlayerClass500DA0
		},
		loadObjectX: func(uint64) float32 { return 10 },
		loadObjectY: func(uint64) float32 { return 20 },
		loadTargetX: func(uint64) float32 { return 110 },
		loadTargetY: func(uint64) float32 { return 20 },
		trace: func(from, to types.Pointf, flags MapTraceFlags) int32 {
			tracedFrom, tracedTo = from, to
			if flags != summonTraceFlags500F40 {
				t.Fatalf("trace flags = %#x", flags)
			}
			return 1
		},
		tileAllow: func(point *types.Pointf) int32 {
			if *point != (types.Ptf(60, 20)) {
				t.Fatalf("tile point = %v, want clamped 60/20", *point)
			}
			return 2
		},
		storeX: func(out *types.Pointf, value float32) { out.X = value },
		storeY: func(out *types.Pointf, value float32) { out.Y = value },
	}
	if got := summonPlacement500F40(record, &destination, hooks); got != 1 {
		t.Fatalf("result = %d, want exact-tile-non-one success", got)
	}
	if tracedFrom != (types.Ptf(10, 20)) || tracedTo != (types.Ptf(60, 20)) {
		t.Fatalf("trace = %v -> %v", tracedFrom, tracedTo)
	}
	if destination != (types.Ptf(60, 20)) {
		t.Fatalf("destination = %v", destination)
	}
}

func TestSummonFinishAndCancel5010D0(t *testing.T) {
	const (
		record   = uint64(0x100000101)
		caster   = uint64(0x7f0012345678)
		summoned = uint64(0x7f009abcdef0)
	)
	position := types.Ptf(12.5, -7.25)
	var events []string
	var completed byte
	hooks := summonFinishHooks5010D0[uint64, uint64, uint64, uint64]{
		loadCaster: func(uint64) uint64 {
			events = append(events, "caster")
			return caster
		},
		loadObjectFlags: func(uint64) uint32 {
			events = append(events, "flags")
			return 0
		},
		loadDeadline: func(uint64) uint32 {
			events = append(events, "deadline")
			return 100
		},
		frame: func() uint32 {
			events = append(events, "frame")
			return 99
		},
		loadSpell: func(uint64) uint32 {
			events = append(events, "spell")
			return 80
		},
		loadClassLowByte: func(uint64) byte {
			events = append(events, "class")
			return 0
		},
		loadPayloadDirection: func(uint64) byte {
			events = append(events, "direction")
			return 0x3c
		},
		loadPayloadTypeID: func(uint64) uint16 {
			events = append(events, "type")
			return 0x1234
		},
		loadPayloadPosition: func(uint64) types.Pointf {
			events = append(events, "position")
			return position
		},
		summon: func(typeID uint16, gotPosition types.Pointf, gotCaster uint64, direction byte) uint64 {
			events = append(events, "summon")
			if typeID != 0x1234 || gotPosition != position || gotCaster != caster || direction != 0x3c {
				t.Fatalf("summon args = %#x/%v/%#x/%#x", typeID, gotPosition, gotCaster, direction)
			}
			return summoned
		},
		audioObject: func(got uint64) {
			events = append(events, "audio")
			if got != summoned {
				t.Fatalf("audio object = %#x", got)
			}
		},
		storeComplete: func(got uint64, value byte) {
			events = append(events, "complete")
			completed = value
		},
	}
	if got := summonFinish5010D0(record, hooks); got != 1 {
		t.Fatalf("finish result = %d, want 1", got)
	}
	want := []string{
		"caster", "flags", "deadline", "frame", "spell", "class",
		"direction", "caster", "type", "position", "summon", "audio", "complete",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("finish events = %q, want %q", events, want)
	}
	if completed != 1 {
		t.Fatalf("complete = %d, want 1", completed)
	}

	events = nil
	cancelHooks := summonCancelHooks5011C0[uint64]{
		loadComplete: func(uint64) byte {
			events = append(events, "complete")
			return 0
		},
		loadEffectID: func(uint64) uint16 {
			events = append(events, "effect")
			return 0xabcd
		},
		sendCancelEffect: func(effect uint16) {
			events = append(events, "send")
			if effect != 0xabcd {
				t.Fatalf("cancel effect = %#x", effect)
			}
		},
		loadPayloadPosition: func(uint64) types.Pointf {
			events = append(events, "position")
			return position
		},
		audioPosition: func(got types.Pointf) {
			events = append(events, "audio")
			if got != position {
				t.Fatalf("cancel position = %v", got)
			}
		},
	}
	summonCancel5011C0(record, cancelHooks)
	if want := []string{"complete", "effect", "send", "position", "audio"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("cancel events = %q, want %q", events, want)
	}
}
