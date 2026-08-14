package server

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func TestTeleportWakeCollide4EAE30NativeLayout(t *testing.T) {
	if got := unsafe.Sizeof(TeleportWakeCollideData{}); got != 8 {
		t.Fatalf("TeleportWakeCollideData size = %d, want 8", got)
	}
	if got := unsafe.Offsetof(TeleportWakeCollideData{}.Destination); got != 0 {
		t.Fatalf("Destination offset = %d, want 0", got)
	}
	if got := unsafe.Sizeof(types.Pointf{}); got != 8 {
		t.Fatalf("Pointf size = %d, want 8", got)
	}
	if got := unsafe.Offsetof(types.Pointf{}.X); got != 0 {
		t.Fatalf("Pointf.X offset = %d, want 0", got)
	}
	if got := unsafe.Offsetof(types.Pointf{}.Y); got != 4 {
		t.Fatalf("Pointf.Y offset = %d, want 4", got)
	}

	type objectLayout struct {
		size        uintptr
		class       uintptr
		position    uintptr
		buffs       uintptr
		owner       uintptr
		collideData uintptr
	}
	wantByPointerSize := map[uintptr]objectLayout{
		4: {size: 780, class: 8, position: 56, buffs: 340, owner: 508, collideData: 700},
		8: {size: 928, class: 12, position: 60, buffs: 344, owner: 552, collideData: 776},
	}
	pointerSize := unsafe.Sizeof(uintptr(0))
	want, ok := wantByPointerSize[pointerSize]
	if !ok {
		t.Fatalf("unsupported pointer size %d", pointerSize)
	}
	got := objectLayout{
		size:        unsafe.Sizeof(Object{}),
		class:       unsafe.Offsetof(Object{}.ObjClass),
		position:    unsafe.Offsetof(Object{}.PosVec),
		buffs:       unsafe.Offsetof(Object{}.Buffs),
		owner:       unsafe.Offsetof(Object{}.ObjOwner),
		collideData: unsafe.Offsetof(Object{}.CollideData),
	}
	if got != want {
		t.Fatalf("Object layout = %+v, want %+v", got, want)
	}
}

func TestTeleportWakeCollideNative4EAE30CachedDestinationAndLivePosition(t *testing.T) {
	data := &TeleportWakeCollideData{Destination: types.Ptf(10, -20)}
	replacement := &TeleportWakeCollideData{Destination: types.Ptf(30, -40)}
	oldOwner := &Object{ObjClass: object.ClassMonster}
	playerOwner := &Object{ObjClass: object.ClassPlayer | object.Class(0x91020000)}
	source := &Object{CollideData: unsafe.Pointer(data), ObjOwner: oldOwner}
	target := &Object{PosVec: types.Ptf(1.5, -2.5)}
	collision := &types.Pointf{X: 123, Y: -456}
	wantCollision := *collision
	entryPosition := target.PosVec
	postPosition := types.Ptf(-123.5, 456.25)
	destinationPointer := (*types.Pointf)(unsafe.Pointer(data))

	var events []string
	invisible := false
	invisibleCalls := 0
	audioCalls := 0

	teleportWakeCollideNative4EAE30(source, target, collision, teleportWakeCollideNativeDeps4EAE30{
		hasEnchant: func(obj *Object, enchant EnchantID) bool {
			if obj != target {
				t.Fatalf("enchant object = %p, want %p", obj, target)
			}
			switch enchant {
			case ENCHANT_ANCHORED:
				events = append(events, "anchored")
				return false
			case ENCHANT_INVISIBLE:
				invisibleCalls++
				events = append(events, "invisible")
				return invisible
			default:
				t.Fatalf("enchant = %d", enchant)
				return false
			}
		},
		questMode: func() bool {
			events = append(events, "quest")
			source.ObjOwner = playerOwner
			target.ObjClass = object.ClassPlayer
			return true
		},
		pointFX: func(id uint32, pos *types.Pointf) {
			events = append(events, "fx")
			if pos != &target.PosVec {
				t.Fatalf("position pointer = %p, want live field %p", pos, &target.PosVec)
			}
			switch id {
			case teleportWakePreFX4EAE30:
				if *pos != entryPosition {
					t.Fatalf("pre position = %+v, want %+v", *pos, entryPosition)
				}
				source.CollideData = unsafe.Pointer(replacement)
				target.ObjClass = 0
				invisible = true
			case teleportWakePostFX4EAE30:
				if *pos != postPosition {
					t.Fatalf("post position = %+v, want live %+v", *pos, postPosition)
				}
			default:
				t.Fatalf("point FX = %d", id)
			}
		},
		audio: func(id uint32, obj *Object) {
			events = append(events, "audio")
			if id != teleportWakeSound4EAE30 || obj != target {
				t.Fatalf("audio = %d/%p", id, obj)
			}
			audioCalls++
			if audioCalls == 1 {
				data.Destination = types.Ptf(-123.5, 456.25)
			}
		},
		teleport: func(obj *Object, destination *types.Pointf) {
			events = append(events, "teleport")
			if obj != target || destination != destinationPointer {
				t.Fatalf("teleport = %p/%p, want target/cached %p/%p", obj, destination, target, destinationPointer)
			}
			if *destination != postPosition {
				t.Fatalf("live destination = %+v, want %+v", *destination, postPosition)
			}
			obj.PosVec = *destination
			invisible = false
		},
	})

	wantEvents := []string{
		"anchored", "quest", "invisible", "fx", "audio", "teleport",
		"invisible", "fx", "audio",
	}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %#v, want %#v", events, wantEvents)
	}
	if invisibleCalls != 2 || audioCalls != 2 {
		t.Fatalf("invisible/audio calls = %d/%d, want 2/2", invisibleCalls, audioCalls)
	}
	if source.CollideData != unsafe.Pointer(replacement) {
		t.Fatal("source collide-data mutation was overwritten")
	}
	if *collision != wantCollision {
		t.Fatalf("collision = %+v, want unchanged %+v", *collision, wantCollision)
	}
}

func TestTeleportWakeCollideNative4EAE30NilDestinationPassedWithoutEarlyDereference(t *testing.T) {
	source := &Object{}
	target := &Object{ObjClass: object.ClassPlayer}
	teleportCalls := 0
	teleportWakeCollideNative4EAE30(source, target, nil, teleportWakeCollideNativeDeps4EAE30{
		hasEnchant: func(*Object, EnchantID) bool { return false },
		questMode:  func() bool { return false },
		pointFX:    func(uint32, *types.Pointf) {},
		audio:      func(uint32, *Object) {},
		teleport: func(obj *Object, destination *types.Pointf) {
			teleportCalls++
			if obj != target || destination != nil {
				t.Fatalf("teleport = %p/%p, want %p/nil", obj, destination, target)
			}
		},
	})
	if teleportCalls != 1 {
		t.Fatalf("teleport calls = %d, want 1", teleportCalls)
	}
}

func TestTeleportWakeCollideNative4EAE30EntryFaultAndNilTarget(t *testing.T) {
	t.Run("nil source faults before nil target", func(t *testing.T) {
		var recovered any
		func() {
			defer func() { recovered = recover() }()
			teleportWakeCollideNative4EAE30(nil, nil, nil, teleportWakeCollideNativeDeps4EAE30{})
		}()
		if recovered == nil {
			t.Fatal("nil source did not fault on entry collide-data load")
		}
	})

	t.Run("nil target does not use cached nil destination", func(t *testing.T) {
		teleportWakeCollideNative4EAE30(&Object{}, nil, nil, teleportWakeCollideNativeDeps4EAE30{
			hasEnchant: func(*Object, EnchantID) bool {
				t.Fatal("enchant queried for nil target")
				return false
			},
		})
	})
}
