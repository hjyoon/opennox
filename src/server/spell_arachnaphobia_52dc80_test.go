package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

func arachnaphobiaCastAlloc52DC80[T comparable](t *testing.T, value T) *T {
	t.Helper()
	ptr, free := alloc.New(value)
	t.Cleanup(free)
	*ptr = value
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(ptr)) <= math.MaxUint32 {
		t.Fatalf("native allocation %p must be above 4 GiB", ptr)
	}
	return ptr
}

func TestArachnaphobiaCastNativePointersLiveOrder52DC80(t *testing.T) {
	second := arachnaphobiaCastAlloc52DC80(t, Object{Field29: 123})
	owner := arachnaphobiaCastAlloc52DC80(t, Object{Field129: second})
	caster := arachnaphobiaCastAlloc52DC80(t, Object{PosVec: types.Ptf(10, 20)})
	arg := arachnaphobiaCastAlloc52DC80(t, SpellAcceptArg{Obj: second, Pos: types.Ptf(30, 40)})
	focus := arachnaphobiaCastAlloc52DC80(t, Object{Field5: 0x1234, ObjFlags: 0x100, Field29: 0x12345678})
	cache := arachnaphobiaCastAlloc52DC80(t, uint32(0))
	beforeSecond, beforeOwner, beforeFocus := *second, *owner, *focus
	var events []string
	h := arachnaphobiaCastDeps52DC80{
		loadCache: func() uint32 { events = append(events, "load"); return *cache },
		lookupType: func(key string) uint32 {
			events = append(events, "type:"+key)
			caster.PosVec, arg.Pos = types.Ptf(50, 60), types.Ptf(70, 80)
			return 11
		},
		storeCache: func(kind uint32) { events = append(events, "store"); *cache = kind },
		traceRay: func(from, to types.Pointf, flags MapTraceFlags) bool {
			events = append(events, "trace")
			if from != types.Ptf(50, 60) || to != types.Ptf(70, 80) || flags != 9 || *cache != 11 {
				t.Fatalf("trace=%v/%v/%d cache=%08x", from, to, flags, *cache)
			}
			*cache = 0x8001000b
			arg.Pos = types.Ptf(90, 100)
			return true
		},
		newObject: func(kind uint32) *Object {
			events = append(events, "new")
			if kind != 0x8001000b {
				t.Fatalf("live DWORD type=%08x", kind)
			}
			arg.Pos = types.Ptf(110, 120)
			return focus
		},
		createAt: func(got, by *Object, pos types.Pointf) {
			events = append(events, "create")
			if got != focus || by != owner || by == caster || pos != types.Ptf(110, 120) || arg.Obj != second {
				t.Fatalf("placement=%p/%p/%v", got, by, pos)
			}
		},
		inform: func(uint8, byte, int32) { t.Fatal("unexpected notification") },
	}
	got := arachnaphobiaCast52DC80(owner, caster, arg, h)
	want := []string{"load", "type:ArachnaphobiaFocus", "store", "trace", "load", "new", "create"}
	if got != 1 || !reflect.DeepEqual(events, want) || *second != beforeSecond || *owner != beforeOwner ||
		*focus != beforeFocus || *caster != (Object{PosVec: types.Ptf(50, 60)}) ||
		*arg != (SpellAcceptArg{Obj: second, Pos: types.Ptf(110, 120)}) {
		t.Fatalf("result/order/raw records=%d/%v", got, events)
	}
}

func TestArachnaphobiaCastCacheAndAllocation52DC80(t *testing.T) {
	for _, word := range []uint32{0, 11, 0x1000b, 0x7fffffff, 0x80000000, 0xffffffff} {
		for _, allocated := range []bool{false, true} {
			t.Run(fmt.Sprintf("cache-%08x/allocated-%t", word, allocated), func(t *testing.T) {
				cache := word
				focus, owner, caster := &Object{}, &Object{}, &Object{}
				arg := &SpellAcceptArg{Obj: caster, Pos: types.Ptf(1, 2)}
				beforeOwner, beforeCaster, beforeFocus, beforeArg := *owner, *caster, *focus, *arg
				lookups, stores, traces, allocations, placements := 0, 0, 0, 0, 0
				h := arachnaphobiaCastDeps52DC80{
					loadCache: func() uint32 { return cache },
					lookupType: func(key string) uint32 {
						lookups++
						if key != "ArachnaphobiaFocus" {
							t.Fatal(key)
						}
						return 11
					},
					storeCache: func(kind uint32) { stores++; cache = kind },
					traceRay: func(from, to types.Pointf, flags MapTraceFlags) bool {
						traces++
						if from != caster.PosVec || to != arg.Pos || flags != 9 {
							t.Fatal("wrong trace")
						}
						return true
					},
					newObject: func(kind uint32) *Object {
						allocations++
						if kind != cache {
							t.Fatalf("type=%08x want=%08x", kind, cache)
						}
						if allocated {
							return focus
						}
						return nil
					},
					createAt: func(got, by *Object, pos types.Pointf) {
						placements++
						if got != focus || by != owner || pos != arg.Pos {
							t.Fatal("wrong placement")
						}
					},
					inform: func(uint8, byte, int32) { t.Fatal("unexpected notification") },
				}
				got := arachnaphobiaCast52DC80(owner, caster, arg, h)
				wantLookup, wantPlacement := 0, 0
				if word == 0 {
					wantLookup = 1
				}
				if allocated {
					wantPlacement = 1
				}
				if got != 1 || lookups != wantLookup || stores != wantLookup || traces != 1 || allocations != 1 ||
					placements != wantPlacement || *owner != beforeOwner || *caster != beforeCaster || *focus != beforeFocus || *arg != beforeArg {
					t.Fatalf("result/callback counts=%d/%d/%d/%d/%d/%d", got, lookups, stores, traces, allocations, placements)
				}
			})
		}
	}
}

func TestArachnaphobiaCastBlockedLivePlayer52DC80(t *testing.T) {
	for _, initialPlayer := range []bool{false, true} {
		for _, livePlayer := range []bool{false, true} {
			for _, index := range []uint8{0, 31, 255} {
				t.Run(fmt.Sprintf("initial-%t/live-%t/index-%d", initialPlayer, livePlayer, index), func(t *testing.T) {
					player := arachnaphobiaCastAlloc52DC80(t, Player{PlayerInd: index})
					data := arachnaphobiaCastAlloc52DC80(t, PlayerUpdateData{Player: player})
					caster := arachnaphobiaCastAlloc52DC80(t, Object{ObjClass: object.ClassMonster})
					if initialPlayer {
						caster.ObjClass = object.ClassPlayer
					}
					arg := arachnaphobiaCastAlloc52DC80(t, SpellAcceptArg{Pos: types.Ptf(30, 40)})
					cache, calls := uint32(11), 0
					h := arachnaphobiaCastDeps52DC80{
						loadCache: func() uint32 { return cache },
						traceRay: func(types.Pointf, types.Pointf, MapTraceFlags) bool {
							caster.ObjClass = object.ClassMonster
							if livePlayer {
								caster.ObjClass = object.ClassPlayer
							}
							caster.UpdateData = unsafe.Pointer(data)
							return false
						},
						inform: func(got uint8, code byte, value int32) {
							calls++
							if got != index || code != 0 || value != 2 {
								t.Fatalf("inform=%d/%d/%d", got, code, value)
							}
						},
					}
					beforePlayer, beforeData, beforeArg := *player, *data, *arg
					got := arachnaphobiaCast52DC80(nil, caster, arg, h)
					wantCalls := 0
					if livePlayer {
						wantCalls = 1
					}
					if got != 0 || calls != wantCalls || cache != 11 || *player != beforePlayer || *data != beforeData || *arg != beforeArg {
						t.Fatalf("blocked result=%d calls=%d", got, calls)
					}
				})
			}
		}
	}
}

func TestArachnaphobiaCastZeroCacheRetries52DC80(t *testing.T) {
	cache, lookups, stores, traces := uint32(0), 0, 0, 0
	h := arachnaphobiaCastDeps52DC80{
		loadCache: func() uint32 { return cache },
		lookupType: func(key string) uint32 {
			lookups++
			if key != "ArachnaphobiaFocus" {
				t.Fatal(key)
			}
			return 0
		},
		storeCache: func(kind uint32) { stores++; cache = kind },
		traceRay:   func(types.Pointf, types.Pointf, MapTraceFlags) bool { traces++; return false },
	}
	for range 2 {
		if got := arachnaphobiaCast52DC80(nil, &Object{}, &SpellAcceptArg{}, h); got != 0 {
			t.Fatal(got)
		}
	}
	if cache != 0 || lookups != 2 || stores != 2 || traces != 2 {
		t.Fatalf("retry=%d/%d/%d/%d", cache, lookups, stores, traces)
	}
}

func TestArachnaphobiaCastFaultPrefix52DC80(t *testing.T) {
	for _, fault := range []string{"caster", "argument", "player-update", "player"} {
		t.Run(fault, func(t *testing.T) {
			var events []string
			cache := uint32(0)
			caster, arg := &Object{}, &SpellAcceptArg{}
			h := arachnaphobiaCastDeps52DC80{
				loadCache:  func() uint32 { events = append(events, "load"); return cache },
				lookupType: func(string) uint32 { events = append(events, "lookup"); return 11 },
				storeCache: func(kind uint32) { events = append(events, "store"); cache = kind },
				traceRay:   func(types.Pointf, types.Pointf, MapTraceFlags) bool { events = append(events, "trace"); return false },
			}
			want := []string{"load", "lookup", "store"}
			switch fault {
			case "caster":
				caster = nil
			case "argument":
				arg = nil
			case "player-update":
				caster.ObjClass = object.ClassPlayer
				want = append(want, "trace")
			case "player":
				caster.ObjClass = object.ClassPlayer
				caster.UpdateData = unsafe.Pointer(&PlayerUpdateData{})
				want = append(want, "trace")
			}
			panicked := false
			func() { defer func() { panicked = recover() != nil }(); arachnaphobiaCast52DC80(nil, caster, arg, h) }()
			if !panicked || cache != 11 || !reflect.DeepEqual(events, want) {
				t.Fatalf("fault=%t cache=%d events=%v want=%v", panicked, cache, events, want)
			}
		})
	}
}
