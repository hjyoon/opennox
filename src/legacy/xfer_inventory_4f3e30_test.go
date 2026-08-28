package legacy

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/server"
)

type inventoryXferTestObject4F3E30 struct {
	id       string
	previous *inventoryXferTestObject4F3E30
	next     *inventoryXferTestObject4F3E30
	first    *inventoryXferTestObject4F3E30
	holder   *inventoryXferTestObject4F3E30
}

func TestXferInventory4F3E30LegacyContractAndLiveHeadReads(t *testing.T) {
	owner := &inventoryXferTestObject4F3E30{id: "owner"}
	item := &inventoryXferTestObject4F3E30{id: "item", previous: new(inventoryXferTestObject4F3E30)}
	firstRead := &inventoryXferTestObject4F3E30{id: "first-read"}
	secondRead := &inventoryXferTestObject4F3E30{id: "second-read"}
	var events []string
	loads := 0

	err := xferInventory4F3E30(59, owner, 1, inventoryXferDeps4F3E30[*inventoryXferTestObject4F3E30]{
		readName: func() (string, error) {
			events = append(events, "name")
			return "RedPotion", nil
		},
		readTOC: func() (uint16, error) {
			t.Fatal("TOC branch reached for version 59")
			return 0, nil
		},
		lookupName: func(name string) int32 {
			events = append(events, "lookup-name:"+name)
			return int32(0x1234002A)
		},
		lookupTOC: func(uint16) int32 {
			t.Fatal("TOC lookup reached for version 59")
			return 0
		},
		readCRC: func() error {
			events = append(events, "crc")
			return nil
		},
		newObject: func(typeInd uint16) *inventoryXferTestObject4F3E30 {
			events = append(events, "new")
			if typeInd != 0x2A {
				t.Fatalf("low-word type = %#x, want 0x2a", typeInd)
			}
			return item
		},
		callXfer: func(got *inventoryXferTestObject4F3E30) error {
			events = append(events, "xfer")
			if got != item {
				t.Fatalf("xfer item = %p, want %p", got, item)
			}
			return nil
		},
		storePrevious: func(object, previous *inventoryXferTestObject4F3E30) {
			if previous == nil {
				events = append(events, "previous:nil")
			} else {
				events = append(events, "previous:"+object.id+":"+previous.id)
			}
			object.previous = previous
		},
		loadFirst: func(gotOwner *inventoryXferTestObject4F3E30) *inventoryXferTestObject4F3E30 {
			loads++
			events = append(events, "first")
			if gotOwner != owner {
				t.Fatalf("first owner = %p, want %p", gotOwner, owner)
			}
			if loads == 1 {
				return firstRead
			}
			return secondRead
		},
		storeNext: func(object, next *inventoryXferTestObject4F3E30) {
			events = append(events, "next")
			object.next = next
		},
		storeFirst: func(gotOwner, first *inventoryXferTestObject4F3E30) {
			events = append(events, "store-first")
			gotOwner.first = first
		},
		storeHolder: func(object, holder *inventoryXferTestObject4F3E30) {
			events = append(events, "holder")
			object.holder = holder
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	wantEvents := []string{
		"name", "lookup-name:RedPotion", "crc", "new", "xfer",
		"previous:nil", "first", "next", "first",
		"previous:second-read:item", "store-first", "holder",
	}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %v, want %v", events, wantEvents)
	}
	if item.previous != nil || item.next != firstRead || secondRead.previous != item {
		t.Fatalf("live links = previous %p next %p second.previous %p", item.previous, item.next, secondRead.previous)
	}
	if owner.first != item || item.holder != owner || loads != 2 {
		t.Fatalf("owner/head/holder/loads = %p/%p/%d, want %p/%p/2", owner.first, item.holder, loads, item, owner)
	}
}

func TestXferInventory4F3E30ModernMultipleItemsPrepend(t *testing.T) {
	old := &inventoryXferTestObject4F3E30{id: "old"}
	owner := &inventoryXferTestObject4F3E30{id: "owner", first: old}
	items := []*inventoryXferTestObject4F3E30{{id: "one"}, {id: "two"}}
	tocs := []uint16{0x1111, 0x2222}
	index := 0
	crcCount := 0

	err := xferInventory4F3E30(60, owner, 2, inventoryXferDeps4F3E30[*inventoryXferTestObject4F3E30]{
		readName: func() (string, error) {
			t.Fatal("name branch reached for version 60")
			return "", nil
		},
		readTOC: func() (uint16, error) {
			got := tocs[index]
			return got, nil
		},
		lookupName: func(string) int32 {
			t.Fatal("name lookup reached for version 60")
			return 0
		},
		lookupTOC: func(toc uint16) int32 {
			if toc != tocs[index] {
				t.Fatalf("TOC = %#x, want %#x", toc, tocs[index])
			}
			return int32(uint32(0xABCD0000) | uint32(index+1))
		},
		readCRC: func() error {
			crcCount++
			return nil
		},
		newObject: func(typeInd uint16) *inventoryXferTestObject4F3E30 {
			if typeInd != uint16(index+1) {
				t.Fatalf("type = %d, want %d", typeInd, index+1)
			}
			return items[index]
		},
		callXfer: func(item *inventoryXferTestObject4F3E30) error {
			if item != items[index] || owner.first == nil {
				t.Fatalf("xfer item/head = %p/%p", item, owner.first)
			}
			index++
			return nil
		},
		storePrevious: func(object, previous *inventoryXferTestObject4F3E30) {
			object.previous = previous
		},
		loadFirst: func(owner *inventoryXferTestObject4F3E30) *inventoryXferTestObject4F3E30 {
			return owner.first
		},
		storeNext: func(object, next *inventoryXferTestObject4F3E30) {
			object.next = next
		},
		storeFirst: func(owner, first *inventoryXferTestObject4F3E30) {
			owner.first = first
		},
		storeHolder: func(object, holder *inventoryXferTestObject4F3E30) {
			object.holder = holder
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if index != 2 || crcCount != 2 {
		t.Fatalf("items/CRC = %d/%d, want 2/2", index, crcCount)
	}
	if owner.first != items[1] || items[1].next != items[0] || items[0].next != old {
		t.Fatalf("forward list = %p -> %p -> %p", owner.first, items[1].next, items[0].next)
	}
	if items[1].previous != nil || items[0].previous != items[1] || old.previous != items[0] {
		t.Fatalf("previous list = %p/%p/%p", items[1].previous, items[0].previous, old.previous)
	}
	if items[0].holder != owner || items[1].holder != owner {
		t.Fatalf("holders = %p/%p, want %p", items[0].holder, items[1].holder, owner)
	}
}

func TestXferInventory4F3E30SignedNonpositiveCountReadsNothing(t *testing.T) {
	for _, count := range []int32{0, -1, -0x40000000, -0x80000000} {
		if err := xferInventory4F3E30[*inventoryXferTestObject4F3E30](
			0xffff, nil, count, inventoryXferDeps4F3E30[*inventoryXferTestObject4F3E30]{},
		); err != nil {
			t.Fatalf("count %d: %v", count, err)
		}
	}
}

func TestXferInventory4F3E30EntryPointsSkipGlobalStateForNonpositiveCount(t *testing.T) {
	oldGetServer := GetServer
	GetServer = func() Server {
		t.Fatal("GetServer reached for a nonpositive signed count")
		return nil
	}
	t.Cleanup(func() { GetServer = oldGetServer })

	for _, count := range []int32{0, -1, -0x40000000, -0x80000000} {
		if got := xferInventoryCall4F3E30(nil, 0xffff, count); got != 1 {
			t.Fatalf("call count %d = %d, want 1", count, got)
		}
		if err := monsterXferInventory4F3E30(nil, nil, 0xffff, uint32(count)); err != nil {
			t.Fatalf("monster count %d: %v", count, err)
		}
	}
}

func TestXferInventory4F3E30FailurePrefixes(t *testing.T) {
	stop := errors.New("stop")
	tests := []struct {
		name       string
		fail       string
		lookup     int32
		wantEvents []string
	}{
		{name: "name read", fail: "name", lookup: 7, wantEvents: []string{"name"}},
		{name: "unknown low word", lookup: int32(0x12340000), wantEvents: []string{"name", "lookup"}},
		{name: "CRC", fail: "crc", lookup: 7, wantEvents: []string{"name", "lookup", "crc"}},
		{name: "allocation", fail: "new", lookup: 7, wantEvents: []string{"name", "lookup", "crc", "new"}},
		{name: "recursive xfer", fail: "xfer", lookup: 7, wantEvents: []string{"name", "lookup", "crc", "new", "xfer"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			owner := &inventoryXferTestObject4F3E30{id: "owner"}
			item := &inventoryXferTestObject4F3E30{id: "item"}
			var events []string
			err := xferInventory4F3E30(59, owner, 1, inventoryXferDeps4F3E30[*inventoryXferTestObject4F3E30]{
				readName: func() (string, error) {
					events = append(events, "name")
					if tc.fail == "name" {
						return "", stop
					}
					return "item", nil
				},
				lookupName: func(string) int32 {
					events = append(events, "lookup")
					return tc.lookup
				},
				readCRC: func() error {
					events = append(events, "crc")
					if tc.fail == "crc" {
						return stop
					}
					return nil
				},
				newObject: func(uint16) *inventoryXferTestObject4F3E30 {
					events = append(events, "new")
					if tc.fail == "new" {
						return nil
					}
					return item
				},
				callXfer: func(*inventoryXferTestObject4F3E30) error {
					events = append(events, "xfer")
					if tc.fail == "xfer" {
						return stop
					}
					return nil
				},
				storePrevious: func(*inventoryXferTestObject4F3E30, *inventoryXferTestObject4F3E30) {
					events = append(events, "link")
				},
			})
			if err == nil {
				t.Fatal("failure was accepted")
			}
			if !reflect.DeepEqual(events, tc.wantEvents) {
				t.Fatalf("events = %v, want %v", events, tc.wantEvents)
			}
			if owner.first != nil || item.previous != nil || item.next != nil || item.holder != nil {
				t.Fatalf("failure linked item: owner=%p previous=%p next=%p holder=%p", owner.first, item.previous, item.next, item.holder)
			}
		})
	}
}

func TestXferInventory4F3E30DefersOwnerFaultUntilAfterRecursiveXfer(t *testing.T) {
	item := &inventoryXferTestObject4F3E30{id: "item", previous: new(inventoryXferTestObject4F3E30)}
	var events []string
	defer func() {
		if recover() == nil {
			t.Fatal("nil owner did not fault at first inventory-head load")
		}
		want := []string{"name", "lookup", "crc", "new", "xfer", "previous", "first"}
		if !reflect.DeepEqual(events, want) {
			t.Fatalf("fault events = %v, want %v", events, want)
		}
		if item.previous != nil {
			t.Fatalf("pre-fault previous = %p, want nil", item.previous)
		}
	}()
	_ = xferInventory4F3E30(59, (*inventoryXferTestObject4F3E30)(nil), 1, inventoryXferDeps4F3E30[*inventoryXferTestObject4F3E30]{
		readName:   func() (string, error) { events = append(events, "name"); return "item", nil },
		lookupName: func(string) int32 { events = append(events, "lookup"); return 1 },
		readCRC:    func() error { events = append(events, "crc"); return nil },
		newObject:  func(uint16) *inventoryXferTestObject4F3E30 { events = append(events, "new"); return item },
		callXfer:   func(*inventoryXferTestObject4F3E30) error { events = append(events, "xfer"); return nil },
		storePrevious: func(object, previous *inventoryXferTestObject4F3E30) {
			events = append(events, "previous")
			object.previous = previous
		},
		loadFirst: func(owner *inventoryXferTestObject4F3E30) *inventoryXferTestObject4F3E30 {
			events = append(events, "first")
			return owner.first
		},
	})
}

func TestReadInventoryTypeName4F3E30UsesCStringAndConsumesFullField(t *testing.T) {
	path := filepath.Join(t.TempDir(), "inventory-name.bin")
	payload := []byte{5, 'A', 0, 'B', 'C', 'D', 0xEE}
	if err := os.WriteFile(path, payload, 0o600); err != nil {
		t.Fatal(err)
	}
	cf, err := cryptfile.OpenFile(path, cryptfile.ReadOnly, -1)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = cf.Close() })
	name, err := readInventoryTypeName4F3E30(cf)
	if err != nil {
		t.Fatal(err)
	}
	if name != "A" {
		t.Fatalf("name = %q, want C string %q", name, "A")
	}
	tail, err := cf.ReadU8()
	if err != nil || tail != 0xEE {
		t.Fatalf("tail = %#x, %v, want 0xee", tail, err)
	}
	if _, err := cf.ReadU8(); err != io.EOF {
		t.Fatalf("trailing read error = %v, want EOF", err)
	}
}

func TestXferInventory4F3E30NativeLayout(t *testing.T) {
	type check struct {
		name string
		got  uintptr
		want uintptr
	}
	checks32 := []check{
		{"Object.size", unsafe.Sizeof(server.Object{}), 780},
		{"Object.InvHolder", unsafe.Offsetof(server.Object{}.InvHolder), 492},
		{"Object.InvNextItem", unsafe.Offsetof(server.Object{}.InvNextItem), 496},
		{"Object.Field125", unsafe.Offsetof(server.Object{}.Field125), 500},
		{"Object.InvFirstItem", unsafe.Offsetof(server.Object{}.InvFirstItem), 504},
		{"Object.Xfer", unsafe.Offsetof(server.Object{}.Xfer), 704},
	}
	checks64 := []check{
		{"Object.size", unsafe.Sizeof(server.Object{}), 928},
		{"Object.InvHolder", unsafe.Offsetof(server.Object{}.InvHolder), 520},
		{"Object.InvNextItem", unsafe.Offsetof(server.Object{}.InvNextItem), 528},
		{"Object.Field125", unsafe.Offsetof(server.Object{}.Field125), 536},
		{"Object.InvFirstItem", unsafe.Offsetof(server.Object{}.InvFirstItem), 544},
		{"Object.Xfer", unsafe.Offsetof(server.Object{}.Xfer), 784},
	}
	checks := checks64
	if unsafe.Sizeof(uintptr(0)) == 4 {
		checks = checks32
	}
	for _, tc := range checks {
		if tc.got != tc.want {
			t.Errorf("%s on %s/%s = %d, want %d", tc.name, runtime.GOOS, runtime.GOARCH, tc.got, tc.want)
		}
	}
}
