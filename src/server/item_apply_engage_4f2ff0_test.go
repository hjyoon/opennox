package server

import (
	"fmt"
	"reflect"
	"testing"
)

type itemApplyEngageTestCallback4F2FF0 struct {
	name string
}

type itemApplyEngageTestModifier4F2FF0 struct {
	name     string
	callback *itemApplyEngageTestCallback4F2FF0
}

type itemApplyEngageTestData4F2FF0 struct {
	name      string
	modifiers [4]*itemApplyEngageTestModifier4F2FF0
}

type itemApplyEngageTestObject4F2FF0 struct {
	name string
	data *itemApplyEngageTestData4F2FF0
}

type itemApplyEngageTestWorld4F2FF0 struct {
	item    *itemApplyEngageTestObject4F2FF0
	owner   *itemApplyEngageTestObject4F2FF0
	events  []string
	faultAt int

	afterLoadEngage func(*itemApplyEngageTestWorld4F2FF0, *itemApplyEngageTestModifier4F2FF0)
	afterCall       func(*itemApplyEngageTestWorld4F2FF0, *itemApplyEngageTestModifier4F2FF0)
}

func itemApplyEngageObjectName4F2FF0(obj *itemApplyEngageTestObject4F2FF0) string {
	if obj == nil {
		return "nil"
	}
	return obj.name
}

func itemApplyEngageDataName4F2FF0(data *itemApplyEngageTestData4F2FF0) string {
	if data == nil {
		return "nil"
	}
	return data.name
}

func itemApplyEngageModifierName4F2FF0(modifier *itemApplyEngageTestModifier4F2FF0) string {
	if modifier == nil {
		return "nil"
	}
	return modifier.name
}

func itemApplyEngageCallbackName4F2FF0(callback *itemApplyEngageTestCallback4F2FF0) string {
	if callback == nil {
		return "nil"
	}
	return callback.name
}

func (w *itemApplyEngageTestWorld4F2FF0) record(format string, args ...any) {
	event := fmt.Sprintf(format, args...)
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
}

func (w *itemApplyEngageTestWorld4F2FF0) hooks() itemApplyEngageEffectHooks4F2FF0[
	*itemApplyEngageTestObject4F2FF0,
	*itemApplyEngageTestData4F2FF0,
	*itemApplyEngageTestModifier4F2FF0,
	*itemApplyEngageTestCallback4F2FF0,
] {
	return itemApplyEngageEffectHooks4F2FF0[
		*itemApplyEngageTestObject4F2FF0,
		*itemApplyEngageTestData4F2FF0,
		*itemApplyEngageTestModifier4F2FF0,
		*itemApplyEngageTestCallback4F2FF0,
	]{
		loadInitData: func(item *itemApplyEngageTestObject4F2FF0) *itemApplyEngageTestData4F2FF0 {
			w.record("data:%s", itemApplyEngageObjectName4F2FF0(item))
			return item.data
		},
		loadModifier: func(data *itemApplyEngageTestData4F2FF0, slot int) *itemApplyEngageTestModifier4F2FF0 {
			modifier := data.modifiers[slot]
			w.record("modifier:%s[%d]=%s", itemApplyEngageDataName4F2FF0(data), slot, itemApplyEngageModifierName4F2FF0(modifier))
			return modifier
		},
		loadEngage: func(modifier *itemApplyEngageTestModifier4F2FF0) *itemApplyEngageTestCallback4F2FF0 {
			callback := modifier.callback
			w.record("engage:%s=%s", itemApplyEngageModifierName4F2FF0(modifier), itemApplyEngageCallbackName4F2FF0(callback))
			if w.afterLoadEngage != nil {
				w.afterLoadEngage(w, modifier)
			}
			return callback
		},
		callEngage: func(callback *itemApplyEngageTestCallback4F2FF0, modifier *itemApplyEngageTestModifier4F2FF0, owner, item *itemApplyEngageTestObject4F2FF0) {
			w.record("call:%s(%s,%s,%s)", itemApplyEngageCallbackName4F2FF0(callback), itemApplyEngageModifierName4F2FF0(modifier), itemApplyEngageObjectName4F2FF0(owner), itemApplyEngageObjectName4F2FF0(item))
			if w.afterCall != nil {
				w.afterCall(w, modifier)
			}
		},
	}
}

func newItemApplyEngageTestWorld4F2FF0() *itemApplyEngageTestWorld4F2FF0 {
	second := &itemApplyEngageTestModifier4F2FF0{
		name:     "second",
		callback: &itemApplyEngageTestCallback4F2FF0{name: "engage-second"},
	}
	third := &itemApplyEngageTestModifier4F2FF0{
		name:     "third",
		callback: &itemApplyEngageTestCallback4F2FF0{name: "engage-third"},
	}
	data := &itemApplyEngageTestData4F2FF0{name: "entry"}
	data.modifiers[2] = second
	data.modifiers[3] = third
	return &itemApplyEngageTestWorld4F2FF0{
		item:  &itemApplyEngageTestObject4F2FF0{name: "item", data: data},
		owner: &itemApplyEngageTestObject4F2FF0{name: "owner"},
	}
}

func itemApplyEngageExpectedEvents4F2FF0() []string {
	return []string{
		"data:item",
		"modifier:entry[2]=second",
		"engage:second=engage-second",
		"call:engage-second(second,owner,item)",
		"modifier:entry[3]=third",
		"engage:third=engage-third",
		"call:engage-third(third,owner,item)",
	}
}

func TestItemApplyEngageEffect4F2FF0ExactOrderAndArguments(t *testing.T) {
	w := newItemApplyEngageTestWorld4F2FF0()
	itemApplyEngageEffect4F2FF0(w.item, w.owner, w.hooks())
	if want := itemApplyEngageExpectedEvents4F2FF0(); !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want %q", w.events, want)
	}
}

func TestItemApplyEngageEffect4F2FF0NilSkipsAreOrdered(t *testing.T) {
	w := newItemApplyEngageTestWorld4F2FF0()
	w.item.data.modifiers[2] = nil
	w.item.data.modifiers[3].callback = nil
	itemApplyEngageEffect4F2FF0(w.item, nil, w.hooks())
	want := []string{
		"data:item",
		"modifier:entry[2]=nil",
		"modifier:entry[3]=third",
		"engage:third=nil",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want %q", w.events, want)
	}
}

func TestItemApplyEngageEffect4F2FF0CachesDataButReloadsSlotAndCallback(t *testing.T) {
	w := newItemApplyEngageTestWorld4F2FF0()
	entry := w.item.data
	replacementData := &itemApplyEngageTestData4F2FF0{name: "replacement-data"}
	replacementModifier := &itemApplyEngageTestModifier4F2FF0{
		name:     "replacement-modifier",
		callback: &itemApplyEngageTestCallback4F2FF0{name: "replacement-engage"},
	}
	w.afterCall = func(w *itemApplyEngageTestWorld4F2FF0, modifier *itemApplyEngageTestModifier4F2FF0) {
		if modifier.name != "second" {
			return
		}
		w.item.data = replacementData
		entry.modifiers[3] = replacementModifier
	}
	itemApplyEngageEffect4F2FF0(w.item, w.owner, w.hooks())
	wantSuffix := []string{
		"modifier:entry[3]=replacement-modifier",
		"engage:replacement-modifier=replacement-engage",
		"call:replacement-engage(replacement-modifier,owner,item)",
	}
	if got := w.events[len(w.events)-len(wantSuffix):]; !reflect.DeepEqual(got, wantSuffix) {
		t.Fatalf("event suffix = %q, want %q", got, wantSuffix)
	}
	if w.events[0] != "data:item" {
		t.Fatalf("InitData was not the first and only entry load: %q", w.events)
	}
	loads := 0
	for _, event := range w.events {
		if event == "data:item" {
			loads++
		}
	}
	if loads != 1 {
		t.Fatalf("InitData loads = %d, want 1", loads)
	}
}

func TestItemApplyEngageEffect4F2FF0CachesCallbackBeforeCall(t *testing.T) {
	w := newItemApplyEngageTestWorld4F2FF0()
	old := w.item.data.modifiers[2].callback
	w.afterLoadEngage = func(_ *itemApplyEngageTestWorld4F2FF0, modifier *itemApplyEngageTestModifier4F2FF0) {
		if modifier.name == "second" {
			modifier.callback = &itemApplyEngageTestCallback4F2FF0{name: "late"}
		}
	}
	itemApplyEngageEffect4F2FF0(w.item, w.owner, w.hooks())
	if got := w.events[3]; got != "call:engage-second(second,owner,item)" {
		t.Fatalf("first call = %q, want cached callback %q", got, old.name)
	}
}

func TestItemApplyEngageEffect4F2FF0PreservesUnguardedFaults(t *testing.T) {
	t.Run("nil item", func(t *testing.T) {
		w := newItemApplyEngageTestWorld4F2FF0()
		w.item = nil
		defer func() {
			if recover() == nil {
				t.Fatal("nil item did not preserve the original fault")
			}
			if want := []string{"data:nil"}; !reflect.DeepEqual(w.events, want) {
				t.Fatalf("events = %q, want %q", w.events, want)
			}
		}()
		itemApplyEngageEffect4F2FF0(w.item, w.owner, w.hooks())
	})

	t.Run("nil InitData", func(t *testing.T) {
		w := newItemApplyEngageTestWorld4F2FF0()
		w.item.data = nil
		defer func() {
			if recover() == nil {
				t.Fatal("nil InitData did not preserve the original fault")
			}
			if want := []string{"data:item"}; !reflect.DeepEqual(w.events, want) {
				t.Fatalf("events = %q, want %q", w.events, want)
			}
		}()
		itemApplyEngageEffect4F2FF0(w.item, w.owner, w.hooks())
	})
}

func TestItemApplyEngageEffect4F2FF0AllFaultPrefixes(t *testing.T) {
	wantEvents := itemApplyEngageExpectedEvents4F2FF0()
	for faultAt := 1; faultAt <= len(wantEvents); faultAt++ {
		t.Run(fmt.Sprintf("event-%02d", faultAt), func(t *testing.T) {
			w := newItemApplyEngageTestWorld4F2FF0()
			w.faultAt = faultAt
			defer func() {
				if got := recover(); got != wantEvents[faultAt-1] {
					t.Fatalf("panic = %v, want %q", got, wantEvents[faultAt-1])
				}
				if want := wantEvents[:faultAt]; !reflect.DeepEqual(w.events, want) {
					t.Fatalf("events = %q, want %q", w.events, want)
				}
			}()
			itemApplyEngageEffect4F2FF0(w.item, w.owner, w.hooks())
		})
	}
}
