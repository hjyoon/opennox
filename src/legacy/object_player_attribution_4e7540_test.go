package legacy

import (
	"reflect"
	"testing"
)

type attributionObject4E7540 struct {
	name   string
	class  uint32
	update *attributionUpdate4E7540
}

type attributionUpdate4E7540 struct {
	name   string
	player *attributionPlayer4E7540
}

type attributionPlayer4E7540 struct {
	name        string
	index       byte
	pending     uint32
	aggressor   uint32
	frame       uint32
	neighboring uint32
}

func attributionHooks4E7540(events *[]string) playerAttributionHooks4E7540[
	*attributionObject4E7540,
	*attributionUpdate4E7540,
	*attributionPlayer4E7540,
] {
	return playerAttributionHooks4E7540[*attributionObject4E7540, *attributionUpdate4E7540, *attributionPlayer4E7540]{
		class: func(obj *attributionObject4E7540) uint32 {
			*events = append(*events, "class:"+obj.name)
			return obj.class
		},
		updateData: func(obj *attributionObject4E7540) *attributionUpdate4E7540 {
			*events = append(*events, "update:"+obj.name)
			return obj.update
		},
		player: func(update *attributionUpdate4E7540) *attributionPlayer4E7540 {
			if update == nil {
				*events = append(*events, "player:nil")
				panic("nil update data")
			}
			*events = append(*events, "player:"+update.name+":"+update.player.name)
			return update.player
		},
		playerIndex: func(player *attributionPlayer4E7540) byte {
			*events = append(*events, "index:"+player.name)
			return player.index
		},
		setPlayerIndex: func(player *attributionPlayer4E7540, index uint32) {
			*events = append(*events, "set-index:"+player.name)
			player.aggressor = index
		},
		frame: func() uint32 {
			*events = append(*events, "frame")
			return 0x89abcdef
		},
		setFrame: func(player *attributionPlayer4E7540, frame uint32) {
			*events = append(*events, "set-frame:"+player.name)
			player.frame = frame
		},
		setPending: func(player *attributionPlayer4E7540, pending uint32) {
			*events = append(*events, "set-pending:"+player.name)
			player.pending = pending
		},
	}
}

func TestRecordPlayerAttribution4E7540GuardOrder(t *testing.T) {
	player := &attributionPlayer4E7540{name: "player"}
	update := &attributionUpdate4E7540{name: "update", player: player}
	playerSource := &attributionObject4E7540{name: "source", class: 4, update: update}
	playerTarget := &attributionObject4E7540{name: "target", class: 4, update: update}
	nonPlayerSource := &attributionObject4E7540{name: "source", class: 0x10002, update: update}
	nonPlayerTarget := &attributionObject4E7540{name: "target", class: 0x20002, update: update}

	for _, tc := range []struct {
		name           string
		source, target *attributionObject4E7540
		want           []string
	}{
		{name: "nil source", source: nil, target: playerTarget},
		{name: "nil target", source: playerSource, target: nil},
		{name: "source is not player", source: nonPlayerSource, target: playerTarget, want: []string{"class:source"}},
		{name: "target is not player", source: playerSource, target: nonPlayerTarget, want: []string{"class:source", "class:target"}},
		{name: "same player object", source: playerSource, target: playerSource, want: []string{"class:source", "class:source"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var events []string
			recordPlayerAttribution4E7540(tc.source, tc.target, attributionHooks4E7540(&events))
			if !reflect.DeepEqual(events, tc.want) {
				t.Fatalf("events = %v, want %v", events, tc.want)
			}
		})
	}
}

func TestRecordPlayerAttribution4E7540ReloadsTargetPlayerForEveryStore(t *testing.T) {
	sourcePlayer := &attributionPlayer4E7540{name: "source", index: 0xfe, neighboring: 0x11111111}
	firstTarget := &attributionPlayer4E7540{name: "target-1", neighboring: 0x22222222}
	secondTarget := &attributionPlayer4E7540{name: "target-2", neighboring: 0x33333333}
	thirdTarget := &attributionPlayer4E7540{name: "target-3", neighboring: 0x44444444}
	sourceUpdate := &attributionUpdate4E7540{name: "source", player: sourcePlayer}
	targetUpdate := &attributionUpdate4E7540{name: "target", player: firstTarget}
	source := &attributionObject4E7540{name: "source", class: 0xff000004, update: sourceUpdate}
	target := &attributionObject4E7540{name: "target", class: 0xaa000004, update: targetUpdate}
	var events []string
	hooks := attributionHooks4E7540(&events)
	originalSetIndex := hooks.setPlayerIndex
	hooks.setPlayerIndex = func(player *attributionPlayer4E7540, index uint32) {
		originalSetIndex(player, index)
		targetUpdate.player = secondTarget
	}
	originalSetFrame := hooks.setFrame
	hooks.setFrame = func(player *attributionPlayer4E7540, frame uint32) {
		originalSetFrame(player, frame)
		targetUpdate.player = thirdTarget
	}

	recordPlayerAttribution4E7540(source, target, hooks)
	wantEvents := []string{
		"class:source", "class:target",
		"update:source", "update:target",
		"player:source:source", "index:source",
		"player:target:target-1", "set-index:target-1",
		"player:target:target-2", "frame", "set-frame:target-2",
		"player:target:target-3", "set-pending:target-3",
	}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %v, want %v", events, wantEvents)
	}
	if firstTarget.aggressor != 0xfe || firstTarget.frame != 0 || firstTarget.pending != 0 {
		t.Fatalf("first target state = (%#x, %#x, %#x), want (0xfe, 0, 0)", firstTarget.aggressor, firstTarget.frame, firstTarget.pending)
	}
	if secondTarget.aggressor != 0 || secondTarget.frame != 0x89abcdef || secondTarget.pending != 0 {
		t.Fatalf("second target state = (%#x, %#x, %#x), want (0, 0x89abcdef, 0)", secondTarget.aggressor, secondTarget.frame, secondTarget.pending)
	}
	if thirdTarget.aggressor != 0 || thirdTarget.frame != 0 || thirdTarget.pending != 1 {
		t.Fatalf("third target state = (%#x, %#x, %#x), want (0, 0, 1)", thirdTarget.aggressor, thirdTarget.frame, thirdTarget.pending)
	}
	if sourcePlayer.neighboring != 0x11111111 || firstTarget.neighboring != 0x22222222 || secondTarget.neighboring != 0x33333333 || thirdTarget.neighboring != 0x44444444 {
		t.Fatal("attribution changed an adjacent player field")
	}
}

func TestRecordPlayerAttribution4E7540LoadsBothUpdatesBeforeSourcePlayer(t *testing.T) {
	targetPlayer := &attributionPlayer4E7540{name: "target"}
	source := &attributionObject4E7540{name: "source", class: 4}
	target := &attributionObject4E7540{
		name:   "target",
		class:  4,
		update: &attributionUpdate4E7540{name: "target", player: targetPlayer},
	}
	var events []string
	defer func() {
		if recover() == nil {
			t.Fatal("nil source update returned without a panic")
		}
		want := []string{"class:source", "class:target", "update:source", "update:target", "player:nil"}
		if !reflect.DeepEqual(events, want) {
			t.Fatalf("events = %v, want %v", events, want)
		}
	}()
	recordPlayerAttribution4E7540(source, target, attributionHooks4E7540(&events))
}
