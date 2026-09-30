package legacy

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

func newQuestStageFixture450980(t *testing.T, key string, flags *uint8) (*questStartFixture450A30, questStageBriefingHooks450980[questStartRefModel450A30, *int, *int]) {
	t.Helper()
	f := newQuestStartFixture450A30(t, key)
	h := questStageBriefingHooks450980[questStartRefModel450A30, *int, *int]{questStartBriefingHooks450A30: f.h}
	h.storeState = func(v uint32) { f.record(fmt.Sprintf("state:%d", v)); f.state = v }
	h.loadText = func(p questStartRefModel450A30, source string, line int32) *int {
		f.record("load-text")
		if p.packet != f.packet || p.off != 37 || source != `C:\NoxPost\src\client\Gui\GUIBrief.c` || line != 1714 {
			t.Fatalf("lookup=%+v %q %d", p, source, line)
		}
		return f.text
	}
	h.flags = func(p questStartRefModel450A30) uint8 {
		f.record("flags")
		if p.packet != f.packet || p.off != 0 {
			t.Fatalf("flags packet=%+v", p)
		}
		return *flags
	}
	h.lock = func(screen, mode int32, flags int8) int32 {
		f.record("lock")
		if screen != 254 || mode != 1 || flags != 2 {
			t.Fatalf("lock=%d/%d/%d", screen, mode, flags)
		}
		return math.MinInt32 + 1
	}
	return f, h
}

func questStageTrace450980(hasKey, stateOne, show bool) []string {
	want := []string{"state:0", "reset", "hide", "prepare", "image-address", "load-image", "set-image", "key-address", "strlen"}
	if hasKey {
		want = append(want, "load-text")
	} else {
		want = append(want, "empty-text")
	}
	want = append(want, "set-text", "stage", "set-stage", "flags")
	if stateOne {
		want = append(want, "state:1")
	}
	if show {
		want = append(want, "lock")
	}
	return want
}

func TestQuestStageBriefing450980BranchesOrderAndReturns(t *testing.T) {
	for _, key := range []string{"", "Description"} {
		for _, flags := range []uint8{0, 1, 2, 3, 4, 0x80, 0xff} {
			for _, show := range []int32{0, 1, 256, -256, math.MinInt32, math.MaxInt32} {
				f, h := newQuestStageFixture450980(t, key, &flags)
				before := *f.packet
				got := questStageBriefing450980(questStartRefModel450A30{packet: f.packet}, show, h)
				wantReturn, wantState := int32(0), uint32(0)
				if show != 0 {
					wantReturn = math.MinInt32 + 1
				}
				if flags&2 != 0 {
					wantState = 1
				}
				wantText := f.text
				if key == "" {
					wantText = f.empty
				}
				if got != wantReturn || f.state != wantState || f.storedStage != 0xfedc || f.storedImage != f.image || f.storedText != wantText || *f.packet != before {
					t.Fatalf("key=%q flags=%#x show=%d return=%d state=%d stage=%#x image=%p text=%p packet=%+v", key, flags, show, got, f.state, f.storedStage, f.storedImage, f.storedText, f.packet)
				}
				if want := questStageTrace450980(key != "", flags&2 != 0, show != 0); !reflect.DeepEqual(f.events, want) {
					t.Fatalf("events=%v want=%v", f.events, want)
				}
			}
		}
	}
}

func TestQuestStageBriefing450980LiveReadsAndNilSetters(t *testing.T) {
	flags := uint8(0)
	f, h := newQuestStageFixture450980(t, "original", &flags)
	f.image, f.text = nil, nil
	h.prepare = func() { f.record("prepare"); f.packet.image = "prepared" }
	h.loadImage = func(p questStartRefModel450A30) *int {
		f.record("load-image")
		if p.packet.image != "prepared" || p.off != 5 {
			t.Fatal("image was read too early")
		}
		return nil
	}
	h.setImage = func(image *int) { f.record("set-image"); f.storedImage = image; f.packet.key = "setter-key" }
	h.stringLength = func(p questStartRefModel450A30) uintptr {
		f.record("strlen")
		if p.packet.key != "setter-key" || p.off != 37 {
			t.Fatal("key was read too early")
		}
		p.packet.key = "" // Nonzero length still chooses lookup at the same address.
		return 10
	}
	h.loadText = func(p questStartRefModel450A30, source string, line int32) *int {
		f.record("load-text")
		if p.packet.key != "" || p.off != 37 || source != questStageBriefingSource450980 || line != 1714 {
			t.Fatal("lookup snapshot or wrong source")
		}
		return nil
	}
	h.setText = func(text *int) { f.record("set-text"); f.storedText = text; f.packet.stage = math.MaxUint16 }
	h.setStage = func(v uint32) { f.record("set-stage"); f.storedStage = v; flags = 2 }
	if got := questStageBriefing450980(questStartRefModel450A30{packet: f.packet}, -256, h); got != math.MinInt32+1 || f.state != 1 || f.storedStage != math.MaxUint16 || f.storedImage != nil || f.storedText != nil {
		t.Fatalf("return=%d fixture=%+v", got, f)
	}
	if want := questStageTrace450980(true, true, true); !reflect.DeepEqual(f.events, want) {
		t.Fatalf("events=%v want=%v", f.events, want)
	}
	// A clear bit does not reset state changes made by a previous callback.
	f, h = newQuestStageFixture450980(t, "", &flags)
	h.setStage = func(v uint32) { f.record("set-stage"); f.storedStage = v; flags = 0; f.state = 9 }
	questStageBriefing450980(questStartRefModel450A30{packet: f.packet}, 0, h)
	if f.state != 9 {
		t.Fatalf("state cleared after callback: %d", f.state)
	}
}

func TestQuestStageBriefing450980EveryDependencyFaultPrefix(t *testing.T) {
	for _, key := range []string{"", "Description"} {
		for _, flags := range []uint8{0, 2} {
			trace := questStageTrace450980(key != "", flags&2 != 0, true)
			for i, event := range trace {
				t.Run(fmt.Sprintf("%s/%d/%s", key, flags, event), func(t *testing.T) {
					f, h := newQuestStageFixture450980(t, key, &flags)
					f.failAt = event
					defer func() {
						if got := recover(); got != "injected "+event {
							t.Fatalf("fault=%v", got)
						}
						if !reflect.DeepEqual(f.events, trace[:i+1]) {
							t.Fatalf("events=%v want=%v", f.events, trace[:i+1])
						}
						wantState := uint32(0)
						if i == 0 {
							wantState = math.MaxUint32
						} else if event == "lock" && flags&2 != 0 {
							wantState = 1
						}
						if f.state != wantState {
							t.Fatalf("partial state=%d want=%d", f.state, wantState)
						}
					}()
					questStageBriefing450980(questStartRefModel450A30{packet: f.packet}, 1, h)
				})
			}
		}
	}
}

func TestQuestStageBriefing450980NilPacketFaultsAfterPreparation(t *testing.T) {
	flags := uint8(0)
	f, h := newQuestStageFixture450980(t, "", &flags)
	defer func() {
		if recover() == nil {
			t.Fatal("nil packet silently ignored")
		}
		if want := questStageTrace450980(false, false, false)[:6]; !reflect.DeepEqual(f.events, want) {
			t.Fatalf("events=%v want=%v", f.events, want)
		}
	}()
	questStageBriefing450980(questStartRefModel450A30{}, 0, h)
}
