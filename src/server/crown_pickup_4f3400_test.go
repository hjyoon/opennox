package server

import (
	"fmt"
	"reflect"
	"testing"
)

type crownPickupTestObject4F3400 struct {
	name       string
	class      uint8
	update     *crownPickupTestUpdate4F3400
	playerData *crownPickupTestPlayerUpdate4F3400
	netCode    uint32
	team       bool
	teamID     uint8
}

type crownPickupTestUpdate4F3400 struct {
	name    string
	pending *crownPickupTestObject4F3400
}

type crownPickupTestPlayerUpdate4F3400 struct {
	name  string
	frame uint32
}

type crownPickupTestState4F3400 struct {
	events       []string
	pickupResult uint32
	frame        uint32
	afterDefault func()
	afterAudio   func()
	afterHasTeam func()
}

func crownPickupTestObjectName4F3400(obj *crownPickupTestObject4F3400) string {
	if obj == nil {
		return "nil"
	}
	return obj.name
}

func (s *crownPickupTestState4F3400) hooks() crownPickupHooks4F3400[
	*crownPickupTestObject4F3400,
	*crownPickupTestUpdate4F3400,
	*crownPickupTestPlayerUpdate4F3400,
] {
	return crownPickupHooks4F3400[
		*crownPickupTestObject4F3400,
		*crownPickupTestUpdate4F3400,
		*crownPickupTestPlayerUpdate4F3400,
	]{
		loadCrownUpdate: func(obj *crownPickupTestObject4F3400) *crownPickupTestUpdate4F3400 {
			s.events = append(s.events, "crown-update:"+crownPickupTestObjectName4F3400(obj))
			return obj.update
		},
		loadClassLow: func(obj *crownPickupTestObject4F3400) uint8 {
			s.events = append(s.events, "class:"+crownPickupTestObjectName4F3400(obj))
			return obj.class
		},
		defaultPickup: func(who, crown *crownPickupTestObject4F3400, flag1, flag2 int32) uint32 {
			s.events = append(s.events, fmt.Sprintf(
				"default:%s:%s:%d:%d",
				crownPickupTestObjectName4F3400(who),
				crownPickupTestObjectName4F3400(crown),
				flag1,
				flag2,
			))
			if s.afterDefault != nil {
				s.afterDefault()
			}
			return s.pickupResult
		},
		loadPlayerUpdate: func(obj *crownPickupTestObject4F3400) *crownPickupTestPlayerUpdate4F3400 {
			s.events = append(s.events, "player-update:"+crownPickupTestObjectName4F3400(obj))
			return obj.playerData
		},
		loadFrame: func() uint32 {
			s.events = append(s.events, "frame")
			return s.frame
		},
		storePickupFrame: func(update *crownPickupTestPlayerUpdate4F3400, frame uint32) {
			s.events = append(s.events, fmt.Sprintf("store-frame:%s:%#x", update.name, frame))
			update.frame = frame
		},
		setOwner: func(owner, crown *crownPickupTestObject4F3400) {
			s.events = append(s.events, "owner:"+owner.name+":"+crown.name)
		},
		applyEnchant: func(obj *crownPickupTestObject4F3400, enchant, duration, power uint32) {
			s.events = append(s.events, fmt.Sprintf(
				"enchant:%s:%d:%d:%d", obj.name, enchant, duration, power,
			))
		},
		playAudio: func(id uint32, obj *crownPickupTestObject4F3400, kind int32, code uint32) {
			s.events = append(s.events, fmt.Sprintf(
				"audio:%d:%s:%d:%d", id, obj.name, kind, code,
			))
			if s.afterAudio != nil {
				s.afterAudio()
			}
		},
		loadNetCode: func(obj *crownPickupTestObject4F3400) uint32 {
			s.events = append(s.events, "net-code:"+obj.name)
			return obj.netCode
		},
		hasTeam: func(obj *crownPickupTestObject4F3400) bool {
			s.events = append(s.events, "has-team:"+obj.name)
			if s.afterHasTeam != nil {
				s.afterHasTeam()
			}
			return obj.team
		},
		loadTeamID: func(obj *crownPickupTestObject4F3400) uint8 {
			s.events = append(s.events, "team-id:"+obj.name)
			return obj.teamID
		},
		informPickup: func(code uint8, netCode, teamID uint32) {
			s.events = append(s.events, fmt.Sprintf(
				"inform:%d:%#x:%#x", code, netCode, teamID,
			))
		},
		unmarkMinimap: func(obj *crownPickupTestObject4F3400, flags uint32) {
			s.events = append(s.events, fmt.Sprintf("unmark:%s:%d", obj.name, flags))
		},
		clearPending: func(update *crownPickupTestUpdate4F3400) {
			s.events = append(s.events, "clear:"+update.name)
			update.pending = nil
		},
	}
}

func assertCrownPickupEvents4F3400(t *testing.T, got, want []string) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("events = %#v, want %#v", got, want)
	}
}

func TestCrownPickup4F3400NonPlayerCachesUpdateWithoutClearing(t *testing.T) {
	pending := &crownPickupTestObject4F3400{name: "pending"}
	update := &crownPickupTestUpdate4F3400{name: "cached", pending: pending}
	crown := &crownPickupTestObject4F3400{name: "crown", update: update}
	who := &crownPickupTestObject4F3400{name: "monster", class: 0x82}
	state := &crownPickupTestState4F3400{pickupResult: 1}

	if got := crownPickup4F3400(who, crown, 7, 9, state.hooks()); got != 0 {
		t.Fatalf("result = %#x, want 0", got)
	}
	if update.pending != pending {
		t.Fatal("non-Player path cleared cached pending target")
	}
	assertCrownPickupEvents4F3400(t, state.events, []string{
		"crown-update:crown",
		"class:monster",
	})
}

func TestCrownPickup4F3400SuccessPreservesOrderArgumentsAndResult(t *testing.T) {
	pending := &crownPickupTestObject4F3400{name: "pending"}
	update := &crownPickupTestUpdate4F3400{name: "cached", pending: pending}
	playerUpdate := &crownPickupTestPlayerUpdate4F3400{name: "player-data"}
	crown := &crownPickupTestObject4F3400{name: "crown", update: update}
	who := &crownPickupTestObject4F3400{
		name:       "player",
		class:      0x84,
		playerData: playerUpdate,
		netCode:    0xf1234567,
		team:       true,
		teamID:     0xab,
	}
	state := &crownPickupTestState4F3400{
		pickupResult: 0x81234567,
		frame:        0xfedcba98,
	}

	got := crownPickup4F3400(who, crown, -7, 0x12345678, state.hooks())
	if got != state.pickupResult {
		t.Fatalf("result = %#x, want %#x", got, state.pickupResult)
	}
	if playerUpdate.frame != state.frame {
		t.Fatalf("stored frame = %#x, want %#x", playerUpdate.frame, state.frame)
	}
	if update.pending != nil {
		t.Fatal("success path did not clear pending target")
	}
	assertCrownPickupEvents4F3400(t, state.events, []string{
		"crown-update:crown",
		"class:player",
		"default:player:crown:-7:305419896",
		"player-update:player",
		"frame",
		"store-frame:player-data:0xfedcba98",
		"owner:player:crown",
		"enchant:player:30:0:5",
		"audio:313:player:0:0",
		"net-code:player",
		"has-team:player",
		"team-id:player",
		"inform:10:0xf1234567:0xab",
		"unmark:crown:1",
		"clear:cached",
	})
}

func TestCrownPickup4F3400FailureClearsCachedUpdateOnly(t *testing.T) {
	oldPending := &crownPickupTestObject4F3400{name: "old-pending"}
	oldUpdate := &crownPickupTestUpdate4F3400{name: "old", pending: oldPending}
	newPending := &crownPickupTestObject4F3400{name: "new-pending"}
	newUpdate := &crownPickupTestUpdate4F3400{name: "new", pending: newPending}
	crown := &crownPickupTestObject4F3400{name: "crown", update: oldUpdate}
	who := &crownPickupTestObject4F3400{name: "player", class: 4}
	state := &crownPickupTestState4F3400{
		afterDefault: func() {
			crown.update = newUpdate
		},
	}

	if got := crownPickup4F3400(who, crown, 1, 2, state.hooks()); got != 0 {
		t.Fatalf("result = %#x, want 0", got)
	}
	if oldUpdate.pending != nil {
		t.Fatal("cached update was not cleared")
	}
	if newUpdate.pending != newPending {
		t.Fatal("live replacement update was cleared instead of cached update")
	}
	assertCrownPickupEvents4F3400(t, state.events, []string{
		"crown-update:crown",
		"class:player",
		"default:player:crown:1:2",
		"clear:old",
	})
}

func TestCrownPickup4F3400UsesLivePostCallbackFields(t *testing.T) {
	initialPlayerData := &crownPickupTestPlayerUpdate4F3400{name: "initial"}
	livePlayerData := &crownPickupTestPlayerUpdate4F3400{name: "live"}
	update := &crownPickupTestUpdate4F3400{name: "crown-data"}
	crown := &crownPickupTestObject4F3400{name: "crown", update: update}
	who := &crownPickupTestObject4F3400{
		name:       "player",
		class:      4,
		playerData: initialPlayerData,
		netCode:    1,
		team:       true,
		teamID:     2,
	}
	state := &crownPickupTestState4F3400{
		pickupResult: 1,
		frame:        77,
		afterDefault: func() {
			who.playerData = livePlayerData
		},
		afterAudio: func() {
			who.netCode = 0xaabbccdd
		},
		afterHasTeam: func() {
			who.teamID = 0xee
		},
	}

	if got := crownPickup4F3400(who, crown, 3, 4, state.hooks()); got != 1 {
		t.Fatalf("result = %#x, want 1", got)
	}
	if initialPlayerData.frame != 0 || livePlayerData.frame != 77 {
		t.Fatalf("frames = (%d, %d), want (0, 77)", initialPlayerData.frame, livePlayerData.frame)
	}
	wantTail := []string{
		"audio:313:player:0:0",
		"net-code:player",
		"has-team:player",
		"team-id:player",
		"inform:10:0xaabbccdd:0xee",
		"unmark:crown:1",
		"clear:crown-data",
	}
	assertCrownPickupEvents4F3400(t, state.events[len(state.events)-len(wantTail):], wantTail)
}

func TestCrownPickup4F3400NoTeamSkipsTeamIDLoad(t *testing.T) {
	update := &crownPickupTestUpdate4F3400{name: "crown-data"}
	crown := &crownPickupTestObject4F3400{name: "crown", update: update}
	who := &crownPickupTestObject4F3400{
		name:       "player",
		class:      4,
		playerData: &crownPickupTestPlayerUpdate4F3400{name: "player-data"},
		netCode:    0x89abcdef,
		teamID:     0xff,
	}
	state := &crownPickupTestState4F3400{pickupResult: 1}

	_ = crownPickup4F3400(who, crown, 1, 1, state.hooks())
	for _, event := range state.events {
		if event == "team-id:player" {
			t.Fatal("no-team path loaded team ID")
		}
	}
	if !reflect.DeepEqual(state.events[len(state.events)-4:], []string{
		"has-team:player",
		"inform:10:0x89abcdef:0x0",
		"unmark:crown:1",
		"clear:crown-data",
	}) {
		t.Fatalf("event tail = %#v", state.events[len(state.events)-4:])
	}
}

func TestCrownPickup4F3400CrownFaultPrecedesWhoRead(t *testing.T) {
	state := &crownPickupTestState4F3400{}
	defer func() {
		if recover() == nil {
			t.Fatal("nil crown did not fault")
		}
		assertCrownPickupEvents4F3400(t, state.events, []string{"crown-update:nil"})
	}()
	crownPickup4F3400(
		(*crownPickupTestObject4F3400)(nil),
		(*crownPickupTestObject4F3400)(nil),
		1,
		1,
		state.hooks(),
	)
}
