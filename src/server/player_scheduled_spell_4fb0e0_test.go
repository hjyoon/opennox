package server

import (
	"fmt"
	"reflect"
	"testing"
)

type playerScheduledSpellTestUnit4FB0E0 struct {
	name string
}

type playerScheduledSpellTestPlayer4FB0E0 struct {
	index uint8
}

type playerScheduledSpellTestUpdate4FB0E0 struct {
	spells [5]uint32
	count  uint32
	posX   int32
	posY   int32
	player *playerScheduledSpellTestPlayer4FB0E0
}

func playerScheduledSpellTestHooks4FB0E0(
	t *testing.T,
	update *playerScheduledSpellTestUpdate4FB0E0,
	events *[]string,
) playerScheduledSpellHooks4FB0E0[
	*playerScheduledSpellTestUnit4FB0E0,
	*playerScheduledSpellTestUpdate4FB0E0,
	*playerScheduledSpellTestPlayer4FB0E0,
] {
	t.Helper()
	record := func(event string) { *events = append(*events, event) }
	return playerScheduledSpellHooks4FB0E0[
		*playerScheduledSpellTestUnit4FB0E0,
		*playerScheduledSpellTestUpdate4FB0E0,
		*playerScheduledSpellTestPlayer4FB0E0,
	]{
		loadUpdateData: func(unit *playerScheduledSpellTestUnit4FB0E0) *playerScheduledSpellTestUpdate4FB0E0 {
			record("update:" + unit.name)
			return update
		},
		loadCountLow: func(got *playerScheduledSpellTestUpdate4FB0E0) uint8 {
			if got != update {
				t.Fatalf("count update = %p, want %p", got, update)
			}
			value := uint8(got.count)
			record(fmt.Sprintf("count:%d", value))
			return value
		},
		loadSpell: func(got *playerScheduledSpellTestUpdate4FB0E0, index int) uint32 {
			if got != update {
				t.Fatalf("spell update = %p, want %p", got, update)
			}
			value := got.spells[index]
			record(fmt.Sprintf("spell:%d=%d", index, value))
			return value
		},
		checkSpell: func(unit *playerScheduledSpellTestUnit4FB0E0, spell uint32, bypass int32) int32 {
			record(fmt.Sprintf("check:%s:%d:%d", unit.name, spell, bypass))
			return 0
		},
		loadPosX: func(got *playerScheduledSpellTestUpdate4FB0E0) int32 {
			record(fmt.Sprintf("x:%d", got.posX))
			return got.posX
		},
		loadPosY: func(got *playerScheduledSpellTestUpdate4FB0E0) int32 {
			record(fmt.Sprintf("y:%d", got.posY))
			return got.posY
		},
		loadPlayer: func(got *playerScheduledSpellTestUpdate4FB0E0) *playerScheduledSpellTestPlayer4FB0E0 {
			record("player")
			return got.player
		},
		loadPlayerInd: func(player *playerScheduledSpellTestPlayer4FB0E0) uint8 {
			record(fmt.Sprintf("player-index:%d", player.index))
			return player.index
		},
		informText: func(index, code uint8, value int32) {
			record(fmt.Sprintf("inform:%d:%d:%d", index, code, value))
		},
		audioEvent: func(sound int32, unit *playerScheduledSpellTestUnit4FB0E0, kind, code int32) {
			record(fmt.Sprintf("audio:%d:%s:%d:%d", sound, unit.name, kind, code))
		},
		castSpell: func(spell uint32, unit *playerScheduledSpellTestUnit4FB0E0, arg playerScheduledSpellArg4FB0E0[*playerScheduledSpellTestUnit4FB0E0]) {
			record(fmt.Sprintf("cast:%d:%s:%s:%g:%g", spell, unit.name, arg.target.name, arg.posX, arg.posY))
		},
		storeSpell: func(got *playerScheduledSpellTestUpdate4FB0E0, index int, value uint32) {
			record(fmt.Sprintf("store-spell:%d=%d", index, value))
			got.spells[index] = value
		},
		storeCountLow: func(got *playerScheduledSpellTestUpdate4FB0E0, value uint8) {
			record(fmt.Sprintf("store-count:%d", value))
			got.count = got.count&^0xff | uint32(value)
		},
	}
}

func TestPlayerDoScheduledSpell4FB0E0Empty(t *testing.T) {
	unit := &playerScheduledSpellTestUnit4FB0E0{name: "unit"}
	update := &playerScheduledSpellTestUpdate4FB0E0{count: 0xaabbcc00}
	var events []string
	hooks := playerScheduledSpellTestHooks4FB0E0(t, update, &events)

	if got := playerDoScheduledSpell4FB0E0(unit, nil, hooks); got != 0 {
		t.Fatalf("FIFO result = %d, want 0", got)
	}
	if want := []string{"update:unit", "count:0"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("FIFO events = %v, want %v", events, want)
	}

	events = nil
	if got := playerDoScheduledSpellQueue4FB1D0(unit, nil, hooks); got != 0 {
		t.Fatalf("LIFO result = %d, want 0", got)
	}
	if want := []string{"update:unit", "count:0"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("LIFO events = %v, want %v", events, want)
	}
}

func TestPlayerDoScheduledSpell4FB0E0ReloadsAndShiftsFIFO(t *testing.T) {
	unit := &playerScheduledSpellTestUnit4FB0E0{name: "unit"}
	target := &playerScheduledSpellTestUnit4FB0E0{name: "target"}
	update := &playerScheduledSpellTestUpdate4FB0E0{
		spells: [5]uint32{11, 22, 33, 44, 55},
		count:  0xaabbcc03,
		posX:   -2147483648,
		posY:   16777217,
	}
	var events []string
	hooks := playerScheduledSpellTestHooks4FB0E0(t, update, &events)
	hooks.checkSpell = func(gotUnit *playerScheduledSpellTestUnit4FB0E0, gotSpell uint32, bypass int32) int32 {
		events = append(events, fmt.Sprintf("check:%s:%d:%d", gotUnit.name, gotSpell, bypass))
		update.spells[0] = 99
		return 0
	}
	hooks.castSpell = func(spell uint32, gotUnit *playerScheduledSpellTestUnit4FB0E0, arg playerScheduledSpellArg4FB0E0[*playerScheduledSpellTestUnit4FB0E0]) {
		events = append(events, fmt.Sprintf("cast:%d:%s:%s:%g:%g", spell, gotUnit.name, arg.target.name, arg.posX, arg.posY))
		update.count = update.count&^0xff | 4
	}

	if got := playerDoScheduledSpell4FB0E0(unit, target, hooks); got != 1 {
		t.Fatalf("FIFO result = %d, want 1", got)
	}
	if got, want := update.spells, ([5]uint32{22, 33, 44, 0, 55}); got != want {
		t.Fatalf("FIFO spells = %v, want %v", got, want)
	}
	if got := update.count; got != 0xaabbcc03 {
		t.Fatalf("FIFO count = %#x, want 0xaabbcc03", got)
	}
	want := []string{
		"update:unit", "count:3", "spell:0=11", "check:unit:11:0",
		"x:-2147483648", "y:16777217", "spell:0=99",
		"cast:99:unit:target:-2.1474836e+09:1.6777216e+07",
		"count:4", "spell:1=22", "store-spell:0=22", "count:4",
		"spell:2=33", "store-spell:1=33", "count:4",
		"spell:3=44", "store-spell:2=44", "count:4",
		"store-spell:3=0", "count:4", "store-count:3",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("FIFO events:\n got %v\nwant %v", events, want)
	}
}

func TestPlayerDoScheduledSpell4FB0E0RejectedStillConsumes(t *testing.T) {
	unit := &playerScheduledSpellTestUnit4FB0E0{name: "unit"}
	target := &playerScheduledSpellTestUnit4FB0E0{name: "target"}
	update := &playerScheduledSpellTestUpdate4FB0E0{
		spells: [5]uint32{71, 72, 73, 74, 75},
		count:  0x12345601,
		posX:   -5,
		posY:   6,
		player: &playerScheduledSpellTestPlayer4FB0E0{index: 9},
	}
	var events []string
	hooks := playerScheduledSpellTestHooks4FB0E0(t, update, &events)
	hooks.checkSpell = func(*playerScheduledSpellTestUnit4FB0E0, uint32, int32) int32 {
		events = append(events, "check:reject")
		return 14
	}
	hooks.castSpell = func(uint32, *playerScheduledSpellTestUnit4FB0E0, playerScheduledSpellArg4FB0E0[*playerScheduledSpellTestUnit4FB0E0]) {
		t.Fatal("rejected spell was cast")
	}

	if got := playerDoScheduledSpell4FB0E0(unit, target, hooks); got != 1 {
		t.Fatalf("FIFO result = %d, want 1", got)
	}
	if got, want := update.spells, ([5]uint32{0, 72, 73, 74, 75}); got != want {
		t.Fatalf("FIFO spells = %v, want %v", got, want)
	}
	if got := update.count; got != 0x12345600 {
		t.Fatalf("FIFO count = %#x, want 0x12345600", got)
	}
	want := []string{
		"update:unit", "count:1", "spell:0=71", "check:reject", "x:-5", "y:6",
		"player", "player-index:9", "inform:9:0:14", "audio:231:unit:0:0",
		"count:1", "store-spell:0=0", "count:1", "store-count:0",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("FIFO rejected events:\n got %v\nwant %v", events, want)
	}
}

func TestPlayerDoScheduledSpellQueue4FB1D0ReloadsNewestWithoutClearing(t *testing.T) {
	unit := &playerScheduledSpellTestUnit4FB0E0{name: "unit"}
	target := &playerScheduledSpellTestUnit4FB0E0{name: "target"}
	update := &playerScheduledSpellTestUpdate4FB0E0{
		spells: [5]uint32{10, 20, 30, 40, 50},
		count:  0xcafeba03,
		posX:   -7,
		posY:   8,
	}
	var events []string
	hooks := playerScheduledSpellTestHooks4FB0E0(t, update, &events)
	hooks.checkSpell = func(_ *playerScheduledSpellTestUnit4FB0E0, gotSpell uint32, _ int32) int32 {
		events = append(events, fmt.Sprintf("check:%d", gotSpell))
		update.count = update.count&^0xff | 2
		update.spells[1] = 77
		return 0
	}
	hooks.castSpell = func(spell uint32, _ *playerScheduledSpellTestUnit4FB0E0, arg playerScheduledSpellArg4FB0E0[*playerScheduledSpellTestUnit4FB0E0]) {
		events = append(events, fmt.Sprintf("cast:%d:%s:%g:%g", spell, arg.target.name, arg.posX, arg.posY))
		update.count = update.count&^0xff | 1
	}
	hooks.storeSpell = func(*playerScheduledSpellTestUpdate4FB0E0, int, uint32) {
		t.Fatal("LIFO path cleared a spell slot")
	}

	if got := playerDoScheduledSpellQueue4FB1D0(unit, target, hooks); got != 1 {
		t.Fatalf("LIFO result = %d, want 1", got)
	}
	if got, want := update.spells, ([5]uint32{10, 77, 30, 40, 50}); got != want {
		t.Fatalf("LIFO spells = %v, want %v", got, want)
	}
	if got := update.count; got != 0xcafeba00 {
		t.Fatalf("LIFO count = %#x, want 0xcafeba00", got)
	}
	want := []string{
		"update:unit", "count:3", "spell:2=30", "check:30", "x:-7", "y:8",
		"count:2", "spell:1=77", "cast:77:target:-7:8", "count:1", "store-count:0",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("LIFO events:\n got %v\nwant %v", events, want)
	}
}

func TestPlayerDoScheduledSpellQueue4FB1D0RejectedUsesLiveCount(t *testing.T) {
	unit := &playerScheduledSpellTestUnit4FB0E0{name: "unit"}
	update := &playerScheduledSpellTestUpdate4FB0E0{
		spells: [5]uint32{10, 20, 30, 40, 50},
		count:  0x10203002,
		player: &playerScheduledSpellTestPlayer4FB0E0{index: 4},
	}
	var events []string
	hooks := playerScheduledSpellTestHooks4FB0E0(t, update, &events)
	hooks.checkSpell = func(*playerScheduledSpellTestUnit4FB0E0, uint32, int32) int32 { return 17 }
	hooks.informText = func(index, code uint8, value int32) {
		events = append(events, fmt.Sprintf("inform:%d:%d:%d", index, code, value))
		update.count = update.count&^0xff | 4
	}
	hooks.audioEvent = func(sound int32, _ *playerScheduledSpellTestUnit4FB0E0, kind, code int32) {
		events = append(events, fmt.Sprintf("audio:%d:%d:%d", sound, kind, code))
		update.count = update.count&^0xff | 3
	}
	hooks.castSpell = func(uint32, *playerScheduledSpellTestUnit4FB0E0, playerScheduledSpellArg4FB0E0[*playerScheduledSpellTestUnit4FB0E0]) {
		t.Fatal("rejected LIFO spell was cast")
	}
	hooks.storeSpell = func(*playerScheduledSpellTestUpdate4FB0E0, int, uint32) {
		t.Fatal("rejected LIFO path cleared a spell slot")
	}

	if got := playerDoScheduledSpellQueue4FB1D0(unit, nil, hooks); got != 1 {
		t.Fatalf("LIFO result = %d, want 1", got)
	}
	if got := update.count; got != 0x10203002 {
		t.Fatalf("LIFO count = %#x, want 0x10203002", got)
	}
	if got, want := update.spells, ([5]uint32{10, 20, 30, 40, 50}); got != want {
		t.Fatalf("LIFO spells = %v, want %v", got, want)
	}
}
