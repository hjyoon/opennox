package server

import (
	"reflect"
	"testing"
	"unsafe"
)

func TestItemApplyAttackEffect538840CachedArrayLiveSlots(t *testing.T) {
	var tokens [4]byte
	mods := [4]*ModifierEff{}
	for i := range mods {
		mods[i] = &ModifierEff{Attack40: ModifierEffFnc{Fnc: unsafe.Pointer(&tokens[i])}}
	}
	data := &ModifierInitData{Modifiers: [4]*ModifierEff{mods[0], nil, mods[2], mods[3]}}
	item := &Object{InitData: unsafe.Pointer(data)}
	owner := &Object{}
	context := unsafe.Pointer(&tokens[0])
	var calls []int
	got := itemApplyAttackEffect538840(item, owner, context, func(fn unsafe.Pointer, mod *ModifierEff, it, own *Object, ctx unsafe.Pointer) {
		if it != item || own != owner || ctx != context {
			t.Fatal("native callback arguments changed")
		}
		for i, want := range mods {
			if mod == want {
				calls = append(calls, i)
				if fn != unsafe.Pointer(&tokens[i]) {
					t.Fatal("callback identity changed")
				}
			}
		}
		if mod == mods[0] {
			// A replacement InitData is not followed, but later entries in the
			// entry-cached array must be read after this callback.
			item.InitData = unsafe.Pointer(&ModifierInitData{})
			data.Modifiers[1] = mods[1]
			data.Modifiers[2] = nil
		}
		if mod == mods[1] {
			mods[3].Attack40.Fnc = nil
		}
	})
	if got != 0 || !reflect.DeepEqual(calls, []int{0, 1}) {
		t.Fatalf("result/calls = %d/%v, want 0/[0 1]", got, calls)
	}
}

func TestItemApplyAttackEffect538840RequiredInputs(t *testing.T) {
	for _, tc := range []struct {
		name string
		item *Object
	}{
		{"item", nil}, {"modifier array", &Object{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("missing required input silently skipped")
				}
			}()
			itemApplyAttackEffect538840(tc.item, nil, nil, nil)
		})
	}
	if got := itemApplyAttackEffect538840(&Object{InitData: unsafe.Pointer(&ModifierInitData{})}, nil, nil, nil); got != 0 {
		t.Fatalf("empty chain result = %d", got)
	}
}
