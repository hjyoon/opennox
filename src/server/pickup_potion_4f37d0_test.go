package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type pickupPotionTestUse4F37D0 struct {
	name  string
	value int32
}

type pickupPotionTestHealth4F37D0 struct {
	name     string
	cur, max uint16
}

type pickupPotionTestPlayer4F37D0 struct {
	name  string
	class uint8
}

type pickupPotionTestUpdate4F37D0 struct {
	name    string
	player  *pickupPotionTestPlayer4F37D0
	manaCur uint16
	manaMax uint16
}

type pickupPotionTestObject4F37D0 struct {
	name     string
	classLow uint8
	subClass uint8
	netCode  uint32
	poison   uint8
	use      *pickupPotionTestUse4F37D0
	health   *pickupPotionTestHealth4F37D0
	update   *pickupPotionTestUpdate4F37D0
}

type pickupPotionTestWorld4F37D0 struct {
	flags         map[uint32]int32
	playerState   int32
	canUse        int32
	healthScale   map[uint8]int32
	manaScale     map[uint8]int32
	spellSound    uint32
	defaultResult int32
	arg3          int32
	arg4          int32
	events        []string
	faultAt       int

	afterClassFailure func()
	afterHealthScale  func()
	afterAdjustHealth func()
	afterManaScale    func()
	afterAddMana      func()
	afterRemovePoison func()
	afterDecay        func()
	afterLoadArg4     func()
}

func newPickupPotionTestWorld4F37D0() *pickupPotionTestWorld4F37D0 {
	return &pickupPotionTestWorld4F37D0{
		flags:       make(map[uint32]int32),
		canUse:      1,
		healthScale: map[uint8]int32{0: 11, 1: 12, 2: 13},
		manaScale:   map[uint8]int32{0: 21, 1: 22, 2: 23},
		spellSound:  0x456,
		arg3:        math.MinInt32,
		arg4:        math.MaxInt32,
	}
}

func pickupPotionObjectName4F37D0(obj *pickupPotionTestObject4F37D0) string {
	if obj == nil {
		return "nil"
	}
	return obj.name
}

func pickupPotionUseName4F37D0(use *pickupPotionTestUse4F37D0) string {
	if use == nil {
		return "nil"
	}
	return use.name
}

func pickupPotionHealthName4F37D0(health *pickupPotionTestHealth4F37D0) string {
	if health == nil {
		return "nil"
	}
	return health.name
}

func pickupPotionUpdateName4F37D0(update *pickupPotionTestUpdate4F37D0) string {
	if update == nil {
		return "nil"
	}
	return update.name
}

func pickupPotionPlayerName4F37D0(player *pickupPotionTestPlayer4F37D0) string {
	if player == nil {
		return "nil"
	}
	return player.name
}

func (w *pickupPotionTestWorld4F37D0) event(format string, args ...any) {
	event := fmt.Sprintf(format, args...)
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
}

func (w *pickupPotionTestWorld4F37D0) hooks() pickupPotionHooks4F37D0[
	*pickupPotionTestObject4F37D0,
	*pickupPotionTestUse4F37D0,
	*pickupPotionTestHealth4F37D0,
	*pickupPotionTestUpdate4F37D0,
	*pickupPotionTestPlayer4F37D0,
] {
	return pickupPotionHooks4F37D0[
		*pickupPotionTestObject4F37D0,
		*pickupPotionTestUse4F37D0,
		*pickupPotionTestHealth4F37D0,
		*pickupPotionTestUpdate4F37D0,
		*pickupPotionTestPlayer4F37D0,
	]{
		loadPotionUseData: func(potion *pickupPotionTestObject4F37D0) *pickupPotionTestUse4F37D0 {
			w.event("use-data:%s", pickupPotionObjectName4F37D0(potion))
			return potion.use
		},
		loadPotionValue: func(use *pickupPotionTestUse4F37D0) int32 {
			w.event("use-value:%s", pickupPotionUseName4F37D0(use))
			return use.value
		},
		gameFlag: func(flag uint32) int32 {
			value := w.flags[flag]
			w.event("game:%08x=%08x", flag, uint32(value))
			return value
		},
		loadOwnerClassLow: func(owner *pickupPotionTestObject4F37D0) uint8 {
			w.event("class:%s", pickupPotionObjectName4F37D0(owner))
			return owner.classLow
		},
		loadOwnerUpdate: func(owner *pickupPotionTestObject4F37D0) *pickupPotionTestUpdate4F37D0 {
			update := owner.update
			w.event("update:%s=%s", pickupPotionObjectName4F37D0(owner), pickupPotionUpdateName4F37D0(update))
			return update
		},
		loadUpdatePlayer: func(update *pickupPotionTestUpdate4F37D0) *pickupPotionTestPlayer4F37D0 {
			player := update.player
			w.event("player:%s=%s", pickupPotionUpdateName4F37D0(update), pickupPotionPlayerName4F37D0(player))
			return player
		},
		loadPlayerClass: func(player *pickupPotionTestPlayer4F37D0) uint8 {
			w.event("player-class:%s=%02x", pickupPotionPlayerName4F37D0(player), player.class)
			return player.class
		},
		playerClassCanUse: func(potion *pickupPotionTestObject4F37D0, class uint8) int32 {
			w.event("can-use:%s:%02x=%08x", pickupPotionObjectName4F37D0(potion), class, uint32(w.canUse))
			return w.canUse
		},
		classFailureMessage: func(owner *pickupPotionTestObject4F37D0, message string, value uint8) {
			w.event("message:%s:%s:%02x", pickupPotionObjectName4F37D0(owner), message, value)
			if w.afterClassFailure != nil {
				w.afterClassFailure()
			}
		},
		loadOwnerNetCode: func(owner *pickupPotionTestObject4F37D0) uint32 {
			w.event("netcode:%s=%08x", pickupPotionObjectName4F37D0(owner), owner.netCode)
			return owner.netCode
		},
		audio: func(sound uint32, owner *pickupPotionTestObject4F37D0, kind int32, code uint32) {
			w.event("audio:%d:%s:%d:%08x", sound, pickupPotionObjectName4F37D0(owner), kind, code)
		},
		loadPlayerState: func(owner *pickupPotionTestObject4F37D0) int32 {
			w.event("state:%s=%08x", pickupPotionObjectName4F37D0(owner), uint32(w.playerState))
			return w.playerState
		},
		loadPotionSubClassLow: func(potion *pickupPotionTestObject4F37D0) uint8 {
			w.event("subclass:%s=%02x", pickupPotionObjectName4F37D0(potion), potion.subClass)
			return potion.subClass
		},
		loadOwnerHealth: func(owner *pickupPotionTestObject4F37D0) *pickupPotionTestHealth4F37D0 {
			health := owner.health
			w.event("health:%s=%s", pickupPotionObjectName4F37D0(owner), pickupPotionHealthName4F37D0(health))
			return health
		},
		loadHealthCur: func(health *pickupPotionTestHealth4F37D0) uint16 {
			w.event("health-cur:%s=%04x", pickupPotionHealthName4F37D0(health), health.cur)
			return health.cur
		},
		loadHealthMax: func(health *pickupPotionTestHealth4F37D0) uint16 {
			w.event("health-max:%s=%04x", pickupPotionHealthName4F37D0(health), health.max)
			return health.max
		},
		scaleHealth: func(base int32, class uint8) int32 {
			value := w.healthScale[class]
			w.event("scale-health:%08x:%02x=%08x", uint32(base), class, uint32(value))
			if w.afterHealthScale != nil {
				w.afterHealthScale()
			}
			return value
		},
		adjustHealth: func(owner *pickupPotionTestObject4F37D0, amount int32) {
			w.event("adjust-health:%s:%08x", pickupPotionObjectName4F37D0(owner), uint32(amount))
			if w.afterAdjustHealth != nil {
				w.afterAdjustHealth()
			}
		},
		scaleMana: func(base int32, class uint8) int32 {
			value := w.manaScale[class]
			w.event("scale-mana:%08x:%02x=%08x", uint32(base), class, uint32(value))
			if w.afterManaScale != nil {
				w.afterManaScale()
			}
			return value
		},
		loadManaCur: func(update *pickupPotionTestUpdate4F37D0) uint16 {
			w.event("mana-cur:%s=%04x", pickupPotionUpdateName4F37D0(update), update.manaCur)
			return update.manaCur
		},
		loadManaMax: func(update *pickupPotionTestUpdate4F37D0) uint16 {
			w.event("mana-max:%s=%04x", pickupPotionUpdateName4F37D0(update), update.manaMax)
			return update.manaMax
		},
		addMana: func(owner *pickupPotionTestObject4F37D0, amount int32) {
			w.event("add-mana:%s:%08x", pickupPotionObjectName4F37D0(owner), uint32(amount))
			if w.afterAddMana != nil {
				w.afterAddMana()
			}
		},
		loadOwnerPoison: func(owner *pickupPotionTestObject4F37D0) uint8 {
			w.event("poison:%s=%02x", pickupPotionObjectName4F37D0(owner), owner.poison)
			return owner.poison
		},
		removePoison: func(owner *pickupPotionTestObject4F37D0) {
			w.event("remove-poison:%s", pickupPotionObjectName4F37D0(owner))
			if w.afterRemovePoison != nil {
				w.afterRemovePoison()
			}
		},
		spellAudio: func(spell, field int32) uint32 {
			w.event("spell-audio:%d:%d=%d", spell, field, w.spellSound)
			return w.spellSound
		},
		delayedDelete: func(potion *pickupPotionTestObject4F37D0) {
			w.event("delete:%s", pickupPotionObjectName4F37D0(potion))
		},
		decay: func(potion *pickupPotionTestObject4F37D0) {
			w.event("decay:%s", pickupPotionObjectName4F37D0(potion))
			if w.afterDecay != nil {
				w.afterDecay()
			}
		},
		loadArg4: func() int32 {
			value := w.arg4
			w.event("arg4:%08x", uint32(value))
			if w.afterLoadArg4 != nil {
				w.afterLoadArg4()
			}
			return value
		},
		loadArg3: func() int32 {
			w.event("arg3:%08x", uint32(w.arg3))
			return w.arg3
		},
		defaultPickup: func(owner, potion *pickupPotionTestObject4F37D0, arg3, arg4 int32) int32 {
			w.event("default:%s:%s:%08x:%08x=%08x", pickupPotionObjectName4F37D0(owner), pickupPotionObjectName4F37D0(potion), uint32(arg3), uint32(arg4), uint32(w.defaultResult))
			return w.defaultResult
		},
	}
}

func pickupPotionFullTrace4F37D0() (*pickupPotionTestWorld4F37D0, *pickupPotionTestObject4F37D0, *pickupPotionTestObject4F37D0, []string) {
	w := newPickupPotionTestWorld4F37D0()
	w.flags[pickupPotionClassRestrictionFlag4F37D0] = 1
	player := &pickupPotionTestPlayer4F37D0{name: "player-1", class: pickupPotionWarriorClass4F37D0}
	update := &pickupPotionTestUpdate4F37D0{name: "update-1", player: player, manaCur: 5, manaMax: 100}
	owner := &pickupPotionTestObject4F37D0{
		name: "owner", classLow: pickupPotionPlayerClassLow4F37D0,
		poison: 1, health: &pickupPotionTestHealth4F37D0{name: "health-1", cur: 20, max: 100}, update: update,
	}
	potion := &pickupPotionTestObject4F37D0{
		name: "potion", subClass: 0x70,
		use: &pickupPotionTestUse4F37D0{name: "use-1", value: 10},
	}
	want := []string{
		"use-data:potion",
		"use-value:use-1",
		"game:00002000=00000001",
		"game:00001000=00000000",
		"class:owner",
		"update:owner=update-1",
		"player:update-1=player-1",
		"player-class:player-1=00",
		"can-use:potion:00=00000001",
		"state:owner=00000000",
		"subclass:potion=70",
		"health:owner=health-1",
		"class:owner",
		"update:owner=update-1",
		"player:update-1=player-1",
		"player-class:player-1=00",
		"scale-health:0000000a:00=0000000b",
		"health:owner=health-1",
		"health-cur:health-1=0014",
		"health-max:health-1=0064",
		"adjust-health:owner:0000000b",
		"audio:754:owner:0:00000000",
		"subclass:potion=70",
		"class:owner",
		"update:owner=update-1",
		"player:update-1=player-1",
		"player-class:player-1=00",
		"scale-mana:0000000a:00=00000015",
		"mana-cur:update-1=0005",
		"mana-max:update-1=0064",
		"add-mana:owner:00000015",
		"audio:755:owner:0:00000000",
		"subclass:potion=70",
		"class:owner",
		"poison:owner=01",
		"remove-poison:owner",
		"spell-audio:14:1=1110",
		"audio:1110:owner:0:00000000",
		"delete:potion",
	}
	return w, owner, potion, want
}

func TestPickupPotion4F37D0FullTraceAndEveryFaultPrefix(t *testing.T) {
	w, owner, potion, want := pickupPotionFullTrace4F37D0()
	if got := pickupPotion4F37D0(owner, potion, w.hooks()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %v, want %v", w.events, want)
	}

	for faultAt := 1; faultAt <= len(want); faultAt++ {
		t.Run(fmt.Sprintf("fault-%d", faultAt), func(t *testing.T) {
			w, owner, potion, _ := pickupPotionFullTrace4F37D0()
			w.faultAt = faultAt
			defer func() {
				if got := recover(); got != want[faultAt-1] {
					t.Fatalf("panic = %v, want %q", got, want[faultAt-1])
				}
				if !reflect.DeepEqual(w.events, want[:faultAt]) {
					t.Fatalf("events = %v, want %v", w.events, want[:faultAt])
				}
			}()
			pickupPotion4F37D0(owner, potion, w.hooks())
		})
	}
}

func TestPickupPotion4F37D0EagerUseDataFaultsBeforeFlagsAndOwner(t *testing.T) {
	owner := &pickupPotionTestObject4F37D0{name: "owner"}
	for _, tc := range []struct {
		name   string
		potion *pickupPotionTestObject4F37D0
		want   []string
	}{
		{name: "nil-potion", want: []string{"use-data:nil"}},
		{name: "nil-use-data", potion: &pickupPotionTestObject4F37D0{name: "potion"}, want: []string{"use-data:potion", "use-value:nil"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := newPickupPotionTestWorld4F37D0()
			defer func() {
				if recover() == nil {
					t.Fatal("expected eager nil dereference")
				}
				if !reflect.DeepEqual(w.events, tc.want) {
					t.Fatalf("events = %v, want %v", w.events, tc.want)
				}
			}()
			pickupPotion4F37D0(owner, tc.potion, w.hooks())
		})
	}
}

func TestPickupPotion4F37D0ClassFailureUsesLiveNetCodeAndCanonicalZero(t *testing.T) {
	w := newPickupPotionTestWorld4F37D0()
	w.flags[pickupPotionClassRestrictionFlag4F37D0] = math.MinInt32
	w.canUse = 0
	player := &pickupPotionTestPlayer4F37D0{name: "player", class: pickupPotionWizardClass4F37D0}
	owner := &pickupPotionTestObject4F37D0{
		name: "owner", classLow: pickupPotionPlayerClassLow4F37D0, netCode: 7,
		update: &pickupPotionTestUpdate4F37D0{name: "update", player: player},
	}
	potion := &pickupPotionTestObject4F37D0{name: "potion", use: &pickupPotionTestUse4F37D0{name: "use", value: 3}}
	w.afterClassFailure = func() { owner.netCode = 0x89abcdef }
	if got := pickupPotion4F37D0(owner, potion, w.hooks()); got != 0 {
		t.Fatalf("result = %d, want canonical zero", got)
	}
	wantSuffix := []string{
		"can-use:potion:01=00000000",
		"message:owner:pickup.c:ObjectEquipClassFail:00",
		"netcode:owner=89abcdef",
		"audio:925:owner:2:89abcdef",
	}
	if !reflect.DeepEqual(w.events[len(w.events)-len(wantSuffix):], wantSuffix) {
		t.Fatalf("suffix = %v, want %v", w.events[len(w.events)-len(wantSuffix):], wantSuffix)
	}
	for _, event := range w.events {
		if event[:min(len(event), len("state:"))] == "state:" {
			t.Fatalf("class rejection reached player-state path: %v", w.events)
		}
	}
}

func TestPickupPotion4F37D0RestrictionShortCircuitsInOriginalOrder(t *testing.T) {
	for _, tc := range []struct {
		name       string
		classFlag  int32
		questFlag  int32
		ownerClass uint8
		prefix     []string
	}{
		{name: "class-flag-off", prefix: []string{"game:00002000=00000000", "state:owner=00000001"}},
		{name: "quest-flag-on", classFlag: 1, questFlag: -1, prefix: []string{"game:00002000=00000001", "game:00001000=ffffffff", "state:owner=00000001"}},
		{name: "non-player", classFlag: 1, ownerClass: 2, prefix: []string{"game:00002000=00000001", "game:00001000=00000000", "class:owner", "state:owner=00000001"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := newPickupPotionTestWorld4F37D0()
			w.flags[pickupPotionClassRestrictionFlag4F37D0] = tc.classFlag
			w.flags[pickupPotionQuestFlag4F37D0] = tc.questFlag
			w.playerState = 1
			owner := &pickupPotionTestObject4F37D0{name: "owner", classLow: tc.ownerClass}
			potion := &pickupPotionTestObject4F37D0{name: "potion", use: &pickupPotionTestUse4F37D0{name: "use"}}
			_ = pickupPotion4F37D0(owner, potion, w.hooks())
			got := w.events[2 : 2+len(tc.prefix)]
			if !reflect.DeepEqual(got, tc.prefix) {
				t.Fatalf("restriction prefix = %v, want %v; all = %v", got, tc.prefix, w.events)
			}
		})
	}
}

func TestPickupPotion4F37D0BlockedPathDecayArgsExactResultAndAudio(t *testing.T) {
	for _, result := range []int32{0, 1, 2, -1, math.MinInt32, math.MaxInt32} {
		t.Run(fmt.Sprintf("%08x", uint32(result)), func(t *testing.T) {
			w := newPickupPotionTestWorld4F37D0()
			w.playerState = math.MinInt32
			w.defaultResult = result
			owner := &pickupPotionTestObject4F37D0{name: "owner"}
			potion := &pickupPotionTestObject4F37D0{name: "potion", use: &pickupPotionTestUse4F37D0{name: "use", value: 99}}
			w.afterLoadArg4 = func() {
				w.arg3 = 0x12345678
				w.arg4 = -2
			}
			if got := pickupPotion4F37D0(owner, potion, w.hooks()); got != result {
				t.Fatalf("result = %d, want %d", got, result)
			}
			wantTail := []string{
				"decay:potion",
				"arg4:7fffffff",
				"arg3:12345678",
				fmt.Sprintf("default:owner:potion:12345678:7fffffff=%08x", uint32(result)),
			}
			if result == 1 {
				wantTail = append(wantTail, "audio:832:owner:0:00000000")
			}
			gotTail := w.events[len(w.events)-len(wantTail):]
			if !reflect.DeepEqual(gotTail, wantTail) {
				t.Fatalf("tail = %v, want %v", gotTail, wantTail)
			}
		})
	}
}

func TestPickupPotion4F37D0HealthReloadStrictBoundaryAndSignedWrap(t *testing.T) {
	for _, tc := range []struct {
		name       string
		value      int32
		cur, max   uint16
		wantAdjust bool
	}{
		{name: "equal-does-not-consume", value: 10, cur: 90, max: 100},
		{name: "below-consumes", value: 9, cur: 90, max: 100, wantAdjust: true},
		{name: "signed-wrap-consumes", value: math.MaxInt32, cur: 1, max: 1, wantAdjust: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := newPickupPotionTestWorld4F37D0()
			w.defaultResult = -7
			owner := &pickupPotionTestObject4F37D0{
				name: "owner", classLow: 2,
				health: &pickupPotionTestHealth4F37D0{name: "health", cur: tc.cur, max: tc.max},
			}
			potion := &pickupPotionTestObject4F37D0{
				name: "potion", subClass: pickupPotionHealthSubClassLow4F37D0,
				use: &pickupPotionTestUse4F37D0{name: "use", value: tc.value},
			}
			got := pickupPotion4F37D0(owner, potion, w.hooks())
			if tc.wantAdjust && got != 1 {
				t.Fatalf("result = %d, want consumed success", got)
			}
			if !tc.wantAdjust && got != -7 {
				t.Fatalf("result = %d, want default -7", got)
			}
			adjusted := false
			for _, event := range w.events {
				if len(event) >= len("adjust-health:") && event[:len("adjust-health:")] == "adjust-health:" {
					adjusted = true
				}
			}
			if adjusted != tc.wantAdjust {
				t.Fatalf("adjusted = %v, want %v; events = %v", adjusted, tc.wantAdjust, w.events)
			}
		})
	}

	w := newPickupPotionTestWorld4F37D0()
	player := &pickupPotionTestPlayer4F37D0{name: "player", class: 0}
	health1 := &pickupPotionTestHealth4F37D0{name: "health-1", cur: 99, max: 100}
	health2 := &pickupPotionTestHealth4F37D0{name: "health-2", cur: 1, max: 100}
	owner := &pickupPotionTestObject4F37D0{
		name: "owner", classLow: pickupPotionPlayerClassLow4F37D0,
		health: health1, update: &pickupPotionTestUpdate4F37D0{name: "update", player: player},
	}
	potion := &pickupPotionTestObject4F37D0{
		name: "potion", subClass: pickupPotionHealthSubClassLow4F37D0,
		use: &pickupPotionTestUse4F37D0{name: "use", value: 10},
	}
	w.afterHealthScale = func() { owner.health = health2 }
	if got := pickupPotion4F37D0(owner, potion, w.hooks()); got != 1 {
		t.Fatalf("reloaded health result = %d, want 1", got)
	}
	want := []string{
		"health:owner=health-1",
		"class:owner",
		"update:owner=update",
		"player:update=player",
		"player-class:player=00",
		"scale-health:0000000a:00=0000000b",
		"health:owner=health-2",
		"health-cur:health-2=0001",
		"health-max:health-2=0064",
	}
	for i := 0; i+len(want) <= len(w.events); i++ {
		if reflect.DeepEqual(w.events[i:i+len(want)], want) {
			return
		}
	}
	t.Fatalf("health reload sequence not found: %v", w.events)
}

func TestPickupPotion4F37D0ManaCachesUpdateAndInvalidClassRetainsHealthAmount(t *testing.T) {
	w := newPickupPotionTestWorld4F37D0()
	player := &pickupPotionTestPlayer4F37D0{name: "player", class: 0}
	update1 := &pickupPotionTestUpdate4F37D0{name: "update-1", player: player, manaCur: 1, manaMax: 100}
	update2 := &pickupPotionTestUpdate4F37D0{name: "update-2", player: player, manaCur: 99, manaMax: 100}
	owner := &pickupPotionTestObject4F37D0{
		name: "owner", classLow: pickupPotionPlayerClassLow4F37D0,
		health: &pickupPotionTestHealth4F37D0{name: "health", cur: 1, max: 100}, update: update1,
	}
	potion := &pickupPotionTestObject4F37D0{
		name: "potion", subClass: pickupPotionHealthSubClassLow4F37D0 | pickupPotionManaSubClassLow4F37D0,
		use: &pickupPotionTestUse4F37D0{name: "use", value: 10},
	}
	w.healthScale[0] = 30
	w.afterAdjustHealth = func() { player.class = 7 }
	w.afterManaScale = func() { t.Fatal("invalid class must not call mana scaling") }
	loadUpdateCount := 0
	hooks := w.hooks()
	originalLoadUpdate := hooks.loadOwnerUpdate
	hooks.loadOwnerUpdate = func(owner *pickupPotionTestObject4F37D0) *pickupPotionTestUpdate4F37D0 {
		update := originalLoadUpdate(owner)
		loadUpdateCount++
		if loadUpdateCount == 2 {
			owner.update = update2
		}
		return update
	}
	if got := pickupPotion4F37D0(owner, potion, hooks); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	wantTail := []string{
		"player:update-1=player",
		"player-class:player=07",
		"mana-cur:update-1=0001",
		"mana-max:update-1=0064",
		"add-mana:owner:0000001e",
	}
	found := false
	for i := 0; i+len(wantTail) <= len(w.events); i++ {
		if reflect.DeepEqual(w.events[i:i+len(wantTail)], wantTail) {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("cached-update/retained-amount sequence missing: %v", w.events)
	}
}

func TestPickupPotion4F37D0CureDeletesImmediatelyAfterLivePoison(t *testing.T) {
	w := newPickupPotionTestWorld4F37D0()
	owner := &pickupPotionTestObject4F37D0{
		name: "owner", classLow: pickupPotionPlayerClassLow4F37D0, poison: 2,
	}
	potion := &pickupPotionTestObject4F37D0{
		name: "potion", subClass: pickupPotionCureSubClassLow4F37D0,
		use: &pickupPotionTestUse4F37D0{name: "use", value: 1},
	}
	w.afterRemovePoison = func() { owner.poison = 0; potion.subClass = 0 }
	if got := pickupPotion4F37D0(owner, potion, w.hooks()); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	wantTail := []string{
		"poison:owner=02",
		"remove-poison:owner",
		"spell-audio:14:1=1110",
		"audio:1110:owner:0:00000000",
		"delete:potion",
	}
	if !reflect.DeepEqual(w.events[len(w.events)-len(wantTail):], wantTail) {
		t.Fatalf("tail = %v, want %v", w.events[len(w.events)-len(wantTail):], wantTail)
	}
}
