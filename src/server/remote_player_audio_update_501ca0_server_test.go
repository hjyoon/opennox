package server

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/sound"
)

func TestRemotePlayerAudioNativeLayout501CA0(t *testing.T) {
	if got := int32(sound.SoundSpellPhonemeUp); got != remotePlayerAudioFirstPhoneme501CA0 {
		t.Fatalf("first phoneme = %d, want %d", got, remotePlayerAudioFirstPhoneme501CA0)
	}
	if got := int32(sound.SoundSpellPhonemeUpLeft); got != remotePlayerAudioLastPhoneme501CA0 {
		t.Fatalf("last phoneme = %d, want %d", got, remotePlayerAudioLastPhoneme501CA0)
	}
	wantPhonemeState := uintptr(188)
	wantUpdatePlayer := uintptr(276)
	wantCameraTarget := uintptr(3628)
	wantPlayerPosition := uintptr(3632)
	wantCurrentPolygon := uintptr(3664)
	wantPlayerZone := uintptr(3668)
	wantPlayerFlags := uintptr(3680)
	wantAudioEventSize := uintptr(36)
	wantAudioEventSound := uintptr(4)
	wantAudioEventPosition := uintptr(8)
	wantAudioEventObject := uintptr(16)
	wantAudioEventKind := uintptr(20)
	wantAudioEventCode := uintptr(24)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantPhonemeState = 240
		wantUpdatePlayer = 336
		wantCameraTarget = 4912
		wantPlayerPosition = 4920
		wantCurrentPolygon = 4960
		wantPlayerZone = 4964
		wantPlayerFlags = 4976
		wantAudioEventSize = 64
		wantAudioEventSound = 8
		wantAudioEventPosition = 16
		wantAudioEventObject = 24
		wantAudioEventKind = 32
		wantAudioEventCode = 40
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"PlayerUpdateData.Field47_0", unsafe.Offsetof(PlayerUpdateData{}.Field47_0), wantPhonemeState},
		{"PlayerUpdateData.Player", unsafe.Offsetof(PlayerUpdateData{}.Player), wantUpdatePlayer},
		{"Player.CameraFollowObj", unsafe.Offsetof(Player{}.CameraFollowObj), wantCameraTarget},
		{"Player.Pos3632Vec", unsafe.Offsetof(Player{}.Pos3632Vec), wantPlayerPosition},
		{"Player.field3664", unsafe.Offsetof(Player{}.field3664), wantCurrentPolygon},
		{"Player.field3668", unsafe.Offsetof(Player{}.field3668), wantPlayerZone},
		{"Player.Field3680", unsafe.Offsetof(Player{}.Field3680), wantPlayerFlags},
		{"AudioEvent size", unsafe.Sizeof(AudioEvent{}), wantAudioEventSize},
		{"AudioEvent.Sound", unsafe.Offsetof(AudioEvent{}.Sound), wantAudioEventSound},
		{"AudioEvent.Pos", unsafe.Offsetof(AudioEvent{}.Pos), wantAudioEventPosition},
		{"AudioEvent.Obj", unsafe.Offsetof(AudioEvent{}.Obj), wantAudioEventObject},
		{"AudioEvent.Kind", unsafe.Offsetof(AudioEvent{}.Kind), wantAudioEventKind},
		{"AudioEvent.Code", unsafe.Offsetof(AudioEvent{}.Code), wantAudioEventCode},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestRemotePlayerAudioNative501CA0PreservesPointers(t *testing.T) {
	listenerPosition := types.Ptf(11, 22)
	player := &Player{
		Pos3632Vec: listenerPosition,
		field3664:  1,
		field3668:  0xe7,
	}
	update := &PlayerUpdateData{Field47_0: 1, Player: player}
	unit := &Object{
		ObjClass:   object.ClassPlayer,
		NetCode:    0xf1234567,
		UpdateData: unsafe.Pointer(update),
	}
	eventObject := &Object{ObjClass: object.ClassMonster}
	event := &AudioEvent{
		Sound: 42,
		Pos:   types.Ptf(33, 44),
		Obj:   eventObject,
	}
	srv := &Server{}
	srv.Audio.head = event

	if unsafe.Sizeof(uintptr(0)) == 8 {
		for name, pointer := range map[string]unsafe.Pointer{
			"unit":              unsafe.Pointer(unit),
			"update":            unsafe.Pointer(update),
			"player":            unsafe.Pointer(player),
			"event":             unsafe.Pointer(event),
			"event object":      unsafe.Pointer(eventObject),
			"event position":    unsafe.Pointer(&event.Pos),
			"listener position": unsafe.Pointer(&player.Pos3632Vec),
		} {
			if uintptr(pointer) <= math.MaxUint32 {
				t.Fatalf("%s pointer = %p, want native address above 4 GiB", name, pointer)
			}
		}
	}

	var zoneCalls, fadeCalls, directCalls, flushCalls int
	srv.UpdateRemotePlayerAudio501CA0(unit, RemotePlayerAudioRuntime501CA0{
		QuestCheckSecretArea: func(*Object) {
			t.Fatal("initialized listener called QuestCheckSecretArea")
		},
		PolygonAtPoint: func([2]int32, uint32) unsafe.Pointer {
			t.Fatal("normal listener called PolygonAtPoint")
			return nil
		},
		PolygonZone: func(unsafe.Pointer) uint8 {
			t.Fatal("normal listener called PolygonZone")
			return 0
		},
		EventZone: func(position *types.Pointf, object *Object) uint8 {
			zoneCalls++
			if position != &event.Pos || object != eventObject {
				t.Fatalf("event zone args = %p/%p, want %p/%p", position, object, &event.Pos, eventObject)
			}
			return 0
		},
		Fade: func(id sound.ID, eventPosition, gotListenerPosition *types.Pointf) int32 {
			fadeCalls++
			if id != event.Sound || eventPosition != &event.Pos || gotListenerPosition != &player.Pos3632Vec {
				t.Fatalf("fade args = %d/%p/%p, want %d/%p/%p", id, eventPosition, gotListenerPosition, event.Sound, &event.Pos, &player.Pos3632Vec)
			}
			return 11
		},
		SendDirect: func(gotUnit *Object, gotEvent *AudioEvent, fade int32) {
			directCalls++
			if gotUnit != unit || gotEvent != event || fade != 5 {
				t.Fatalf("direct args = %p/%p/%d, want %p/%p/5", gotUnit, gotEvent, fade, unit, event)
			}
		},
		Flush: func(got *Object) {
			flushCalls++
			if got != unit {
				t.Fatalf("flush unit = %p, want %p", got, unit)
			}
		},
	})
	if zoneCalls != 1 || fadeCalls != 1 || directCalls != 1 || flushCalls != 1 {
		t.Fatalf("calls zone/fade/direct/flush = %d/%d/%d/%d, want 1/1/1/1", zoneCalls, fadeCalls, directCalls, flushCalls)
	}
	runtime.KeepAlive(eventObject)
	runtime.KeepAlive(event)
	runtime.KeepAlive(player)
	runtime.KeepAlive(update)
	runtime.KeepAlive(unit)
}

func TestRemotePlayerAudioNative501CA0ReloadsPlayerAfterQuest(t *testing.T) {
	oldPlayer := &Player{field3664: remotePlayerAudioUninitialized501CA0, field3668: 1}
	newPlayer := &Player{field3664: 7, field3668: 0xd5}
	update := &PlayerUpdateData{Player: oldPlayer}
	unit := &Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update)}
	srv := &Server{}
	var flushed bool
	srv.UpdateRemotePlayerAudio501CA0(unit, RemotePlayerAudioRuntime501CA0{
		QuestCheckSecretArea: func(got *Object) {
			if got != unit {
				t.Fatalf("quest unit = %p, want %p", got, unit)
			}
			update.Player = newPlayer
		},
		PolygonAtPoint: func([2]int32, uint32) unsafe.Pointer {
			t.Fatal("normal listener called PolygonAtPoint")
			return nil
		},
		PolygonZone: func(unsafe.Pointer) uint8 {
			t.Fatal("normal listener called PolygonZone")
			return 0
		},
		EventZone: func(*types.Pointf, *Object) uint8 {
			t.Fatal("empty audio list called EventZone")
			return 0
		},
		Fade: func(sound.ID, *types.Pointf, *types.Pointf) int32 {
			t.Fatal("empty audio list called Fade")
			return 0
		},
		SendDirect: func(*Object, *AudioEvent, int32) {
			t.Fatal("empty audio list called SendDirect")
		},
		Flush: func(got *Object) {
			flushed = got == unit
		},
	})
	if !flushed {
		t.Fatal("empty audio list was not flushed")
	}
	runtime.KeepAlive(newPlayer)
	runtime.KeepAlive(oldPlayer)
	runtime.KeepAlive(update)
	runtime.KeepAlive(unit)
}
