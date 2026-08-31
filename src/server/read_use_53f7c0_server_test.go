package server

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func TestReadUseNative53F7C0LayoutAndPointerIdentity(t *testing.T) {
	wantObjectSize := uintptr(780)
	wantObjClass := uintptr(8)
	wantUseData := uintptr(736)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantObjectSize = 928
		wantObjClass = 12
		wantUseData = 848
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantObjectSize},
		{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), wantObjClass},
		{"Object.UseData", unsafe.Offsetof(Object{}.UseData), wantUseData},
		{"ReadableUseData size", unsafe.Sizeof(ReadableUseData{}), 260},
		{"ReadableUseData.Text", unsafe.Offsetof(ReadableUseData{}.Text), 0},
		{"ReadableUseData.TransientReadState", unsafe.Offsetof(ReadableUseData{}.TransientReadState), 256},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}

	owner := &Object{ObjClass: object.ClassPlayer}
	data := &ReadableUseData{}
	copy(data.Text[:], "JournalEntry\x00")
	readable := &Object{}
	readable.UseData.SetPtr(unsafe.Pointer(data))

	frame := uint32(0x10203040)
	var (
		mapOwner     *Object
		mapReadable  *Object
		messageOwner *Object
		messageData  *ReadableUseData
		messageValue uint8
	)
	deps := readUseNativeDeps53F7C0{
		loadFPS:   func() uint32 { return 30 },
		loadFrame: func() uint32 { return frame },
		mapCheck: func(gotOwner, gotReadable *Object) int32 {
			mapOwner, mapReadable = gotOwner, gotReadable
			return 1
		},
		primaryMessage: func(gotOwner *Object, gotData *ReadableUseData, value uint8) {
			messageOwner, messageData, messageValue = gotOwner, gotData, value
			frame = 0x89abcdef
		},
	}
	if got := readUseNative53F7C0(owner, readable, deps); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if mapOwner != owner || mapReadable != readable {
		t.Fatalf("map args = %p/%p, want %p/%p", mapOwner, mapReadable, owner, readable)
	}
	if messageOwner != owner || messageData != data || messageValue != 1 {
		t.Fatalf("message args = %p/%p/%d, want %p/%p/1", messageOwner, messageData, messageValue, owner, data)
	}
	if data.TransientReadState != 0x89abcdef || data.TextString() != "JournalEntry" {
		t.Fatalf("data state/text = %#x/%q", data.TransientReadState, data.TextString())
	}

	if unsafe.Sizeof(uintptr(0)) == 8 {
		for name, ptr := range map[string]uintptr{
			"owner":    uintptr(unsafe.Pointer(owner)),
			"readable": uintptr(unsafe.Pointer(readable)),
			"data":     uintptr(unsafe.Pointer(data)),
		} {
			if ptr <= math.MaxUint32 {
				t.Fatalf("%s pointer = %#x, want native address above 4 GiB", name, ptr)
			}
		}
	}
	runtime.KeepAlive(owner)
	runtime.KeepAlive(readable)
	runtime.KeepAlive(data)
}

func TestReadUseNative53F7C0NonPlayerNeedsNoServerState(t *testing.T) {
	s := &Server{}
	if got := s.ReadUse53F7C0(&Object{ObjClass: object.ClassImmobile}, nil); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
}
