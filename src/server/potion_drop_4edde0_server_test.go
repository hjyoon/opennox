package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/sound"
)

func defaultPotionDropNativeDeps4EDDE0() potionDropNativeDeps4EDDE0 {
	return potionDropNativeDeps4EDDE0{
		defaultDrop: func(*Object, *Object, *types.Pointf) int32 { return 0 },
		audio:       func(uint32, *Object, int32, uint32) {},
		gameFlag:    func(uint32) int32 { return 0 },
		loadGameFPS: func() uint32 { return 0 },
		setDecay:    func(*Object, uint32) {},
	}
}

func TestPotionDrop4EDDE0NativeLayout(t *testing.T) {
	if got := unsafe.Sizeof(types.Pointf{}); got != 8 {
		t.Fatalf("Pointf size = %d, want 8", got)
	}
	if got := unsafe.Offsetof(types.Pointf{}.X); got != 0 {
		t.Fatalf("Pointf.X offset = %d, want 0", got)
	}
	if got := unsafe.Offsetof(types.Pointf{}.Y); got != 4 {
		t.Fatalf("Pointf.Y offset = %d, want 4", got)
	}
}

func TestPotionDropNative4EDDE0BindsPointersAndServices(t *testing.T) {
	owner := &Object{}
	item := &Object{}
	point := &types.Pointf{X: 3.5, Y: -9.25}
	events := make([]string, 0, 5)
	deps := defaultPotionDropNativeDeps4EDDE0()
	deps.defaultDrop = func(gotOwner, gotItem *Object, gotPoint *types.Pointf) int32 {
		events = append(events, "default")
		if gotOwner != owner || gotItem != item || gotPoint != point {
			t.Fatalf("default args = %p/%p/%p", gotOwner, gotItem, gotPoint)
		}
		return -1
	}
	deps.audio = func(id uint32, gotItem *Object, kind int32, code uint32) {
		events = append(events, "audio")
		if id != 833 || gotItem != item || kind != 0 || code != 0 {
			t.Fatalf("audio args = %d/%p/%d/%08x", id, gotItem, kind, code)
		}
	}
	deps.gameFlag = func(flag uint32) int32 {
		events = append(events, fmt.Sprintf("flag:%04x", flag))
		return 0
	}
	deps.loadGameFPS = func() uint32 {
		events = append(events, "fps")
		return math.MaxUint32
	}
	deps.setDecay = func(gotItem *Object, delay uint32) {
		events = append(events, "decay")
		if gotItem != item || delay != 0xffffffe7 {
			t.Fatalf("decay args = %p/%08x", gotItem, delay)
		}
	}

	if got := potionDropNative4EDDE0(owner, item, point, deps); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	want := []string{"default", "audio", "flag:0800", "flag:1000", "fps", "decay"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestPotionDropNative4EDDE0DefaultFailureAcceptsNilPointers(t *testing.T) {
	deps := defaultPotionDropNativeDeps4EDDE0()
	deps.defaultDrop = func(owner, item *Object, point *types.Pointf) int32 {
		if owner != nil || item != nil || point != nil {
			t.Fatalf("default args = %p/%p/%p", owner, item, point)
		}
		return 0
	}
	deps.audio = func(uint32, *Object, int32, uint32) { t.Fatal("audio called") }
	deps.gameFlag = func(uint32) int32 { t.Fatal("game flag read"); return 0 }
	deps.loadGameFPS = func() uint32 { t.Fatal("FPS read"); return 0 }
	deps.setDecay = func(*Object, uint32) { t.Fatal("decay called") }
	if got := potionDropNative4EDDE0(nil, nil, nil, deps); got != 0 {
		t.Fatalf("result = %d, want 0", got)
	}
}

func TestPotionDrop4EDDE0ServerBindingUsesFlagsTickRateAudioAndNativeDecay(t *testing.T) {
	oldFlags := noxflags.GetGame()
	noxflags.ResetGame()
	t.Cleanup(func() {
		noxflags.ResetGame()
		noxflags.SetGame(oldFlags)
	})

	s := &Server{}
	s.SetFrame(100)
	s.SetTickRate(73)
	owner := &Object{}
	item := &Object{}
	point := &types.Pointf{X: 7, Y: 11}
	runtime := PotionDropRuntime4EDDE0{
		DefaultDrop: func(gotOwner, gotItem *Object, gotPoint *types.Pointf) int32 {
			if gotOwner != owner || gotItem != item || gotPoint != point {
				t.Fatalf("default args = %p/%p/%p", gotOwner, gotItem, gotPoint)
			}
			return 1
		},
	}
	if got := s.PotionDrop4EDDE0(owner, item, point, runtime); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if len(s.Audio.delayedObj) != 1 {
		t.Fatalf("queued audio count = %d, want 1", len(s.Audio.delayedObj))
	}
	audio := s.Audio.delayedObj[0]
	if audio.ID != sound.SoundPotionDrop || audio.Obj != item || audio.Kind != 0 || audio.Code != 0 {
		t.Fatalf("queued audio = %#v", audio)
	}
	wantDelay := uint32(73 * 25)
	if s.decay.head != item || s.decay.next[item] != nil {
		t.Fatalf("decay links = head %p next %p", s.decay.head, s.decay.next[item])
	}
	if item.Field34 != 100+wantDelay || uint32(item.ObjFlags)&decayListedFlag511660 == 0 {
		t.Fatalf("decay state = deadline %d flags %08x", item.Field34, uint32(item.ObjFlags))
	}
}

func TestPotionDropServerDeps4EDDE0ObserveLiveGameFlags(t *testing.T) {
	oldFlags := noxflags.GetGame()
	noxflags.ResetGame()
	t.Cleanup(func() {
		noxflags.ResetGame()
		noxflags.SetGame(oldFlags)
	})

	deps := potionDropServerDeps4EDDE0(&Server{}, PotionDropRuntime4EDDE0{})
	if got := deps.gameFlag(potionDropCoopFlag4EDDE0); got != 0 {
		t.Fatalf("clear coop = %d, want 0", got)
	}
	noxflags.SetGame(noxflags.GameModeCoop)
	if got := deps.gameFlag(potionDropCoopFlag4EDDE0); got != 1 {
		t.Fatalf("set coop = %d, want 1", got)
	}
	if got := deps.gameFlag(potionDropQuestFlag4EDDE0); got != 0 {
		t.Fatalf("clear quest = %d, want 0", got)
	}
	noxflags.SetGame(noxflags.GameModeQuest)
	if got := deps.gameFlag(potionDropQuestFlag4EDDE0); got != 1 {
		t.Fatalf("set quest = %d, want 1", got)
	}
}
