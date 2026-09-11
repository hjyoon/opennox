package server

import (
	"reflect"
	"runtime"
	"testing"
	"unsafe"
)

func TestAudioEventFlushNativeLayout502060(t *testing.T) {
	wantDescriptorSize := uintptr(28)
	wantDescriptorLimit := uintptr(20)
	wantDescriptorHead := uintptr(24)
	wantEventNext := uintptr(28)
	wantEventPercentage := uintptr(32)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantDescriptorSize = 40
		wantDescriptorLimit = 24
		wantDescriptorHead = 32
		wantEventNext = 48
		wantEventPercentage = 56
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"bitmap size", unsafe.Sizeof(soundBitmap501E80{}), 0x80},
		{"descriptor count", uintptr(len(serverAudio{}.bySound)), 1024},
		{"descriptor size", unsafe.Sizeof(audioEvent2{}), wantDescriptorSize},
		{"descriptor limit", unsafe.Offsetof(audioEvent2{}.Field20), wantDescriptorLimit},
		{"descriptor head", unsafe.Offsetof(audioEvent2{}.Field24), wantDescriptorHead},
		{"event next", unsafe.Offsetof(AudioEvent{}.list28), wantEventNext},
		{"event percentage", unsafe.Offsetof(AudioEvent{}.Perc), wantEventPercentage},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestFlushAudioEventsNative502060PreservesPointersAndNarrowsDwords(t *testing.T) {
	const soundID = 33
	audio := &serverAudio{}
	listener := &Object{}
	second := &AudioEvent{Perc: 99}
	first := &AudioEvent{Perc: -17, list28: second}
	descriptor := &audio.bySound[soundID]
	descriptor.Field20 = 1
	descriptor.Field24 = first
	audio.bitmap[soundID/32] = uint32(1) << (soundID % 32)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		lowPercentage := int32(-17)
		highLimit := uint64(1) << 32
		highPercentage := uint64(0x12345678) << 32
		descriptor.Field20 = int(highLimit | 1)
		first.Perc = int(highPercentage | uint64(uint32(lowPercentage)))
	}

	type call struct {
		listener   *Object
		event      *AudioEvent
		percentage int32
	}
	var calls []call
	audio.FlushAudioEvents502060(listener, func(gotListener *Object, event *AudioEvent, percentage int32) bool {
		calls = append(calls, call{gotListener, event, percentage})
		return false
	})

	want := []call{{listener, first, -17}}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls = %+v, want %+v", calls, want)
	}
	runtime.KeepAlive(audio)
	runtime.KeepAlive(listener)
	runtime.KeepAlive(first)
	runtime.KeepAlive(second)
}
