package opennox

import (
	"fmt"
	"math"
	"reflect"
	"testing"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/server"
)

type playerAbilityRuntimeTestPlayer4FBEE0 struct {
	name  string
	unit  *playerAbilityRuntimeTestUnit4FBEE0
	class uint8
	index uint8
	next  *playerAbilityRuntimeTestPlayer4FBEE0
}

type playerAbilityRuntimeTestUnit4FBEE0 struct {
	name  string
	flags object.Flags
}

type playerAbilityRuntimeTestExec4FBEE0 struct {
	name    string
	unit    *playerAbilityRuntimeTestUnit4FBEE0
	ability server.Ability
	frame   uint32
	next    *playerAbilityRuntimeTestExec4FBEE0
	prev    *playerAbilityRuntimeTestExec4FBEE0
	freed   bool
}

type playerAbilityRuntimeCooldownKey4FBEE0 struct {
	index   uint8
	ability server.Ability
}

type playerAbilityRuntimeTestWorld4FBEE0 struct {
	first     *playerAbilityRuntimeTestPlayer4FBEE0
	head      *playerAbilityRuntimeTestExec4FBEE0
	frame     uint32
	cooldowns map[playerAbilityRuntimeCooldownKey4FBEE0]int32
	events    []string
	after     map[string]func()
}

func playerAbilityRuntimePlayerName4FBEE0(player *playerAbilityRuntimeTestPlayer4FBEE0) string {
	if player == nil {
		return "nil"
	}
	return player.name
}

func playerAbilityRuntimeUnitName4FBEE0(unit *playerAbilityRuntimeTestUnit4FBEE0) string {
	if unit == nil {
		return "nil"
	}
	return unit.name
}

func playerAbilityRuntimeExecName4FBEE0(exec *playerAbilityRuntimeTestExec4FBEE0) string {
	if exec == nil {
		return "nil"
	}
	return exec.name
}

func (w *playerAbilityRuntimeTestWorld4FBEE0) record(event string) {
	w.events = append(w.events, event)
	if after := w.after[event]; after != nil {
		after()
	}
}

func (w *playerAbilityRuntimeTestWorld4FBEE0) hooks() playerAbilityRuntimeTickHooks4FBEE0[
	*playerAbilityRuntimeTestPlayer4FBEE0,
	*playerAbilityRuntimeTestUnit4FBEE0,
	*playerAbilityRuntimeTestExec4FBEE0,
] {
	return playerAbilityRuntimeTickHooks4FBEE0[
		*playerAbilityRuntimeTestPlayer4FBEE0,
		*playerAbilityRuntimeTestUnit4FBEE0,
		*playerAbilityRuntimeTestExec4FBEE0,
	]{
		firstPlayer: func() *playerAbilityRuntimeTestPlayer4FBEE0 {
			w.record("first=" + playerAbilityRuntimePlayerName4FBEE0(w.first))
			return w.first
		},
		nextPlayer: func(player *playerAbilityRuntimeTestPlayer4FBEE0) *playerAbilityRuntimeTestPlayer4FBEE0 {
			next := player.next
			w.record("next:" + player.name + "=" + playerAbilityRuntimePlayerName4FBEE0(next))
			return next
		},
		loadPlayerUnit: func(player *playerAbilityRuntimeTestPlayer4FBEE0) *playerAbilityRuntimeTestUnit4FBEE0 {
			unit := player.unit
			w.record("player-unit:" + player.name + "=" + playerAbilityRuntimeUnitName4FBEE0(unit))
			return unit
		},
		loadPlayerClass: func(player *playerAbilityRuntimeTestPlayer4FBEE0) uint8 {
			w.record(fmt.Sprintf("player-class:%s=%d", player.name, player.class))
			return player.class
		},
		loadPlayerIndex: func(player *playerAbilityRuntimeTestPlayer4FBEE0) uint8 {
			w.record(fmt.Sprintf("player-index:%s=%d", player.name, player.index))
			return player.index
		},
		loadCooldown: func(index uint8, ability server.Ability) int32 {
			value := w.cooldowns[playerAbilityRuntimeCooldownKey4FBEE0{index: index, ability: ability}]
			w.record(fmt.Sprintf("cooldown:%d:%d=%d", index, ability, value))
			return value
		},
		storeCooldown: func(index uint8, ability server.Ability, value int32) {
			w.cooldowns[playerAbilityRuntimeCooldownKey4FBEE0{index: index, ability: ability}] = value
			w.record(fmt.Sprintf("cooldown-store:%d:%d=%d", index, ability, value))
		},
		reportState: func(unit *playerAbilityRuntimeTestUnit4FBEE0, ability server.Ability, state uint8) {
			w.record(fmt.Sprintf("state:%s:%d=%d", playerAbilityRuntimeUnitName4FBEE0(unit), ability, state))
		},
		loadExecHead: func() *playerAbilityRuntimeTestExec4FBEE0 {
			w.record("head=" + playerAbilityRuntimeExecName4FBEE0(w.head))
			return w.head
		},
		storeExecHead: func(exec *playerAbilityRuntimeTestExec4FBEE0) {
			w.head = exec
			w.record("head-store=" + playerAbilityRuntimeExecName4FBEE0(exec))
		},
		loadExecUnit: func(exec *playerAbilityRuntimeTestExec4FBEE0) *playerAbilityRuntimeTestUnit4FBEE0 {
			unit := exec.unit
			w.record("exec-unit:" + exec.name + "=" + playerAbilityRuntimeUnitName4FBEE0(unit))
			return unit
		},
		loadExecAbility: func(exec *playerAbilityRuntimeTestExec4FBEE0) server.Ability {
			w.record(fmt.Sprintf("exec-ability:%s=%d", exec.name, exec.ability))
			return exec.ability
		},
		loadExecFrame: func(exec *playerAbilityRuntimeTestExec4FBEE0) uint32 {
			w.record(fmt.Sprintf("exec-frame:%s=%d", exec.name, exec.frame))
			return exec.frame
		},
		loadExecNext: func(exec *playerAbilityRuntimeTestExec4FBEE0) *playerAbilityRuntimeTestExec4FBEE0 {
			next := exec.next
			w.record("exec-next:" + exec.name + "=" + playerAbilityRuntimeExecName4FBEE0(next))
			return next
		},
		loadExecPrev: func(exec *playerAbilityRuntimeTestExec4FBEE0) *playerAbilityRuntimeTestExec4FBEE0 {
			prev := exec.prev
			w.record("exec-prev:" + exec.name + "=" + playerAbilityRuntimeExecName4FBEE0(prev))
			return prev
		},
		storeExecNext: func(exec, next *playerAbilityRuntimeTestExec4FBEE0) {
			exec.next = next
			w.record("exec-next-store:" + exec.name + "=" + playerAbilityRuntimeExecName4FBEE0(next))
		},
		storeExecPrev: func(exec, prev *playerAbilityRuntimeTestExec4FBEE0) {
			exec.prev = prev
			w.record("exec-prev-store:" + exec.name + "=" + playerAbilityRuntimeExecName4FBEE0(prev))
		},
		loadUnitFlags: func(unit *playerAbilityRuntimeTestUnit4FBEE0) object.Flags {
			if unit == nil {
				w.record("flags:nil")
				panic("nil unit")
			}
			w.record(fmt.Sprintf("flags:%s=%08x", unit.name, uint32(unit.flags)))
			return unit.flags
		},
		loadFrame: func() uint32 {
			w.record(fmt.Sprintf("frame=%d", w.frame))
			return w.frame
		},
		loadEndingSound: func(ability server.Ability, slot int32) int32 {
			w.record(fmt.Sprintf("sound:%d:%d", ability, slot))
			return 300 + int32(ability)
		},
		audio: func(soundID int32, unit *playerAbilityRuntimeTestUnit4FBEE0, kind int32, code uint32) {
			w.record(fmt.Sprintf("audio:%d:%s:%d:%d", soundID, playerAbilityRuntimeUnitName4FBEE0(unit), kind, code))
		},
		reportActive: func(unit *playerAbilityRuntimeTestUnit4FBEE0, ability server.Ability, active uint8) {
			w.record(fmt.Sprintf("active:%s:%d=%d", playerAbilityRuntimeUnitName4FBEE0(unit), ability, active))
		},
		setPlayerState: func(unit *playerAbilityRuntimeTestUnit4FBEE0, state uint8) {
			w.record(fmt.Sprintf("player-state:%s=%d", playerAbilityRuntimeUnitName4FBEE0(unit), state))
		},
		freeExec: func(exec *playerAbilityRuntimeTestExec4FBEE0) {
			exec.freed = true
			w.record("free=" + exec.name)
		},
	}
}

func TestPlayerAbilityRuntimeTick4FBEE0CooldownReloadAndInt32Wrap(t *testing.T) {
	unitInitial := &playerAbilityRuntimeTestUnit4FBEE0{name: "initial"}
	unitLive := &playerAbilityRuntimeTestUnit4FBEE0{name: "live"}
	warrior := &playerAbilityRuntimeTestPlayer4FBEE0{name: "warrior", unit: unitInitial, index: 7}
	nonWarrior := &playerAbilityRuntimeTestPlayer4FBEE0{name: "wizard", unit: unitInitial, class: 1}
	withoutUnit := &playerAbilityRuntimeTestPlayer4FBEE0{name: "empty"}
	warrior.next = nonWarrior
	nonWarrior.next = withoutUnit
	world := &playerAbilityRuntimeTestWorld4FBEE0{
		first: warrior,
		cooldowns: map[playerAbilityRuntimeCooldownKey4FBEE0]int32{
			{index: 7, ability: server.AbilityInvalid}: math.MinInt32,
			{index: 7, ability: server.AbilityBerserk}: 1,
		},
		after: make(map[string]func()),
	}
	world.after["cooldown-store:7:1=0"] = func() {
		warrior.index = 8
	}
	world.after["cooldown:8:1=0"] = func() {
		warrior.unit = unitLive
	}
	world.after["state:live:1=1"] = func() {
		warrior.index = 7
	}

	playerAbilityRuntimeTick4FBEE0(world.hooks())

	if got, want := world.cooldowns[playerAbilityRuntimeCooldownKey4FBEE0{index: 7, ability: server.AbilityInvalid}], int32(math.MaxInt32); got != want {
		t.Fatalf("INT32_MIN decrement = %d, want %d", got, want)
	}
	if got := world.cooldowns[playerAbilityRuntimeCooldownKey4FBEE0{index: 7, ability: server.AbilityBerserk}]; got != 0 {
		t.Fatalf("Berserk cooldown = %d, want 0", got)
	}
	want := []string{
		"first=warrior",
		"player-unit:warrior=initial", "player-class:warrior=0",
		"player-index:warrior=7", "cooldown:7:0=-2147483648", "cooldown-store:7:0=2147483647", "player-index:warrior=7", "cooldown:7:0=2147483647",
		"player-index:warrior=7", "cooldown:7:1=1", "cooldown-store:7:1=0", "player-index:warrior=8", "cooldown:8:1=0", "player-unit:warrior=live", "state:live:1=1",
		"player-index:warrior=7", "cooldown:7:2=0",
		"player-index:warrior=7", "cooldown:7:3=0",
		"player-index:warrior=7", "cooldown:7:4=0",
		"player-index:warrior=7", "cooldown:7:5=0",
		"next:warrior=wizard", "player-unit:wizard=initial", "player-class:wizard=1",
		"next:wizard=empty", "player-unit:empty=nil", "next:empty=nil", "head=nil",
	}
	if !reflect.DeepEqual(world.events, want) {
		t.Fatalf("event order:\n got: %q\nwant: %q", world.events, want)
	}
}

func TestPlayerAbilityRuntimeTick4FBEE0ActiveCallbacksAndLiveUnlink(t *testing.T) {
	deadUnit := &playerAbilityRuntimeTestUnit4FBEE0{name: "dead", flags: object.FlagDead}
	startUnit := &playerAbilityRuntimeTestUnit4FBEE0{name: "start"}
	reportUnit := &playerAbilityRuntimeTestUnit4FBEE0{name: "report"}
	stateUnit := &playerAbilityRuntimeTestUnit4FBEE0{name: "state"}
	futureUnit := &playerAbilityRuntimeTestUnit4FBEE0{name: "future"}
	mutatedUnit := &playerAbilityRuntimeTestUnit4FBEE0{name: "mutated"}

	mutated := &playerAbilityRuntimeTestExec4FBEE0{name: "mutated", unit: mutatedUnit, frame: 0}
	future := &playerAbilityRuntimeTestExec4FBEE0{name: "future", unit: futureUnit, ability: server.AbilityHarpoon, frame: 100}
	expired := &playerAbilityRuntimeTestExec4FBEE0{name: "expired", unit: startUnit, ability: server.AbilityBerserk, frame: 99, next: future}
	dead := &playerAbilityRuntimeTestExec4FBEE0{name: "dead", unit: deadUnit, ability: server.AbilityWarcry, next: expired}
	expired.prev = dead
	future.prev = expired
	world := &playerAbilityRuntimeTestWorld4FBEE0{
		head:      dead,
		frame:     100,
		cooldowns: make(map[playerAbilityRuntimeCooldownKey4FBEE0]int32),
		after:     make(map[string]func()),
	}
	world.after["audio:301:start:0:0"] = func() {
		expired.ability = server.AbilityWarcry
		expired.unit = reportUnit
	}
	world.after["active:report:2=0"] = func() {
		expired.ability = server.AbilityBerserk
		expired.unit = stateUnit
		expired.next = mutated
	}

	playerAbilityRuntimeTick4FBEE0(world.hooks())

	if !dead.freed || !expired.freed || future.freed || mutated.freed {
		t.Fatalf("freed state dead/expired/future/mutated = %v/%v/%v/%v", dead.freed, expired.freed, future.freed, mutated.freed)
	}
	if world.head != mutated || mutated.prev != nil {
		t.Fatalf("live unlink head/prev = %s/%s, want mutated/nil", playerAbilityRuntimeExecName4FBEE0(world.head), playerAbilityRuntimeExecName4FBEE0(mutated.prev))
	}
	want := []string{
		"first=nil", "head=dead",
		"exec-unit:dead=dead", "exec-next:dead=expired", "flags:dead=00008000",
		"exec-next:dead=expired", "exec-prev:dead=nil", "exec-prev-store:expired=nil", "exec-prev:dead=nil", "exec-next:dead=expired", "head-store=expired", "free=dead",
		"exec-unit:expired=start", "exec-next:expired=future", "flags:start=00000000", "frame=100", "exec-frame:expired=99",
		"exec-ability:expired=1", "sound:1:2", "audio:301:start:0:0",
		"exec-ability:expired=2", "exec-unit:expired=report", "active:report:2=0",
		"exec-ability:expired=1", "exec-unit:expired=state", "player-state:state=13",
		"exec-next:expired=mutated", "exec-prev:expired=nil", "exec-prev-store:mutated=nil", "exec-prev:expired=nil", "exec-next:expired=mutated", "head-store=mutated", "free=expired",
		"exec-unit:future=future", "exec-next:future=nil", "flags:future=00000000", "frame=100", "exec-frame:future=100",
	}
	if !reflect.DeepEqual(world.events, want) {
		t.Fatalf("event order:\n got: %q\nwant: %q", world.events, want)
	}
}

func TestPlayerAbilityRuntimeTick4FBEE0UnsignedFrameAndNilExecUnitFault(t *testing.T) {
	t.Run("unsigned preserve", func(t *testing.T) {
		unit := &playerAbilityRuntimeTestUnit4FBEE0{name: "unit"}
		exec := &playerAbilityRuntimeTestExec4FBEE0{name: "exec", unit: unit, frame: math.MaxUint32}
		world := &playerAbilityRuntimeTestWorld4FBEE0{
			head:      exec,
			cooldowns: make(map[playerAbilityRuntimeCooldownKey4FBEE0]int32),
			after:     make(map[string]func()),
		}
		playerAbilityRuntimeTick4FBEE0(world.hooks())
		if exec.freed {
			t.Fatal("deadline UINT32_MAX was treated as signed and expired at frame zero")
		}
	})

	t.Run("nil unit faults after cached successor", func(t *testing.T) {
		exec := &playerAbilityRuntimeTestExec4FBEE0{name: "exec"}
		world := &playerAbilityRuntimeTestWorld4FBEE0{
			head:      exec,
			cooldowns: make(map[playerAbilityRuntimeCooldownKey4FBEE0]int32),
			after:     make(map[string]func()),
		}
		defer func() {
			if recover() == nil {
				t.Fatal("nil active-record Unit did not fault")
			}
			want := []string{"first=nil", "head=exec", "exec-unit:exec=nil", "exec-next:exec=nil", "flags:nil"}
			if !reflect.DeepEqual(world.events, want) {
				t.Fatalf("fault prefix:\n got: %q\nwant: %q", world.events, want)
			}
		}()
		playerAbilityRuntimeTick4FBEE0(world.hooks())
	})
}
