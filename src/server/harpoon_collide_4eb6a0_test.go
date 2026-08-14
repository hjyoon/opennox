package server

import (
	"fmt"
	"reflect"
	"testing"
)

type harpoonCollideTestObject4EB6A0 struct {
	name  string
	owner *harpoonCollideTestObject4EB6A0
	data  *harpoonCollideTestData4EB6A0
	flags uint32
	class uint8
	newX  float32
	newY  float32
	posX  float32
	posY  float32
}

type harpoonCollideTestData4EB6A0 struct {
	name   string
	target *harpoonCollideTestObject4EB6A0
	bolt   *harpoonCollideTestObject4EB6A0
	x      float32
	y      float32
	frame  uint32
}

type harpoonCollideTestFixture4EB6A0 struct {
	events       []string
	damage       int32
	balance      float32
	damageResult int32
	enemy        bool
	gameplay     bool
	frame        uint32
	onDamage     func()
	onSound      func()
	onDisable    func()
}

func harpoonObjectName4EB6A0(obj *harpoonCollideTestObject4EB6A0) string {
	if obj == nil {
		return "nil"
	}
	return obj.name
}

func harpoonDataName4EB6A0(data *harpoonCollideTestData4EB6A0) string {
	if data == nil {
		return "nil"
	}
	return data.name
}

func (f *harpoonCollideTestFixture4EB6A0) event(format string, args ...any) {
	f.events = append(f.events, fmt.Sprintf(format, args...))
}

func (f *harpoonCollideTestFixture4EB6A0) hooks() harpoonCollideHooks4EB6A0[
	*harpoonCollideTestObject4EB6A0,
	*harpoonCollideTestData4EB6A0,
] {
	return harpoonCollideHooks4EB6A0[
		*harpoonCollideTestObject4EB6A0,
		*harpoonCollideTestData4EB6A0,
	]{
		loadDamage: func() int32 {
			f.event("damage:load=%d", f.damage)
			return f.damage
		},
		loadOwner: func(obj *harpoonCollideTestObject4EB6A0) *harpoonCollideTestObject4EB6A0 {
			f.event("owner:%s=%s", harpoonObjectName4EB6A0(obj), harpoonObjectName4EB6A0(obj.owner))
			return obj.owner
		},
		loadPlayerData: func(obj *harpoonCollideTestObject4EB6A0) *harpoonCollideTestData4EB6A0 {
			f.event("data:%s=%s", harpoonObjectName4EB6A0(obj), harpoonDataName4EB6A0(obj.data))
			return obj.data
		},
		loadBalanceDamage: func() float32 {
			f.event("balance=%g", f.balance)
			return f.balance
		},
		floatToInt: func(value float32) int32 {
			f.event("convert=%g", value)
			return int32(value)
		},
		storeDamage: func(value int32) {
			f.event("damage:store=%d", value)
			f.damage = value
		},
		loadTargetFlags: func(obj *harpoonCollideTestObject4EB6A0) uint32 {
			f.event("targetFlags:%s=%#x", harpoonObjectName4EB6A0(obj), obj.flags)
			return obj.flags
		},
		findParentPlayer: func(obj *harpoonCollideTestObject4EB6A0) *harpoonCollideTestObject4EB6A0 {
			f.event("parent:%s", harpoonObjectName4EB6A0(obj))
			return obj.owner
		},
		targetDamage: func(target, parent, source *harpoonCollideTestObject4EB6A0, damage int32, damageType uint32) int32 {
			f.event("hit:%s:%s:%s:%d:%d", harpoonObjectName4EB6A0(target), harpoonObjectName4EB6A0(parent), harpoonObjectName4EB6A0(source), damage, damageType)
			if f.onDamage != nil {
				f.onDamage()
			}
			return f.damageResult
		},
		isEnemy: func(owner, target *harpoonCollideTestObject4EB6A0) bool {
			f.event("enemy:%s:%s=%t", harpoonObjectName4EB6A0(owner), harpoonObjectName4EB6A0(target), f.enemy)
			return f.enemy
		},
		gameplayFlag: func(flag uint32) bool {
			f.event("gameplay:%d=%t", flag, f.gameplay)
			return f.gameplay
		},
		loadClassLo: func(obj *harpoonCollideTestObject4EB6A0) uint8 {
			f.event("class:%s=%#x", harpoonObjectName4EB6A0(obj), obj.class)
			return obj.class
		},
		loadNewPosY: func(obj *harpoonCollideTestObject4EB6A0) float32 {
			f.event("newY:%s=%g", harpoonObjectName4EB6A0(obj), obj.newY)
			return obj.newY
		},
		loadNewPosX: func(obj *harpoonCollideTestObject4EB6A0) float32 {
			f.event("newX:%s=%g", harpoonObjectName4EB6A0(obj), obj.newX)
			return obj.newX
		},
		damageMap: func(x, y, damage int32, damageType uint32, source *harpoonCollideTestObject4EB6A0) {
			f.event("map:%d:%d:%d:%d:%s", x, y, damage, damageType, harpoonObjectName4EB6A0(source))
		},
		defaultDamageSound: func(target, source *harpoonCollideTestObject4EB6A0) {
			f.event("sound:%s:%s", harpoonObjectName4EB6A0(target), harpoonObjectName4EB6A0(source))
			if f.onSound != nil {
				f.onSound()
			}
		},
		storeTarget: func(data *harpoonCollideTestData4EB6A0, target *harpoonCollideTestObject4EB6A0) {
			f.event("target:%s=%s", harpoonDataName4EB6A0(data), harpoonObjectName4EB6A0(target))
			data.target = target
		},
		disableAbility: func(owner *harpoonCollideTestObject4EB6A0, ability int32) {
			f.event("disable:%s:%d", harpoonObjectName4EB6A0(owner), ability)
			if f.onDisable != nil {
				f.onDisable()
			}
		},
		delayedDelete: func(source *harpoonCollideTestObject4EB6A0) {
			f.event("delete:%s", harpoonObjectName4EB6A0(source))
		},
		storeBolt: func(data *harpoonCollideTestData4EB6A0, bolt *harpoonCollideTestObject4EB6A0) {
			f.event("bolt:%s=%s", harpoonDataName4EB6A0(data), harpoonObjectName4EB6A0(bolt))
			data.bolt = bolt
		},
		loadPosX: func(obj *harpoonCollideTestObject4EB6A0) float32 {
			f.event("posX:%s=%g", harpoonObjectName4EB6A0(obj), obj.posX)
			return obj.posX
		},
		loadPosY: func(obj *harpoonCollideTestObject4EB6A0) float32 {
			f.event("posY:%s=%g", harpoonObjectName4EB6A0(obj), obj.posY)
			return obj.posY
		},
		loadFrame: func() uint32 {
			f.event("frame=%d", f.frame)
			return f.frame
		},
		storeTargetX: func(data *harpoonCollideTestData4EB6A0, value float32) {
			f.event("storeX:%s=%g", harpoonDataName4EB6A0(data), value)
			data.x = value
		},
		storeTargetY: func(data *harpoonCollideTestData4EB6A0, value float32) {
			f.event("storeY:%s=%g", harpoonDataName4EB6A0(data), value)
			data.y = value
		},
		storeFrame: func(data *harpoonCollideTestData4EB6A0, value uint32) {
			f.event("storeFrame:%s=%d", harpoonDataName4EB6A0(data), value)
			data.frame = value
		},
		loadSourceFlags: func(obj *harpoonCollideTestObject4EB6A0) uint32 {
			f.event("sourceFlags:%s=%#x", harpoonObjectName4EB6A0(obj), obj.flags)
			return obj.flags
		},
		storeSourceFlags: func(obj *harpoonCollideTestObject4EB6A0, value uint32) {
			f.event("sourceFlags:store:%s=%#x", harpoonObjectName4EB6A0(obj), value)
			obj.flags = value
		},
		markRelation: func(owner, target *harpoonCollideTestObject4EB6A0) {
			f.event("relation:%s:%s", harpoonObjectName4EB6A0(owner), harpoonObjectName4EB6A0(target))
		},
		audio: func(sound uint32, owner *harpoonCollideTestObject4EB6A0) {
			f.event("audio:%d:%s", sound, harpoonObjectName4EB6A0(owner))
		},
	}
}

func newHarpoonCollideFixture4EB6A0() (
	*harpoonCollideTestFixture4EB6A0,
	*harpoonCollideTestObject4EB6A0,
	*harpoonCollideTestObject4EB6A0,
	*harpoonCollideTestObject4EB6A0,
	*harpoonCollideTestData4EB6A0,
) {
	data := &harpoonCollideTestData4EB6A0{name: "old-data"}
	owner := &harpoonCollideTestObject4EB6A0{name: "owner", data: data}
	source := &harpoonCollideTestObject4EB6A0{name: "source", owner: owner, newX: 69, newY: -46, flags: 0x08}
	target := &harpoonCollideTestObject4EB6A0{name: "target", class: harpoonUnitClassMask4EB6A0, posX: 12.5, posY: -4.25}
	f := &harpoonCollideTestFixture4EB6A0{balance: 7.75, damageResult: 1, enemy: true, frame: 1234}
	return f, source, target, owner, data
}

func assertHarpoonEvents4EB6A0(t *testing.T, got, want []string) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("event order mismatch\n got: %#v\nwant: %#v", got, want)
	}
}

func TestHarpoonCollide4EB6A0CachesEntryStateBeforeDamageInit(t *testing.T) {
	f, source, target, _, _ := newHarpoonCollideFixture4EB6A0()
	target.flags = harpoonTargetRejectFlags4EB6A0
	harpoonCollide4EB6A0(source, target, new(int), f.hooks())
	assertHarpoonEvents4EB6A0(t, f.events, []string{
		"damage:load=0", "owner:source=owner", "data:owner=old-data",
		"balance=7.75", "convert=7.75", "damage:store=7", "targetFlags:target=0x8020",
	})
}

func TestHarpoonCollide4EB6A0NonzeroDamageSkipsBalance(t *testing.T) {
	f, source, target, _, _ := newHarpoonCollideFixture4EB6A0()
	f.damage = -9
	target.flags = harpoonTargetRejectFlags4EB6A0
	harpoonCollide4EB6A0(source, target, nil, f.hooks())
	assertHarpoonEvents4EB6A0(t, f.events, []string{
		"damage:load=-9", "owner:source=owner", "data:owner=old-data", "targetFlags:target=0x8020",
	})
}

func TestHarpoonCollide4EB6A0WallUsesYBeforeXAndCachedCleanup(t *testing.T) {
	f, source, _, owner, data := newHarpoonCollideFixture4EB6A0()
	f.damage = 12
	replacement := &harpoonCollideTestData4EB6A0{name: "replacement"}
	f.onDisable = func() {
		owner.data = replacement
		data.bolt = source
	}
	harpoonCollide4EB6A0(source, (*harpoonCollideTestObject4EB6A0)(nil), &struct{}{}, f.hooks())
	assertHarpoonEvents4EB6A0(t, f.events, []string{
		"damage:load=12", "owner:source=owner", "data:owner=old-data",
		"newY:source=-46", "convert=-2", "newX:source=69", "convert=3",
		"map:3:-2:12:11:source", "target:old-data=nil", "disable:owner:3",
		"delete:source", "bolt:old-data=nil",
	})
	if data.bolt != nil || owner.data != replacement {
		t.Fatalf("cached cleanup mismatch: old=%+v replacement=%+v", data, replacement)
	}
}

func TestHarpoonCollide4EB6A0RejectedTargetsDoNotTouchData(t *testing.T) {
	for _, tc := range []struct {
		name  string
		flags uint32
		owner bool
	}{
		{name: "flags", flags: harpoonTargetRejectFlags4EB6A0},
		{name: "owner", owner: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f, source, target, _, _ := newHarpoonCollideFixture4EB6A0()
			f.damage = 4
			target.flags = tc.flags
			if tc.owner {
				target = source.owner
			}
			harpoonCollide4EB6A0(source, target, nil, f.hooks())
			want := []string{"damage:load=4", "owner:source=owner", "data:owner=old-data", fmt.Sprintf("targetFlags:%s=%#x", target.name, tc.flags)}
			assertHarpoonEvents4EB6A0(t, f.events, want)
		})
	}
}

func TestHarpoonCollide4EB6A0DamageUsesFullResultAndFailureOrder(t *testing.T) {
	for _, result := range []int32{0, 0x100} {
		t.Run(fmt.Sprintf("result_%#x", result), func(t *testing.T) {
			f, source, target, owner, data := newHarpoonCollideFixture4EB6A0()
			f.damage, f.damageResult = 15, result
			f.enemy = true
			replacement := &harpoonCollideTestData4EB6A0{name: "replacement"}
			f.onDamage = func() { owner.data = replacement }
			f.onSound = func() { source.owner = target }
			harpoonCollide4EB6A0(source, target, nil, f.hooks())
			if result == 0 {
				assertHarpoonEvents4EB6A0(t, f.events, []string{
					"damage:load=15", "owner:source=owner", "data:owner=old-data", "targetFlags:target=0x0",
					"parent:source", "hit:target:owner:source:15:11", "sound:target:source",
					"target:old-data=nil", "disable:owner:3", "delete:source", "bolt:old-data=nil",
				})
				return
			}
			if data.target != target {
				t.Fatalf("full nonzero damage result did not attach target: events=%#v", f.events)
			}
		})
	}
}

func TestHarpoonCollide4EB6A0PolicyShortCircuit(t *testing.T) {
	for _, tc := range []struct {
		name     string
		enemy    bool
		gameplay bool
		class    uint8
		wantTail []string
	}{
		{name: "enemy", enemy: true, wantTail: []string{"enemy:owner:target=true"}},
		{name: "gameplay off", wantTail: []string{"enemy:owner:target=false", "gameplay:1=false"}},
		{name: "non-unit", gameplay: true, wantTail: []string{"enemy:owner:target=false", "gameplay:1=true", "class:target=0x0"}},
		{name: "unit override", gameplay: true, class: harpoonUnitClassMask4EB6A0, wantTail: []string{"enemy:owner:target=false", "gameplay:1=true", "class:target=0x6"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f, source, target, _, _ := newHarpoonCollideFixture4EB6A0()
			f.damage, f.damageResult = 5, 1
			f.enemy, f.gameplay, target.class = tc.enemy, tc.gameplay, tc.class
			harpoonCollide4EB6A0(source, target, nil, f.hooks())
			start := 6
			if len(f.events) < start+len(tc.wantTail) {
				t.Fatalf("events too short: %#v", f.events)
			}
			assertHarpoonEvents4EB6A0(t, f.events[start:start+len(tc.wantTail)], tc.wantTail)
		})
	}
}

func TestHarpoonCollide4EB6A0SuccessUsesCachedAndLiveOwners(t *testing.T) {
	f, source, target, owner, data := newHarpoonCollideFixture4EB6A0()
	f.damage, f.damageResult, f.enemy = 22, 1, true
	liveOwner := &harpoonCollideTestObject4EB6A0{name: "live-owner"}
	f.onDamage = func() {
		target.posX = 19.5
		target.posY = -8.75
		f.frame = 9876
		source.flags = 0x21
		source.owner = liveOwner
		owner.data = &harpoonCollideTestData4EB6A0{name: "replacement"}
	}
	harpoonCollide4EB6A0(source, target, &struct{}{}, f.hooks())
	assertHarpoonEvents4EB6A0(t, f.events, []string{
		"damage:load=22", "owner:source=owner", "data:owner=old-data", "targetFlags:target=0x0",
		"parent:source", "hit:target:owner:source:22:11", "enemy:owner:target=true",
		"target:old-data=target", "posX:target=19.5", "storeX:old-data=19.5",
		"posY:target=-8.75", "storeY:old-data=-8.75", "frame=9876", "storeFrame:old-data=9876",
		"sourceFlags:source=0x21", "owner:source=live-owner", "sourceFlags:store:source=0x61",
		"relation:live-owner:target", "audio:999:owner",
	})
	if data.target != target || data.x != 19.5 || data.y != -8.75 || data.frame != 9876 || source.flags != 0x61 {
		t.Fatalf("success state mismatch: data=%+v source=%+v", data, source)
	}
}
