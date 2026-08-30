package server

import (
	"reflect"
	"testing"
)

type playerFollowLastTargetTestObject4F9E10 struct {
	name   string
	last   *playerFollowLastTargetTestObject4F9E10
	owner  *playerFollowLastTargetTestObject4F9E10
	flags  uint8
	class  uint8
	update *playerFollowLastTargetTestUpdate4F9E10
}

type playerFollowLastTargetTestUpdate4F9E10 struct {
	player *playerFollowLastTargetTestPlayer4F9E10
}

type playerFollowLastTargetTestPlayer4F9E10 struct {
	status uint32
}

func playerFollowLastTargetTestName4F9E10(obj *playerFollowLastTargetTestObject4F9E10) string {
	if obj == nil {
		return "nil"
	}
	return obj.name
}

func playerFollowLastTargetTestHooks4F9E10(calls *[]string) playerFollowLastTargetHooks4F9E10[
	*playerFollowLastTargetTestObject4F9E10,
	*playerFollowLastTargetTestUpdate4F9E10,
	*playerFollowLastTargetTestPlayer4F9E10,
] {
	return playerFollowLastTargetHooks4F9E10[
		*playerFollowLastTargetTestObject4F9E10,
		*playerFollowLastTargetTestUpdate4F9E10,
		*playerFollowLastTargetTestPlayer4F9E10,
	]{
		loadLastTarget: func(obj *playerFollowLastTargetTestObject4F9E10) *playerFollowLastTargetTestObject4F9E10 {
			*calls = append(*calls, "last:"+playerFollowLastTargetTestName4F9E10(obj))
			return obj.last
		},
		findOwnerChainPlayer: func(obj *playerFollowLastTargetTestObject4F9E10) *playerFollowLastTargetTestObject4F9E10 {
			*calls = append(*calls, "owner-chain:"+playerFollowLastTargetTestName4F9E10(obj))
			for obj.owner != nil && obj.class&playerFollowPlayerClass4F9E10 == 0 {
				obj = obj.owner
			}
			return obj
		},
		loadFlagsByte: func(obj *playerFollowLastTargetTestObject4F9E10) uint8 {
			*calls = append(*calls, "flags:"+obj.name)
			return obj.flags
		},
		loadClassByte: func(obj *playerFollowLastTargetTestObject4F9E10) uint8 {
			*calls = append(*calls, "class:"+obj.name)
			return obj.class
		},
		loadUpdateData: func(obj *playerFollowLastTargetTestObject4F9E10) *playerFollowLastTargetTestUpdate4F9E10 {
			*calls = append(*calls, "update:"+obj.name)
			return obj.update
		},
		loadPlayer: func(update *playerFollowLastTargetTestUpdate4F9E10) *playerFollowLastTargetTestPlayer4F9E10 {
			*calls = append(*calls, "player")
			return update.player
		},
		loadPlayerStatus: func(player *playerFollowLastTargetTestPlayer4F9E10) uint32 {
			*calls = append(*calls, "status")
			return player.status
		},
		cameraFollow: func(unit, target *playerFollowLastTargetTestObject4F9E10) {
			*calls = append(*calls, "follow:"+unit.name+":"+target.name)
		},
	}
}

func TestPlayerFollowLastTarget4F9E10ExactPlayerTrace(t *testing.T) {
	pl := &playerFollowLastTargetTestPlayer4F9E10{status: 0x80000000}
	target := &playerFollowLastTargetTestObject4F9E10{
		name: "target", class: playerFollowPlayerClass4F9E10,
		update: &playerFollowLastTargetTestUpdate4F9E10{player: pl},
	}
	leaf := &playerFollowLastTargetTestObject4F9E10{name: "leaf", owner: target}
	unit := &playerFollowLastTargetTestObject4F9E10{name: "unit", last: leaf}
	var calls []string

	if got := playerFollowLastTarget4F9E10(unit, playerFollowLastTargetTestHooks4F9E10(&calls)); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	want := []string{
		"last:unit", "owner-chain:leaf", "flags:target", "class:target",
		"update:target", "player", "status", "follow:unit:target",
	}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls = %v, want %v", calls, want)
	}
}

func TestPlayerFollowLastTarget4F9E10ShortCircuits(t *testing.T) {
	tests := []struct {
		name       string
		unit       *playerFollowLastTargetTestObject4F9E10
		wantCalls  []string
		wantResult int32
	}{
		{name: "nil unit"},
		{
			name:      "nil attribution",
			unit:      &playerFollowLastTargetTestObject4F9E10{name: "unit"},
			wantCalls: []string{"last:unit"},
		},
		{
			name: "rejected flags",
			unit: &playerFollowLastTargetTestObject4F9E10{name: "unit", last: &playerFollowLastTargetTestObject4F9E10{
				name: "target", flags: playerFollowRejectedFlags4F9E10,
			}},
			wantCalls: []string{"last:unit", "owner-chain:target", "flags:target"},
		},
		{
			name: "monster",
			unit: &playerFollowLastTargetTestObject4F9E10{name: "unit", last: &playerFollowLastTargetTestObject4F9E10{
				name: "target", class: playerFollowMonsterClass4F9E10,
			}},
			wantCalls: []string{"last:unit", "owner-chain:target", "flags:target", "class:target"},
		},
		{
			name: "observing player",
			unit: &playerFollowLastTargetTestObject4F9E10{name: "unit", last: &playerFollowLastTargetTestObject4F9E10{
				name: "target", class: playerFollowPlayerClass4F9E10,
				update: &playerFollowLastTargetTestUpdate4F9E10{
					player: &playerFollowLastTargetTestPlayer4F9E10{status: playerFollowObserverBit4F9E10},
				},
			}},
			wantCalls: []string{
				"last:unit", "owner-chain:target", "flags:target", "class:target",
				"update:target", "player", "status",
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var calls []string
			got := playerFollowLastTarget4F9E10(tc.unit, playerFollowLastTargetTestHooks4F9E10(&calls))
			if got != tc.wantResult {
				t.Fatalf("result = %d, want %d", got, tc.wantResult)
			}
			if !reflect.DeepEqual(calls, tc.wantCalls) {
				t.Fatalf("calls = %v, want %v", calls, tc.wantCalls)
			}
		})
	}
}

func TestPlayerFollowLastTarget4F9E10NonPlayerSkipsPlayerLoads(t *testing.T) {
	target := &playerFollowLastTargetTestObject4F9E10{name: "target", class: 0x80}
	unit := &playerFollowLastTargetTestObject4F9E10{name: "unit", last: target}
	var calls []string
	if got := playerFollowLastTarget4F9E10(unit, playerFollowLastTargetTestHooks4F9E10(&calls)); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	want := []string{"last:unit", "owner-chain:target", "flags:target", "class:target", "follow:unit:target"}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls = %v, want %v", calls, want)
	}
}

func TestPlayerFollowLastTarget4F9E10DoesNotGuardPlayerUpdate(t *testing.T) {
	target := &playerFollowLastTargetTestObject4F9E10{name: "target", class: playerFollowPlayerClass4F9E10}
	unit := &playerFollowLastTargetTestObject4F9E10{name: "unit", last: target}
	defer func() {
		if recover() == nil {
			t.Fatal("nil Player update did not preserve the original fault contract")
		}
	}()
	var calls []string
	_ = playerFollowLastTarget4F9E10(unit, playerFollowLastTargetTestHooks4F9E10(&calls))
}
