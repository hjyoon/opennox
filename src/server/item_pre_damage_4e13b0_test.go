package server

import (
	"fmt"
	"math"
	"slices"
	"testing"
	"unsafe"
)

type itemPreDamageModifier4E13B0 struct {
	name string
	fnc  bool
}

type itemPreDamageInit4E13B0 struct {
	name  string
	slots [4]*itemPreDamageModifier4E13B0
}

type itemPreDamageObject4E13B0 struct {
	init *itemPreDamageInit4E13B0
}

func TestItemPreDamage4E13B0ReadCallAndFaultOrder(t *testing.T) {
	want := []string{"init", "slot:cached:0", "function:first", "call:first", "slot:cached:1", "function:empty", "slot:cached:2", "slot:cached:3", "function:last", "call:last"}
	for faultAt := -1; faultAt < len(want); faultAt++ {
		t.Run(fmt.Sprintf("fault-%02d", faultAt), func(t *testing.T) {
			first, last := &itemPreDamageModifier4E13B0{"first", true}, &itemPreDamageModifier4E13B0{"last", true}
			weapon := &itemPreDamageObject4E13B0{&itemPreDamageInit4E13B0{"cached", [4]*itemPreDamageModifier4E13B0{first, {"empty", false}, nil, last}}}
			target, source := &itemPreDamageObject4E13B0{}, &itemPreDamageObject4E13B0{}
			damage := int32(math.MinInt32)
			var events []string
			event := func(name string) {
				events = append(events, name)
				if len(events)-1 == faultAt {
					panic("004E13B0 injected fault")
				}
			}
			if faultAt >= 0 {
				defer func() {
					if got := recover(); got != "004E13B0 injected fault" || !slices.Equal(events, want[:faultAt+1]) {
						t.Fatalf("fault=%v events=%v, want exact prefix %v", got, events, want[:faultAt+1])
					}
					wantDamage := int32(math.MinInt32)
					if faultAt > 3 {
						wantDamage++ // Callback directly committed its word.
					}
					if damage != wantDamage {
						t.Fatalf("damage=%d, want completed prefix %d", damage, wantDamage)
					}
				}()
			}
			got := itemPreDamage4E13B0(target, source, weapon, &damage, itemPreDamageHooks4E13B0[*itemPreDamageObject4E13B0, *itemPreDamageInit4E13B0, *itemPreDamageModifier4E13B0, *int32]{
				initData: func(w *itemPreDamageObject4E13B0) *itemPreDamageInit4E13B0 { event("init"); return w.init },
				modifier: func(init *itemPreDamageInit4E13B0, slot int) *itemPreDamageModifier4E13B0 {
					event(fmt.Sprintf("slot:%s:%d", init.name, slot))
					return init.slots[slot]
				},
				hasPreDamage: func(m *itemPreDamageModifier4E13B0) bool { event("function:" + m.name); return m.fnc },
				applyPreDamage: func(m *itemPreDamageModifier4E13B0, w, s, trg *itemPreDamageObject4E13B0, d *int32) {
					event("call:" + m.name)
					if w != weapon || s != source || trg != target || d != &damage {
						t.Fatal("original modifier/weapon/source/target/damage-address arguments changed")
					}
					*d++
				},
			})
			if faultAt >= 0 || got != 0 || damage != math.MinInt32+2 || !slices.Equal(events, want) {
				t.Fatalf("return=%d damage=%d events=%v", got, damage, events)
			}
		})
	}
}

func TestItemPreDamage4E13B0NativeLiveSlotsCachedBaseAndRawWord(t *testing.T) {
	first := &ModifierEff{AttackPreDmg64: ModifierEffFnc{Fnc: unsafe.Pointer(new(byte))}}
	stale := &ModifierEff{AttackPreDmg64: ModifierEffFnc{Fnc: unsafe.Pointer(new(byte))}}
	second := &ModifierEff{AttackPreDmg64: ModifierEffFnc{Fnc: unsafe.Pointer(new(byte))}}
	third := &ModifierEff{AttackPreDmg64: ModifierEffFnc{Fnc: unsafe.Pointer(new(byte))}}
	fourth := &ModifierEff{AttackPreDmg64: ModifierEffFnc{Fnc: unsafe.Pointer(new(byte))}}
	base := &ModifierInitData{Modifiers: [4]*ModifierEff{first, stale, nil, stale}}
	weapon := &Object{InitData: unsafe.Pointer(base)} // No class, flags, HP or update gate.
	damage := int32(0)
	var calls []*ModifierEff
	got := ItemPreDamage4E13B0(nil, nil, weapon, &damage, ItemPreDamageRuntime4E13B0{
		ApplyPreDamage: func(m *ModifierEff, w, source, target *Object, d *int32) {
			if w != weapon || source != nil || target != nil || d != &damage {
				t.Fatal("native objects or original damage address changed")
			}
			calls = append(calls, m)
			switch m {
			case first:
				if *d != 0 {
					t.Fatal("zero word was clamped")
				}
				base.Modifiers[1] = second
				weapon.InitData = unsafe.Pointer(&ModifierInitData{Modifiers: [4]*ModifierEff{stale, stale, stale, stale}})
				*d = -7
			case second:
				if *d != -7 {
					t.Fatal("negative word was clamped")
				}
				base.Modifiers[2] = third
				*d = math.MinInt32
			case third:
				if *d != math.MinInt32 {
					t.Fatal("signed DWORD was narrowed")
				}
				base.Modifiers[3] = fourth
				*d = math.MaxInt32
			case fourth:
				if *d != math.MaxInt32 {
					t.Fatal("preceding callback word not visible")
				}
			default:
				t.Fatal("cached contents or replacement base reached")
			}
		},
	})
	if got != 0 || damage != math.MaxInt32 || !slices.Equal(calls, []*ModifierEff{first, second, third, fourth}) {
		t.Fatalf("return=%d damage=%d calls=%v", got, damage, calls)
	}
}

func TestItemPreDamage4E13B0NilBoundaries(t *testing.T) {
	t.Run("empty slots never read damage or callback", func(t *testing.T) {
		weapon := &Object{InitData: unsafe.Pointer(&ModifierInitData{Modifiers: [4]*ModifierEff{nil, {}, nil, {}}})}
		if got := ItemPreDamage4E13B0(nil, nil, weapon, nil, ItemPreDamageRuntime4E13B0{}); got != 0 {
			t.Fatalf("empty counter=%d", got)
		}
	})
	t.Run("callback may accept nil damage", func(t *testing.T) {
		mod := &ModifierEff{AttackPreDmg64: ModifierEffFnc{Fnc: unsafe.Pointer(new(byte))}}
		weapon := &Object{InitData: unsafe.Pointer(&ModifierInitData{Modifiers: [4]*ModifierEff{3: mod}})}
		calls := 0
		if got := ItemPreDamage4E13B0(nil, nil, weapon, nil, ItemPreDamageRuntime4E13B0{ApplyPreDamage: func(m *ModifierEff, w, s, target *Object, d *int32) {
			if m != mod || w != weapon || s != nil || target != nil || d != nil {
				t.Fatal("nil argument contract changed")
			}
			calls++
		}}); got != 0 || calls != 1 {
			t.Fatalf("counter=%d calls=%d", got, calls)
		}
	})
	for _, tc := range []struct {
		name   string
		weapon *Object
	}{
		{"nil weapon", nil},
		{"nil init base", &Object{}},
		{"nil callback at use", &Object{InitData: unsafe.Pointer(&ModifierInitData{Modifiers: [4]*ModifierEff{2: {AttackPreDmg64: ModifierEffFnc{Fnc: unsafe.Pointer(new(byte))}}}})}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("original at-use nil fault was suppressed")
				}
			}()
			ItemPreDamage4E13B0(nil, nil, tc.weapon, nil, ItemPreDamageRuntime4E13B0{})
		})
	}
}
