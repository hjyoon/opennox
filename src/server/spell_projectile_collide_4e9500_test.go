package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type spellProjectileTestObject4E9500 struct {
	name         string
	flags        uint32
	classLow     uint8
	direction    int16
	update       *spellProjectileTestUpdate4E9500
	playerUpdate *spellProjectileTestPlayerUpdate4E9500
}

type spellProjectileTestUpdate4E9500 struct {
	target *spellProjectileTestObject4E9500
	owner  *spellProjectileTestObject4E9500
	source *spellProjectileTestObject4E9500
	spell  int32
	level  int32
}

type spellProjectileTestPlayerUpdate4E9500 struct {
	state  uint8
	player *spellProjectileTestPlayer4E9500
	frame  uint8
}

type spellProjectileTestPlayer4E9500 struct {
	weapon uint32
}

type spellProjectileTestState4E9500 struct {
	events             []string
	previousResult     int32
	currentResult      int32
	inversionResult    int32
	enchantResult      int32
	randomResult       int32
	actionResult       int32
	animStart          int32
	animEnd            int32
	acceptResult       int32
	onReflect          func()
	acceptedSpell      int32
	acceptedSource     *spellProjectileTestObject4E9500
	acceptedOwner      *spellProjectileTestObject4E9500
	acceptedProjectile *spellProjectileTestObject4E9500
	acceptedTarget     *spellProjectileTestObject4E9500
	acceptedLevel      int32
}

func (s *spellProjectileTestState4E9500) hooks() spellProjectileCollideHooks4E9500[
	*spellProjectileTestObject4E9500,
	*spellProjectileTestUpdate4E9500,
	*spellProjectileTestPlayerUpdate4E9500,
	*spellProjectileTestPlayer4E9500,
	*int,
] {
	return spellProjectileCollideHooks4E9500[
		*spellProjectileTestObject4E9500,
		*spellProjectileTestUpdate4E9500,
		*spellProjectileTestPlayerUpdate4E9500,
		*spellProjectileTestPlayer4E9500,
		*int,
	]{
		loadProjectileUpdate: func(obj *spellProjectileTestObject4E9500) *spellProjectileTestUpdate4E9500 {
			s.events = append(s.events, "projectile-update")
			return obj.update
		},
		wallReflect: func(*int, *spellProjectileTestObject4E9500) {
			s.events = append(s.events, "wall-reflect")
		},
		loadFlags: func(obj *spellProjectileTestObject4E9500) uint32 {
			s.events = append(s.events, "flags")
			return obj.flags
		},
		loadClassLow: func(obj *spellProjectileTestObject4E9500) uint8 {
			s.events = append(s.events, "class")
			return obj.classLow
		},
		loadPlayerUpdate: func(obj *spellProjectileTestObject4E9500) *spellProjectileTestPlayerUpdate4E9500 {
			s.events = append(s.events, "player-update")
			return obj.playerUpdate
		},
		loadPlayerState: func(update *spellProjectileTestPlayerUpdate4E9500) uint8 {
			s.events = append(s.events, "state")
			return update.state
		},
		loadDirection: func(obj *spellProjectileTestObject4E9500) int16 {
			s.events = append(s.events, "direction")
			return obj.direction
		},
		checkPrevious: func(target *spellProjectileTestObject4E9500, direction int32, projectile *spellProjectileTestObject4E9500) int32 {
			s.events = append(s.events, fmt.Sprintf("previous:%d", direction))
			return s.previousResult
		},
		audio: func(id uint32, obj *spellProjectileTestObject4E9500) {
			s.events = append(s.events, fmt.Sprintf("audio:%d", id))
		},
		projectileReflect: func(projectile, other *spellProjectileTestObject4E9500) {
			s.events = append(s.events, "projectile-reflect")
			if s.onReflect != nil {
				s.onReflect()
			}
		},
		changeOwner: func(projectile, other *spellProjectileTestObject4E9500) {
			s.events = append(s.events, "change-owner")
		},
		loadPlayer: func(update *spellProjectileTestPlayerUpdate4E9500) *spellProjectileTestPlayer4E9500 {
			s.events = append(s.events, "player")
			return update.player
		},
		loadWeaponEquip: func(player *spellProjectileTestPlayer4E9500) uint32 {
			s.events = append(s.events, "weapon")
			return player.weapon
		},
		randomInt: func(minimum, maximum int32) int32 {
			s.events = append(s.events, fmt.Sprintf("random:%d:%d", minimum, maximum))
			return s.randomResult
		},
		setPlayerState: func(obj *spellProjectileTestObject4E9500, state int32) {
			s.events = append(s.events, fmt.Sprintf("set-state:%d", state))
			obj.playerUpdate.state = uint8(state)
		},
		mapPlayerAction: func(obj *spellProjectileTestObject4E9500) int32 {
			s.events = append(s.events, "map-action")
			return s.actionResult
		},
		playerAnimFrames: func(action int32) (int32, int32) {
			s.events = append(s.events, fmt.Sprintf("anim:%d", action))
			return s.animStart, s.animEnd
		},
		storeAnimFrame: func(update *spellProjectileTestPlayerUpdate4E9500, frame uint8) {
			s.events = append(s.events, fmt.Sprintf("store-frame:%d", frame))
			update.frame = frame
		},
		checkInversion: func(target, projectile *spellProjectileTestObject4E9500) int32 {
			s.events = append(s.events, "inversion")
			return s.inversionResult
		},
		hasEnchant: func(obj *spellProjectileTestObject4E9500, enchant uint32) int32 {
			s.events = append(s.events, fmt.Sprintf("enchant:%d", enchant))
			return s.enchantResult
		},
		checkCurrent: func(target *spellProjectileTestObject4E9500, direction int32, projectile *spellProjectileTestObject4E9500) int32 {
			s.events = append(s.events, fmt.Sprintf("current:%d", direction))
			return s.currentResult
		},
		loadTarget: func(update *spellProjectileTestUpdate4E9500) *spellProjectileTestObject4E9500 {
			s.events = append(s.events, "target")
			return update.target
		},
		loadLevel: func(update *spellProjectileTestUpdate4E9500) int32 {
			s.events = append(s.events, "level")
			return update.level
		},
		loadOwner: func(update *spellProjectileTestUpdate4E9500) *spellProjectileTestObject4E9500 {
			s.events = append(s.events, "owner")
			return update.owner
		},
		loadSource: func(update *spellProjectileTestUpdate4E9500) *spellProjectileTestObject4E9500 {
			s.events = append(s.events, "source")
			return update.source
		},
		loadSpell: func(update *spellProjectileTestUpdate4E9500) int32 {
			s.events = append(s.events, "spell")
			return update.spell
		},
		spellAccept: func(spellID int32, source, owner, projectile, target *spellProjectileTestObject4E9500, level int32) int32 {
			s.events = append(s.events, "accept")
			s.acceptedSpell = spellID
			s.acceptedSource = source
			s.acceptedOwner = owner
			s.acceptedProjectile = projectile
			s.acceptedTarget = target
			s.acceptedLevel = level
			return s.acceptResult
		},
		delayedDelete: func(obj *spellProjectileTestObject4E9500) {
			s.events = append(s.events, "delete")
		},
	}
}

func TestSpellProjectileCollide4E9500CachesUpdateBeforeNilOther(t *testing.T) {
	projectile := &spellProjectileTestObject4E9500{update: &spellProjectileTestUpdate4E9500{}}
	state := &spellProjectileTestState4E9500{}
	spellProjectileCollide4E9500(projectile, nil, (*int)(nil), state.hooks())
	if want := []string{"projectile-update"}; !reflect.DeepEqual(state.events, want) {
		t.Fatalf("nil collision events = %v, want %v", state.events, want)
	}

	state = &spellProjectileTestState4E9500{}
	collision := 7
	spellProjectileCollide4E9500(projectile, nil, &collision, state.hooks())
	if want := []string{"projectile-update", "wall-reflect"}; !reflect.DeepEqual(state.events, want) {
		t.Fatalf("wall events = %v, want %v", state.events, want)
	}
}

func TestSpellProjectileCollide4E9500EntryGatesAndCachedUpdate(t *testing.T) {
	other := &spellProjectileTestObject4E9500{}
	oldUpdate := &spellProjectileTestUpdate4E9500{target: other, spell: 71, level: -9}
	projectile := &spellProjectileTestObject4E9500{update: oldUpdate}

	state := &spellProjectileTestState4E9500{}
	other.flags = 0x8000
	spellProjectileCollide4E9500(projectile, other, (*int)(nil), state.hooks())
	if want := []string{"projectile-update", "flags"}; !reflect.DeepEqual(state.events, want) {
		t.Fatalf("flag-gate events = %v, want %v", state.events, want)
	}

	other.flags = 0
	state = &spellProjectileTestState4E9500{}
	state.onReflect = func() {
		projectile.update = &spellProjectileTestUpdate4E9500{}
	}
	oldUpdate.target = nil
	spellProjectileCollide4E9500(projectile, other, (*int)(nil), state.hooks())
	if want := []string{"projectile-update", "flags", "class", "target", "projectile-reflect"}; !reflect.DeepEqual(state.events, want) {
		t.Fatalf("mismatch events = %v, want %v", state.events, want)
	}
}

func TestSpellProjectileCollide4E9500State16ReflectOrder(t *testing.T) {
	projectile := &spellProjectileTestObject4E9500{update: &spellProjectileTestUpdate4E9500{}}
	other := &spellProjectileTestObject4E9500{
		classLow:     4,
		direction:    -2,
		playerUpdate: &spellProjectileTestPlayerUpdate4E9500{state: 16},
	}
	state := &spellProjectileTestState4E9500{previousResult: 9}
	spellProjectileCollide4E9500(projectile, other, (*int)(nil), state.hooks())
	want := []string{
		"projectile-update", "flags", "class", "player-update", "state",
		"direction", "previous:-2", "audio:878", "projectile-reflect", "change-owner",
	}
	if !reflect.DeepEqual(state.events, want) {
		t.Fatalf("events = %v, want %v", state.events, want)
	}
}

func TestSpellProjectileCollide4E9500GreatSwordFullOrder(t *testing.T) {
	owner := &spellProjectileTestObject4E9500{name: "owner"}
	source := &spellProjectileTestObject4E9500{name: "source"}
	other := &spellProjectileTestObject4E9500{
		name:      "other",
		classLow:  4,
		direction: 0x1234,
		playerUpdate: &spellProjectileTestPlayerUpdate4E9500{
			state:  13,
			player: &spellProjectileTestPlayer4E9500{weapon: 0x400},
		},
	}
	oldUpdate := &spellProjectileTestUpdate4E9500{
		target: other,
		owner:  owner,
		source: source,
		spell:  -17,
		level:  math.MinInt32,
	}
	projectile := &spellProjectileTestObject4E9500{name: "projectile", update: oldUpdate}
	state := &spellProjectileTestState4E9500{
		previousResult: 1,
		randomResult:   20,
		actionResult:   47,
		animStart:      0x101,
		animEnd:        0x202,
		acceptResult:   99,
	}
	state.onReflect = func() {
		projectile.update = &spellProjectileTestUpdate4E9500{}
	}
	spellProjectileCollide4E9500(projectile, other, (*int)(nil), state.hooks())
	want := []string{
		"projectile-update", "flags", "class", "player-update", "state", "state",
		"player", "weapon", "direction", "previous:4660", "random:18:20",
		"audio:890", "projectile-reflect", "change-owner", "set-state:20",
		"map-action", "anim:47", "store-frame:0", "inversion", "enchant:27",
		"target", "level", "owner", "source", "spell", "accept", "delete",
	}
	if !reflect.DeepEqual(state.events, want) {
		t.Fatalf("events = %v, want %v", state.events, want)
	}
	if other.playerUpdate.frame != 0 {
		t.Fatalf("stored frame = %d, want low byte of 0x100", other.playerUpdate.frame)
	}
	if state.acceptedSpell != oldUpdate.spell || state.acceptedSource != source || state.acceptedOwner != owner ||
		state.acceptedProjectile != projectile || state.acceptedTarget != other || state.acceptedLevel != oldUpdate.level {
		t.Fatalf("accept args = (%d,%p,%p,%p,%p,%d)", state.acceptedSpell, state.acceptedSource,
			state.acceptedOwner, state.acceptedProjectile, state.acceptedTarget, state.acceptedLevel)
	}
}

func TestSpellProjectileCollide4E9500InversionAndEnchantReturns(t *testing.T) {
	projectile := &spellProjectileTestObject4E9500{update: &spellProjectileTestUpdate4E9500{}}
	other := &spellProjectileTestObject4E9500{
		classLow:     4,
		direction:    -32768,
		playerUpdate: &spellProjectileTestPlayerUpdate4E9500{},
	}

	state := &spellProjectileTestState4E9500{inversionResult: -1}
	spellProjectileCollide4E9500(projectile, other, (*int)(nil), state.hooks())
	want := []string{
		"projectile-update", "flags", "class", "player-update", "state", "state",
		"inversion", "change-owner",
	}
	if !reflect.DeepEqual(state.events, want) {
		t.Fatalf("inversion events = %v, want %v", state.events, want)
	}

	state = &spellProjectileTestState4E9500{enchantResult: 1, currentResult: 3}
	spellProjectileCollide4E9500(projectile, other, (*int)(nil), state.hooks())
	want = []string{
		"projectile-update", "flags", "class", "player-update", "state", "state",
		"inversion", "enchant:27", "direction", "current:-32768", "change-owner", "audio:122",
	}
	if !reflect.DeepEqual(state.events, want) {
		t.Fatalf("enchant events = %v, want %v", state.events, want)
	}
}

func TestSpellProjectileCollide4E9500NilProjectileFaultsAtCachedLoad(t *testing.T) {
	state := &spellProjectileTestState4E9500{}
	defer func() {
		if recover() == nil {
			t.Fatal("nil projectile did not fault")
		}
		if want := []string{"projectile-update"}; !reflect.DeepEqual(state.events, want) {
			t.Fatalf("events = %v, want %v", state.events, want)
		}
	}()
	spellProjectileCollide4E9500(
		(*spellProjectileTestObject4E9500)(nil),
		(*spellProjectileTestObject4E9500)(nil),
		(*int)(nil),
		state.hooks(),
	)
}

type spellProjectileInversionTestItem4FA4F0 struct {
	name  string
	flags uint32
	class uint32
	data  *spellProjectileInversionTestData4FA4F0
	next  *spellProjectileInversionTestItem4FA4F0
}

type spellProjectileInversionTestData4FA4F0 struct {
	modifiers [4]*spellProjectileInversionTestModifier4FA4F0
}

type spellProjectileInversionTestModifier4FA4F0 struct {
	name     string
	function int
	strength int32
}

func TestSpellProjectileInversion4FA4F0ExactFiltersSlotsAndOrder(t *testing.T) {
	wrongFunction := &spellProjectileInversionTestModifier4FA4F0{name: "wrong", function: 88, strength: 9}
	weak := &spellProjectileInversionTestModifier4FA4F0{name: "weak", function: 77, strength: 0}
	strong := &spellProjectileInversionTestModifier4FA4F0{name: "strong", function: 77, strength: 1}
	ignoredSlot := &spellProjectileInversionTestModifier4FA4F0{name: "ignored", function: 77, strength: 9}
	fourth := &spellProjectileInversionTestItem4FA4F0{
		name: "fourth", flags: 0x100, class: 0x1000,
		data: &spellProjectileInversionTestData4FA4F0{modifiers: [4]*spellProjectileInversionTestModifier4FA4F0{nil, nil, strong}},
	}
	third := &spellProjectileInversionTestItem4FA4F0{
		name: "third", flags: 0x100, class: 0x10000000,
		data: &spellProjectileInversionTestData4FA4F0{modifiers: [4]*spellProjectileInversionTestModifier4FA4F0{ignoredSlot, nil, weak, wrongFunction}},
		next: fourth,
	}
	second := &spellProjectileInversionTestItem4FA4F0{name: "second", flags: 0x100, class: 0x800, next: third}
	first := &spellProjectileInversionTestItem4FA4F0{name: "first", next: second}
	target := &spellProjectileTestObject4E9500{name: "target"}
	projectile := &spellProjectileTestObject4E9500{name: "projectile"}
	events := make([]string, 0, 32)
	got := spellProjectileInversion4FA4F0(target, projectile, spellProjectileInversionHooks4FA4F0[
		*spellProjectileTestObject4E9500,
		*spellProjectileInversionTestItem4FA4F0,
		*spellProjectileInversionTestData4FA4F0,
		*spellProjectileInversionTestModifier4FA4F0,
		int,
	]{
		firstItem: func(*spellProjectileTestObject4E9500) *spellProjectileInversionTestItem4FA4F0 {
			events = append(events, "first")
			return first
		},
		loadFlags: func(item *spellProjectileInversionTestItem4FA4F0) uint32 {
			events = append(events, "flags:"+item.name)
			return item.flags
		},
		loadClass: func(item *spellProjectileInversionTestItem4FA4F0) uint32 {
			events = append(events, "class:"+item.name)
			return item.class
		},
		loadInitData: func(item *spellProjectileInversionTestItem4FA4F0) *spellProjectileInversionTestData4FA4F0 {
			events = append(events, "data:"+item.name)
			return item.data
		},
		loadModifier: func(data *spellProjectileInversionTestData4FA4F0, slot int) *spellProjectileInversionTestModifier4FA4F0 {
			events = append(events, fmt.Sprintf("slot:%d", slot))
			return data.modifiers[slot]
		},
		loadDefendCollide: func(modifier *spellProjectileInversionTestModifier4FA4F0) int {
			events = append(events, "function:"+modifier.name)
			return modifier.function
		},
		inversionEffect: 77,
		findParent: func(obj *spellProjectileTestObject4E9500) *spellProjectileTestObject4E9500 {
			events = append(events, "parent:"+obj.name)
			return obj
		},
		loadInversionStrength: func(modifier *spellProjectileInversionTestModifier4FA4F0) int32 {
			events = append(events, "strength:"+modifier.name)
			return modifier.strength
		},
		nextItem: func(item *spellProjectileInversionTestItem4FA4F0) *spellProjectileInversionTestItem4FA4F0 {
			events = append(events, "next:"+item.name)
			return item.next
		},
	})
	if got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	want := []string{
		"first", "flags:first", "next:first",
		"flags:second", "class:second", "next:second",
		"flags:third", "class:third", "data:third", "slot:2", "function:weak",
		"parent:projectile", "strength:weak", "slot:3", "function:wrong", "next:third",
		"flags:fourth", "class:fourth", "data:fourth", "slot:2", "function:strong",
		"parent:projectile", "strength:strong",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}
