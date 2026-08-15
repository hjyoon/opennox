package server

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

func TestRespawnScheduler4EC720PrologueOrder(t *testing.T) {
	t.Run("lookup-zero-is-stored-before-game-flags-return", func(t *testing.T) {
		var events []string
		RespawnScheduler4EC720(RespawnSchedulerHooks4EC720[int, int, int, int]{
			LoadCrown: func() uint32 {
				events = append(events, "load-crown")
				return 0
			},
			LookupTypeByName: func(name string) uint32 {
				events = append(events, "lookup:"+name)
				return 0
			},
			StoreCrown: func(v uint32) {
				events = append(events, fmt.Sprintf("store-crown:%d", v))
			},
			GameFlagsCheck: func(mask uint32) uint32 {
				events = append(events, fmt.Sprintf("game-flags:%#x", mask))
				return 1
			},
		})
		want := []string{"load-crown", "lookup:Crown", "store-crown:0", "game-flags:0x1200"}
		if !reflect.DeepEqual(events, want) {
			t.Fatalf("events = %v, want %v", events, want)
		}
	})

	t.Run("head-is-loaded-before-allow-is-cleared", func(t *testing.T) {
		var events []string
		RespawnScheduler4EC720(RespawnSchedulerHooks4EC720[int, int, int, int]{
			LoadCrown: func() uint32 {
				events = append(events, "load-crown")
				return 7
			},
			GameFlagsCheck: func(mask uint32) uint32 {
				events = append(events, fmt.Sprintf("game-flags:%#x", mask))
				return 0
			},
			LoadHead: func() int {
				events = append(events, "load-head")
				return 0
			},
			StoreAllow: func(v uint32) {
				events = append(events, fmt.Sprintf("store-allow:%d", v))
			},
		})
		want := []string{"load-crown", "game-flags:0x1200", "load-head", "store-allow:0"}
		if !reflect.DeepEqual(events, want) {
			t.Fatalf("events = %v, want %v", events, want)
		}
	})
}

func TestRespawnSchedule4EC720UsesFPSBeforeFrameAndWraps(t *testing.T) {
	var events []string
	var gotAt uint32
	respawnSchedule4EC720(3, RespawnSchedulerHooks4EC720[int, int, int, int]{
		StorePending: func(rec int, value uint32) {
			events = append(events, fmt.Sprintf("store-pending:%d:%d", rec, value))
		},
		LoadFPS: func() uint32 {
			events = append(events, "load-fps")
			return 1
		},
		LoadFrame: func() uint32 {
			events = append(events, "load-frame")
			return 0xfffffff0
		},
		StoreRespawnAt: func(rec int, value uint32) {
			events = append(events, fmt.Sprintf("store-respawn-at:%d:%#x", rec, value))
			gotAt = value
		},
	})
	if gotAt != 0x0e {
		t.Fatalf("respawn at = %#x, want uint32-wrapped 0xe", gotAt)
	}
	want := []string{"store-pending:3:1", "load-fps", "load-frame", "store-respawn-at:3:0xe"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestRespawnClassify4EC720CoreBranches(t *testing.T) {
	tests := []struct {
		name           string
		object         int
		class          uint32
		flags          uint32
		allowed        bool
		wantObject     int
		wantPending    uint32
		wantRespawnAt  uint32
		wantEventParts []string
	}{
		{
			name:          "missing-object",
			object:        0,
			wantObject:    0,
			wantPending:   1,
			wantRespawnAt: 650,
			wantEventParts: []string{
				"load-object:1", "store-pending:1", "load-fps", "load-frame", "store-respawn-at:650",
			},
		},
		{
			name:          "destroyed-monster-clears-then-schedules",
			object:        4,
			class:         0x00000102,
			flags:         0x20,
			wantObject:    0,
			wantPending:   1,
			wantRespawnAt: 650,
			wantEventParts: []string{
				"load-object:1", "load-class:4", "load-flags:4", "store-object:0",
				"store-pending:1", "load-fps", "load-frame", "store-respawn-at:650",
			},
		},
		{
			name:          "no-update-monster-retains-object",
			object:        4,
			class:         0x00000102,
			flags:         0x8000,
			wantObject:    4,
			wantPending:   1,
			wantRespawnAt: 650,
			wantEventParts: []string{
				"load-object:1", "load-class:4", "load-flags:4",
				"store-pending:1", "load-fps", "load-frame", "store-respawn-at:650",
			},
		},
		{
			name:          "destroyed-nonmonster-clears-after-unit-def",
			object:        4,
			class:         0x00000100,
			flags:         0x20,
			allowed:       true,
			wantObject:    0,
			wantPending:   1,
			wantRespawnAt: 650,
			wantEventParts: []string{
				"load-object:1", "load-class:4", "load-flags:4", "load-type:4",
				"unit-def:9", "store-object:0", "store-pending:1", "load-fps", "load-frame", "store-respawn-at:650",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			object := tc.object
			pending := uint32(0)
			respawnAt := uint32(0)
			var events []string
			hooks := RespawnSchedulerHooks4EC720[int, int, int, int]{
				LoadObject: func(rec int) int {
					events = append(events, fmt.Sprintf("load-object:%d", rec))
					return object
				},
				StoreObject: func(rec, value int) {
					events = append(events, fmt.Sprintf("store-object:%d", value))
					object = value
				},
				LoadClass: func(obj int) uint32 {
					events = append(events, fmt.Sprintf("load-class:%d", obj))
					return tc.class
				},
				LoadFlags: func(obj int) uint32 {
					events = append(events, fmt.Sprintf("load-flags:%d", obj))
					return tc.flags
				},
				LoadObjectTypeInd: func(obj int) uint16 {
					events = append(events, fmt.Sprintf("load-type:%d", obj))
					return 9
				},
				UnitDefAllowed: func(ind uint32) bool {
					events = append(events, fmt.Sprintf("unit-def:%d", ind))
					if tc.name == "destroyed-nonmonster-clears-after-unit-def" && object != 4 {
						t.Fatalf("object cleared before UnitDef callback")
					}
					return tc.allowed
				},
				StorePending: func(rec int, value uint32) {
					events = append(events, fmt.Sprintf("store-pending:%d", value))
					pending = value
				},
				LoadFPS: func() uint32 {
					events = append(events, "load-fps")
					return 20
				},
				LoadFrame: func() uint32 {
					events = append(events, "load-frame")
					return 50
				},
				StoreRespawnAt: func(rec int, value uint32) {
					events = append(events, fmt.Sprintf("store-respawn-at:%d", value))
					respawnAt = value
				},
			}
			respawnClassify4EC720(1, hooks)
			if object != tc.wantObject || pending != tc.wantPending || respawnAt != tc.wantRespawnAt {
				t.Fatalf("object/pending/at = (%d, %d, %d), want (%d, %d, %d)", object, pending, respawnAt, tc.wantObject, tc.wantPending, tc.wantRespawnAt)
			}
			if !reflect.DeepEqual(events, tc.wantEventParts) {
				t.Fatalf("events = %v, want %v", events, tc.wantEventParts)
			}
		})
	}
}

func TestRespawnClassify4EC720LiveInventoryReloads(t *testing.T) {
	object := 1
	unitDefCalls := 0
	var events []string
	hooks := RespawnSchedulerHooks4EC720[int, int, int, int]{
		LoadObject: func(rec int) int {
			events = append(events, fmt.Sprintf("load-object:%d", object))
			return object
		},
		LoadClass: func(obj int) uint32 {
			events = append(events, fmt.Sprintf("load-class:%d", obj))
			return 0x1000
		},
		LoadFlags: func(obj int) uint32 {
			events = append(events, fmt.Sprintf("load-flags:%d", obj))
			return 0
		},
		LoadInvHolder: func(obj int) int {
			events = append(events, fmt.Sprintf("load-holder:%d", obj))
			if obj == 1 {
				return 0
			}
			return 90 + obj
		},
		LoadObjectTypeInd: func(obj int) uint16 {
			events = append(events, fmt.Sprintf("load-type:%d", obj))
			return uint16(10 + obj)
		},
		UnitDefAllowed: func(ind uint32) bool {
			unitDefCalls++
			events = append(events, fmt.Sprintf("unit-def:%d", ind))
			if unitDefCalls == 1 {
				object = 2
				return false
			}
			object = 3
			return true
		},
		LoadCrown: func() uint32 {
			events = append(events, "load-crown")
			return 99
		},
		SpecialModeCheck: func(mode uint32) uint32 {
			events = append(events, fmt.Sprintf("special:%d", mode))
			return 1
		},
		StorePending: func(rec int, value uint32) {
			events = append(events, fmt.Sprintf("store-pending:%d", value))
		},
		LoadFPS: func() uint32 {
			events = append(events, "load-fps")
			return 10
		},
		LoadFrame: func() uint32 {
			events = append(events, "load-frame")
			return 20
		},
		StoreRespawnAt: func(rec int, value uint32) {
			events = append(events, fmt.Sprintf("store-respawn-at:%d", value))
		},
	}

	respawnClassify4EC720(7, hooks)

	want := []string{
		"load-object:1", "load-class:1", "load-flags:1", "load-holder:1", "load-type:1", "unit-def:11",
		"load-object:2", "load-holder:2", "load-type:2", "unit-def:12",
		"load-object:3", "load-crown", "load-type:3", "special:2",
		"store-pending:1", "load-fps", "load-frame", "store-respawn-at:320",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events =\n%v\nwant =\n%v", events, want)
	}
}

func TestRespawnClassify4EC720CrownAndNoneligibleOrder(t *testing.T) {
	var events []string
	respawnClassify4EC720(1, RespawnSchedulerHooks4EC720[int, int, int, int]{
		LoadObject: func(int) int {
			events = append(events, "load-object")
			return 2
		},
		LoadClass: func(int) uint32 {
			events = append(events, "load-class")
			return 0
		},
		LoadFlags: func(int) uint32 {
			events = append(events, "load-flags")
			return 0
		},
		LoadCrown: func() uint32 {
			events = append(events, "load-crown")
			return 7
		},
		LoadObjectTypeInd: func(int) uint16 {
			events = append(events, "load-type")
			return 8
		},
		LoadInvHolder: func(int) int {
			events = append(events, "load-holder")
			return 3
		},
		StorePending: func(int, uint32) { events = append(events, "store-pending") },
		LoadFPS: func() uint32 {
			events = append(events, "load-fps")
			return 0
		},
		LoadFrame: func() uint32 {
			events = append(events, "load-frame")
			return 0
		},
		StoreRespawnAt: func(int, uint32) { events = append(events, "store-at") },
	})
	want := []string{
		"load-object", "load-class", "load-flags", "load-crown", "load-type", "load-holder",
		"store-pending", "load-fps", "load-frame", "store-at",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestRespawnRestoreMovedObject4EC720DistanceGate(t *testing.T) {
	tests := []struct {
		name string
		x    float32
		y    float32
		want bool
	}{
		{name: "equal-is-rejected", x: 50, want: false},
		{name: "unordered-is-rejected", x: float32(math.NaN()), want: false},
		{name: "greater-is-accepted", x: 50, y: 1, want: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			called := false
			respawnRestoreMovedObject4EC720(1, RespawnSchedulerHooks4EC720[int, int, int, int]{
				LoadFPS:     func() uint32 { return 1 },
				LoadObject:  func(int) int { return 2 },
				LoadField32: func(int) uint32 { return 0 },
				LoadFrame:   func() uint32 { return 6 },
				LoadRecordX: func(int) float32 { return tc.x },
				LoadObjectX: func(int) float32 { return 0 },
				LoadRecordY: func(int) float32 { return tc.y },
				LoadObjectY: func(int) float32 { return 0 },
				PointFXAtObject: func(uint32, int) {
					called = true
				},
				AudioAtObjectPosition: func(uint32, int) {},
				MoveToRecord:          func(int, int) {},
				LoadClass:             func(int) uint32 { return 0 },
				LoadHealthData:        func(int) int { return 0 },
				PointFXAtRecord:       func(uint32, int) {},
				AudioAtRecordPosition: func(uint32, int) {},
			})
			if called != tc.want {
				t.Fatalf("distance accepted = %v, want %v", called, tc.want)
			}
		})
	}
}

func TestRespawnRestoreMovedObject4EC720LiveObjectOrder(t *testing.T) {
	objectLoads := 0
	object := 10
	var events []string
	hooks := RespawnSchedulerHooks4EC720[int, int, int, int]{
		LoadFPS: func() uint32 {
			events = append(events, "load-fps")
			return 2
		},
		LoadObject: func(int) int {
			objectLoads++
			events = append(events, fmt.Sprintf("load-object:%d", object))
			return object
		},
		LoadField32: func(obj int) uint32 {
			events = append(events, fmt.Sprintf("load-field32:%d", obj))
			return 10
		},
		LoadFrame: func() uint32 {
			events = append(events, "load-frame")
			return 21
		},
		LoadRecordX: func(int) float32 {
			events = append(events, "load-rec-x")
			return 51
		},
		LoadObjectX: func(obj int) float32 {
			events = append(events, fmt.Sprintf("load-obj-x:%d", obj))
			return 0
		},
		LoadRecordY: func(int) float32 {
			events = append(events, "load-rec-y")
			return 0
		},
		LoadObjectY: func(obj int) float32 {
			events = append(events, fmt.Sprintf("load-obj-y:%d", obj))
			return 0
		},
		PointFXAtObject: func(code uint32, obj int) {
			events = append(events, fmt.Sprintf("point-object:%d:%d", code, obj))
			object = 20
		},
		AudioAtObjectPosition: func(code uint32, obj int) {
			events = append(events, fmt.Sprintf("audio-object-pos:%d:%d", code, obj))
			object = 30
		},
		MoveToRecord: func(obj, rec int) {
			events = append(events, fmt.Sprintf("move:%d:%d", obj, rec))
			object = 40
		},
		LoadClass: func(obj int) uint32 {
			events = append(events, fmt.Sprintf("load-class:%d", obj))
			return 0x01000000
		},
		WeaponEquipFlags: func(obj int) uint32 {
			events = append(events, fmt.Sprintf("weapon-flags:%d", obj))
			object = 50
			return 0x82
		},
		LoadCharge1: func(int) uint8 {
			events = append(events, "load-charge1")
			return 11
		},
		LoadUseData: func(obj int) int {
			events = append(events, fmt.Sprintf("load-use:%d", obj))
			return 500 + obj
		},
		StoreUseByte: func(use int, index uint32, value uint8) {
			events = append(events, fmt.Sprintf("store-use:%d:%d:%d", use, index, value))
		},
		LoadCharge0: func(int) uint8 {
			events = append(events, "load-charge0")
			return 10
		},
		LoadHealthData: func(obj int) int {
			events = append(events, fmt.Sprintf("load-health:%d", obj))
			return 70
		},
		LoadHealthMax: func(health int) uint16 {
			events = append(events, fmt.Sprintf("load-health-max:%d", health))
			return 123
		},
		SetHP: func(obj int, hp uint16) {
			events = append(events, fmt.Sprintf("set-hp:%d:%d", obj, hp))
		},
		PointFXAtRecord: func(code uint32, rec int) {
			events = append(events, fmt.Sprintf("point-record:%d:%d", code, rec))
		},
		AudioAtRecordPosition: func(code uint32, rec int) {
			events = append(events, fmt.Sprintf("audio-record-pos:%d:%d", code, rec))
		},
	}

	respawnRestoreMovedObject4EC720(7, hooks)

	want := []string{
		"load-fps", "load-object:10", "load-field32:10", "load-frame",
		"load-rec-x", "load-obj-x:10", "load-rec-y", "load-obj-y:10",
		"point-object:129:10", "load-object:20", "audio-object-pos:283:20",
		"load-object:30", "move:30:7", "load-object:40", "load-class:40", "weapon-flags:40",
		"load-object:50", "load-charge1", "load-use:50", "store-use:550:1:11",
		"load-charge0", "store-use:550:0:10", "load-object:50", "load-health:50",
		"load-health-max:70", "set-hp:50:123", "point-record:129:7", "audio-record-pos:283:7",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events =\n%v\nwant =\n%v", events, want)
	}
	if objectLoads != 6 {
		t.Fatalf("object loads = %d, want 6", objectLoads)
	}
}

func TestRespawnScheduler4EC720DuePathOrderAndLiveNext(t *testing.T) {
	recordType := uint32(7)
	newClassLoads := 0
	next := 0
	object := 4
	var events []string
	hooks := RespawnSchedulerHooks4EC720[int, int, int, int]{
		LoadCrown:      func() uint32 { events = append(events, "load-crown"); return 1 },
		GameFlagsCheck: func(uint32) uint32 { events = append(events, "game-flags"); return 0 },
		LoadHead:       func() int { events = append(events, "load-head"); return 1 },
		StoreAllow:     func(uint32) { events = append(events, "store-allow") },
		LoadPending:    func(int) uint32 { events = append(events, "load-pending"); return 1 },
		LoadFrame:      func() uint32 { events = append(events, "load-frame"); return 0 },
		LoadRespawnAt: func(rec int) uint32 {
			events = append(events, "load-at")
			if rec == 2 {
				return 0xffffffff
			}
			return 0
		},
		LoadRecordTypeInd: func(int) uint32 {
			events = append(events, fmt.Sprintf("load-record-type:%d", recordType))
			return recordType
		},
		UnitDefAllowed: func(ind uint32) bool {
			events = append(events, fmt.Sprintf("unit-def:%d", ind))
			recordType = 8
			return true
		},
		NewObjectByTypeInd: func(ind uint32) int {
			events = append(events, fmt.Sprintf("new-object:%d", ind))
			return 9
		},
		LoadRecordY: func(int) float32 { events = append(events, "load-y"); return 2 },
		LoadRecordX: func(int) float32 { events = append(events, "load-x"); return 1 },
		CreateAt: func(obj, owner int, x, y float32) {
			events = append(events, fmt.Sprintf("create:%d:%d:%.0f:%.0f", obj, owner, x, y))
		},
		PointFXAtRecord: func(code uint32, rec int) {
			events = append(events, fmt.Sprintf("point:%d:%d", code, rec))
		},
		LoadDirection: func(int) uint16 { events = append(events, "load-direction"); return 33 },
		StoreDirection1: func(obj int, value uint16) {
			events = append(events, fmt.Sprintf("store-direction1:%d:%d", obj, value))
		},
		StoreDirection2: func(obj int, value uint16) {
			events = append(events, fmt.Sprintf("store-direction2:%d:%d", obj, value))
		},
		LoadClass: func(obj int) uint32 {
			events = append(events, fmt.Sprintf("load-class:%d", obj))
			if obj == 9 {
				newClassLoads++
				if newClassLoads == 1 {
					return 0x1000
				}
				return 0x01000000
			}
			return 2
		},
		CopyModifierAttrs: func(obj, rec int) {
			events = append(events, fmt.Sprintf("copy-attrs:%d:%d", obj, rec))
		},
		WeaponEquipFlags: func(obj int) uint32 {
			events = append(events, fmt.Sprintf("weapon-flags:%d", obj))
			return 0x82
		},
		LoadUseData: func(obj int) int {
			events = append(events, fmt.Sprintf("load-use:%d", obj))
			return 90
		},
		LoadCharge1: func(int) uint8 { events = append(events, "load-charge1"); return 12 },
		LoadCharge0: func(int) uint8 { events = append(events, "load-charge0"); return 11 },
		StoreUseByte: func(use int, index uint32, value uint8) {
			events = append(events, fmt.Sprintf("store-use:%d:%d:%d", use, index, value))
		},
		AudioOnObject: func(code uint32, obj int) {
			events = append(events, fmt.Sprintf("audio-object:%d:%d", code, obj))
			next = 2
		},
		LoadObject: func(int) int {
			events = append(events, fmt.Sprintf("load-old-object:%d", object))
			return object
		},
		DelayedDelete: func(obj int) {
			events = append(events, fmt.Sprintf("delayed-delete:%d", obj))
		},
		StorePending: func(rec int, value uint32) {
			events = append(events, fmt.Sprintf("store-pending:%d:%d", rec, value))
		},
		StoreObject: func(rec, obj int) {
			events = append(events, fmt.Sprintf("store-object:%d:%d", rec, obj))
			object = obj
		},
		LoadNext: func(rec int) int {
			events = append(events, fmt.Sprintf("load-next:%d:%d", rec, next))
			if rec == 1 {
				return next
			}
			return 0
		},
	}

	RespawnScheduler4EC720(hooks)

	want := []string{
		"load-crown", "game-flags", "load-head", "store-allow", "load-pending",
		"load-frame", "load-at", "load-record-type:7", "unit-def:7", "load-record-type:8", "new-object:8",
		"load-y", "load-x", "create:9:0:1:2", "point:129:1", "load-direction",
		"store-direction1:9:33", "store-direction2:9:33", "load-class:9", "copy-attrs:9:1",
		"load-class:9", "weapon-flags:9", "load-use:9", "load-charge1", "store-use:90:1:12",
		"load-charge0", "store-use:90:0:11", "audio-object:283:9",
		"load-old-object:4", "load-class:4", "delayed-delete:4", "store-pending:1:0", "store-object:1:9",
		"load-next:1:2", "load-pending", "load-frame", "load-at", "load-next:2:2",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events =\n%#v\nwant =\n%#v", events, want)
	}
}

func TestRespawnScheduler4EC720UnsignedDueGateAndNilCreation(t *testing.T) {
	t.Run("unsigned-frame-before-deadline", func(t *testing.T) {
		var events []string
		RespawnScheduler4EC720(RespawnSchedulerHooks4EC720[int, int, int, int]{
			LoadCrown:      func() uint32 { return 1 },
			GameFlagsCheck: func(uint32) uint32 { return 0 },
			LoadHead:       func() int { return 1 },
			StoreAllow:     func(uint32) {},
			LoadPending:    func(int) uint32 { return 1 },
			LoadFrame: func() uint32 {
				events = append(events, "load-frame")
				return 0
			},
			LoadRespawnAt: func(int) uint32 {
				events = append(events, "load-at")
				return 0xffffffff
			},
			LoadNext: func(int) int {
				events = append(events, "load-next")
				return 0
			},
		})
		want := []string{"load-frame", "load-at", "load-next"}
		if !reflect.DeepEqual(events, want) {
			t.Fatalf("events = %v, want %v", events, want)
		}
	})

	t.Run("nil-new-object-still-replaces-and-cleans-old", func(t *testing.T) {
		var events []string
		RespawnScheduler4EC720(RespawnSchedulerHooks4EC720[int, int, int, int]{
			LoadCrown:          func() uint32 { return 1 },
			GameFlagsCheck:     func(uint32) uint32 { return 0 },
			LoadHead:           func() int { return 1 },
			StoreAllow:         func(uint32) {},
			LoadPending:        func(int) uint32 { return 1 },
			LoadFrame:          func() uint32 { return 4 },
			LoadRespawnAt:      func(int) uint32 { return 4 },
			LoadRecordTypeInd:  func(int) uint32 { return 7 },
			UnitDefAllowed:     func(uint32) bool { return true },
			NewObjectByTypeInd: func(uint32) int { events = append(events, "new:nil"); return 0 },
			LoadObject:         func(int) int { events = append(events, "load-old"); return 2 },
			LoadClass:          func(int) uint32 { events = append(events, "load-old-class"); return 2 },
			DelayedDelete:      func(int) { events = append(events, "delete-old") },
			StorePending:       func(int, uint32) { events = append(events, "clear-pending") },
			StoreObject:        func(rec, obj int) { events = append(events, fmt.Sprintf("store-object:%d", obj)) },
			LoadNext:           func(int) int { events = append(events, "load-next"); return 0 },
		})
		want := []string{"new:nil", "load-old", "load-old-class", "delete-old", "clear-pending", "store-object:0", "load-next"}
		if !reflect.DeepEqual(events, want) {
			t.Fatalf("events = %v, want %v", events, want)
		}
	})
}
