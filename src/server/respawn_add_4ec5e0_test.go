package server

import (
	"reflect"
	"testing"
)

type respawnAddObject4EC5E0 struct {
	name string
}

func TestRespawnAdd4EC5E0GateAndAllocationFailure(t *testing.T) {
	obj := &respawnAddObject4EC5E0{name: "object"}
	t.Run("disabled", func(t *testing.T) {
		var events []string
		got := RespawnAdd4EC5E0(obj, RespawnAddHooks4EC5E0[*respawnAddObject4EC5E0, int, string, string]{
			LoadAllow: func() uint32 {
				events = append(events, "load-allow")
				return 0
			},
		})
		if got != 0 {
			t.Fatalf("result = %d, want zero", got)
		}
		if want := []string{"load-allow"}; !reflect.DeepEqual(events, want) {
			t.Fatalf("events = %#v, want %#v", events, want)
		}
	})

	t.Run("allocation-failure", func(t *testing.T) {
		allocator := "original"
		var events []string
		got := RespawnAdd4EC5E0(obj, RespawnAddHooks4EC5E0[*respawnAddObject4EC5E0, int, string, string]{
			LoadAllow: func() uint32 {
				events = append(events, "load-allow")
				return 0x80000000
			},
			LoadAllocator: func() string {
				events = append(events, "load-allocator")
				return allocator
			},
			AllocZero: func(got string) int {
				events = append(events, "alloc-zero:"+got)
				allocator = "replacement"
				return 0
			},
		})
		if got != 0 {
			t.Fatalf("result = %d, want zero", got)
		}
		want := []string{"load-allow", "load-allocator", "alloc-zero:original"}
		if !reflect.DeepEqual(events, want) {
			t.Fatalf("events = %#v, want %#v", events, want)
		}
	})
}

func TestRespawnAdd4EC5E0OrderAndLiveReloads(t *testing.T) {
	obj := &respawnAddObject4EC5E0{name: "object"}
	const rec = 1
	const firstHead = 2
	const secondHead = 3
	class := uint32(0x1000)
	useData := "entry-use"
	headLoad := 0
	var events []string

	got := RespawnAdd4EC5E0(obj, RespawnAddHooks4EC5E0[*respawnAddObject4EC5E0, int, string, string]{
		LoadAllow: func() uint32 {
			events = append(events, "load-allow")
			return 7
		},
		LoadAllocator: func() string {
			events = append(events, "load-allocator")
			return "allocator"
		},
		AllocZero: func(allocator string) int {
			events = append(events, "alloc-zero:"+allocator)
			return rec
		},
		LoadTypeInd: func(got *respawnAddObject4EC5E0) uint16 {
			events = append(events, "load-type:"+got.name)
			return 0xfedc
		},
		StoreObject: func(gotRec int, gotObj *respawnAddObject4EC5E0) {
			events = append(events, "store-object")
			if gotRec != rec || gotObj != obj {
				t.Fatalf("store object = (%d, %p), want (%d, %p)", gotRec, gotObj, rec, obj)
			}
		},
		StoreTypeInd: func(gotRec int, value uint32) {
			events = append(events, "store-type")
			if gotRec != rec || value != 0xfedc {
				t.Fatalf("store type = (%d, %#x), want (%d, %#x)", gotRec, value, rec, uint32(0xfedc))
			}
		},
		LoadPositionXBits: func(*respawnAddObject4EC5E0) uint32 {
			events = append(events, "load-x")
			return 0x7fa54321
		},
		StorePositionXBits: func(gotRec int, value uint32) {
			events = append(events, "store-x")
			if gotRec != rec || value != 0x7fa54321 {
				t.Fatalf("store x = (%d, %#x)", gotRec, value)
			}
		},
		LoadPositionYBits: func(*respawnAddObject4EC5E0) uint32 {
			events = append(events, "load-y")
			return 0x80000000
		},
		StorePositionYBits: func(gotRec int, value uint32) {
			events = append(events, "store-y")
			if gotRec != rec || value != 0x80000000 {
				t.Fatalf("store y = (%d, %#x)", gotRec, value)
			}
		},
		LoadDirection: func(*respawnAddObject4EC5E0) uint16 {
			events = append(events, "load-direction")
			return 0xf123
		},
		StoreDirection: func(gotRec int, value uint16) {
			events = append(events, "store-direction")
			if gotRec != rec || value != 0xf123 {
				t.Fatalf("store direction = (%d, %#x)", gotRec, value)
			}
		},
		LoadClass: func(*respawnAddObject4EC5E0) uint32 {
			events = append(events, "load-class")
			return class
		},
		CopyModifierAttrs: func(gotRec int, gotObj *respawnAddObject4EC5E0) {
			events = append(events, "copy-attrs")
			if gotRec != rec || gotObj != obj {
				t.Fatalf("copy attrs = (%d, %p)", gotRec, gotObj)
			}
			class = 0x01000000
		},
		WeaponEquipFlags: func(gotObj *respawnAddObject4EC5E0) uint32 {
			events = append(events, "weapon-flags")
			if gotObj != obj {
				t.Fatalf("weapon object = %p, want %p", gotObj, obj)
			}
			useData = "live-use"
			return 0x80000082
		},
		LoadUseData: func(*respawnAddObject4EC5E0) string {
			events = append(events, "load-use-data:"+useData)
			return useData
		},
		LoadUseByte: func(gotUse string, index uint32) uint8 {
			events = append(events, "load-use-byte")
			if gotUse != "live-use" {
				t.Fatalf("use data = %q, want live-use", gotUse)
			}
			return map[uint32]uint8{0: 0x10, 1: 0x11}[index]
		},
		StoreCharge1: func(gotRec int, value uint8) {
			events = append(events, "store-charge-1")
			if gotRec != rec || value != 0x11 {
				t.Fatalf("charge 1 = (%d, %#x)", gotRec, value)
			}
		},
		StoreCharge0: func(gotRec int, value uint8) {
			events = append(events, "store-charge-0")
			if gotRec != rec || value != 0x10 {
				t.Fatalf("charge 0 = (%d, %#x)", gotRec, value)
			}
		},
		StorePrev: func(target, value int) {
			events = append(events, "store-prev")
			switch {
			case target == rec && value == 0:
			case target == secondHead && value == rec:
			default:
				t.Fatalf("store prev = (%d, %d)", target, value)
			}
		},
		LoadHead: func() int {
			headLoad++
			events = append(events, "load-head")
			if headLoad == 1 {
				return firstHead
			}
			return secondHead
		},
		StoreNext: func(target, value int) {
			events = append(events, "store-next")
			if target != rec || value != firstHead {
				t.Fatalf("store next = (%d, %d), want (%d, %d)", target, value, rec, firstHead)
			}
		},
		StoreHead: func(value int) {
			events = append(events, "store-head")
			if value != rec {
				t.Fatalf("store head = %d, want %d", value, rec)
			}
		},
	})

	if got != secondHead {
		t.Fatalf("result = %d, want second live head %d", got, secondHead)
	}
	want := []string{
		"load-allow", "load-allocator", "alloc-zero:allocator",
		"load-type:object", "store-object", "store-type",
		"load-x", "store-x", "load-y", "store-y", "load-direction", "store-direction",
		"load-class", "copy-attrs", "load-class", "weapon-flags",
		"load-use-data:live-use", "load-use-byte", "store-charge-1", "load-use-byte", "store-charge-0",
		"store-prev", "load-head", "store-next", "load-head", "store-prev", "store-head",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events =\n%#v\nwant =\n%#v", events, want)
	}
}

func TestRespawnAdd4EC5E0IndependentHeadLoads(t *testing.T) {
	tests := []struct {
		name       string
		first      int
		second     int
		wantReturn int
		wantPrev   bool
	}{
		{name: "first-only", first: 2, second: 0, wantReturn: 0, wantPrev: false},
		{name: "second-only", first: 0, second: 3, wantReturn: 3, wantPrev: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			headLoads := 0
			prevLinked := false
			got := RespawnAdd4EC5E0(1, RespawnAddHooks4EC5E0[int, int, int, int]{
				LoadAllow:          func() uint32 { return 1 },
				LoadAllocator:      func() int { return 1 },
				AllocZero:          func(int) int { return 1 },
				LoadTypeInd:        func(int) uint16 { return 0 },
				StoreObject:        func(int, int) {},
				StoreTypeInd:       func(int, uint32) {},
				LoadPositionXBits:  func(int) uint32 { return 0 },
				StorePositionXBits: func(int, uint32) {},
				LoadPositionYBits:  func(int) uint32 { return 0 },
				StorePositionYBits: func(int, uint32) {},
				LoadDirection:      func(int) uint16 { return 0 },
				StoreDirection:     func(int, uint16) {},
				LoadClass:          func(int) uint32 { return 0 },
				StorePrev: func(target, value int) {
					if target != 1 || value != 0 {
						prevLinked = true
						if target != tc.second || value != 1 {
							t.Fatalf("store prev = (%d, %d)", target, value)
						}
					}
				},
				LoadHead: func() int {
					headLoads++
					if headLoads == 1 {
						return tc.first
					}
					return tc.second
				},
				StoreNext: func(target, value int) {
					if target != 1 || value != tc.first {
						t.Fatalf("store next = (%d, %d)", target, value)
					}
				},
				StoreHead: func(value int) {
					if value != 1 {
						t.Fatalf("store head = %d", value)
					}
				},
			})
			if got != tc.wantReturn || prevLinked != tc.wantPrev || headLoads != 2 {
				t.Fatalf("result/prev/head-loads = (%d, %v, %d), want (%d, %v, 2)", got, prevLinked, headLoads, tc.wantReturn, tc.wantPrev)
			}
		})
	}
}

func TestRespawnAdd4EC5E0FaultSkipsFinalHeadStore(t *testing.T) {
	stop := &struct{}{}
	var events []string
	var recovered any
	func() {
		defer func() { recovered = recover() }()
		RespawnAdd4EC5E0(1, RespawnAddHooks4EC5E0[int, int, int, int]{
			LoadAllow:          func() uint32 { return 1 },
			LoadAllocator:      func() int { return 1 },
			AllocZero:          func(int) int { return 1 },
			LoadTypeInd:        func(int) uint16 { return 0 },
			StoreObject:        func(int, int) {},
			StoreTypeInd:       func(int, uint32) {},
			LoadPositionXBits:  func(int) uint32 { return 0 },
			StorePositionXBits: func(int, uint32) {},
			LoadPositionYBits:  func(int) uint32 { return 0 },
			StorePositionYBits: func(int, uint32) {},
			LoadDirection:      func(int) uint16 { return 0 },
			StoreDirection:     func(int, uint16) {},
			LoadClass:          func(int) uint32 { return 0 },
			StorePrev: func(target, value int) {
				events = append(events, "store-prev")
				if target == 2 {
					panic(stop)
				}
			},
			LoadHead: func() int {
				events = append(events, "load-head")
				return 2
			},
			StoreNext: func(int, int) { events = append(events, "store-next") },
			StoreHead: func(int) { events = append(events, "store-head") },
		})
	}()
	if recovered != stop {
		t.Fatalf("recovered = %#v, want sentinel", recovered)
	}
	want := []string{"store-prev", "load-head", "store-next", "load-head", "store-prev"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %#v, want %#v", events, want)
	}
}
