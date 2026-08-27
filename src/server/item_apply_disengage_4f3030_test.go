package server

import (
	"fmt"
	"reflect"
	"testing"
)

type itemApplyDisengageTestCallback4F3030 struct {
	name string
}

type itemApplyDisengageTestModifier4F3030 struct {
	name     string
	callback *itemApplyDisengageTestCallback4F3030
}

type itemApplyDisengageTestData4F3030 struct {
	name      string
	modifiers [4]*itemApplyDisengageTestModifier4F3030
}

type itemApplyDisengageTestObject4F3030 struct {
	name string
	data *itemApplyDisengageTestData4F3030
}

type itemApplyDisengageTestWorld4F3030 struct {
	item    *itemApplyDisengageTestObject4F3030
	owner   *itemApplyDisengageTestObject4F3030
	events  []string
	faultAt int

	afterLoadDisengage func(*itemApplyDisengageTestWorld4F3030, *itemApplyDisengageTestModifier4F3030)
	afterCall          func(*itemApplyDisengageTestWorld4F3030, *itemApplyDisengageTestModifier4F3030)
}

func itemApplyDisengageObjectName4F3030(obj *itemApplyDisengageTestObject4F3030) string {
	if obj == nil {
		return "nil"
	}
	return obj.name
}

func itemApplyDisengageDataName4F3030(data *itemApplyDisengageTestData4F3030) string {
	if data == nil {
		return "nil"
	}
	return data.name
}

func itemApplyDisengageModifierName4F3030(modifier *itemApplyDisengageTestModifier4F3030) string {
	if modifier == nil {
		return "nil"
	}
	return modifier.name
}

func itemApplyDisengageCallbackName4F3030(callback *itemApplyDisengageTestCallback4F3030) string {
	if callback == nil {
		return "nil"
	}
	return callback.name
}

func (w *itemApplyDisengageTestWorld4F3030) record(format string, args ...any) {
	event := fmt.Sprintf(format, args...)
	w.events = append(w.events, event)
	if w.faultAt != 0 && len(w.events) == w.faultAt {
		panic(event)
	}
}

func (w *itemApplyDisengageTestWorld4F3030) hooks() itemApplyDisengageEffectHooks4F3030[
	*itemApplyDisengageTestObject4F3030,
	*itemApplyDisengageTestData4F3030,
	*itemApplyDisengageTestModifier4F3030,
	*itemApplyDisengageTestCallback4F3030,
] {
	return itemApplyDisengageEffectHooks4F3030[
		*itemApplyDisengageTestObject4F3030,
		*itemApplyDisengageTestData4F3030,
		*itemApplyDisengageTestModifier4F3030,
		*itemApplyDisengageTestCallback4F3030,
	]{
		loadInitData: func(item *itemApplyDisengageTestObject4F3030) *itemApplyDisengageTestData4F3030 {
			w.record("data:%s", itemApplyDisengageObjectName4F3030(item))
			return item.data
		},
		loadModifier: func(data *itemApplyDisengageTestData4F3030, slot int) *itemApplyDisengageTestModifier4F3030 {
			modifier := data.modifiers[slot]
			w.record("modifier:%s[%d]=%s", itemApplyDisengageDataName4F3030(data), slot, itemApplyDisengageModifierName4F3030(modifier))
			return modifier
		},
		loadDisengage: func(modifier *itemApplyDisengageTestModifier4F3030) *itemApplyDisengageTestCallback4F3030 {
			callback := modifier.callback
			w.record("disengage:%s=%s", itemApplyDisengageModifierName4F3030(modifier), itemApplyDisengageCallbackName4F3030(callback))
			if w.afterLoadDisengage != nil {
				w.afterLoadDisengage(w, modifier)
			}
			return callback
		},
		callDisengage: func(callback *itemApplyDisengageTestCallback4F3030, modifier *itemApplyDisengageTestModifier4F3030, owner, item *itemApplyDisengageTestObject4F3030) {
			w.record("call:%s(%s,%s,%s)", itemApplyDisengageCallbackName4F3030(callback), itemApplyDisengageModifierName4F3030(modifier), itemApplyDisengageObjectName4F3030(owner), itemApplyDisengageObjectName4F3030(item))
			if w.afterCall != nil {
				w.afterCall(w, modifier)
			}
		},
	}
}

func newItemApplyDisengageTestWorld4F3030() *itemApplyDisengageTestWorld4F3030 {
	second := &itemApplyDisengageTestModifier4F3030{
		name:     "second",
		callback: &itemApplyDisengageTestCallback4F3030{name: "disengage-second"},
	}
	third := &itemApplyDisengageTestModifier4F3030{
		name:     "third",
		callback: &itemApplyDisengageTestCallback4F3030{name: "disengage-third"},
	}
	data := &itemApplyDisengageTestData4F3030{name: "entry"}
	data.modifiers[2] = second
	data.modifiers[3] = third
	return &itemApplyDisengageTestWorld4F3030{
		item:  &itemApplyDisengageTestObject4F3030{name: "item", data: data},
		owner: &itemApplyDisengageTestObject4F3030{name: "owner"},
	}
}

func itemApplyDisengageExpectedEvents4F3030() []string {
	return []string{
		"data:item",
		"modifier:entry[2]=second",
		"disengage:second=disengage-second",
		"call:disengage-second(second,owner,item)",
		"modifier:entry[3]=third",
		"disengage:third=disengage-third",
		"call:disengage-third(third,owner,item)",
	}
}

func TestItemApplyDisengageEffect4F3030ExactOrderAndArguments(t *testing.T) {
	w := newItemApplyDisengageTestWorld4F3030()
	itemApplyDisengageEffect4F3030(w.item, w.owner, w.hooks())
	if want := itemApplyDisengageExpectedEvents4F3030(); !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want %q", w.events, want)
	}
}

func TestItemApplyDisengageEffect4F3030NilSkipsAreOrdered(t *testing.T) {
	w := newItemApplyDisengageTestWorld4F3030()
	w.item.data.modifiers[2] = nil
	w.item.data.modifiers[3].callback = nil
	itemApplyDisengageEffect4F3030(w.item, nil, w.hooks())
	want := []string{
		"data:item",
		"modifier:entry[2]=nil",
		"modifier:entry[3]=third",
		"disengage:third=nil",
	}
	if !reflect.DeepEqual(w.events, want) {
		t.Fatalf("events = %q, want %q", w.events, want)
	}
}

func TestItemApplyDisengageEffect4F3030CachesDataButReloadsSlotAndCallback(t *testing.T) {
	w := newItemApplyDisengageTestWorld4F3030()
	entry := w.item.data
	replacementData := &itemApplyDisengageTestData4F3030{name: "replacement-data"}
	replacementModifier := &itemApplyDisengageTestModifier4F3030{
		name:     "replacement-modifier",
		callback: &itemApplyDisengageTestCallback4F3030{name: "replacement-disengage"},
	}
	w.afterCall = func(w *itemApplyDisengageTestWorld4F3030, modifier *itemApplyDisengageTestModifier4F3030) {
		if modifier.name != "second" {
			return
		}
		w.item.data = replacementData
		entry.modifiers[3] = replacementModifier
	}
	itemApplyDisengageEffect4F3030(w.item, w.owner, w.hooks())
	wantSuffix := []string{
		"modifier:entry[3]=replacement-modifier",
		"disengage:replacement-modifier=replacement-disengage",
		"call:replacement-disengage(replacement-modifier,owner,item)",
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

func TestItemApplyDisengageEffect4F3030CachesCallbackBeforeCall(t *testing.T) {
	w := newItemApplyDisengageTestWorld4F3030()
	old := w.item.data.modifiers[2].callback
	w.afterLoadDisengage = func(_ *itemApplyDisengageTestWorld4F3030, modifier *itemApplyDisengageTestModifier4F3030) {
		if modifier.name == "second" {
			modifier.callback = &itemApplyDisengageTestCallback4F3030{name: "late"}
		}
	}
	itemApplyDisengageEffect4F3030(w.item, w.owner, w.hooks())
	if got := w.events[3]; got != "call:disengage-second(second,owner,item)" {
		t.Fatalf("first call = %q, want cached callback %q", got, old.name)
	}
}

func TestItemApplyDisengageEffect4F3030PreservesUnguardedFaults(t *testing.T) {
	t.Run("nil item", func(t *testing.T) {
		w := newItemApplyDisengageTestWorld4F3030()
		w.item = nil
		defer func() {
			if recover() == nil {
				t.Fatal("nil item did not preserve the original fault")
			}
			if want := []string{"data:nil"}; !reflect.DeepEqual(w.events, want) {
				t.Fatalf("events = %q, want %q", w.events, want)
			}
		}()
		itemApplyDisengageEffect4F3030(w.item, w.owner, w.hooks())
	})

	t.Run("nil InitData", func(t *testing.T) {
		w := newItemApplyDisengageTestWorld4F3030()
		w.item.data = nil
		defer func() {
			if recover() == nil {
				t.Fatal("nil InitData did not preserve the original fault")
			}
			if want := []string{"data:item"}; !reflect.DeepEqual(w.events, want) {
				t.Fatalf("events = %q, want %q", w.events, want)
			}
		}()
		itemApplyDisengageEffect4F3030(w.item, w.owner, w.hooks())
	})
}

func TestItemApplyDisengageEffect4F3030AllFaultPrefixes(t *testing.T) {
	wantEvents := itemApplyDisengageExpectedEvents4F3030()
	for faultAt := 1; faultAt <= len(wantEvents); faultAt++ {
		t.Run(fmt.Sprintf("event-%02d", faultAt), func(t *testing.T) {
			w := newItemApplyDisengageTestWorld4F3030()
			w.faultAt = faultAt
			defer func() {
				if got := recover(); got != wantEvents[faultAt-1] {
					t.Fatalf("panic = %v, want %q", got, wantEvents[faultAt-1])
				}
				if want := wantEvents[:faultAt]; !reflect.DeepEqual(w.events, want) {
					t.Fatalf("events = %q, want %q", w.events, want)
				}
			}()
			itemApplyDisengageEffect4F3030(w.item, w.owner, w.hooks())
		})
	}
}
