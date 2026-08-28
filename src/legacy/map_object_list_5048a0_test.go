package legacy

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestMapObjectList5048A0PreservesNativePointersAndBothLinkSets(t *testing.T) {
	oldHead := Get_dword_5d4594_1599540()
	Set_dword_5d4594_1599540(nil)
	first, freeFirst := alloc.New(server.Object{})
	second, freeSecond := alloc.New(server.Object{})
	t.Cleanup(func() {
		for node := Get_dword_5d4594_1599540(); node != nil; {
			next := MapObjectListNodeNext5048A0(node)
			FreeMapObjectListNode5048A0(node)
			node = next
		}
		Set_dword_5d4594_1599540(oldHead)
		freeSecond()
		freeFirst()
	})

	firstNode := mapObjectListAdd5048A0(first)
	if firstNode == nil || Get_dword_5d4594_1599540() != firstNode {
		t.Fatalf("first node/head = %p/%p", firstNode, Get_dword_5d4594_1599540())
	}
	if got := MapObjectListNodeObject5048A0(firstNode); got != first {
		t.Fatalf("first object = %p, want %p", got, first)
	}
	if next := MapObjectListNodeNext5048A0(firstNode); next != nil {
		t.Fatalf("first next node = %p, want nil", next)
	}
	if first.ObjNext != nil || first.ObjPrev != nil {
		t.Fatalf("first object links = %p/%p, want nil/nil", first.ObjNext, first.ObjPrev)
	}

	secondNode := mapObjectListAdd5048A0(second)
	if secondNode == nil || Get_dword_5d4594_1599540() != secondNode {
		t.Fatalf("second node/head = %p/%p", secondNode, Get_dword_5d4594_1599540())
	}
	if got := MapObjectListNodeObject5048A0(secondNode); got != second {
		t.Fatalf("second object = %p, want %p", got, second)
	}
	if next := MapObjectListNodeNext5048A0(secondNode); next != firstNode {
		t.Fatalf("second next node = %p, want %p", next, firstNode)
	}
	if second.ObjNext != first || second.ObjPrev != nil || first.ObjPrev != second {
		t.Fatalf("object links second=%p/%p first.previous=%p", second.ObjNext, second.ObjPrev, first.ObjPrev)
	}

	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(first)) <= math.MaxUint32 {
		t.Logf("allocator returned a low pointer %p; native round-trip checks still passed", first)
	}
}
