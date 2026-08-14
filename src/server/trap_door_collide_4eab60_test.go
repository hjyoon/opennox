package server

import (
	"math"
	"reflect"
	"testing"
)

type trapDoorCollideTestData4EAB60 struct {
	fallVelocityX int32
	fallVelocityY int32
	nextFrame     uint32
	delay         uint16
	activated     uint32
}

type trapDoorCollideTestObject4EAB60 struct {
	data      *trapDoorCollideTestData4EAB60
	class     uint32
	flags     uint32
	shapeKind uint32
	boxWidth  float32
	boxHeight float32
	circleR   float32
	posX      float32
	posY      float32
	fallVelX  float32
	fallVelY  float32
	fallPosX  float32
	fallPosY  float32
}

func defaultTrapDoorCollideHooks4EAB60() trapDoorCollideHooks4EAB60[
	*trapDoorCollideTestObject4EAB60,
	*trapDoorCollideTestData4EAB60,
] {
	return trapDoorCollideHooks4EAB60[
		*trapDoorCollideTestObject4EAB60,
		*trapDoorCollideTestData4EAB60,
	]{
		loadCollideData: func(obj *trapDoorCollideTestObject4EAB60) *trapDoorCollideTestData4EAB60 {
			return obj.data
		},
		loadClass: func(obj *trapDoorCollideTestObject4EAB60) uint32 { return obj.class },
		loadFlags: func(obj *trapDoorCollideTestObject4EAB60) uint32 { return obj.flags },
		loadShapeKind: func(obj *trapDoorCollideTestObject4EAB60) uint32 {
			return obj.shapeKind
		},
		loadBoxWidth: func(obj *trapDoorCollideTestObject4EAB60) float32 { return obj.boxWidth },
		loadBoxHeight: func(obj *trapDoorCollideTestObject4EAB60) float32 {
			return obj.boxHeight
		},
		loadCircleRadius: func(obj *trapDoorCollideTestObject4EAB60) float32 {
			return obj.circleR
		},
		mapPointInBox: func(*trapDoorCollideTestObject4EAB60, *trapDoorCollideTestObject4EAB60) bool {
			return true
		},
		orFlags: func(obj *trapDoorCollideTestObject4EAB60, flags uint32) { obj.flags |= flags },
		loadFallVelocityX: func(data *trapDoorCollideTestData4EAB60) int32 {
			return data.fallVelocityX
		},
		storeFallVelocityX: func(obj *trapDoorCollideTestObject4EAB60, value float32) {
			obj.fallVelX = value
		},
		loadFallVelocityY: func(data *trapDoorCollideTestData4EAB60) int32 {
			return data.fallVelocityY
		},
		storeFallVelocityY: func(obj *trapDoorCollideTestObject4EAB60, value float32) {
			obj.fallVelY = value
		},
		loadPosX: func(obj *trapDoorCollideTestObject4EAB60) float32 { return obj.posX },
		storeFallPosX: func(obj *trapDoorCollideTestObject4EAB60, value float32) {
			obj.fallPosX = value
		},
		loadPosY: func(obj *trapDoorCollideTestObject4EAB60) float32 { return obj.posY },
		storeFallPosY: func(obj *trapDoorCollideTestObject4EAB60, value float32) {
			obj.fallPosY = value
		},
		loadActivated: func(data *trapDoorCollideTestData4EAB60) uint32 {
			return data.activated
		},
		abilityActive: func(*trapDoorCollideTestObject4EAB60, int32) int32 { return 0 },
		loadDelay:     func(data *trapDoorCollideTestData4EAB60) uint16 { return data.delay },
		gameFrame:     func() uint32 { return 0 },
		storeNextFrame: func(data *trapDoorCollideTestData4EAB60, value uint32) {
			data.nextFrame = value
		},
		scriptCallback: func(
			*trapDoorCollideTestData4EAB60,
			*trapDoorCollideTestObject4EAB60,
			*trapDoorCollideTestObject4EAB60,
			int32,
		) {
		},
		storeActivated: func(data *trapDoorCollideTestData4EAB60, value uint32) {
			data.activated = value
		},
	}
}

func TestTrapDoorCollide4EAB60CachesDataBeforeTargetGuards(t *testing.T) {
	data := &trapDoorCollideTestData4EAB60{}
	source := &trapDoorCollideTestObject4EAB60{data: data}
	collision := &struct{ guard uint32 }{guard: 0x31415926}

	for _, tc := range []struct {
		name       string
		target     *trapDoorCollideTestObject4EAB60
		wantEvents []string
	}{
		{name: "nil target", wantEvents: []string{"data"}},
		{
			name:       "door target",
			target:     &trapDoorCollideTestObject4EAB60{class: 0x80000080},
			wantEvents: []string{"data", "class"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var events []string
			hooks := defaultTrapDoorCollideHooks4EAB60()
			hooks.loadCollideData = func(obj *trapDoorCollideTestObject4EAB60) *trapDoorCollideTestData4EAB60 {
				events = append(events, "data")
				return obj.data
			}
			hooks.loadClass = func(obj *trapDoorCollideTestObject4EAB60) uint32 {
				events = append(events, "class")
				return obj.class
			}
			hooks.loadFlags = func(*trapDoorCollideTestObject4EAB60) uint32 {
				t.Fatal("source flags reached")
				return 0
			}
			trapDoorCollide4EAB60(source, tc.target, collision, hooks)
			if !reflect.DeepEqual(events, tc.wantEvents) {
				t.Fatalf("events = %v, want %v", events, tc.wantEvents)
			}
			if collision.guard != 0x31415926 {
				t.Fatalf("collision guard = %#x", collision.guard)
			}
		})
	}
}

func TestTrapDoorCollide4EAB60EnabledShapeGates(t *testing.T) {
	qnan := math.Float32frombits(0x7fc12345)
	for _, tc := range []struct {
		name       string
		configure  func(source, target *trapDoorCollideTestObject4EAB60)
		wantMapped bool
	}{
		{
			name: "box width smaller",
			configure: func(source, target *trapDoorCollideTestObject4EAB60) {
				target.shapeKind, target.boxWidth = trapDoorCollideShapeBox4EAB60, 11
			},
		},
		{
			name: "box unordered width rejects",
			configure: func(source, target *trapDoorCollideTestObject4EAB60) {
				target.shapeKind, source.boxWidth = trapDoorCollideShapeBox4EAB60, qnan
			},
		},
		{
			name: "box height smaller",
			configure: func(source, target *trapDoorCollideTestObject4EAB60) {
				target.shapeKind, target.boxHeight = trapDoorCollideShapeBox4EAB60, 11
			},
		},
		{
			name: "box dimensions accepted",
			configure: func(source, target *trapDoorCollideTestObject4EAB60) {
				target.shapeKind, target.boxWidth, target.boxHeight = trapDoorCollideShapeBox4EAB60, 10, 10
			},
			wantMapped: true,
		},
		{
			name: "circle diameter too wide",
			configure: func(source, target *trapDoorCollideTestObject4EAB60) {
				target.shapeKind, target.circleR = trapDoorCollideShapeCircle4EAB60, 5.5
			},
		},
		{
			name: "circle diameter too tall",
			configure: func(source, target *trapDoorCollideTestObject4EAB60) {
				target.shapeKind, target.circleR, source.boxWidth, source.boxHeight = trapDoorCollideShapeCircle4EAB60, 5.5, 12, 10
			},
		},
		{
			name: "circle unordered radius continues",
			configure: func(source, target *trapDoorCollideTestObject4EAB60) {
				target.shapeKind, target.circleR = trapDoorCollideShapeCircle4EAB60, qnan
			},
			wantMapped: true,
		},
		{
			name: "circle unordered source width continues",
			configure: func(source, target *trapDoorCollideTestObject4EAB60) {
				target.shapeKind, target.circleR, source.boxWidth = trapDoorCollideShapeCircle4EAB60, 2, qnan
			},
			wantMapped: true,
		},
		{
			name: "other shape continues",
			configure: func(source, target *trapDoorCollideTestObject4EAB60) {
				target.shapeKind = 1
			},
			wantMapped: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source := &trapDoorCollideTestObject4EAB60{
				flags:     trapDoorCollideEnabled4EAB60,
				boxWidth:  10,
				boxHeight: 10,
			}
			target := &trapDoorCollideTestObject4EAB60{}
			tc.configure(source, target)
			mapped := false
			hooks := defaultTrapDoorCollideHooks4EAB60()
			hooks.mapPointInBox = func(gotSource, gotTarget *trapDoorCollideTestObject4EAB60) bool {
				mapped = true
				if gotSource != source || gotTarget != target {
					t.Fatalf("map args = %p/%p", gotSource, gotTarget)
				}
				return false
			}
			trapDoorCollide4EAB60(source, target, 0, hooks)
			if mapped != tc.wantMapped {
				t.Fatalf("mapped = %t, want %t", mapped, tc.wantMapped)
			}
		})
	}
}

func TestTrapDoorCollide4EAB60EnabledStoreOrderAndCachedData(t *testing.T) {
	oldData := &trapDoorCollideTestData4EAB60{
		fallVelocityX: 1<<24 + 1,
		fallVelocityY: -(1<<24 + 1),
	}
	replacement := &trapDoorCollideTestData4EAB60{fallVelocityX: 99, fallVelocityY: 101}
	source := &trapDoorCollideTestObject4EAB60{
		data:      oldData,
		flags:     trapDoorCollideEnabled4EAB60,
		shapeKind: 1,
		posX:      1.25,
		posY:      -2.5,
	}
	target := &trapDoorCollideTestObject4EAB60{flags: 0x20}
	collision := &struct{ guard uint32 }{guard: 0xa5a55a5a}
	var events []string
	hooks := defaultTrapDoorCollideHooks4EAB60()
	hooks.loadCollideData = func(obj *trapDoorCollideTestObject4EAB60) *trapDoorCollideTestData4EAB60 {
		events = append(events, "data")
		return obj.data
	}
	hooks.loadClass = func(obj *trapDoorCollideTestObject4EAB60) uint32 {
		events = append(events, "class")
		return obj.class
	}
	hooks.loadFlags = func(obj *trapDoorCollideTestObject4EAB60) uint32 {
		events = append(events, "flags")
		return obj.flags
	}
	hooks.loadShapeKind = func(obj *trapDoorCollideTestObject4EAB60) uint32 {
		events = append(events, "kind")
		return obj.shapeKind
	}
	hooks.mapPointInBox = func(*trapDoorCollideTestObject4EAB60, *trapDoorCollideTestObject4EAB60) bool {
		events = append(events, "map")
		source.data = replacement
		source.posX, source.posY = 7.5, -8.25
		return true
	}
	hooks.orFlags = func(obj *trapDoorCollideTestObject4EAB60, flags uint32) {
		events = append(events, "or flags")
		obj.flags |= flags
	}
	hooks.loadFallVelocityX = func(data *trapDoorCollideTestData4EAB60) int32 {
		events = append(events, "load vx")
		if data != oldData {
			t.Fatalf("X velocity data = %p", data)
		}
		return data.fallVelocityX
	}
	hooks.storeFallVelocityX = func(obj *trapDoorCollideTestObject4EAB60, value float32) {
		events = append(events, "store vx")
		obj.fallVelX = value
	}
	hooks.loadFallVelocityY = func(data *trapDoorCollideTestData4EAB60) int32 {
		events = append(events, "load vy")
		if data != oldData {
			t.Fatalf("Y velocity data = %p", data)
		}
		return data.fallVelocityY
	}
	hooks.storeFallVelocityY = func(obj *trapDoorCollideTestObject4EAB60, value float32) {
		events = append(events, "store vy")
		obj.fallVelY = value
	}
	hooks.loadPosX = func(obj *trapDoorCollideTestObject4EAB60) float32 {
		events = append(events, "load x")
		return obj.posX
	}
	hooks.storeFallPosX = func(obj *trapDoorCollideTestObject4EAB60, value float32) {
		events = append(events, "store x")
		obj.fallPosX = value
	}
	hooks.loadPosY = func(obj *trapDoorCollideTestObject4EAB60) float32 {
		events = append(events, "load y")
		return obj.posY
	}
	hooks.storeFallPosY = func(obj *trapDoorCollideTestObject4EAB60, value float32) {
		events = append(events, "store y")
		obj.fallPosY = value
	}

	trapDoorCollide4EAB60(source, target, collision, hooks)
	wantEvents := []string{
		"data", "class", "flags", "kind", "map", "or flags",
		"load vx", "store vx", "load vy", "store vy",
		"load x", "store x", "load y", "store y",
	}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %v, want %v", events, wantEvents)
	}
	if target.flags != 0x60020 || target.fallVelX != 1<<24 || target.fallVelY != -(1<<24) ||
		target.fallPosX != 7.5 || target.fallPosY != -8.25 {
		t.Fatalf("target state = %+v", target)
	}
	if source.data != replacement || collision.guard != 0xa5a55a5a {
		t.Fatalf("source/collision state = %p/%#x", source.data, collision.guard)
	}
}

func TestTrapDoorCollide4EAB60InactiveLiveDelayAndPostScriptActivation(t *testing.T) {
	oldData := &trapDoorCollideTestData4EAB60{}
	replacement := &trapDoorCollideTestData4EAB60{activated: 77}
	source := &trapDoorCollideTestObject4EAB60{data: oldData}
	target := &trapDoorCollideTestObject4EAB60{class: 0x80000004}
	var events []string
	hooks := defaultTrapDoorCollideHooks4EAB60()
	hooks.loadCollideData = func(obj *trapDoorCollideTestObject4EAB60) *trapDoorCollideTestData4EAB60 {
		events = append(events, "data")
		return obj.data
	}
	hooks.loadClass = func(obj *trapDoorCollideTestObject4EAB60) uint32 {
		events = append(events, "class")
		return obj.class
	}
	hooks.loadFlags = func(obj *trapDoorCollideTestObject4EAB60) uint32 {
		events = append(events, "flags")
		return obj.flags
	}
	hooks.loadActivated = func(data *trapDoorCollideTestData4EAB60) uint32 {
		events = append(events, "activated")
		return data.activated
	}
	hooks.abilityActive = func(obj *trapDoorCollideTestObject4EAB60, ability int32) int32 {
		events = append(events, "ability")
		if obj != target || ability != trapDoorCollideTreadLightly4EAB60 {
			t.Fatalf("ability args = %p/%d", obj, ability)
		}
		oldData.delay = 7
		source.data = replacement
		return 0
	}
	hooks.loadDelay = func(data *trapDoorCollideTestData4EAB60) uint16 {
		events = append(events, "delay")
		if data != oldData {
			t.Fatalf("delay data = %p", data)
		}
		return data.delay
	}
	hooks.gameFrame = func() uint32 {
		events = append(events, "frame")
		return 0xfffffffc
	}
	hooks.storeNextFrame = func(data *trapDoorCollideTestData4EAB60, value uint32) {
		events = append(events, "next frame")
		data.nextFrame = value
	}
	hooks.scriptCallback = func(
		data *trapDoorCollideTestData4EAB60,
		caller, trigger *trapDoorCollideTestObject4EAB60,
		event int32,
	) {
		events = append(events, "script")
		if data != oldData || caller != target || trigger != source || event != trapDoorCollideEvent4EAB60 {
			t.Fatalf("script args = %p/%p/%p/%d", data, caller, trigger, event)
		}
		if data.nextFrame != 3 || data.activated != 0 {
			t.Fatalf("state at script = %+v", data)
		}
		data.activated = 99
	}
	hooks.storeActivated = func(data *trapDoorCollideTestData4EAB60, value uint32) {
		events = append(events, "store activated")
		data.activated = value
	}

	trapDoorCollide4EAB60(source, target, 0, hooks)
	wantEvents := []string{
		"data", "class", "flags", "activated", "ability", "delay",
		"frame", "next frame", "script", "store activated",
	}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %v, want %v", events, wantEvents)
	}
	if oldData.nextFrame != 3 || oldData.activated != 1 || replacement.activated != 77 || source.data != replacement {
		t.Fatalf("final data = old %+v replacement %+v source %p", oldData, replacement, source.data)
	}
}

func TestTrapDoorCollide4EAB60InactiveEarlyReturnsAndNonPlayerSkip(t *testing.T) {
	for _, tc := range []struct {
		name          string
		data          trapDoorCollideTestData4EAB60
		class         uint32
		abilityResult int32
		wantAbility   int
		wantScript    int
	}{
		{name: "already activated", data: trapDoorCollideTestData4EAB60{activated: 2}},
		{
			name:          "player tread lightly",
			class:         trapDoorCollidePlayerClass4EAB60,
			abilityResult: -7,
			wantAbility:   1,
		},
		{name: "non player skips ability", class: 0x400, wantScript: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source := &trapDoorCollideTestObject4EAB60{data: &tc.data}
			target := &trapDoorCollideTestObject4EAB60{class: tc.class}
			abilityCalls, scriptCalls, frameCalls := 0, 0, 0
			hooks := defaultTrapDoorCollideHooks4EAB60()
			hooks.abilityActive = func(*trapDoorCollideTestObject4EAB60, int32) int32 {
				abilityCalls++
				return tc.abilityResult
			}
			hooks.gameFrame = func() uint32 {
				frameCalls++
				return 10
			}
			hooks.scriptCallback = func(
				*trapDoorCollideTestData4EAB60,
				*trapDoorCollideTestObject4EAB60,
				*trapDoorCollideTestObject4EAB60,
				int32,
			) {
				scriptCalls++
			}
			trapDoorCollide4EAB60(source, target, 0, hooks)
			if abilityCalls != tc.wantAbility || scriptCalls != tc.wantScript || frameCalls != 0 {
				t.Fatalf("calls = ability %d script %d frame %d", abilityCalls, scriptCalls, frameCalls)
			}
		})
	}
}

func TestTrapDoorCollide4EAB60FaultOrder(t *testing.T) {
	t.Run("nil source faults before nil target branch", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Fatal("nil source returned")
			}
		}()
		trapDoorCollide4EAB60[*trapDoorCollideTestObject4EAB60](
			nil, nil, 0, defaultTrapDoorCollideHooks4EAB60(),
		)
	})

	t.Run("nil data is safe for nil target", func(t *testing.T) {
		trapDoorCollide4EAB60(
			&trapDoorCollideTestObject4EAB60{},
			nil,
			0,
			defaultTrapDoorCollideHooks4EAB60(),
		)
	})

	t.Run("enabled nil data faults after fall flags", func(t *testing.T) {
		source := &trapDoorCollideTestObject4EAB60{
			flags:     trapDoorCollideEnabled4EAB60,
			shapeKind: 1,
		}
		target := &trapDoorCollideTestObject4EAB60{flags: 0x20}
		defer func() {
			if recover() == nil {
				t.Fatal("nil data returned")
			}
			if target.flags != 0x60020 {
				t.Fatalf("target flags at fault = %#x", target.flags)
			}
		}()
		trapDoorCollide4EAB60(source, target, 0, defaultTrapDoorCollideHooks4EAB60())
	})

	t.Run("failed map does not dereference nil data", func(t *testing.T) {
		source := &trapDoorCollideTestObject4EAB60{
			flags:     trapDoorCollideEnabled4EAB60,
			shapeKind: 1,
		}
		target := &trapDoorCollideTestObject4EAB60{flags: 0x20}
		hooks := defaultTrapDoorCollideHooks4EAB60()
		hooks.mapPointInBox = func(*trapDoorCollideTestObject4EAB60, *trapDoorCollideTestObject4EAB60) bool {
			return false
		}
		trapDoorCollide4EAB60(source, target, 0, hooks)
		if target.flags != 0x20 {
			t.Fatalf("target flags = %#x", target.flags)
		}
	})
}
