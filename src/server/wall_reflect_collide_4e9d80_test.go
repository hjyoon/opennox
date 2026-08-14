package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type wallReflectTestData4E9D80 struct {
	damage int32
}

type wallReflectTestObject4E9D80 struct {
	data    *wallReflectTestData4E9D80
	collide int
	x       float32
	y       float32
}

func TestWallReflectCollide4E9D80QuestYellowStarOrderAndInt32Return(t *testing.T) {
	oldData := &wallReflectTestData4E9D80{damage: math.MaxInt32}
	newData := &wallReflectTestData4E9D80{damage: 19}
	source := &wallReflectTestObject4E9D80{data: oldData, collide: 77}
	target := &wallReflectTestObject4E9D80{}
	parent := &wallReflectTestObject4E9D80{}
	collision := &struct{ unread int }{unread: 99}
	var events []string

	wallReflectCollide4E9D80(source, target, collision, wallReflectCollideHooks4E9D80[
		*wallReflectTestObject4E9D80,
		*struct{ unread int },
		int,
		*wallReflectTestData4E9D80,
	]{
		loadCollideData: func(got *wallReflectTestObject4E9D80) *wallReflectTestData4E9D80 {
			events = append(events, "data")
			if got != source {
				t.Fatalf("data source = %p, want %p", got, source)
			}
			return got.data
		},
		sameTeam: func(first, second *wallReflectTestObject4E9D80) int32 {
			events = append(events, "team")
			if first != source || second != target {
				t.Fatalf("team args = %p/%p", first, second)
			}
			source.data = newData
			return 0
		},
		gameFlagsCheck: func(flag uint32) int32 {
			events = append(events, fmt.Sprintf("flag:%#x", flag))
			return -1
		},
		loadCollide: func(got *wallReflectTestObject4E9D80) int {
			events = append(events, "collide")
			return got.collide
		},
		yellowStarCollide: 77,
		loadDamage: func(data *wallReflectTestData4E9D80) int32 {
			events = append(events, "damage")
			if data != oldData {
				t.Fatalf("damage data = %p, want cached %p", data, oldData)
			}
			return data.damage
		},
		findParent: func(got *wallReflectTestObject4E9D80) *wallReflectTestObject4E9D80 {
			events = append(events, "parent")
			if got != source {
				t.Fatalf("parent source = %p", got)
			}
			return parent
		},
		targetDamage: func(gotTarget, gotParent, gotSource *wallReflectTestObject4E9D80, damage int32, damageType uint32) int32 {
			events = append(events, "target-damage")
			if gotTarget != target || gotParent != parent || gotSource != source ||
				damage != math.MaxInt32-2 || damageType != wallReflectDamageType4E9D80 {
				t.Fatalf("damage args = %p/%p/%p/%d/%d", gotTarget, gotParent, gotSource, damage, damageType)
			}
			// The original tests the complete EAX value, not just AL.
			return 0x100
		},
		delayedDelete: func(got *wallReflectTestObject4E9D80) {
			events = append(events, "delete")
			if got != source {
				t.Fatalf("delete source = %p", got)
			}
		},
	})

	want := []string{"data", "team", "flag:0x1000", "collide", "damage", "parent", "target-damage", "delete"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %#v, want %#v", events, want)
	}
}

func TestWallReflectCollide4E9D80TargetShortCircuits(t *testing.T) {
	tests := []struct {
		name     string
		sameTeam int32
		quest    int32
		want     []string
	}{
		{name: "same team", sameTeam: -7, want: []string{"data", "team"}},
		{name: "non quest", quest: 0, want: []string{"data", "team", "flag", "damage", "parent", "target-damage"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var events []string
			wallReflectCollide4E9D80(1, 2, 99, wallReflectCollideHooks4E9D80[int, int, int, int]{
				loadCollideData: func(int) int { events = append(events, "data"); return 13 },
				sameTeam: func(int, int) int32 {
					events = append(events, "team")
					return tc.sameTeam
				},
				gameFlagsCheck: func(uint32) int32 {
					events = append(events, "flag")
					return tc.quest
				},
				loadCollide: func(int) int {
					t.Fatal("Collide read outside enabled Quest mode")
					return 0
				},
				yellowStarCollide: 7,
				loadDamage: func(data int) int32 {
					events = append(events, "damage")
					return int32(data)
				},
				findParent: func(int) int { events = append(events, "parent"); return 3 },
				targetDamage: func(_, _, _ int, damage int32, damageType uint32) int32 {
					events = append(events, "target-damage")
					if damage != 13 || damageType != wallReflectDamageType4E9D80 {
						t.Fatalf("damage = %d/%d", damage, damageType)
					}
					return 0
				},
				delayedDelete: func(int) { t.Fatal("zero damage return deleted source") },
				wallReflect:   func(int, int) { t.Fatal("target branch read collision") },
			})
			if !reflect.DeepEqual(events, tc.want) {
				t.Fatalf("events = %#v, want %#v", events, tc.want)
			}
		})
	}
}

func TestWallReflectCollide4E9D80WallReloadOrderAfterReflection(t *testing.T) {
	oldData := &wallReflectTestData4E9D80{damage: 5}
	newData := &wallReflectTestData4E9D80{damage: 99}
	source := &wallReflectTestObject4E9D80{data: oldData}
	collision := &struct{}{}
	var events []string
	convertCalls := 0

	wallReflectCollide4E9D80(source, nil, collision, wallReflectCollideHooks4E9D80[
		*wallReflectTestObject4E9D80,
		*struct{},
		int,
		*wallReflectTestData4E9D80,
	]{
		loadCollideData: func(got *wallReflectTestObject4E9D80) *wallReflectTestData4E9D80 {
			events = append(events, "data")
			return got.data
		},
		wallReflect: func(gotCollision *struct{}, gotSource *wallReflectTestObject4E9D80) {
			events = append(events, "reflect")
			if gotCollision != collision || gotSource != source {
				t.Fatalf("reflect args = %p/%p", gotCollision, gotSource)
			}
			source.data = newData
			oldData.damage = -17
			source.y = 46
			source.x = 69
		},
		loadNewPosY: func(got *wallReflectTestObject4E9D80) float32 {
			events = append(events, "y")
			return got.y
		},
		loadDamage: func(data *wallReflectTestData4E9D80) int32 {
			events = append(events, "damage")
			if data != oldData {
				t.Fatalf("wall damage data = %p, want cached %p", data, oldData)
			}
			return data.damage
		},
		floatToInt: func(value float32) int32 {
			convertCalls++
			if convertCalls == 1 {
				events = append(events, "round-y")
				if value != 2 {
					t.Fatalf("Y grid input = %g, want 2", value)
				}
				// X is loaded after the first conversion returns.
				source.x = -92
				return 2
			}
			events = append(events, "round-x")
			if value != -4 {
				t.Fatalf("X grid input = %g, want -4", value)
			}
			return -4
		},
		loadNewPosX: func(got *wallReflectTestObject4E9D80) float32 {
			events = append(events, "x")
			return got.x
		},
		damageMap: func(x, y, damage int32, damageType uint32, gotSource *wallReflectTestObject4E9D80) {
			events = append(events, "map")
			if x != -4 || y != 2 || damage != -17 || damageType != wallReflectDamageType4E9D80 || gotSource != source {
				t.Fatalf("map args = %d/%d/%d/%d/%p", x, y, damage, damageType, gotSource)
			}
		},
		delayedDelete: func(*wallReflectTestObject4E9D80) {
			t.Fatal("wall branch deleted source")
		},
	})

	want := []string{"data", "reflect", "y", "damage", "round-y", "x", "round-x", "map"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %#v, want %#v", events, want)
	}
}

func TestWallReflectCollide4E9D80NilWallStillCachesData(t *testing.T) {
	var events []string
	wallReflectCollide4E9D80(1, 0, 0, wallReflectCollideHooks4E9D80[int, int, int, int]{
		loadCollideData: func(int) int { events = append(events, "data"); return 7 },
		wallReflect:     func(int, int) { t.Fatal("nil collision reflected") },
	})
	if !reflect.DeepEqual(events, []string{"data"}) {
		t.Fatalf("events = %#v", events)
	}
}

func TestYellowStarShotCollide4E9E50FlagFXAndForwarding(t *testing.T) {
	tests := []struct {
		name   string
		source int
		flag   int32
		want   []string
	}{
		{name: "fx enabled", source: 1, want: []string{"flag:0x4", "fx:136:1", "wall:1:2:3"}},
		{name: "fx suppressed", source: 1, flag: -1, want: []string{"flag:0x4", "wall:1:2:3"}},
		{name: "nil source forwarded", source: 0, want: []string{"wall:0:2:3"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var events []string
			yellowStarShotCollide4E9E50(tc.source, 2, 3, yellowStarShotCollideHooks4E9E50[int, int]{
				gameFlagsCheck: func(flag uint32) int32 {
					events = append(events, fmt.Sprintf("flag:%#x", flag))
					return tc.flag
				},
				pointFX: func(id uint32, source int) {
					events = append(events, fmt.Sprintf("fx:%d:%d", id, source))
				},
				wallCollide: func(source, target, collision int) {
					events = append(events, fmt.Sprintf("wall:%d:%d:%d", source, target, collision))
				},
			})
			if !reflect.DeepEqual(events, tc.want) {
				t.Fatalf("events = %#v, want %#v", events, tc.want)
			}
		})
	}
}
