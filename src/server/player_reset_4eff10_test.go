package server

import (
	"fmt"
	"reflect"
	"testing"
)

type playerResetObject4EFF10 struct {
	name      string
	update    *playerResetUpdate4EFF10
	flags     uint32
	field541  uint8
	object130 *playerResetObject4EFF10
}

type playerResetUpdate4EFF10 struct {
	name         string
	player       *playerResetPlayer4EFF10
	manaCurrent  uint16
	manaPrevious uint16
	manaMaximum  uint16
	trapSpells   [5]uint32
	trapCount    uint32
}

type playerResetPlayer4EFF10 struct {
	name       string
	level      uint8
	index      uint8
	manaToken  uint32
	marker3660 uint32
	marker3664 uint32
}

type playerResetWorld4EFF10 struct {
	events  []string
	faultAt int
	onEvent func(string)

	unit           *playerResetObject4EFF10
	protectedToken uint32
	protectedMana  uint16
	reports        []string
}

func (w *playerResetWorld4EFF10) record(event string) {
	w.events = append(w.events, event)
	if w.onEvent != nil {
		w.onEvent(event)
	}
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic("injected fault")
	}
}

func playerResetPlayerName4EFF10(player *playerResetPlayer4EFF10) string {
	if player == nil {
		return "nil"
	}
	return player.name
}

func (w *playerResetWorld4EFF10) hooks() playerResetHooks4EFF10[
	*playerResetObject4EFF10,
	*playerResetUpdate4EFF10,
	*playerResetPlayer4EFF10,
] {
	return playerResetHooks4EFF10[
		*playerResetObject4EFF10,
		*playerResetUpdate4EFF10,
		*playerResetPlayer4EFF10,
	]{
		loadUnitArg: func() *playerResetObject4EFF10 {
			w.record("arg:" + w.unit.name)
			return w.unit
		},
		loadUpdateData: func(unit *playerResetObject4EFF10) *playerResetUpdate4EFF10 {
			update := unit.update
			w.record("update:" + unit.name + "=" + update.name)
			return update
		},
		loadPlayer: func(update *playerResetUpdate4EFF10) *playerResetPlayer4EFF10 {
			player := update.player
			w.record("player:" + update.name + "=" + playerResetPlayerName4EFF10(player))
			return player
		},
		awardBeastScrolls: func(player *playerResetPlayer4EFF10) {
			w.record("award-beast:" + playerResetPlayerName4EFF10(player))
		},
		awardSpells: func(player *playerResetPlayer4EFF10) {
			w.record("award-spells:" + playerResetPlayerName4EFF10(player))
		},
		storePlayerLevel: func(player *playerResetPlayer4EFF10, value uint8) {
			w.record(fmt.Sprintf("level:%s=%d", playerResetPlayerName4EFF10(player), value))
			player.level = value
		},
		cancelAbilities: func(unit *playerResetObject4EFF10) {
			w.record("cancel-abilities:" + unit.name)
		},
		readValues: func(unit *playerResetObject4EFF10, reward int32) {
			w.record(fmt.Sprintf("read-values:%s=%d", unit.name, reward))
		},
		awardWarriorAbilities: func(player *playerResetPlayer4EFF10) {
			w.record("award-warrior:" + playerResetPlayerName4EFF10(player))
		},
		loadManaMaximum: func(update *playerResetUpdate4EFF10) uint16 {
			value := update.manaMaximum
			w.record(fmt.Sprintf("mana-maximum:%s=%#04x", update.name, value))
			return value
		},
		storeManaCurrent: func(update *playerResetUpdate4EFF10, value uint16) {
			w.record(fmt.Sprintf("mana-current:%s=%#04x", update.name, value))
			update.manaCurrent = value
		},
		storeManaPrevious: func(update *playerResetUpdate4EFF10, value uint16) {
			w.record(fmt.Sprintf("mana-previous:%s=%#04x", update.name, value))
			update.manaPrevious = value
		},
		loadManaToken: func(player *playerResetPlayer4EFF10) uint32 {
			value := player.manaToken
			w.record(fmt.Sprintf("mana-token:%s=%#08x", playerResetPlayerName4EFF10(player), value))
			return value
		},
		protectMana: func(token uint32, value uint16) {
			w.record(fmt.Sprintf("protect-mana:%#08x=%#04x", token, value))
			w.protectedToken = token
			w.protectedMana = value
		},
		storeTrapSpell: func(update *playerResetUpdate4EFF10, index int, value uint32) {
			w.record(fmt.Sprintf("trap:%s:%d=%d", update.name, index, value))
			update.trapSpells[index] = value
		},
		storeTrapCountLow: func(update *playerResetUpdate4EFF10, value uint8) {
			w.record(fmt.Sprintf("trap-count-low:%s=%d", update.name, value))
			update.trapCount = update.trapCount&^0xff | uint32(value)
		},
		setHealthMaximum: func(unit *playerResetObject4EFF10) {
			w.record("health-max:" + unit.name)
		},
		loadObjectFlags: func(unit *playerResetObject4EFF10) uint32 {
			value := unit.flags
			w.record(fmt.Sprintf("flags:%s=%#08x", unit.name, value))
			return value
		},
		storeObjectField541: func(unit *playerResetObject4EFF10, value uint8) {
			w.record(fmt.Sprintf("field541:%s=%d", unit.name, value))
			unit.field541 = value
		},
		storeObjectFlags: func(unit *playerResetObject4EFF10, value uint32) {
			w.record(fmt.Sprintf("store-flags:%s=%#08x", unit.name, value))
			unit.flags = value
		},
		setPlayerState: func(unit *playerResetObject4EFF10, state PlayerState) {
			w.record(fmt.Sprintf("state:%s=%d", unit.name, state))
		},
		clearBuffs: func(unit *playerResetObject4EFF10) {
			w.record("clear-buffs:" + unit.name)
		},
		cancelSpells: func(unit *playerResetObject4EFF10) {
			w.record("cancel-spells:" + unit.name)
		},
		removePoison: func(unit *playerResetObject4EFF10) {
			w.record("remove-poison:" + unit.name)
		},
		resetPlayerRuntime: func(unit *playerResetObject4EFF10) {
			w.record("reset-runtime:" + unit.name)
		},
		loadPlayerIndex: func(player *playerResetPlayer4EFF10) uint8 {
			value := player.index
			w.record(fmt.Sprintf("index:%s=%d", playerResetPlayerName4EFF10(player), value))
			return value
		},
		reportTotalHealth: func(index uint8, unit *playerResetObject4EFF10) {
			event := fmt.Sprintf("report-health:%d:%s", index, unit.name)
			w.record(event)
			w.reports = append(w.reports, event)
		},
		reportTotalMana: func(index uint8, unit *playerResetObject4EFF10) {
			event := fmt.Sprintf("report-mana:%d:%s", index, unit.name)
			w.record(event)
			w.reports = append(w.reports, event)
		},
		storeObject130: func(unit, value *playerResetObject4EFF10) {
			w.record("object130:" + unit.name + "=nil")
			unit.object130 = value
		},
		storePlayerMarker3664: func(player *playerResetPlayer4EFF10, value uint32) {
			w.record(fmt.Sprintf("marker3664:%s=%#08x", playerResetPlayerName4EFF10(player), value))
			player.marker3664 = value
		},
		storePlayerMarker3660: func(player *playerResetPlayer4EFF10, value uint32) {
			w.record(fmt.Sprintf("marker3660:%s=%#08x", playerResetPlayerName4EFF10(player), value))
			player.marker3660 = value
		},
	}
}

func newPlayerResetWorld4EFF10() (*playerResetWorld4EFF10, *playerResetUpdate4EFF10, *playerResetPlayer4EFF10) {
	player := &playerResetPlayer4EFF10{name: "p0", index: 7, manaToken: 0x12345678}
	update := &playerResetUpdate4EFF10{
		name:         "u0",
		player:       player,
		manaCurrent:  1,
		manaPrevious: 2,
		manaMaximum:  0xabcd,
		trapSpells:   [5]uint32{1, 2, 3, 4, 5},
		trapCount:    0xa1b2c3dd,
	}
	old := &playerResetObject4EFF10{name: "old"}
	unit := &playerResetObject4EFF10{
		name:      "unit",
		update:    update,
		flags:     0xffffffff,
		field541:  0xff,
		object130: old,
	}
	return &playerResetWorld4EFF10{unit: unit}, update, player
}

func playerResetExpectedEvents4EFF10() []string {
	return []string{
		"arg:unit", "update:unit=u0",
		"player:u0=p0", "award-beast:p0",
		"player:u0=p0", "award-spells:p0",
		"player:u0=p0", "level:p0=1",
		"cancel-abilities:unit", "read-values:unit=0",
		"player:u0=p0", "award-warrior:p0",
		"mana-maximum:u0=0xabcd", "player:u0=p0",
		"mana-current:u0=0xabcd", "mana-previous:u0=0xabcd",
		"mana-token:p0=0x12345678", "protect-mana:0x12345678=0xabcd",
		"trap:u0:0=0", "trap:u0:1=0", "trap:u0:2=0", "trap:u0:3=0", "trap:u0:4=0",
		"trap-count-low:u0=0", "health-max:unit",
		"flags:unit=0xffffffff", "field541:unit=0", "store-flags:unit=0xffeb3fe7",
		"state:unit=13", "clear-buffs:unit", "cancel-spells:unit", "remove-poison:unit", "reset-runtime:unit",
		"player:u0=p0", "index:p0=7", "report-health:7:unit",
		"player:u0=p0", "index:p0=7", "report-mana:7:unit",
		"object130:unit=nil",
		"player:u0=p0", "marker3664:p0=0xdeadface",
		"player:u0=p0", "marker3660:p0=0xdeadface",
	}
}

func TestPlayerReset4EFF10ExactOrderAndState(t *testing.T) {
	world, update, player := newPlayerResetWorld4EFF10()
	got := playerReset4EFF10(world.hooks())

	if got != playerResetResult4EFF10 {
		t.Fatalf("result = %#x, want %#x", uint32(got), uint32(playerResetMarker4EFF10))
	}
	if !reflect.DeepEqual(world.events, playerResetExpectedEvents4EFF10()) {
		t.Fatalf("events = %v", world.events)
	}
	if player.level != 1 || update.manaCurrent != 0xabcd || update.manaPrevious != 0xabcd {
		t.Fatalf("level/mana = %d/%#x/%#x", player.level, update.manaCurrent, update.manaPrevious)
	}
	if update.trapSpells != [5]uint32{} || update.trapCount != 0xa1b2c300 {
		t.Fatalf("trap state = %v/%#x", update.trapSpells, update.trapCount)
	}
	if world.unit.flags != playerResetObjectFlagMask4EFF10 || world.unit.field541 != 0 || world.unit.object130 != nil {
		t.Fatalf("object state = %#x/%d/%p", world.unit.flags, world.unit.field541, world.unit.object130)
	}
	if player.marker3664 != playerResetMarker4EFF10 || player.marker3660 != playerResetMarker4EFF10 {
		t.Fatalf("markers = %#x/%#x", player.marker3664, player.marker3660)
	}
}

func TestPlayerReset4EFF10CachesUpdateAndReloadsEveryPlayer(t *testing.T) {
	world, cached, first := newPlayerResetWorld4EFF10()
	replacementUpdate := &playerResetUpdate4EFF10{name: "replacement", trapCount: 0xffffffff}
	players := make([]*playerResetPlayer4EFF10, 9)
	players[0] = first
	for index := 1; index < len(players); index++ {
		players[index] = &playerResetPlayer4EFF10{
			name:      fmt.Sprintf("p%d", index),
			index:     uint8(10 + index),
			manaToken: uint32(0x1000 + index),
		}
	}
	loadIndex := 0
	hooks := world.hooks()
	originalLoadPlayer := hooks.loadPlayer
	hooks.loadPlayer = func(update *playerResetUpdate4EFF10) *playerResetPlayer4EFF10 {
		if update != cached {
			t.Fatalf("load used replacement UpdateData %p", update)
		}
		cached.player = players[loadIndex]
		loadIndex++
		return originalLoadPlayer(update)
	}
	world.onEvent = func(event string) {
		if event == "award-beast:p0" {
			world.unit.update = replacementUpdate
		}
	}

	playerReset4EFF10(hooks)
	if loadIndex != len(players) {
		t.Fatalf("Player loads = %d, want %d", loadIndex, len(players))
	}
	if players[2].level != 1 {
		t.Fatalf("third live Player level = %d", players[2].level)
	}
	if world.protectedToken != players[4].manaToken {
		t.Fatalf("mana token = %#x, want fifth Player %#x", world.protectedToken, players[4].manaToken)
	}
	if len(world.reports) != 2 || world.reports[0] != "report-health:15:unit" || world.reports[1] != "report-mana:16:unit" {
		t.Fatalf("reports = %v", world.reports)
	}
	if players[7].marker3664 != playerResetMarker4EFF10 || players[8].marker3660 != playerResetMarker4EFF10 {
		t.Fatalf("live markers = %#x/%#x", players[7].marker3664, players[8].marker3660)
	}
	if replacementUpdate.trapCount != 0xffffffff {
		t.Fatalf("replacement update changed = %#x", replacementUpdate.trapCount)
	}
}

func TestPlayerReset4EFF10EveryObservableFaultPrefix(t *testing.T) {
	want := playerResetExpectedEvents4EFF10()
	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("event_%02d", faultAt), func(t *testing.T) {
			world, _, _ := newPlayerResetWorld4EFF10()
			world.faultAt = faultAt
			func() {
				defer func() {
					if recover() == nil {
						t.Fatal("expected injected fault")
					}
				}()
				playerReset4EFF10(world.hooks())
			}()
			if !reflect.DeepEqual(world.events, want[:faultAt]) {
				t.Fatalf("events = %v, want prefix %v", world.events, want[:faultAt])
			}
		})
	}
}
