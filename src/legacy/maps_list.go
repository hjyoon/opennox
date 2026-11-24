package legacy

/*
#include "defs.h"
void nox_common_list_clear_425760(nox_list_item_t* list);
extern nox_list_item_t nox_common_maplist;
void nox_common_scanAddMap(char* filename);
void nox_common_maplist_add_4D0760(nox_map_list_item* mp);
void* sub_425770(void* a1p);
*/
import "C"
import (
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

var _ = [1]struct{}{}[unsafe.Sizeof(C.nox_map_list_item{})-unsafe.Sizeof(Nox_map_list_item{})]

type nox_map_list_item = C.nox_map_list_item
type nativeListItem struct {
	next *nativeListItem
	prev *nativeListItem
	head *nativeListItem
}

type Nox_map_list_item struct {
	list      C.nox_list_item_t // 0, 0
	Name      [12]byte          // 3, 12
	Field_6   int32             // 6, 24
	Field_7   uint32            // 7, 28
	Field_8_0 byte              // 8, 32
	Field_8_1 byte              // 8, 33
	Field_8_2 uint16            // 8, 34
}

//export nox_xxx_mapGetTypeMB_4CFFA0
func nox_xxx_mapGetTypeMB_4CFFA0(record unsafe.Pointer) C.int {
	flags := *(*uint32)(unsafe.Add(record, 1392))
	return C.int(nox_mapToGameFlags_4CFF50(int32(flags)))
}

//export sub_4CFFC0
func sub_4CFFC0(entry *C.nox_map_list_item) C.int {
	return C.int(nox_mapToGameFlags_4CFF50(int32(entry.field_7)))
}

//export nox_common_maplist_first_4D09B0
func nox_common_maplist_first_4D09B0() *C.nox_map_list_item {
	link := nox_common_list_getFirstSafe_425890(&C.nox_common_maplist)
	return (*C.nox_map_list_item)(unsafe.Pointer(link))
}

//export nox_common_maplist_next_4D09C0
func nox_common_maplist_next_4D09C0(entry *C.nox_map_list_item) *C.nox_map_list_item {
	link := nox_common_list_getNextSafe_4258A0((*C.nox_list_item_t)(unsafe.Pointer(entry)))
	return (*C.nox_map_list_item)(unsafe.Pointer(link))
}

func Get_nox_common_maplist() unsafe.Pointer {
	return unsafe.Pointer(&C.nox_common_maplist)
}

func Nox_common_maplist_add_4D0760(p *Nox_map_list_item) {
	C.nox_common_maplist_add_4D0760((*nox_map_list_item)(unsafe.Pointer(p)))
}

func Sub_425770(p *Nox_map_list_item) {
	intrusiveListInit425770((*intrusiveListNode425760)(unsafe.Pointer(p)))
}

func noxCommonMaplistFreeNative() {
	head := (*nativeListItem)(unsafe.Pointer(&C.nox_common_maplist))
	for it := head.next; it != nil && it != head; {
		next := it.next
		alloc.FreePtr(unsafe.Pointer(it))
		it = next
	}
	head.next = head
	head.prev = head
	head.head = head
}

// ConsoleMapNames4D09B0 reads the native list and bounded Name field, rather
// than the original PE32 byte offset twelve used by the old console command.
func ConsoleMapNames4D09B0() []string {
	head := (*nativeListItem)(unsafe.Pointer(&C.nox_common_maplist))
	var names []string
	for it := head.next; it != nil && it != head; it = it.next {
		mp := (*Nox_map_list_item)(unsafe.Pointer(it))
		names = append(names, alloc.GoStringS(mp.Name[:]))
	}
	return names
}
