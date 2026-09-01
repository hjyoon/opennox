package server

import (
	"math"
	"reflect"
	"testing"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

type monsterGeneratorDamageTestUpdate4E27D0 struct{ name string }

type monsterGeneratorDamageTestHealth4E27D0 struct {
	name         string
	current, max uint16
}

type monsterGeneratorDamageTestObject4E27D0 struct {
	name        string
	flags       uint32
	update      *monsterGeneratorDamageTestUpdate4E27D0
	effectFrame uint32
	pos         types.Pointf
	hp          uint16
	health      *monsterGeneratorDamageTestHealth4E27D0
	xstatus     uint32
}

type monsterGeneratorDamageTestWorld4E27D0 struct {
	events      []string
	frameValue  uint32
	fxOp        byte
	fxSubtype   byte
	fxPoint     types.Pointf
	fxCalls     int
	audioID     int
	audioTarget *monsterGeneratorDamageTestObject4E27D0
	audioCalls  int

	defaultResult bool
	defaultCall   func(
		*monsterGeneratorDamageTestObject4E27D0,
		*monsterGeneratorDamageTestObject4E27D0,
		*monsterGeneratorDamageTestObject4E27D0,
		int32,
		object.DamageType,
	)
	scriptCall func(
		*monsterGeneratorDamageTestUpdate4E27D0,
		*monsterGeneratorDamageTestObject4E27D0,
		*monsterGeneratorDamageTestObject4E27D0,
		ScriptEventType,
	)
	healthLoads []*monsterGeneratorDamageTestHealth4E27D0
	setCall     func(*monsterGeneratorDamageTestObject4E27D0, uint32)
	unsetCall   func(*monsterGeneratorDamageTestObject4E27D0, uint32)
}

func monsterGeneratorDamageTestObjectName4E27D0(object *monsterGeneratorDamageTestObject4E27D0) string {
	if object == nil {
		return "nil"
	}
	return object.name
}

func (world *monsterGeneratorDamageTestWorld4E27D0) hooks() monsterGeneratorDamageHooks4E27D0[
	*monsterGeneratorDamageTestObject4E27D0,
	*monsterGeneratorDamageTestUpdate4E27D0,
	*monsterGeneratorDamageTestHealth4E27D0,
] {
	return monsterGeneratorDamageHooks4E27D0[
		*monsterGeneratorDamageTestObject4E27D0,
		*monsterGeneratorDamageTestUpdate4E27D0,
		*monsterGeneratorDamageTestHealth4E27D0,
	]{
		loadFlags: func(object *monsterGeneratorDamageTestObject4E27D0) uint32 {
			world.events = append(world.events, "flags:"+monsterGeneratorDamageTestObjectName4E27D0(object))
			return object.flags
		},
		loadUpdate: func(object *monsterGeneratorDamageTestObject4E27D0) *monsterGeneratorDamageTestUpdate4E27D0 {
			world.events = append(world.events, "update:"+monsterGeneratorDamageTestObjectName4E27D0(object))
			return object.update
		},
		frame: func() uint32 {
			world.events = append(world.events, "frame")
			return world.frameValue
		},
		loadEffectFrame: func(object *monsterGeneratorDamageTestObject4E27D0) uint32 {
			world.events = append(world.events, "effect-frame:"+monsterGeneratorDamageTestObjectName4E27D0(object))
			return object.effectFrame
		},
		loadPosX: func(object *monsterGeneratorDamageTestObject4E27D0) float32 {
			world.events = append(world.events, "x:"+monsterGeneratorDamageTestObjectName4E27D0(object))
			return object.pos.X
		},
		loadPosY: func(object *monsterGeneratorDamageTestObject4E27D0) float32 {
			world.events = append(world.events, "y:"+monsterGeneratorDamageTestObjectName4E27D0(object))
			return object.pos.Y
		},
		normalize: func(point *types.Pointf) {
			world.events = append(world.events, "normalize")
			chestOpenNormalizeVector509F20(point)
		},
		pointFX: func(op, subtype byte, point types.Pointf) {
			world.events = append(world.events, "fx")
			world.fxOp, world.fxSubtype, world.fxPoint = op, subtype, point
			world.fxCalls++
		},
		audio: func(id int, target *monsterGeneratorDamageTestObject4E27D0) {
			world.events = append(world.events, "audio")
			world.audioID, world.audioTarget = id, target
			world.audioCalls++
		},
		getHP: func(object *monsterGeneratorDamageTestObject4E27D0) uint16 {
			world.events = append(world.events, "hp:"+monsterGeneratorDamageTestObjectName4E27D0(object))
			return object.hp
		},
		defaultDamage: func(target, source, weapon *monsterGeneratorDamageTestObject4E27D0, damage int32, typ object.DamageType) bool {
			world.events = append(world.events, "default")
			if world.defaultCall != nil {
				world.defaultCall(target, source, weapon, damage, typ)
			}
			return world.defaultResult
		},
		scriptDamage: func(update *monsterGeneratorDamageTestUpdate4E27D0, source, target *monsterGeneratorDamageTestObject4E27D0, event ScriptEventType) {
			world.events = append(world.events, "script")
			if world.scriptCall != nil {
				world.scriptCall(update, source, target, event)
			}
		},
		loadHealth: func(object *monsterGeneratorDamageTestObject4E27D0) *monsterGeneratorDamageTestHealth4E27D0 {
			world.events = append(world.events, "health:"+monsterGeneratorDamageTestObjectName4E27D0(object))
			if len(world.healthLoads) != 0 {
				health := world.healthLoads[0]
				world.healthLoads = world.healthLoads[1:]
				return health
			}
			return object.health
		},
		loadHealthMax: func(health *monsterGeneratorDamageTestHealth4E27D0) uint16 {
			world.events = append(world.events, "max:"+health.name)
			return health.max
		},
		loadHealthCur: func(health *monsterGeneratorDamageTestHealth4E27D0) uint16 {
			world.events = append(world.events, "cur:"+health.name)
			return health.current
		},
		loadXStatus: func(object *monsterGeneratorDamageTestObject4E27D0) uint32 {
			world.events = append(world.events, "xstatus")
			return object.xstatus
		},
		setXStatus: func(object *monsterGeneratorDamageTestObject4E27D0, status uint32) {
			world.events = append(world.events, "set")
			if world.setCall != nil {
				world.setCall(object, status)
				return
			}
			object.xstatus |= status
		},
		unsetXStatus: func(object *monsterGeneratorDamageTestObject4E27D0, status uint32) {
			world.events = append(world.events, "unset")
			if world.unsetCall != nil {
				world.unsetCall(object, status)
				return
			}
			object.xstatus &^= status
		},
	}
}

func TestMonsterGeneratorDamageEarlyFlagsCacheUpdateFirst4E27D0(t *testing.T) {
	for _, flag := range []uint32{uint32(object.FlagDead), uint32(object.FlagDestroyed), uint32(object.FlagDead | object.FlagDestroyed)} {
		world := new(monsterGeneratorDamageTestWorld4E27D0)
		target := &monsterGeneratorDamageTestObject4E27D0{name: "target", flags: flag}
		if got := monsterGeneratorDamage4E27D0(target, nil, nil, 7, object.DamageBlade, world.hooks()); got {
			t.Fatalf("flags %#x result = true, want false", flag)
		}
		want := []string{"flags:target", "update:target"}
		if !reflect.DeepEqual(world.events, want) {
			t.Fatalf("flags %#x events = %v, want %v", flag, world.events, want)
		}
	}
}

func TestMonsterGeneratorDamageOrderedHighHealthPath4E27D0(t *testing.T) {
	health := &monsterGeneratorDamageTestHealth4E27D0{name: "health", current: 250, max: 300}
	target := &monsterGeneratorDamageTestObject4E27D0{name: "target", effectFrame: 10, hp: 250, health: health}
	world := &monsterGeneratorDamageTestWorld4E27D0{frameValue: 11, defaultResult: true}
	if got := monsterGeneratorDamage4E27D0(target, nil, nil, -17, object.DamageElectric, world.hooks()); !got {
		t.Fatal("result = false, want true")
	}
	want := []string{
		"flags:target", "update:target", "frame", "effect-frame:target",
		"hp:target", "default", "hp:target", "flags:target",
		"health:target", "max:health", "cur:health", "health:target", "max:health",
	}
	if !reflect.DeepEqual(world.events, want) {
		t.Fatalf("events = %v, want %v", world.events, want)
	}
}

func TestMonsterGeneratorDamageEffectCadenceAndPositions4E27D0(t *testing.T) {
	health := &monsterGeneratorDamageTestHealth4E27D0{name: "health", current: 250, max: 300}
	for _, test := range []struct {
		name        string
		frame       uint32
		effectFrame uint32
		weapon      *monsterGeneratorDamageTestObject4E27D0
		wantFX      bool
		wantXBits   uint32
		wantYBits   uint32
	}{
		{name: "recent", frame: 29, effectFrame: 10},
		{name: "unsigned wrap remains recent", frame: 1, effectFrame: math.MaxUint32 - 15},
		{name: "periodic target position", frame: 30, effectFrame: 30, wantFX: true, wantXBits: 0x41200000, wantYBits: 0x41a00000},
		{name: "old hit weapon direction", frame: 31, effectFrame: 10, weapon: &monsterGeneratorDamageTestObject4E27D0{name: "weapon", pos: types.Ptf(13, 24)}, wantFX: true, wantXBits: 0x41b9a7f9, wantYBits: 0x42166102},
	} {
		t.Run(test.name, func(t *testing.T) {
			target := &monsterGeneratorDamageTestObject4E27D0{
				name: "target", effectFrame: test.effectFrame, pos: types.Ptf(10, 20), hp: 250, health: health,
			}
			world := &monsterGeneratorDamageTestWorld4E27D0{frameValue: test.frame}
			monsterGeneratorDamage4E27D0(target, nil, test.weapon, 1, object.DamageBlade, world.hooks())
			if got := world.fxCalls != 0; got != test.wantFX {
				t.Fatalf("fx called = %t, want %t", got, test.wantFX)
			}
			if world.audioCalls != world.fxCalls {
				t.Fatalf("audio/fx calls = %d/%d", world.audioCalls, world.fxCalls)
			}
			if !test.wantFX {
				return
			}
			if world.fxOp != monsterGeneratorDamageFxOpcode4E27D0 || world.fxSubtype != monsterGeneratorDamageFxSubtype4E27D0 ||
				world.audioID != monsterGeneratorDamageSound4E27D0 || world.audioTarget != target {
				t.Fatalf("effect args = %#x/%d/%d/%p", world.fxOp, world.fxSubtype, world.audioID, world.audioTarget)
			}
			if math.Float32bits(world.fxPoint.X) != test.wantXBits || math.Float32bits(world.fxPoint.Y) != test.wantYBits {
				t.Fatalf("point bits = %#08x/%#08x, want %#08x/%#08x",
					math.Float32bits(world.fxPoint.X), math.Float32bits(world.fxPoint.Y), test.wantXBits, test.wantYBits)
			}
		})
	}
}

func TestMonsterGeneratorDamageCachesScriptAndReloadsHealth4E27D0(t *testing.T) {
	initialUpdate := &monsterGeneratorDamageTestUpdate4E27D0{name: "initial"}
	replacementUpdate := &monsterGeneratorDamageTestUpdate4E27D0{name: "replacement"}
	firstHealth := &monsterGeneratorDamageTestHealth4E27D0{name: "first", current: 201, max: 600}
	secondHealth := &monsterGeneratorDamageTestHealth4E27D0{name: "second", current: 1, max: 200}
	target := &monsterGeneratorDamageTestObject4E27D0{
		name: "target", update: initialUpdate, effectFrame: 10, hp: 100, health: firstHealth,
	}
	source := &monsterGeneratorDamageTestObject4E27D0{name: "source"}
	weapon := &monsterGeneratorDamageTestObject4E27D0{name: "weapon"}
	scriptCalls := 0
	world := &monsterGeneratorDamageTestWorld4E27D0{
		frameValue:    11,
		defaultResult: true,
		healthLoads:   []*monsterGeneratorDamageTestHealth4E27D0{firstHealth, secondHealth},
		defaultCall: func(gotTarget, gotSource, gotWeapon *monsterGeneratorDamageTestObject4E27D0, damage int32, typ object.DamageType) {
			if gotTarget != target || gotSource != source || gotWeapon != weapon || damage != -23 || typ != object.DamageFlame {
				t.Fatalf("default args = %p/%p/%p/%d/%d", gotTarget, gotSource, gotWeapon, damage, typ)
			}
			target.update = replacementUpdate
			target.hp = 90
		},
		scriptCall: func(update *monsterGeneratorDamageTestUpdate4E27D0, gotSource, gotTarget *monsterGeneratorDamageTestObject4E27D0, event ScriptEventType) {
			scriptCalls++
			if update != initialUpdate || gotSource != source || gotTarget != target || event != NoxEventGeneratorDamage {
				t.Fatalf("script args = %p/%p/%p/%d", update, gotSource, gotTarget, event)
			}
		},
	}
	if got := monsterGeneratorDamage4E27D0(target, source, weapon, -23, object.DamageFlame, world.hooks()); !got {
		t.Fatal("result = false, want true")
	}
	if scriptCalls != 1 {
		t.Fatalf("script calls = %d, want 1", scriptCalls)
	}
	for _, event := range world.events {
		if event == "set" || event == "unset" || event == "xstatus" {
			t.Fatalf("second live health pointer was not used: events = %v", world.events)
		}
	}
}

func TestMonsterGeneratorDamageHealthBandsAndLiveStatus4E27D0(t *testing.T) {
	if got := monsterGeneratorDamageRoundThreshold4E27D0(300, monsterGeneratorDamageThirdBits4E27D0); got != 100 {
		t.Fatalf("one-third threshold = %d, want 100", got)
	}
	if got := monsterGeneratorDamageRoundThreshold4E27D0(300, monsterGeneratorDamageTwoThirdBits4E27D0); got != 200 {
		t.Fatalf("two-thirds threshold = %d, want 200", got)
	}

	for _, test := range []struct {
		name       string
		current    uint16
		status     uint32
		wantStatus uint32
		wantCalls  []string
	}{
		{name: "above two thirds", current: 201, wantStatus: 0},
		{name: "at two thirds", current: 200, wantStatus: 0x100, wantCalls: []string{"set"}},
		{name: "middle already marked", current: 101, status: 0x100, wantStatus: 0x100},
		{name: "at one third", current: 100, status: 0x100, wantStatus: 0x200, wantCalls: []string{"unset", "set"}},
		{name: "low already marked", current: 10, status: 0x200, wantStatus: 0x200},
	} {
		t.Run(test.name, func(t *testing.T) {
			health := &monsterGeneratorDamageTestHealth4E27D0{name: "health", current: test.current, max: 300}
			target := &monsterGeneratorDamageTestObject4E27D0{name: "target", effectFrame: 10, hp: test.current, health: health, xstatus: test.status}
			world := &monsterGeneratorDamageTestWorld4E27D0{frameValue: 11}
			monsterGeneratorDamage4E27D0(target, nil, nil, 1, object.DamageBlade, world.hooks())
			if target.xstatus != test.wantStatus {
				t.Fatalf("status = %#x, want %#x", target.xstatus, test.wantStatus)
			}
			var calls []string
			for _, event := range world.events {
				if event == "set" || event == "unset" {
					calls = append(calls, event)
				}
			}
			if !reflect.DeepEqual(calls, test.wantCalls) {
				t.Fatalf("status calls = %v, want %v", calls, test.wantCalls)
			}
		})
	}

	health := &monsterGeneratorDamageTestHealth4E27D0{name: "health", current: 10, max: 300}
	target := &monsterGeneratorDamageTestObject4E27D0{name: "target", effectFrame: 10, hp: 10, health: health, xstatus: 0x100}
	setCalls := 0
	world := &monsterGeneratorDamageTestWorld4E27D0{
		frameValue: 11,
		unsetCall: func(object *monsterGeneratorDamageTestObject4E27D0, status uint32) {
			object.xstatus &^= status
			object.xstatus |= monsterGeneratorDamageStatusLow4E27D0
		},
		setCall: func(*monsterGeneratorDamageTestObject4E27D0, uint32) { setCalls++ },
	}
	monsterGeneratorDamage4E27D0(target, nil, nil, 1, object.DamageBlade, world.hooks())
	if target.xstatus != 0x200 || setCalls != 0 {
		t.Fatalf("live status/set calls = %#x/%d, want 0x200/0", target.xstatus, setCalls)
	}
}
