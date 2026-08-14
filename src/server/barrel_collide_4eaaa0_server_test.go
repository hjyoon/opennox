package server

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"
)

func defaultBarrelCollideNativeDeps4EAAA0() barrelCollideNativeDeps4EAAA0 {
	return barrelCollideNativeDeps4EAAA0{
		frame: func() uint32 { return 0 },
		audio: func(uint32, *Object) {},
	}
}

func TestBarrelCollide4EAAA0NativeLayout(t *testing.T) {
	wantFrame := uintptr(136)
	wantSize := uintptr(780)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantFrame = 140
		wantSize = 928
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantSize},
		{"Object.Field34", unsafe.Offsetof(Object{}.Field34), wantFrame},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestBarrelCollideNative4EAAA0UsesNativeFieldAndOrder(t *testing.T) {
	source := &Object{Field34: 5}
	target := &Object{Field34: 0x11223344}
	collision := &types.Pointf{X: 1, Y: 2}
	events := make([]string, 0, 2)
	deps := defaultBarrelCollideNativeDeps4EAAA0()
	deps.frame = func() uint32 {
		events = append(events, "frame")
		return 9
	}
	deps.audio = func(id uint32, obj *Object) {
		events = append(events, "audio")
		if id != barrelCollideSound4EAAA0 || obj != source || obj.Field34 != 9 {
			t.Fatalf("audio = %d/%p frame %#x", id, obj, obj.Field34)
		}
	}

	barrelCollideNative4EAAA0(source, target, collision, deps)
	if want := []string{"frame", "audio"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
	if source.Field34 != 9 || target.Field34 != 0x11223344 || collision.X != 1 || collision.Y != 2 {
		t.Fatalf("state = source %#x target %#x collision %+v", source.Field34, target.Field34, collision)
	}
}

func TestBarrelCollide4EAAA0ServerBinding(t *testing.T) {
	s := &Server{}
	s.SetFrame(9)
	source := &Object{Field34: 5}
	target := &Object{Field34: 0x55667788}
	collision := &types.Pointf{X: 3, Y: 4}
	s.BarrelCollide4EAAA0(source, target, collision)
	if source.Field34 != 9 || target.Field34 != 0x55667788 || collision.X != 3 || collision.Y != 4 {
		t.Fatalf("state = source %#x target %#x collision %+v", source.Field34, target.Field34, collision)
	}
	if len(s.Audio.delayedObj) != 1 {
		t.Fatalf("delayed audio = %d", len(s.Audio.delayedObj))
	}
	event := s.Audio.delayedObj[0]
	if uint32(event.ID) != barrelCollideSound4EAAA0 || event.Obj != source || event.Kind != 0 || event.Code != 0 {
		t.Fatalf("audio event = %+v", event)
	}
}
