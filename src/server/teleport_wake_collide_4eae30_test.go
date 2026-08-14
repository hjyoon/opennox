package server

import (
	"fmt"
	"reflect"
	"testing"
)

type teleportWakeTestData4EAE30 struct {
	name string
	x    float32
	y    float32
}

type teleportWakeTestPosition4EAE30 struct {
	name string
}

type teleportWakeTestObject4EAE30 struct {
	name      string
	data      *teleportWakeTestData4EAE30
	owner     *teleportWakeTestObject4EAE30
	class     uint32
	anchored  bool
	invisible bool
	position  *teleportWakeTestPosition4EAE30
}

func TestTeleportWakeCollide4EAE30SuccessOrderAndLiveReloads(t *testing.T) {
	data := &teleportWakeTestData4EAE30{name: "entry", x: 10, y: 20}
	replacement := &teleportWakeTestData4EAE30{name: "replacement", x: 30, y: 40}
	entryPosition := &teleportWakeTestPosition4EAE30{name: "entry"}
	postPosition := &teleportWakeTestPosition4EAE30{name: "post"}
	oldOwner := &teleportWakeTestObject4EAE30{name: "old-owner"}
	playerOwner := &teleportWakeTestObject4EAE30{name: "player-owner", class: 0x91020304}
	source := &teleportWakeTestObject4EAE30{name: "source", data: data, owner: oldOwner}
	target := &teleportWakeTestObject4EAE30{name: "target", position: entryPosition}
	collision := &struct{ unread [2]uint32 }{unread: [2]uint32{0xdeadbeef, 0x89abcdef}}

	var events []string
	invisibleCalls := 0
	audioCalls := 0
	positionCalls := 0

	hooks := teleportWakeCollideHooks4EAE30[
		*teleportWakeTestObject4EAE30,
		*teleportWakeTestData4EAE30,
		*teleportWakeTestPosition4EAE30,
	]{
		loadCollideData: func(obj *teleportWakeTestObject4EAE30) *teleportWakeTestData4EAE30 {
			events = append(events, "data")
			return obj.data
		},
		hasEnchant: func(obj *teleportWakeTestObject4EAE30, enchant uint32) bool {
			switch enchant {
			case teleportWakeAnchoredEnchant4EAE30:
				events = append(events, "anchored")
				return obj.anchored
			case teleportWakeInvisibleEnchant4EAE30:
				invisibleCalls++
				events = append(events, fmt.Sprintf("invisible-%d", invisibleCalls))
				return obj.invisible
			default:
				t.Fatalf("unexpected enchant %d", enchant)
				return false
			}
		},
		questMode: func() bool {
			events = append(events, "quest")
			source.owner = playerOwner
			target.class = teleportWakeTargetClassMask4EAE30
			return true
		},
		loadOwner: func(obj *teleportWakeTestObject4EAE30) *teleportWakeTestObject4EAE30 {
			events = append(events, "owner")
			return obj.owner
		},
		loadOwnerClassLo: func(obj *teleportWakeTestObject4EAE30) uint8 {
			events = append(events, "owner-class-low")
			return uint8(obj.class)
		},
		loadTargetClass: func(obj *teleportWakeTestObject4EAE30) uint32 {
			events = append(events, "target-class")
			return obj.class
		},
		position: func(obj *teleportWakeTestObject4EAE30) *teleportWakeTestPosition4EAE30 {
			positionCalls++
			events = append(events, fmt.Sprintf("position-%d", positionCalls))
			return obj.position
		},
		pointFX: func(id uint32, pos *teleportWakeTestPosition4EAE30) {
			events = append(events, fmt.Sprintf("fx-%d", id))
			switch id {
			case teleportWakePreFX4EAE30:
				if pos != entryPosition {
					t.Fatalf("pre position = %p, want %p", pos, entryPosition)
				}
				source.data = replacement
				target.class = 0
				target.invisible = true
			case teleportWakePostFX4EAE30:
				if pos != postPosition {
					t.Fatalf("post position = %p, want live %p", pos, postPosition)
				}
			default:
				t.Fatalf("point FX = %d", id)
			}
		},
		audio: func(id uint32, obj *teleportWakeTestObject4EAE30) {
			audioCalls++
			events = append(events, fmt.Sprintf("audio-%d", audioCalls))
			if id != teleportWakeSound4EAE30 || obj != target {
				t.Fatalf("audio = %d/%p", id, obj)
			}
			if audioCalls == 1 {
				data.x, data.y = -123.5, 456.25
			}
		},
		teleport: func(obj *teleportWakeTestObject4EAE30, got *teleportWakeTestData4EAE30) {
			events = append(events, "teleport")
			if obj != target || got != data {
				t.Fatalf("teleport = %p/%p, want target/cached data %p/%p", obj, got, target, data)
			}
			if got.x != -123.5 || got.y != 456.25 {
				t.Fatalf("teleport data = %+v; cached pointer did not retain live contents", got)
			}
			obj.position = postPosition
			obj.invisible = false
		},
	}

	teleportWakeCollide4EAE30(source, target, collision, hooks)

	want := []string{
		"data", "anchored", "quest", "owner", "owner-class-low", "target-class",
		"invisible-1", "position-1", "fx-138", "audio-1", "teleport",
		"invisible-2", "position-2", "fx-137", "audio-2",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %#v, want %#v", events, want)
	}
	if source.data != replacement || target.class != 0 {
		t.Fatal("callback mutations were unexpectedly overwritten")
	}
}

func TestTeleportWakeCollide4EAE30GateOrder(t *testing.T) {
	tests := []struct {
		name        string
		targetNil   bool
		anchored    bool
		quest       bool
		ownerNil    bool
		ownerClass  uint32
		targetClass uint32
		want        []string
	}{
		{
			name:      "nil target after data cache",
			targetNil: true,
			want:      []string{"data"},
		},
		{
			name:     "anchored before quest",
			anchored: true,
			want:     []string{"data", "anchored"},
		},
		{
			name:        "quest non-player owner",
			quest:       true,
			ownerClass:  0xfffffff8,
			targetClass: teleportWakeTargetClassMask4EAE30,
			want:        []string{"data", "anchored", "quest", "owner", "owner-class-low"},
		},
		{
			name:        "quest nil owner reaches target class",
			quest:       true,
			ownerNil:    true,
			targetClass: 0,
			want:        []string{"data", "anchored", "quest", "owner", "target-class"},
		},
		{
			name:        "non-quest skips owner and rejects target class",
			targetClass: 0x80000000,
			want:        []string{"data", "anchored", "quest", "target-class"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			data := &teleportWakeTestData4EAE30{name: "data"}
			owner := &teleportWakeTestObject4EAE30{name: "owner", class: tc.ownerClass}
			if tc.ownerNil {
				owner = nil
			}
			source := &teleportWakeTestObject4EAE30{name: "source", data: data, owner: owner}
			target := &teleportWakeTestObject4EAE30{
				name:     "target",
				class:    tc.targetClass,
				anchored: tc.anchored,
			}
			if tc.targetNil {
				target = nil
			}
			var events []string
			hooks := teleportWakeCollideHooks4EAE30[
				*teleportWakeTestObject4EAE30,
				*teleportWakeTestData4EAE30,
				*teleportWakeTestPosition4EAE30,
			]{
				loadCollideData: func(obj *teleportWakeTestObject4EAE30) *teleportWakeTestData4EAE30 {
					events = append(events, "data")
					return obj.data
				},
				hasEnchant: func(obj *teleportWakeTestObject4EAE30, enchant uint32) bool {
					if enchant != teleportWakeAnchoredEnchant4EAE30 {
						t.Fatalf("unexpected success-path enchant %d", enchant)
					}
					events = append(events, "anchored")
					return obj.anchored
				},
				questMode: func() bool {
					events = append(events, "quest")
					return tc.quest
				},
				loadOwner: func(obj *teleportWakeTestObject4EAE30) *teleportWakeTestObject4EAE30 {
					events = append(events, "owner")
					return obj.owner
				},
				loadOwnerClassLo: func(obj *teleportWakeTestObject4EAE30) uint8 {
					events = append(events, "owner-class-low")
					return uint8(obj.class)
				},
				loadTargetClass: func(obj *teleportWakeTestObject4EAE30) uint32 {
					events = append(events, "target-class")
					return obj.class
				},
			}

			teleportWakeCollide4EAE30(source, target, (*int)(nil), hooks)
			if !reflect.DeepEqual(events, tc.want) {
				t.Fatalf("events = %#v, want %#v", events, tc.want)
			}
		})
	}
}

func TestTeleportWakeCollide4EAE30RechecksInvisibility(t *testing.T) {
	tests := []struct {
		name      string
		invisible [2]bool
		want      []string
	}{
		{
			name:      "pre hidden post visible",
			invisible: [2]bool{true, false},
			want:      []string{"invisible-1", "audio", "teleport", "invisible-2", "position", "fx-137", "audio"},
		},
		{
			name:      "pre visible post hidden",
			invisible: [2]bool{false, true},
			want:      []string{"invisible-1", "position", "fx-138", "audio", "teleport", "invisible-2", "audio"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			source := &teleportWakeTestObject4EAE30{data: &teleportWakeTestData4EAE30{}}
			target := &teleportWakeTestObject4EAE30{
				class:    teleportWakeTargetClassMask4EAE30,
				position: &teleportWakeTestPosition4EAE30{name: "position"},
			}
			var events []string
			invisibleCall := 0
			hooks := teleportWakeCollideHooks4EAE30[
				*teleportWakeTestObject4EAE30,
				*teleportWakeTestData4EAE30,
				*teleportWakeTestPosition4EAE30,
			]{
				loadCollideData: func(obj *teleportWakeTestObject4EAE30) *teleportWakeTestData4EAE30 {
					return obj.data
				},
				hasEnchant: func(_ *teleportWakeTestObject4EAE30, enchant uint32) bool {
					if enchant == teleportWakeAnchoredEnchant4EAE30 {
						return false
					}
					invisibleCall++
					events = append(events, fmt.Sprintf("invisible-%d", invisibleCall))
					return tc.invisible[invisibleCall-1]
				},
				questMode: func() bool { return false },
				loadTargetClass: func(obj *teleportWakeTestObject4EAE30) uint32 {
					return obj.class
				},
				position: func(obj *teleportWakeTestObject4EAE30) *teleportWakeTestPosition4EAE30 {
					events = append(events, "position")
					return obj.position
				},
				pointFX: func(id uint32, _ *teleportWakeTestPosition4EAE30) {
					events = append(events, fmt.Sprintf("fx-%d", id))
				},
				audio: func(uint32, *teleportWakeTestObject4EAE30) {
					events = append(events, "audio")
				},
				teleport: func(*teleportWakeTestObject4EAE30, *teleportWakeTestData4EAE30) {
					events = append(events, "teleport")
				},
			}

			teleportWakeCollide4EAE30(source, target, &struct{ unread bool }{unread: true}, hooks)
			if !reflect.DeepEqual(events, tc.want) {
				t.Fatalf("events = %#v, want %#v", events, tc.want)
			}
		})
	}
}

func TestTeleportWakeCollide4EAE30NilSourceFaultsBeforeNilTarget(t *testing.T) {
	var events []string
	hooks := teleportWakeCollideHooks4EAE30[
		*teleportWakeTestObject4EAE30,
		*teleportWakeTestData4EAE30,
		*teleportWakeTestPosition4EAE30,
	]{
		loadCollideData: func(obj *teleportWakeTestObject4EAE30) *teleportWakeTestData4EAE30 {
			events = append(events, "data")
			return obj.data
		},
	}
	var recovered any
	func() {
		defer func() { recovered = recover() }()
		teleportWakeCollide4EAE30(
			(*teleportWakeTestObject4EAE30)(nil),
			(*teleportWakeTestObject4EAE30)(nil),
			(*int)(nil),
			hooks,
		)
	}()
	if recovered == nil {
		t.Fatal("nil source did not fault during entry collide-data load")
	}
	if want := []string{"data"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %#v, want %#v", events, want)
	}
}
