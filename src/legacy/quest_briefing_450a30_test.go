package legacy

import (
	"math"
	"reflect"
	"testing"
)

type questStartPacketModel450A30 struct {
	image, key string
	stage      uint16
}

type questStartRefModel450A30 struct {
	packet *questStartPacketModel450A30
	off    uintptr
}

type questStartFixture450A30 struct {
	packet             *questStartPacketModel450A30
	image, text, empty *int
	storedImage        *int
	storedText         *int
	storedStage        uint32
	state              uint32
	events             []string
	failAt             string
	h                  questStartBriefingHooks450A30[questStartRefModel450A30, *int, *int]
}

func newQuestStartFixture450A30(t *testing.T, key string) *questStartFixture450A30 {
	t.Helper()
	f := &questStartFixture450A30{
		packet: &questStartPacketModel450A30{image: "QuestStartImage", key: key, stage: 0xfedc},
		image:  new(int), text: new(int), empty: new(int), state: math.MaxUint32,
	}
	f.h = questStartBriefingHooks450A30[questStartRefModel450A30, *int, *int]{
		storeState:     func(v uint32) { f.record("state"); f.state = v },
		resetParticles: func() { f.record("reset") },
		hideBook: func(v int32) int32 {
			f.record("hide")
			if v != 1 {
				t.Fatalf("hide=%d", v)
			}
			return math.MinInt32 // Must not affect the final return.
		},
		prepare: func() { f.record("prepare") },
		offset: func(p questStartRefModel450A30, off uintptr) questStartRefModel450A30 {
			if off == 5 {
				f.record("image-address")
			} else if off == 37 {
				f.record("key-address")
			} else {
				t.Fatalf("offset=%d", off)
			}
			return questStartRefModel450A30{p.packet, off}
		},
		loadImage: func(p questStartRefModel450A30) *int {
			f.record("load-image")
			if p.off != 5 || p.packet.image != "QuestStartImage" {
				t.Fatalf("image ref=%+v", p)
			}
			return f.image
		},
		setImage: func(p *int) { f.record("set-image"); f.storedImage = p },
		stringLength: func(p questStartRefModel450A30) uintptr {
			f.record("strlen")
			if p.off != 37 {
				t.Fatalf("key ref=%+v", p)
			}
			return uintptr(len(p.packet.key))
		},
		loadText: func(p questStartRefModel450A30, source string, line int32) *int {
			f.record("load-text")
			if p.off != 37 || source != `C:\NoxPost\src\client\Gui\GUIBrief.c` || line != 1756 {
				t.Fatalf("lookup=%+v %q %d", p, source, line)
			}
			return f.text
		},
		emptyText: func() *int { f.record("empty-text"); return f.empty },
		setText:   func(p *int) { f.record("set-text"); f.storedText = p },
		stage:     func(p questStartRefModel450A30) uint16 { f.record("stage"); return p.packet.stage },
		setStage:  func(v uint32) { f.record("set-stage"); f.storedStage = v },
		lock: func(screen, mode int32, flags int8) int32 {
			f.record("lock")
			if screen != 254 || mode != 1 || flags != 4 {
				t.Fatalf("lock=%d/%d/%d", screen, mode, flags)
			}
			return math.MinInt32 + 1
		},
	}
	return f
}

func (f *questStartFixture450A30) record(event string) {
	f.events = append(f.events, event)
	if f.failAt == event {
		panic("injected " + event)
	}
}

func questStartTrace450A30(hasKey, show bool) []string {
	want := []string{"state", "reset", "hide", "prepare", "image-address", "load-image", "set-image", "key-address", "strlen"}
	if hasKey {
		want = append(want, "load-text")
	} else {
		want = append(want, "empty-text")
	}
	want = append(want, "set-text", "stage", "set-stage")
	if show {
		want = append(want, "lock")
	}
	return want
}

func TestQuestStartBriefing450A30BranchesOrderAndReturns(t *testing.T) {
	for _, key := range []string{"", "QuestDescription"} {
		for _, show := range []int32{0, 1, 256, -1, math.MinInt32, math.MaxInt32} {
			f := newQuestStartFixture450A30(t, key)
			before := *f.packet
			got := questStartBriefing450A30(questStartRefModel450A30{packet: f.packet}, show, f.h)
			wantReturn := int32(0)
			if show != 0 {
				wantReturn = math.MinInt32 + 1
			}
			wantText := f.text
			if key == "" {
				wantText = f.empty
			}
			if got != wantReturn || f.state != 0 || f.storedStage != 0xfedc || f.storedImage != f.image || f.storedText != wantText || *f.packet != before {
				t.Fatalf("key=%q show=%d return=%d state=%#x stage=%#x image=%p text=%p packet=%+v", key, show, got, f.state, f.storedStage, f.storedImage, f.storedText, f.packet)
			}
			if want := questStartTrace450A30(key != "", show != 0); !reflect.DeepEqual(f.events, want) {
				t.Fatalf("events=%v want=%v", f.events, want)
			}
		}
	}
}

func TestQuestStartBriefing450A30NilResultsStillReachSetters(t *testing.T) {
	f := newQuestStartFixture450A30(t, "QuestDescription")
	f.image, f.text = nil, nil
	if got := questStartBriefing450A30(questStartRefModel450A30{packet: f.packet}, 0, f.h); got != 0 || f.storedImage != nil || f.storedText != nil || f.storedStage != 0xfedc {
		t.Fatalf("return=%d fixture=%+v", got, f)
	}
	if want := questStartTrace450A30(true, false); !reflect.DeepEqual(f.events, want) {
		t.Fatalf("events=%v want=%v", f.events, want)
	}
}

func TestQuestStartBriefing450A30LiveReadsAcrossCallbacks(t *testing.T) {
	f := newQuestStartFixture450A30(t, "original")
	f.h.prepare = func() {
		f.record("prepare")
		f.packet.image, f.packet.key, f.packet.stage = "prepared", "prepared-key", 1
	}
	f.h.loadImage = func(p questStartRefModel450A30) *int {
		f.record("load-image")
		if p.packet != f.packet || p.off != 5 || p.packet.image != "prepared" {
			t.Fatalf("early image read: %+v", p)
		}
		f.packet.key, f.packet.stage = "image-key", 2
		return f.image
	}
	f.h.setImage = func(image *int) { f.record("set-image"); f.packet.key, f.packet.stage = "setter-key", 3 }
	f.h.stringLength = func(p questStartRefModel450A30) uintptr {
		f.record("strlen")
		if p.packet != f.packet || p.off != 37 || p.packet.key != "setter-key" {
			t.Fatalf("early key read: %+v", p)
		}
		f.packet.key, f.packet.stage = "", 4 // The cached nonzero length still selects loadText.
		return 10
	}
	f.h.loadText = func(p questStartRefModel450A30, source string, line int32) *int {
		f.record("load-text")
		if p.packet.key != "" || p.off != 37 || source != questStartBriefingSource450A30 || line != 1756 {
			t.Fatal("cached key address was not re-read by lookup")
		}
		f.packet.stage = 5
		return f.text
	}
	f.h.setText = func(text *int) { f.record("set-text"); f.packet.stage = math.MaxUint16 }
	if got := questStartBriefing450A30(questStartRefModel450A30{packet: f.packet}, -256, f.h); got != math.MinInt32+1 || f.storedStage != math.MaxUint16 {
		t.Fatalf("return=%d stage=%#x", got, f.storedStage)
	}
	if want := questStartTrace450A30(true, true); !reflect.DeepEqual(f.events, want) {
		t.Fatalf("events=%v want=%v", f.events, want)
	}
}

func TestQuestStartBriefing450A30DependencyFaultPrefixes(t *testing.T) {
	for _, key := range []string{"", "QuestDescription"} {
		trace := questStartTrace450A30(key != "", true)
		for index, event := range trace {
			t.Run(key+"/"+event, func(t *testing.T) {
				f := newQuestStartFixture450A30(t, key)
				f.failAt = event
				defer func() {
					if got := recover(); got != "injected "+event {
						t.Fatalf("fault=%v", got)
					}
					if !reflect.DeepEqual(f.events, trace[:index+1]) {
						t.Fatalf("events=%v want=%v", f.events, trace[:index+1])
					}
				}()
				questStartBriefing450A30(questStartRefModel450A30{packet: f.packet}, 1, f.h)
			})
		}
	}
}

func TestQuestStartBriefing450A30NilPacketFaultAfterPreparation(t *testing.T) {
	f := newQuestStartFixture450A30(t, "")
	defer func() {
		if recover() == nil {
			t.Fatal("nil packet was silently ignored")
		}
		if want := questStartTrace450A30(false, false)[:6]; !reflect.DeepEqual(f.events, want) {
			t.Fatalf("events=%v want=%v", f.events, want)
		}
	}()
	questStartBriefing450A30(questStartRefModel450A30{}, 0, f.h)
}
