package server

import (
	"math"
	"reflect"
	"testing"
)

type deathBallFragmentTestObject4E9FE0 struct {
	name          string
	newX          float32
	newY          float32
	damageVersion int
}

func TestDeathBallFragmentCollide4E9FE0TargetOrderArgumentsAndDelete(t *testing.T) {
	source := &deathBallFragmentTestObject4E9FE0{name: "source"}
	target := &deathBallFragmentTestObject4E9FE0{name: "target", damageVersion: 1}
	parent := &deathBallFragmentTestObject4E9FE0{name: "parent"}
	events := make([]string, 0, 3)

	deathBallFragmentCollide4E9FE0(
		source,
		target,
		77,
		deathBallFragmentCollideHooks4E9FE0[*deathBallFragmentTestObject4E9FE0, int]{
			findParent: func(got *deathBallFragmentTestObject4E9FE0) *deathBallFragmentTestObject4E9FE0 {
				events = append(events, "parent")
				if got != source {
					t.Fatalf("parent source = %p, want %p", got, source)
				}
				target.damageVersion = 2
				return parent
			},
			targetDamage: func(gotTarget, gotParent, gotSource *deathBallFragmentTestObject4E9FE0, damage int32, damageType uint32) int32 {
				events = append(events, "damage")
				if gotTarget != target || gotParent != parent || gotSource != source {
					t.Fatalf("damage objects = %p/%p/%p", gotTarget, gotParent, gotSource)
				}
				if damage != deathBallFragmentDamage4E9FE0 || damageType != deathBallFragmentDamageType4E9FE0 {
					t.Fatalf("damage/type = %d/%d", damage, damageType)
				}
				if target.damageVersion != 2 {
					t.Fatalf("Damage callback was not observed after parent lookup")
				}
				return math.MinInt32
			},
			wallReflect: func(int, *deathBallFragmentTestObject4E9FE0) {
				t.Fatal("target path read collision")
			},
			audio: func(uint32, *deathBallFragmentTestObject4E9FE0) {
				t.Fatal("target path emitted audio")
			},
			loadNewPosY: func(*deathBallFragmentTestObject4E9FE0) float32 {
				t.Fatal("target path read NewPos.Y")
				return 0
			},
			loadNewPosX: func(*deathBallFragmentTestObject4E9FE0) float32 {
				t.Fatal("target path read NewPos.X")
				return 0
			},
			floatToInt: func(float32) int32 {
				t.Fatal("target path converted a map coordinate")
				return 0
			},
			damageMap: func(int32, int32, int32, uint32, *deathBallFragmentTestObject4E9FE0) {
				t.Fatal("target path damaged map")
			},
			delayedDelete: func(got *deathBallFragmentTestObject4E9FE0) {
				events = append(events, "delete")
				if got != source {
					t.Fatalf("deleted = %p, want %p", got, source)
				}
			},
		},
	)

	if want := []string{"parent", "damage", "delete"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestDeathBallFragmentCollide4E9FE0WallOrderLiveCoordinatesAndNoDelete(t *testing.T) {
	source := &deathBallFragmentTestObject4E9FE0{name: "source", newX: 1, newY: 2}
	const collision = 73
	events := make([]string, 0, 7)
	inputs := make([]uint32, 0, 2)
	results := []int32{-17, 29}
	var gotMap struct {
		x, y, damage int32
		damageType   uint32
		source       *deathBallFragmentTestObject4E9FE0
	}

	deathBallFragmentCollide4E9FE0(
		source,
		nil,
		collision,
		deathBallFragmentCollideHooks4E9FE0[*deathBallFragmentTestObject4E9FE0, int]{
			findParent: func(*deathBallFragmentTestObject4E9FE0) *deathBallFragmentTestObject4E9FE0 {
				t.Fatal("wall path looked up parent")
				return nil
			},
			targetDamage: func(*deathBallFragmentTestObject4E9FE0, *deathBallFragmentTestObject4E9FE0, *deathBallFragmentTestObject4E9FE0, int32, uint32) int32 {
				t.Fatal("wall path damaged target")
				return 0
			},
			wallReflect: func(gotCollision int, gotSource *deathBallFragmentTestObject4E9FE0) {
				events = append(events, "reflect")
				if gotCollision != collision || gotSource != source {
					t.Fatalf("reflect = %d/%p", gotCollision, gotSource)
				}
				source.newY = 46
			},
			audio: func(id uint32, gotSource *deathBallFragmentTestObject4E9FE0) {
				events = append(events, "audio")
				if id != deathBallFragmentWallAudio4E9FE0 || gotSource != source {
					t.Fatalf("audio = %d/%p", id, gotSource)
				}
				source.newY = 69
			},
			loadNewPosY: func(got *deathBallFragmentTestObject4E9FE0) float32 {
				events = append(events, "new-y")
				return got.newY
			},
			loadNewPosX: func(got *deathBallFragmentTestObject4E9FE0) float32 {
				events = append(events, "new-x")
				return got.newX
			},
			floatToInt: func(value float32) int32 {
				events = append(events, "float")
				inputs = append(inputs, math.Float32bits(value))
				if len(inputs) == 1 {
					source.newX = 92
				}
				result := results[0]
				results = results[1:]
				return result
			},
			damageMap: func(x, y, damage int32, damageType uint32, gotSource *deathBallFragmentTestObject4E9FE0) {
				events = append(events, "map")
				gotMap.x, gotMap.y, gotMap.damage = x, y, damage
				gotMap.damageType, gotMap.source = damageType, gotSource
			},
			delayedDelete: func(*deathBallFragmentTestObject4E9FE0) {
				t.Fatal("wall path deleted fragment")
			},
		},
	)

	wantEvents := []string{"reflect", "audio", "new-y", "float", "new-x", "float", "map"}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %v, want %v", events, wantEvents)
	}
	gridInverse := math.Float32frombits(deathBallFragmentGridInverseBits4E9FE0)
	wantInputs := []uint32{
		math.Float32bits(float32(69) * gridInverse),
		math.Float32bits(float32(92) * gridInverse),
	}
	if !reflect.DeepEqual(inputs, wantInputs) {
		t.Fatalf("conversion inputs = %#v, want %#v", inputs, wantInputs)
	}
	if gotMap.x != 29 || gotMap.y != -17 || gotMap.damage != deathBallFragmentDamage4E9FE0 ||
		gotMap.damageType != deathBallFragmentDamageType4E9FE0 || gotMap.source != source {
		t.Fatalf("map = %#v", gotMap)
	}
}

func TestDeathBallFragmentCollide4E9FE0NoTargetOrWallDeletesOnly(t *testing.T) {
	events := make([]string, 0, 1)
	hooks := deathBallFragmentCollideHooks4E9FE0[int, int]{
		findParent: func(int) int { t.Fatal("looked up parent"); return 0 },
		targetDamage: func(int, int, int, int32, uint32) int32 {
			t.Fatal("damaged target")
			return 0
		},
		wallReflect: func(int, int) { t.Fatal("reflected wall") },
		audio:       func(uint32, int) { t.Fatal("emitted audio") },
		loadNewPosY: func(int) float32 { t.Fatal("read Y"); return 0 },
		loadNewPosX: func(int) float32 { t.Fatal("read X"); return 0 },
		floatToInt:  func(float32) int32 { t.Fatal("converted coordinate"); return 0 },
		damageMap:   func(int32, int32, int32, uint32, int) { t.Fatal("damaged map") },
		delayedDelete: func(source int) {
			events = append(events, "delete")
			if source != 11 {
				t.Fatalf("deleted source = %d", source)
			}
		},
	}

	deathBallFragmentCollide4E9FE0(11, 0, 0, hooks)
	if !reflect.DeepEqual(events, []string{"delete"}) {
		t.Fatalf("events = %v", events)
	}
}
