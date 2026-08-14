package server

import (
	"math"
	"reflect"
	"strconv"
	"testing"

	"github.com/opennox/libs/types"
)

type boomCollideTestState4E9770 struct {
	events       []string
	ready        uint32
	direct       int32
	splash       int32
	radius       float32
	pushRange    float32
	force        float32
	balance      map[string]float64
	classes      map[string]uint8
	velocityX    float32
	velocityY    float32
	direction2   uint16
	pointX       int32
	pointY       int32
	quest        int32
	parent       string
	enemy        int32
	inversion    int32
	enchant      int32
	direction    int16
	directionHit int32
}

func (s *boomCollideTestState4E9770) event(name string) {
	s.events = append(s.events, name)
}

func (s *boomCollideTestState4E9770) hooks() boomCollideHooks4E9770[string, string, string] {
	return boomCollideHooks4E9770[string, string, string]{
		loadBalanceReady: func() uint32 {
			s.event("ready")
			return s.ready
		},
		gameDataFloat: func(name string) float64 {
			s.event("balance:" + name)
			return s.balance[name]
		},
		floatToInt: func(value float32) int32 {
			s.event("round:" + strconv.FormatFloat(float64(value), 'g', -1, 32))
			return playerCollideRound4E8460(value)
		},
		storeDirectDamage: func(value int32) {
			s.event("store-direct")
			s.direct = value
		},
		storeSplashDamage: func(value int32) {
			s.event("store-splash")
			s.splash = value
		},
		storeRange: func(value float32) {
			s.event("store-range")
			s.radius = value
		},
		storePushRange: func(value float32) {
			s.event("store-push-range")
			s.pushRange = value
		},
		storeForce: func(value float32) {
			s.event("store-force")
			s.force = value
		},
		storeBalanceReady: func(value uint32) {
			s.event("store-ready")
			s.ready = value
		},
		gameFlagsCheck: func(flag uint32) int32 {
			s.event("flags:" + strconv.FormatUint(uint64(flag), 10))
			return s.quest
		},
		findParent: func(obj string) string {
			s.event("parent:" + obj)
			return s.parent
		},
		classLow: func(obj string) uint8 {
			s.event("class:" + obj)
			return s.classes[obj]
		},
		isEnemy: func(first, second string) int32 {
			s.event("enemy:" + first + ":" + second)
			return s.enemy
		},
		pointFX: func(id uint32, obj string) {
			s.event("point-fx:" + strconv.FormatUint(uint64(id), 10) + ":" + obj)
		},
		inversion: func(target, source string) int32 {
			s.event("inversion:" + target + ":" + source)
			return s.inversion
		},
		changeOwner: func(source, target string) {
			s.event("change-owner:" + source + ":" + target)
		},
		hasEnchant: func(target string, enchant uint32) int32 {
			s.event("enchant:" + target + ":" + strconv.FormatUint(uint64(enchant), 10))
			return s.enchant
		},
		loadDirection: func(target string) int16 {
			s.event("direction:" + target)
			return s.direction
		},
		checkDirection: func(target string, direction int16, source string) int32 {
			s.event("check-direction:" + target + ":" + strconv.Itoa(int(direction)) + ":" + source)
			return s.directionHit
		},
		audio: func(id uint32, obj string, kind int32, code uint32) {
			s.event("audio:" + strconv.FormatUint(uint64(id), 10) + ":" + obj + ":" + strconv.Itoa(int(kind)) + ":" + strconv.FormatUint(uint64(code), 10))
		},
		loadDirectDamage: func() int32 {
			s.event("load-direct")
			return s.direct
		},
		targetDamage: func(target, parent, source string, damage int32, damageType uint32) int32 {
			s.event("damage:" + target + ":" + parent + ":" + source + ":" + strconv.Itoa(int(damage)) + ":" + strconv.FormatUint(uint64(damageType), 10))
			return 0
		},
		scorch: func(target string, kind int32) {
			s.event("scorch:" + target + ":" + strconv.Itoa(int(kind)))
		},
		wallReflect: func(collision, source string) {
			s.event("wall-reflect:" + collision + ":" + source)
		},
		loadVelocityX: func(source string) float32 {
			s.event("velocity-x:" + source)
			return s.velocityX
		},
		loadVelocityY: func(source string) float32 {
			s.event("velocity-y:" + source)
			return s.velocityY
		},
		vectorDirection: func(x, y float32) int32 {
			s.event("vector-direction:" + strconv.FormatFloat(float64(x), 'g', -1, 32) + ":" + strconv.FormatFloat(float64(y), 'g', -1, 32))
			return directionFromVector509ED0(x, y)
		},
		storeDirection2: func(source string, value uint16) {
			s.event("store-direction:" + source + ":" + strconv.Itoa(int(value)))
			s.direction2 = value
		},
		storeVelocityX: func(source string, value float32) {
			s.event("store-x:" + source + ":" + strconv.FormatFloat(float64(value), 'g', -1, 32))
			s.velocityX = value
		},
		storeVelocityY: func(source string, value float32) {
			s.event("store-y:" + source + ":" + strconv.FormatFloat(float64(value), 'g', -1, 32))
			s.velocityY = value
		},
		traceHitPoint: func() string {
			s.event("trace-point")
			return ""
		},
		loadPointY: func(point string) int32 {
			s.event("point-y:" + point)
			return s.pointY
		},
		loadPointX: func(point string) int32 {
			s.event("point-x:" + point)
			return s.pointX
		},
		damageMap: func(x, y, damage int32, damageType uint32, source string) {
			s.event("damage-map:" + strconv.Itoa(int(x)) + ":" + strconv.Itoa(int(y)) + ":" + strconv.Itoa(int(damage)) + ":" + strconv.FormatUint(uint64(damageType), 10) + ":" + source)
		},
		loadSplashDamage: func() int32 {
			s.event("load-splash")
			return s.splash
		},
		loadRange: func() float32 {
			s.event("load-range")
			return s.radius
		},
		mapDamageUnits: func(pos string, radius, inner float32, damage int32, damageType uint32, source, excluded string) {
			s.event("map-damage:" + pos + ":" + strconv.FormatFloat(float64(radius), 'g', -1, 32) + ":" + strconv.FormatFloat(float64(inner), 'g', -1, 32) + ":" + strconv.Itoa(int(damage)) + ":" + strconv.FormatUint(uint64(damageType), 10) + ":" + source + ":" + excluded)
		},
		loadForce: func() float32 {
			s.event("load-force")
			return s.force
		},
		loadPushRange: func() float32 {
			s.event("load-push-range")
			return s.pushRange
		},
		mapPushUnits: func(pos string, first, second, force float32, source string, arg6, arg7 int32) {
			s.event("map-push:" + pos + ":" + strconv.FormatFloat(float64(first), 'g', -1, 32) + ":" + strconv.FormatFloat(float64(second), 'g', -1, 32) + ":" + strconv.FormatFloat(float64(force), 'g', -1, 32) + ":" + source + ":" + strconv.Itoa(int(arg6)) + ":" + strconv.Itoa(int(arg7)))
		},
		delayedDelete: func(source string) {
			s.event("delete:" + source)
		},
	}
}

func TestBoomCollide4E9770InitializesBalanceInOrderThenExplodes(t *testing.T) {
	state := &boomCollideTestState4E9770{
		balance: map[string]float64{
			boomCollideDamageBalance4E9770:    10.5,
			boomCollideSplashBalance4E9770:    20.5,
			boomCollideRangeBalance4E9770:     30.25,
			boomCollidePushRangeBalance4E9770: 40.5,
			boomCollideForceBalance4E9770:     50.75,
		},
		classes: map[string]uint8{},
	}
	boomCollide4E9770("source", "", "", state.hooks())

	want := []string{
		"ready",
		"balance:MagicMissileDamage", "round:10.5", "store-direct",
		"balance:MagicMissileSplashDamage", "round:20.5", "store-splash",
		"balance:MagicMissileRange", "store-range",
		"balance:MagicMissilePushRange", "store-push-range",
		"balance:MagicMissileForce", "store-force", "store-ready",
		"flags:4096", "point-fx:134:source",
		"load-splash", "load-range", "map-damage:source:30.25:5:20:7:source:",
		"load-force", "load-push-range", "map-push:source:40.5:40.5:50.75:source:0:0",
		"audio:84:source:0:0", "delete:source",
	}
	if !reflect.DeepEqual(state.events, want) {
		t.Fatalf("events =\n%q\nwant\n%q", state.events, want)
	}
	if state.ready != 1 || state.direct != 10 || state.splash != 20 || state.radius != 30.25 || state.pushRange != 40.5 || state.force != 50.75 {
		t.Fatalf("cache = ready:%d direct:%d splash:%d range:%g push:%g force:%g",
			state.ready, state.direct, state.splash, state.radius, state.pushRange, state.force)
	}
}

func TestBoomCollide4E9770QuestFriendlyPlayersSuppressAllEffects(t *testing.T) {
	state := &boomCollideTestState4E9770{
		ready:   1,
		quest:   9,
		parent:  "player-owner",
		classes: map[string]uint8{"player-owner": 4, "target": 4},
	}
	boomCollide4E9770("source", "target", "collision", state.hooks())
	want := []string{
		"ready", "flags:4096", "parent:source", "class:player-owner", "class:target", "enemy:player-owner:target",
	}
	if !reflect.DeepEqual(state.events, want) {
		t.Fatalf("events = %q, want %q", state.events, want)
	}
}

func TestBoomCollide4E9770QuestShortCircuitOrder(t *testing.T) {
	tests := []struct {
		name    string
		source  string
		target  string
		parent  string
		classes map[string]uint8
		enemy   int32
		prefix  []string
	}{
		{name: "nil parent", source: "source", target: "target", prefix: []string{"ready", "flags:4096", "parent:source"}},
		{name: "nil target", source: "source", parent: "owner", prefix: []string{"ready", "flags:4096", "parent:source"}},
		{name: "non-player parent", source: "source", target: "target", parent: "owner", classes: map[string]uint8{"owner": 2}, prefix: []string{"ready", "flags:4096", "parent:source", "class:owner"}},
		{name: "non-player target", source: "source", target: "target", parent: "owner", classes: map[string]uint8{"owner": 4, "target": 2}, prefix: []string{"ready", "flags:4096", "parent:source", "class:owner", "class:target"}},
		{name: "enemy", source: "source", target: "target", parent: "owner", classes: map[string]uint8{"owner": 4, "target": 4}, enemy: 1, prefix: []string{"ready", "flags:4096", "parent:source", "class:owner", "class:target", "enemy:owner:target"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			state := &boomCollideTestState4E9770{
				ready: 1, quest: 1, parent: tc.parent, classes: tc.classes, enemy: tc.enemy,
			}
			if state.classes == nil {
				state.classes = map[string]uint8{}
			}
			boomCollide4E9770(tc.source, tc.target, "", state.hooks())
			if len(state.events) < len(tc.prefix) || !reflect.DeepEqual(state.events[:len(tc.prefix)], tc.prefix) {
				t.Fatalf("event prefix = %q, want %q (all %q)", state.events[:min(len(state.events), len(tc.prefix))], tc.prefix, state.events)
			}
		})
	}
}

func TestBoomCollide4E9770PlayerReflectionReturnsImmediately(t *testing.T) {
	tests := []struct {
		name         string
		inversion    int32
		enchant      int32
		directionHit int32
		want         []string
	}{
		{
			name:      "inversion equipment",
			inversion: 1,
			want: []string{
				"ready", "flags:4096", "point-fx:134:source", "class:target",
				"inversion:target:source", "change-owner:source:target",
			},
		},
		{
			name: "active inversion shield direction", enchant: 1, directionHit: 5,
			want: []string{
				"ready", "flags:4096", "point-fx:134:source", "class:target",
				"inversion:target:source", "enchant:target:27", "direction:target",
				"check-direction:target:-123:source", "change-owner:source:target", "audio:122:target:0:0",
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			state := &boomCollideTestState4E9770{
				ready: 1, classes: map[string]uint8{"target": 4}, inversion: tc.inversion,
				enchant: tc.enchant, direction: -123, directionHit: tc.directionHit,
			}
			boomCollide4E9770("source", "target", "collision", state.hooks())
			if !reflect.DeepEqual(state.events, tc.want) {
				t.Fatalf("events = %q, want %q", state.events, tc.want)
			}
		})
	}
}

func TestBoomCollide4E9770TargetDamageReloadsThenScorchesAndExplodes(t *testing.T) {
	state := &boomCollideTestState4E9770{
		ready: 1, direct: 37, splash: 19, radius: 60, pushRange: 70, force: 80,
		parent: "owner", classes: map[string]uint8{"target": 2},
	}
	boomCollide4E9770("source", "target", "ignored", state.hooks())
	want := []string{
		"ready", "flags:4096", "point-fx:134:source", "class:target",
		"load-direct", "parent:source", "damage:target:owner:source:37:7", "scorch:target:0",
		"load-splash", "load-range", "map-damage:source:60:5:19:7:source:",
		"load-force", "load-push-range", "map-push:source:70:70:80:source:0:0",
		"audio:84:source:0:0", "delete:source",
	}
	if !reflect.DeepEqual(state.events, want) {
		t.Fatalf("events = %q, want %q", state.events, want)
	}
}

func TestBoomCollide4E9770WallPathPreservesReloadOrder(t *testing.T) {
	state := &boomCollideTestState4E9770{
		ready: 1, direct: 91, classes: map[string]uint8{}, velocityX: 2, velocityY: 4,
		pointX: 123, pointY: -456,
	}
	hooks := state.hooks()
	hooks.wallReflect = func(collision, source string) {
		state.event("wall-reflect:" + collision + ":" + source)
		state.velocityX, state.velocityY = 10, 20
	}
	hooks.vectorDirection = func(x, y float32) int32 {
		state.event("vector-direction:" + strconv.FormatFloat(float64(x), 'g', -1, 32) + ":" + strconv.FormatFloat(float64(y), 'g', -1, 32))
		state.velocityX, state.velocityY = 30, 40
		return 0x12345
	}
	hooks.storeDirection2 = func(source string, value uint16) {
		state.event("store-direction:" + source + ":" + strconv.Itoa(int(value)))
		state.direction2 = value
		state.velocityX, state.velocityY = 99, 50
	}
	hooks.traceHitPoint = func() string {
		state.event("trace-point")
		return "hit"
	}
	hooks.loadPointY = func(point string) int32 {
		state.event("point-y:" + point)
		state.pointX = 789
		return state.pointY
	}

	boomCollide4E9770("source", "", "normal", hooks)
	want := []string{
		"ready", "flags:4096", "point-fx:134:source", "wall-reflect:normal:source",
		"velocity-x:source", "velocity-y:source", "vector-direction:10:20",
		"velocity-x:source", "store-direction:source:9029", "store-x:source:15",
		"velocity-y:source", "store-y:source:25", "trace-point", "load-direct",
		"point-y:hit", "point-x:hit", "damage-map:789:-456:91:7:source",
	}
	if !reflect.DeepEqual(state.events, want) {
		t.Fatalf("events = %q, want %q", state.events, want)
	}
	if state.direction2 != 0x2345 || state.velocityX != 15 || state.velocityY != 25 {
		t.Fatalf("direction/velocity = %#x (%g,%g), want 0x2345 (15,25)", state.direction2, state.velocityX, state.velocityY)
	}
}

func TestBoomCollide4E9770WallPathWithoutTracePointStillReturns(t *testing.T) {
	state := &boomCollideTestState4E9770{
		ready: 1, classes: map[string]uint8{}, velocityX: 1, velocityY: 0,
	}
	boomCollide4E9770("source", "", "normal", state.hooks())
	if got := state.events[len(state.events)-1]; got != "trace-point" {
		t.Fatalf("last event = %q, want trace-point (all %q)", got, state.events)
	}
	for _, event := range state.events {
		if event == "load-splash" || event == "delete:source" {
			t.Fatalf("wall path fell through into common explosion: %q", state.events)
		}
	}
}

func TestDirectionFromVector509ED0CardinalsAndSpecialValues(t *testing.T) {
	tests := []struct {
		name string
		x    float32
		y    float32
		want int32
	}{
		{name: "east", x: 1, y: 0, want: 0},
		{name: "south east", x: 1, y: 1, want: 32},
		{name: "south", x: 0, y: 1, want: 64},
		{name: "west", x: -1, y: 0, want: 128},
		{name: "north", x: 0, y: -1, want: 192},
		{name: "zero", x: 0, y: 0, want: 0},
		{name: "nan x", x: float32(math.NaN()), y: 1, want: 0},
		{name: "nan y", x: 1, y: float32(math.NaN()), want: 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := directionFromVector509ED0(tc.x, tc.y); got != tc.want {
				t.Fatalf("direction = %d, want %d", got, tc.want)
			}
			if got := DirFromVec(types.Ptf(tc.x, tc.y)); got != Dir16(tc.want) {
				t.Fatalf("DirFromVec = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestDirectionFromVector509ED0UsesExactConstantBits(t *testing.T) {
	if got := math.Float32bits(math.Float32frombits(boomDirectionHalfBits509ED0)); got != 0x3f000000 {
		t.Fatalf("half bits = %08x", got)
	}
	if got := math.Float32bits(math.Float32frombits(boomDirectionScaleBits509ED0)); got != 0x4222f983 {
		t.Fatalf("scale bits = %08x", got)
	}
	if got := math.Float32bits(math.Float32frombits(boomDirectionTauBits509ED0)); got != 0x40c90fdb {
		t.Fatalf("tau bits = %08x", got)
	}
}
