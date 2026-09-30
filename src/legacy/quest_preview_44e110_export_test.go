package legacy

import (
	"bytes"
	"fmt"
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func preserveQuestPreviewSlots44E110(t *testing.T) [12]**client.Drawable {
	t.Helper()
	slots := questPreviewSlots44E110()
	var old [12]*client.Drawable
	for i, p := range slots {
		old[i], *p = *p, nil
	}
	font, call := *questPreviewFontSlot44E110(), questPreviewCall44E110
	*questPreviewFontSlot44E110() = nil
	t.Cleanup(func() {
		for i, p := range slots {
			*p = old[i]
		}
		*questPreviewFontSlot44E110(), questPreviewCall44E110 = font, call
	})
	return slots
}

func TestQuestPreview44E110CEntryNativeStorageAndFields(t *testing.T) {
	slots := preserveQuestPreviewSlots44E110(t)
	font, freeFont := alloc.Malloc(8)
	t.Cleanup(freeFont)
	if unsafe.Sizeof(uintptr(0)) > 4 && uintptr(font) <= math.MaxUint32 {
		t.Fatalf("font=%p, want above 4 GiB", font)
	}
	var drawables [12]*client.Drawable
	var before [12][]byte
	for i := range drawables {
		dr, free := alloc.New(client.Drawable{})
		t.Cleanup(free)
		// Fill the whole foreign object so wrong PE32 offset writes show up.
		buf := unsafe.Slice((*byte)(unsafe.Pointer(dr)), int(unsafe.Sizeof(*dr)))
		for n := range buf {
			buf[n] = 0xa5
		}
		dr.ObjFlags = object.Flags(0x80000003)
		drawables[i], before[i] = dr, bytes.Clone(buf)
		if unsafe.Sizeof(uintptr(0)) > 4 && uintptr(unsafe.Pointer(dr)) <= math.MaxUint32 {
			t.Fatalf("drawable=%p, want above 4 GiB", dr)
		}
	}
	h := questPreviewNativeHooks44E110()
	var trace []string
	h.fontByName = func(name string) unsafe.Pointer {
		trace = append(trace, "font:"+name)
		return font
	}
	h.thingByName = func(name string) int32 {
		for _, spec := range questPreviewSpecs44E110 {
			if name == spec.name {
				trace = append(trace, "thing:"+name)
				return int32(spec.slot + 101)
			}
		}
		t.Fatalf("unexpected type %q", name)
		return 0
	}
	h.create = func(typ int32) *client.Drawable {
		trace = append(trace, fmt.Sprintf("create:%d", typ))
		return drawables[typ-101]
	}
	questPreviewCall44E110 = func() *client.Drawable { return questPreview44E110(h) }
	if got := questPreviewCEntry44E110(); got != drawables[11] || *questPreviewFontSlot44E110() != font {
		t.Fatalf("return=%p want=%p font=%p", got, drawables[11], *questPreviewFontSlot44E110())
	}
	want := []string{"font:default"}
	names := [...]string{"BeholderGenerator", "GauntletExitB", "Ankh", "SoulGate",
		"SilverKey", "GoldKey", "QuestGoldChest", "QuestGoldPile",
		"DunMirChest4", "WarHammer", "HastePotion", "ConjurerSpellBook"}
	for _, slot := range []int{1, 0, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11} {
		want = append(want, "thing:"+names[slot])
		want = append(want, fmt.Sprintf("create:%d", slot+101))
	}
	if !reflect.DeepEqual(trace, want) {
		t.Fatalf("trace=%v want=%v", trace, want)
	}
	for i, dr := range drawables {
		if *slots[i] != dr || dr.ObjFlags != object.Flags(0x81000003) {
			t.Fatalf("slot %d=%p flags=%#x", i, *slots[i], dr.ObjFlags)
		}
		// Comparing all bytes independently of the flag field catches a
		// native pointer-size or old +120 flag offset regression.
		dr.ObjFlags = object.Flags(0x80000003)
		if got := unsafe.Slice((*byte)(unsafe.Pointer(dr)), int(unsafe.Sizeof(*dr))); !bytes.Equal(got, before[i]) {
			t.Fatalf("non-flag bytes modified in drawable %d", i)
		}
	}
	trace = nil
	if got := questPreviewCEntry44E110(); got != drawables[11] || len(trace) != 0 {
		t.Fatalf("cache recreated: return=%p trace=%v", got, trace)
	}
	for i, dr := range drawables {
		if *slots[i] != dr || dr.ObjFlags != object.Flags(0x81000003) {
			t.Fatalf("cached slot %d=%p flags=%#x", i, *slots[i], dr.ObjFlags)
		}
	}
}

func TestQuestPreview44E110CEntryReturnsFullPointerAndNil(t *testing.T) {
	preserveQuestPreviewSlots44E110(t)
	dr, free := alloc.New(client.Drawable{})
	t.Cleanup(free)
	if unsafe.Sizeof(uintptr(0)) > 4 && uintptr(unsafe.Pointer(dr)) <= math.MaxUint32 {
		t.Fatalf("drawable=%p, want above 4 GiB", dr)
	}
	questPreviewCall44E110 = func() *client.Drawable { return dr }
	if got := questPreviewCEntry44E110(); got != dr {
		t.Fatalf("return=%p want=%p", got, dr)
	}
	questPreviewCall44E110 = func() *client.Drawable { return nil }
	if got := questPreviewCEntry44E110(); got != nil {
		t.Fatalf("nil return=%p", got)
	}
}
