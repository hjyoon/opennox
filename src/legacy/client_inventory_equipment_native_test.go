package legacy

import (
	"testing"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func TestClientInventoryEquippedListKeepsLayerOrderAndNativeLinks(t *testing.T) {
	objectStatusTestServer(t, 1)
	items, free := alloc.CloneSliceHigh([]client.Drawable{
		{ObjClass: object.ClassArmor, ObjSubClass: 0x40, NetCode32: 0x900001, TypeIDVal: 1},
		{ObjClass: object.ClassArmor, ObjSubClass: 0x100, NetCode32: 0x900002, TypeIDVal: 1},
		{ObjClass: object.ClassArmor, ObjSubClass: 0x40, NetCode32: 0x900003, TypeIDVal: 1},
	})
	t.Cleanup(free)
	first, last, middle := &items[0], &items[1], &items[2]
	heads := inventoryStateNativeHeads()
	previous := *heads
	clear(heads[:])
	t.Cleanup(func() { *heads = previous })
	a, b, c := (*nox_drawable)(first.C()), (*nox_drawable)(last.C()), (*nox_drawable)(middle.C())
	const layer = 2
	if got := sub_4622E0(a); got != layer || sub_4622E0(b) != layer {
		t.Fatalf("armor layer = %d", got)
	}
	if sub_4623E0(a, layer) != nil || sub_4623E0(b, layer) != a || sub_4623E0(c, layer) != a ||
		heads[layer] != first || first.NextPtr != middle || middle.Field_93 != first ||
		middle.NextPtr != last || last.Field_93 != middle {
		t.Fatalf("equipped layer links differ: head %p, first next %p, middle next %p, last prev %p", heads[layer], first.NextPtr, middle.NextPtr, last.Field_93)
	}
	if got := sub_461F90(0x900003); got != c || first.NextPtr != last || last.Field_93 != first ||
		middle.NextPtr != nil || middle.Field_93 != nil {
		t.Fatalf("middle unlink = %p, list %p/%p", got, first.NextPtr, last.Field_93)
	}
	if got := sub_461F90(0x900001); got != a || heads[layer] != last || last.Field_93 != nil {
		t.Fatalf("head unlink = %p, head %p", got, heads[layer])
	}
	if sub_461F90(0x900002) != b || heads[layer] != nil || sub_461F90(0x900002) != nil {
		t.Fatal("last unlink did not empty the layer")
	}
}
