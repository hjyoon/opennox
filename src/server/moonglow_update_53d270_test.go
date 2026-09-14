package server

import (
	"image"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func TestMoonglowUpdate53D270MissingAndNonPlayerOwner(t *testing.T) {
	visual := &Object{}
	deleted := 0
	runtime := MoonglowUpdateRuntime53D270{
		DelayedDelete: func(got *Object) {
			deleted++
			if got != visual {
				t.Errorf("deleted = %p, want %p", got, visual)
			}
		},
	}
	MoonglowUpdate53D270(visual, runtime)
	if deleted != 1 {
		t.Fatalf("missing-owner deletes = %d, want 1", deleted)
	}
	visual.ObjOwner = &Object{ObjClass: object.ClassMonster}
	MoonglowUpdate53D270(visual, runtime)
	if deleted != 1 {
		t.Fatalf("non-player deletes = %d, want 0 more", deleted)
	}
}

func TestMoonglowUpdate53D270PlayerMovementAndExpiry(t *testing.T) {
	player := &Player{CursorVec: image.Point{X: -321, Y: 654}}
	update := &PlayerUpdateData{Player: player}
	owner := &Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update)}
	visual := &Object{ObjOwner: owner, Field32: 100}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(owner)) <= 0xffffffff {
		t.Fatal("expected owner pointer above 4 GiB")
	}
	frame := uint32(110)
	valid := true
	var events []string
	runtime := MoonglowUpdateRuntime53D270{
		Frame: func() uint32 { return frame },
		FPS:   func() uint32 { return 30 },
		ValidPosition: func(point types.Pointf) bool {
			events = append(events, "valid")
			if point != (types.Pointf{X: -321, Y: 654}) {
				t.Errorf("valid position = %v", point)
			}
			return valid
		},
		Move: func(got *Object, point types.Pointf) {
			events = append(events, "move")
			if got != visual || point != (types.Pointf{X: -321, Y: 654}) {
				t.Errorf("move = %p/%v", got, point)
			}
		},
		DelayedDelete: func(got *Object) {
			events = append(events, "delete")
			if got != visual {
				t.Errorf("delete = %p", got)
			}
		},
		BuffOff: func(got *Object, buff EnchantID) {
			events = append(events, "off")
			if got != owner || buff != ENCHANT_MOONGLOW {
				t.Errorf("buff-off = %p/%d", got, buff)
			}
		},
	}
	MoonglowUpdate53D270(visual, runtime)
	valid = false
	MoonglowUpdate53D270(visual, runtime)
	frame = 100 + 300*30
	MoonglowUpdate53D270(visual, runtime)
	frame++
	MoonglowUpdate53D270(visual, runtime)
	want := []string{"valid", "move", "valid", "valid", "delete", "off"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestMoonglowUpdate53D270DestroyedOwnerShortCircuitsTime(t *testing.T) {
	owner := &Object{ObjClass: object.ClassPlayer, ObjFlags: object.FlagDestroyed}
	visual := &Object{ObjOwner: owner}
	var events []string
	MoonglowUpdate53D270(visual, MoonglowUpdateRuntime53D270{
		DelayedDelete: func(*Object) { events = append(events, "delete") },
		BuffOff: func(*Object, EnchantID) {
			events = append(events, "off")
		},
	})
	if !reflect.DeepEqual(events, []string{"delete", "off"}) {
		t.Fatalf("events = %v", events)
	}
}
