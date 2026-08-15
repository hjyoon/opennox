package server

import (
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
)

func defaultSoulGateCollideNativeDeps4EBE40() soulGateCollideNativeDeps4EBE40 {
	return soulGateCollideNativeDeps4EBE40{
		gameFlagsCheck:  func(uint32) uint32 { return 0 },
		setQuestMode:    func(int32) {},
		firstPlayerUnit: func() *Object { return nil },
		nextPlayerUnit:  func(*Object) *Object { return nil },
		loadFrame:       func() uint32 { return 0 },
		setQuestTimer:   func(uint32) {},
		loadFPS:         func() uint32 { return 0 },
		audio:           func(uint32, *Object, int32, int32) {},
		pointFX:         func(uint32, *types.Pointf) uint32 { return 0 },
		priorityMessage: func(*Object, string, int32) {},
	}
}

func TestSoulGateCollide4EBE40NativeLayout(t *testing.T) {
	wantObjectSize := uintptr(780)
	wantClass := uintptr(8)
	wantPosition := uintptr(56)
	wantCollideData := uintptr(700)
	wantUpdateData := uintptr(748)
	wantUpdateSize := uintptr(556)
	wantPlayerPointer := uintptr(276)
	wantSoulGate := uintptr(308)
	wantPlayerSize := uintptr(4828)
	wantQuestState := uintptr(4792)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize = 928
		wantClass = 12
		wantPosition = 60
		wantCollideData = 776
		wantUpdateData = 872
		wantUpdateSize = 640
		wantPlayerPointer = 320
		wantSoulGate = 376
		wantPlayerSize = 6160
		wantQuestState = 6096
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), wantClass},
		{"Object.PosVec", unsafe.Offsetof(Object{}.PosVec), wantPosition},
		{"Object.CollideData", unsafe.Offsetof(Object{}.CollideData), wantCollideData},
		{"Object.UpdateData", unsafe.Offsetof(Object{}.UpdateData), wantUpdateData},
		{"SoulGateCollideData size", unsafe.Sizeof(SoulGateCollideData{}), 4},
		{"SoulGateCollideData.LastUsedFrame", unsafe.Offsetof(SoulGateCollideData{}.LastUsedFrame), 0},
		{"PlayerUpdateData size", unsafe.Sizeof(PlayerUpdateData{}), wantUpdateSize},
		{"PlayerUpdateData.Player", unsafe.Offsetof(PlayerUpdateData{}.Player), wantPlayerPointer},
		{"PlayerUpdateData.SoulGate", unsafe.Offsetof(PlayerUpdateData{}.SoulGate), wantSoulGate},
		{"Player size", unsafe.Sizeof(Player{}), wantPlayerSize},
		{"Player.Field4792", unsafe.Offsetof(Player{}.Field4792), wantQuestState},
		{"Pointf size", unsafe.Sizeof(types.Pointf{}), 8},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestSoulGateCollideNative4EBE40BindsNativeFieldsAndCachesPointers(t *testing.T) {
	entryData := &SoulGateCollideData{LastUsedFrame: 17}
	liveData := &SoulGateCollideData{LastUsedFrame: 99}
	source := &Object{
		PosVec:      types.Pointf{X: 12.5, Y: -7.25},
		CollideData: unsafe.Pointer(entryData),
	}
	targetUpdate := &PlayerUpdateData{}
	replacementUpdate := &PlayerUpdateData{}
	target := &Object{
		ObjClass:   object.ClassPlayer | object.Class(0x40000000),
		UpdateData: unsafe.Pointer(targetUpdate),
	}
	readyUpdate := &PlayerUpdateData{
		Player:   &Player{Field4792: soulGateReadyState4EBE40},
		SoulGate: source,
	}
	ready := &Object{UpdateData: unsafe.Pointer(readyUpdate)}
	collision := &types.Pointf{
		X: math.Float32frombits(0x7fc12345),
		Y: math.Float32frombits(0x80000000),
	}
	events := make([]string, 0, 10)

	deps := defaultSoulGateCollideNativeDeps4EBE40()
	deps.gameFlagsCheck = func(flag uint32) uint32 {
		events = append(events, "game")
		if flag != soulGateQuestFlag4EBE40 {
			t.Fatalf("game flag = %#x", flag)
		}
		return 0x1000
	}
	deps.setQuestMode = func(value int32) {
		events = append(events, "quest-mode")
		if value != 0 {
			t.Fatalf("Quest mode = %d", value)
		}
		source.CollideData = unsafe.Pointer(liveData)
	}
	deps.firstPlayerUnit = func() *Object {
		events = append(events, "first")
		return ready
	}
	deps.nextPlayerUnit = func(got *Object) *Object {
		events = append(events, "next")
		if got != ready {
			t.Fatalf("next input = %p, want ready %p", got, ready)
		}
		return nil
	}
	deps.setQuestTimer = func(uint32) {
		t.Fatal("ready SoulGate unexpectedly refreshed Quest timer")
	}
	deps.loadFPS = func() uint32 {
		t.Fatal("different target gate unexpectedly read FPS")
		return 0
	}
	deps.audio = func(id uint32, got *Object, first, second int32) {
		events = append(events, "audio")
		if id != soulGateAudio4EBE40 || got != source || first != 0 || second != 0 {
			t.Fatalf("audio = %d/%p/%d/%d", id, got, first, second)
		}
		target.UpdateData = unsafe.Pointer(replacementUpdate)
	}
	deps.pointFX = func(id uint32, position *types.Pointf) uint32 {
		events = append(events, "fx")
		if id != soulGatePointFX4EBE40 || position != &source.PosVec {
			t.Fatalf("point FX = %d/%p, want %d/%p", id, position, soulGatePointFX4EBE40, &source.PosVec)
		}
		position.Y = 44
		return 0xf1234567
	}
	deps.priorityMessage = func(got *Object, message string, value int32) {
		events = append(events, "message")
		if got != target || message != soulGatePriorityMessage4EBE40 || value != 0 {
			t.Fatalf("message = %p/%q/%d", got, message, value)
		}
	}
	deps.loadFrame = func() uint32 {
		events = append(events, "frame")
		if targetUpdate.SoulGate != source {
			t.Fatal("final frame read preceded target SoulGate store")
		}
		return 1234
	}

	soulGateCollideNative4EBE40(source, target, collision, deps)
	if entryData.LastUsedFrame != 1234 || liveData.LastUsedFrame != 99 {
		t.Fatalf("entry/live last frame = %d/%d, want 1234/99", entryData.LastUsedFrame, liveData.LastUsedFrame)
	}
	if targetUpdate.SoulGate != source || replacementUpdate.SoulGate != nil {
		t.Fatalf("cached/replacement target gate = %p/%p, want source/nil", targetUpdate.SoulGate, replacementUpdate.SoulGate)
	}
	if source.PosVec != (types.Pointf{X: 12.5, Y: 44}) {
		t.Fatalf("source position = %v, want native PointFX mutation", source.PosVec)
	}
	if math.Float32bits(collision.X) != 0x7fc12345 || math.Float32bits(collision.Y) != 0x80000000 {
		t.Fatalf("collision changed: %#v", *collision)
	}
	wantEvents := []string{"game", "quest-mode", "first", "next", "audio", "fx", "message", "frame"}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %#v, want %#v", events, wantEvents)
	}
}

func TestSoulGateCollide4EBE40ServerDepsBindFrameRateAndRuntime(t *testing.T) {
	srv := new(Server)
	srv.SetFrame(0xfedcba98)
	srv.SetTickRate(73)
	events := make([]string, 0, 3)
	position := types.Pointf{X: 1, Y: 2}
	runtime := SoulGateCollideRuntime4EBE40{
		SetQuestMode: func(value int32) {
			events = append(events, "mode")
			if value != -7 {
				t.Fatalf("mode = %d", value)
			}
		},
		SetQuestTimer: func(frame uint32) {
			events = append(events, "timer")
			if frame != 0x87654321 {
				t.Fatalf("timer = %#x", frame)
			}
		},
		PointFX: func(id uint32, got *types.Pointf) uint32 {
			events = append(events, "fx")
			if id != 55 || got != &position {
				t.Fatalf("point FX = %d/%p", id, got)
			}
			return 0x89abcdef
		},
	}
	deps := soulGateCollideServerDeps4EBE40(srv, runtime)
	if got := deps.loadFrame(); got != 0xfedcba98 {
		t.Fatalf("frame = %#x", got)
	}
	if got := deps.loadFPS(); got != 73 {
		t.Fatalf("FPS = %d", got)
	}
	if got, want := deps.gameFlagsCheck(soulGateQuestFlag4EBE40) != 0, noxflags.HasGame(noxflags.GameModeQuest); got != want {
		t.Fatalf("Quest flag check = %v, want %v", got, want)
	}
	deps.setQuestMode(-7)
	deps.setQuestTimer(0x87654321)
	if got := deps.pointFX(55, &position); got != 0x89abcdef {
		t.Fatalf("PointFX return = %#x", got)
	}
	if !reflect.DeepEqual(events, []string{"mode", "timer", "fx"}) {
		t.Fatalf("runtime events = %#v", events)
	}
}
