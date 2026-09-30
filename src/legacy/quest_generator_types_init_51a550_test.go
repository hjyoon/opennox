package legacy

import (
	"crypto/sha256"
	"fmt"
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/things"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func runQuestGeneratorTypesInitCEntry51A550(t *testing.T, d questGeneratorTypesInitDeps51A550) *byte {
	t.Helper()
	old := questGeneratorTypesInitDepsFactory51A550
	questGeneratorTypesInitDepsFactory51A550 = func() questGeneratorTypesInitDeps51A550 { return d }
	defer func() { questGeneratorTypesInitDepsFactory51A550 = old }()
	return questGeneratorTypesInitCEntry51A550()
}

func TestQuestGeneratorTypesInit51A550CallbackOrder(t *testing.T) {
	for _, cEntry := range []bool{false, true} {
		t.Run(fmt.Sprintf("CEntry=%t", cEntry), func(t *testing.T) {
			var trace []string
			generators := []*byte{alloc.InternCString("g0"), alloc.InternCString("stale-g1"), nil}
			creatures := []*byte{alloc.InternCString("stale-c0"), alloc.InternCString("c1")}
			var generatorIDs, creatureIDs [2]uint32
			ready := uint32(0x12345678)
			d := questGeneratorTypesInitDeps51A550{
				generatorName: func(i int) *byte {
					trace = append(trace, fmt.Sprintf("generator-name:%d", i))
					return generators[i]
				},
				creatureName: func(i int) *byte {
					trace = append(trace, fmt.Sprintf("creature-name:%d", i))
					return creatures[i]
				},
				lookupType: func(p *byte) uint32 {
					if unsafe.Sizeof(uintptr(0)) > 4 && uintptr(unsafe.Pointer(p)) <= math.MaxUint32 {
						t.Fatalf("lookup name = %p, want above 4 GiB", p)
					}
					name := alloc.GoString(p)
					trace = append(trace, "lookup:"+name)
					switch name {
					case "g0":
						return 0xFEDCBA98
					case "live-c0":
						generators[1] = alloc.InternCString("live-g1")
						return 0
					case "live-g1":
						return 0
					case "c1":
						return 0x87654321
					default:
						t.Fatalf("lookup of stale or unknown name %q", name)
						return 0
					}
				},
				storeGenerator: func(i int, id uint32) {
					trace = append(trace, fmt.Sprintf("store-generator:%d:%08x", i, id))
					generatorIDs[i] = id
					if i == 0 {
						creatures[0] = alloc.InternCString("live-c0")
					}
				},
				storeCreature: func(i int, id uint32) {
					trace = append(trace, fmt.Sprintf("store-creature:%d:%08x", i, id))
					creatureIDs[i] = id
				},
				storeReady: func(v uint32) {
					trace = append(trace, fmt.Sprintf("ready:%d", v))
					ready = v
				},
			}
			var got *byte
			if cEntry {
				got = runQuestGeneratorTypesInitCEntry51A550(t, d)
			} else {
				got = questGeneratorTypesInitNative51A550(d)
			}
			if got != nil || ready != 1 || generatorIDs != [2]uint32{0xFEDCBA98, 0} || creatureIDs != [2]uint32{0, 0x87654321} {
				t.Fatalf("result=%p ready=%d generator IDs=%x creature IDs=%x", got, ready, generatorIDs, creatureIDs)
			}
			want := []string{
				"generator-name:0", "lookup:g0", "store-generator:0:fedcba98", "creature-name:0",
				"lookup:live-c0", "store-creature:0:00000000", "generator-name:1", "lookup:live-g1",
				"store-generator:1:00000000", "creature-name:1", "lookup:c1", "store-creature:1:87654321",
				"generator-name:2", "ready:1",
			}
			if !reflect.DeepEqual(trace, want) {
				t.Fatalf("trace = %v, want %v", trace, want)
			}
		})
	}
}

func TestQuestGeneratorTypesInit51A550EmptyAndRepeatedCalls(t *testing.T) {
	var trace []string
	d := questGeneratorTypesInitDeps51A550{
		generatorName: func(i int) *byte {
			if i != 0 {
				t.Fatalf("empty table advanced to %d", i)
			}
			trace = append(trace, "name")
			return nil
		},
		storeReady: func(v uint32) {
			if v != 1 {
				t.Fatalf("ready = %d", v)
			}
			trace = append(trace, "ready")
		},
	}
	for range 2 {
		if got := runQuestGeneratorTypesInitCEntry51A550(t, d); got != nil {
			t.Fatalf("empty result = %p", got)
		}
	}
	if want := []string{"name", "ready", "name", "ready"}; !reflect.DeepEqual(trace, want) {
		t.Fatalf("empty/repeated trace = %v, want %v", trace, want)
	}
}

func TestQuestGeneratorTypesInit51A550SealedNamesAndNativeStorage(t *testing.T) {
	InitBlobData()
	var names []byte
	var pointers [54][2]*byte
	for i := range 53 {
		pointers[i] = [2]*byte{questGeneratorCreatureName51A550(i), questGeneratorName51A550(i)}
		for _, p := range pointers[i] {
			if p == nil {
				t.Fatalf("nil name at record %d", i)
			}
			if unsafe.Sizeof(uintptr(0)) > 4 && uintptr(unsafe.Pointer(p)) <= math.MaxUint32 {
				t.Fatalf("name at record %d = %p, want above 4 GiB", i, p)
			}
			names = append(names, alloc.GoString(p)...)
			names = append(names, 0)
		}
	}
	pointers[53] = [2]*byte{questGeneratorCreatureName51A550(53), questGeneratorName51A550(53)}
	if pointers[53] != [2]*byte{} {
		t.Fatalf("terminator names = %p/%p", pointers[53][0], pointers[53][1])
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(names)); len(names) != 1449 || got != "8674612fb9e7d8d37ef911c9d426b7713247fd82acd5d094a8ba6d5c6c0b0160" {
		t.Fatalf("sealed name pairs = %d bytes / %s", len(names), got)
	}

	srv := server.New(nil, nil, strman.New())
	t.Cleanup(srv.Close)
	srv.Types.Free()
	for i := range 53 {
		for _, p := range pointers[i] {
			id := alloc.GoString(p)
			if srv.Types.ByID(id) != nil {
				continue
			}
			if err := srv.Types.ReadObjectType(&things.Thing{Name: id}); err != nil {
				t.Fatal(err)
			}
		}
	}
	t.Cleanup(srv.Types.Free)
	oldGetServer := GetServer
	GetServer = func() Server { return &questCanJoinLegacyServer4E4100{srv: srv} }
	t.Cleanup(func() { GetServer = oldGetServer })
	packed := memmap.Slice(0x587000, 249896)[:864]
	saved := append([]byte(nil), packed...)
	ready := memmap.PtrUint32(0x5D4594, 2388664)
	savedReady := *ready
	t.Cleanup(func() { copy(packed, saved); *ready = savedReady })
	*ready = 0x80000001 // The initializer itself must not skip a ready table.
	if got := questGeneratorTypesInitCEntry51A550(); got != nil || *ready != 1 {
		t.Fatalf("real C entry result=%p ready=%d", got, *ready)
	}
	for i := range 54 {
		if questGeneratorCreatureName51A550(i) != pointers[i][0] || questGeneratorName51A550(i) != pointers[i][1] {
			t.Fatalf("native name pointer changed at record %d", i)
		}
		for _, off := range []int{0, 8} {
			if !reflect.DeepEqual(packed[16*i+off:16*i+off+4], saved[16*i+off:16*i+off+4]) {
				t.Fatalf("PE32 name slot %d/%d was overwritten", i, off)
			}
		}
		if i == 53 {
			if !reflect.DeepEqual(packed[16*i:16*(i+1)], saved[16*i:16*(i+1)]) {
				t.Fatal("terminator record overwritten")
			}
			continue
		}
		for j, off := range []int{4, 12} {
			want := uint32(srv.Types.IndByID(alloc.GoString(pointers[i][j])))
			if got := memmap.Uint32(0x587000, uintptr(249896+16*i+off)); got != want || want == 0 {
				t.Fatalf("record %d ID offset %d = %d, want %d", i, off, got, want)
			}
		}
	}
}
