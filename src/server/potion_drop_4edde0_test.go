package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type potionDropTestWorld4EDDE0 struct {
	pointArg     string
	ownerArg     string
	itemArg      string
	defaultValue int32
	gameFlags    map[uint32]int32
	fps          uint32
	events       []string
	faultAt      int
	afterDefault func(*potionDropTestWorld4EDDE0)
	afterAudio   func(*potionDropTestWorld4EDDE0)
	afterFlag    func(*potionDropTestWorld4EDDE0, uint32)
}

func newPotionDropTestWorld4EDDE0() *potionDropTestWorld4EDDE0 {
	return &potionDropTestWorld4EDDE0{
		pointArg:     "point-a",
		ownerArg:     "owner-a",
		itemArg:      "item-a",
		defaultValue: 1,
		gameFlags:    make(map[uint32]int32),
		fps:          30,
	}
}

func (w *potionDropTestWorld4EDDE0) event(value string) {
	w.events = append(w.events, value)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(value)
	}
}

func (w *potionDropTestWorld4EDDE0) hooks() potionDropHooks4EDDE0[string, string] {
	return potionDropHooks4EDDE0[string, string]{
		loadPointArg: func() string {
			w.event("point-arg:" + w.pointArg)
			return w.pointArg
		},
		loadOwnerArg: func() string {
			w.event("owner-arg:" + w.ownerArg)
			return w.ownerArg
		},
		loadItemArg: func() string {
			w.event("item-arg:" + w.itemArg)
			return w.itemArg
		},
		defaultDrop: func(owner, item, point string) int32 {
			w.event("default:" + owner + ":" + item + ":" + point)
			value := w.defaultValue
			if w.afterDefault != nil {
				w.afterDefault(w)
			}
			return value
		},
		audio: func(id uint32, item string, kind int32, code uint32) {
			w.event(fmt.Sprintf("audio:%d:%s:%d:%08x", id, item, kind, code))
			if w.afterAudio != nil {
				w.afterAudio(w)
			}
		},
		gameFlag: func(flag uint32) int32 {
			value := w.gameFlags[flag]
			w.event(fmt.Sprintf("game-flag:%08x=%08x", flag, uint32(value)))
			if w.afterFlag != nil {
				w.afterFlag(w, flag)
			}
			return value
		},
		loadGameFPS: func() uint32 {
			w.event(fmt.Sprintf("fps:%08x", w.fps))
			return w.fps
		},
		setDecay: func(item string, delay uint32) {
			w.event(fmt.Sprintf("decay:%s:%08x", item, delay))
		},
	}
}

func potionDropSuccessEvents4EDDE0(fps uint32) []string {
	return []string{
		"point-arg:point-a",
		"owner-arg:owner-a",
		"item-arg:item-a",
		"default:owner-a:item-a:point-a",
		"audio:833:item-a:0:00000000",
		"game-flag:00000800=00000000",
		"game-flag:00001000=00000000",
		fmt.Sprintf("fps:%08x", fps),
		fmt.Sprintf("decay:item-a:%08x", fps*potionDropSeconds4EDDE0),
	}
}

func verifyPotionDropFaultPrefixes4EDDE0(
	t *testing.T,
	want []string,
	build func() *potionDropTestWorld4EDDE0,
) {
	t.Helper()
	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("fault-%d", faultAt), func(t *testing.T) {
			w := build()
			w.faultAt = faultAt
			defer func() {
				if got := recover(); got != want[faultAt-1] {
					t.Fatalf("panic = %v, want %q", got, want[faultAt-1])
				}
				if !reflect.DeepEqual(w.events, want[:faultAt]) {
					t.Fatalf("events = %v, want %v", w.events, want[:faultAt])
				}
			}()
			potionDrop4EDDE0(w.hooks())
		})
	}
}

func TestPotionDrop4EDDE0ExactSuccessTraceAndUint32Wrap(t *testing.T) {
	build := func() *potionDropTestWorld4EDDE0 {
		w := newPotionDropTestWorld4EDDE0()
		w.fps = math.MaxUint32
		return w
	}
	want := potionDropSuccessEvents4EDDE0(math.MaxUint32)
	w := build()
	if got := potionDrop4EDDE0(w.hooks()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
	if want[len(want)-1] != "decay:item-a:ffffffe7" {
		t.Fatalf("wrapped decay = %q", want[len(want)-1])
	}
	verifyPotionDropFaultPrefixes4EDDE0(t, want, build)
}

func TestPotionDrop4EDDE0DefaultGateUsesWholeEAX(t *testing.T) {
	for _, value := range []int32{0, 1, -1, math.MinInt32} {
		t.Run(fmt.Sprintf("%08x", uint32(value)), func(t *testing.T) {
			w := newPotionDropTestWorld4EDDE0()
			w.defaultValue = value
			got := potionDrop4EDDE0(w.hooks())
			if value == 0 {
				want := []string{
					"point-arg:point-a", "owner-arg:owner-a", "item-arg:item-a",
					"default:owner-a:item-a:point-a",
				}
				if got != 0 || !reflect.DeepEqual(w.events, want) {
					t.Fatalf("result/events = %d/%v, want 0/%v", got, w.events, want)
				}
				return
			}
			if got != 1 || !reflect.DeepEqual(w.events, potionDropSuccessEvents4EDDE0(30)) {
				t.Fatalf("result/events = %d/%v", got, w.events)
			}
		})
	}
}

func TestPotionDrop4EDDE0FlagGatesUseWholeEAXAndShortCircuit(t *testing.T) {
	tests := []struct {
		name      string
		coop      int32
		quest     int32
		wantTail  []string
		wantCount int
	}{
		{
			name:      "coop",
			coop:      -1,
			wantTail:  []string{"game-flag:00000800=ffffffff"},
			wantCount: 6,
		},
		{
			name:      "quest",
			quest:     math.MinInt32,
			wantTail:  []string{"game-flag:00000800=00000000", "game-flag:00001000=80000000"},
			wantCount: 7,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			build := func() *potionDropTestWorld4EDDE0 {
				w := newPotionDropTestWorld4EDDE0()
				w.gameFlags[potionDropCoopFlag4EDDE0] = tc.coop
				w.gameFlags[potionDropQuestFlag4EDDE0] = tc.quest
				return w
			}
			w := build()
			if got := potionDrop4EDDE0(w.hooks()); got != 1 {
				t.Fatalf("result = %d, want 1", got)
			}
			prefix := []string{
				"point-arg:point-a", "owner-arg:owner-a", "item-arg:item-a",
				"default:owner-a:item-a:point-a", "audio:833:item-a:0:00000000",
			}
			want := append(append([]string{}, prefix...), tc.wantTail...)
			if len(want) != tc.wantCount || !reflect.DeepEqual(w.events, want) {
				t.Fatalf("events = %v, want %v", w.events, want)
			}
			verifyPotionDropFaultPrefixes4EDDE0(t, want, build)
		})
	}
}

func TestPotionDrop4EDDE0CachesArgumentsButReadsPostAudioState(t *testing.T) {
	w := newPotionDropTestWorld4EDDE0()
	w.gameFlags[potionDropCoopFlag4EDDE0] = -1
	w.gameFlags[potionDropQuestFlag4EDDE0] = -1
	w.afterDefault = func(w *potionDropTestWorld4EDDE0) {
		w.pointArg = "point-b"
		w.ownerArg = "owner-b"
		w.itemArg = "item-b"
	}
	w.afterAudio = func(w *potionDropTestWorld4EDDE0) {
		w.gameFlags[potionDropCoopFlag4EDDE0] = 0
	}
	w.afterFlag = func(w *potionDropTestWorld4EDDE0, flag uint32) {
		switch flag {
		case potionDropCoopFlag4EDDE0:
			w.gameFlags[potionDropQuestFlag4EDDE0] = 0
		case potionDropQuestFlag4EDDE0:
			w.fps = 0x80000001
		}
	}
	if got := potionDrop4EDDE0(w.hooks()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	want := potionDropSuccessEvents4EDDE0(0x80000001)
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}
}

func TestPotionDrop4EDDE0DoesNotDereferenceNilArguments(t *testing.T) {
	owner := new(int)
	point := new(int)
	var item *int
	events := make([]string, 0, 4)
	hooks := potionDropHooks4EDDE0[*int, *int]{
		loadPointArg: func() *int { events = append(events, "point"); return point },
		loadOwnerArg: func() *int { events = append(events, "owner"); return owner },
		loadItemArg:  func() *int { events = append(events, "item"); return item },
		defaultDrop: func(gotOwner, gotItem *int, gotPoint *int) int32 {
			if gotOwner != owner || gotItem != nil || gotPoint != point {
				t.Fatal("default arguments changed")
			}
			events = append(events, "default")
			return 1
		},
		audio: func(id uint32, gotItem *int, kind int32, code uint32) {
			if id != 833 || gotItem != nil || kind != 0 || code != 0 {
				t.Fatal("audio arguments changed")
			}
			events = append(events, "audio")
		},
		gameFlag:    func(uint32) int32 { return 1 },
		loadGameFPS: func() uint32 { t.Fatal("FPS read"); return 0 },
		setDecay:    func(*int, uint32) { t.Fatal("decay called") },
	}
	if got := potionDrop4EDDE0(hooks); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if want := []string{"point", "owner", "item", "default", "audio"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}
