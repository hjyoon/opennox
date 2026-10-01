package server

import (
	"reflect"
	"testing"
	"unsafe"
)

func TestPlayerPreAttackEffects538290Gates(t *testing.T) {
	for _, tc := range []struct {
		name                                      string
		noItem, noOwner, nonUnit, gameplay, enemy bool
		buff23, buff27, result                    int32
		calls                                     []string
	}{
		{name: "no item", noItem: true},
		{name: "friend", calls: []string{"gameplay", "enemy"}},
		{name: "enemy", enemy: true, calls: []string{"gameplay", "enemy", "23", "27"}},
		{name: "friendly fire", gameplay: true, calls: []string{"gameplay", "23", "27"}},
		{name: "no owner", noOwner: true, calls: []string{"gameplay", "23", "27"}},
		{name: "non-unit", nonUnit: true, calls: []string{"gameplay", "23", "27"}},
		{name: "buff 23", gameplay: true, buff23: 1, result: 1, calls: []string{"gameplay", "23"}},
		{name: "buff 27", gameplay: true, buff27: 1, result: 1, calls: []string{"gameplay", "23", "27"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			item := &Object{InitData: unsafe.Pointer(&ModifierInitData{})}
			owner, target := &Object{ObjClass: 4}, &Object{}
			if tc.noItem {
				item = nil
			}
			if tc.noOwner {
				owner = nil
			}
			if tc.nonUnit {
				owner.ObjClass = 0x400
			}
			var calls []string
			r := preAttackEffectsRuntime538290{
				gameplay: func() bool { calls = append(calls, "gameplay"); return tc.gameplay },
				enemy: func(own, tgt *Object) bool {
					if own != owner || tgt != target {
						t.Fatal("enemy arguments")
					}
					calls = append(calls, "enemy")
					return tc.enemy
				},
				buff: func(tgt *Object, id int32) int32 {
					if tgt != target {
						t.Fatal("buff target")
					}
					if id == 23 {
						calls = append(calls, "23")
						return tc.buff23
					}
					if id != 27 {
						t.Fatal(id)
					}
					calls = append(calls, "27")
					return tc.buff27
				},
				invoke: func(unsafe.Pointer, *ModifierEff, *Object, *Object, *Object, unsafe.Pointer) {
					t.Fatal("empty chain invoked")
				},
			}
			if got := playerPreAttackEffects538290(target, owner, item, nil, r); got != tc.result || !reflect.DeepEqual(calls, tc.calls) {
				t.Fatalf("result/calls=%d/%v, want %d/%v", got, calls, tc.result, tc.calls)
			}
		})
	}
}

func TestPlayerPreAttackEffects538290CachedArrayLiveSlots(t *testing.T) {
	var token byte
	fn := unsafe.Pointer(&token)
	mods := [4]*ModifierEff{}
	for i := range mods {
		mods[i] = &ModifierEff{AttackPreHit52: ModifierEffFnc{Fnc: fn}}
	}
	data := &ModifierInitData{Modifiers: [4]*ModifierEff{mods[0], mods[1], mods[2], nil}}
	item := &Object{InitData: unsafe.Pointer(data)}
	owner, target := &Object{}, &Object{}
	var calls []*ModifierEff
	r := preAttackEffectsRuntime538290{
		gameplay: func() bool { item.InitData = nil; return true },
		buff:     func(*Object, int32) int32 { return 0 },
		invoke: func(gotfn unsafe.Pointer, mod *ModifierEff, it, own, tgt *Object, ctx unsafe.Pointer) {
			if gotfn != fn || it != item || own != owner || tgt != target || ctx != fn {
				t.Fatal("callback arguments")
			}
			calls = append(calls, mod)
			if mod == mods[2] {
				data.Modifiers[3] = mods[3]
			}
		},
	}
	if got := playerPreAttackEffects538290(target, owner, item, fn, r); got != 0 || !reflect.DeepEqual(calls, []*ModifierEff{mods[2], mods[3]}) {
		t.Fatalf("result/calls=%d/%v", got, calls)
	}
}

func TestPlayerPreAttackEffects538290RequiredArray(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("required array silently skipped")
		}
	}()
	playerPreAttackEffects538290(nil, nil, &Object{}, nil, preAttackEffectsRuntime538290{
		gameplay: func() bool { return true }, buff: func(*Object, int32) int32 { return 0 },
	})
}
