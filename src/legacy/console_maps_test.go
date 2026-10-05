package legacy

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func TestConsoleMapNamesUseNativeListLinksAndBoundedNames(t *testing.T) {
	head := (*nativeListItem)(Get_nox_common_maplist())
	oldHead := *head
	t.Cleanup(func() { *head = oldHead })
	head.next, head.prev, head.head = head, head, head
	names := []string{"war03b", "con02a", "012345678901"}
	for _, name := range names {
		mp, free := alloc.New(Nox_map_list_item{})
		t.Cleanup(free)
		copy(mp.Name[:], name)
		it := (*nativeListItem)(unsafe.Pointer(mp))
		it.next, it.prev, it.head = head, head.prev, head
		head.prev.next, head.prev = it, it
	}
	if got := ConsoleMapNames4D09B0(); !reflect.DeepEqual(got, names) {
		t.Fatalf("native list names = %#v want %#v", got, names)
	}
	head.next, head.prev = head, head
	if got := ConsoleMapNames4D09B0(); len(got) != 0 {
		t.Fatalf("empty list = %#v", got)
	}
}
