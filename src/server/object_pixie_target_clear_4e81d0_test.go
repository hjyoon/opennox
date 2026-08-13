package server

import (
	"fmt"
	"reflect"
	"testing"
)

type objectPixieTargetClearTestObject4E81D0 struct {
	typeInd uint16
	update  *objectPixieTargetClearTestUpdate4E81D0
}

type objectPixieTargetClearTestUpdate4E81D0 struct {
	target string
}

func TestObjectPixieTargetClear4E81D0CachePrecedesObject(t *testing.T) {
	t.Run("cached nil object", func(t *testing.T) {
		var events []string
		got := objectPixieTargetClear4E81D0[*objectPixieTargetClearTestObject4E81D0, *objectPixieTargetClearTestUpdate4E81D0](
			nil,
			objectPixieTargetClearHooks4E81D0[*objectPixieTargetClearTestObject4E81D0, *objectPixieTargetClearTestUpdate4E81D0]{
				loadPixieTypeID: func() uint32 {
					events = append(events, "cache")
					return 0x12345678
				},
				lookupObjectType: func(string) uint32 {
					t.Fatal("nonzero cache reached lookup")
					return 0
				},
				storePixieTypeID: func(uint32) { t.Fatal("nonzero cache reached store") },
				loadTypeInd:      func(*objectPixieTargetClearTestObject4E81D0) uint16 { t.Fatal("nil object was read"); return 0 },
				loadUpdateData: func(*objectPixieTargetClearTestObject4E81D0) *objectPixieTargetClearTestUpdate4E81D0 {
					t.Fatal("nil object update was read")
					return nil
				},
				clearTarget: func(*objectPixieTargetClearTestUpdate4E81D0) { t.Fatal("nil object target was cleared") },
			},
		)
		if got.typeID != 0x12345678 || got.returnsUpdate || got.updateData != nil {
			t.Fatalf("result = %#v, want scalar type ID", got)
		}
		if want := []string{"cache"}; !reflect.DeepEqual(events, want) {
			t.Fatalf("events = %v, want %v", events, want)
		}
	})

	t.Run("empty cache populates before type load", func(t *testing.T) {
		obj := &objectPixieTargetClearTestObject4E81D0{}
		update := &objectPixieTargetClearTestUpdate4E81D0{target: "enemy"}
		obj.update = update
		var events []string
		got := objectPixieTargetClear4E81D0(obj, objectPixieTargetClearHooks4E81D0[*objectPixieTargetClearTestObject4E81D0, *objectPixieTargetClearTestUpdate4E81D0]{
			loadPixieTypeID: func() uint32 {
				events = append(events, "cache")
				return 0
			},
			lookupObjectType: func(name string) uint32 {
				events = append(events, "lookup:"+name)
				return 31
			},
			storePixieTypeID: func(typeID uint32) {
				events = append(events, fmt.Sprintf("store:%d", typeID))
				obj.typeInd = uint16(typeID)
			},
			loadTypeInd: func(obj *objectPixieTargetClearTestObject4E81D0) uint16 {
				events = append(events, "type")
				return obj.typeInd
			},
			loadUpdateData: func(obj *objectPixieTargetClearTestObject4E81D0) *objectPixieTargetClearTestUpdate4E81D0 {
				events = append(events, "update")
				return obj.update
			},
			clearTarget: func(update *objectPixieTargetClearTestUpdate4E81D0) {
				events = append(events, "clear")
				update.target = ""
			},
		})
		if got.typeID != 31 || !got.returnsUpdate || got.updateData != update || update.target != "" {
			t.Fatalf("result/update = (%#v, %#v), want matched cleared update", got, update)
		}
		want := []string{"cache", "lookup:Pixie", "store:31", "type", "update", "clear"}
		if !reflect.DeepEqual(events, want) {
			t.Fatalf("events = %v, want %v", events, want)
		}
	})
}

func TestObjectPixieTargetClear4E81D0ComparesFullCachedType(t *testing.T) {
	obj := &objectPixieTargetClearTestObject4E81D0{typeInd: 1}
	var events []string
	got := objectPixieTargetClear4E81D0(obj, objectPixieTargetClearHooks4E81D0[*objectPixieTargetClearTestObject4E81D0, *objectPixieTargetClearTestUpdate4E81D0]{
		loadPixieTypeID:  func() uint32 { events = append(events, "cache"); return 0x10001 },
		lookupObjectType: func(string) uint32 { t.Fatal("nonzero cache reached lookup"); return 0 },
		storePixieTypeID: func(uint32) { t.Fatal("nonzero cache reached store") },
		loadTypeInd: func(obj *objectPixieTargetClearTestObject4E81D0) uint16 {
			events = append(events, "type")
			return obj.typeInd
		},
		loadUpdateData: func(*objectPixieTargetClearTestObject4E81D0) *objectPixieTargetClearTestUpdate4E81D0 {
			t.Fatal("16-bit alias loaded update data")
			return nil
		},
		clearTarget: func(*objectPixieTargetClearTestUpdate4E81D0) { t.Fatal("16-bit alias cleared target") },
	})
	if got.typeID != 0x10001 || got.returnsUpdate {
		t.Fatalf("result = %#v, want scalar mismatch", got)
	}
	if want := []string{"cache", "type"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestObjectPixieTargetClear4E81D0ZeroLookupCanMatchZeroType(t *testing.T) {
	update := &objectPixieTargetClearTestUpdate4E81D0{target: "enemy"}
	obj := &objectPixieTargetClearTestObject4E81D0{update: update}
	stores := 0
	got := objectPixieTargetClear4E81D0(obj, objectPixieTargetClearHooks4E81D0[*objectPixieTargetClearTestObject4E81D0, *objectPixieTargetClearTestUpdate4E81D0]{
		loadPixieTypeID: func() uint32 { return 0 },
		lookupObjectType: func(name string) uint32 {
			if name != "Pixie" {
				t.Fatalf("lookup name = %q", name)
			}
			return 0
		},
		storePixieTypeID: func(value uint32) {
			if value != 0 {
				t.Fatalf("stored = %d", value)
			}
			stores++
		},
		loadTypeInd: func(obj *objectPixieTargetClearTestObject4E81D0) uint16 { return obj.typeInd },
		loadUpdateData: func(obj *objectPixieTargetClearTestObject4E81D0) *objectPixieTargetClearTestUpdate4E81D0 {
			return obj.update
		},
		clearTarget: func(update *objectPixieTargetClearTestUpdate4E81D0) { update.target = "" },
	})
	if stores != 1 || !got.returnsUpdate || got.updateData != update || update.target != "" {
		t.Fatalf("stores/result/update = (%d, %#v, %#v)", stores, got, update)
	}
}
