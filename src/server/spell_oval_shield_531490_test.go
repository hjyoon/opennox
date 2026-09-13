package server

import (
	"reflect"
	"testing"
)

type ovalShieldTestObject531490 struct {
	class uint8
	flags uint32
}

type ovalShieldTestRecord531490 struct {
	level  uint32
	target *ovalShieldTestObject531490
	frame  uint32
}

func ovalShieldTestHooks531490(events *[]string) spellOvalShieldHooks531490[*ovalShieldTestRecord531490, *ovalShieldTestObject531490] {
	return spellOvalShieldHooks531490[*ovalShieldTestRecord531490, *ovalShieldTestObject531490]{
		loadFPS: func() uint32 {
			*events = append(*events, "fps")
			return 30
		},
		loadLevel: func(record *ovalShieldTestRecord531490) uint32 {
			*events = append(*events, "level")
			return record.level
		},
		loadTarget: func(record *ovalShieldTestRecord531490) *ovalShieldTestObject531490 {
			*events = append(*events, "target")
			return record.target
		},
		loadClass: func(target *ovalShieldTestObject531490) uint8 {
			*events = append(*events, "class")
			return target.class
		},
		loadFlags: func(target *ovalShieldTestObject531490) uint32 {
			*events = append(*events, "flags")
			return target.flags
		},
		cancelOffensive: func(*ovalShieldTestObject531490) {
			*events = append(*events, "cancel")
		},
		applyBuff: func(_ *ovalShieldTestObject531490, buff int32, duration int16, power int8) {
			*events = append(*events, "apply")
			if buff != 27 || duration != 600 || power != 1 {
				panic("unexpected buff arguments")
			}
		},
		loadFrame: func() uint32 {
			*events = append(*events, "frame")
			return 100
		},
		storeFrame: func(record *ovalShieldTestRecord531490, frame uint32) {
			*events = append(*events, "store")
			record.frame = frame
		},
		testBuff: func(_ *ovalShieldTestObject531490, buff int32) int32 {
			*events = append(*events, "test-buff")
			if buff != 8 {
				panic("unexpected buff ID")
			}
			return 0
		},
		positionDelta: func(*ovalShieldTestObject531490, *ovalShieldTestRecord531490) int32 {
			*events = append(*events, "position")
			return 0
		},
		buffOff: func(_ *ovalShieldTestObject531490, buff int32) {
			*events = append(*events, "buff-off")
			if buff != 27 {
				panic("unexpected buff ID")
			}
		},
	}
}

func TestSpellOvalShieldCreate531490LiveReadsAndFrame(t *testing.T) {
	var events []string
	first := &ovalShieldTestObject531490{class: 4}
	second := &ovalShieldTestObject531490{class: 4}
	record := &ovalShieldTestRecord531490{level: 2, target: first}
	h := ovalShieldTestHooks531490(&events)
	h.cancelOffensive = func(target *ovalShieldTestObject531490) {
		events = append(events, "cancel")
		if target != first {
			t.Fatal("cancel target changed before call")
		}
		record.level = 1
		record.target = second
	}
	h.applyBuff = func(target *ovalShieldTestObject531490, buff int32, duration int16, power int8) {
		events = append(events, "apply")
		if target != second || buff != 27 || duration != 1200 || power != 1 {
			t.Fatalf("apply = %p/%d/%d/%d", target, buff, duration, power)
		}
	}
	if got := spellOvalShieldCreate531490(record, h); got != 0 {
		t.Fatalf("create = %d, want 0", got)
	}
	if record.frame != 1300 {
		t.Fatalf("frame = %d, want 1300", record.frame)
	}
	want := []string{"fps", "level", "target", "class", "cancel", "level", "target", "apply", "frame", "store"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestSpellOvalShieldCreate531490RejectsNonPlayerAndWrapsDword(t *testing.T) {
	for _, target := range []*ovalShieldTestObject531490{nil, {class: 2}} {
		var events []string
		record := &ovalShieldTestRecord531490{level: 1, target: target, frame: 55}
		if got := spellOvalShieldCreate531490(record, ovalShieldTestHooks531490(&events)); got != 1 || record.frame != 55 {
			t.Fatalf("rejected create = %d, frame = %d", got, record.frame)
		}
		for _, event := range events {
			if event == "cancel" || event == "apply" || event == "store" {
				t.Fatalf("rejected create performed %s", event)
			}
		}
	}
	var events []string
	record := &ovalShieldTestRecord531490{level: 0xffffffff, target: &ovalShieldTestObject531490{class: 4}}
	h := ovalShieldTestHooks531490(&events)
	h.loadFPS = func() uint32 { return 0xffffffff }
	h.loadFrame = func() uint32 { return 0xfffffff0 }
	h.applyBuff = func(_ *ovalShieldTestObject531490, buff int32, duration int16, power int8) {
		if buff != 27 || duration != 20 || power != -1 {
			t.Fatalf("wrapped apply = %d/%d/%d", buff, duration, power)
		}
	}
	if got := spellOvalShieldCreate531490(record, h); got != 0 || record.frame != 4 {
		t.Fatalf("wrapped create = %d, frame = %#x", got, record.frame)
	}
}

func TestSpellOvalShieldUpdate5314F0Branches(t *testing.T) {
	for _, tc := range []struct {
		name    string
		class   uint8
		flags   uint32
		blocked bool
		moved   bool
		want    int32
		calls   []string
	}{
		{"player active", 4, 0, false, false, 0, []string{"target", "test-buff", "target", "class", "target", "flags"}},
		{"blocked buff", 4, 0, true, false, 1, []string{"target", "test-buff"}},
		{"monster active", 2, 0, false, false, 0, []string{"target", "test-buff", "target", "class", "position", "target", "flags"}},
		{"monster moved", 2, 0, false, true, 1, []string{"target", "test-buff", "target", "class", "position"}},
		{"dead flag", 4, 0x20, false, false, 1, []string{"target", "test-buff", "target", "class", "target", "flags"}},
		{"removed flag", 4, 0x8000, false, false, 1, []string{"target", "test-buff", "target", "class", "target", "flags"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var events []string
			record := &ovalShieldTestRecord531490{target: &ovalShieldTestObject531490{class: tc.class, flags: tc.flags}}
			h := ovalShieldTestHooks531490(&events)
			h.testBuff = func(*ovalShieldTestObject531490, int32) int32 {
				events = append(events, "test-buff")
				if tc.blocked {
					return 1
				}
				return 0
			}
			h.positionDelta = func(*ovalShieldTestObject531490, *ovalShieldTestRecord531490) int32 {
				events = append(events, "position")
				if tc.moved {
					return 1
				}
				return 0
			}
			if got := spellOvalShieldUpdate5314F0(record, h); got != tc.want {
				t.Fatalf("update = %d, want %d", got, tc.want)
			}
			if !reflect.DeepEqual(events, tc.calls) {
				t.Fatalf("events = %v, want %v", events, tc.calls)
			}
		})
	}
}

func TestSpellOvalShieldUpdate5314F0ReloadsTarget(t *testing.T) {
	var events []string
	record := &ovalShieldTestRecord531490{target: &ovalShieldTestObject531490{class: 4}}
	h := ovalShieldTestHooks531490(&events)
	h.testBuff = func(*ovalShieldTestObject531490, int32) int32 {
		events = append(events, "test-buff")
		record.target = &ovalShieldTestObject531490{class: 2, flags: 0x20}
		return 0
	}
	if got := spellOvalShieldUpdate5314F0(record, h); got != 1 {
		t.Fatalf("reloaded target update = %d, want 1", got)
	}
	if !reflect.DeepEqual(events, []string{"target", "test-buff", "target", "class", "position", "target", "flags"}) {
		t.Fatalf("events = %v", events)
	}
	record.target = nil
	if got := spellOvalShieldUpdate5314F0(record, h); got != 1 {
		t.Fatalf("nil target update = %d, want 1", got)
	}
}

func TestSpellOvalShieldDestroy531560OptionalTarget(t *testing.T) {
	var events []string
	record := &ovalShieldTestRecord531490{}
	h := ovalShieldTestHooks531490(&events)
	spellOvalShieldDestroy531560(record, h)
	if !reflect.DeepEqual(events, []string{"target"}) {
		t.Fatalf("nil destroy events = %v", events)
	}
	record.target = &ovalShieldTestObject531490{}
	events = nil
	spellOvalShieldDestroy531560(record, h)
	if !reflect.DeepEqual(events, []string{"target", "buff-off"}) {
		t.Fatalf("destroy events = %v", events)
	}
}
