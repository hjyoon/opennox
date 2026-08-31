package opennox

import (
	"fmt"
	"math"
	"reflect"
	"testing"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/server"
)

type playerExecuteAbilityTestUnit4FBB70 struct {
	name   string
	flags  object.Flags
	class  uint8
	update *playerExecuteAbilityTestUpdate4FBB70
	first  *playerExecuteAbilityTestItem4FBB70
}

type playerExecuteAbilityTestUpdate4FBB70 struct {
	name   string
	state  uint8
	player *playerExecuteAbilityTestPlayer4FBB70
}

type playerExecuteAbilityTestPlayer4FBB70 struct {
	name         string
	class        uint8
	index        uint8
	levels       [server.AbilityMax]uint32
	berserkBlock uint32
}

type playerExecuteAbilityTestItem4FBB70 struct {
	name  string
	class object.Class
	next  *playerExecuteAbilityTestItem4FBB70
}

type playerExecuteAbilityTestExec4FBB70 struct {
	name    string
	unit    *playerExecuteAbilityTestUnit4FBB70
	ability server.Ability
	frame   uint32
	active  uint32
	next    *playerExecuteAbilityTestExec4FBB70
	prev    *playerExecuteAbilityTestExec4FBB70
}

type playerExecuteAbilityActiveKey4FBB70 struct {
	value   bool
	ability server.Ability
}

type playerExecuteAbilityCooldownKey4FBB70 struct {
	index   uint8
	ability server.Ability
}

type playerExecuteAbilityTestWorld4FBB70 struct {
	events      []string
	faultAt     int
	afterEvent  map[string]func()
	game        map[uint32][]int32
	gameCalls   map[uint32]int
	active      map[playerExecuteAbilityActiveKey4FBB70][]int32
	activeCalls map[playerExecuteAbilityActiveKey4FBB70]int
	cooldowns   map[playerExecuteAbilityCooldownKey4FBB70]int32
	delays      []int32
	delayCalls  int
	duration    int32
	frame       uint32
	sound       int32
	allocated   *playerExecuteAbilityTestExec4FBB70
	head        *playerExecuteAbilityTestExec4FBB70
}

func playerExecuteAbilityUnitName4FBB70(unit *playerExecuteAbilityTestUnit4FBB70) string {
	if unit == nil {
		return "nil"
	}
	return unit.name
}

func playerExecuteAbilityUpdateName4FBB70(update *playerExecuteAbilityTestUpdate4FBB70) string {
	if update == nil {
		return "nil"
	}
	return update.name
}

func playerExecuteAbilityPlayerName4FBB70(player *playerExecuteAbilityTestPlayer4FBB70) string {
	if player == nil {
		return "nil"
	}
	return player.name
}

func playerExecuteAbilityItemName4FBB70(item *playerExecuteAbilityTestItem4FBB70) string {
	if item == nil {
		return "nil"
	}
	return item.name
}

func playerExecuteAbilityExecName4FBB70(exec *playerExecuteAbilityTestExec4FBB70) string {
	if exec == nil {
		return "nil"
	}
	return exec.name
}

func (w *playerExecuteAbilityTestWorld4FBB70) record(event string) {
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
}

func (w *playerExecuteAbilityTestWorld4FBB70) finish(event string) {
	if after := w.afterEvent[event]; after != nil {
		after()
	}
}

func playerExecuteAbilityQueued4FBB70[K comparable](values map[K][]int32, calls map[K]int, key K) int32 {
	index := calls[key]
	calls[key] = index + 1
	queue := values[key]
	if len(queue) == 0 {
		return 0
	}
	if index >= len(queue) {
		return queue[len(queue)-1]
	}
	return queue[index]
}

func (w *playerExecuteAbilityTestWorld4FBB70) hooks() playerExecuteAbilityHooks4FBB70[
	*playerExecuteAbilityTestUnit4FBB70,
	*playerExecuteAbilityTestUpdate4FBB70,
	*playerExecuteAbilityTestPlayer4FBB70,
	*playerExecuteAbilityTestItem4FBB70,
	*playerExecuteAbilityTestExec4FBB70,
] {
	return playerExecuteAbilityHooks4FBB70[
		*playerExecuteAbilityTestUnit4FBB70,
		*playerExecuteAbilityTestUpdate4FBB70,
		*playerExecuteAbilityTestPlayer4FBB70,
		*playerExecuteAbilityTestItem4FBB70,
		*playerExecuteAbilityTestExec4FBB70,
	]{
		loadFlags: func(unit *playerExecuteAbilityTestUnit4FBB70) object.Flags {
			if unit == nil {
				event := "flags:nil"
				w.record(event)
				panic(event)
			}
			event := fmt.Sprintf("flags:%s=%08x", unit.name, uint32(unit.flags))
			w.record(event)
			w.finish(event)
			return unit.flags
		},
		loadClassLow: func(unit *playerExecuteAbilityTestUnit4FBB70) uint8 {
			if unit == nil {
				event := "class:nil"
				w.record(event)
				panic(event)
			}
			event := fmt.Sprintf("class:%s=%02x", unit.name, unit.class)
			w.record(event)
			w.finish(event)
			return unit.class
		},
		loadUpdateData: func(unit *playerExecuteAbilityTestUnit4FBB70) *playerExecuteAbilityTestUpdate4FBB70 {
			if unit == nil {
				event := "update:nil-unit"
				w.record(event)
				panic(event)
			}
			update := unit.update
			event := "update:" + unit.name + "=" + playerExecuteAbilityUpdateName4FBB70(update)
			w.record(event)
			w.finish(event)
			return update
		},
		loadPlayer: func(update *playerExecuteAbilityTestUpdate4FBB70) *playerExecuteAbilityTestPlayer4FBB70 {
			if update == nil {
				event := "player:nil-update"
				w.record(event)
				panic(event)
			}
			player := update.player
			event := "player:" + update.name + "=" + playerExecuteAbilityPlayerName4FBB70(player)
			w.record(event)
			w.finish(event)
			return player
		},
		loadPlayerClassLow: func(player *playerExecuteAbilityTestPlayer4FBB70) uint8 {
			if player == nil {
				event := "player-class:nil"
				w.record(event)
				panic(event)
			}
			event := fmt.Sprintf("player-class:%s=%02x", player.name, player.class)
			w.record(event)
			w.finish(event)
			return player.class
		},
		gameFlag: func(mask uint32) int32 {
			result := playerExecuteAbilityQueued4FBB70(w.game, w.gameCalls, mask)
			event := fmt.Sprintf("game:%08x=%d", mask, result)
			w.record(event)
			w.finish(event)
			return result
		},
		firstItem: func(unit *playerExecuteAbilityTestUnit4FBB70) *playerExecuteAbilityTestItem4FBB70 {
			item := unit.first
			event := "first:" + unit.name + "=" + playerExecuteAbilityItemName4FBB70(item)
			w.record(event)
			w.finish(event)
			return item
		},
		nextItem: func(item *playerExecuteAbilityTestItem4FBB70) *playerExecuteAbilityTestItem4FBB70 {
			if item == nil {
				event := "next:nil"
				w.record(event)
				panic(event)
			}
			next := item.next
			event := "next:" + item.name + "=" + playerExecuteAbilityItemName4FBB70(next)
			w.record(event)
			w.finish(event)
			return next
		},
		loadItemClass: func(item *playerExecuteAbilityTestItem4FBB70) object.Class {
			if item == nil {
				event := "item-class:nil"
				w.record(event)
				panic(event)
			}
			event := fmt.Sprintf("item-class:%s=%08x", item.name, uint32(item.class))
			w.record(event)
			w.finish(event)
			return item.class
		},
		isActive: func(unit *playerExecuteAbilityTestUnit4FBB70, ability server.Ability) int32 {
			key := playerExecuteAbilityActiveKey4FBB70{ability: ability}
			result := playerExecuteAbilityQueued4FBB70(w.active, w.activeCalls, key)
			event := fmt.Sprintf("active:%s:%d=%d", playerExecuteAbilityUnitName4FBB70(unit), ability, result)
			w.record(event)
			w.finish(event)
			return result
		},
		isActiveVal: func(unit *playerExecuteAbilityTestUnit4FBB70, ability server.Ability) int32 {
			key := playerExecuteAbilityActiveKey4FBB70{value: true, ability: ability}
			result := playerExecuteAbilityQueued4FBB70(w.active, w.activeCalls, key)
			event := fmt.Sprintf("active-val:%s:%d=%d", playerExecuteAbilityUnitName4FBB70(unit), ability, result)
			w.record(event)
			w.finish(event)
			return result
		},
		loadState: func(update *playerExecuteAbilityTestUpdate4FBB70) uint8 {
			if update == nil {
				event := "state:nil"
				w.record(event)
				panic(event)
			}
			event := fmt.Sprintf("state:%s=%02x", update.name, update.state)
			w.record(event)
			w.finish(event)
			return update.state
		},
		loadSpellLevel: func(player *playerExecuteAbilityTestPlayer4FBB70, ability server.Ability) uint32 {
			if player == nil {
				event := fmt.Sprintf("level:nil:%d", ability)
				w.record(event)
				panic(event)
			}
			level := player.levels[ability]
			event := fmt.Sprintf("level:%s:%d=%08x", player.name, ability, level)
			w.record(event)
			w.finish(event)
			return level
		},
		loadBerserkBlock: func(player *playerExecuteAbilityTestPlayer4FBB70) uint32 {
			if player == nil {
				event := "berserk-block:nil"
				w.record(event)
				panic(event)
			}
			value := player.berserkBlock
			event := fmt.Sprintf("berserk-block:%s=%08x", player.name, value)
			w.record(event)
			w.finish(event)
			return value
		},
		loadPlayerIndex: func(player *playerExecuteAbilityTestPlayer4FBB70) uint8 {
			if player == nil {
				event := "index:nil"
				w.record(event)
				panic(event)
			}
			index := player.index
			event := fmt.Sprintf("index:%s=%02x", player.name, index)
			w.record(event)
			w.finish(event)
			return index
		},
		reportText: func(index, kind uint8, code *int32) {
			event := fmt.Sprintf("report:%02x:%d=%d", index, kind, *code)
			w.record(event)
			w.finish(event)
		},
		loadCooldown: func(index uint8, ability server.Ability) int32 {
			key := playerExecuteAbilityCooldownKey4FBB70{index: index, ability: ability}
			value := w.cooldowns[key]
			event := fmt.Sprintf("cooldown:%02x:%d=%08x", index, ability, uint32(value))
			w.record(event)
			w.finish(event)
			return value
		},
		storeCooldown: func(index uint8, ability server.Ability, value int32) {
			event := fmt.Sprintf("store-cooldown:%02x:%d=%08x", index, ability, uint32(value))
			w.record(event)
			w.cooldowns[playerExecuteAbilityCooldownKey4FBB70{index: index, ability: ability}] = value
			w.finish(event)
		},
		loadDelay: func(ability server.Ability) int32 {
			index := w.delayCalls
			w.delayCalls++
			value := int32(0)
			if len(w.delays) != 0 {
				if index >= len(w.delays) {
					index = len(w.delays) - 1
				}
				value = w.delays[index]
			}
			event := fmt.Sprintf("delay:%d=%08x", ability, uint32(value))
			w.record(event)
			w.finish(event)
			return value
		},
		reportState: func(unit *playerExecuteAbilityTestUnit4FBB70, ability server.Ability, state uint8) {
			event := fmt.Sprintf("state-report:%s:%d=%d", playerExecuteAbilityUnitName4FBB70(unit), ability, state)
			w.record(event)
			w.finish(event)
		},
		loadDuration: func(ability server.Ability) int32 {
			event := fmt.Sprintf("duration:%d=%08x", ability, uint32(w.duration))
			w.record(event)
			w.finish(event)
			return w.duration
		},
		allocExec: func() *playerExecuteAbilityTestExec4FBB70 {
			exec := w.allocated
			event := "alloc=" + playerExecuteAbilityExecName4FBB70(exec)
			w.record(event)
			w.finish(event)
			return exec
		},
		loadFrame: func() uint32 {
			event := fmt.Sprintf("frame=%08x", w.frame)
			w.record(event)
			w.finish(event)
			return w.frame
		},
		storeExecUnit: func(exec *playerExecuteAbilityTestExec4FBB70, unit *playerExecuteAbilityTestUnit4FBB70) {
			event := "exec-unit:" + playerExecuteAbilityExecName4FBB70(exec) + "=" + playerExecuteAbilityUnitName4FBB70(unit)
			w.record(event)
			exec.unit = unit
			w.finish(event)
		},
		storeExecAbility: func(exec *playerExecuteAbilityTestExec4FBB70, ability server.Ability) {
			event := fmt.Sprintf("exec-ability:%s=%d", playerExecuteAbilityExecName4FBB70(exec), ability)
			w.record(event)
			exec.ability = ability
			w.finish(event)
		},
		storeExecFrame: func(exec *playerExecuteAbilityTestExec4FBB70, frame uint32) {
			event := fmt.Sprintf("exec-frame:%s=%08x", playerExecuteAbilityExecName4FBB70(exec), frame)
			w.record(event)
			exec.frame = frame
			w.finish(event)
		},
		loadExecHead: func() *playerExecuteAbilityTestExec4FBB70 {
			head := w.head
			event := "head=" + playerExecuteAbilityExecName4FBB70(head)
			w.record(event)
			w.finish(event)
			return head
		},
		storeExecNext: func(exec, next *playerExecuteAbilityTestExec4FBB70) {
			event := "exec-next:" + playerExecuteAbilityExecName4FBB70(exec) + "=" + playerExecuteAbilityExecName4FBB70(next)
			w.record(event)
			exec.next = next
			w.finish(event)
		},
		storeExecActive: func(exec *playerExecuteAbilityTestExec4FBB70, active uint32) {
			event := fmt.Sprintf("exec-active:%s=%08x", playerExecuteAbilityExecName4FBB70(exec), active)
			w.record(event)
			exec.active = active
			w.finish(event)
		},
		storeExecPrev: func(exec, prev *playerExecuteAbilityTestExec4FBB70) {
			event := "exec-prev:" + playerExecuteAbilityExecName4FBB70(exec) + "=" + playerExecuteAbilityExecName4FBB70(prev)
			w.record(event)
			exec.prev = prev
			w.finish(event)
		},
		storeExecHead: func(exec *playerExecuteAbilityTestExec4FBB70) {
			event := "store-head=" + playerExecuteAbilityExecName4FBB70(exec)
			w.record(event)
			w.head = exec
			w.finish(event)
		},
		invoke: func(unit *playerExecuteAbilityTestUnit4FBB70, ability server.Ability) {
			event := fmt.Sprintf("invoke:%s:%d", playerExecuteAbilityUnitName4FBB70(unit), ability)
			w.record(event)
			w.finish(event)
		},
		loadSound: func(ability server.Ability, slot int32) int32 {
			event := fmt.Sprintf("sound:%d:%d=%08x", ability, slot, uint32(w.sound))
			w.record(event)
			w.finish(event)
			return w.sound
		},
		audio: func(soundID int32, unit *playerExecuteAbilityTestUnit4FBB70, kind int32, code uint32) {
			event := fmt.Sprintf("audio:%08x:%s:%d:%08x", uint32(soundID), playerExecuteAbilityUnitName4FBB70(unit), kind, code)
			w.record(event)
			w.finish(event)
		},
	}
}

func playerExecuteAbilityFixture4FBB70() (
	*playerExecuteAbilityTestUnit4FBB70,
	*playerExecuteAbilityTestUpdate4FBB70,
	*playerExecuteAbilityTestPlayer4FBB70,
	*playerExecuteAbilityTestWorld4FBB70,
) {
	player := &playerExecuteAbilityTestPlayer4FBB70{name: "player", index: 7}
	for ability := server.AbilityBerserk; ability <= server.AbilityInfravis; ability++ {
		player.levels[ability] = 1
	}
	update := &playerExecuteAbilityTestUpdate4FBB70{name: "update", player: player}
	unit := &playerExecuteAbilityTestUnit4FBB70{
		name:   "unit",
		flags:  object.FlagSelected,
		class:  playerExecuteAbilityPlayerClass4FBB70,
		update: update,
	}
	world := &playerExecuteAbilityTestWorld4FBB70{
		afterEvent:  make(map[string]func()),
		game:        make(map[uint32][]int32),
		gameCalls:   make(map[uint32]int),
		active:      make(map[playerExecuteAbilityActiveKey4FBB70][]int32),
		activeCalls: make(map[playerExecuteAbilityActiveKey4FBB70]int),
		cooldowns:   make(map[playerExecuteAbilityCooldownKey4FBB70]int32),
		delays:      []int32{-0x7edcba99, 9},
		duration:    7,
		frame:       math.MaxUint32 - 3,
		sound:       int32(0x76543210),
		allocated:   &playerExecuteAbilityTestExec4FBB70{name: "new"},
		head:        &playerExecuteAbilityTestExec4FBB70{name: "old"},
	}
	return unit, update, player, world
}

func playerExecuteAbilitySuccessTrace4FBB70() []string {
	return []string{
		"flags:unit=40000000",
		"class:unit=04",
		"update:unit=update",
		"player:update=player",
		"player-class:player=00",
		"game:00000020=0",
		"active:unit:4=0",
		"state:update=00",
		"game:00000800=0",
		"flags:unit=40000000",
		"player:update=player",
		"game:00002000=0",
		"level:player:4=00000001",
		"index:player=07",
		"cooldown:07:4=00000000",
		"delay:4=81234567",
		"index:player=07",
		"store-cooldown:07:4=81234567",
		"delay:4=00000009",
		"state-report:unit:4=0",
		"duration:4=00000007",
		"alloc=new",
		"frame=fffffffc",
		"exec-unit:new=unit",
		"exec-ability:new=4",
		"exec-frame:new=00000003",
		"head=old",
		"exec-next:new=old",
		"exec-active:new=00000001",
		"exec-prev:new=nil",
		"head=old",
		"exec-prev:old=new",
		"store-head=new",
		"invoke:unit:4",
		"sound:4:0=76543210",
		"audio:76543210:unit:0:00000000",
	}
}

func TestPlayerExecuteAbility4FBB70ExactSuccessOrderAndWrap(t *testing.T) {
	unit, _, _, world := playerExecuteAbilityFixture4FBB70()
	old := world.head
	playerExecuteAbility4FBB70(unit, server.AbilityTreadLightly, world.hooks())

	want := playerExecuteAbilitySuccessTrace4FBB70()
	if !reflect.DeepEqual(world.events, want) {
		t.Fatalf("events =\n%v\nwant\n%v", world.events, want)
	}
	if got := world.cooldowns[playerExecuteAbilityCooldownKey4FBB70{index: 7, ability: server.AbilityTreadLightly}]; got != int32(-0x7edcba99) {
		t.Fatalf("stored cooldown = %#08x, want %#08x", uint32(got), uint32(0x81234567))
	}
	if world.head != world.allocated || world.allocated.next != old || old.prev != world.allocated {
		t.Fatalf("execution list = head %p, new.next %p, old.prev %p", world.head, world.allocated.next, old.prev)
	}
	if world.allocated.frame != 3 || world.allocated.active != 1 || world.allocated.unit != unit || world.allocated.ability != server.AbilityTreadLightly {
		t.Fatalf("execution record = %+v", world.allocated)
	}
}

func TestPlayerExecuteAbility4FBB70NilAndSignedAbilityGuardsPrecedeReads(t *testing.T) {
	for _, ability := range []server.Ability{
		server.Ability(math.MinInt32), -1, server.AbilityInvalid,
		server.AbilityMax, 6, server.Ability(math.MaxInt32),
	} {
		t.Run(fmt.Sprintf("ability-%08x", uint32(ability)), func(t *testing.T) {
			unit, _, _, world := playerExecuteAbilityFixture4FBB70()
			playerExecuteAbility4FBB70(unit, ability, world.hooks())
			if len(world.events) != 0 {
				t.Fatalf("events = %v, want none", world.events)
			}
		})
	}
	_, _, _, world := playerExecuteAbilityFixture4FBB70()
	playerExecuteAbility4FBB70(
		(*playerExecuteAbilityTestUnit4FBB70)(nil),
		server.AbilityBerserk,
		world.hooks(),
	)
	if len(world.events) != 0 {
		t.Fatalf("nil events = %v, want none", world.events)
	}
}

func TestPlayerExecuteAbility4FBB70ClassAndFlagGates(t *testing.T) {
	for _, flags := range []object.Flags{
		object.FlagDestroyed,
		object.FlagDead,
		object.FlagDestroyed | object.FlagDead,
	} {
		unit, _, _, world := playerExecuteAbilityFixture4FBB70()
		unit.flags = flags
		playerExecuteAbility4FBB70(unit, server.AbilityBerserk, world.hooks())
		want := []string{fmt.Sprintf("flags:unit=%08x", uint32(flags))}
		if !reflect.DeepEqual(world.events, want) {
			t.Fatalf("flags %#x events = %v, want %v", uint32(flags), world.events, want)
		}
	}

	unit, _, _, world := playerExecuteAbilityFixture4FBB70()
	unit.class = 0xf8
	playerExecuteAbility4FBB70(unit, server.AbilityBerserk, world.hooks())
	want := []string{"flags:unit=40000000", "class:unit=f8"}
	if !reflect.DeepEqual(world.events, want) {
		t.Fatalf("non-player events = %v, want %v", world.events, want)
	}

	unit, _, player, world := playerExecuteAbilityFixture4FBB70()
	player.class = 0x80
	playerExecuteAbility4FBB70(unit, server.AbilityBerserk, world.hooks())
	want = []string{
		"flags:unit=40000000", "class:unit=04", "update:unit=update",
		"player:update=player", "player-class:player=80",
	}
	if !reflect.DeepEqual(world.events, want) {
		t.Fatalf("non-Warrior events = %v, want %v", world.events, want)
	}
}

func playerExecuteAbilityHasSuffix4FBB70(events, suffix []string) bool {
	return len(events) >= len(suffix) && reflect.DeepEqual(events[len(events)-len(suffix):], suffix)
}

func TestPlayerExecuteAbility4FBB70ExactFailureReportsAndFizzle(t *testing.T) {
	tests := []struct {
		name      string
		ability   server.Ability
		configure func(*playerExecuteAbilityTestUnit4FBB70, *playerExecuteAbilityTestUpdate4FBB70, *playerExecuteAbilityTestPlayer4FBB70, *playerExecuteAbilityTestWorld4FBB70)
		code      int32
		fizzle    bool
	}{
		{
			name: "flag carrier", ability: server.AbilityBerserk, code: 5, fizzle: true,
			configure: func(unit *playerExecuteAbilityTestUnit4FBB70, _ *playerExecuteAbilityTestUpdate4FBB70, _ *playerExecuteAbilityTestPlayer4FBB70, world *playerExecuteAbilityTestWorld4FBB70) {
				world.game[playerExecuteAbilityCTF4FBB70] = []int32{-7}
				unit.first = &playerExecuteAbilityTestItem4FBB70{name: "flag", class: object.ClassFlag}
			},
		},
		{
			name: "mutual exclusion", ability: server.AbilityWarcry, code: 2, fizzle: true,
			configure: func(_ *playerExecuteAbilityTestUnit4FBB70, _ *playerExecuteAbilityTestUpdate4FBB70, _ *playerExecuteAbilityTestPlayer4FBB70, world *playerExecuteAbilityTestWorld4FBB70) {
				world.active[playerExecuteAbilityActiveKey4FBB70{ability: server.AbilityBerserk}] = []int32{1}
			},
		},
		{
			name: "same ability", ability: server.AbilityInfravis, code: 2, fizzle: true,
			configure: func(_ *playerExecuteAbilityTestUnit4FBB70, _ *playerExecuteAbilityTestUpdate4FBB70, _ *playerExecuteAbilityTestPlayer4FBB70, world *playerExecuteAbilityTestWorld4FBB70) {
				world.active[playerExecuteAbilityActiveKey4FBB70{ability: server.AbilityInfravis}] = []int32{-1}
			},
		},
		{
			name: "state", ability: server.AbilityTreadLightly, code: 6, fizzle: true,
			configure: func(_ *playerExecuteAbilityTestUnit4FBB70, update *playerExecuteAbilityTestUpdate4FBB70, _ *playerExecuteAbilityTestPlayer4FBB70, _ *playerExecuteAbilityTestWorld4FBB70) {
				update.state = playerExecuteAbilityBlockedState4FBB70
			},
		},
		{
			name: "airborne", ability: server.AbilityTreadLightly, code: 6, fizzle: true,
			configure: func(unit *playerExecuteAbilityTestUnit4FBB70, _ *playerExecuteAbilityTestUpdate4FBB70, _ *playerExecuteAbilityTestPlayer4FBB70, _ *playerExecuteAbilityTestWorld4FBB70) {
				unit.flags |= object.FlagAirborne
			},
		},
		{
			name: "unknown", ability: server.AbilityTreadLightly, code: 3, fizzle: true,
			configure: func(_ *playerExecuteAbilityTestUnit4FBB70, _ *playerExecuteAbilityTestUpdate4FBB70, player *playerExecuteAbilityTestPlayer4FBB70, _ *playerExecuteAbilityTestWorld4FBB70) {
				player.levels[server.AbilityTreadLightly] = 0
			},
		},
		{
			name: "overweight", ability: server.AbilityBerserk, code: 7, fizzle: false,
			configure: func(_ *playerExecuteAbilityTestUnit4FBB70, _ *playerExecuteAbilityTestUpdate4FBB70, player *playerExecuteAbilityTestPlayer4FBB70, world *playerExecuteAbilityTestWorld4FBB70) {
				player.berserkBlock = 1
				world.game[playerExecuteAbilityOnline4FBB70] = []int32{1}
				world.game[playerExecuteAbilityQuest4FBB70] = []int32{0}
			},
		},
		{
			name: "cooldown", ability: server.AbilityTreadLightly, code: 2, fizzle: true,
			configure: func(_ *playerExecuteAbilityTestUnit4FBB70, _ *playerExecuteAbilityTestUpdate4FBB70, player *playerExecuteAbilityTestPlayer4FBB70, _ *playerExecuteAbilityTestWorld4FBB70) {
				player.levels[server.AbilityTreadLightly] = 1
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			unit, update, player, world := playerExecuteAbilityFixture4FBB70()
			world.duration = 0
			tc.configure(unit, update, player, world)
			if tc.name == "cooldown" {
				world.cooldowns[playerExecuteAbilityCooldownKey4FBB70{index: player.index, ability: tc.ability}] = 44
			}
			playerExecuteAbility4FBB70(unit, tc.ability, world.hooks())
			suffix := []string{
				"player:update=player",
				"index:player=07",
				fmt.Sprintf("report:07:2=%d", tc.code),
			}
			if tc.fizzle {
				suffix = append(suffix, "audio:000000e7:unit:0:00000000")
			}
			if !playerExecuteAbilityHasSuffix4FBB70(world.events, suffix) {
				t.Fatalf("events = %v, want suffix %v", world.events, suffix)
			}
			for _, event := range world.events {
				if event == fmt.Sprintf("invoke:unit:%d", tc.ability) {
					t.Fatalf("rejected ability invoked: %v", world.events)
				}
			}
		})
	}
}

func TestPlayerExecuteAbility4FBB70MutualExclusionOrder(t *testing.T) {
	tests := []struct {
		ability server.Ability
		want    []string
	}{
		{server.AbilityBerserk, []string{"active-val:unit:2=0", "active:unit:3=0", "active:unit:1=0"}},
		{server.AbilityWarcry, []string{"active:unit:1=0", "active:unit:3=0", "active:unit:2=0"}},
		{server.AbilityHarpoon, []string{"active-val:unit:2=0", "active:unit:1=0", "active:unit:3=0"}},
		{server.AbilityTreadLightly, []string{"active:unit:4=0"}},
		{server.AbilityInfravis, []string{"active:unit:5=0"}},
	}
	for _, tc := range tests {
		t.Run(tc.ability.String(), func(t *testing.T) {
			unit, _, _, world := playerExecuteAbilityFixture4FBB70()
			world.duration = 0
			world.game[playerExecuteAbilityOnline4FBB70] = []int32{1}
			world.game[playerExecuteAbilityQuest4FBB70] = []int32{0}
			playerExecuteAbility4FBB70(unit, tc.ability, world.hooks())
			start := 6
			if got := world.events[start : start+len(tc.want)]; !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("active order = %v, want %v (all events %v)", got, tc.want, world.events)
			}
		})
	}
}

func TestPlayerExecuteAbility4FBB70LivePlayerReloadsAndCooldownIndex(t *testing.T) {
	t.Run("report reload", func(t *testing.T) {
		unit, update, initial, world := playerExecuteAbilityFixture4FBB70()
		snapshot := &playerExecuteAbilityTestPlayer4FBB70{name: "snapshot", index: 2}
		snapshot.levels[server.AbilityTreadLightly] = 1
		report := &playerExecuteAbilityTestPlayer4FBB70{name: "report", index: 3}
		world.afterEvent["player-class:player=00"] = func() { update.player = snapshot }
		world.cooldowns[playerExecuteAbilityCooldownKey4FBB70{index: 2, ability: server.AbilityTreadLightly}] = 99
		world.afterEvent["cooldown:02:4=00000063"] = func() { update.player = report }

		playerExecuteAbility4FBB70(unit, server.AbilityTreadLightly, world.hooks())
		if initial == snapshot || initial == report {
			t.Fatal("fixture players unexpectedly alias")
		}
		want := []string{
			"player:update=player", "player-class:player=00",
			"player:update=snapshot", "index:snapshot=02", "cooldown:02:4=00000063",
			"player:update=report", "index:report=03", "report:03:2=2",
		}
		cursor := 0
		for _, event := range world.events {
			if cursor < len(want) && event == want[cursor] {
				cursor++
			}
		}
		if cursor != len(want) {
			t.Fatalf("events = %v, missing ordered subsequence %v at %d", world.events, want, cursor)
		}
	})

	t.Run("index reloaded before store", func(t *testing.T) {
		unit, _, player, world := playerExecuteAbilityFixture4FBB70()
		world.duration = 0
		world.afterEvent["index:player=07"] = func() {
			delete(world.afterEvent, "index:player=07")
			player.index = 9
		}
		playerExecuteAbility4FBB70(unit, server.AbilityTreadLightly, world.hooks())
		want := []string{
			"index:player=07", "cooldown:07:4=00000000", "delay:4=81234567",
			"index:player=09", "store-cooldown:09:4=81234567",
		}
		cursor := 0
		for _, event := range world.events {
			if cursor < len(want) && event == want[cursor] {
				cursor++
			}
		}
		if cursor != len(want) {
			t.Fatalf("events = %v, missing ordered subsequence %v", world.events, want)
		}
	})

	t.Run("Berserk block and report reload independently", func(t *testing.T) {
		unit, update, _, world := playerExecuteAbilityFixture4FBB70()
		block := &playerExecuteAbilityTestPlayer4FBB70{name: "block", index: 4, berserkBlock: 1}
		report := &playerExecuteAbilityTestPlayer4FBB70{name: "report", index: 5}
		world.game[playerExecuteAbilityOnline4FBB70] = []int32{1}
		world.game[playerExecuteAbilityQuest4FBB70] = []int32{0}
		world.afterEvent["game:00001000=0"] = func() { update.player = block }
		world.afterEvent["berserk-block:block=00000001"] = func() { update.player = report }
		playerExecuteAbility4FBB70(unit, server.AbilityBerserk, world.hooks())
		want := []string{
			"player:update=block", "berserk-block:block=00000001",
			"player:update=report", "index:report=05", "report:05:2=7",
		}
		if !playerExecuteAbilityHasSuffix4FBB70(world.events, want) {
			t.Fatalf("events = %v, want suffix %v", world.events, want)
		}
	})
}

func TestPlayerExecuteAbility4FBB70IntegerGamePredicatesAndDualDelay(t *testing.T) {
	t.Run("Online must equal one", func(t *testing.T) {
		unit, _, player, world := playerExecuteAbilityFixture4FBB70()
		player.levels[server.AbilityTreadLightly] = 0
		world.game[playerExecuteAbilityOnline4FBB70] = []int32{2}
		playerExecuteAbility4FBB70(unit, server.AbilityTreadLightly, world.hooks())
		for _, forbidden := range []string{"game:00001000=0", "invoke:unit:4"} {
			for _, event := range world.events {
				if event == forbidden {
					t.Fatalf("unexpected event %q in %v", forbidden, world.events)
				}
			}
		}
		if !playerExecuteAbilityHasSuffix4FBB70(world.events, []string{
			"player:update=player", "index:player=07", "report:07:2=3",
			"audio:000000e7:unit:0:00000000",
		}) {
			t.Fatalf("events = %v", world.events)
		}
	})

	t.Run("Online one and Quest zero skip level", func(t *testing.T) {
		unit, _, player, world := playerExecuteAbilityFixture4FBB70()
		player.levels[server.AbilityTreadLightly] = 0
		world.game[playerExecuteAbilityOnline4FBB70] = []int32{1}
		world.game[playerExecuteAbilityQuest4FBB70] = []int32{0}
		world.duration = 0
		playerExecuteAbility4FBB70(unit, server.AbilityTreadLightly, world.hooks())
		for _, event := range world.events {
			if event == "level:player:4=00000000" {
				t.Fatalf("level unexpectedly read: %v", world.events)
			}
		}
		if !playerExecuteAbilityHasSuffix4FBB70(world.events, []string{
			"invoke:unit:4", "sound:4:0=76543210", "audio:76543210:unit:0:00000000",
		}) {
			t.Fatalf("events = %v", world.events)
		}
	})

	for _, tc := range []struct {
		name       string
		delays     []int32
		wantReport bool
		wantStored uint32
	}{
		{"first stored, second zero", []int32{math.MinInt32 + 17, 0}, false, 0x80000011},
		{"second nonzero reports", []int32{0, -9}, true, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			unit, _, _, world := playerExecuteAbilityFixture4FBB70()
			world.delays = tc.delays
			world.duration = 0
			playerExecuteAbility4FBB70(unit, server.AbilityTreadLightly, world.hooks())
			stored := world.cooldowns[playerExecuteAbilityCooldownKey4FBB70{index: 7, ability: server.AbilityTreadLightly}]
			if uint32(stored) != tc.wantStored {
				t.Fatalf("stored cooldown = %#08x, want %#08x", uint32(stored), tc.wantStored)
			}
			found := false
			for _, event := range world.events {
				if event == "state-report:unit:4=0" {
					found = true
				}
			}
			if found != tc.wantReport {
				t.Fatalf("state report = %v, want %v; events %v", found, tc.wantReport, world.events)
			}
		})
	}
}

func TestPlayerExecuteAbility4FBB70DurationAllocationAndLiveHead(t *testing.T) {
	for _, duration := range []int32{math.MinInt32, -1, 0} {
		t.Run(fmt.Sprintf("duration-%08x", uint32(duration)), func(t *testing.T) {
			unit, _, _, world := playerExecuteAbilityFixture4FBB70()
			world.duration = duration
			playerExecuteAbility4FBB70(unit, server.AbilityTreadLightly, world.hooks())
			for _, event := range world.events {
				if event == "alloc=new" {
					t.Fatalf("nonpositive duration allocated: %v", world.events)
				}
			}
		})
	}

	t.Run("allocation failure skips every record access", func(t *testing.T) {
		unit, _, _, world := playerExecuteAbilityFixture4FBB70()
		world.allocated = nil
		playerExecuteAbility4FBB70(unit, server.AbilityTreadLightly, world.hooks())
		want := []string{
			"duration:4=00000007", "alloc=nil", "invoke:unit:4",
			"sound:4:0=76543210", "audio:76543210:unit:0:00000000",
		}
		if !playerExecuteAbilityHasSuffix4FBB70(world.events, want) {
			t.Fatalf("events = %v, want suffix %v", world.events, want)
		}
	})

	t.Run("head is reloaded for backlink", func(t *testing.T) {
		unit, _, _, world := playerExecuteAbilityFixture4FBB70()
		old := world.head
		replacement := &playerExecuteAbilityTestExec4FBB70{name: "replacement"}
		world.afterEvent["head=old"] = func() {
			delete(world.afterEvent, "head=old")
			world.head = replacement
		}
		playerExecuteAbility4FBB70(unit, server.AbilityTreadLightly, world.hooks())
		if world.allocated.next != old || replacement.prev != world.allocated || old.prev != nil || world.head != world.allocated {
			t.Fatalf("live head result: new.next=%s replacement.prev=%s old.prev=%s head=%s",
				playerExecuteAbilityExecName4FBB70(world.allocated.next),
				playerExecuteAbilityExecName4FBB70(replacement.prev),
				playerExecuteAbilityExecName4FBB70(old.prev),
				playerExecuteAbilityExecName4FBB70(world.head),
			)
		}
	})
}

func TestPlayerExecuteAbility4FBB70EveryObservableSuccessFaultPrefix(t *testing.T) {
	want := playerExecuteAbilitySuccessTrace4FBB70()
	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("fault-%d", faultAt), func(t *testing.T) {
			unit, _, _, world := playerExecuteAbilityFixture4FBB70()
			world.faultAt = faultAt
			deferred := false
			defer func() {
				deferred = true
				if got := recover(); got != want[faultAt-1] {
					t.Fatalf("panic = %v, want %q", got, want[faultAt-1])
				}
				if prefix := want[:faultAt]; !reflect.DeepEqual(world.events, prefix) {
					t.Fatalf("events = %v, want prefix %v", world.events, prefix)
				}
			}()
			playerExecuteAbility4FBB70(unit, server.AbilityTreadLightly, world.hooks())
			if !deferred {
				t.Fatal("fault did not panic")
			}
		})
	}
}
