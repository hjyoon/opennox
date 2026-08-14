package server

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func TestAudioEventCollide4EAAD0NativeLayout(t *testing.T) {
	wantSize := uintptr(780)
	wantClass := uintptr(8)
	wantFrame := uintptr(136)
	wantData := uintptr(700)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantSize = 928
		wantClass = 12
		wantFrame = 140
		wantData = 776
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantSize},
		{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), wantClass},
		{"Object.Field34", unsafe.Offsetof(Object{}.Field34), wantFrame},
		{"Object.CollideData", unsafe.Offsetof(Object{}.CollideData), wantData},
		{"AudioEventCollideData size", unsafe.Sizeof(AudioEventCollideData{}), 4},
		{"AudioEventCollideData.Sound", unsafe.Offsetof(AudioEventCollideData{}.Sound), 0},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestAudioEventCollideNative4EAAD0DataFaultAfterFrameStore(t *testing.T) {
	source := &Object{Field34: 0}
	target := &Object{ObjClass: object.ClassPlayer}
	deps := audioEventCollideNativeDeps4EAAD0{
		frame: func() uint32 { return 31 },
		audio: func(uint32, *Object) {
			t.Fatal("audio reached")
		},
	}
	defer func() {
		if recover() == nil {
			t.Fatal("nil collide data did not fault")
		}
		if source.Field34 != 31 {
			t.Fatalf("frame = %d, want 31", source.Field34)
		}
	}()
	audioEventCollideNative4EAAD0(source, target, nil, deps)
}

func TestAudioEventCollide4EAAD0ServerBinding(t *testing.T) {
	s := &Server{}
	s.SetFrame(31)
	data := &AudioEventCollideData{Sound: 417}
	source := &Object{CollideData: unsafe.Pointer(data)}
	target := &Object{ObjClass: object.ClassPlayer, Field34: 0x55667788}
	collision := &types.Pointf{X: 3, Y: 4}
	s.AudioEventCollide4EAAD0(source, target, collision)
	if source.Field34 != 31 || target.Field34 != 0x55667788 || collision.X != 3 || collision.Y != 4 {
		t.Fatalf("state = source %#x target %#x collision %+v", source.Field34, target.Field34, collision)
	}
	if len(s.Audio.delayedObj) != 1 {
		t.Fatalf("delayed audio = %d", len(s.Audio.delayedObj))
	}
	event := s.Audio.delayedObj[0]
	if uint32(event.ID) != 417 || event.Obj != source || event.Kind != 0 || event.Code != 0 {
		t.Fatalf("audio event = %+v", event)
	}
}
