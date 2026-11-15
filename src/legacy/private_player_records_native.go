package legacy

import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

// These are native foreign owners, not the packed PE32 records. In particular
// neither list header may overlap a name, an identity, or the raw init flag.
type privatePlayerRecord425AD0 struct {
	link     intrusiveListNode425760
	name     [10]uint16
	id       int32
	team     unsafe.Pointer
	children intrusiveListNode425760
}

type privatePlayerChild425B30 struct {
	link intrusiveListNode425760
	peer int32
}

type privatePlayerRecordStore4259C0 struct {
	head     *intrusiveListNode425760
	allocate func(uintptr) unsafe.Pointer
	release  func(unsafe.Pointer)
}

var privatePlayerRecords4259C0 privatePlayerRecordStore4259C0

func (s *privatePlayerRecordStore4259C0) root() *intrusiveListNode425760 {
	if s.head == nil {
		// The original root is static. Retain the foreign replacement for the
		// process lifetime, including free-all and server stop/start cycles.
		s.head, _ = alloc.New(intrusiveListNode425760{})
	}
	return s.head
}

func (s *privatePlayerRecordStore4259C0) malloc(size uintptr) unsafe.Pointer {
	if s.allocate != nil {
		return s.allocate(size)
	}
	p, _ := alloc.Malloc(size) // Retains the decoded C's zeroed allocation.
	return p
}

func (s *privatePlayerRecordStore4259C0) free(p unsafe.Pointer) {
	if s.release != nil {
		s.release(p)
		return
	}
	alloc.FreePtr(p)
}

func (s *privatePlayerRecordStore4259C0) init() {
	flag := memmap.PtrUint32(0x5D4594, 599472)
	if *flag == 0 {
		intrusiveListClear425760(s.root())
		*flag = 1
	}
}

func (s *privatePlayerRecordStore4259C0) first() *privatePlayerRecord425AD0 {
	return (*privatePlayerRecord425AD0)(unsafe.Pointer(intrusiveListFirstSafe425890(s.root())))
}

func privatePlayerRecordNext425A60(p *privatePlayerRecord425AD0) *privatePlayerRecord425AD0 {
	return (*privatePlayerRecord425AD0)(unsafe.Pointer(intrusiveListNextSafe4258A0((*intrusiveListNode425760)(unsafe.Pointer(p)))))
}

func (s *privatePlayerRecordStore4259C0) find(id int32) *privatePlayerRecord425AD0 {
	for p := s.first(); p != nil; p = privatePlayerRecordNext425A60(p) {
		if p.id == id {
			return p
		}
	}
	return nil
}

func (s *privatePlayerRecordStore4259C0) create(id int32, name *uint16) *privatePlayerRecord425AD0 {
	if s.find(id) != nil {
		return nil
	}
	p := (*privatePlayerRecord425AD0)(s.malloc(unsafe.Sizeof(privatePlayerRecord425AD0{})))
	p.id = id
	// Preserve raw UTF16 and the original name/ID overlap at full capacity.
	// StrCopy16 and a Go string round trip would change that behavior.
	dst := unsafe.Pointer(&p.name[0])
	for {
		unit := *name
		*(*uint16)(dst) = unit
		if unit == 0 {
			break
		}
		dst = unsafe.Add(dst, 2)
		name = (*uint16)(unsafe.Add(unsafe.Pointer(name), 2))
	}
	p.team = nil
	intrusiveListInit425770(&p.link)
	intrusiveListClear425760(&p.children)
	intrusiveListBefore4258E0(s.root(), &p.link)
	return p
}

func (s *privatePlayerRecordStore4259C0) addChild(p *privatePlayerRecord425AD0, peer int32) {
	child := (*privatePlayerChild425B30)(s.malloc(unsafe.Sizeof(privatePlayerChild425B30{})))
	child.peer = peer
	intrusiveListInit425770(&child.link)
	intrusiveListBefore4258E0(&p.children, &child.link)
}

func privatePlayerFirstChild425BC0(p *privatePlayerRecord425AD0) *privatePlayerChild425B30 {
	return (*privatePlayerChild425B30)(unsafe.Pointer(intrusiveListNext425940(&p.children)))
}

func privatePlayerNextChild425BE0(p *privatePlayerChild425B30) *privatePlayerChild425B30 {
	return (*privatePlayerChild425B30)(unsafe.Pointer(intrusiveListNextSafe4258A0((*intrusiveListNode425760)(unsafe.Pointer(p)))))
}

func (s *privatePlayerRecordStore4259C0) removeChild(p *privatePlayerRecord425AD0, peer int32) unsafe.Pointer {
	for child := privatePlayerFirstChild425BC0(p); child != nil; child = privatePlayerNextChild425BE0(child) {
		if child.peer == peer {
			intrusiveListRemove425920(&child.link)
			s.free(unsafe.Pointer(child))
			break
		}
	}
	// Keep the decoded C pointer return, computed before releasing its owner.
	result := unsafe.Pointer(&p.children)
	if p.children.prev == &p.children {
		intrusiveListRemove425920(&p.link)
		s.free(unsafe.Pointer(p))
	}
	return result
}

func (s *privatePlayerRecordStore4259C0) freeAll() *privatePlayerRecord425AD0 {
	first := s.first()
	for p := first; p != nil; {
		next := privatePlayerRecordNext425A60(p)
		for child := privatePlayerFirstChild425BC0(p); child != nil; {
			nextChild := privatePlayerNextChild425BE0(child)
			intrusiveListRemove425920(&child.link)
			s.free(unsafe.Pointer(child))
			child = nextChild
		}
		intrusiveListRemove425920(&p.link)
		s.free(unsafe.Pointer(p))
		p = next
	}
	return first
}
