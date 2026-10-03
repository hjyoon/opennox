package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

type itemDefendRecord4E1320 struct {
	name        string
	flags       uint32
	first, next *itemDefendRecord4E1320
	init        *itemDefendInit4E1320
}

type itemDefendInit4E1320 struct {
	name  string
	slots [4]*itemDefendModifier4E1320
}

type itemDefendModifier4E1320 struct {
	name      string
	hasDefend bool
}

type itemDefendFixture4E1320 struct {
	target, source, weapon, item *itemDefendRecord4E1320
	first, second                *itemDefendModifier4E1320
	damage                       int32
	typ                          int32
	events                       []string
	faultAt                      int
	contexts                     []*[2]int32
	apply                        func(*itemDefendModifier4E1320, *itemDefendRecord4E1320, *[2]int32)
}

func newItemDefendFixture4E1320() *itemDefendFixture4E1320 {
	f := &itemDefendFixture4E1320{
		target: &itemDefendRecord4E1320{name: "target"}, source: &itemDefendRecord4E1320{name: "source"},
		weapon: &itemDefendRecord4E1320{name: "weapon"}, item: &itemDefendRecord4E1320{name: "item", flags: 0x100},
		first: &itemDefendModifier4E1320{name: "first", hasDefend: true}, second: &itemDefendModifier4E1320{name: "second", hasDefend: true},
		damage: math.MinInt32, typ: math.MaxInt32, faultAt: -1,
	}
	f.item.init = &itemDefendInit4E1320{name: "cached", slots: [4]*itemDefendModifier4E1320{2: f.first, 3: f.second}}
	f.target.first = f.item
	return f
}

func (f *itemDefendFixture4E1320) event(event string) {
	f.events = append(f.events, event)
	if len(f.events)-1 == f.faultAt {
		panic("004E1320 injected fault")
	}
}

func (f *itemDefendFixture4E1320) hooks(t *testing.T) itemDefendEffectsHooks4E1320[*itemDefendRecord4E1320, *itemDefendInit4E1320, *itemDefendModifier4E1320] {
	return itemDefendEffectsHooks4E1320[*itemDefendRecord4E1320, *itemDefendInit4E1320, *itemDefendModifier4E1320]{
		lowDWORD: func(rec *itemDefendRecord4E1320) int32 { f.event("low:" + rec.name); return -16 },
		firstItem: func(rec *itemDefendRecord4E1320) *itemDefendRecord4E1320 {
			f.event("inventory:" + rec.name)
			return rec.first
		},
		flags:    func(rec *itemDefendRecord4E1320) uint32 { f.event("flags:" + rec.name); return rec.flags },
		initData: func(rec *itemDefendRecord4E1320) *itemDefendInit4E1320 { f.event("init:" + rec.name); return rec.init },
		modifier: func(init *itemDefendInit4E1320, slot int) *itemDefendModifier4E1320 {
			f.event(fmt.Sprintf("slot:%s:%d", init.name, slot))
			return init.slots[slot]
		},
		hasDefend:  func(mod *itemDefendModifier4E1320) bool { f.event("function:" + mod.name); return mod.hasDefend },
		loadDamage: func() int32 { f.event("load-damage"); return f.damage },
		storeDamage: func(value int32) {
			f.event("store-damage")
			f.damage = value
		},
		applyDefend: func(mod *itemDefendModifier4E1320, item, target, weapon, source *itemDefendRecord4E1320, context *[2]int32) {
			f.event("call:" + mod.name)
			if target != f.target || weapon != f.weapon || source != f.source || context[0] != f.damage || context[1] != f.typ {
				t.Fatal("original object argument order or raw DWORD context changed")
			}
			f.contexts = append(f.contexts, context)
			if f.apply != nil {
				f.apply(mod, item, context)
			} else {
				context[0]++
			}
			context[1] = -123 // Never carried into the next callback's type.
		},
		nextItem: func(rec *itemDefendRecord4E1320) *itemDefendRecord4E1320 {
			f.event("next:" + rec.name)
			return rec.next
		},
	}
}

func TestItemDefendEffects4E1320OriginalReadAndWriteOrder(t *testing.T) {
	f := newItemDefendFixture4E1320()
	if got := itemDefendEffects4E1320(f.target, f.source, f.weapon, f.typ, f.hooks(t)); got != 0 || f.damage != math.MinInt32+2 {
		t.Fatalf("return=%d damage=%#x", got, uint32(f.damage))
	}
	want := []string{
		"low:target", "inventory:target", "flags:item", "init:item", "slot:cached:2", "function:first",
		"load-damage", "call:first", "store-damage", "slot:cached:3", "function:second",
		"load-damage", "call:second", "store-damage", "next:item",
	}
	if !reflect.DeepEqual(f.events, want) {
		t.Fatalf("events=%v, want %v", f.events, want)
	}
	if len(f.contexts) != 2 || f.contexts[0] != f.contexts[1] {
		t.Fatal("original single stack context address was not reused")
	}
	for at := range want {
		t.Run(fmt.Sprintf("fault-prefix-%02d", at), func(t *testing.T) {
			f := newItemDefendFixture4E1320()
			f.faultAt = at
			defer func() {
				if got := recover(); got != "004E1320 injected fault" {
					t.Fatalf("fault=%v, want injected boundary", got)
				}
				if !reflect.DeepEqual(f.events, want[:at+1]) {
					t.Fatalf("fault events=%v, want %v", f.events, want[:at+1])
				}
				wantDamage := int32(math.MinInt32)
				if at >= 9 {
					wantDamage++ // First store completed before the next slot.
				}
				if at >= 14 {
					wantDamage++
				}
				if f.damage != wantDamage {
					t.Fatalf("damage=%#x, want committed prefix %#x", uint32(f.damage), uint32(wantDamage))
				}
			}()
			itemDefendEffects4E1320(f.target, f.source, f.weapon, f.typ, f.hooks(t))
		})
	}
}

func TestItemDefendEffects4E1320LiveSlotsAndLinks(t *testing.T) {
	f := newItemDefendFixture4E1320()
	stale := &itemDefendRecord4E1320{name: "stale", flags: 0x100}
	laterMod := &itemDefendModifier4E1320{name: "later", hasDefend: true}
	later := &itemDefendRecord4E1320{name: "later", init: &itemDefendInit4E1320{name: "later", slots: [4]*itemDefendModifier4E1320{2: laterMod}}}
	replacement := &itemDefendModifier4E1320{name: "replacement", hasDefend: true}
	oldInit := f.item.init
	f.item.next = stale
	f.apply = func(mod *itemDefendModifier4E1320, item *itemDefendRecord4E1320, context *[2]int32) {
		context[0]++
		switch mod {
		case f.first:
			f.item.init = &itemDefendInit4E1320{name: "must-not-reload"}
			oldInit.slots[3] = replacement // Cached base, live contents.
			f.target.first = stale         // First item must not be reloaded.
		case replacement:
			f.item.next = later // Link is read only after slot three returns.
			later.flags = 0x100
		case laterMod:
			if item != later {
				t.Fatal("wrong live inventory item")
			}
		default:
			t.Fatal("stale modifier reached")
		}
	}
	if got := itemDefendEffects4E1320(f.target, f.source, f.weapon, f.typ, f.hooks(t)); got != 0 || f.damage != math.MinInt32+3 {
		t.Fatalf("return=%d damage=%#x", got, uint32(f.damage))
	}
	want := []string{
		"low:target", "inventory:target", "flags:item", "init:item", "slot:cached:2", "function:first", "load-damage", "call:first", "store-damage",
		"slot:cached:3", "function:replacement", "load-damage", "call:replacement", "store-damage", "next:item",
		"flags:later", "init:later", "slot:later:2", "function:later", "load-damage", "call:later", "store-damage", "slot:later:3", "next:later",
	}
	if !reflect.DeepEqual(f.events, want) {
		t.Fatalf("events=%v, want %v", f.events, want)
	}
	for _, context := range f.contexts {
		if context != f.contexts[0] {
			t.Fatal("context address changed between items")
		}
	}
}

func TestItemDefendEffects4E1320NativeFlagsOnlyAndRawContext(t *testing.T) {
	for _, damage := range []int32{0, -1, math.MinInt32, math.MaxInt32} {
		t.Run(fmt.Sprintf("damage-%08x", uint32(damage)), func(t *testing.T) {
			mod := &ModifierEff{Defend76: ModifierEffFnc{Fnc: unsafe.Pointer(new(byte))}}
			init := &ModifierInitData{Modifiers: [4]*ModifierEff{2: mod, 3: mod}}
			item := &Object{ObjFlags: object.FlagEquipped, InitData: unsafe.Pointer(init)}
			target := &Object{InvFirstItem: item} // Deliberately no class, health or update data.
			if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(target)) <= math.MaxUint32 {
				t.Fatal("fixture does not exercise a native address above 4 GiB")
			}
			calls := 0
			var firstContext *[2]int32
			got := ItemDefendEffects4E1320(target, nil, nil, &damage, math.MinInt32, ItemDefendEffectsRuntime4E1320{
				ApplyDefend: func(gotMod *ModifierEff, gotItem, gotTarget, weapon, source *Object, context *[2]int32) {
					calls++
					if gotMod != mod || gotItem != item || gotTarget != target || weapon != nil || source != nil || context[0] != damage || context[1] != math.MinInt32 {
						t.Fatal("native records, nil object arguments or signed DWORD context changed")
					}
					if firstContext == nil {
						firstContext = context
					} else if context != firstContext {
						t.Fatal("context address changed")
					}
					context[0] = -1
					context[1] = 99
				},
			})
			if got != 0 || calls != 2 || damage != -1 {
				t.Fatalf("return=%d calls=%d damage=%d, want 0/2/-1", got, calls, damage)
			}
		})
	}
}

func TestItemDefendEffects4E1320ReturnAndSkippedContextReads(t *testing.T) {
	for _, tc := range []struct {
		name  string
		first *Object
		want  uint32
	}{
		{name: "empty"},
		{name: "non-equipped flags", first: &Object{ObjFlags: object.Flags(0xf0000000)}, want: 0xf0000000},
		{name: "two nil modifiers", first: &Object{ObjFlags: object.FlagEquipped, InitData: unsafe.Pointer(&ModifierInitData{})}},
		{name: "two nil functions", first: &Object{ObjFlags: object.FlagEquipped, InitData: unsafe.Pointer(&ModifierInitData{Modifiers: [4]*ModifierEff{2: {}, 3: {}}})}},
		{name: "last non-equipped flags", first: &Object{ObjFlags: object.FlagEquipped, InitData: unsafe.Pointer(&ModifierInitData{}), InvNextItem: &Object{ObjFlags: object.Flags(0x80000000)}}, want: 0x80000000},
		{name: "last equipped counter", first: &Object{InvNextItem: &Object{ObjFlags: object.FlagEquipped, InitData: unsafe.Pointer(&ModifierInitData{})}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			target := &Object{InvFirstItem: tc.first}
			want := tc.want
			if tc.first == nil {
				want = uint32(uintptr(unsafe.Pointer(target)))
			}
			// Missing damage and callback are legal until a nonnil Defend
			// function is reached. Missing init is legal for non-equipped items.
			if got := uint32(ItemDefendEffects4E1320(target, nil, nil, nil, math.MinInt32, ItemDefendEffectsRuntime4E1320{})); got != want {
				t.Fatalf("return=%#x, want %#x", got, want)
			}
		})
	}
}

func TestItemDefendEffects4E1320UnguardedFaultBoundaries(t *testing.T) {
	mod := &ModifierEff{Defend76: ModifierEffFnc{Fnc: unsafe.Pointer(new(byte))}}
	for _, tc := range []struct {
		name   string
		target *Object
		damage *int32
	}{
		{name: "nil target"},
		{name: "equipped nil init", target: &Object{InvFirstItem: &Object{ObjFlags: object.FlagEquipped}}},
		{name: "callback nil damage", target: &Object{InvFirstItem: &Object{ObjFlags: object.FlagEquipped, InitData: unsafe.Pointer(&ModifierInitData{Modifiers: [4]*ModifierEff{2: mod}})}}},
		{name: "missing runtime callback", target: &Object{InvFirstItem: &Object{ObjFlags: object.FlagEquipped, InitData: unsafe.Pointer(&ModifierInitData{Modifiers: [4]*ModifierEff{3: mod}})}}, damage: new(int32)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("original unguarded access was silently accepted")
				}
			}()
			ItemDefendEffects4E1320(tc.target, nil, nil, tc.damage, 0, ItemDefendEffectsRuntime4E1320{})
		})
	}
}
