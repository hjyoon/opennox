package server

import (
	"fmt"
	"reflect"
	"testing"
)

const (
	durationRayRecordParent4FEF90 = uint64(0x1000000aa)
	durationRayRecordA4FEF90      = uint64(0x2000000aa)
	durationRayRecordB4FEF90      = uint64(0x3000000aa)
	durationRayRecordC4FEF90      = uint64(0x4000000aa)

	durationRayCaster4FEF90  = uint64(0x10089abcdef)
	durationRayCaster24FEF90 = uint64(0x20089abcdef)
	durationRayCaster34FEF90 = uint64(0x30089abcdef)
	durationRayCaster44FEF90 = uint64(0x40089abcdef)
	durationRayWho4FEF90     = uint64(0x50089abcdef)
	durationRayTarget4FEF90  = uint64(0x60089abcdef)
	durationRayTargetA4FEF90 = uint64(0x70089abcdef)
	durationRayTargetB4FEF90 = uint64(0x80089abcdef)
	durationRayTargetC4FEF90 = uint64(0x90089abcdef)
)

type durationRayRecordState4FEF90 struct {
	spell  uint32
	level  uint32
	caster uint64
	target uint64
	sub108 uint64
	next   uint64
}

type durationRayObjectState4FEF90 struct {
	direction uint32
	code      uint32
}

type durationRayPacket4FEF90 struct {
	recipient int32
	packet    [7]byte
	related   uint64
	remove    int32
}

type durationRayWorld4FEF90 struct {
	events     []string
	after      map[string]func()
	faultAt    int
	sendResult int32
	records    map[uint64]*durationRayRecordState4FEF90
	objects    map[uint64]*durationRayObjectState4FEF90
	packets    []durationRayPacket4FEF90
	unmarks    []string
}

func durationRayRecordName4FEF90(record uint64) string {
	switch record {
	case 0:
		return "nil"
	case durationRayRecordParent4FEF90:
		return "P"
	case durationRayRecordA4FEF90:
		return "A"
	case durationRayRecordB4FEF90:
		return "B"
	case durationRayRecordC4FEF90:
		return "C"
	default:
		return fmt.Sprintf("%x", record)
	}
}

func durationRayObjectName4FEF90(object uint64) string {
	switch object {
	case 0:
		return "nil"
	case durationRayCaster4FEF90:
		return "caster"
	case durationRayCaster24FEF90:
		return "caster2"
	case durationRayCaster34FEF90:
		return "caster3"
	case durationRayCaster44FEF90:
		return "caster4"
	case durationRayWho4FEF90:
		return "who"
	case durationRayTarget4FEF90:
		return "target"
	case durationRayTargetA4FEF90:
		return "targetA"
	case durationRayTargetB4FEF90:
		return "targetB"
	case durationRayTargetC4FEF90:
		return "targetC"
	default:
		return fmt.Sprintf("%x", object)
	}
}

func (w *durationRayWorld4FEF90) observe(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
	if after := w.after[event]; after != nil {
		delete(w.after, event)
		after()
	}
}

func (w *durationRayWorld4FEF90) hooks() DurationRayStopHooks4FEF90[uint64, uint64] {
	return DurationRayStopHooks4FEF90[uint64, uint64]{
		LoadCaster: func(record uint64) uint64 {
			value := w.records[record].caster
			w.observe(fmt.Sprintf("caster:%s=%s", durationRayRecordName4FEF90(record), durationRayObjectName4FEF90(value)))
			return value
		},
		LoadSpell: func(record uint64) uint32 {
			value := w.records[record].spell
			w.observe(fmt.Sprintf("spell:%s=%d", durationRayRecordName4FEF90(record), value))
			return value
		},
		LoadLevelLow: func(record uint64) byte {
			value := byte(w.records[record].level)
			w.observe(fmt.Sprintf("level:%s=%02x", durationRayRecordName4FEF90(record), value))
			return value
		},
		LoadDirectionLow: func(object uint64) byte {
			value := byte(w.objects[object].direction)
			w.observe(fmt.Sprintf("direction:%s=%02x", durationRayObjectName4FEF90(object), value))
			return value
		},
		LoadTarget: func(record uint64) uint64 {
			value := w.records[record].target
			w.observe(fmt.Sprintf("target:%s=%s", durationRayRecordName4FEF90(record), durationRayObjectName4FEF90(value)))
			return value
		},
		LoadSub108: func(record uint64) uint64 {
			value := w.records[record].sub108
			w.observe(fmt.Sprintf("sub108:%s=%s", durationRayRecordName4FEF90(record), durationRayRecordName4FEF90(value)))
			return value
		},
		LoadNext: func(record uint64) uint64 {
			value := w.records[record].next
			w.observe(fmt.Sprintf("next:%s=%s", durationRayRecordName4FEF90(record), durationRayRecordName4FEF90(value)))
			return value
		},
		UnitCode: func(object uint64) uint32 {
			value := w.objects[object].code
			w.observe(fmt.Sprintf("code:%s=%08x", durationRayObjectName4FEF90(object), value))
			return value
		},
		SendPacket: func(recipient int32, packet [7]byte, related uint64, remove int32) int32 {
			w.packets = append(w.packets, durationRayPacket4FEF90{recipient, packet, related, remove})
			w.observe(fmt.Sprintf("send:%d:%x:%s:%d", recipient, packet, durationRayObjectName4FEF90(related), remove))
			return w.sendResult
		},
		UnmarkMinimap: func(object uint64, flags uint32) {
			event := fmt.Sprintf("unmark:%s:%d", durationRayObjectName4FEF90(object), flags)
			w.unmarks = append(w.unmarks, event)
			w.observe(event)
		},
	}
}

func newDurationRayWorld4FEF90(spell uint32) *durationRayWorld4FEF90 {
	return &durationRayWorld4FEF90{
		after: make(map[string]func()),
		records: map[uint64]*durationRayRecordState4FEF90{
			durationRayRecordA4FEF90: {
				spell:  spell,
				level:  0x123456ab,
				caster: durationRayCaster4FEF90,
				target: durationRayTarget4FEF90,
			},
		},
		objects: map[uint64]*durationRayObjectState4FEF90{
			durationRayCaster4FEF90:  {direction: 0x123456cd, code: 0xabcdef01},
			durationRayCaster24FEF90: {direction: 0x22334455, code: 0x11112222},
			durationRayCaster34FEF90: {direction: 0x33445566, code: 0x33334444},
			durationRayCaster44FEF90: {direction: 0x44556677, code: 0x55556666},
			durationRayWho4FEF90:     {code: 0x12345678},
			durationRayTarget4FEF90:  {code: 0x87654321},
			durationRayTargetA4FEF90: {code: 0xaaaa1001},
			durationRayTargetB4FEF90: {code: 0xbbbb2002},
			durationRayTargetC4FEF90: {code: 0xcccc3003},
		},
	}
}

func TestDurationRayStop4FEF90DispatchAndPacket(t *testing.T) {
	tests := []struct {
		spell uint32
		typ   byte
		value byte
	}{
		{7, 10, 0xab},
		{9, 9, 0xab},
		{22, 12, 0xab},
		{24, 11, 0xab},
		{35, 13, 0xab},
		{59, 8, 0xcd},
	}
	for _, tc := range tests {
		t.Run(fmt.Sprintf("spell-%d", tc.spell), func(t *testing.T) {
			w := newDurationRayWorld4FEF90(tc.spell)
			w.sendResult = -123456789
			DurationRayStop4FEF90(durationRayRecordA4FEF90, durationRayWho4FEF90, w.hooks())
			wantPacket := durationRayPacket4FEF90{
				recipient: 255,
				packet:    [7]byte{0x9e, tc.typ, tc.value, 0x78, 0x56, 0x01, 0xef},
				remove:    1,
			}
			if !reflect.DeepEqual(w.packets, []durationRayPacket4FEF90{wantPacket}) {
				t.Fatalf("packets = %#v, want %#v", w.packets, []durationRayPacket4FEF90{wantPacket})
			}
			wantUnmarks := []string{"unmark:caster:2", "unmark:who:2"}
			if !reflect.DeepEqual(w.unmarks, wantUnmarks) {
				t.Fatalf("unmarks = %v, want %v", w.unmarks, wantUnmarks)
			}
		})
	}
}

func TestDurationRayStop4FEF90RejectsNilAndDefaultInExactOrder(t *testing.T) {
	t.Run("nil-record", func(t *testing.T) {
		w := newDurationRayWorld4FEF90(24)
		DurationRayStop4FEF90(uint64(0), durationRayWho4FEF90, w.hooks())
		if len(w.events) != 0 {
			t.Fatalf("events = %v, want none", w.events)
		}
	})
	t.Run("nil-caster", func(t *testing.T) {
		w := newDurationRayWorld4FEF90(24)
		w.records[durationRayRecordA4FEF90].caster = 0
		DurationRayStop4FEF90(durationRayRecordA4FEF90, durationRayWho4FEF90, w.hooks())
		want := []string{"caster:A=nil"}
		if !reflect.DeepEqual(w.events, want) {
			t.Fatalf("events = %v, want %v", w.events, want)
		}
	})
	t.Run("nil-who", func(t *testing.T) {
		w := newDurationRayWorld4FEF90(24)
		DurationRayStop4FEF90(durationRayRecordA4FEF90, uint64(0), w.hooks())
		want := []string{"caster:A=caster"}
		if !reflect.DeepEqual(w.events, want) {
			t.Fatalf("events = %v, want %v", w.events, want)
		}
	})

	for spellID := uint32(0); spellID <= 70; spellID++ {
		if spellID == 7 || spellID == 9 || spellID == 22 || spellID == 24 ||
			spellID == 35 || spellID == 43 || spellID == 59 {
			continue
		}
		w := newDurationRayWorld4FEF90(spellID)
		DurationRayStop4FEF90(durationRayRecordA4FEF90, durationRayWho4FEF90, w.hooks())
		want := []string{"caster:A=caster", fmt.Sprintf("spell:A=%d", spellID)}
		if !reflect.DeepEqual(w.events, want) {
			t.Fatalf("spell %d events = %v, want %v", spellID, w.events, want)
		}
	}
}

func TestDurationRayStop4FEF90CachesPlasmaCasterAndReloadsForCallbacks(t *testing.T) {
	w := newDurationRayWorld4FEF90(59)
	w.after["caster:A=caster"] = func() {
		w.records[durationRayRecordA4FEF90].caster = durationRayCaster24FEF90
	}
	w.after["code:who=12345678"] = func() {
		w.records[durationRayRecordA4FEF90].caster = durationRayCaster34FEF90
	}
	w.after["send:255:9e08cd78564444:nil:1"] = func() {
		w.records[durationRayRecordA4FEF90].caster = durationRayCaster44FEF90
	}
	DurationRayStop4FEF90(durationRayRecordA4FEF90, durationRayWho4FEF90, w.hooks())
	want := []string{
		"caster:A=caster",
		"spell:A=59",
		"direction:caster=cd",
		"code:who=12345678",
		"caster:A=caster3",
		"code:caster3=33334444",
		"send:255:9e08cd78564444:nil:1",
		"caster:A=caster4",
		"unmark:caster4:2",
		"unmark:who:2",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events =\n%v\nwant\n%v", w.events, want)
	}
}

func TestDurationRayStop4FEF90GreaterHealOrderAndFullIdentity(t *testing.T) {
	t.Run("same-full-identity", func(t *testing.T) {
		w := newDurationRayWorld4FEF90(35)
		w.records[durationRayRecordA4FEF90].target = durationRayCaster4FEF90
		DurationRayStop4FEF90(durationRayRecordA4FEF90, durationRayWho4FEF90, w.hooks())
		want := []string{"caster:A=caster", "spell:A=35", "target:A=caster"}
		if !reflect.DeepEqual(w.events, want) {
			t.Fatalf("events = %v, want %v", w.events, want)
		}
	})

	t.Run("same-low-dword-is-distinct", func(t *testing.T) {
		w := newDurationRayWorld4FEF90(35)
		w.records[durationRayRecordA4FEF90].target = durationRayCaster24FEF90
		w.after["code:who=12345678"] = func() {
			w.records[durationRayRecordA4FEF90].caster = durationRayCaster34FEF90
		}
		w.after["code:caster3=33334444"] = func() {
			w.records[durationRayRecordA4FEF90].level = 0xfeedface
		}
		DurationRayStop4FEF90(durationRayRecordA4FEF90, durationRayWho4FEF90, w.hooks())
		want := []string{
			"caster:A=caster",
			"spell:A=35",
			"target:A=caster2",
			"code:who=12345678",
			"caster:A=caster3",
			"code:caster3=33334444",
			"level:A=ce",
			"send:255:9e0dce78564444:nil:1",
			"caster:A=caster3",
			"unmark:caster3:2",
			"unmark:who:2",
		}
		if !reflect.DeepEqual(w.events, want) {
			t.Fatalf("events =\n%v\nwant\n%v", w.events, want)
		}
	})
}

func TestDurationRayStop4FEF90ChainLightningUsesLiveNextAfterRecursion(t *testing.T) {
	w := newDurationRayWorld4FEF90(43)
	w.records = map[uint64]*durationRayRecordState4FEF90{
		durationRayRecordParent4FEF90: {
			spell: 43, caster: durationRayCaster4FEF90, sub108: durationRayRecordA4FEF90,
		},
		durationRayRecordA4FEF90: {
			spell: 7, level: 1, caster: durationRayCaster24FEF90,
			target: durationRayTargetA4FEF90, next: durationRayRecordB4FEF90,
		},
		durationRayRecordB4FEF90: {
			spell: 9, level: 2, caster: durationRayCaster34FEF90,
			target: durationRayTargetB4FEF90,
		},
		durationRayRecordC4FEF90: {
			spell: 59, caster: durationRayCaster44FEF90,
			target: durationRayTargetC4FEF90,
		},
	}
	w.after["send:255:9e0a0101102222:nil:1"] = func() {
		w.records[durationRayRecordA4FEF90].next = durationRayRecordC4FEF90
	}
	DurationRayStop4FEF90(durationRayRecordParent4FEF90, durationRayWho4FEF90, w.hooks())

	wantPackets := []durationRayPacket4FEF90{
		{255, [7]byte{0x9e, 10, 1, 0x01, 0x10, 0x22, 0x22}, 0, 1},
		{255, [7]byte{0x9e, 8, 0x77, 0x03, 0x30, 0x66, 0x66}, 0, 1},
	}
	if !reflect.DeepEqual(w.packets, wantPackets) {
		t.Fatalf("packets = %#v, want %#v", w.packets, wantPackets)
	}
	for _, event := range w.events {
		if event == "target:B=targetB" || event == "caster:B=caster3" {
			t.Fatalf("saved Next used instead of live post-recursion link: %v", w.events)
		}
	}
	wantTail := []string{
		"unmark:caster2:2", "unmark:targetA:2", "next:A=C",
		"target:C=targetC", "caster:C=caster4", "spell:C=59",
	}
	if len(w.events) < len(wantTail) {
		t.Fatalf("events too short: %v", w.events)
	}
	start := 0
	for _, want := range wantTail {
		found := false
		for start < len(w.events) {
			if w.events[start] == want {
				found = true
				start++
				break
			}
			start++
		}
		if !found {
			t.Fatalf("event %q not found in order: %v", want, w.events)
		}
	}
}

func TestDurationRayStop4FEF90FaultPrefixes(t *testing.T) {
	tests := []struct {
		name   string
		spell  uint32
		before func(*durationRayWorld4FEF90)
		want   []string
	}{
		{
			name:  "ordinary",
			spell: 24,
			want: []string{
				"caster:A=caster", "spell:A=24", "level:A=ab",
				"code:who=12345678", "caster:A=caster", "code:caster=abcdef01",
				"send:255:9e0bab785601ef:nil:1", "caster:A=caster",
				"unmark:caster:2", "unmark:who:2",
			},
		},
		{
			name:  "greater-heal",
			spell: 35,
			want: []string{
				"caster:A=caster", "spell:A=35", "target:A=target",
				"code:who=12345678", "caster:A=caster", "code:caster=abcdef01",
				"level:A=ab", "send:255:9e0dab785601ef:nil:1", "caster:A=caster",
				"unmark:caster:2", "unmark:who:2",
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for faultAt := 1; faultAt <= len(tc.want); faultAt++ {
				w := newDurationRayWorld4FEF90(tc.spell)
				if tc.before != nil {
					tc.before(w)
				}
				w.faultAt = faultAt
				panicked := false
				func() {
					defer func() {
						panicked = recover() != nil
					}()
					DurationRayStop4FEF90(durationRayRecordA4FEF90, durationRayWho4FEF90, w.hooks())
				}()
				if !panicked {
					t.Fatalf("fault %d did not panic", faultAt)
				}
				if want := tc.want[:faultAt]; !reflect.DeepEqual(w.events, want) {
					t.Fatalf("fault %d events = %v, want %v", faultAt, w.events, want)
				}
			}
		})
	}
}
