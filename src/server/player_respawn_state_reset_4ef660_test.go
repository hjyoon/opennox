package server

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

type playerRespawnStateResetObject4EF660 struct {
	name        string
	update      *playerRespawnStateResetUpdate4EF660
	attribution *playerRespawnStateResetObject4EF660
}

type playerRespawnStateResetUpdate4EF660 struct {
	pending [4]*playerRespawnStateResetObject4EF660
	soul    *playerRespawnStateResetObject4EF660
	player  *playerRespawnStateResetPlayer4EF660
	traps   uint32
	field66 uint32
}

type playerRespawnStateResetPlayer4EF660 struct {
	name    string
	ankhs   [5]*playerRespawnStateResetObject4EF660
	marker0 uint32
	marker1 uint32
}

type playerRespawnStateResetWorld4EF660 struct {
	events       []string
	faultAt      int
	coop         int32
	glyphs       int32
	markerResult *playerRespawnStateResetPlayer4EF660
	onEvent      func(string)
}

func (w *playerRespawnStateResetWorld4EF660) event(event string) {
	w.events = append(w.events, event)
	if w.onEvent != nil {
		w.onEvent(event)
	}
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic("injected fault")
	}
}

func (w *playerRespawnStateResetWorld4EF660) hooks() playerRespawnStateResetHooks4EF660[
	*playerRespawnStateResetObject4EF660,
	*playerRespawnStateResetUpdate4EF660,
	*playerRespawnStateResetPlayer4EF660,
] {
	return playerRespawnStateResetHooks4EF660[
		*playerRespawnStateResetObject4EF660,
		*playerRespawnStateResetUpdate4EF660,
		*playerRespawnStateResetPlayer4EF660,
	]{
		loadUpdateData: func(unit *playerRespawnStateResetObject4EF660) *playerRespawnStateResetUpdate4EF660 {
			w.event("update")
			return unit.update
		},
		storePendingObject: func(update *playerRespawnStateResetUpdate4EF660, index int, value *playerRespawnStateResetObject4EF660) {
			w.event(fmt.Sprintf("pending:%d", index))
			update.pending[index] = value
		},
		storeSoulGate: func(update *playerRespawnStateResetUpdate4EF660, value *playerRespawnStateResetObject4EF660) {
			w.event("soul")
			update.soul = value
		},
		loadPlayer: func(update *playerRespawnStateResetUpdate4EF660) *playerRespawnStateResetPlayer4EF660 {
			w.event("player")
			return update.player
		},
		storeQuestAnkh: func(player *playerRespawnStateResetPlayer4EF660, index int, value *playerRespawnStateResetObject4EF660) {
			w.event(fmt.Sprintf("ankh:%d:%s", index, player.name))
			player.ankhs[index] = value
		},
		gameFlag: func(flag uint32) int32 {
			w.event(fmt.Sprintf("flag:%#x", flag))
			return w.coop
		},
		countGlyphs: func(unit *playerRespawnStateResetObject4EF660) int32 {
			w.event("glyphs:" + unit.name)
			return w.glyphs
		},
		storeCurTrapsLowByte: func(update *playerRespawnStateResetUpdate4EF660, value uint8) {
			w.event(fmt.Sprintf("traps:%#02x", value))
			update.traps = update.traps&^uint32(0xff) | uint32(value)
		},
		storeField66: func(update *playerRespawnStateResetUpdate4EF660, value uint32) {
			w.event("field66")
			update.field66 = value
		},
		storeAttribution: func(unit, value *playerRespawnStateResetObject4EF660) {
			w.event("attribution:" + unit.name)
			unit.attribution = value
		},
		resetPlayerMarkers: func(player *playerRespawnStateResetPlayer4EF660) *playerRespawnStateResetPlayer4EF660 {
			name := "nil"
			if player != nil {
				name = player.name
			}
			w.event("markers:" + name)
			if player != nil {
				player.marker0 = 0xdeadface
				player.marker1 = 0xdeadface
			}
			if w.markerResult != nil {
				return w.markerResult
			}
			return player
		},
	}
}

func newPlayerRespawnStateResetWorld4EF660() (
	*playerRespawnStateResetWorld4EF660,
	*playerRespawnStateResetObject4EF660,
	*playerRespawnStateResetUpdate4EF660,
	*playerRespawnStateResetPlayer4EF660,
) {
	value := &playerRespawnStateResetObject4EF660{name: "value"}
	player := &playerRespawnStateResetPlayer4EF660{name: "p0"}
	for index := range player.ankhs {
		player.ankhs[index] = value
	}
	update := &playerRespawnStateResetUpdate4EF660{
		soul:    value,
		player:  player,
		traps:   0xa1b2c3dd,
		field66: 0xffffffff,
	}
	for index := range update.pending {
		update.pending[index] = value
	}
	unit := &playerRespawnStateResetObject4EF660{
		name:        "unit",
		update:      update,
		attribution: value,
	}
	world := &playerRespawnStateResetWorld4EF660{glyphs: 0x123}
	return world, unit, update, player
}

func TestPlayerRespawnStateReset4EF660ExactNonCoopOrderAndState(t *testing.T) {
	world, unit, update, player := newPlayerRespawnStateResetWorld4EF660()
	markerResult := &playerRespawnStateResetPlayer4EF660{name: "result"}
	world.markerResult = markerResult

	got := playerRespawnStateReset4EF660(unit, world.hooks())
	wantEvents := []string{
		"update",
		"pending:0", "pending:1", "pending:2", "pending:3", "soul",
		"player", "ankh:0:p0", "player", "ankh:1:p0", "player", "ankh:2:p0",
		"player", "ankh:3:p0", "player", "ankh:4:p0",
		"flag:0x800", "glyphs:unit", "traps:0x23", "field66",
		"attribution:unit", "player", "markers:p0",
	}
	if !reflect.DeepEqual(world.events, wantEvents) {
		t.Fatalf("events = %v, want %v", world.events, wantEvents)
	}
	if got != markerResult {
		t.Fatalf("result = %p, want helper result %p", got, markerResult)
	}
	if update.pending != [4]*playerRespawnStateResetObject4EF660{} || update.soul != nil {
		t.Fatalf("cached update pointers were not cleared: pending=%v soul=%p", update.pending, update.soul)
	}
	if player.ankhs != [5]*playerRespawnStateResetObject4EF660{} {
		t.Fatalf("Quest Ankh slots = %v, want all nil", player.ankhs)
	}
	if update.traps != 0xa1b2c323 || update.field66 != 0 || unit.attribution != nil {
		t.Fatalf("traps/field66/attribution = %#x/%#x/%p", update.traps, update.field66, unit.attribution)
	}
	if player.marker0 != 0xdeadface || player.marker1 != 0xdeadface {
		t.Fatalf("markers = %#x/%#x", player.marker0, player.marker1)
	}
}

func TestPlayerRespawnStateReset4EF660CoopPreservesWholeTrapWord(t *testing.T) {
	world, unit, update, _ := newPlayerRespawnStateResetWorld4EF660()
	world.coop = -1

	playerRespawnStateReset4EF660(unit, world.hooks())
	if update.traps != 0xa1b2c3dd {
		t.Fatalf("CurTraps = %#x, want unchanged", update.traps)
	}
	for _, event := range world.events {
		if event == "glyphs:unit" || strings.HasPrefix(event, "traps:") {
			t.Fatalf("Coop path unexpectedly executed %q", event)
		}
	}
}

func TestPlayerRespawnStateReset4EF660UsesCachedUpdateAndLivePlayers(t *testing.T) {
	world, unit, cached, first := newPlayerRespawnStateResetWorld4EF660()
	replacementUpdate := &playerRespawnStateResetUpdate4EF660{traps: 0xffffffff, field66: 7}
	players := []*playerRespawnStateResetPlayer4EF660{first}
	for index := 1; index <= 5; index++ {
		players = append(players, &playerRespawnStateResetPlayer4EF660{name: fmt.Sprintf("p%d", index)})
	}
	value := &playerRespawnStateResetObject4EF660{name: "ankh"}
	for index := 0; index < 5; index++ {
		players[index].ankhs[index] = value
	}

	world.onEvent = func(event string) {
		switch event {
		case "pending:0":
			unit.update = replacementUpdate
		case "ankh:0:p0":
			cached.player = players[1]
		case "ankh:1:p1":
			cached.player = players[2]
		case "ankh:2:p2":
			cached.player = players[3]
		case "ankh:3:p3":
			cached.player = players[4]
		case "ankh:4:p4":
			cached.player = players[5]
		}
	}

	got := playerRespawnStateReset4EF660(unit, world.hooks())
	for index := 0; index < 5; index++ {
		if players[index].ankhs[index] != nil {
			t.Fatalf("live player %d slot %d was not cleared", index, index)
		}
	}
	if got != players[5] || players[5].marker0 != 0xdeadface || players[5].marker1 != 0xdeadface {
		t.Fatalf("final live player/result = %p/%#x/%#x", got, players[5].marker0, players[5].marker1)
	}
	if replacementUpdate.traps != 0xffffffff || replacementUpdate.field66 != 7 {
		t.Fatalf("replacement update was mutated: traps=%#x field66=%#x", replacementUpdate.traps, replacementUpdate.field66)
	}
	if cached.traps != 0xa1b2c323 || cached.field66 != 0 {
		t.Fatalf("cached update = traps %#x field66 %#x", cached.traps, cached.field66)
	}
}

func TestPlayerRespawnStateReset4EF660EveryObservableFaultPrefix(t *testing.T) {
	base, unit, _, _ := newPlayerRespawnStateResetWorld4EF660()
	playerRespawnStateReset4EF660(unit, base.hooks())
	want := append([]string(nil), base.events...)

	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("event_%02d", faultAt), func(t *testing.T) {
			world, unit, _, _ := newPlayerRespawnStateResetWorld4EF660()
			world.faultAt = faultAt
			func() {
				defer func() {
					if recover() == nil {
						t.Fatal("expected injected fault")
					}
				}()
				playerRespawnStateReset4EF660(unit, world.hooks())
			}()
			if !reflect.DeepEqual(world.events, want[:faultAt]) {
				t.Fatalf("events = %v, want prefix %v", world.events, want[:faultAt])
			}
		})
	}
}
