package server

import (
	"reflect"
	"testing"

	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/sound"
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
