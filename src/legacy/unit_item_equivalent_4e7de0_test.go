package legacy

import (
	"reflect"
	"testing"
)

type itemEquivalentObject4E7DE0 struct {
	name     string
	typeInd  uint16
	class    uint32
	subclass uint32
	attrs    *itemEquivalentAttrs4E7DE0
	use      *itemEquivalentUse4E7DE0
}

type itemEquivalentAttrs4E7DE0 struct {
	name      string
	modifiers [4]uintptr
}

type itemEquivalentUse4E7DE0 struct {
	name  string
	bytes []byte
}

func itemEquivalentTestHooks4E7DE0(events *[]string) itemEquivalentHooks4E7DE0[
	*itemEquivalentObject4E7DE0, *itemEquivalentAttrs4E7DE0, *itemEquivalentUse4E7DE0, uintptr,
] {
	return itemEquivalentHooks4E7DE0[
		*itemEquivalentObject4E7DE0, *itemEquivalentAttrs4E7DE0, *itemEquivalentUse4E7DE0, uintptr,
	]{
		loadType: func(obj *itemEquivalentObject4E7DE0) uint16 {
			*events = append(*events, "type:"+obj.name)
			return obj.typeInd
		},
		loadClass: func(obj *itemEquivalentObject4E7DE0) uint32 {
			*events = append(*events, "class:"+obj.name)
			return obj.class
		},
		loadInitData: func(obj *itemEquivalentObject4E7DE0) *itemEquivalentAttrs4E7DE0 {
			*events = append(*events, "init:"+obj.name)
			return obj.attrs
		},
		loadModifier: func(attrs *itemEquivalentAttrs4E7DE0, index int) uintptr {
			*events = append(*events, "modifier:"+attrs.name+":"+string(rune('0'+index)))
			return attrs.modifiers[index]
		},
		loadSubclass: func(obj *itemEquivalentObject4E7DE0) uint32 {
			*events = append(*events, "subclass:"+obj.name)
			return obj.subclass
		},
		loadUseData: func(obj *itemEquivalentObject4E7DE0) *itemEquivalentUse4E7DE0 {
			*events = append(*events, "use:"+obj.name)
			return obj.use
		},
		loadUseByte: func(data *itemEquivalentUse4E7DE0, index int) byte {
			*events = append(*events, "byte:"+data.name+":"+string(rune('0'+index)))
			return data.bytes[index]
		},
	}
}

func TestItemEquivalent4E7DE0GuardsTypeAndOrdinaryItem(t *testing.T) {
	candidate := &itemEquivalentObject4E7DE0{name: "candidate", typeInd: 0xffff}
	item := &itemEquivalentObject4E7DE0{name: "item", typeInd: 0xffff, class: ^uint32(0), subclass: ^uint32(0)}
	for _, tc := range []struct {
		name       string
		candidate  *itemEquivalentObject4E7DE0
		item       *itemEquivalentObject4E7DE0
		want       bool
		wantEvents []string
	}{
		{name: "nil candidate"},
		{name: "nil item", candidate: candidate},
		{
			name: "type mismatch", candidate: candidate,
			item:       &itemEquivalentObject4E7DE0{name: "item", typeInd: 0xfffe},
			wantEvents: []string{"type:candidate", "type:item"},
		},
		{
			name: "ordinary match ignores item mode fields", candidate: candidate, item: item, want: true,
			wantEvents: []string{"type:candidate", "type:item", "class:candidate"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var events []string
			got := itemEquivalent4E7DE0(tc.candidate, tc.item, itemEquivalentTestHooks4E7DE0(&events))
			if got != tc.want {
				t.Fatalf("result = %t, want %t", got, tc.want)
			}
			if !reflect.DeepEqual(events, tc.wantEvents) {
				t.Fatalf("events = %v, want %v", events, tc.wantEvents)
			}
		})
	}
}

func TestItemEquivalent4E7DE0ModifierPointerOrderAndShortCircuit(t *testing.T) {
	candidateAttrs := &itemEquivalentAttrs4E7DE0{name: "candidate", modifiers: [4]uintptr{1, 2, 3, 4}}
	itemAttrs := &itemEquivalentAttrs4E7DE0{name: "item", modifiers: [4]uintptr{1, 2, 9, 4}}
	candidate := &itemEquivalentObject4E7DE0{
		name: "candidate", typeInd: 7, class: itemEquivalentModifierClass4E7DE0, attrs: candidateAttrs,
	}
	item := &itemEquivalentObject4E7DE0{name: "item", typeInd: 7, attrs: itemAttrs}
	var events []string

	if itemEquivalent4E7DE0(candidate, item, itemEquivalentTestHooks4E7DE0(&events)) {
		t.Fatal("modifier mismatch returned true")
	}
	want := []string{
		"type:candidate", "type:item", "class:candidate", "init:candidate", "init:item",
		"modifier:candidate:0", "modifier:item:0",
		"modifier:candidate:1", "modifier:item:1",
		"modifier:candidate:2", "modifier:item:2",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}

	itemAttrs.modifiers[2] = 3
	events = nil
	if !itemEquivalent4E7DE0(candidate, item, itemEquivalentTestHooks4E7DE0(&events)) {
		t.Fatal("four equal modifier pointers returned false")
	}
	want = append(want,
		"modifier:candidate:3", "modifier:item:3",
	)
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("equal events = %v, want %v", events, want)
	}
}

func TestItemEquivalent4E7DE0ExactModifierClassMask(t *testing.T) {
	candidateAttrs := &itemEquivalentAttrs4E7DE0{name: "candidate", modifiers: [4]uintptr{1}}
	itemAttrs := &itemEquivalentAttrs4E7DE0{name: "item", modifiers: [4]uintptr{2}}
	for _, tc := range []struct {
		name        string
		objectClass uint32
		want        bool
	}{
		{name: "Wand", objectClass: 0x00001000},
		{name: "Weapon", objectClass: 0x01000000},
		{name: "Armor", objectClass: 0x02000000},
		{name: "Flag", objectClass: 0x10000000},
		{name: "adjacent low bit", objectClass: 0x00002000, want: true},
		{name: "adjacent high bit", objectClass: 0x04000000, want: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			candidate := &itemEquivalentObject4E7DE0{
				name: "candidate", typeInd: 8, class: tc.objectClass, attrs: candidateAttrs,
			}
			item := &itemEquivalentObject4E7DE0{name: "item", typeInd: 8, attrs: itemAttrs}
			var events []string
			got := itemEquivalent4E7DE0(candidate, item, itemEquivalentTestHooks4E7DE0(&events))
			if got != tc.want {
				t.Fatalf("result = %t, want %t; events = %v", got, tc.want, events)
			}
		})
	}
}

func TestItemEquivalent4E7DE0CachesClassButLoadsSubclassLate(t *testing.T) {
	attrs := &itemEquivalentAttrs4E7DE0{name: "attrs", modifiers: [4]uintptr{1, 2, 3, 4}}
	candidate := &itemEquivalentObject4E7DE0{
		name: "candidate", typeInd: 9,
		class: itemEquivalentModifierClass4E7DE0 | itemEquivalentInfoBookClass4E7DE0,
		attrs: attrs,
		use:   &itemEquivalentUse4E7DE0{name: "candidate", bytes: []byte{'x', 0}},
	}
	item := &itemEquivalentObject4E7DE0{
		name: "item", typeInd: 9, attrs: attrs,
		use: &itemEquivalentUse4E7DE0{name: "item", bytes: []byte{'x', 0}},
	}
	var events []string
	hooks := itemEquivalentTestHooks4E7DE0(&events)
	loadClass := hooks.loadClass
	hooks.loadClass = func(obj *itemEquivalentObject4E7DE0) uint32 {
		objectClass := loadClass(obj)
		obj.class = 0
		return objectClass
	}
	loadModifier := hooks.loadModifier
	hooks.loadModifier = func(got *itemEquivalentAttrs4E7DE0, index int) uintptr {
		modifier := loadModifier(got, index)
		if index == 3 {
			candidate.subclass = itemEquivalentFieldGuideSubclass4E7DE0
		}
		return modifier
	}

	if !itemEquivalent4E7DE0(candidate, item, hooks) {
		t.Fatalf("result = false; events = %v", events)
	}
	if candidate.class != 0 || candidate.subclass != itemEquivalentFieldGuideSubclass4E7DE0 {
		t.Fatalf("live mutations were not observed: class=%#x subclass=%#x", candidate.class, candidate.subclass)
	}
	wantSuffix := []string{
		"subclass:candidate", "use:item", "use:candidate",
		"byte:candidate:0", "byte:item:0", "byte:candidate:1", "byte:item:1",
	}
	if len(events) < len(wantSuffix) || !reflect.DeepEqual(events[len(events)-len(wantSuffix):], wantSuffix) {
		t.Fatalf("event suffix = %v, want %v; all events = %v", events, wantSuffix, events)
	}
}

func TestItemEquivalent4E7DE0SpellBookTakesPrecedence(t *testing.T) {
	candidate := &itemEquivalentObject4E7DE0{
		name: "candidate", typeInd: 11, class: itemEquivalentInfoBookClass4E7DE0,
		subclass: itemEquivalentSpellBookSubclass4E7DE0 | itemEquivalentFieldGuideSubclass4E7DE0,
		use:      &itemEquivalentUse4E7DE0{name: "candidate", bytes: []byte{7, 'x', 0}},
	}
	item := &itemEquivalentObject4E7DE0{
		name: "item", typeInd: 11,
		use: &itemEquivalentUse4E7DE0{name: "item", bytes: []byte{7, 'y', 0}},
	}
	var events []string
	if !itemEquivalent4E7DE0(candidate, item, itemEquivalentTestHooks4E7DE0(&events)) {
		t.Fatal("equal SpellBook ID returned false")
	}
	want := []string{
		"type:candidate", "type:item", "class:candidate", "subclass:candidate",
		"use:candidate", "use:item", "byte:candidate:0", "byte:item:0",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestItemEquivalent4E7DE0FieldGuideStringOrder(t *testing.T) {
	for _, tc := range []struct {
		name       string
		candidate  []byte
		item       []byte
		want       bool
		wantEvents []string
	}{
		{
			name: "mismatch", candidate: []byte{'a', 'b', 'c', 0}, item: []byte{'a', 'b', 'd', 0},
			wantEvents: []string{
				"type:candidate", "type:item", "class:candidate", "subclass:candidate",
				"use:item", "use:candidate",
				"byte:candidate:0", "byte:item:0", "byte:candidate:1", "byte:item:1",
				"byte:candidate:2", "byte:item:2",
			},
		},
		{
			name: "candidate NUL terminates equality", candidate: []byte{'a', 0, 'x'}, item: []byte{'a', 0, 'y'}, want: true,
			wantEvents: []string{
				"type:candidate", "type:item", "class:candidate", "subclass:candidate",
				"use:item", "use:candidate",
				"byte:candidate:0", "byte:item:0", "byte:candidate:1", "byte:item:1",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			candidate := &itemEquivalentObject4E7DE0{
				name: "candidate", typeInd: 12, class: itemEquivalentInfoBookClass4E7DE0,
				subclass: itemEquivalentFieldGuideSubclass4E7DE0,
				use:      &itemEquivalentUse4E7DE0{name: "candidate", bytes: tc.candidate},
			}
			item := &itemEquivalentObject4E7DE0{
				name: "item", typeInd: 12,
				use: &itemEquivalentUse4E7DE0{name: "item", bytes: tc.item},
			}
			var events []string
			got := itemEquivalent4E7DE0(candidate, item, itemEquivalentTestHooks4E7DE0(&events))
			if got != tc.want {
				t.Fatalf("result = %t, want %t", got, tc.want)
			}
			if !reflect.DeepEqual(events, tc.wantEvents) {
				t.Fatalf("events = %v, want %v", events, tc.wantEvents)
			}
		})
	}
}

func TestItemEquivalent4E7DE0OtherInfoBookUsesFirstByte(t *testing.T) {
	candidate := &itemEquivalentObject4E7DE0{
		name: "candidate", typeInd: 13, class: itemEquivalentInfoBookClass4E7DE0,
		use: &itemEquivalentUse4E7DE0{name: "candidate", bytes: []byte{42, 1}},
	}
	item := &itemEquivalentObject4E7DE0{
		name: "item", typeInd: 13,
		use: &itemEquivalentUse4E7DE0{name: "item", bytes: []byte{42, 2}},
	}
	var events []string
	if !itemEquivalent4E7DE0(candidate, item, itemEquivalentTestHooks4E7DE0(&events)) {
		t.Fatal("equal first use byte returned false")
	}
	want := []string{
		"type:candidate", "type:item", "class:candidate", "subclass:candidate",
		"use:candidate", "use:item", "byte:candidate:0", "byte:item:0",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestItemEquivalent4E7DE0FaultOrder(t *testing.T) {
	for _, tc := range []struct {
		name       string
		candidate  *itemEquivalentObject4E7DE0
		item       *itemEquivalentObject4E7DE0
		wantEvents []string
	}{
		{
			name: "modifier loads both init pointers before candidate fault",
			candidate: &itemEquivalentObject4E7DE0{
				name: "candidate", typeInd: 21, class: itemEquivalentModifierClass4E7DE0,
			},
			item: &itemEquivalentObject4E7DE0{
				name: "item", typeInd: 21, attrs: &itemEquivalentAttrs4E7DE0{name: "item"},
			},
			wantEvents: []string{
				"type:candidate", "type:item", "class:candidate", "init:candidate", "init:item",
			},
		},
		{
			name: "SpellBook loads both use pointers before candidate byte fault",
			candidate: &itemEquivalentObject4E7DE0{
				name: "candidate", typeInd: 22, class: itemEquivalentInfoBookClass4E7DE0,
				subclass: itemEquivalentSpellBookSubclass4E7DE0,
			},
			item: &itemEquivalentObject4E7DE0{
				name: "item", typeInd: 22,
				use: &itemEquivalentUse4E7DE0{name: "item", bytes: []byte{1}},
			},
			wantEvents: []string{
				"type:candidate", "type:item", "class:candidate", "subclass:candidate",
				"use:candidate", "use:item",
			},
		},
		{
			name: "FieldGuide loads item pointer first but candidate byte first",
			candidate: &itemEquivalentObject4E7DE0{
				name: "candidate", typeInd: 23, class: itemEquivalentInfoBookClass4E7DE0,
				subclass: itemEquivalentFieldGuideSubclass4E7DE0,
				use:      &itemEquivalentUse4E7DE0{name: "candidate", bytes: []byte{'a', 0}},
			},
			item: &itemEquivalentObject4E7DE0{name: "item", typeInd: 23},
			wantEvents: []string{
				"type:candidate", "type:item", "class:candidate", "subclass:candidate",
				"use:item", "use:candidate", "byte:candidate:0",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var events []string
			panicked := false
			func() {
				defer func() {
					panicked = recover() != nil
				}()
				itemEquivalent4E7DE0(tc.candidate, tc.item, itemEquivalentTestHooks4E7DE0(&events))
			}()
			if !panicked {
				t.Fatal("result did not fault")
			}
			if !reflect.DeepEqual(events, tc.wantEvents) {
				t.Fatalf("events = %v, want %v", events, tc.wantEvents)
			}
		})
	}
}
