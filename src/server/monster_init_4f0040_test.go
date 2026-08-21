package server

import (
	"fmt"
	"math"
	"reflect"
	"strings"
	"testing"
)

type monsterInitTestObject4F0040 struct {
	name      string
	update    *monsterInitTestUpdate4F0040
	flags     uint32
	typeIndex uint16
	radius    float32
	positionX uint32
	positionY uint32
	direction int16
	health    *monsterInitTestHealth4F0040
	subclass  uint8
	speedBase float32
}

type monsterInitTestUpdate4F0040 struct {
	name        string
	monsterDef  *monsterInitTestDef4F0040
	aiAction    uint32
	sightRange  float32
	aggression  float32
	status      uint32
	direction   uint32
	positionX   uint32
	positionY   uint32
	healthGraph [32]uint16
	field332    float32
	field333    uint32
	healthScale float32
	fleeRange   float32
}

type monsterInitTestDef4F0040 struct {
	name       string
	meleeRange float32
}

type monsterInitTestHealth4F0040 struct {
	name              string
	current, previous uint16
	maximum           uint16
}

type monsterInitTestAction4F0040 struct {
	name   string
	action uint32
	args   [4]uint32
}

type monsterInitTestWorld4F0040 struct {
	events  []string
	faultAt int
	onEvent func(string)

	unit        *monsterInitTestObject4F0040
	plantTypeID uint32
	rat         bool
	fish        bool
	frog        bool
	canAttack   bool
	canCastFlag bool
	frame       uint32
	random      float64
	nilPush     map[uint32]bool
	pushed      []*monsterInitTestAction4F0040
}

func (w *monsterInitTestWorld4F0040) record(event string) {
	w.events = append(w.events, event)
	if w.onEvent != nil {
		w.onEvent(event)
	}
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic("injected fault")
	}
}

func (w *monsterInitTestWorld4F0040) hooks() monsterInitHooks4F0040[
	*monsterInitTestObject4F0040,
	*monsterInitTestUpdate4F0040,
	*monsterInitTestHealth4F0040,
	*monsterInitTestDef4F0040,
	*monsterInitTestAction4F0040,
] {
	return monsterInitHooks4F0040[
		*monsterInitTestObject4F0040,
		*monsterInitTestUpdate4F0040,
		*monsterInitTestHealth4F0040,
		*monsterInitTestDef4F0040,
		*monsterInitTestAction4F0040,
	]{
		loadUnitArg: func() *monsterInitTestObject4F0040 {
			w.record("arg:" + w.unit.name)
			return w.unit
		},
		loadUpdateData: func(unit *monsterInitTestObject4F0040) *monsterInitTestUpdate4F0040 {
			update := unit.update
			w.record("update:" + unit.name + "=" + update.name)
			return update
		},
		loadPlantTypeID: func() uint32 {
			w.record(fmt.Sprintf("plant-id=%d", w.plantTypeID))
			return w.plantTypeID
		},
		loadObjectFlags: func(unit *monsterInitTestObject4F0040) uint32 {
			w.record(fmt.Sprintf("flags:%s=%#08x", unit.name, unit.flags))
			return unit.flags
		},
		loadTypeIndex: func(unit *monsterInitTestObject4F0040) uint16 {
			w.record(fmt.Sprintf("type:%s=%d", unit.name, unit.typeIndex))
			return unit.typeIndex
		},
		isRat: func(unit *monsterInitTestObject4F0040) bool {
			w.record(fmt.Sprintf("rat:%s=%t", unit.name, w.rat))
			return w.rat
		},
		isFish: func(unit *monsterInitTestObject4F0040) bool {
			w.record(fmt.Sprintf("fish:%s=%t", unit.name, w.fish))
			return w.fish
		},
		isFrog: func(unit *monsterInitTestObject4F0040) bool {
			w.record(fmt.Sprintf("frog:%s=%t", unit.name, w.frog))
			return w.frog
		},
		clearActionStack: func(unit *monsterInitTestObject4F0040) {
			w.record("clear:" + unit.name)
		},
		loadMonsterDef: func(update *monsterInitTestUpdate4F0040) *monsterInitTestDef4F0040 {
			def := update.monsterDef
			w.record("monster-def:" + update.name + "=" + def.name)
			return def
		},
		loadMeleeRange: func(def *monsterInitTestDef4F0040) float32 {
			w.record(fmt.Sprintf("melee:%s=%#08x", def.name, math.Float32bits(def.meleeRange)))
			return def.meleeRange
		},
		loadCircleRadius: func(unit *monsterInitTestObject4F0040) float32 {
			w.record(fmt.Sprintf("radius:%s=%#08x", unit.name, math.Float32bits(unit.radius)))
			return unit.radius
		},
		loadAIAction: func(update *monsterInitTestUpdate4F0040) uint32 {
			w.record(fmt.Sprintf("ai-load:%s=%d", update.name, update.aiAction))
			return update.aiAction
		},
		storeAIAction: func(update *monsterInitTestUpdate4F0040, action uint32) {
			w.record(fmt.Sprintf("ai-store:%s=%d", update.name, action))
			update.aiAction = action
		},
		storeSightRange: func(update *monsterInitTestUpdate4F0040, value float32) {
			w.record(fmt.Sprintf("sight:%s=%#08x", update.name, math.Float32bits(value)))
			update.sightRange = value
		},
		storeAggression: func(update *monsterInitTestUpdate4F0040, value float32) {
			w.record(fmt.Sprintf("aggression:%s=%#08x", update.name, math.Float32bits(value)))
			update.aggression = value
		},
		loadStatus: func(update *monsterInitTestUpdate4F0040) uint32 {
			w.record(fmt.Sprintf("status-load:%s=%#08x", update.name, update.status))
			return update.status
		},
		storeStatus: func(update *monsterInitTestUpdate4F0040, status uint32) {
			w.record(fmt.Sprintf("status-store:%s=%#08x", update.name, status))
			update.status = status
		},
		pushAction: func(unit *monsterInitTestObject4F0040, action uint32) *monsterInitTestAction4F0040 {
			w.record(fmt.Sprintf("push:%s=%d", unit.name, action))
			if w.nilPush[action] {
				return nil
			}
			item := &monsterInitTestAction4F0040{
				name:   fmt.Sprintf("action-%d-%d", action, len(w.pushed)),
				action: action,
				args:   [4]uint32{0xa0a0a0a0, 0xb0b0b0b0, 0xc0ffee00, 0xd0d0d0d0},
			}
			w.pushed = append(w.pushed, item)
			return item
		},
		storeActionArg: func(item *monsterInitTestAction4F0040, index int, value uint32) {
			w.record(fmt.Sprintf("arg-store:%s:%d=%#08x", item.name, index, value))
			item.args[index] = value
		},
		storeActionArgLow: func(item *monsterInitTestAction4F0040, index int, value uint8) {
			w.record(fmt.Sprintf("arg-low:%s:%d=%#02x", item.name, index, value))
			item.args[index] = item.args[index]&^0xff | uint32(value)
		},
		canAttackAtWill: func(unit *monsterInitTestObject4F0040) bool {
			w.record(fmt.Sprintf("can-attack:%s=%t", unit.name, w.canAttack))
			return w.canAttack
		},
		loadPositionXBits: func(unit *monsterInitTestObject4F0040) uint32 {
			w.record(fmt.Sprintf("position-x:%s=%#08x", unit.name, unit.positionX))
			return unit.positionX
		},
		loadPositionYBits: func(unit *monsterInitTestObject4F0040) uint32 {
			w.record(fmt.Sprintf("position-y:%s=%#08x", unit.name, unit.positionY))
			return unit.positionY
		},
		loadFrame: func() uint32 {
			w.record(fmt.Sprintf("frame=%#08x", w.frame))
			return w.frame
		},
		loadDirection: func(unit *monsterInitTestObject4F0040) int16 {
			w.record(fmt.Sprintf("direction:%s=%d", unit.name, unit.direction))
			return unit.direction
		},
		storeDirection: func(update *monsterInitTestUpdate4F0040, value uint32) {
			w.record(fmt.Sprintf("direction-store:%s=%#08x", update.name, value))
			update.direction = value
		},
		storePositionX: func(update *monsterInitTestUpdate4F0040, value uint32) {
			w.record(fmt.Sprintf("position-x-store:%s=%#08x", update.name, value))
			update.positionX = value
		},
		storePositionY: func(update *monsterInitTestUpdate4F0040, value uint32) {
			w.record(fmt.Sprintf("position-y-store:%s=%#08x", update.name, value))
			update.positionY = value
		},
		loadHealth: func(unit *monsterInitTestObject4F0040) *monsterInitTestHealth4F0040 {
			health := unit.health
			w.record("health:" + unit.name + "=" + health.name)
			return health
		},
		loadHealthMaximum: func(health *monsterInitTestHealth4F0040) uint16 {
			w.record(fmt.Sprintf("health-max:%s=%d", health.name, health.maximum))
			return health.maximum
		},
		loadHealthCurrent: func(health *monsterInitTestHealth4F0040) uint16 {
			w.record(fmt.Sprintf("health-cur:%s=%d", health.name, health.current))
			return health.current
		},
		loadHealthScale: func(update *monsterInitTestUpdate4F0040) float32 {
			w.record(fmt.Sprintf("health-scale:%s=%#08x", update.name, math.Float32bits(update.healthScale)))
			return update.healthScale
		},
		setHealth: func(unit *monsterInitTestObject4F0040, value uint16) {
			w.record(fmt.Sprintf("health-set:%s=%d", unit.name, value))
			unit.health.current = value
		},
		storeHealthPrev: func(health *monsterInitTestHealth4F0040, value uint16) {
			w.record(fmt.Sprintf("health-prev:%s=%d", health.name, value))
			health.previous = value
		},
		storeHealthGraph: func(update *monsterInitTestUpdate4F0040, index int, value uint16) {
			w.record(fmt.Sprintf("graph:%s:%d=%d", update.name, index, value))
			update.healthGraph[index] = value
		},
		loadSubclassLow: func(unit *monsterInitTestObject4F0040) uint8 {
			w.record(fmt.Sprintf("subclass:%s=%#02x", unit.name, unit.subclass))
			return unit.subclass
		},
		loadSpeedField332: func(update *monsterInitTestUpdate4F0040) float32 {
			w.record(fmt.Sprintf("field332:%s=%#08x", update.name, math.Float32bits(update.field332)))
			return update.field332
		},
		loadSpeedField333: func(update *monsterInitTestUpdate4F0040) uint32 {
			w.record(fmt.Sprintf("field333:%s=%#08x", update.name, update.field333))
			return update.field333
		},
		randomFloat: func(minimum, maximum float32, source string, line int) float64 {
			w.record(fmt.Sprintf(
				"random:%#08x:%#08x:%s:%d",
				math.Float32bits(minimum), math.Float32bits(maximum), source, line,
			))
			return w.random
		},
		loadSpeedBase: func(unit *monsterInitTestObject4F0040) float32 {
			w.record(fmt.Sprintf("speed-load:%s=%#08x", unit.name, math.Float32bits(unit.speedBase)))
			return unit.speedBase
		},
		storeSpeedBase: func(unit *monsterInitTestObject4F0040, value float32) {
			w.record(fmt.Sprintf("speed-store:%s=%#08x", unit.name, math.Float32bits(value)))
			unit.speedBase = value
		},
		canCast: func(unit *monsterInitTestObject4F0040) bool {
			w.record(fmt.Sprintf("can-cast:%s=%t", unit.name, w.canCastFlag))
			return w.canCastFlag
		},
		storeFleeRange: func(update *monsterInitTestUpdate4F0040, value float32) {
			w.record(fmt.Sprintf("flee:%s=%#08x", update.name, math.Float32bits(value)))
			update.fleeRange = value
		},
	}
}

func newMonsterInitTestWorld4F0040() *monsterInitTestWorld4F0040 {
	def := &monsterInitTestDef4F0040{name: "def", meleeRange: 12.5}
	health := &monsterInitTestHealth4F0040{name: "h0", current: 100, previous: 1, maximum: 100}
	update := &monsterInitTestUpdate4F0040{
		name:        "update",
		monsterDef:  def,
		aiAction:    0xaaaaaaaa,
		status:      monsterInitStatusHold4F0040 | monsterInitStatusAlwaysRun4F0040,
		field332:    3,
		field333:    0x1234567e,
		healthScale: 1.5,
		fleeRange:   -1,
	}
	unit := &monsterInitTestObject4F0040{
		name:      "unit",
		update:    update,
		typeIndex: 7,
		radius:    3.25,
		positionX: 0x7fc12345,
		positionY: 0x80000000,
		direction: -32767,
		health:    health,
		subclass:  monsterInitNPCSubclassMask4F0040,
		speedBase: 9,
	}
	return &monsterInitTestWorld4F0040{
		unit:        unit,
		plantTypeID: 7,
		canCastFlag: true,
		frame:       0x89abcdef,
		random:      1.0,
		nilPush:     make(map[uint32]bool),
	}
}

func monsterInitExpectedPlantEvents4F0040() []string {
	events := []string{
		"arg:unit", "update:unit=update", "plant-id=7", "flags:unit=0x00000000",
		"type:unit=7", "clear:unit", "monster-def:update=def", "ai-store:update=4",
		"melee:def=0x41480000", "radius:unit=0x40500000", "sight:update=0x41ce0000",
		"ai-load:update=4", "push:unit=4",
		"position-x:unit=0x7fc12345", "arg-store:action-4-0:0=0x7fc12345",
		"position-y:unit=0x80000000", "arg-store:action-4-0:1=0x80000000",
		"direction:unit=-32767", "arg-store:action-4-0:2=0xffff8001",
		"ai-store:update=38",
		"direction:unit=-32767", "direction-store:update=0xffff8001",
		"position-x:unit=0x7fc12345", "position-x-store:update=0x7fc12345",
		"position-y:unit=0x80000000", "position-y-store:update=0x80000000",
		"health:unit=h0", "health-max:h0=100", "health-cur:h0=100",
		"health-scale:update=0x3fc00000", "health-set:unit=150",
		"health:unit=h0", "health-cur:h0=150", "health-prev:h0=150",
	}
	for index := 0; index < 32; index++ {
		events = append(events,
			"health:unit=h0",
			"health-cur:h0=150",
			fmt.Sprintf("graph:update:%d=150", index),
		)
	}
	return append(events,
		"subclass:unit=0x30", "field332:update=0x40400000", "speed-store:unit=0x404ccccd",
		"can-cast:unit=true", "flee:update=0x42c80000",
		"status-load:update=0x00008040", "flee:update=0x00000000",
		"status-store:update=0x0000c040",
	)
}

func TestMonsterInit4F0040ExactPlantOrderAndState(t *testing.T) {
	world := newMonsterInitTestWorld4F0040()
	monsterInit4F0040(world.hooks())

	if want := monsterInitExpectedPlantEvents4F0040(); !reflect.DeepEqual(world.events, want) {
		t.Fatalf("events =\n%q\nwant =\n%q", world.events, want)
	}
	update := world.unit.update
	if update.aiAction != monsterInitActionInvalid4F0040 || math.Float32bits(update.sightRange) != 0x41ce0000 {
		t.Fatalf("action/sight = %d/%#08x", update.aiAction, math.Float32bits(update.sightRange))
	}
	if len(world.pushed) != 1 || world.pushed[0].action != monsterInitActionGuard4F0040 ||
		world.pushed[0].args != [4]uint32{0x7fc12345, 0x80000000, 0xffff8001, 0xd0d0d0d0} {
		t.Fatalf("guard action = %#v", world.pushed)
	}
	if update.direction != 0xffff8001 || update.positionX != 0x7fc12345 || update.positionY != 0x80000000 {
		t.Fatalf("cached direction/position = %#08x/%#08x/%#08x", update.direction, update.positionX, update.positionY)
	}
	if world.unit.health.current != 150 || world.unit.health.previous != 150 {
		t.Fatalf("health = %+v", world.unit.health)
	}
	for index, value := range update.healthGraph {
		if value != 150 {
			t.Fatalf("health graph %d = %d", index, value)
		}
	}
	if math.Float32bits(world.unit.speedBase) != 0x404ccccd || math.Float32bits(update.fleeRange) != 0 {
		t.Fatalf("speed/flee = %#08x/%#08x", math.Float32bits(world.unit.speedBase), math.Float32bits(update.fleeRange))
	}
	if update.status != monsterInitStatusHold4F0040|monsterInitStatusAlwaysRun4F0040|monsterInitStatusRunning4F0040 {
		t.Fatalf("status = %#08x", update.status)
	}
}

func TestMonsterInit4F0040EveryObservablePlantFaultPrefix(t *testing.T) {
	want := monsterInitExpectedPlantEvents4F0040()
	for faultAt := 1; faultAt <= len(want); faultAt++ {
		world := newMonsterInitTestWorld4F0040()
		world.faultAt = faultAt
		panicked := false
		func() {
			defer func() { panicked = recover() != nil }()
			monsterInit4F0040(world.hooks())
		}()
		if !panicked {
			t.Fatalf("fault %d did not panic", faultAt)
		}
		if !reflect.DeepEqual(world.events, want[:faultAt]) {
			t.Fatalf("fault %d events = %q, want %q", faultAt, world.events, want[:faultAt])
		}
	}
}

func TestMonsterInit4F0040SpeciesBranchesAndSkipFlags(t *testing.T) {
	tests := []struct {
		name       string
		flags      uint32
		rat        bool
		fish       bool
		frog       bool
		wantChecks []string
		wantAction uint32
		wantStatus uint32
	}{
		{name: "ordinary", wantChecks: []string{"rat:unit=false", "fish:unit=false", "frog:unit=false"}, wantAction: math.MaxUint32},
		{name: "rat", rat: true, wantChecks: []string{"rat:unit=true"}, wantAction: monsterInitActionRandomWalk4F0040},
		{name: "fish", fish: true, wantChecks: []string{"rat:unit=false", "fish:unit=true"}, wantAction: monsterInitActionRoam4F0040},
		{name: "frog", frog: true, wantChecks: []string{"rat:unit=false", "fish:unit=false", "frog:unit=true"}, wantAction: monsterInitActionIdle4F0040, wantStatus: monsterInitStatusAlert4F0040},
		{name: "dead", flags: 0x8000, wantAction: math.MaxUint32},
		{name: "destroyed", flags: 0x20, wantAction: math.MaxUint32},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			world := newMonsterInitTestWorld4F0040()
			world.unit.typeIndex = 99
			world.unit.flags = tc.flags
			world.unit.update.aiAction = monsterInitActionInvalid4F0040
			world.unit.update.status = 0
			world.unit.health.current = 99
			world.rat, world.fish, world.frog = tc.rat, tc.fish, tc.frog
			world.canCastFlag = false
			monsterInit4F0040(world.hooks())

			var checks []string
			for _, event := range world.events {
				if strings.HasPrefix(event, "rat:") || strings.HasPrefix(event, "fish:") || strings.HasPrefix(event, "frog:") {
					checks = append(checks, event)
				}
			}
			if !reflect.DeepEqual(checks, tc.wantChecks) {
				t.Fatalf("predicate checks = %v, want %v", checks, tc.wantChecks)
			}
			if tc.wantAction == math.MaxUint32 {
				if len(world.pushed) != 0 {
					t.Fatalf("pushed = %#v", world.pushed)
				}
			} else if len(world.pushed) != 1 || world.pushed[0].action != tc.wantAction {
				t.Fatalf("pushed = %#v, want %d", world.pushed, tc.wantAction)
			}
			if tc.fish && world.pushed[0].args != [4]uint32{0, 0xb0b0b0b0, 0xc0ffeeff, 0xd0d0d0d0} {
				t.Fatalf("fish arguments = %#v", world.pushed[0].args)
			}
			if math.Float32bits(world.unit.update.aggression) != func() uint32 {
				if tc.rat || tc.fish || tc.frog {
					return monsterInitAggressionBits4F0040
				}
				return 0
			}() {
				t.Fatalf("aggression = %#08x", math.Float32bits(world.unit.update.aggression))
			}
			if world.unit.update.status != tc.wantStatus {
				t.Fatalf("status = %#08x, want %#08x", world.unit.update.status, tc.wantStatus)
			}
		})
	}
}

func TestMonsterInit4F0040ActionJumpTable(t *testing.T) {
	tests := []struct {
		name      string
		action    uint32
		canAttack bool
		wantPush  uint32
		wantArgs  [4]uint32
		wantNone  bool
	}{
		{name: "escort", action: 3, wantPush: 3, wantArgs: [4]uint32{0x7fc12345, 0x80000000, 0, 0xd0d0d0d0}},
		{name: "guard", action: 4, wantPush: 4, wantArgs: [4]uint32{0x7fc12345, 0x80000000, 0xffff8001, 0xd0d0d0d0}},
		{name: "roam", action: 10, wantPush: 10, wantArgs: [4]uint32{0, 0xb0b0b0b0, 0xc0ffee7e, 0xd0d0d0d0}},
		{name: "roam hunts", action: 10, canAttack: true, wantPush: 5, wantArgs: [4]uint32{0xa0a0a0a0, 0xb0b0b0b0, 0xc0ffee00, 0xd0d0d0d0}},
		{name: "fight", action: 15, wantPush: 15, wantArgs: [4]uint32{0x7fc12345, 0x80000000, 0x89abcdef, 0xd0d0d0d0}},
		{name: "invalid", action: 38, wantNone: true},
		{name: "below table", action: 0, wantPush: 0, wantArgs: [4]uint32{0xa0a0a0a0, 0xb0b0b0b0, 0xc0ffee00, 0xd0d0d0d0}},
		{name: "inside default", action: 5, wantPush: 0, wantArgs: [4]uint32{0xa0a0a0a0, 0xb0b0b0b0, 0xc0ffee00, 0xd0d0d0d0}},
		{name: "above table", action: 39, wantPush: 0, wantArgs: [4]uint32{0xa0a0a0a0, 0xb0b0b0b0, 0xc0ffee00, 0xd0d0d0d0}},
		{name: "unsigned maximum", action: math.MaxUint32, wantPush: 0, wantArgs: [4]uint32{0xa0a0a0a0, 0xb0b0b0b0, 0xc0ffee00, 0xd0d0d0d0}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			world := newMonsterInitTestWorld4F0040()
			world.unit.typeIndex = 99
			world.unit.update.aiAction = tc.action
			world.unit.update.status = 0
			world.unit.health.current = 99
			world.canAttack = tc.canAttack
			world.canCastFlag = false
			monsterInit4F0040(world.hooks())

			if tc.wantNone {
				if len(world.pushed) != 0 {
					t.Fatalf("pushed = %#v", world.pushed)
				}
				return
			}
			if len(world.pushed) != 1 || world.pushed[0].action != tc.wantPush || world.pushed[0].args != tc.wantArgs {
				t.Fatalf("pushed = %#v, want action %d args %#v", world.pushed, tc.wantPush, tc.wantArgs)
			}
			if world.unit.update.aiAction != monsterInitActionInvalid4F0040 {
				t.Fatalf("final AI action = %d", world.unit.update.aiAction)
			}
		})
	}
}

func TestMonsterInit4F0040CheckedActionPointers(t *testing.T) {
	for _, action := range []uint32{
		monsterInitActionEscort4F0040,
		monsterInitActionGuard4F0040,
		monsterInitActionRoam4F0040,
		monsterInitActionFight4F0040,
	} {
		t.Run(fmt.Sprintf("action-%d", action), func(t *testing.T) {
			world := newMonsterInitTestWorld4F0040()
			world.unit.typeIndex = 99
			world.unit.update.aiAction = action
			world.unit.update.status = 0
			world.unit.health.current = 99
			world.canCastFlag = false
			world.nilPush[action] = true
			monsterInit4F0040(world.hooks())
			if len(world.pushed) != 0 {
				t.Fatalf("nil action was retained = %#v", world.pushed)
			}
			for _, event := range world.events {
				if strings.HasPrefix(event, "arg-store:") || strings.HasPrefix(event, "arg-low:") {
					t.Fatalf("argument store after nil PushAction: %s", event)
				}
			}
		})
	}

	t.Run("fish", func(t *testing.T) {
		world := newMonsterInitTestWorld4F0040()
		world.unit.typeIndex = 99
		world.unit.update.aiAction = monsterInitActionInvalid4F0040
		world.unit.update.status = 0
		world.unit.health.current = 99
		world.fish = true
		world.canCastFlag = false
		world.nilPush[monsterInitActionRoam4F0040] = true
		monsterInit4F0040(world.hooks())
		for _, event := range world.events {
			if strings.HasPrefix(event, "arg-store:") || strings.HasPrefix(event, "arg-low:") {
				t.Fatalf("fish argument store after nil PushAction: %s", event)
			}
		}
	})
}

func TestMonsterInit4F0040CachesUpdateAndReloadsEveryHealthSample(t *testing.T) {
	world := newMonsterInitTestWorld4F0040()
	cached := world.unit.update
	replacement := &monsterInitTestUpdate4F0040{name: "replacement", aiAction: monsterInitActionInvalid4F0040}
	world.unit.typeIndex = 99
	cached.aiAction = monsterInitActionInvalid4F0040
	cached.status = 0
	world.unit.health.current = 11
	world.unit.health.maximum = 12
	world.canCastFlag = false

	graphHealth := make([]*monsterInitTestHealth4F0040, 32)
	for index := range graphHealth {
		graphHealth[index] = &monsterInitTestHealth4F0040{
			name:    fmt.Sprintf("graph-%d", index),
			current: uint16(1000 + index),
			maximum: 2000,
		}
	}
	world.onEvent = func(event string) {
		switch event {
		case "update:unit=update":
			world.unit.update = replacement
		case "health-prev:h0=11":
			world.unit.health = graphHealth[0]
		default:
			for index := 0; index < len(graphHealth)-1; index++ {
				if event == fmt.Sprintf("graph:update:%d=%d", index, 1000+index) {
					world.unit.health = graphHealth[index+1]
					break
				}
			}
		}
	}

	monsterInit4F0040(world.hooks())
	if replacement != world.unit.update || replacement.direction != 0 || replacement.positionX != 0 || replacement.positionY != 0 {
		t.Fatalf("replacement UpdateData changed = %+v", replacement)
	}
	if cached.aiAction != monsterInitActionInvalid4F0040 || cached.direction != 0xffff8001 || cached.positionX != 0x7fc12345 {
		t.Fatalf("cached UpdateData state = %+v", cached)
	}
	for index, value := range cached.healthGraph {
		if want := uint16(1000 + index); value != want {
			t.Fatalf("health graph %d = %d, want %d", index, value, want)
		}
	}
	if graphHealth[31] != world.unit.health {
		t.Fatalf("last live HealthData = %p, want %p", world.unit.health, graphHealth[31])
	}
}

func TestMonsterInit4F0040NumericContracts(t *testing.T) {
	if got := math.Float32bits(monsterInitPlantSight4F0040(12.5, 3.25)); got != 0x41ce0000 {
		t.Fatalf("plant sight bits = %#08x", got)
	}
	if got := math.Float32bits(monsterInitNPCSpeed4F0040(3)); got != 0x404ccccd {
		t.Fatalf("NPC speed bits = %#08x", got)
	}
	if got := math.Float32bits(monsterInitRandomSpeed4F0040(1.05, 7.25)); got != math.Float32bits(float32(1.05*7.25)) {
		t.Fatalf("random speed bits = %#08x", got)
	}

	tests := []struct {
		name    string
		maximum uint16
		scale   float32
		want    uint16
	}{
		{name: "truncate", maximum: 9, scale: 1.5, want: 13},
		{name: "low word", maximum: 65535, scale: 1.5, want: 32766},
		{name: "negative wraps", maximum: 65535, scale: -1, want: 1},
		{name: "nan indefinite", maximum: 1, scale: math.Float32frombits(0x7fc12345), want: 0},
		{name: "positive infinity indefinite", maximum: 1, scale: float32(math.Inf(1)), want: 0},
		{name: "negative infinity indefinite", maximum: 1, scale: float32(math.Inf(-1)), want: 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := monsterInitScaledHealth4F0040(tc.maximum, tc.scale); got != tc.want {
				t.Fatalf("scaled health = %#04x, want %#04x", got, tc.want)
			}
		})
	}
}

func TestMonsterInit4F0040RandomSpeedCallAndLiveSpeedLoad(t *testing.T) {
	world := newMonsterInitTestWorld4F0040()
	world.unit.typeIndex = 99
	world.unit.update.aiAction = monsterInitActionInvalid4F0040
	world.unit.update.status = 0
	world.unit.health.current = 99
	world.unit.subclass = 0
	world.unit.speedBase = 4
	world.random = 1.25
	world.canCastFlag = false
	world.onEvent = func(event string) {
		if strings.HasPrefix(event, "random:") {
			world.unit.speedBase = 8
		}
	}
	monsterInit4F0040(world.hooks())

	wantRandom := fmt.Sprintf(
		"random:%#08x:%#08x:%s:%d",
		monsterInitSpeedMinBits4F0040,
		monsterInitSpeedMaxBits4F0040,
		monsterInitRandomSource4F0040,
		monsterInitRandomLine4F0040,
	)
	found := false
	for index, event := range world.events {
		if event == wantRandom {
			found = true
			if index+1 >= len(world.events) || world.events[index+1] != "speed-load:unit=0x41000000" {
				t.Fatalf("events after random = %q", world.events[index:])
			}
			break
		}
	}
	if !found {
		t.Fatalf("random call missing from %q", world.events)
	}
	if math.Float32bits(world.unit.speedBase) != math.Float32bits(10) {
		t.Fatalf("live speed result = %#08x", math.Float32bits(world.unit.speedBase))
	}
}
