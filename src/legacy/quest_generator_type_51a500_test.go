package legacy

import (
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

func runQuestGeneratorTypeCEntry51A500(t *testing.T, unit *server.Object, d questGeneratorTypeDeps51A500) int32 {
	t.Helper()
	old := questGeneratorTypeDepsFactory51A500
	questGeneratorTypeDepsFactory51A500 = func() questGeneratorTypeDeps51A500 { return d }
	defer func() { questGeneratorTypeDepsFactory51A500 = old }()
	return questGeneratorTypeCEntry51A500(unit)
}

func TestQuestGeneratorType51A500InitializationPrecedesNilAndTableGates(t *testing.T) {
	for _, ready := range []uint32{0, 1, 0x80000000, math.MaxUint32} {
		for _, nilUnit := range []bool{false, true} {
			t.Run(fmt.Sprintf("ready=%08x/nil=%t", ready, nilUnit), func(t *testing.T) {
				var trace []string
				var unit *server.Object
				if !nilUnit {
					var free func()
					unit, free = alloc.New(server.Object{})
					t.Cleanup(free)
				}
				d := questGeneratorTypeDeps51A500{
					ready: func() uint32 { trace = append(trace, "ready"); return ready },
					initialize: func() {
						trace = append(trace, "init")
						// Initialization is not followed by a second ready read.
					},
					generatorName: func(i int) *byte {
						trace = append(trace, fmt.Sprintf("name:%d", i))
						return nil
					},
					objectType: func(*server.Object) uint16 {
						t.Fatal("empty table read object type")
						return 0
					},
				}
				if got := runQuestGeneratorTypeCEntry51A500(t, unit, d); got != 0 {
					t.Fatalf("gated result = %d", got)
				}
				want := []string{"ready"}
				if ready == 0 {
					want = append(want, "init")
				}
				if !nilUnit {
					want = append(want, "name:0")
				}
				if !reflect.DeepEqual(trace, want) {
					t.Fatalf("trace = %v, want %v", trace, want)
				}
			})
		}
	}
}

func TestQuestGeneratorType51A500CachesZeroExtendedTypeAndFirstMatch(t *testing.T) {
	for _, cEntry := range []bool{false, true} {
		t.Run(fmt.Sprintf("CEntry=%t", cEntry), func(t *testing.T) {
			unit, free := alloc.New(server.Object{})
			t.Cleanup(free)
			unit.TypeInd = 77
			if unsafe.Sizeof(uintptr(0)) > 4 && uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 {
				t.Fatalf("unit = %p, want above 4 GiB", unit)
			}
			var trace []string
			name := alloc.InternCString("") // Non-nil empty names do not terminate.
			creatureIDs := []uint32{0x100004D, 1, 77, 77}
			d := questGeneratorTypeDeps51A500{
				ready:      func() uint32 { trace = append(trace, "ready"); return 0 },
				initialize: func() { trace = append(trace, "init"); unit.TypeInd = 77 },
				generatorName: func(i int) *byte {
					trace = append(trace, fmt.Sprintf("name:%d", i))
					if i < len(creatureIDs) {
						return name
					}
					return nil
				},
				objectType: func(got *server.Object) uint16 {
					if got != unit {
						t.Fatalf("object identity = %p, want %p", got, unit)
					}
					trace = append(trace, "object-type")
					return got.TypeInd
				},
				creatureType: func(i int) uint32 {
					trace = append(trace, fmt.Sprintf("creature:%d", i))
					unit.TypeInd = 1 // Recovered C's live reload would wrongly match row 1.
					return creatureIDs[i]
				},
				generatorType: func(i int) uint32 {
					trace = append(trace, fmt.Sprintf("generator:%d", i))
					return 0xFEDCBA98
				},
			}
			var got int32
			if cEntry {
				got = runQuestGeneratorTypeCEntry51A500(t, unit, d)
			} else {
				got = questGeneratorTypeNative51A500(unit, d)
			}
			if uint32(got) != 0xFEDCBA98 || unit.TypeInd != 1 {
				t.Fatalf("return bits = %08x, unit type = %d", uint32(got), unit.TypeInd)
			}
			want := []string{"ready", "init", "name:0", "object-type", "creature:0", "name:1", "creature:1", "name:2", "creature:2", "generator:2"}
			if !reflect.DeepEqual(trace, want) {
				t.Fatalf("trace = %v, want %v", trace, want)
			}
		})
	}
}

func TestQuestGeneratorType51A500UnknownAndZeroResult(t *testing.T) {
	for _, objectType := range []uint16{0, 2, math.MaxUint16} {
		t.Run(fmt.Sprint(objectType), func(t *testing.T) {
			unit, free := alloc.New(server.Object{})
			t.Cleanup(free)
			unit.TypeInd = objectType
			var trace []string
			d := questGeneratorTypeDeps51A500{
				ready: func() uint32 { return 1 },
				generatorName: func(i int) *byte {
					trace = append(trace, fmt.Sprintf("name:%d", i))
					if i < 2 {
						return alloc.InternCString("generator")
					}
					return nil
				},
				objectType: func(u *server.Object) uint16 { return u.TypeInd },
				creatureType: func(i int) uint32 {
					trace = append(trace, fmt.Sprintf("creature:%d", i))
					return uint32(i + 1)
				},
				generatorType: func(i int) uint32 {
					trace = append(trace, fmt.Sprintf("generator:%d", i))
					return 0
				},
			}
			if got := runQuestGeneratorTypeCEntry51A500(t, unit, d); got != 0 {
				t.Fatalf("zero/unknown result = %d", got)
			}
			want := []string{"name:0", "creature:0", "name:1", "creature:1", "name:2"}
			if objectType == 2 {
				want[len(want)-1] = "generator:1"
			}
			if !reflect.DeepEqual(trace, want) {
				t.Fatalf("trace = %v, want %v", trace, want)
			}
		})
	}
}

func TestQuestGeneratorType51A500RealTableCEntry(t *testing.T) {
	InitBlobData()
	srv := server.New(nil, nil, strman.New())
	t.Cleanup(srv.Close)
	srv.Types.Free()
	for i := range 53 {
		for _, p := range []*byte{questGeneratorCreatureName51A550(i), questGeneratorName51A550(i)} {
			id := alloc.GoString(p)
			if srv.Types.ByID(id) == nil {
				if err := srv.Types.ReadObjectType(&things.Thing{Name: id}); err != nil {
					t.Fatal(err)
				}
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
	*ready = 0
	if got := questGeneratorTypeCEntry51A500(nil); got != 0 || *ready != 1 {
		t.Fatalf("cold nil C entry = %d, ready = %d", got, *ready)
	}
	initialized := append([]byte(nil), packed...)
	unit, free := alloc.New(server.Object{})
	t.Cleanup(free)
	if unsafe.Sizeof(uintptr(0)) > 4 && uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 {
		t.Fatalf("unit = %p, want above 4 GiB", unit)
	}
	for i := range 53 {
		id := srv.Types.IndByID(alloc.GoString(questGeneratorCreatureName51A550(i)))
		if id == 0 || id > math.MaxUint16 {
			t.Fatalf("record %d creature ID = %d", i, id)
		}
		unit.TypeInd = uint16(id)
		var want uint32
		for first := 0; first < 53; first++ {
			if memmap.Uint32(0x587000, uintptr(249900+16*first)) == uint32(id) {
				want = memmap.Uint32(0x587000, uintptr(249908+16*first))
				break
			}
		}
		if got := questGeneratorTypeCEntry51A500(unit); uint32(got) != want || want == 0 || unit.TypeInd != uint16(id) {
			t.Fatalf("record %d result = %d, want %d, object type = %d", i, got, want, unit.TypeInd)
		}
	}
	unit.TypeInd = math.MaxUint16
	if got := questGeneratorTypeCEntry51A500(unit); got != 0 {
		t.Fatalf("unknown type result = %d", got)
	}
	if !reflect.DeepEqual(packed, initialized) || *ready != 1 {
		t.Fatal("ready lookup mutated the packed table or ready flag")
	}
}
