package server

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func TestResetAudioEventsNative502100OnlyClearsOriginalState(t *testing.T) {
	audio := &serverAudio{}
	audio.Init(&Server{})
	defer audio.Free()

	class := audio.alloc.Class
	event := audio.alloc.NewObject()
	if event == nil {
		t.Fatal("allocate audio event: got nil")
	}
	audio.head = event
	audio.inAudio = false
	audio.bitmap[31] = 0x80000000
	audio.bySound[1023].Field20 = -17
	audio.bySound[1023].Field24 = event
	audio.onSound = append(audio.onSound, func(sound.ID, int, *Object, types.Pointf) {})
	audio.delayedObj = append(audio.delayedObj, audioEventObj{ID: 1, Kind: 2, Code: 3})
	audio.delayedPos = append(audio.delayedPos, audioEventPos{ID: 4, Kind: 5, Code: 6})

	audio.resetAudioEvents502100()

	if audio.head != nil {
		t.Fatalf("head = %p, want nil", audio.head)
	}
	if !audio.inited || audio.alloc.Class != class {
		t.Fatal("00502100 freed or replaced the initialized allocation class")
	}
	if audio.inAudio {
		t.Fatal("00502100 changed the Go-owned audio collection phase")
	}
	if audio.bitmap[31] != 0x80000000 || audio.bySound[1023].Field20 != -17 || audio.bySound[1023].Field24 != event {
		t.Fatal("00502100 changed bitmap or per-sound descriptor state")
	}
	if len(audio.onSound) != 1 || len(audio.delayedObj) != 1 || len(audio.delayedPos) != 1 {
		t.Fatal("00502100 changed listeners or delayed queues")
	}
	if reused := audio.alloc.NewObject(); reused != event {
		t.Fatalf("reused event = %p, want returned active event %p", reused, event)
	}
}

func TestResetAudioEventsNative502100NilClass(t *testing.T) {
	event := &AudioEvent{Sound: 17}
	audio := serverAudio{head: event, inAudio: true}

	audio.resetAudioEvents502100()

	if audio.head != nil {
		t.Fatalf("head = %p, want nil", audio.head)
	}
	if !audio.inAudio || audio.inited || audio.alloc.Class != nil {
		t.Fatal("nil-class cleanup changed unrelated audio lifecycle state")
	}
}

func TestServerAudioResetReplaysDelayedEventsAfter502100(t *testing.T) {
	oldFlags := noxflags.GetGame()
	noxflags.ResetGame()
	defer func() {
		noxflags.ResetGame()
		noxflags.SetGame(oldFlags)
	}()

	audio := &serverAudio{}
	audio.Init(&Server{})
	defer audio.Free()
	oldEvent := audio.alloc.NewObject()
	if oldEvent == nil {
		t.Fatal("allocate old audio event: got nil")
	}
	oldEvent.Sound = 99
	audio.head = oldEvent

	const (
		firstID  = sound.ID(11)
		secondID = sound.ID(12)
	)
	audio.bySound[firstID].Field12 = 1
	audio.bySound[secondID].Field12 = 1
	audio.inAudio = false
	audio.delayedPos = append(audio.delayedPos,
		audioEventPos{ID: firstID, Pos: types.Pointf{X: 1, Y: 2}, Kind: 3, Code: 4},
		audioEventPos{ID: secondID, Pos: types.Pointf{X: 5, Y: 6}, Kind: 7, Code: 8},
	)
	var observed []sound.ID
	audio.onSound = append(audio.onSound, func(id sound.ID, _ int, _ *Object, _ types.Pointf) {
		if !audio.inAudio {
			t.Fatal("delayed replay ran before opening the Go audio collection")
		}
		if audio.head == nil || audio.head.Sound != id {
			t.Fatalf("head during callback = %v, want newly replayed sound %d", audio.head, id)
		}
		observed = append(observed, id)
	})

	audio.Reset()

	if !audio.inAudio {
		t.Fatal("Reset left the Go audio collection closed")
	}
	if len(audio.delayedPos) != 0 {
		t.Fatalf("delayed positions = %d, want zero", len(audio.delayedPos))
	}
	if want := []sound.ID{firstID, secondID}; !reflect.DeepEqual(observed, want) {
		t.Fatalf("replay order = %v, want %v", observed, want)
	}
	if audio.head == nil || audio.head.Sound != secondID || audio.head.next0 == nil ||
		audio.head.next0.Sound != firstID || audio.head.next0.next0 != nil {
		t.Fatalf("replayed list = %#v, want second then first", audio.head)
	}
}

func TestAudioEventObjNative501960ReturnsLiveHead(t *testing.T) {
	oldFlags := noxflags.GetGame()
	noxflags.ResetGame()
	t.Cleanup(func() { noxflags.ResetGame(); noxflags.SetGame(oldFlags) })

	audio := &serverAudio{alloc: alloc.NewClassTHigh("AudEvent", AudioEvent{}, 3)}
	t.Cleanup(audio.alloc.Free)
	obj := nativeTradeTestValue(t, Object{PosVec: types.Ptf(17, 23)})
	const id = sound.SoundNoCanDo
	const code = uint32(0xa5b61234)
	disabled := int32(-31)
	audio.bySound[id].Field12 = uint32(disabled)
	if got := audio.EventObjNative501960(id, obj, 2, code); got != uintptr(disabled) || audio.head != nil {
		t.Fatalf("unavailable sound returned %#x, head %p", got, audio.head)
	}
	noxflags.SetGame(noxflags.GameFlag20)
	if got := audio.EventObjNative501960(id, obj, 2, code); got != 1 || audio.head != nil {
		t.Fatalf("muted sound returned %#x, head %p", got, audio.head)
	}
	noxflags.ResetGame()
	audio.bySound[id].Field12 = 1
	if got := audio.EventObjNative501960(id, obj, 2, code); got != 0 || audio.head == nil {
		t.Fatalf("first sound returned %#x, head %p", got, audio.head)
	}
	first := audio.head
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(first)) <= uintptr(^uint32(0)) {
		t.Fatalf("event address %p does not exercise native pointer width", first)
	}
	var nested *AudioEvent
	audio.OnSound(func(gotID sound.ID, kind int, unit *Object, pos types.Pointf) {
		if audio.head != first || gotID != id || kind != 0 || unit != obj || pos != obj.PosVec {
			t.Fatalf("sound callback saw head %p, sound %d, kind %d, object %p, position %v", audio.head, gotID, kind, unit, pos)
		}
		if got := audio.EventObjNative501960(id, obj, 2, code); got != uintptr(unsafe.Pointer(first)) {
			t.Fatalf("nested sound returned %#x, want %p", got, first)
		}
		nested = audio.head
	})
	obj.ObjFlags = object.FlagDestroyed
	if got := audio.EventObjNative501960(id, obj, 0, code); nested == nil || got != uintptr(unsafe.Pointer(nested)) {
		t.Fatalf("outer sound returned %#x, want %p", got, nested)
	}
	head := audio.head
	if head.next0 != nested || nested.next0 != first || head.Sound != id || head.Obj != nil || head.Pos != obj.PosVec ||
		head.Kind != 0 || head.Code != code || first.Obj != obj || first.Kind != 2 || first.Code != code {
		t.Fatalf("event list = %#v, nested %#v, first %#v", head, nested, first)
	}
	if got := audio.EventObjNative501960(id, obj, 0, code); got != 0 || audio.head != head {
		t.Fatalf("exhausted event pool returned %#x, head %p", got, audio.head)
	}
}
