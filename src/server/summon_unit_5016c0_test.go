package server

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/opennox/libs/types"
)

const (
	summonUnitTestOwner5016C0         = uint64(0x7fae173753d0)
	summonUnitTestCreated5016C0       = uint64(0x7fae175b47d0)
	summonUnitTestPosition5016C0      = uint64(0x7fae20920d80)
	summonUnitTestMonsterUpdate5016C0 = uint64(0x7fae300002ec)
	summonUnitTestPlayerUpdate5016C0  = uint64(0x7fae400002ec)
	summonUnitTestTeam5016C0          = uint64(0x7fae17375400)
)

type summonUnitTestWorld5016C0 struct {
	events          []string
	created         uint64
	positionX       float32
	positionY       float32
	monsterUpdate   uint64
	status          uint32
	classLow        uint8
	playerUpdate    uint64
	players         []uint64
	playerLoads     int
	orders          map[uint64]uint32
	indices         map[uint64]uint8
	subclass        uint32
	team            uint64
	hasTeamValue    bool
	netCode         uint32
	teamID          uint8
	storedDirection [2]uint16
	storedAction    map[uint64]uint32
}

func newSummonUnitTestWorld5016C0() *summonUnitTestWorld5016C0 {
	return &summonUnitTestWorld5016C0{
		created:       summonUnitTestCreated5016C0,
		positionX:     12.5,
		positionY:     -7.25,
		monsterUpdate: summonUnitTestMonsterUpdate5016C0,
		status:        0x12345601,
		classLow:      summonUnitPlayerClass5016C0,
		playerUpdate:  summonUnitTestPlayerUpdate5016C0,
		players: []uint64{
			0x7fae50000000,
			0x7fae50000100,
			0x7fae50000200,
			0x7fae50000300,
		},
		orders: map[uint64]uint32{
			0x7fae50000000: 0x89abcdef,
		},
		indices: map[uint64]uint8{
			0x7fae50000100: 0x11,
			0x7fae50000200: 0x22,
			0x7fae50000300: 0x33,
		},
		subclass:     0x40000000,
		team:         summonUnitTestTeam5016C0,
		hasTeamValue: true,
		netCode:      0x89abcdef,
		teamID:       0x7e,
		storedAction: make(map[uint64]uint32),
	}
}

func (w *summonUnitTestWorld5016C0) observe(format string, args ...any) {
	w.events = append(w.events, fmt.Sprintf(format, args...))
}

func (w *summonUnitTestWorld5016C0) hooks(t *testing.T) summonUnitHooks5016C0[
	uint64,
	uint64,
	uint64,
	uint64,
	uint64,
	uint64,
] {
	t.Helper()
	return summonUnitHooks5016C0[uint64, uint64, uint64, uint64, uint64, uint64]{
		newObject: func(typeID int32) uint64 {
			w.observe("new:%08x", uint32(typeID))
			return w.created
		},
		loadPositionY: func(position uint64) float32 {
			w.observe("y:%x", position)
			return w.positionY
		},
		loadPositionX: func(position uint64) float32 {
			w.observe("x:%x", position)
			return w.positionX
		},
		createObjectAt: func(created, owner uint64, position types.Pointf) {
			w.observe("create:%x:%x:%g:%g", created, owner, position.X, position.Y)
			if position != (types.Pointf{X: 12.5, Y: -7.25}) {
				t.Fatalf("create position = %+v", position)
			}
		},
		loadDirection: func(direction uint8) uint16 {
			w.observe("direction:%02x", direction)
			return uint16(direction)
		},
		loadMonsterUpdate: func(created uint64) uint64 {
			w.observe("monster-update:%x", created)
			return w.monsterUpdate
		},
		storeDirection1: func(created uint64, direction uint16) {
			w.observe("direction1:%x:%04x", created, direction)
			w.storedDirection[0] = direction
		},
		storeDirection2: func(created uint64, direction uint16) {
			w.observe("direction2:%x:%04x", created, direction)
			w.storedDirection[1] = direction
		},
		loadMonsterStatus: func(update uint64) uint32 {
			w.observe("status:%x", update)
			return w.status
		},
		storeMonsterStatus: func(update uint64, status uint32) {
			w.observe("store-status:%x:%08x", update, status)
			w.status = status
		},
		loadClassLow: func(owner uint64) uint8 {
			w.observe("class:%x", owner)
			return w.classLow
		},
		loadPlayerUpdate: func(owner uint64) uint64 {
			w.observe("player-update:%x", owner)
			return w.playerUpdate
		},
		loadPlayer: func(update uint64) uint64 {
			w.observe("player:%x", update)
			if w.playerLoads >= len(w.players) {
				t.Fatalf("too many player reloads: %d", w.playerLoads+1)
			}
			player := w.players[w.playerLoads]
			w.playerLoads++
			return player
		},
		loadSummonOrder: func(player uint64) uint32 {
			w.observe("summon-order:%x", player)
			return w.orders[player]
		},
		orderUnit: func(owner, created uint64, order uint32) {
			w.observe("order:%x:%x:%08x", owner, created, order)
			w.monsterUpdate = 0x7fae600002ec
		},
		storeAIAction: func(update uint64, action uint32) {
			w.observe("action:%x:%08x", update, action)
			w.storedAction[update] = action
		},
		loadSubclass: func(created uint64) uint32 {
			w.observe("subclass:%x", created)
			return w.subclass
		},
		storeSubclass: func(created uint64, subclass uint32) {
			w.observe("store-subclass:%x:%08x", created, subclass)
			w.subclass = subclass
		},
		loadPlayerIndex: func(player uint64) uint8 {
			w.observe("index:%x", player)
			return w.indices[player]
		},
		reportAcquire: func(index uint8, created uint64) {
			w.observe("acquire:%02x:%x", index, created)
		},
		markMinimap: func(index uint8, created uint64, flags uint32) {
			w.observe("minimap:%02x:%x:%08x", index, created, flags)
		},
		sendSimpleObject: func(index uint8, created uint64) {
			w.observe("simple:%02x:%x", index, created)
		},
		loadTeam: func(owner uint64) uint64 {
			w.observe("team:%x", owner)
			return w.team
		},
		hasTeam: func(team uint64) bool {
			w.observe("has-team:%x", team)
			w.team = 0x7fae70000030
			w.netCode = 0xfedcba98
			w.teamID = 0xa5
			return w.hasTeamValue
		},
		loadNetCode: func(created uint64) uint32 {
			w.observe("net-code:%x", created)
			return w.netCode
		},
		loadTeamID: func(owner uint64) uint8 {
			w.observe("team-id:%x", owner)
			return w.teamID
		},
		createTeam: func(id uint8, team uint64, active, netCode, flags uint32) {
			w.observe("create-team:%02x:%x:%d:%08x:%d", id, team, active, netCode, flags)
			w.subclass = 0x20000000
		},
	}
}

func TestSummonUnit5016C0PlayerExactOrderAndNativeHandles(t *testing.T) {
	w := newSummonUnitTestWorld5016C0()
	got := summonUnitAt5016C0(
		int32(-0x76543211),
		summonUnitTestPosition5016C0,
		summonUnitTestOwner5016C0,
		0xab,
		w.hooks(t),
	)
	if got != summonUnitTestCreated5016C0 {
		t.Fatalf("created = %#x, want %#x", got, uint64(summonUnitTestCreated5016C0))
	}

	wantEvents := []string{
		"new:89abcdef",
		"y:7fae20920d80",
		"x:7fae20920d80",
		"create:7fae175b47d0:7fae173753d0:12.5:-7.25",
		"direction:ab",
		"monster-update:7fae175b47d0",
		"direction2:7fae175b47d0:00ab",
		"direction1:7fae175b47d0:00ab",
		"status:7fae300002ec",
		"store-status:7fae300002ec:12345681",
		"class:7fae173753d0",
		"player-update:7fae173753d0",
		"player:7fae400002ec",
		"summon-order:7fae50000000",
		"order:7fae173753d0:7fae175b47d0:89abcdef",
		"action:7fae300002ec:00000026",
		"subclass:7fae175b47d0",
		"store-subclass:7fae175b47d0:40000080",
		"player:7fae400002ec",
		"index:7fae50000100",
		"acquire:11:7fae175b47d0",
		"player:7fae400002ec",
		"index:7fae50000200",
		"minimap:22:7fae175b47d0:00000001",
		"player:7fae400002ec",
		"index:7fae50000300",
		"simple:33:7fae175b47d0",
		"team:7fae173753d0",
		"has-team:7fae17375400",
		"net-code:7fae175b47d0",
		"team-id:7fae173753d0",
		"create-team:a5:7fae17375400:1:fedcba98:0",
		"subclass:7fae175b47d0",
		"store-subclass:7fae175b47d0:20000100",
	}
	if !reflect.DeepEqual(w.events, wantEvents) {
		t.Fatalf("events =\n%q\nwant\n%q", w.events, wantEvents)
	}
	if w.storedDirection != [2]uint16{0xab, 0xab} {
		t.Fatalf("directions = %#v", w.storedDirection)
	}
	if got := w.storedAction[summonUnitTestMonsterUpdate5016C0]; got != summonUnitInvalidAction5016C0 {
		t.Fatalf("cached update action = %#x, want %#x", got, summonUnitInvalidAction5016C0)
	}
	if _, ok := w.storedAction[w.monsterUpdate]; ok {
		t.Fatal("order callback replacement update received the action store")
	}
	if w.status != 0x12345681 || w.subclass != 0x20000100 || w.playerLoads != 4 {
		t.Fatalf("final state = status %#x, subclass %#x, player loads %d", w.status, w.subclass, w.playerLoads)
	}
}

func TestSummonUnit5016C0AllocationFailureDoesNotReadOtherArguments(t *testing.T) {
	var events []string
	hooks := summonUnitHooks5016C0[uint64, uint64, uint64, uint64, uint64, uint64]{
		newObject: func(typeID int32) uint64 {
			events = append(events, fmt.Sprintf("new:%08x", uint32(typeID)))
			return 0
		},
		loadPositionY: func(uint64) float32 {
			t.Fatal("allocation failure read position Y")
			return 0
		},
		loadPositionX: func(uint64) float32 {
			t.Fatal("allocation failure read position X")
			return 0
		},
	}
	if got := summonUnitAt5016C0(int32(-1), summonUnitTestPosition5016C0, summonUnitTestOwner5016C0, 0xff, hooks); got != 0 {
		t.Fatalf("created = %#x, want nil handle", got)
	}
	if want := []string{"new:ffffffff"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %q, want %q", events, want)
	}
}

func TestSummonUnit5016C0NonPlayerUsesFixedOrderAndCachedUpdate(t *testing.T) {
	w := newSummonUnitTestWorld5016C0()
	w.classLow = 0x82
	w.hasTeamValue = false
	got := summonUnitAt5016C0(7, summonUnitTestPosition5016C0, summonUnitTestOwner5016C0, 0x7f, w.hooks(t))
	if got != summonUnitTestCreated5016C0 {
		t.Fatalf("created = %#x", got)
	}
	wantSuffix := []string{
		"class:7fae173753d0",
		"order:7fae173753d0:7fae175b47d0:00000004",
		"action:7fae300002ec:00000026",
	}
	if len(w.events) < len(wantSuffix) || !reflect.DeepEqual(w.events[len(w.events)-len(wantSuffix):], wantSuffix) {
		t.Fatalf("event suffix = %q, want %q", w.events, wantSuffix)
	}
	if w.playerLoads != 0 {
		t.Fatalf("non-player path reloaded %d players", w.playerLoads)
	}
}

func TestSummonUnit5016C0NilOwnerStopsAfterSummonedStatus(t *testing.T) {
	w := newSummonUnitTestWorld5016C0()
	got := summonUnitAt5016C0(7, summonUnitTestPosition5016C0, uint64(0), 0x35, w.hooks(t))
	if got != summonUnitTestCreated5016C0 {
		t.Fatalf("created = %#x", got)
	}
	if got := w.events[len(w.events)-1]; got != "store-status:7fae300002ec:12345681" {
		t.Fatalf("last event = %q", got)
	}
}
