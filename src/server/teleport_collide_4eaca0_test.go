package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type teleportCollideTestData4EACA0 struct {
	x int32
	y int32
}

type teleportCollideTestPosition4EACA0 struct {
	name string
}

type teleportCollideTestObject4EACA0 struct {
	class       uint32
	data        *teleportCollideTestData4EACA0
	position    *teleportCollideTestPosition4EACA0
	destination *teleportCollideTestPosition4EACA0
	storedX     float32
	storedY     float32
}

func TestTeleportCollide4EACA0CachesDataBeforeTargetGuards(t *testing.T) {
	t.Run("nil source faults before nil target branch", func(t *testing.T) {
		var events []string
		hooks := teleportCollideHooks4EACA0[
			*teleportCollideTestObject4EACA0,
			*teleportCollideTestData4EACA0,
			*teleportCollideTestPosition4EACA0,
			*teleportCollideTestPosition4EACA0,
		]{
			loadCollideData: func(obj *teleportCollideTestObject4EACA0) *teleportCollideTestData4EACA0 {
				events = append(events, "data")
				return obj.data
			},
		}

		var recovered any
		func() {
			defer func() { recovered = recover() }()
			teleportCollide4EACA0(
				(*teleportCollideTestObject4EACA0)(nil),
				(*teleportCollideTestObject4EACA0)(nil),
				struct{}{}, hooks,
			)
		}()
		if recovered == nil {
			t.Fatal("nil source did not fault on collide-data load")
		}
		if !reflect.DeepEqual(events, []string{"data"}) {
			t.Fatalf("events = %#v", events)
		}
	})

	t.Run("nil target returns with nil data still cached", func(t *testing.T) {
		source := &teleportCollideTestObject4EACA0{}
		var events []string
		hooks := teleportCollideHooks4EACA0[
			*teleportCollideTestObject4EACA0,
			*teleportCollideTestData4EACA0,
			*teleportCollideTestPosition4EACA0,
			*teleportCollideTestPosition4EACA0,
		]{
			loadCollideData: func(obj *teleportCollideTestObject4EACA0) *teleportCollideTestData4EACA0 {
				events = append(events, "data")
				return obj.data
			},
			loadClass: func(*teleportCollideTestObject4EACA0) uint32 {
				t.Fatal("class read for nil target")
				return 0
			},
		}

		teleportCollide4EACA0(
			source,
			(*teleportCollideTestObject4EACA0)(nil),
			struct{}{}, hooks,
		)
		if !reflect.DeepEqual(events, []string{"data"}) {
			t.Fatalf("events = %#v", events)
		}
	})

	t.Run("low-byte bit 0x80 rejects before position", func(t *testing.T) {
		source := &teleportCollideTestObject4EACA0{}
		target := &teleportCollideTestObject4EACA0{class: 0x12345680}
		var events []string
		hooks := teleportCollideHooks4EACA0[
			*teleportCollideTestObject4EACA0,
			*teleportCollideTestData4EACA0,
			*teleportCollideTestPosition4EACA0,
			*teleportCollideTestPosition4EACA0,
		]{
			loadCollideData: func(obj *teleportCollideTestObject4EACA0) *teleportCollideTestData4EACA0 {
				events = append(events, "data")
				return obj.data
			},
			loadClass: func(obj *teleportCollideTestObject4EACA0) uint32 {
				events = append(events, "class")
				return obj.class
			},
			cachePosition: func(*teleportCollideTestObject4EACA0) *teleportCollideTestPosition4EACA0 {
				t.Fatal("position cached for rejected target")
				return nil
			},
		}

		teleportCollide4EACA0(source, target, struct{}{}, hooks)
		if !reflect.DeepEqual(events, []string{"data", "class"}) {
			t.Fatalf("events = %#v", events)
		}
	})
}

func TestTeleportCollide4EACA0SuccessOrderAndLiveValues(t *testing.T) {
	data := &teleportCollideTestData4EACA0{x: 1, y: 2}
	replacement := &teleportCollideTestData4EACA0{x: 7, y: 8}
	prePosition := &teleportCollideTestPosition4EACA0{name: "entry-position"}
	replacementPosition := &teleportCollideTestPosition4EACA0{name: "replacement-position"}
	destination := &teleportCollideTestPosition4EACA0{name: "field41-field42"}
	source := &teleportCollideTestObject4EACA0{data: data}
	target := &teleportCollideTestObject4EACA0{
		class:       0x80000004,
		position:    prePosition,
		destination: destination,
	}
	collision := &struct{ poison bool }{poison: true}
	var events []string
	audioCount := 0

	hooks := teleportCollideHooks4EACA0[
		*teleportCollideTestObject4EACA0,
		*teleportCollideTestData4EACA0,
		*teleportCollideTestPosition4EACA0,
		*teleportCollideTestPosition4EACA0,
	]{
		loadCollideData: func(obj *teleportCollideTestObject4EACA0) *teleportCollideTestData4EACA0 {
			events = append(events, "data")
			return obj.data
		},
		loadClass: func(obj *teleportCollideTestObject4EACA0) uint32 {
			events = append(events, "class")
			return obj.class
		},
		cachePosition: func(obj *teleportCollideTestObject4EACA0) *teleportCollideTestPosition4EACA0 {
			events = append(events, "position")
			return obj.position
		},
		pointFX: func(id uint32, position *teleportCollideTestPosition4EACA0) {
			events = append(events, fmt.Sprintf("fx%d", id))
			if position != prePosition {
				t.Fatalf("point position = %p, want cached %p", position, prePosition)
			}
			switch id {
			case teleportCollidePreFX4EACA0:
				source.data = replacement
				data.x = 1<<24 + 1
				target.class |= teleportCollideDoorClass4EACA0
			case teleportCollidePostFX4EACA0:
				if target.position != replacementPosition {
					t.Fatalf("teleport did not mutate live target position reference")
				}
			default:
				t.Fatalf("point FX = %d", id)
			}
		},
		audio: func(id uint32, obj *teleportCollideTestObject4EACA0) {
			events = append(events, fmt.Sprintf("audio%d", id))
			if id != teleportCollideSound4EACA0 || obj != target {
				t.Fatalf("audio = %d/%p", id, obj)
			}
			audioCount++
			if audioCount == 1 {
				data.y = 123
			}
		},
		loadDestinationX: func(got *teleportCollideTestData4EACA0) int32 {
			events = append(events, "load-x")
			if got != data {
				t.Fatalf("X data = %p, want cached %p", got, data)
			}
			return got.x
		},
		cacheDestination: func(obj *teleportCollideTestObject4EACA0) *teleportCollideTestPosition4EACA0 {
			events = append(events, "destination")
			return obj.destination
		},
		storeDestinationX: func(obj *teleportCollideTestObject4EACA0, value float32) {
			events = append(events, "store-x")
			obj.storedX = value
			data.y = math.MinInt32
		},
		loadDestinationY: func(got *teleportCollideTestData4EACA0) int32 {
			events = append(events, "load-y")
			if got != data {
				t.Fatalf("Y data = %p, want cached %p", got, data)
			}
			return got.y
		},
		storeDestinationY: func(obj *teleportCollideTestObject4EACA0, value float32) {
			events = append(events, "store-y")
			obj.storedY = value
		},
		teleport: func(obj *teleportCollideTestObject4EACA0, got *teleportCollideTestPosition4EACA0) {
			events = append(events, "teleport")
			if obj != target || got != destination {
				t.Fatalf("teleport = %p/%p, want %p/%p", obj, got, target, destination)
			}
			obj.position = replacementPosition
		},
	}

	teleportCollide4EACA0(source, target, collision, hooks)
	wantEvents := []string{
		"data", "class", "position", "fx138", "audio147",
		"load-x", "destination", "store-x", "load-y", "store-y",
		"teleport", "fx137", "audio147",
	}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %#v, want %#v", events, wantEvents)
	}
	if got, want := math.Float32bits(target.storedX), math.Float32bits(float32(1<<24+1)); got != want {
		t.Fatalf("stored X bits = %#08x, want %#08x", got, want)
	}
	if got, want := math.Float32bits(target.storedY), math.Float32bits(float32(math.MinInt32)); got != want {
		t.Fatalf("stored Y bits = %#08x, want %#08x", got, want)
	}
	if audioCount != 2 {
		t.Fatalf("audio count = %d", audioCount)
	}
}

func TestTeleportCollide4EACA0FaultAndStoreOrder(t *testing.T) {
	t.Run("nil data faults after pre effect and audio", func(t *testing.T) {
		source := &teleportCollideTestObject4EACA0{}
		target := &teleportCollideTestObject4EACA0{
			position: &teleportCollideTestPosition4EACA0{},
		}
		var events []string
		hooks := teleportCollideHooks4EACA0[
			*teleportCollideTestObject4EACA0,
			*teleportCollideTestData4EACA0,
			*teleportCollideTestPosition4EACA0,
			*teleportCollideTestPosition4EACA0,
		]{
			loadCollideData: func(obj *teleportCollideTestObject4EACA0) *teleportCollideTestData4EACA0 {
				events = append(events, "data")
				return obj.data
			},
			loadClass: func(obj *teleportCollideTestObject4EACA0) uint32 {
				events = append(events, "class")
				return obj.class
			},
			cachePosition: func(obj *teleportCollideTestObject4EACA0) *teleportCollideTestPosition4EACA0 {
				events = append(events, "position")
				return obj.position
			},
			pointFX: func(uint32, *teleportCollideTestPosition4EACA0) {
				events = append(events, "fx")
			},
			audio: func(uint32, *teleportCollideTestObject4EACA0) {
				events = append(events, "audio")
			},
			loadDestinationX: func(data *teleportCollideTestData4EACA0) int32 {
				events = append(events, "load-x")
				return data.x
			},
		}

		var recovered any
		func() {
			defer func() { recovered = recover() }()
			teleportCollide4EACA0(source, target, struct{}{}, hooks)
		}()
		if recovered == nil {
			t.Fatal("nil collide data did not fault on X load")
		}
		want := []string{"data", "class", "position", "fx", "audio", "load-x"}
		if !reflect.DeepEqual(events, want) {
			t.Fatalf("events = %#v, want %#v", events, want)
		}
	})

	t.Run("Y fault observes X store but suppresses teleport and post calls", func(t *testing.T) {
		data := &teleportCollideTestData4EACA0{x: 41, y: 42}
		source := &teleportCollideTestObject4EACA0{data: data}
		target := &teleportCollideTestObject4EACA0{
			position:    &teleportCollideTestPosition4EACA0{},
			destination: &teleportCollideTestPosition4EACA0{},
		}
		var events []string
		hooks := teleportCollideHooks4EACA0[
			*teleportCollideTestObject4EACA0,
			*teleportCollideTestData4EACA0,
			*teleportCollideTestPosition4EACA0,
			*teleportCollideTestPosition4EACA0,
		]{
			loadCollideData: func(obj *teleportCollideTestObject4EACA0) *teleportCollideTestData4EACA0 {
				events = append(events, "data")
				return obj.data
			},
			loadClass: func(obj *teleportCollideTestObject4EACA0) uint32 {
				events = append(events, "class")
				return obj.class
			},
			cachePosition: func(obj *teleportCollideTestObject4EACA0) *teleportCollideTestPosition4EACA0 {
				events = append(events, "position")
				return obj.position
			},
			pointFX: func(uint32, *teleportCollideTestPosition4EACA0) {
				events = append(events, "fx")
			},
			audio: func(uint32, *teleportCollideTestObject4EACA0) {
				events = append(events, "audio")
			},
			loadDestinationX: func(data *teleportCollideTestData4EACA0) int32 {
				events = append(events, "load-x")
				return data.x
			},
			cacheDestination: func(obj *teleportCollideTestObject4EACA0) *teleportCollideTestPosition4EACA0 {
				events = append(events, "destination")
				return obj.destination
			},
			storeDestinationX: func(obj *teleportCollideTestObject4EACA0, value float32) {
				events = append(events, "store-x")
				obj.storedX = value
			},
			loadDestinationY: func(*teleportCollideTestData4EACA0) int32 {
				events = append(events, "load-y")
				panic("Y fault")
			},
			storeDestinationY: func(*teleportCollideTestObject4EACA0, float32) {
				t.Fatal("Y stored after Y fault")
			},
			teleport: func(*teleportCollideTestObject4EACA0, *teleportCollideTestPosition4EACA0) {
				t.Fatal("teleport called after Y fault")
			},
		}

		var recovered any
		func() {
			defer func() { recovered = recover() }()
			teleportCollide4EACA0(source, target, struct{}{}, hooks)
		}()
		if recovered == nil {
			t.Fatal("Y load did not fault")
		}
		want := []string{
			"data", "class", "position", "fx", "audio", "load-x",
			"destination", "store-x", "load-y",
		}
		if !reflect.DeepEqual(events, want) {
			t.Fatalf("events = %#v, want %#v", events, want)
		}
		if target.storedX != 41 {
			t.Fatalf("stored X = %v", target.storedX)
		}
	})

	t.Run("teleport fault follows both stores and suppresses post calls", func(t *testing.T) {
		data := &teleportCollideTestData4EACA0{x: -3, y: 5}
		source := &teleportCollideTestObject4EACA0{data: data}
		target := &teleportCollideTestObject4EACA0{
			position:    &teleportCollideTestPosition4EACA0{},
			destination: &teleportCollideTestPosition4EACA0{},
		}
		var events []string
		fxCount := 0
		audioCount := 0
		hooks := teleportCollideHooks4EACA0[
			*teleportCollideTestObject4EACA0,
			*teleportCollideTestData4EACA0,
			*teleportCollideTestPosition4EACA0,
			*teleportCollideTestPosition4EACA0,
		]{
			loadCollideData: func(obj *teleportCollideTestObject4EACA0) *teleportCollideTestData4EACA0 { return obj.data },
			loadClass:       func(obj *teleportCollideTestObject4EACA0) uint32 { return obj.class },
			cachePosition:   func(obj *teleportCollideTestObject4EACA0) *teleportCollideTestPosition4EACA0 { return obj.position },
			pointFX: func(uint32, *teleportCollideTestPosition4EACA0) {
				fxCount++
			},
			audio: func(uint32, *teleportCollideTestObject4EACA0) {
				audioCount++
			},
			loadDestinationX: func(data *teleportCollideTestData4EACA0) int32 { return data.x },
			cacheDestination: func(obj *teleportCollideTestObject4EACA0) *teleportCollideTestPosition4EACA0 {
				return obj.destination
			},
			storeDestinationX: func(obj *teleportCollideTestObject4EACA0, value float32) { obj.storedX = value },
			loadDestinationY:  func(data *teleportCollideTestData4EACA0) int32 { return data.y },
			storeDestinationY: func(obj *teleportCollideTestObject4EACA0, value float32) { obj.storedY = value },
			teleport: func(*teleportCollideTestObject4EACA0, *teleportCollideTestPosition4EACA0) {
				events = append(events, "teleport")
				panic("teleport fault")
			},
		}

		var recovered any
		func() {
			defer func() { recovered = recover() }()
			teleportCollide4EACA0(source, target, struct{}{}, hooks)
		}()
		if recovered == nil {
			t.Fatal("teleport did not fault")
		}
		if target.storedX != -3 || target.storedY != 5 {
			t.Fatalf("stored destination = %v/%v", target.storedX, target.storedY)
		}
		if fxCount != 1 || audioCount != 1 || !reflect.DeepEqual(events, []string{"teleport"}) {
			t.Fatalf("callbacks = fx:%d audio:%d events:%#v", fxCount, audioCount, events)
		}
	})
}
