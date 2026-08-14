package server

import (
	"fmt"
	"reflect"
	"testing"
)

type chakramCollideTestObject4EAF00 struct {
	name      string
	flags     uint32
	material  uint8
	owner     *chakramCollideTestObject4EAF00
	class     uint8
	hasWeapon bool
	typeIndex uint16
	posX      float32
	posY      float32
	radius    float32
	firstItem *chakramCollideTestObject4EAF00
}

type chakramCollideTestUpdate4EAF00 struct {
	reflections uint8
	state       uint8
	returnTo    *chakramCollideTestObject4EAF00
	lastHit     *chakramCollideTestObject4EAF00
}

type chakramCollideTestModifier4EAF00 struct{ name string }

type chakramCollideTestFixture4EAF00 struct {
	events          []string
	update          *chakramCollideTestUpdate4EAF00
	ownerReads      []*chakramCollideTestObject4EAF00
	inventoryReads  []*chakramCollideTestObject4EAF00
	traceX          int32
	traceY          int32
	traceOK         bool
	projectileClass *chakramCollideTestModifier4EAF00
	attack          chakramAttack4EAF00[*chakramCollideTestObject4EAF00]
	roundInput      float64
	damage          int32
	damageTarget    *chakramCollideTestObject4EAF00
	onApply         func(*chakramAttack4EAF00[*chakramCollideTestObject4EAF00])
	onPre           func(*chakramAttack4EAF00[*chakramCollideTestObject4EAF00])
	onDamage        func(*chakramCollideTestObject4EAF00)
}

func (f *chakramCollideTestFixture4EAF00) event(format string, args ...any) {
	f.events = append(f.events, fmt.Sprintf(format, args...))
}

func (f *chakramCollideTestFixture4EAF00) nextOwner(obj *chakramCollideTestObject4EAF00) *chakramCollideTestObject4EAF00 {
	if len(f.ownerReads) == 0 {
		return obj.owner
	}
	owner := f.ownerReads[0]
	f.ownerReads = f.ownerReads[1:]
	return owner
}

func (f *chakramCollideTestFixture4EAF00) nextItem(obj *chakramCollideTestObject4EAF00) *chakramCollideTestObject4EAF00 {
	if len(f.inventoryReads) == 0 {
		return obj.firstItem
	}
	item := f.inventoryReads[0]
	f.inventoryReads = f.inventoryReads[1:]
	return item
}

func (f *chakramCollideTestFixture4EAF00) hooks() chakramCollideHooks4EAF00[
	*chakramCollideTestObject4EAF00,
	*chakramCollideTestUpdate4EAF00,
	string,
	string,
	*int,
	*chakramCollideTestModifier4EAF00,
] {
	return chakramCollideHooks4EAF00[
		*chakramCollideTestObject4EAF00,
		*chakramCollideTestUpdate4EAF00,
		string,
		string,
		*int,
		*chakramCollideTestModifier4EAF00,
	]{
		loadUpdateData: func(obj *chakramCollideTestObject4EAF00) *chakramCollideTestUpdate4EAF00 {
			f.event("update:%s", obj.name)
			return f.update
		},
		inventoryFirst: func(obj *chakramCollideTestObject4EAF00) *chakramCollideTestObject4EAF00 {
			f.event("inventory:%s", obj.name)
			return f.nextItem(obj)
		},
		loadFlags: func(obj *chakramCollideTestObject4EAF00) uint32 {
			f.event("flags:%s", obj.name)
			return obj.flags
		},
		loadMaterialLo: func(obj *chakramCollideTestObject4EAF00) uint8 {
			f.event("material:%s", obj.name)
			return obj.material
		},
		loadOwner: func(obj *chakramCollideTestObject4EAF00) *chakramCollideTestObject4EAF00 {
			f.event("owner:%s", obj.name)
			return f.nextOwner(obj)
		},
		loadClassLo: func(obj *chakramCollideTestObject4EAF00) uint8 {
			f.event("class:%s", obj.name)
			return obj.class
		},
		ownerHasWeapon: func(obj *chakramCollideTestObject4EAF00) bool {
			f.event("has-weapon:%s", obj.name)
			return obj.hasWeapon
		},
		loadTypeIndex: func(obj *chakramCollideTestObject4EAF00) uint16 {
			f.event("type:%s", obj.name)
			return obj.typeIndex
		},
		loadPosX: func(obj *chakramCollideTestObject4EAF00) float32 {
			f.event("pos-x:%s", obj.name)
			return obj.posX
		},
		loadPosY: func(obj *chakramCollideTestObject4EAF00) float32 {
			f.event("pos-y:%s", obj.name)
			return obj.posY
		},
		loadRadius: func(obj *chakramCollideTestObject4EAF00) float32 {
			f.event("radius:%s", obj.name)
			return obj.radius
		},
		position: func(obj *chakramCollideTestObject4EAF00) string {
			f.event("position:%s", obj.name)
			return "pos(" + obj.name + ")"
		},
		velocity: func(obj *chakramCollideTestObject4EAF00) string {
			f.event("velocity:%s", obj.name)
			return "vel(" + obj.name + ")"
		},
		loadReflections: func(update *chakramCollideTestUpdate4EAF00) uint8 {
			f.event("reflections:%d", update.reflections)
			return update.reflections
		},
		storeReflections: func(update *chakramCollideTestUpdate4EAF00, value uint8) {
			f.event("store-reflections:%d", value)
			update.reflections = value
		},
		loadReturnState: func(update *chakramCollideTestUpdate4EAF00) uint8 {
			f.event("state:%d", update.state)
			return update.state
		},
		storeReturnState: func(update *chakramCollideTestUpdate4EAF00, value uint8) {
			f.event("store-state:%d", value)
			update.state = value
		},
		storeReturnTarget: func(update *chakramCollideTestUpdate4EAF00, obj *chakramCollideTestObject4EAF00) {
			f.event("store-return:%s", obj.name)
			update.returnTo = obj
		},
		storeLastHit: func(update *chakramCollideTestUpdate4EAF00, obj *chakramCollideTestObject4EAF00) {
			f.event("store-last:%s", obj.name)
			update.lastHit = obj
		},
		pointFX: func(id uint32, pos string) { f.event("fx:%d:%s", id, pos) },
		wallReflect: func(_ *int, velocity string) {
			f.event("wall-reflect:%s", velocity)
		},
		randomReflect: func(obj *chakramCollideTestObject4EAF00) {
			f.event("random-reflect:%s", obj.name)
		},
		tracePoint: func() (int32, int32, bool) {
			f.event("trace")
			return f.traceX, f.traceY, f.traceOK
		},
		damageMap: func(x, y, damage int32, typ uint32, obj *chakramCollideTestObject4EAF00) {
			f.event("map-damage:%d:%d:%d:%d:%s", x, y, damage, typ, obj.name)
		},
		drop: func(owner, item *chakramCollideTestObject4EAF00, pos string) {
			f.event("drop:%s:%s:%s", owner.name, item.name, pos)
		},
		delayedDelete: func(obj *chakramCollideTestObject4EAF00) {
			f.event("delete:%s", obj.name)
		},
		retarget: func(obj *chakramCollideTestObject4EAF00) { f.event("retarget:%s", obj.name) },
		detach: func(owner, item *chakramCollideTestObject4EAF00) {
			f.event("detach:%s:%s", owner.name, item.name)
		},
		inventoryPut: func(owner, item *chakramCollideTestObject4EAF00, mode uint32) {
			f.event("put:%s:%s:%d", owner.name, item.name, mode)
		},
		equipWeapon: func(owner, item *chakramCollideTestObject4EAF00, a3, a4 uint32) {
			f.event("equip:%s:%s:%d:%d", owner.name, item.name, a3, a4)
		},
		audio: func(id uint32, obj *chakramCollideTestObject4EAF00) {
			f.event("audio:%d:%s", id, obj.name)
		},
		sameTeam: func(source, target *chakramCollideTestObject4EAF00) bool {
			f.event("same-team:%s:%s", source.name, target.name)
			return false
		},
		lookupProjectileClass: func(index uint16) *chakramCollideTestModifier4EAF00 {
			f.event("projectile:%d", index)
			return f.projectileClass
		},
		strength: func(obj *chakramCollideTestObject4EAF00) int32 {
			f.event("strength:%s", obj.name)
			return 37
		},
		calcBoltDamage: func(strength int32, modifier *chakramCollideTestModifier4EAF00) float32 {
			f.event("calc:%d:%s", strength, modifier.name)
			return 7.25
		},
		applyAttackEffect: func(source, owner *chakramCollideTestObject4EAF00, attack *chakramAttack4EAF00[*chakramCollideTestObject4EAF00]) {
			f.event("apply:%s:%s", source.name, owner.name)
			if f.onApply != nil {
				f.onApply(attack)
			}
		},
		preAttackEffects: func(target, owner, source *chakramCollideTestObject4EAF00, attack *chakramAttack4EAF00[*chakramCollideTestObject4EAF00]) {
			f.event("pre:%s:%s:%s", target.name, owner.name, source.name)
			if f.onPre != nil {
				f.onPre(attack)
			}
		},
		floatToInt: func(value float64) int32 {
			f.event("round:%.2f", value)
			f.roundInput = value
			return 10
		},
		targetDamage: func(target, owner, source *chakramCollideTestObject4EAF00, damage int32, typ uint32) {
			f.event("damage:%s:%s:%s:%d:%d", target.name, owner.name, source.name, damage, typ)
			f.damage = damage
			f.damageTarget = target
			if f.onDamage != nil {
				f.onDamage(target)
			}
		},
		projectileReflect: func(source, target *chakramCollideTestObject4EAF00) {
			f.event("projectile-reflect:%s:%s", source.name, target.name)
		},
		createAt: func(item, owner *chakramCollideTestObject4EAF00, pos string) {
			ownerName := "nil"
			if owner != nil {
				ownerName = owner.name
			}
			f.event("create:%s:%s:%s", item.name, ownerName, pos)
		},
	}
}

func assertChakramEvents4EAF00(t *testing.T, got, want []string) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("events mismatch\n got: %q\nwant: %q", got, want)
	}
}

func newChakramFixture4EAF00() (
	*chakramCollideTestFixture4EAF00,
	*chakramCollideTestObject4EAF00,
	*chakramCollideTestObject4EAF00,
	*chakramCollideTestObject4EAF00,
) {
	owner := &chakramCollideTestObject4EAF00{name: "owner", class: chakramPlayerClassBit4EAF00}
	item := &chakramCollideTestObject4EAF00{name: "item"}
	source := &chakramCollideTestObject4EAF00{
		name: "source", owner: owner, firstItem: item, typeIndex: 77,
		posX: 12.5, posY: -4.25, radius: 6,
	}
	target := &chakramCollideTestObject4EAF00{name: "target"}
	fixture := &chakramCollideTestFixture4EAF00{
		update:          &chakramCollideTestUpdate4EAF00{},
		projectileClass: &chakramCollideTestModifier4EAF00{name: "class"},
		traceX:          31, traceY: 47, traceOK: true,
	}
	return fixture, source, target, item
}

func TestChakramCollide4EAF00EntryInventoryAndNilTargetOrder(t *testing.T) {
	t.Run("missing inventory", func(t *testing.T) {
		f, source, target, _ := newChakramFixture4EAF00()
		source.firstItem = nil
		chakramCollide4EAF00(source, target, (*int)(nil), f.hooks())
		assertChakramEvents4EAF00(t, f.events, []string{
			"update:source", "inventory:source", "delete:source",
		})
	})

	t.Run("destroyed inventory", func(t *testing.T) {
		f, source, target, item := newChakramFixture4EAF00()
		item.flags = chakramDestroyedFlag4EAF00
		chakramCollide4EAF00(source, target, (*int)(nil), f.hooks())
		assertChakramEvents4EAF00(t, f.events, []string{
			"update:source", "inventory:source", "flags:item", "delete:source",
		})
	})

	t.Run("nil target and collision", func(t *testing.T) {
		f, source, _, _ := newChakramFixture4EAF00()
		chakramCollide4EAF00(source, (*chakramCollideTestObject4EAF00)(nil), (*int)(nil), f.hooks())
		assertChakramEvents4EAF00(t, f.events, []string{
			"update:source", "inventory:source", "flags:item", "position:source", "fx:150:pos(source)",
		})
	})
}

func TestChakramCollide4EAF00TargetMaterialFXPrecedesOwnerAndTeamGates(t *testing.T) {
	f, source, target, _ := newChakramFixture4EAF00()
	target.material = chakramMaterialFXMask4EAF00
	hooks := f.hooks()
	hooks.sameTeam = func(source, target *chakramCollideTestObject4EAF00) bool {
		f.event("same-team:%s:%s", source.name, target.name)
		return true
	}
	chakramCollide4EAF00(source, target, (*int)(nil), hooks)
	assertChakramEvents4EAF00(t, f.events, []string{
		"update:source", "inventory:source", "flags:item", "material:target",
		"position:source", "fx:150:pos(source)", "owner:source", "same-team:source:target",
	})
}

func TestChakramCollide4EAF00WallStateMachine(t *testing.T) {
	collision := 1
	tests := []struct {
		name        string
		reflections uint8
		state       uint8
		wantState   uint8
		wantReturn  bool
		wantTail    []string
	}{
		{
			name: "last reflected flight returns to owner", reflections: 1, state: 0,
			wantState: 0, wantReturn: true,
			wantTail: []string{
				"reflections:1", "velocity:source", "wall-reflect:vel(source)",
				"reflections:1", "store-reflections:0", "trace", "map-damage:31:47:1:0:source",
				"state:0", "reflections:0", "store-state:0", "owner:source", "store-return:owner",
			},
		},
		{
			name: "first unreflected impact seeks a target", reflections: 0, state: 0,
			wantState: 2,
			wantTail: []string{
				"reflections:0", "random-reflect:source", "reflections:0", "state:0", "store-state:2",
				"trace", "map-damage:31:47:1:0:source", "state:2", "reflections:0", "retarget:source",
			},
		},
		{
			name: "existing seek state returns to owner", reflections: 0, state: 2,
			wantState: 0, wantReturn: true,
			wantTail: []string{
				"reflections:0", "random-reflect:source", "reflections:0", "state:2", "trace",
				"map-damage:31:47:1:0:source", "state:2", "reflections:0", "store-state:0",
				"owner:source", "store-return:owner",
			},
		},
		{
			name: "drop state drops entry item", reflections: 0, state: 1,
			wantState: 1,
			wantTail: []string{
				"reflections:0", "random-reflect:source", "reflections:0", "state:1", "trace",
				"map-damage:31:47:1:0:source", "state:1", "position:source",
				"drop:source:item:pos(source)", "delete:source",
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f, source, _, _ := newChakramFixture4EAF00()
			f.update.reflections = tc.reflections
			f.update.state = tc.state
			chakramCollide4EAF00(source, (*chakramCollideTestObject4EAF00)(nil), &collision, f.hooks())
			prefix := []string{"update:source", "inventory:source", "flags:item", "position:source", "fx:150:pos(source)"}
			assertChakramEvents4EAF00(t, f.events, append(prefix, tc.wantTail...))
			if f.update.reflections != 0 || f.update.state != tc.wantState {
				t.Fatalf("state = (%d, %d), want (0, %d)", f.update.reflections, f.update.state, tc.wantState)
			}
			if got := f.update.returnTo != nil; got != tc.wantReturn {
				t.Fatalf("return target presence = %t, want %t", got, tc.wantReturn)
			}
		})
	}
}

func TestChakramCollide4EAF00OwnerCatchUsesThreeLiveOwnerReads(t *testing.T) {
	f, source, _, item := newChakramFixture4EAF00()
	compareOwner := source.owner
	putOwner := &chakramCollideTestObject4EAF00{name: "put-owner"}
	equipOwner := &chakramCollideTestObject4EAF00{name: "equip-owner", class: chakramPlayerClassBit4EAF00}
	f.ownerReads = []*chakramCollideTestObject4EAF00{compareOwner, putOwner, equipOwner}
	chakramCollide4EAF00(source, compareOwner, (*int)(nil), f.hooks())
	assertChakramEvents4EAF00(t, f.events, []string{
		"update:source", "inventory:source", "flags:item", "material:owner", "owner:source",
		"detach:source:item", "owner:source", "put:put-owner:item:1", "owner:source",
		"class:equip-owner", "has-weapon:equip-owner", "equip:equip-owner:item:1:1",
		"audio:892:source", "delete:source",
	})
	if item.flags != 0 {
		t.Fatal("catch path unexpectedly mutated item flags")
	}
}

func TestChakramCollide4EAF00OwnerAndProjectileGates(t *testing.T) {
	t.Run("nil owner still performs projectile lookup before drop", func(t *testing.T) {
		f, source, target, _ := newChakramFixture4EAF00()
		source.owner = nil
		chakramCollide4EAF00(source, target, (*int)(nil), f.hooks())
		assertChakramEvents4EAF00(t, f.events, []string{
			"update:source", "inventory:source", "flags:item", "material:target", "owner:source",
			"same-team:source:target", "owner:source", "type:source", "projectile:77",
			"position:source", "drop:source:item:pos(source)", "delete:source",
		})
	})

	t.Run("invalid owner drops before target flags", func(t *testing.T) {
		f, source, target, _ := newChakramFixture4EAF00()
		source.owner.flags = chakramInvalidOwnerMask4EAF00
		chakramCollide4EAF00(source, target, (*int)(nil), f.hooks())
		assertChakramEvents4EAF00(t, f.events, []string{
			"update:source", "inventory:source", "flags:item", "material:target", "owner:source",
			"same-team:source:target", "owner:source", "type:source", "projectile:77", "flags:owner",
			"position:source", "drop:source:item:pos(source)", "delete:source",
		})
	})

	t.Run("untargetable target wins over nonnil projectile", func(t *testing.T) {
		f, source, target, _ := newChakramFixture4EAF00()
		target.flags = chakramUntargetableFlag4EAF00
		chakramCollide4EAF00(source, target, (*int)(nil), f.hooks())
		assertChakramEvents4EAF00(t, f.events, []string{
			"update:source", "inventory:source", "flags:item", "material:target", "owner:source",
			"same-team:source:target", "owner:source", "type:source", "projectile:77",
			"flags:owner", "flags:target",
		})
	})

	t.Run("nil projectile returns after target flags", func(t *testing.T) {
		f, source, target, _ := newChakramFixture4EAF00()
		f.projectileClass = nil
		chakramCollide4EAF00(source, target, (*int)(nil), f.hooks())
		assertChakramEvents4EAF00(t, f.events, []string{
			"update:source", "inventory:source", "flags:item", "material:target", "owner:source",
			"same-team:source:target", "owner:source", "type:source", "projectile:77",
			"flags:owner", "flags:target",
		})
	})
}

func TestChakramCollide4EAF00AttackRecordEffectsDamageAndReturnOrder(t *testing.T) {
	f, source, target, _ := newChakramFixture4EAF00()
	f.update.reflections = 1
	f.onApply = func(attack *chakramAttack4EAF00[*chakramCollideTestObject4EAF00]) {
		f.attack = *attack
		attack.Damage = 8.75
	}
	f.onPre = func(attack *chakramAttack4EAF00[*chakramCollideTestObject4EAF00]) {
		attack.Damage = 9.25
	}
	chakramCollide4EAF00(source, target, (*int)(nil), f.hooks())
	assertChakramEvents4EAF00(t, f.events, []string{
		"update:source", "inventory:source", "flags:item", "material:target", "owner:source",
		"same-team:source:target", "owner:source", "type:source", "projectile:77", "flags:owner",
		"flags:target", "strength:owner", "pos-x:source", "pos-y:source", "calc:37:class",
		"radius:source", "apply:source:owner", "pre:target:owner:source", "round:9.75",
		"damage:target:owner:source:10:0", "flags:target", "store-last:target", "reflections:1",
		"projectile-reflect:source:target", "reflections:1", "store-reflections:0", "state:0",
		"reflections:0", "store-return:owner", "store-state:0",
	})
	wantAttack := chakramAttack4EAF00[*chakramCollideTestObject4EAF00]{
		Damage: 7.25, Radius: 36, Owner: source.owner, PosX: 12.5, PosY: -4.25, Source: source,
	}
	if f.attack != wantAttack {
		t.Fatalf("initial attack = %+v, want %+v", f.attack, wantAttack)
	}
	if f.roundInput != 9.75 || f.damage != 10 || f.damageTarget != target {
		t.Fatalf("damage conversion = (%v, %d, %p), want (9.75, 10, %p)", f.roundInput, f.damage, f.damageTarget, target)
	}
	if f.update.lastHit != target || f.update.returnTo != source.owner || f.update.state != 0 {
		t.Fatalf("update = %+v, want last/return target and state zero", f.update)
	}
}

func TestChakramCollide4EAF00PostDamageDestroyedTargetIsNotRemembered(t *testing.T) {
	f, source, target, _ := newChakramFixture4EAF00()
	f.update.reflections = 1
	f.onDamage = func(target *chakramCollideTestObject4EAF00) {
		target.flags = chakramDestroyedFlag4EAF00
	}
	chakramCollide4EAF00(source, target, (*int)(nil), f.hooks())
	if f.update.lastHit != nil {
		t.Fatalf("last hit = %p, want nil after target destruction", f.update.lastHit)
	}
	if got := f.events[len(f.events)-7:]; !reflect.DeepEqual(got, []string{
		"projectile-reflect:source:target", "reflections:1", "store-reflections:0",
		"state:0", "reflections:0", "store-return:owner", "store-state:0",
	}) {
		t.Fatalf("post-damage tail = %q", got)
	}
}

func TestChakramCollide4EAF00UnitDropStateReloadsInventory(t *testing.T) {
	f, source, target, item := newChakramFixture4EAF00()
	liveItem := &chakramCollideTestObject4EAF00{name: "live-item"}
	f.inventoryReads = []*chakramCollideTestObject4EAF00{item, liveItem}
	f.update.reflections = 1
	f.update.state = chakramReturnStateDrop4EAF00
	chakramCollide4EAF00(source, target, (*int)(nil), f.hooks())
	wantTail := []string{
		"state:1", "inventory:source", "audio:893:source", "detach:source:live-item",
		"position:source", "create:live-item:nil:pos(source)", "delete:source",
	}
	if got := f.events[len(f.events)-len(wantTail):]; !reflect.DeepEqual(got, wantTail) {
		t.Fatalf("drop tail = %q, want %q", got, wantTail)
	}
}
