package server

import (
	"image"
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

type projectileTestObject537770 struct {
	name       string
	class      object.Class
	subclass   object.SubClass
	flags      object.Flags
	collide    bool
	callback   int
	owner      *projectileTestObject537770
	team       int
	enemy      bool
	position   types.Pointf
	newPos     types.Pointf
	typeIndex  uint16
	doorUpdate *projectileTestDoorUpdate537770
}

type projectileTestDoorUpdate537770 struct {
	direction int32
}

func projectileFilterTestHooks54E730() projectileCanCollideHooks54E730[*projectileTestObject537770] {
	return projectileCanCollideHooks54E730[*projectileTestObject537770]{
		loadClassLow: func(obj *projectileTestObject537770) uint8 {
			return uint8(obj.class)
		},
		loadSubclassLow: func(obj *projectileTestObject537770) uint8 {
			return uint8(obj.subclass)
		},
		loadFlags: func(obj *projectileTestObject537770) object.Flags {
			return obj.flags
		},
		hasCollide: func(obj *projectileTestObject537770) bool {
			return obj.collide
		},
		loadOwner: func(obj *projectileTestObject537770) *projectileTestObject537770 {
			return obj.owner
		},
		sameTeam: func(first, second *projectileTestObject537770) bool {
			return first != nil && second != nil && first.team != 0 && first.team == second.team
		},
		isEnemy: func(_, second *projectileTestObject537770) bool {
			return second.enemy
		},
	}
}

func TestProjectileCanCollide54E730Filters(t *testing.T) {
	tests := []struct {
		name string
		edit func(source, candidate, owner *projectileTestObject537770)
		want bool
	}{
		{name: "default", want: true},
		{name: "candidate missile", edit: func(_, candidate, _ *projectileTestObject537770) {
			candidate.class = object.ClassMissile
		}},
		{name: "source destroyed", edit: func(source, _, _ *projectileTestObject537770) {
			source.flags = object.FlagDestroyed
		}},
		{name: "candidate destroyed", edit: func(_, candidate, _ *projectileTestObject537770) {
			candidate.flags = object.FlagDestroyed
		}},
		{name: "source callback missing", edit: func(source, _, _ *projectileTestObject537770) {
			source.collide = false
		}},
		{name: "candidate callback missing", edit: func(_, candidate, _ *projectileTestObject537770) {
			candidate.collide = false
		}},
		{name: "candidate no collide", edit: func(_, candidate, _ *projectileTestObject537770) {
			candidate.flags = object.FlagNoCollide
		}},
		{name: "missile hit bypass", edit: func(source, candidate, _ *projectileTestObject537770) {
			source.flags = object.FlagBelow | object.FlagNoCollideOwner
			candidate.flags = object.FlagMissileHit | object.FlagAirborne
			source.team = 1
			candidate.team = 1
		}, want: true},
		{name: "source below candidate airborne", edit: func(source, candidate, _ *projectileTestObject537770) {
			source.flags = object.FlagBelow
			candidate.flags = object.FlagAirborne
		}},
		{name: "source overlap candidate airborne is allowed", edit: func(source, candidate, _ *projectileTestObject537770) {
			source.flags = object.FlagAllowOverlap
			candidate.flags = object.FlagAirborne
		}, want: true},
		{name: "candidate short source airborne", edit: func(source, candidate, _ *projectileTestObject537770) {
			source.flags = object.FlagAirborne
			candidate.flags = object.FlagShort
		}},
		{name: "no collide owner same team", edit: func(source, candidate, _ *projectileTestObject537770) {
			source.flags = object.FlagNoCollideOwner
			source.team = 7
			candidate.team = 7
		}},
		{name: "monster owner non enemy monster", edit: func(source, candidate, owner *projectileTestObject537770) {
			source.owner = owner
			owner.class = object.ClassMonster
			candidate.class = object.ClassMonster
			candidate.enemy = false
		}},
		{name: "subclass two bypasses monster owner rule", edit: func(source, candidate, owner *projectileTestObject537770) {
			source.owner = owner
			source.subclass = 2
			owner.class = object.ClassMonster
			candidate.class = object.ClassMonster
			candidate.enemy = false
		}, want: true},
		{name: "player owner is not monster", edit: func(source, candidate, owner *projectileTestObject537770) {
			source.owner = owner
			owner.class = object.ClassPlayer
			candidate.class = object.ClassMonster
			candidate.enemy = false
		}, want: true},
		{name: "enemy monster on different team", edit: func(source, candidate, owner *projectileTestObject537770) {
			source.owner = owner
			owner.class = object.ClassMonster
			owner.team = 1
			candidate.class = object.ClassMonster
			candidate.team = 2
			candidate.enemy = true
		}, want: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			source := &projectileTestObject537770{class: object.ClassMissile, collide: true}
			candidate := &projectileTestObject537770{class: object.ClassMonster, collide: true, enemy: true}
			owner := &projectileTestObject537770{}
			if test.edit != nil {
				test.edit(source, candidate, owner)
			}
			if got := projectileCanCollide54E730(source, candidate, projectileFilterTestHooks54E730()); got != test.want {
				t.Fatalf("projectileCanCollide54E730() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestProjectileCanCollide54E730OwnerLoadBoundaryAndReload(t *testing.T) {
	t.Run("airborne rejection does not load owner", func(t *testing.T) {
		source := &projectileTestObject537770{
			class:   object.ClassMissile,
			flags:   object.FlagBelow,
			collide: true,
		}
		candidate := &projectileTestObject537770{
			class:   object.ClassMonster,
			flags:   object.FlagAirborne,
			collide: true,
		}
		hooks := projectileFilterTestHooks54E730()
		ownerLoads := 0
		hooks.loadOwner = func(obj *projectileTestObject537770) *projectileTestObject537770 {
			ownerLoads++
			return obj.owner
		}
		if projectileCanCollide54E730(source, candidate, hooks) {
			t.Fatal("airborne candidate unexpectedly accepted")
		}
		if ownerLoads != 0 {
			t.Fatalf("owner loaded %d times before the airborne rejection", ownerLoads)
		}
	})

	t.Run("owner is reloaded after enemy callback", func(t *testing.T) {
		owner := &projectileTestObject537770{class: object.ClassMonster, team: 1}
		replacement := &projectileTestObject537770{team: 7}
		source := &projectileTestObject537770{
			class:   object.ClassMissile,
			collide: true,
			owner:   owner,
		}
		candidate := &projectileTestObject537770{
			class:   object.ClassMonster,
			collide: true,
			team:    7,
		}
		hooks := projectileFilterTestHooks54E730()
		ownerLoads := 0
		hooks.loadOwner = func(obj *projectileTestObject537770) *projectileTestObject537770 {
			ownerLoads++
			return obj.owner
		}
		hooks.isEnemy = func(first, second *projectileTestObject537770) bool {
			if first != owner || second != candidate {
				t.Fatalf("enemy arguments = (%p, %p)", first, second)
			}
			source.owner = replacement
			return true
		}
		if projectileCanCollide54E730(source, candidate, hooks) {
			t.Fatal("reloaded same-team owner did not reject the collision")
		}
		if ownerLoads != 2 {
			t.Fatalf("owner loaded %d times, want initial load plus post-callback reload", ownerLoads)
		}
	})
}

func TestProjectileSampleObject54E810LastMatchAndDoorRollback(t *testing.T) {
	first := &projectileTestObject537770{name: "first"}
	door := &projectileTestObject537770{
		name:       "door",
		class:      0x80,
		position:   types.Ptf(20, 30),
		doorUpdate: &projectileTestDoorUpdate537770{direction: 4},
	}
	last := &projectileTestObject537770{name: "last"}
	current := types.Ptf(10, 12)
	previous := types.Ptf(2, 3)
	var intersectSamples []types.Pointf

	found := projectileSampleObject54E810(first, &current, &previous, projectileSampleObjectHooks54E810[
		*projectileTestObject537770,
		*projectileTestDoorUpdate537770,
	]{
		eachAt: func(point types.Pointf, callback func(*projectileTestObject537770)) {
			if point != (types.Ptf(10, 12)) {
				t.Fatalf("traversal point = %v", point)
			}
			for _, candidate := range []*projectileTestObject537770{first, door, last} {
				callback(candidate)
			}
		},
		loadClassLow:    func(obj *projectileTestObject537770) uint8 { return uint8(obj.class) },
		loadSubclassLow: func(obj *projectileTestObject537770) uint8 { return uint8(obj.subclass) },
		loadDoorUpdate: func(obj *projectileTestObject537770) *projectileTestDoorUpdate537770 {
			return obj.doorUpdate
		},
		loadDoorDirection: func(update *projectileTestDoorUpdate537770) int32 { return update.direction },
		loadPosX:          func(obj *projectileTestObject537770) float32 { return obj.position.X },
		loadPosY:          func(obj *projectileTestObject537770) float32 { return obj.position.Y },
		doorSize: func(direction byte) image.Point {
			if direction != 4 {
				t.Fatalf("door direction = %d", direction)
			}
			return image.Pt(5, -7)
		},
		lineTrace: func(projectile, doorLine types.Rectf) bool {
			if projectile != (types.Rectf{Min: previous, Max: types.Ptf(10, 12)}) {
				t.Fatalf("projectile line = %v", projectile)
			}
			if doorLine != (types.Rectf{Min: door.position, Max: types.Ptf(25, 23)}) {
				t.Fatalf("door line = %v", doorLine)
			}
			return true
		},
		canCollide: func(_, _ *projectileTestObject537770) bool { return true },
		intersects: func(_ *projectileTestObject537770, point *types.Pointf) bool {
			intersectSamples = append(intersectSamples, *point)
			return true
		},
	})

	if found != last {
		t.Fatalf("found = %v, want last candidate", found.name)
	}
	if current != previous {
		t.Fatalf("current = %v, want Door rollback %v", current, previous)
	}
	wantSamples := []types.Pointf{types.Ptf(10, 12), previous}
	if !reflect.DeepEqual(intersectSamples, wantSamples) {
		t.Fatalf("intersection samples = %v, want %v", intersectSamples, wantSamples)
	}
}

func TestProjectileSampleObject54E850LoadsSubclassBeforeDoorUpdate(t *testing.T) {
	door := &projectileTestObject537770{class: 0x80, subclass: 4}
	current := types.Ptf(10, 12)
	previous := types.Ptf(2, 3)
	var events []string
	_ = projectileSampleObject54E810(door, &current, &previous, projectileSampleObjectHooks54E810[
		*projectileTestObject537770,
		*projectileTestDoorUpdate537770,
	]{
		eachAt:       func(_ types.Pointf, callback func(*projectileTestObject537770)) { callback(door) },
		loadClassLow: func(*projectileTestObject537770) uint8 { return 0x80 },
		loadSubclassLow: func(*projectileTestObject537770) uint8 {
			events = append(events, "subclass")
			return 4
		},
		loadDoorUpdate: func(*projectileTestObject537770) *projectileTestDoorUpdate537770 {
			events = append(events, "update")
			return nil
		},
		loadDoorDirection: func(*projectileTestDoorUpdate537770) int32 {
			t.Fatal("direction must not be read for subclass bit 4")
			return 0
		},
	})
	if !reflect.DeepEqual(events, []string{"subclass", "update"}) {
		t.Fatalf("events = %v", events)
	}
}

func projectileTraceTestHooks537850(
	source *projectileTestObject537770,
) projectileTraceHitHooks537850[*projectileTestObject537770] {
	return projectileTraceHitHooks537850[*projectileTestObject537770]{
		loadPosX:       func(obj *projectileTestObject537770) float32 { return obj.position.X },
		loadPosY:       func(obj *projectileTestObject537770) float32 { return obj.position.Y },
		loadNewPosX:    func(obj *projectileTestObject537770) float32 { return obj.newPos.X },
		loadNewPosY:    func(obj *projectileTestObject537770) float32 { return obj.newPos.Y },
		storeNewPosX:   func(obj *projectileTestObject537770, value float32) { obj.newPos.X = value },
		storeNewPosY:   func(obj *projectileTestObject537770, value float32) { obj.newPos.Y = value },
		loadObjectPosX: func(obj *projectileTestObject537770) float32 { return obj.position.X },
		loadObjectPosY: func(obj *projectileTestObject537770) float32 { return obj.position.Y },
		sampleObject: func(*projectileTestObject537770, *types.Pointf, *types.Pointf) *projectileTestObject537770 {
			return nil
		},
		mapTrace: func(types.Pointf, types.Pointf, MapTraceFlags) projectileMapTraceResult537850 {
			return projectileMapTraceResult537850{clear: true}
		},
		storeTraceGrid: func(image.Point) {},
		setTraceReady:  func(uint32) {},
		wallNormal: func(image.Point, types.Rectf) (types.Pointf, bool) {
			return types.Pointf{}, false
		},
	}
}

func TestProjectileTraceHit537850ShortAndNaNPaths(t *testing.T) {
	hit := &projectileTestObject537770{position: types.Ptf(0, 1)}
	source := &projectileTestObject537770{position: types.Ptf(1, 2), newPos: types.Ptf(4, 6)}
	hooks := projectileTraceTestHooks537850(source)
	var gotCurrent, gotPrevious types.Pointf
	hooks.sampleObject = func(_ *projectileTestObject537770, current, previous *types.Pointf) *projectileTestObject537770 {
		gotCurrent, gotPrevious = *current, *previous
		return hit
	}
	gotHit, normal, ok := projectileTraceHit537850(source, hooks)
	if !ok || gotHit != hit || normal != (types.Ptf(1, 1)) {
		t.Fatalf("result = (%p, %v, %v)", gotHit, normal, ok)
	}
	if gotCurrent != source.newPos || gotPrevious != source.position {
		t.Fatalf("sample = current %v previous %v", gotCurrent, gotPrevious)
	}

	source.newPos.X = float32(math.NaN())
	hooks = projectileTraceTestHooks537850(source)
	samples := 0
	hooks.sampleObject = func(_ *projectileTestObject537770, current, _ *types.Pointf) *projectileTestObject537770 {
		samples++
		if !math.IsNaN(float64(current.X)) {
			t.Fatalf("NaN sample X = %v", current.X)
		}
		return nil
	}
	_, _, _ = projectileTraceHit537850(source, hooks)
	if samples != 1 {
		t.Fatalf("NaN path sampled %d times, want short path once", samples)
	}
}

func TestProjectileTraceHit537850ShortSampleLoadOrder(t *testing.T) {
	source := &projectileTestObject537770{position: types.Ptf(1, 2), newPos: types.Ptf(4, 6)}
	hooks := projectileTraceTestHooks537850(source)
	fieldLoads := 0
	sampled := false
	var events []string
	record := func(name string, value float32) float32 {
		fieldLoads++
		// The first four reads form the delta. Record only the subsequent
		// short-path sample construction and stop once sampling begins.
		if fieldLoads > 4 && !sampled {
			events = append(events, name)
		}
		return value
	}
	hooks.loadPosX = func(obj *projectileTestObject537770) float32 {
		return record("pos-x", obj.position.X)
	}
	hooks.loadPosY = func(obj *projectileTestObject537770) float32 {
		return record("pos-y", obj.position.Y)
	}
	hooks.loadNewPosX = func(obj *projectileTestObject537770) float32 {
		return record("new-x", obj.newPos.X)
	}
	hooks.loadNewPosY = func(obj *projectileTestObject537770) float32 {
		return record("new-y", obj.newPos.Y)
	}
	hooks.sampleObject = func(*projectileTestObject537770, *types.Pointf, *types.Pointf) *projectileTestObject537770 {
		sampled = true
		return nil
	}
	_, _, _ = projectileTraceHit537850(source, hooks)
	want := []string{"new-y", "pos-x", "new-x", "pos-y"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("short sample field loads = %v, want %v", events, want)
	}
}

func TestProjectileTraceHit537850LongPathUsesNearestEvenSteps(t *testing.T) {
	source := &projectileTestObject537770{newPos: types.Ptf(15, 0)}
	hooks := projectileTraceTestHooks537850(source)
	var samples []types.Pointf
	hooks.sampleObject = func(_ *projectileTestObject537770, current, _ *types.Pointf) *projectileTestObject537770 {
		samples = append(samples, *current)
		return nil
	}
	_, _, ok := projectileTraceHit537850(source, hooks)
	if ok {
		t.Fatal("clear long trace unexpectedly reported a hit")
	}
	want := []types.Pointf{types.Ptf(3.75, 0), types.Ptf(7.5, 0), types.Ptf(11.25, 0), types.Ptf(15, 0)}
	if !reflect.DeepEqual(samples, want) {
		t.Fatalf("samples = %v, want %v", samples, want)
	}

	hit := &projectileTestObject537770{position: types.Ptf(3, 0)}
	samples = nil
	hooks.sampleObject = func(_ *projectileTestObject537770, current, _ *types.Pointf) *projectileTestObject537770 {
		samples = append(samples, *current)
		if len(samples) == 2 {
			return hit
		}
		return nil
	}
	gotHit, normal, ok := projectileTraceHit537850(source, hooks)
	if !ok || gotHit != hit || normal != (types.Ptf(0.75, 0)) {
		t.Fatalf("second sample result = (%p, %v, %v)", gotHit, normal, ok)
	}
}

func TestProjectileTraceHit537850WallStateAndSelection(t *testing.T) {
	t.Run("wall only", func(t *testing.T) {
		source := &projectileTestObject537770{position: types.Ptf(5, 6), newPos: types.Ptf(1, 2)}
		hooks := projectileTraceTestHooks537850(source)
		grid := image.Pt(7, 9)
		var stored image.Point
		var ready uint32
		hooks.mapTrace = func(from, to types.Pointf, flags MapTraceFlags) projectileMapTraceResult537850 {
			if from != source.position || to != source.newPos || flags != MapTraceFlags(5) {
				t.Fatalf("map trace = %v -> %v flags %d", from, to, flags)
			}
			return projectileMapTraceResult537850{point: types.Ptf(3, 4), grid: grid}
		}
		hooks.storeTraceGrid = func(point image.Point) { stored = point }
		hooks.setTraceReady = func(value uint32) { ready = value }
		hooks.wallNormal = func(point image.Point, ray types.Rectf) (types.Pointf, bool) {
			if point != grid || ray != (types.Rectf{Min: types.Ptf(5, 6), Max: types.Ptf(3, 4)}) {
				t.Fatalf("wall normal input = %v %v", point, ray)
			}
			return types.Ptf(-1, 0), true
		}
		hit, normal, ok := projectileTraceHit537850(source, hooks)
		if !ok || hit != nil || normal != (types.Ptf(-1, 0)) {
			t.Fatalf("wall result = (%p, %v, %v)", hit, normal, ok)
		}
		if stored != grid || ready != 1 || source.newPos != source.position {
			t.Fatalf("wall state = grid %v ready %d new %v", stored, ready, source.newPos)
		}
	})

	t.Run("normal failure still records and resets", func(t *testing.T) {
		source := &projectileTestObject537770{position: types.Ptf(5, 6), newPos: types.Ptf(8, 10)}
		hooks := projectileTraceTestHooks537850(source)
		stored := false
		ready := uint32(0)
		hooks.mapTrace = func(types.Pointf, types.Pointf, MapTraceFlags) projectileMapTraceResult537850 {
			return projectileMapTraceResult537850{point: types.Ptf(7, 8), grid: image.Pt(4, 3)}
		}
		hooks.storeTraceGrid = func(image.Point) { stored = true }
		hooks.setTraceReady = func(value uint32) { ready = value }
		_, _, ok := projectileTraceHit537850(source, hooks)
		if ok || !stored || ready != 1 || source.newPos != source.position {
			t.Fatalf("failure state = ok %v stored %v ready %d new %v", ok, stored, ready, source.newPos)
		}
	})

	t.Run("blocked reset loads both coordinates before stores", func(t *testing.T) {
		source := &projectileTestObject537770{position: types.Ptf(5, 6), newPos: types.Ptf(8, 10)}
		hooks := projectileTraceTestHooks537850(source)
		recording := false
		var events []string
		hooks.loadPosX = func(obj *projectileTestObject537770) float32 {
			if recording {
				events = append(events, "load-x")
			}
			return obj.position.X
		}
		hooks.loadPosY = func(obj *projectileTestObject537770) float32 {
			if recording {
				events = append(events, "load-y")
			}
			return obj.position.Y
		}
		hooks.storeNewPosX = func(obj *projectileTestObject537770, value float32) {
			events = append(events, "store-x")
			obj.newPos.X = value
		}
		hooks.storeNewPosY = func(obj *projectileTestObject537770, value float32) {
			events = append(events, "store-y")
			obj.newPos.Y = value
		}
		hooks.mapTrace = func(types.Pointf, types.Pointf, MapTraceFlags) projectileMapTraceResult537850 {
			return projectileMapTraceResult537850{point: types.Ptf(7, 8), grid: image.Pt(4, 3)}
		}
		hooks.wallNormal = func(image.Point, types.Rectf) (types.Pointf, bool) {
			recording = true
			return types.Pointf{}, false
		}
		_, _, _ = projectileTraceHit537850(source, hooks)
		want := []string{"load-x", "load-y", "store-x", "store-y"}
		if !reflect.DeepEqual(events, want) {
			t.Fatalf("blocked reset events = %v, want %v", events, want)
		}
	})

	tests := []struct {
		name       string
		hitX       float32
		wantObject bool
	}{
		{name: "tie selects wall", hitX: 3},
		{name: "closer object wins", hitX: 2, wantObject: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			source := &projectileTestObject537770{newPos: types.Ptf(4, 0)}
			hit := &projectileTestObject537770{position: types.Ptf(test.hitX, 0)}
			hooks := projectileTraceTestHooks537850(source)
			hooks.sampleObject = func(*projectileTestObject537770, *types.Pointf, *types.Pointf) *projectileTestObject537770 {
				return hit
			}
			hooks.mapTrace = func(types.Pointf, types.Pointf, MapTraceFlags) projectileMapTraceResult537850 {
				return projectileMapTraceResult537850{point: types.Ptf(3, 0), grid: image.Pt(1, 2)}
			}
			hooks.wallNormal = func(image.Point, types.Rectf) (types.Pointf, bool) {
				return types.Ptf(9, 8), true
			}
			gotHit, normal, ok := projectileTraceHit537850(source, hooks)
			if !ok {
				t.Fatal("collision not reported")
			}
			if test.wantObject {
				if gotHit != hit || normal != (types.Ptf(-test.hitX, 0)) {
					t.Fatalf("object result = (%p, %v)", gotHit, normal)
				}
			} else if gotHit != nil || normal != (types.Ptf(9, 8)) {
				t.Fatalf("wall result = (%p, %v)", gotHit, normal)
			}
		})
	}

	testsNaN := []struct {
		name      string
		hitX      float32
		wallPoint types.Pointf
	}{
		{name: "unordered object distance selects object", hitX: float32(math.NaN()), wallPoint: types.Ptf(3, 0)},
		{name: "unordered wall distance selects object", hitX: 2, wallPoint: types.Ptf(float32(math.NaN()), 0)},
	}
	for _, test := range testsNaN {
		t.Run(test.name, func(t *testing.T) {
			source := &projectileTestObject537770{newPos: types.Ptf(4, 0)}
			hit := &projectileTestObject537770{position: types.Ptf(test.hitX, 0)}
			hooks := projectileTraceTestHooks537850(source)
			hooks.sampleObject = func(*projectileTestObject537770, *types.Pointf, *types.Pointf) *projectileTestObject537770 {
				return hit
			}
			hooks.mapTrace = func(types.Pointf, types.Pointf, MapTraceFlags) projectileMapTraceResult537850 {
				return projectileMapTraceResult537850{point: test.wallPoint, grid: image.Pt(1, 2)}
			}
			hooks.wallNormal = func(image.Point, types.Rectf) (types.Pointf, bool) {
				return types.Ptf(9, 8), true
			}
			gotHit, _, ok := projectileTraceHit537850(source, hooks)
			if !ok || gotHit != hit {
				t.Fatalf("unordered result = (%p, %v), want object %p", gotHit, ok, hit)
			}
		})
	}
}

func TestProjectileCollisionDispatch537770CacheOrderBeforeFlagExit(t *testing.T) {
	source := &projectileTestObject537770{flags: object.FlagDestroyed}
	var small, medium, large uint32
	var events []string
	hooks := projectileCollisionDispatchHooks537770[*projectileTestObject537770, int]{
		loadSmallCache:   func() uint32 { events = append(events, "small"); return small },
		storeSmallCache:  func(value uint32) { events = append(events, "store-small"); small = value },
		loadMediumCache:  func() uint32 { return medium },
		storeMediumCache: func(value uint32) { events = append(events, "store-medium"); medium = value },
		loadLargeCache:   func() uint32 { return large },
		storeLargeCache:  func(value uint32) { events = append(events, "store-large"); large = value },
		lookupType: func(name string) uint32 {
			events = append(events, "lookup-"+name)
			return map[string]uint32{"SmallFist": 11, "MediumFist": 12, "LargeFist": 13}[name]
		},
		loadFlagsLow: func(obj *projectileTestObject537770) uint8 {
			events = append(events, "flags")
			return uint8(obj.flags)
		},
		traceHit: func(*projectileTestObject537770) (*projectileTestObject537770, types.Pointf, bool) {
			t.Fatal("destroyed source must not trace")
			return nil, types.Pointf{}, false
		},
	}
	projectileCollisionDispatch537770(source, hooks)
	wantEvents := []string{
		"small", "lookup-SmallFist", "store-small",
		"lookup-MediumFist", "store-medium", "lookup-LargeFist", "store-large", "flags",
	}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %v, want %v", events, wantEvents)
	}
	if small != 11 || medium != 12 || large != 13 {
		t.Fatalf("caches = %d %d %d", small, medium, large)
	}
}

func TestProjectileCollisionDispatch537770FistPreservesTraceReady(t *testing.T) {
	source := &projectileTestObject537770{}
	hit := &projectileTestObject537770{typeIndex: 11}
	ready := uint32(99)
	calls := 0
	var events []string
	hooks := projectileCollisionDispatchHooks537770[*projectileTestObject537770, int]{
		loadSmallCache: func() uint32 {
			events = append(events, "small")
			return 11
		},
		loadMediumCache: func() uint32 { return 12 },
		loadLargeCache:  func() uint32 { return 13 },
		loadFlagsLow:    func(*projectileTestObject537770) uint8 { return 0 },
		traceHit: func(*projectileTestObject537770) (*projectileTestObject537770, types.Pointf, bool) {
			ready = 1
			return hit, types.Ptf(2, 3), true
		},
		loadTypeIndex: func(obj *projectileTestObject537770) uint16 {
			events = append(events, "type")
			return obj.typeIndex
		},
		loadCollide: func(obj *projectileTestObject537770) int { return obj.callback },
		callCollide: func(int, *projectileTestObject537770, *projectileTestObject537770, *types.Pointf) {
			calls++
		},
		setTraceReady: func(value uint32) { ready = value },
	}
	projectileCollisionDispatch537770(source, hooks)
	if ready != 1 || calls != 0 {
		t.Fatalf("ready = %d, calls = %d", ready, calls)
	}
	if !reflect.DeepEqual(events, []string{"small", "small", "type"}) {
		t.Fatalf("cache/type load order = %v", events)
	}
}

func TestProjectileCollisionDispatch537770LiveTargetCallbackAndNormal(t *testing.T) {
	source := &projectileTestObject537770{callback: 1}
	hit := &projectileTestObject537770{callback: 2, typeIndex: 20}
	var readyWrites []uint32
	type call struct {
		callback int
		first    *projectileTestObject537770
		second   *projectileTestObject537770
		normal   types.Pointf
	}
	var calls []call
	hooks := projectileCollisionDispatchHooks537770[*projectileTestObject537770, int]{
		loadSmallCache:  func() uint32 { return 11 },
		loadMediumCache: func() uint32 { return 12 },
		loadLargeCache:  func() uint32 { return 13 },
		loadFlagsLow:    func(*projectileTestObject537770) uint8 { return 0 },
		traceHit: func(*projectileTestObject537770) (*projectileTestObject537770, types.Pointf, bool) {
			return hit, types.Ptf(2, -4), true
		},
		loadTypeIndex: func(obj *projectileTestObject537770) uint16 { return obj.typeIndex },
		loadCollide:   func(obj *projectileTestObject537770) int { return obj.callback },
		callCollide: func(callback int, first, second *projectileTestObject537770, normal *types.Pointf) {
			calls = append(calls, call{callback: callback, first: first, second: second, normal: *normal})
			if first == source {
				hit.callback = 3
			}
		},
		setTraceReady: func(value uint32) { readyWrites = append(readyWrites, value) },
	}
	projectileCollisionDispatch537770(source, hooks)
	wantCalls := []call{
		{callback: 1, first: source, second: hit, normal: types.Ptf(2, -4)},
		{callback: 3, first: hit, second: source, normal: types.Ptf(-2, 4)},
	}
	if !reflect.DeepEqual(calls, wantCalls) {
		t.Fatalf("calls = %#v, want %#v", calls, wantCalls)
	}
	if !reflect.DeepEqual(readyWrites, []uint32{0, 0}) {
		t.Fatalf("trace-ready writes = %v", readyWrites)
	}
}

func TestProjectileCollisionNative537770PreservesPointerWidth(t *testing.T) {
	if unsafe.Sizeof(uintptr(0)) <= 4 {
		t.Skip("native pointer-width assertion is specific to 64-bit hosts")
	}
	server := &Server{}
	server.Types.fast.projectileSmallFist537770 = 11
	server.Types.fast.projectileMediumFist537770 = 12
	server.Types.fast.projectileLargeFist537770 = 13
	var sourceToken, targetToken byte
	source := &Object{Collide: unsafe.Pointer(&sourceToken)}
	target := &Object{TypeInd: 20, Collide: unsafe.Pointer(&targetToken)}
	type call struct {
		callback unsafe.Pointer
		first    uintptr
		second   uintptr
		normal   uintptr
	}
	var calls []call
	server.projectileCollisionDispatchNative537770(source, projectileCollisionNativeDeps537770{
		traceHit: func(*Object) (*Object, types.Pointf, bool) {
			return target, types.Ptf(1, 2), true
		},
		setTraceReady: func(uint32) {},
		callCollide: func(callback unsafe.Pointer, first, second, normal uintptr) {
			calls = append(calls, call{callback: callback, first: first, second: second, normal: normal})
		},
	})
	if len(calls) != 2 {
		t.Fatalf("calls = %d", len(calls))
	}
	if calls[0].first != uintptr(unsafe.Pointer(source)) || calls[0].second != uintptr(unsafe.Pointer(target)) ||
		calls[1].first != uintptr(unsafe.Pointer(target)) || calls[1].second != uintptr(unsafe.Pointer(source)) {
		t.Fatalf("object arguments truncated: %#v", calls)
	}
	if calls[0].first <= math.MaxUint32 || calls[0].second <= math.MaxUint32 {
		t.Fatalf("test objects unexpectedly fit ABI32: %#x %#x", calls[0].first, calls[0].second)
	}
	if calls[0].normal == 0 || calls[1].normal == 0 {
		t.Fatal("normal pointer was not forwarded")
	}
}
