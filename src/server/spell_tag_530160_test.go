package server

import (
	"reflect"
	"testing"
)

type tagTestObject530160 struct {
	class  byte
	flags  uint32
	index  int32
	code   uint16
	typeID uint16
}

type tagTestRecord530160 struct {
	caster *tagTestObject530160
	target *tagTestObject530160
	level  uint32
	frame  uint32
}

func tagTestHooks530160(events *[]string, packet *[7]byte) spellTagHooks530160[*tagTestRecord530160, *tagTestObject530160] {
	return spellTagHooks530160[*tagTestRecord530160, *tagTestObject530160]{
		loadCaster:      func(r *tagTestRecord530160) *tagTestObject530160 { return r.caster },
		loadTarget:      func(r *tagTestRecord530160) *tagTestObject530160 { return r.target },
		loadClass:       func(o *tagTestObject530160) byte { return o.class },
		loadFlags:       func(o *tagTestObject530160) uint32 { return o.flags },
		loadPlayerIndex: func(o *tagTestObject530160) int32 { return o.index },
		loadLevel:       func(r *tagTestRecord530160) uint32 { return r.level },
		loadFrame:       func() uint32 { return 0xfffffff0 },
		storeFrame: func(r *tagTestRecord530160, frame uint32) {
			*events = append(*events, "store-frame")
			r.frame = frame
		},
		loadBalance: func(key string) float64 {
			if key != "TagDurationPerLevel" {
				panic(key)
			}
			*events = append(*events, "balance")
			return 2.5
		},
		floatToInt: func(v float32) int32 {
			if v != 2.5 {
				panic(v)
			}
			*events = append(*events, "round")
			return 2
		},
		mark: func(index int32, o *tagTestObject530160, flags uint32) {
			if index != 7 || o == nil || flags != 1 {
				panic("mark arguments")
			}
			*events = append(*events, "mark")
		},
		unmark: func(index int32, o *tagTestObject530160, flags uint32) {
			if index != 7 || o == nil || flags != 1 {
				panic("unmark arguments")
			}
			*events = append(*events, "unmark")
		},
		unitCode:   func(o *tagTestObject530160) uint16 { return o.code },
		objectType: func(o *tagTestObject530160) uint16 { return o.typeID },
		send: func(index int32, value [7]byte) {
			if index != 7 {
				panic("send recipient")
			}
			*events = append(*events, "send")
			*packet = value
		},
	}
}

func TestSpellTagCreate530160FrameWrapAndPacket(t *testing.T) {
	var events []string
	var packet [7]byte
	record := &tagTestRecord530160{
		caster: &tagTestObject530160{class: 4, index: 7},
		target: &tagTestObject530160{code: 0xa123, typeID: 0xb456},
		level:  16,
	}
	if got := spellTagCreate530160(record, tagTestHooks530160(&events, &packet)); got != 0 {
		t.Fatalf("create returned %d, want 0", got)
	}
	if record.frame != 0x10 {
		t.Fatalf("frame = %#x, want 0x10", record.frame)
	}
	if want := []string{"balance", "round", "store-frame", "mark", "send"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
	if want := [7]byte{0xd2, 0x23, 0xa1, 0x56, 0xb4, 1, 1}; packet != want {
		t.Fatalf("packet = %x, want %x", packet, want)
	}
}

func TestSpellTagCreate530160RejectsInvalidObjectsBeforeEffects(t *testing.T) {
	for _, tc := range []struct {
		name   string
		caster *tagTestObject530160
		target *tagTestObject530160
	}{
		{"nil caster", nil, &tagTestObject530160{}},
		{"dead caster", &tagTestObject530160{class: 4, flags: 0x20}, &tagTestObject530160{}},
		{"nonplayer caster", &tagTestObject530160{}, &tagTestObject530160{}},
		{"nil target", &tagTestObject530160{class: 4}, nil},
		{"dead target", &tagTestObject530160{class: 4}, &tagTestObject530160{flags: 0x8000}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var events []string
			var packet [7]byte
			record := &tagTestRecord530160{caster: tc.caster, target: tc.target, frame: 99}
			if got := spellTagCreate530160(record, tagTestHooks530160(&events, &packet)); got != 1 {
				t.Fatalf("create returned %d, want 1", got)
			}
			if len(events) != 0 || record.frame != 99 || packet != [7]byte{} {
				t.Fatalf("rejected create mutated state: events %v, frame %d, packet %x", events, record.frame, packet)
			}
		})
	}
}

func TestSpellTagDestroy530270UnmarkAndPacket(t *testing.T) {
	for _, tc := range []struct {
		name        string
		targetClass byte
		wantEvents  []string
	}{
		{"nonplayer target", 0, []string{"unmark", "send"}},
		{"player target", 4, []string{"send"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var events []string
			var packet [7]byte
			record := &tagTestRecord530160{
				caster: &tagTestObject530160{class: 4, index: 7},
				target: &tagTestObject530160{class: tc.targetClass, code: 0xa123, typeID: 0xb456},
			}
			spellTagDestroy530270(record, tagTestHooks530160(&events, &packet))
			if !reflect.DeepEqual(events, tc.wantEvents) {
				t.Fatalf("events = %v, want %v", events, tc.wantEvents)
			}
			if want := [7]byte{0xd2, 0x23, 0xa1, 0x56, 0xb4, 2, 1}; packet != want {
				t.Fatalf("packet = %x, want %x", packet, want)
			}
		})
	}
}

func TestSpellTagDestroy530270AbsentCasterOrTarget(t *testing.T) {
	for _, record := range []*tagTestRecord530160{
		{},
		{caster: &tagTestObject530160{}},
		{caster: &tagTestObject530160{class: 4}},
	} {
		var events []string
		var packet [7]byte
		spellTagDestroy530270(record, tagTestHooks530160(&events, &packet))
		if len(events) != 0 || packet != [7]byte{} {
			t.Fatalf("absent object produced effects: %v, %x", events, packet)
		}
	}
}
