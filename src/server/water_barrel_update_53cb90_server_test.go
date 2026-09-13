package server

import (
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/sound"
)

func TestWaterBarrelUpdate53CB90NativeLayout(t *testing.T) {
	wantClass := uintptr(8)
	wantPosition := uintptr(56)
	wantCreated := uintptr(128)
	wantRadius := uintptr(176)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantClass = 12
		wantPosition = 60
		wantCreated = 132
		wantRadius = 180
	}
	for _, field := range []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"class", unsafe.Offsetof(Object{}.ObjClass), wantClass},
		{"position", unsafe.Offsetof(Object{}.PosVec), wantPosition},
		{"creation frame", unsafe.Offsetof(Object{}.Field32), wantCreated},
		{"circle radius", unsafe.Offsetof(Object{}.Shape) + unsafe.Offsetof(Shape{}.Circle) + unsafe.Offsetof(Shape{}.Circle.R), wantRadius},
	} {
		if field.got != field.want {
			t.Errorf("%s offset = %d, want %d", field.name, field.got, field.want)
		}
	}
	if waterBarrelAudioID53CB90 != sound.ID(283) {
		t.Fatalf("water barrel audio = %d, want 283", waterBarrelAudioID53CB90)
	}
}

func TestWaterBarrelUpdate53CB90AgeAndWrap(t *testing.T) {
	tests := []struct {
		name           string
		frame, created uint32
		wantRect       bool
		wantDelete     bool
	}{
		{"before burst", 107, 100, false, false},
		{"burst", 108, 100, true, false},
		{"after burst", 109, 100, false, false},
		{"before expiry", 129, 100, false, false},
		{"expiry", 130, 100, false, true},
		{"wrapped burst", 3, math.MaxUint32 - 4, true, false},
		{"wrapped expiry", 3, math.MaxUint32 - 26, false, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			source := &Object{Field32: tc.created}
			var rectCalls, deletes, audio int
			waterBarrelUpdateNative53CB90(source, waterBarrelUpdateNativeDeps53CB90{
				frame: func() uint32 { return tc.frame },
				eachInRect: func(types.Rectf, func(*Object) bool) {
					rectCalls++
				},
				delayedDelete: func(got *Object) {
					if got != source {
						t.Fatalf("deleted %p, want source %p", got, source)
					}
					deletes++
				},
				audio: func(sound.ID, *Object) { audio++ },
			})
			if (rectCalls != 0) != tc.wantRect || (deletes != 0) != tc.wantDelete || audio != 0 {
				t.Fatalf("rect/deletes/audio = %d/%d/%d, want rect=%t delete=%t", rectCalls, deletes, audio, tc.wantRect, tc.wantDelete)
			}
		})
	}
}

func TestWaterBarrelUpdate53CB90FireFilterRectAndAudioOrder(t *testing.T) {
	source := &Object{Field32: 100, PosVec: types.Ptf(100, 200)}
	fire := &Object{ObjClass: object.ClassFire, PosVec: types.Ptf(100, 200)}
	boundary := &Object{ObjClass: object.ClassFire, PosVec: types.Ptf(145, 200)}
	boundary.Shape.Circle.R = 5
	outside := &Object{ObjClass: object.ClassFire, PosVec: types.Ptf(145.25, 200)}
	outside.Shape.Circle.R = 5
	nonFire := &Object{ObjClass: object.ClassMissile, PosVec: types.Ptf(100, 200)}
	candidates := []*Object{nonFire, fire, outside, boundary}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, obj := range append(candidates, source) {
			if uintptr(unsafe.Pointer(obj)) <= math.MaxUint32 {
				t.Fatalf("native object pointer = %p, want above ABI32 range", obj)
			}
		}
	}
	var events []string
	wantRect := types.Rectf{Min: types.Ptf(60, 160), Max: types.Ptf(140, 240)}
	waterBarrelUpdateNative53CB90(source, waterBarrelUpdateNativeDeps53CB90{
		frame: func() uint32 {
			events = append(events, "frame")
			return 108
		},
		eachInRect: func(rect types.Rectf, fnc func(*Object) bool) {
			if rect != wantRect {
				t.Fatalf("rect = %+v, want %+v", rect, wantRect)
			}
			events = append(events, "rect")
			for _, candidate := range candidates {
				if !fnc(candidate) {
					t.Fatal("callback stopped enumeration")
				}
			}
		},
		delayedDelete: func(got *Object) {
			switch got {
			case fire:
				events = append(events, "fire")
			case boundary:
				events = append(events, "boundary")
			default:
				t.Fatalf("unexpected deletion: %p", got)
			}
		},
		audio: func(id sound.ID, got *Object) {
			if id != waterBarrelAudioID53CB90 || got != source {
				t.Fatalf("audio = %d/%p, want %d/%p", id, got, waterBarrelAudioID53CB90, source)
			}
			events = append(events, "audio")
		},
	})
	if want := []string{"frame", "rect", "fire", "boundary", "audio"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestWaterBarrelUpdate53CB90LoadsLiveSourcePositionDuringEnumeration(t *testing.T) {
	source := &Object{Field32: 100, PosVec: types.Ptf(0, 0)}
	first := &Object{ObjClass: object.ClassFire, PosVec: types.Ptf(0, 0)}
	second := &Object{ObjClass: object.ClassFire, PosVec: types.Ptf(1, 0)}
	var deleted []*Object
	waterBarrelUpdateNative53CB90(source, waterBarrelUpdateNativeDeps53CB90{
		frame: func() uint32 { return 108 },
		eachInRect: func(_ types.Rectf, fnc func(*Object) bool) {
			fnc(first)
			fnc(second)
		},
		delayedDelete: func(got *Object) {
			deleted = append(deleted, got)
			source.PosVec = types.Ptf(100, 100)
		},
		audio: func(sound.ID, *Object) {},
	})
	if !reflect.DeepEqual(deleted, []*Object{first}) {
		t.Fatalf("deleted = %v, want only %p", deleted, first)
	}
}
