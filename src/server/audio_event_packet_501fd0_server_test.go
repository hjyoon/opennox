package server

import (
	"math"
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/internal/netlist"
)

func TestAudioEventPacketNativeLayout501FD0(t *testing.T) {
	wantObjectUpdate := uintptr(748)
	wantUpdatePlayer := uintptr(276)
	wantPlayerIndex := uintptr(2064)
	wantPlayerPosition := uintptr(3632)
	wantEventSize := uintptr(36)
	wantEventSound := uintptr(4)
	wantEventSoundWidth := uintptr(4)
	wantEventPosition := uintptr(8)
	wantEventObject := uintptr(16)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectUpdate = 872
		wantUpdatePlayer = 336
		wantPlayerIndex = 2068
		wantPlayerPosition = 4920
		wantEventSize = 64
		wantEventSound = 8
		wantEventSoundWidth = 8
		wantEventPosition = 16
		wantEventObject = 24
	}
	for _, check := range []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object.UpdateData", unsafe.Offsetof(Object{}.UpdateData), wantObjectUpdate},
		{"PlayerUpdateData.Player", unsafe.Offsetof(PlayerUpdateData{}.Player), wantUpdatePlayer},
		{"Player.PlayerInd", unsafe.Offsetof(Player{}.PlayerInd), wantPlayerIndex},
		{"Player.Pos3632Vec", unsafe.Offsetof(Player{}.Pos3632Vec), wantPlayerPosition},
		{"AudioEvent size", unsafe.Sizeof(AudioEvent{}), wantEventSize},
		{"AudioEvent.Sound", unsafe.Offsetof(AudioEvent{}.Sound), wantEventSound},
		{"AudioEvent.Pos", unsafe.Offsetof(AudioEvent{}.Pos), wantEventPosition},
		{"AudioEvent.Obj", unsafe.Offsetof(AudioEvent{}.Obj), wantEventObject},
		{"AudioEvent.Sound width", unsafe.Sizeof(AudioEvent{}.Sound), wantEventSoundWidth},
		{"AudioEvent.Pos.X width", unsafe.Sizeof(AudioEvent{}.Pos.X), 4},
	} {
		if check.got != check.want {
			t.Errorf("%s on %s/%s = %d, want %d", check.name, runtime.GOOS, runtime.GOARCH, check.got, check.want)
		}
	}
	if byte(netmsg.MSG_AUDIO_EVENT) != 0xa6 || byte(netmsg.MSG_AUDIO_PLAYER_EVENT) != 0xa7 {
		t.Fatalf("audio opcodes = %#x/%#x, want 0xa6/0xa7", netmsg.MSG_AUDIO_EVENT, netmsg.MSG_AUDIO_PLAYER_EVENT)
	}
	if netlist.Kind(audioEventPacketListKind501FD0) != netlist.Kind1 {
		t.Fatalf("message-list kind = %d, want Kind1 (%d)", audioEventPacketListKind501FD0, netlist.Kind1)
	}
}

func TestSendAudioEventNative501FD0PreservesPointersRoundingAndReload(t *testing.T) {
	list := netlist.New()
	list.Init()
	defer list.Free()

	oldPlayer := &Player{PlayerInd: 5, Pos3632Vec: types.Ptf(10, 20)}
	newPlayer := &Player{PlayerInd: 7, Pos3632Vec: types.Ptf(999, 999)}
	update := &PlayerUpdateData{Player: oldPlayer}
	unit := &Object{UpdateData: unsafe.Pointer(update)}
	soundValue := uint64(1)<<32 | 0x12ab
	if unsafe.Sizeof(sound.ID(0)) == 4 {
		soundValue = 0x12ab
	}
	event := &AudioEvent{
		Sound: sound.ID(soundValue),
		Pos:   types.Ptf(13.5, 77),
		Obj:   unit,
	}
	server := &Server{NetList: list}

	if unsafe.Sizeof(uintptr(0)) == 8 {
		for name, pointer := range map[string]unsafe.Pointer{
			"unit":       unsafe.Pointer(unit),
			"event":      unsafe.Pointer(event),
			"update":     unsafe.Pointer(update),
			"old player": unsafe.Pointer(oldPlayer),
			"new player": unsafe.Pointer(newPlayer),
		} {
			if uintptr(pointer) <= math.MaxUint32 {
				t.Fatalf("%s pointer = %p, want native address above 4 GiB", name, pointer)
			}
		}
	}

	widthCalls := 0
	if got := server.SendAudioEvent501FD0(unit, event, -3, AudioEventPacketRuntime501FD0{
		WindowWidth: func() int32 {
			widthCalls++
			update.Player = newPlayer
			return 2
		},
	}); !got {
		t.Fatal("enqueue result = false, want true")
	}
	if widthCalls != 1 {
		t.Fatalf("window width calls = %d, want 1", widthCalls)
	}
	if got := list.CopyPacketsA(ntype.PlayerInd(oldPlayer.PlayerInd), netlist.Kind1); len(got) != 0 {
		t.Fatalf("old player received packet %x", got)
	}
	want := []byte{0xa7, 0xc8, 0xab, 0xf6}
	if got := list.CopyPacketsA(ntype.PlayerInd(newPlayer.PlayerInd), netlist.Kind1); !reflect.DeepEqual(got, want) {
		t.Fatalf("packet = %x, want %x", got, want)
	}

	runtime.KeepAlive(event)
	runtime.KeepAlive(unit)
	runtime.KeepAlive(update)
	runtime.KeepAlive(oldPlayer)
	runtime.KeepAlive(newPlayer)
}
