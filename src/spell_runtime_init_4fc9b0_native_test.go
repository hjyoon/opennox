package opennox

import (
	"fmt"
	"math"
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestSpellRuntimeInitNative4FC9B0PreservesPointersLayoutAndRawIDs(t *testing.T) {
	allocatorClass := new(alloc.Class)
	allocator := alloc.ClassT[server.MagicEntityClass]{Class: allocatorClass}
	caster := new(server.Object)
	wantTypeIDs := [...]int32{
		0,
		-1,
		math.MinInt32,
		math.MaxInt32,
		0x01234567,
		-0x76543211,
		0x5a5aa5a5,
	}
	storedTypeIDs := [len(spellRuntimeObjectTypeNames4FC9B0)]uint32{}
	var events []string

	got := spellRuntimeInitNative4FC9B0(spellRuntimeInitNativeDeps4FC9B0{
		initDurations: func() int32 {
			events = append(events, "durations")
			return -1
		},
		newMagicClass: func(name string, recordSize uintptr, capacity int) alloc.ClassT[server.MagicEntityClass] {
			events = append(events, fmt.Sprintf("new-magic:%s:%d:%d", name, recordSize, capacity))
			wantNativeSize := uintptr(60) + 5*(unsafe.Sizeof(uintptr(0))-4)
			if recordSize != wantNativeSize || recordSize != unsafe.Sizeof(server.MagicEntityClass{}) {
				t.Fatalf("record size = %d, want native %d", recordSize, wantNativeSize)
			}
			return allocator
		},
		storeMagicClass: func(value alloc.ClassT[server.MagicEntityClass]) {
			events = append(events, "store-magic")
			if value.Class != allocatorClass {
				t.Fatalf("allocator pointer = %p, want exact %p", value.Class, allocatorClass)
			}
		},
		newObjectByTypeID: func(name string) *server.Object {
			events = append(events, "new-object:"+name)
			return caster
		},
		storeImaginaryCaster: func(value *server.Object) {
			events = append(events, "store-caster")
			if value != caster {
				t.Fatalf("caster pointer = %p, want exact %p", value, caster)
			}
		},
		createObjectAt: func(object, owner *server.Object, pos types.Pointf) {
			events = append(events, "create-object")
			if object != caster || owner != nil {
				t.Fatalf("create pointers = (%p, %p), want (%p, nil)", object, owner, caster)
			}
			if math.Float32bits(pos.X) != math.Float32bits(2944) || math.Float32bits(pos.Y) != math.Float32bits(2944) {
				t.Fatalf("create position = (%08x, %08x), want exact 2944", math.Float32bits(pos.X), math.Float32bits(pos.Y))
			}
		},
		objectTypeIDByName: func(name string) int32 {
			index := len(events) - 6
			if index%2 != 0 {
				t.Fatalf("lookup event index = %d", index)
			}
			index /= 2
			if index < 0 || index >= len(spellRuntimeObjectTypeNames4FC9B0) || name != spellRuntimeObjectTypeNames4FC9B0[index] {
				t.Fatalf("lookup %d = %q", index, name)
			}
			events = append(events, "lookup:"+name)
			return wantTypeIDs[index]
		},
		storeSpellObjectTypeID: func(index int, value uint32) {
			name := spellRuntimeObjectTypeNames4FC9B0[index]
			events = append(events, "store:"+name)
			if value != uint32(wantTypeIDs[index]) {
				t.Fatalf("stored %s ID = %#x, want raw %#x", name, value, uint32(wantTypeIDs[index]))
			}
			storedTypeIDs[index] = value
		},
	})

	if got != 1 {
		t.Fatalf("result = %d, want canonical 1", got)
	}
	wantEvents := []string{
		"durations",
		fmt.Sprintf("new-magic:magicEntityClass:%d:64", unsafe.Sizeof(server.MagicEntityClass{})),
		"store-magic",
		"new-object:ImaginaryCaster",
		"store-caster",
		"create-object",
	}
	for _, name := range spellRuntimeObjectTypeNames4FC9B0 {
		wantEvents = append(wantEvents, "lookup:"+name, "store:"+name)
	}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %#v, want %#v", events, wantEvents)
	}
	for i, value := range storedTypeIDs {
		if value != uint32(wantTypeIDs[i]) {
			t.Fatalf("stored ID %d = %#x, want %#x", i, value, uint32(wantTypeIDs[i]))
		}
	}
	runtime.KeepAlive(allocatorClass)
	runtime.KeepAlive(caster)
}

func TestSpellRuntimeInitNative4FC9B0StoresNilCasterBeforeStopping(t *testing.T) {
	allocatorClass := new(alloc.Class)
	var events []string
	storedCaster := new(server.Object)

	got := spellRuntimeInitNative4FC9B0(spellRuntimeInitNativeDeps4FC9B0{
		initDurations: func() int32 {
			events = append(events, "durations")
			return 1
		},
		newMagicClass: func(string, uintptr, int) alloc.ClassT[server.MagicEntityClass] {
			events = append(events, "new-magic")
			return alloc.ClassT[server.MagicEntityClass]{Class: allocatorClass}
		},
		storeMagicClass: func(alloc.ClassT[server.MagicEntityClass]) {
			events = append(events, "store-magic")
		},
		newObjectByTypeID: func(name string) *server.Object {
			events = append(events, "new-object:"+name)
			return nil
		},
		storeImaginaryCaster: func(value *server.Object) {
			events = append(events, "store-caster")
			storedCaster = value
		},
		createObjectAt: func(*server.Object, *server.Object, types.Pointf) {
			t.Fatal("nil caster reached create")
		},
		objectTypeIDByName: func(string) int32 {
			t.Fatal("nil caster reached type lookup")
			return 0
		},
		storeSpellObjectTypeID: func(int, uint32) {
			t.Fatal("nil caster reached type store")
		},
	})

	if got != 0 || storedCaster != nil {
		t.Fatalf("result/caster = (%d, %p), want (0, nil)", got, storedCaster)
	}
	wantEvents := []string{"durations", "new-magic", "store-magic", "new-object:ImaginaryCaster", "store-caster"}
	if !reflect.DeepEqual(events, wantEvents) {
		t.Fatalf("events = %#v, want %#v", events, wantEvents)
	}
	runtime.KeepAlive(allocatorClass)
}
