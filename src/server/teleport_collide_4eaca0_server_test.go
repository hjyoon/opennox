package server

import (
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func TestTeleportCollide4EACA0NativeLayout(t *testing.T) {
	if got := unsafe.Sizeof(TeleportCollideData{}); got != 8 {
		t.Fatalf("TeleportCollideData size = %d, want 8", got)
	}
	if got := unsafe.Offsetof(TeleportCollideData{}.DestinationX); got != 0 {
		t.Fatalf("DestinationX offset = %d, want 0", got)
	}
	if got := unsafe.Offsetof(TeleportCollideData{}.DestinationY); got != 4 {
		t.Fatalf("DestinationY offset = %d, want 4", got)
	}
	if got := unsafe.Sizeof(types.Pointf{}); got != 8 {
		t.Fatalf("Pointf size = %d, want 8", got)
	}
	if got := unsafe.Offsetof(Object{}.Field42) - unsafe.Offsetof(Object{}.Field41); got != 4 {
		t.Fatalf("Field42-Field41 = %d, want 4", got)
	}

	type objectLayout struct {
		size        uintptr
		class       uintptr
		position    uintptr
		field41     uintptr
		field42     uintptr
		collideData uintptr
	}
	wantByPointerSize := map[uintptr]objectLayout{
		4: {size: 780, class: 8, position: 56, field41: 164, field42: 168, collideData: 700},
		8: {size: 928, class: 12, position: 60, field41: 168, field42: 172, collideData: 776},
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
		field41:     unsafe.Offsetof(Object{}.Field41),
		field42:     unsafe.Offsetof(Object{}.Field42),
		collideData: unsafe.Offsetof(Object{}.CollideData),
	}
	if got != want {
		t.Fatalf("Object layout = %+v, want %+v", got, want)
	}
}

func TestTeleportCollideNative4EACA0UsesCachedPointersAndLiveData(t *testing.T) {
	data := &TeleportCollideData{DestinationX: 1, DestinationY: 2}
	replacement := &TeleportCollideData{DestinationX: 7, DestinationY: 8}
	source := &Object{CollideData: unsafe.Pointer(data)}
	target := &Object{
		ObjClass: object.ClassPlayer | object.Class(0x80000000),
		PosVec:   types.Ptf(10.5, -20.25),
	}
	collision := &types.Pointf{X: 99, Y: -99}
	entryPosition := target.PosVec
	var events []string
	audioCount := 0
	positionPointer := &target.PosVec
	destinationPointer := teleportCollideDestination4EACA0(target)

	teleportCollideNative4EACA0(source, target, collision, teleportCollideNativeDeps4EACA0{
		pointFX: func(id uint32, pos *types.Pointf) {
			events = append(events, "fx")
			if pos != positionPointer {
				t.Fatalf("point position pointer = %p, want cached %p", pos, positionPointer)
			}
			switch id {
			case teleportCollidePreFX4EACA0:
				if *pos != entryPosition {
					t.Fatalf("pre position = %#v, want %#v", *pos, entryPosition)
				}
				source.CollideData = unsafe.Pointer(replacement)
				data.DestinationX = 1<<24 + 1
				target.ObjClass |= object.Class(teleportCollideDoorClass4EACA0)
			case teleportCollidePostFX4EACA0:
				want := types.Ptf(float32(1<<24+1), float32(math.MinInt32))
				if *pos != want {
					t.Fatalf("post position = %#v, want %#v", *pos, want)
				}
			default:
				t.Fatalf("point FX = %d", id)
			}
		},
		audio: func(id uint32, obj *Object) {
			events = append(events, "audio")
			if id != teleportCollideSound4EACA0 || obj != target {
				t.Fatalf("audio = %d/%p", id, obj)
			}
			audioCount++
			if audioCount == 1 {
				data.DestinationY = math.MinInt32
			}
		},
		teleport: func(obj *Object, destination *types.Pointf) {
			events = append(events, "teleport")
			if obj != target || destination != destinationPointer {
				t.Fatalf("teleport = %p/%p, want %p/%p", obj, destination, target, destinationPointer)
			}
			want := types.Ptf(float32(1<<24+1), float32(math.MinInt32))
			if *destination != want {
				t.Fatalf("destination = %#v, want %#v", *destination, want)
			}
			obj.PosVec = *destination
		},
	})

	if want := []string{"fx", "audio", "teleport", "fx", "audio"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %#v, want %#v", events, want)
	}
	if audioCount != 2 {
		t.Fatalf("audio count = %d", audioCount)
	}
	if got, want := target.Field41, math.Float32bits(float32(1<<24+1)); got != want {
		t.Fatalf("Field41 bits = %#08x, want %#08x", got, want)
	}
	if got, want := target.Field42, math.Float32bits(float32(math.MinInt32)); got != want {
		t.Fatalf("Field42 bits = %#08x, want %#08x", got, want)
	}
}

func TestTeleportCollideNative4EACA0SignedConversionBits(t *testing.T) {
	values := []int32{0, 1, -1, 1<<24 + 1, -(1<<24 + 1), math.MaxInt32, math.MinInt32}
	for _, x := range values {
		for _, y := range values {
			data := &TeleportCollideData{DestinationX: x, DestinationY: y}
			source := &Object{CollideData: unsafe.Pointer(data)}
			target := &Object{}
			teleportCollideNative4EACA0(source, target, nil, teleportCollideNativeDeps4EACA0{
				pointFX: func(uint32, *types.Pointf) {},
				audio:   func(uint32, *Object) {},
				teleport: func(obj *Object, destination *types.Pointf) {
					if got, want := math.Float32bits(destination.X), math.Float32bits(float32(x)); got != want {
						t.Fatalf("x=%d: bits = %#08x, want %#08x", x, got, want)
					}
					if got, want := math.Float32bits(destination.Y), math.Float32bits(float32(y)); got != want {
						t.Fatalf("y=%d: bits = %#08x, want %#08x", y, got, want)
					}
				},
			})
		}
	}
}

func TestTeleportCollideNative4EACA0NilDataFaultOrder(t *testing.T) {
	source := &Object{}
	target := &Object{}
	var events []string
	teleported := false
	var recovered any
	func() {
		defer func() { recovered = recover() }()
		teleportCollideNative4EACA0(source, target, nil, teleportCollideNativeDeps4EACA0{
			pointFX: func(uint32, *types.Pointf) { events = append(events, "fx") },
			audio:   func(uint32, *Object) { events = append(events, "audio") },
			teleport: func(*Object, *types.Pointf) {
				teleported = true
			},
		})
	}()
	if recovered == nil {
		t.Fatal("nil collide data did not fault")
	}
	if want := []string{"fx", "audio"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %#v, want %#v", events, want)
	}
	if teleported {
		t.Fatal("teleport ran after nil-data fault")
	}
}

func TestTeleportCollide4EACA0ServerBindingTargetGuards(t *testing.T) {
	source := &Object{}
	var srv Server
	srv.TeleportCollide4EACA0(source, nil, nil, TeleportCollideRuntime4EACA0{})
	srv.TeleportCollide4EACA0(
		source,
		&Object{ObjClass: object.Class(teleportCollideDoorClass4EACA0)},
		nil,
		TeleportCollideRuntime4EACA0{},
	)
}
