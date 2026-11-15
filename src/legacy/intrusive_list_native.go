package legacy

import "unsafe"

// The third native slot is a sentinel identity or a signed32bit sorting key.
// Keep its bits outside Go's pointer map; scalar keys are not valid Go pointers.
type intrusiveListNode425760 struct {
	next *intrusiveListNode425760
	prev *intrusiveListNode425760
	tag  uintptr
}

func intrusiveListClear425760(node *intrusiveListNode425760) {
	node.next = node
	node.prev = node
	node.tag = uintptr(unsafe.Pointer(node))
}

func intrusiveListInit425770(node *intrusiveListNode425760) *intrusiveListNode425760 {
	node.next = node
	node.prev = node
	node.tag = 0
	return node
}

func intrusiveListKey425790(node *intrusiveListNode425760) int32 {
	return *(*int32)(unsafe.Pointer(&node.tag))
}

func intrusiveListAscending425790(head, node *intrusiveListNode425760) int32 {
	var skipped int32
	key := intrusiveListKey425790(node)
	it := intrusiveListNext425940(head)
	for it != nil && key > intrusiveListKey425790(it) {
		skipped++
		it = intrusiveListNext425940(it)
	}
	if it == nil {
		it = head
	}
	intrusiveListBefore4258E0(it, node)
	return skipped
}

func intrusiveListDescending4257F0(head, node *intrusiveListNode425760) {
	key := intrusiveListKey425790(node)
	it := intrusiveListNext425940(head)
	for it != nil && intrusiveListKey425790(it) > key {
		it = intrusiveListNext425940(it)
	}
	if it == nil {
		it = head
	}
	intrusiveListBefore4258E0(it, node)
}

func intrusiveListFirstSafe425890(node *intrusiveListNode425760) *intrusiveListNode425760 {
	return intrusiveListNextSafe4258A0(node)
}

func intrusiveListNextSafe4258A0(node *intrusiveListNode425760) *intrusiveListNode425760 {
	if node == nil {
		return nil
	}
	return intrusiveListNext425940(node)
}

func intrusiveListAt4258C0(head *intrusiveListNode425760, index int32) *intrusiveListNode425760 {
	it := head.next
	if it == head {
		return nil
	}
	for index != 0 {
		index--
		it = it.next
		if it == head {
			return nil
		}
	}
	return it
}

func intrusiveListBefore4258E0(head, node *intrusiveListNode425760) {
	if head == nil || node == nil {
		intrusiveListAbort4258E0()
		return
	}
	previous := head.prev
	node.next = head
	node.prev = previous
	head.prev = node
	// Retain the existing C compatibility exception for a missing previous
	// link. Do not invent initialization of an invalid/partial list owner.
	if previous != nil {
		previous.next = node
	}
}

func intrusiveListAfter425900(head, node *intrusiveListNode425760) *intrusiveListNode425760 {
	node.prev = head
	node.next = head.next
	head.next = node
	node.next.prev = node
	return node
}

func intrusiveListRemove425920(node *intrusiveListNode425760) {
	if node == nil || node.next == nil || node.prev == nil {
		return
	}
	node.prev.next = node.next
	node.next.prev = node.prev
	node.next = node
	node.prev = node
}

func intrusiveListNext425940(node *intrusiveListNode425760) *intrusiveListNode425760 {
	it := node.next
	if it != nil && it.tag == uintptr(unsafe.Pointer(it)) {
		return nil
	}
	return it
}

func intrusiveListPrevious425960(node *intrusiveListNode425760) *intrusiveListNode425760 {
	it := node.prev
	if it.tag == uintptr(unsafe.Pointer(it)) {
		return nil
	}
	return it
}
