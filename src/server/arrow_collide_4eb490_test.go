package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

type arrowCollideTestObject4EB490 struct {
	name      string
	typeIndex uint16
	flags     uint32
	classLo   uint8
	owner     *arrowCollideTestObject4EB490
	data      *arrowCollideTestData4EB490
	health    *arrowCollideTestHealth4EB490
	posX      float32
	posY      float32
	radius    float32
}

type arrowCollideTestData4EB490 struct {
	name  string
	owner *arrowCollideTestObject4EB490
}

type arrowCollideTestModifier4EB490 struct{ name string }

type arrowCollideTestHealth4EB490 struct {
	cur uint16
	max uint16
}

type arrowCollideTestFixture4EB490 struct {
	events          []string
	projectileClass *arrowCollideTestModifier4EB490
	quest           bool
	parent          *arrowCollideTestObject4EB490
	enemy           bool
	traceX          int32
	traceY          int32
	traceOK         bool
	calcDamage      float64
	archerBoltType  uint32
	lookupType      uint32
	damageResult    int32
	attack          arrowAttack4EB490[*arrowCollideTestObject4EB490]
	onGameFlag      func()
	onApply         func(*arrowAttack4EB490[*arrowCollideTestObject4EB490])
	onPre           func(*arrowAttack4EB490[*arrowCollideTestObject4EB490])
	onConvert       func()
	onDamage        func()
}

func arrowObjectName4EB490(obj *arrowCollideTestObject4EB490) string {
	if obj == nil {
		return "nil"
	}
	return obj.name
}

func arrowDataName4EB490(data *arrowCollideTestData4EB490) string {
	if data == nil {
		return "nil"
	}
	return data.name
}

func (f *arrowCollideTestFixture4EB490) event(format string, args ...any) {
	f.events = append(f.events, fmt.Sprintf(format, args...))
}

func (f *arrowCollideTestFixture4EB490) hooks() arrowCollideHooks4EB490[
	*arrowCollideTestObject4EB490,
	*arrowCollideTestData4EB490,
	*arrowCollideTestModifier4EB490,
	*arrowCollideTestHealth4EB490,
] {
	return arrowCollideHooks4EB490[
		*arrowCollideTestObject4EB490,
		*arrowCollideTestData4EB490,
		*arrowCollideTestModifier4EB490,
		*arrowCollideTestHealth4EB490,
	]{
		loadTypeIndex: func(obj *arrowCollideTestObject4EB490) uint16 {
			f.event("type:%s=%d", arrowObjectName4EB490(obj), obj.typeIndex)
			return obj.typeIndex
		},
		loadCollideData: func(obj *arrowCollideTestObject4EB490) *arrowCollideTestData4EB490 {
			f.event("data:%s=%s", arrowObjectName4EB490(obj), arrowDataName4EB490(obj.data))
			return obj.data
		},
		lookupProjectileClass: func(index uint16) *arrowCollideTestModifier4EB490 {
			f.event("projectile:%d", index)
			return f.projectileClass
		},
		loadOwner: func(obj *arrowCollideTestObject4EB490) *arrowCollideTestObject4EB490 {
			f.event("owner:%s=%s", arrowObjectName4EB490(obj), arrowObjectName4EB490(obj.owner))
			return obj.owner
		},
		strength: func(obj *arrowCollideTestObject4EB490) int32 {
			f.event("strength:%s", arrowObjectName4EB490(obj))
			if obj == nil {
				return -1
			}
			return 40
		},
		gameFlag: func(flag uint32) bool {
			f.event("gameFlag:%#x", flag)
			if f.onGameFlag != nil {
				f.onGameFlag()
			}
			return f.quest
		},
		findParentPlayer: func(obj *arrowCollideTestObject4EB490) *arrowCollideTestObject4EB490 {
			f.event("parent:%s=%s", arrowObjectName4EB490(obj), arrowObjectName4EB490(f.parent))
			return f.parent
		},
		loadClassLo: func(obj *arrowCollideTestObject4EB490) uint8 {
			f.event("class:%s=%#x", arrowObjectName4EB490(obj), obj.classLo)
			return obj.classLo
		},
		isEnemy: func(parent, target *arrowCollideTestObject4EB490) bool {
			f.event("enemy:%s:%s=%t", arrowObjectName4EB490(parent), arrowObjectName4EB490(target), f.enemy)
			return f.enemy
		},
		tracePoint: func() (int32, int32, bool) {
			f.event("trace=%d:%d:%t", f.traceX, f.traceY, f.traceOK)
			return f.traceX, f.traceY, f.traceOK
		},
		calcBoltDamage: func(strength int32, modifier *arrowCollideTestModifier4EB490) float64 {
			f.event("calc:%d:%s", strength, modifier.name)
			return f.calcDamage
		},
		floatToInt: func(value float64) int32 {
			f.event("convert:%g", value)
			if f.onConvert != nil {
				f.onConvert()
			}
			return int32(value)
		},
		damageMap: func(x, y, damage int32, damageType uint32, source *arrowCollideTestObject4EB490) {
			f.event("map:%d:%d:%d:%d:%s", x, y, damage, damageType, arrowObjectName4EB490(source))
		},
		delayedDelete: func(source *arrowCollideTestObject4EB490) {
			f.event("delete:%s", arrowObjectName4EB490(source))
		},
		loadArcherBoltType: func() uint32 {
			f.event("archer:load=%d", f.archerBoltType)
			return f.archerBoltType
		},
		lookupType: func(name string) uint32 {
			f.event("typeLookup:%s", name)
			return f.lookupType
		},
		storeArcherBoltType: func(value uint32) {
			f.event("archer:store=%d", value)
			f.archerBoltType = value
		},
		loadFlags: func(obj *arrowCollideTestObject4EB490) uint32 {
			f.event("flags:%s=%#x", arrowObjectName4EB490(obj), obj.flags)
			return obj.flags
		},
		loadPosX: func(obj *arrowCollideTestObject4EB490) float32 {
			f.event("posX:%s=%g", arrowObjectName4EB490(obj), obj.posX)
			return obj.posX
		},
		loadPosY: func(obj *arrowCollideTestObject4EB490) float32 {
			f.event("posY:%s=%g", arrowObjectName4EB490(obj), obj.posY)
			return obj.posY
		},
		loadRadius: func(obj *arrowCollideTestObject4EB490) float32 {
			f.event("radius:%s=%g", arrowObjectName4EB490(obj), obj.radius)
			return obj.radius
		},
		loadDataOwner: func(data *arrowCollideTestData4EB490) *arrowCollideTestObject4EB490 {
			f.event("dataOwner:%s=%s", arrowDataName4EB490(data), arrowObjectName4EB490(data.owner))
			return data.owner
		},
		applyAttackEffect: func(source, owner *arrowCollideTestObject4EB490, attack *arrowAttack4EB490[*arrowCollideTestObject4EB490]) {
			f.event("apply:%s:%s", arrowObjectName4EB490(source), arrowObjectName4EB490(owner))
			f.attack = *attack
			if f.onApply != nil {
				f.onApply(attack)
			}
		},
		preAttackEffects: func(target, owner, source *arrowCollideTestObject4EB490, attack *arrowAttack4EB490[*arrowCollideTestObject4EB490]) {
			f.event("pre:%s:%s:%s", arrowObjectName4EB490(target), arrowObjectName4EB490(owner), arrowObjectName4EB490(source))
			if f.onPre != nil {
				f.onPre(attack)
			}
		},
		targetDamage: func(target, parent, source *arrowCollideTestObject4EB490, damage int32, damageType uint32) int32 {
			f.event("damage:%s:%s:%s:%d:%d", arrowObjectName4EB490(target), arrowObjectName4EB490(parent), arrowObjectName4EB490(source), damage, damageType)
			if f.onDamage != nil {
				f.onDamage()
			}
			return f.damageResult
		},
		loadHealth: func(obj *arrowCollideTestObject4EB490) *arrowCollideTestHealth4EB490 {
			f.event("health:%s", arrowObjectName4EB490(obj))
			return obj.health
		},
		loadHealthCur: func(health *arrowCollideTestHealth4EB490) uint16 {
			f.event("healthCur=%d", health.cur)
			return health.cur
		},
		loadHealthMax: func(health *arrowCollideTestHealth4EB490) uint16 {
			f.event("healthMax=%d", health.max)
			return health.max
		},
	}
}

func newArrowCollideFixture4EB490() (
	*arrowCollideTestFixture4EB490,
	*arrowCollideTestObject4EB490,
	*arrowCollideTestObject4EB490,
) {
	owner := &arrowCollideTestObject4EB490{name: "owner"}
	dataOwner := &arrowCollideTestObject4EB490{name: "data-owner"}
	source := &arrowCollideTestObject4EB490{
		name: "source", typeIndex: 7, owner: owner,
		data: &arrowCollideTestData4EB490{name: "old-data", owner: dataOwner},
		posX: 12.5, posY: -4.25, radius: 6.75,
	}
	target := &arrowCollideTestObject4EB490{name: "target"}
	fixture := &arrowCollideTestFixture4EB490{
		projectileClass: &arrowCollideTestModifier4EB490{name: "projectile"},
		calcDamage:      12.75,
		archerBoltType:  99,
		lookupType:      99,
		damageResult:    1,
	}
	return fixture, source, target
}

func assertArrowEvents4EB490(t *testing.T, got, want []string) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("event order mismatch\n got: %#v\nwant: %#v", got, want)
	}
}

func arrowEventIndex4EB490(t *testing.T, events []string, value string) int {
	t.Helper()
	for i, event := range events {
		if event == value {
			return i
		}
	}
	t.Fatalf("event %q missing from %#v", value, events)
	return -1
}

func TestArrowCollide4EB490CachesDataBeforeClassLookup(t *testing.T) {
	f, source, target := newArrowCollideFixture4EB490()
	f.projectileClass = nil
	arrowCollide4EB490(source, target, &struct{}{}, f.hooks())
	assertArrowEvents4EB490(t, f.events, []string{
		"type:source=7", "data:source=old-data", "projectile:7",
	})
}

func TestArrowCollide4EB490OwnerTargetEarlyReturn(t *testing.T) {
	f, source, target := newArrowCollideFixture4EB490()
	source.owner = target
	arrowCollide4EB490(source, target, nil, f.hooks())
	assertArrowEvents4EB490(t, f.events, []string{
		"type:source=7", "data:source=old-data", "projectile:7", "owner:source=target",
	})
}

func TestArrowCollide4EB490QuestFriendlyPlayerGate(t *testing.T) {
	f, source, target := newArrowCollideFixture4EB490()
	parent := &arrowCollideTestObject4EB490{name: "parent", classLo: arrowPlayerClassBit4EB490}
	target.classLo = arrowPlayerClassBit4EB490
	f.quest = true
	f.parent = parent
	f.enemy = false
	arrowCollide4EB490(source, target, nil, f.hooks())
	assertArrowEvents4EB490(t, f.events, []string{
		"type:source=7", "data:source=old-data", "projectile:7", "owner:source=owner",
		"strength:owner", "gameFlag:0x1000", "parent:source=parent", "class:parent=0x4",
		"class:target=0x4", "enemy:parent:target=false",
	})
}

func TestArrowCollide4EB490WallImpact(t *testing.T) {
	for _, tc := range []struct {
		name    string
		owner   bool
		traceOK bool
		want    []string
	}{
		{
			name: "trace hit with owner strength", owner: true, traceOK: true,
			want: []string{
				"type:source=7", "data:source=old-data", "projectile:7", "owner:source=owner",
				"strength:owner", "gameFlag:0x1000", "trace=21:-8:true", "calc:40:projectile",
				"convert:12.75", "map:21:-8:12:11:source", "delete:source",
			},
		},
		{
			name: "trace miss with default strength", owner: false, traceOK: false,
			want: []string{
				"type:source=7", "data:source=old-data", "projectile:7", "owner:source=nil",
				"gameFlag:0x1000", "trace=21:-8:false", "delete:source",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f, source, _ := newArrowCollideFixture4EB490()
			if !tc.owner {
				source.owner = nil
			}
			f.traceX, f.traceY, f.traceOK = 21, -8, tc.traceOK
			arrowCollide4EB490(source, (*arrowCollideTestObject4EB490)(nil), new(int), f.hooks())
			assertArrowEvents4EB490(t, f.events, tc.want)
		})
	}
}

func TestArrowCollide4EB490SecondStrengthAndLazyTypePrecedeFlags(t *testing.T) {
	f, source, target := newArrowCollideFixture4EB490()
	liveOwner := &arrowCollideTestObject4EB490{name: "live-owner"}
	f.archerBoltType = 0
	target.flags = arrowUntargetableFlag4EB490
	f.onGameFlag = func() { source.owner = liveOwner }
	arrowCollide4EB490(source, target, nil, f.hooks())
	assertArrowEvents4EB490(t, f.events, []string{
		"type:source=7", "data:source=old-data", "projectile:7", "owner:source=owner",
		"strength:owner", "gameFlag:0x1000", "owner:source=live-owner", "strength:live-owner",
		"archer:load=0", "typeLookup:ArcherBolt", "archer:store=99", "flags:target=0x8000",
	})
}

func TestArrowCollide4EB490AttackUsesCachedDataAndLiveReloads(t *testing.T) {
	f, source, target := newArrowCollideFixture4EB490()
	oldData := source.data
	initialDataOwner := oldData.owner
	liveDataOwner := &arrowCollideTestObject4EB490{name: "live-data-owner"}
	replacementDataOwner := &arrowCollideTestObject4EB490{name: "replacement-data-owner"}
	liveOwner := &arrowCollideTestObject4EB490{name: "live-owner"}
	convertedParent := &arrowCollideTestObject4EB490{name: "converted-parent"}
	f.onApply = func(attack *arrowAttack4EB490[*arrowCollideTestObject4EB490]) {
		oldData.owner = liveDataOwner
		source.data = &arrowCollideTestData4EB490{name: "replacement-data", owner: replacementDataOwner}
		source.owner = liveOwner
		attack.Damage = 9.75
		attack.DamageType = 13
	}
	f.onPre = func(attack *arrowAttack4EB490[*arrowCollideTestObject4EB490]) {
		attack.Radius = 17
	}
	f.onConvert = func() { f.parent = convertedParent }
	arrowCollide4EB490(source, target, &struct{}{}, f.hooks())

	wantAttack := arrowAttack4EB490[*arrowCollideTestObject4EB490]{
		Damage: 12.75, DamageType: arrowDamageType4EB490, Radius: 6.75,
		Owner: initialDataOwner, PosX: 12.5, PosY: -4.25, Source: source,
	}
	if !reflect.DeepEqual(f.attack, wantAttack) {
		t.Fatalf("initial attack mismatch\n got: %#v\nwant: %#v", f.attack, wantAttack)
	}
	assertArrowEvents4EB490(t, f.events, []string{
		"type:source=7", "data:source=old-data", "projectile:7", "owner:source=owner",
		"strength:owner", "gameFlag:0x1000", "owner:source=owner", "strength:owner",
		"archer:load=99", "flags:target=0x0", "posX:source=12.5", "posY:source=-4.25",
		"calc:40:projectile", "radius:source=6.75", "dataOwner:old-data=data-owner",
		"apply:source:data-owner", "owner:source=live-owner", "dataOwner:old-data=live-data-owner",
		"pre:target:live-data-owner:source", "convert:10.25", "parent:source=converted-parent",
		"damage:target:converted-parent:source:10:13", "archer:load=99", "type:source=7",
		"delete:source",
	})
}

func TestArrowCollide4EB490ApplyCanSuppressPreAttackWithLiveOwner(t *testing.T) {
	f, source, target := newArrowCollideFixture4EB490()
	f.damageResult = 0
	f.onApply = func(*arrowAttack4EB490[*arrowCollideTestObject4EB490]) { source.owner = target }
	arrowCollide4EB490(source, target, nil, f.hooks())
	for _, event := range f.events {
		if len(event) >= 4 && event[:4] == "pre:" {
			t.Fatalf("pre-attack unexpectedly called: %#v", f.events)
		}
	}
	ownerIndex := arrowEventIndex4EB490(t, f.events, "owner:source=target")
	convertIndex := arrowEventIndex4EB490(t, f.events, "convert:13.25")
	if ownerIndex >= convertIndex {
		t.Fatalf("live owner check did not precede conversion: %#v", f.events)
	}
	for _, event := range f.events {
		if event == "delete:source" {
			t.Fatalf("AL-zero non-Archer result unexpectedly deleted source: %#v", f.events)
		}
	}
}

func TestArrowCollide4EB490UsesOnlyDamageResultLowByte(t *testing.T) {
	for _, tc := range []struct {
		result     int32
		wantDelete bool
	}{
		{result: 0x100, wantDelete: false},
		{result: 0x101, wantDelete: true},
	} {
		t.Run(fmt.Sprintf("result_%#x", tc.result), func(t *testing.T) {
			f, source, target := newArrowCollideFixture4EB490()
			f.damageResult = tc.result
			arrowCollide4EB490(source, target, nil, f.hooks())
			gotDelete := false
			for _, event := range f.events {
				gotDelete = gotDelete || event == "delete:source"
			}
			if gotDelete != tc.wantDelete {
				t.Fatalf("delete=%t, want %t; events=%#v", gotDelete, tc.wantDelete, f.events)
			}
		})
	}
}

func TestArrowCollide4EB490ArcherHealthPolicyUsesPostDamageState(t *testing.T) {
	for _, tc := range []struct {
		name        string
		health      *arrowCollideTestHealth4EB490
		wantTail    []string
		wantDeleted bool
	}{
		{name: "nil health", wantTail: []string{"health:target", "delete:source"}, wantDeleted: true},
		{name: "alive", health: &arrowCollideTestHealth4EB490{cur: 1, max: 10}, wantTail: []string{"health:target", "healthCur=1", "delete:source"}, wantDeleted: true},
		{name: "dead valid", health: &arrowCollideTestHealth4EB490{cur: 0, max: 10}, wantTail: []string{"health:target", "healthCur=0", "healthMax=10"}},
		{name: "zeroed", health: &arrowCollideTestHealth4EB490{}, wantTail: []string{"health:target", "healthCur=0", "healthMax=0", "delete:source"}, wantDeleted: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f, source, target := newArrowCollideFixture4EB490()
			f.damageResult = 0x100
			f.onDamage = func() {
				source.typeIndex = uint16(f.archerBoltType)
				target.health = tc.health
			}
			arrowCollide4EB490(source, target, nil, f.hooks())
			if len(f.events) < len(tc.wantTail) {
				t.Fatalf("events too short: %#v", f.events)
			}
			assertArrowEvents4EB490(t, f.events[len(f.events)-len(tc.wantTail):], tc.wantTail)
			gotDeleted := f.events[len(f.events)-1] == "delete:source"
			if gotDeleted != tc.wantDeleted {
				t.Fatalf("delete=%t, want %t; events=%#v", gotDeleted, tc.wantDeleted, f.events)
			}
		})
	}
}

func TestArrowCollide4EB490RoundsStoredFloat32Damage(t *testing.T) {
	f, source, target := newArrowCollideFixture4EB490()
	f.calcDamage = math.Float64frombits(0x4028ffffffffffff)
	f.damageResult = 0
	arrowCollide4EB490(source, target, nil, f.hooks())
	stored := float32(f.calcDamage)
	want := fmt.Sprintf("convert:%g", float64(stored)+0.5)
	arrowEventIndex4EB490(t, f.events, want)
	if f.attack.Damage != stored {
		t.Fatalf("stored damage bits=%#x, want %#x", math.Float32bits(f.attack.Damage), math.Float32bits(stored))
	}
}
