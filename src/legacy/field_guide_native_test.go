package legacy

import (
	"slices"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestFieldGuideNative427070PartialLoadAndReload(t *testing.T) {
	old := fieldGuides
	var oldSizes [server.FieldGuideCount]byte
	for id := 1; id < len(oldSizes); id++ {
		oldSizes[id] = *memmap.PtrUint8(0x5D4594, uintptr(740100+28*id))
	}
	t.Cleanup(func() {
		fieldGuides = old
		for id := 1; id < len(oldSizes); id++ {
			*memmap.PtrUint8(0x5D4594, uintptr(740100+28*id)) = oldSizes[id]
		}
	})
	fieldGuides[6] = fieldGuideEntry{thingType: 55, unitSize: 99}
	*memmap.PtrUint8(0x5D4594, 740100+28*6) = 99

	types := make(map[string]*client.ObjectType)
	for id, subclass := range []object.SubClass{
		object.SubClass(object.MonsterSmall | object.MonsterMedium),
		object.SubClass(object.MonsterMedium), 0, 0, 0,
	} {
		name := server.RewardFieldGuideName4F0D20(id + 1)
		types[name] = &client.ObjectType{
			Name: alloc.InternCString(name), Field_1c: int32(101 + id), ObjSubClass: subclass,
		}
	}
	var stringsLoaded, imagesLoaded []string
	lookup := func(name string) *client.ObjectType { return types[name] }
	translate := func(id strman.ID, source string) string {
		if source != fieldGuideSource {
			t.Errorf("string %s loaded from %q", id, source)
		}
		stringsLoaded = append(stringsLoaded, string(id))
		return string(id)
	}
	image := func(name string) *noxrender.Image {
		imagesLoaded = append(imagesLoaded, name)
		return nil
	}
	if loadFieldGuides(lookup, translate, image) {
		t.Fatal("partial field guide load succeeded")
	}
	if first, next, end := nox_xxx_bookGetFirstCreMB_427300(), nox_xxx_bookGetNextCre_427320(1), nox_xxx_bookGetNextCre_427320(4); first != 1 || next != 2 || end != 0 {
		t.Fatalf("field guide iteration returned %d, %d, %d", first, next, end)
	}
	if bat, bomber, missing := nox_xxx_creatureIsCharmableByTT_4272B0(101), nox_xxx_creatureIsCharmableByTT_4272B0(0), nox_xxx_creatureIsCharmableByTT_4272B0(105); bat != 1 || bomber != 5 || missing != 0 {
		t.Fatalf("charm lookups returned %d, %d, %d", bat, bomber, missing)
	}
	if name := nox_xxx_guiCreatureGetName_427240(5); name != nil {
		t.Fatalf("Bomber returned a title at %p", name)
	}
	if desc := alloc.GoString16((*uint16)(unsafe.Pointer(nox_xxx_guideGetDescById_4272E0(5)))); desc != "creature_desc:Bomber" {
		t.Fatalf("Bomber description = %q", desc)
	}
	if small, medium, large := nox_xxx_guideGetUnitSize_427460(1), nox_xxx_guideGetUnitSize_427460(2), nox_xxx_guideGetUnitSize_427460(3); small != 1 || medium != 2 || large != 4 {
		t.Fatalf("field guide sizes = %d, %d, %d", small, medium, large)
	}
	if *memmap.PtrUint8(0x5D4594, 740100+28*3) != 4 ||
		*memmap.PtrUint8(0x5D4594, 740100+28*6) != 99 || fieldGuides[6] != (fieldGuideEntry{}) {
		t.Fatal("partial load changed mirrored sizes for unvisited records or retained a stale typed entry")
	}
	if len(stringsLoaded) != 10 || len(imagesLoaded) != 10 ||
		!slices.Equal(stringsLoaded[:2], []string{"creature:Bat", "creature_desc:Bat"}) ||
		!slices.Equal(imagesLoaded[:2], []string{"CreatureCageBat", "SpellbookBat"}) {
		t.Fatalf("asset loads = %v / %v", stringsLoaded, imagesLoaded)
	}

	if !loadFieldGuides(func(name string) *client.ObjectType {
		return &client.ObjectType{Name: alloc.InternCString(name), Field_1c: 200}
	}, translate, image) {
		t.Fatal("complete field guide load failed")
	}
	if last := nox_xxx_bookGetNextCre_427320(39); last != 40 || nox_xxx_bookGetNextCre_427320(last) != 0 {
		t.Fatalf("last guide ID = %d", last)
	}
	if got := GoString(nox_xxx_guideNameByN_427230(0)); got != "GUIDE_INVALID" {
		t.Fatalf("invalid field guide name = %q", got)
	}
	if nox_xxx_guideNameByN_427230(-1) != nil || nox_xxx_guideNameByN_427230(41) != nil {
		t.Fatal("out-of-range field guide ID returned a name")
	}
}
