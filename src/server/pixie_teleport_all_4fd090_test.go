package server

import (
	"fmt"
	"reflect"
	"testing"
)

func TestPixieTeleportAll4FD090ExactTraceAndLiveSuccessor(t *testing.T) {
	const (
		owner          = uint64(0x1_0000_1000)
		mismatch       = uint64(0x2_0000_2000)
		dead           = uint64(0x3_0000_3000)
		targeted       = uint64(0x4_0000_4000)
		eligible       = uint64(0x5_0000_5000)
		originalEnd    = uint64(0x6_0000_6000)
		replacement    = uint64(0x7_0000_7000)
		targetedUpdate = uint64(0x8_0000_8000)
		target         = uint64(0x9_0000_9000)
		eligibleUpdate = uint64(0xa_0000_a000)
		pixieType      = uint32(0x1234)
	)
	names := map[uint64]string{
		owner:          "owner",
		mismatch:       "mismatch",
		dead:           "dead",
		targeted:       "targeted",
		eligible:       "eligible",
		originalEnd:    "original-end",
		replacement:    "replacement",
		targetedUpdate: "targeted-update",
		target:         "target",
		eligibleUpdate: "eligible-update",
	}
	next := map[uint64]uint64{
		mismatch: dead,
		dead:     targeted,
		targeted: eligible,
		eligible: originalEnd,
	}
	typeInd := map[uint64]uint16{
		mismatch:    uint16(pixieType + 1),
		dead:        uint16(pixieType),
		targeted:    uint16(pixieType),
		eligible:    uint16(pixieType),
		replacement: uint16(pixieType + 1),
		originalEnd: uint16(pixieType),
	}
	flags := map[uint64]uint32{
		dead:     pixieDeadFlag4FD090,
		targeted: ^pixieDeadFlag4FD090,
		eligible: ^pixieDeadFlag4FD090,
	}
	updates := map[uint64]uint64{targeted: targetedUpdate, eligible: eligibleUpdate}
	targets := map[uint64]uint64{targetedUpdate: target}
	var events []string
	teleports := 0

	pixieTeleportAll4FD090(pixieTeleportAllHooks4FD090[uint64, uint64]{
		loadOwnerArg: func() uint64 {
			events = append(events, "owner-arg")
			return owner
		},
		loadFirstOwned: func(got uint64) uint64 {
			events = append(events, "first:"+names[got])
			return mismatch
		},
		loadPixieTypeID: func() uint32 {
			events = append(events, "type-id")
			return pixieType
		},
		loadTypeInd: func(got uint64) uint16 {
			events = append(events, "type:"+names[got])
			return typeInd[got]
		},
		loadFlags: func(got uint64) uint32 {
			events = append(events, "flags:"+names[got])
			return flags[got]
		},
		loadUpdateData: func(got uint64) uint64 {
			events = append(events, "update:"+names[got])
			return updates[got]
		},
		loadTarget: func(got uint64) uint64 {
			events = append(events, "target:"+names[got])
			if got != targetedUpdate && got != eligibleUpdate {
				t.Fatalf("unexpected update token %#x", got)
			}
			return targets[got]
		},
		teleport: func(gotPixie, gotOwner uint64) {
			events = append(events, "teleport:"+names[gotPixie]+":"+names[gotOwner])
			teleports++
			if gotPixie != eligible || gotOwner != owner {
				t.Fatalf("teleport = (%#x, %#x), want (%#x, %#x)", gotPixie, gotOwner, eligible, owner)
			}
			next[eligible] = replacement
		},
		loadNextOwned: func(got uint64) uint64 {
			events = append(events, "next:"+names[got])
			return next[got]
		},
	})

	want := []string{
		"owner-arg", "first:owner",
		"type-id", "type:mismatch", "next:mismatch",
		"type-id", "type:dead", "flags:dead", "next:dead",
		"type-id", "type:targeted", "flags:targeted", "update:targeted", "target:targeted-update", "next:targeted",
		"type-id", "type:eligible", "flags:eligible", "update:eligible", "target:eligible-update", "teleport:eligible:owner", "next:eligible",
		"type-id", "type:replacement", "next:replacement",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
	if teleports != 1 {
		t.Fatalf("teleports = %d, want 1", teleports)
	}
	for _, token := range []uint64{owner, mismatch, dead, targeted, eligible, replacement} {
		if token <= uint64(^uint32(0)) {
			t.Fatalf("token %#x did not retain native-width high bits", token)
		}
	}
}

func TestPixieTeleportAll4FD090ReloadsAndComparesFullTypeID(t *testing.T) {
	const (
		owner  = uint64(0x1_0000_0001)
		first  = uint64(0x2_0000_0002)
		second = uint64(0x3_0000_0003)
	)
	loads := 0
	var teleported []uint64
	pixieTeleportAll4FD090(pixieTeleportAllHooks4FD090[uint64, uint64]{
		loadOwnerArg:   func() uint64 { return owner },
		loadFirstOwned: func(uint64) uint64 { return first },
		loadPixieTypeID: func() uint32 {
			loads++
			if loads == 1 {
				return 0x10001
			}
			return 2
		},
		loadTypeInd: func(got uint64) uint16 {
			if got == first {
				return 1
			}
			return 2
		},
		loadFlags:      func(uint64) uint32 { return 0 },
		loadUpdateData: func(uint64) uint64 { return 0x4_0000_0004 },
		loadTarget:     func(uint64) uint64 { return 0 },
		teleport: func(got, gotOwner uint64) {
			if gotOwner != owner {
				t.Fatalf("owner = %#x, want %#x", gotOwner, owner)
			}
			teleported = append(teleported, got)
		},
		loadNextOwned: func(got uint64) uint64 {
			if got == first {
				return second
			}
			return 0
		},
	})
	if loads != 2 {
		t.Fatalf("type ID loads = %d, want 2", loads)
	}
	if want := []uint64{second}; !reflect.DeepEqual(teleported, want) {
		t.Fatalf("teleported = %#v, want %#v", teleported, want)
	}
}

func TestPixieTeleportAll4FD090ZeroTypeIDCanMatch(t *testing.T) {
	teleports := 0
	pixieTeleportAll4FD090(pixieTeleportAllHooks4FD090[int, int]{
		loadOwnerArg:    func() int { return 1 },
		loadFirstOwned:  func(int) int { return 2 },
		loadPixieTypeID: func() uint32 { return 0 },
		loadTypeInd:     func(int) uint16 { return 0 },
		loadFlags:       func(int) uint32 { return 0 },
		loadUpdateData:  func(int) int { return 3 },
		loadTarget:      func(int) int { return 0 },
		teleport:        func(int, int) { teleports++ },
		loadNextOwned:   func(int) int { return 0 },
	})
	if teleports != 1 {
		t.Fatalf("teleports = %d, want zero TypeInd to match zero type ID", teleports)
	}
}

func TestPixieTeleportAll4FD090EmptyListStopsAfterFirstLoad(t *testing.T) {
	var events []string
	pixieTeleportAll4FD090(pixieTeleportAllHooks4FD090[int, int]{
		loadOwnerArg: func() int {
			events = append(events, "owner")
			return 7
		},
		loadFirstOwned: func(owner int) int {
			events = append(events, fmt.Sprintf("first:%d", owner))
			return 0
		},
		loadPixieTypeID: func() uint32 { t.Fatal("empty list loaded type ID"); return 0 },
		loadTypeInd:     func(int) uint16 { t.Fatal("empty list loaded TypeInd"); return 0 },
		loadFlags:       func(int) uint32 { t.Fatal("empty list loaded flags"); return 0 },
		loadUpdateData:  func(int) int { t.Fatal("empty list loaded update data"); return 0 },
		loadTarget:      func(int) int { t.Fatal("empty list loaded target"); return 0 },
		teleport:        func(int, int) { t.Fatal("empty list teleported") },
		loadNextOwned:   func(int) int { t.Fatal("empty list loaded successor"); return 0 },
	})
	if want := []string{"owner", "first:7"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestPixieTeleportAll4FD090AllFaultPrefixes(t *testing.T) {
	want := []string{
		"owner-arg", "first-owned", "type-id", "type-ind", "flags",
		"update-data", "target", "teleport", "next-owned",
	}
	for failAt := 1; failAt <= len(want); failAt++ {
		t.Run(fmt.Sprintf("step_%02d", failAt), func(t *testing.T) {
			var events []string
			emit := func(event string) {
				events = append(events, event)
				if len(events) == failAt {
					panic(event)
				}
			}
			hooks := pixieTeleportAllHooks4FD090[int, int]{
				loadOwnerArg: func() int {
					emit("owner-arg")
					return 1
				},
				loadFirstOwned: func(int) int {
					emit("first-owned")
					return 2
				},
				loadPixieTypeID: func() uint32 {
					emit("type-id")
					return 3
				},
				loadTypeInd: func(int) uint16 {
					emit("type-ind")
					return 3
				},
				loadFlags: func(int) uint32 {
					emit("flags")
					return 0
				},
				loadUpdateData: func(int) int {
					emit("update-data")
					return 4
				},
				loadTarget: func(int) int {
					emit("target")
					return 0
				},
				teleport: func(int, int) { emit("teleport") },
				loadNextOwned: func(int) int {
					emit("next-owned")
					return 0
				},
			}
			func() {
				defer func() {
					if recover() == nil {
						t.Fatalf("step %d did not panic", failAt)
					}
				}()
				pixieTeleportAll4FD090(hooks)
			}()
			if expected := want[:failAt]; !reflect.DeepEqual(events, expected) {
				t.Fatalf("events = %v, want prefix %v", events, expected)
			}
		})
	}
}
