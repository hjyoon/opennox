package legacy

import (
	"fmt"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
	"github.com/opennox/opennox/v1/server"
)

func TestRespawnRechargeItem4EC720FixedPayload(t *testing.T) {
	if got := unsafe.Sizeof(respawnRechargeData4EC720{}); got != 116 {
		t.Fatalf("payload size = %d, want 116", got)
	}
	if got := unsafe.Offsetof(respawnRechargeData4EC720{}.Charge); got != 108 {
		t.Fatalf("charge offset = %d, want 108", got)
	}
	if got := unsafe.Offsetof(respawnRechargeData4EC720{}.MaxCharge); got != 109 {
		t.Fatalf("max charge offset = %d, want 109", got)
	}
	if got := unsafe.Offsetof(respawnRechargeData4EC720{}.Progress); got != 112 {
		t.Fatalf("progress offset = %d, want 112", got)
	}

	tests := []struct {
		name         string
		data         *respawnRechargeData4EC720
		amount       uint32
		wantChanged  bool
		wantProgress uint32
		wantCharge   uint8
	}{
		{name: "nil-use-data", data: nil, amount: 100},
		{
			name:         "already-full",
			data:         &respawnRechargeData4EC720{Progress: 100, Charge: 3, MaxCharge: 9},
			amount:       100,
			wantProgress: 100,
			wantCharge:   3,
		},
		{
			name:         "partial-progress-and-charge-change",
			data:         &respawnRechargeData4EC720{Progress: 10, MaxCharge: 9},
			amount:       20,
			wantChanged:  true,
			wantProgress: 30,
			wantCharge:   2,
		},
		{
			name:         "progress-changes-with-same-charge",
			data:         &respawnRechargeData4EC720{Progress: 20, Charge: 2, MaxCharge: 10},
			amount:       5,
			wantProgress: 25,
			wantCharge:   2,
		},
		{
			name:         "caps-at-one-hundred",
			data:         &respawnRechargeData4EC720{Progress: 95, Charge: 6, MaxCharge: 7},
			amount:       10,
			wantChanged:  true,
			wantProgress: 100,
			wantCharge:   7,
		},
		{
			name:         "signed-negative-progress",
			data:         &respawnRechargeData4EC720{Progress: 0xffffffff, MaxCharge: 10},
			amount:       100,
			wantChanged:  true,
			wantProgress: 99,
			wantCharge:   9,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			obj := &server.Object{}
			if tc.data != nil {
				obj.UseData.Ptr = unsafe.Pointer(tc.data)
			}
			changed := respawnRechargeItem4EC720(obj, tc.amount)
			if changed != tc.wantChanged {
				t.Fatalf("changed = %v, want %v", changed, tc.wantChanged)
			}
			if tc.data != nil && (tc.data.Progress != tc.wantProgress || tc.data.Charge != tc.wantCharge) {
				t.Fatalf("progress/charge = (%#x, %d), want (%#x, %d)", tc.data.Progress, tc.data.Charge, tc.wantProgress, tc.wantCharge)
			}
		})
	}
}

func withRespawnSchedulerNative4EC720(t *testing.T, fn func()) {
	t.Helper()
	handles.Init()
	t.Cleanup(handles.Release)

	oldAllow := respawnAddLoadAllow4EC5E0()
	oldHead := respawnAddLoadHead4EC5E0()
	oldCrown := respawnSchedulerLoadCrown4EC720()
	defer func() {
		respawnSchedulerStoreCrown4EC720(oldCrown)
		respawnAddStoreHead4EC5E0(oldHead)
		respawnAddStoreAllow4EC5E0(oldAllow)
	}()
	fn()
}

func TestRespawnSchedulerNative4EC720DueRecord(t *testing.T) {
	withRespawnSchedulerNative4EC720(t, func() {
		rec, freeRec := alloc.New(respawnRecord4EC5E0{})
		defer freeRec()
		oldObj, freeOld := alloc.New(server.Object{})
		defer freeOld()
		newObj, freeNew := alloc.New(server.Object{})
		defer freeNew()
		useData, freeUseData := alloc.New([2]uint8{})
		defer freeUseData()

		rec.TypeInd = 7
		rec.Object = oldObj
		rec.X = 12
		rec.Y = 34
		rec.Direction = 0x9876
		rec.RespawnAt = 10
		rec.Pending = 1
		rec.Attrs.Field16 = 0x12345678
		rec.Charge1 = 0x56
		rec.Charge0 = 0x34
		oldObj.ObjClass = object.ClassMonster
		newObj.ObjClass = object.ClassWeapon | object.ClassWand
		newObj.UseData.Ptr = unsafe.Pointer(useData)
		respawnAddStoreHead4EC5E0(rec)
		respawnAddStoreAllow4EC5E0(99)
		respawnSchedulerStoreCrown4EC720(0)

		var events []string
		respawnSchedulerNative4EC720(respawnSchedulerRuntime4EC720{
			LookupTypeByName: func(name string) uint32 {
				events = append(events, "lookup:"+name)
				return 5
			},
			GameFlagsCheck: func(mask uint32) uint32 {
				events = append(events, fmt.Sprintf("game-flags:%#x", mask))
				return 0
			},
			UnitDefAllowed: func(ind uint32) bool {
				events = append(events, fmt.Sprintf("unit-def:%d", ind))
				return true
			},
			NewObjectByTypeInd: func(ind uint32) *server.Object {
				events = append(events, fmt.Sprintf("new-object:%d", ind))
				return newObj
			},
			SpecialModeCheck: func(uint32) uint32 {
				t.Fatal("due record checked special mode")
				return 0
			},
			FPS: func() uint32 {
				t.Fatal("due record loaded FPS")
				return 0
			},
			Frame: func() uint32 {
				events = append(events, "frame")
				return 10
			},
			PointFX: func(code uint32, pos types.Pointf) {
				events = append(events, fmt.Sprintf("point:%d:%.0f:%.0f", code, pos.X, pos.Y))
			},
			AudioAtPosition: func(uint32, types.Pointf) {
				t.Fatal("due record emitted positional audio")
			},
			AudioOnObject: func(code uint32, obj *server.Object) {
				events = append(events, fmt.Sprintf("audio:%d:%t", code, obj == newObj))
			},
			MoveTo: func(*server.Object, types.Pointf) {
				t.Fatal("due record moved old object")
			},
			WeaponEquipFlags: func(obj *server.Object) uint32 {
				events = append(events, fmt.Sprintf("weapon:%t", obj == newObj))
				return 0x82
			},
			SetHP: func(*server.Object, uint16) {
				t.Fatal("due record set old HP")
			},
			CreateAt: func(obj, owner *server.Object, pos types.Pointf) {
				events = append(events, fmt.Sprintf("create:%t:%t:%.0f:%.0f", obj == newObj, owner == nil, pos.X, pos.Y))
			},
			CopyModifierAttrs: func(obj *server.Object, attrs *server.ModifierInitData) {
				events = append(events, fmt.Sprintf("copy:%t:%#x", obj == newObj, attrs.Field16))
			},
			DelayedDelete: func(obj *server.Object) {
				events = append(events, fmt.Sprintf("delete:%t", obj == oldObj))
			},
		})

		wantEvents := []string{
			"lookup:Crown", "game-flags:0x1200", "frame", "unit-def:7", "new-object:7",
			"create:true:true:12:34", "point:129:12:34", "copy:true:0x12345678",
			"weapon:true", "audio:283:true", "delete:true",
		}
		if !reflect.DeepEqual(events, wantEvents) {
			t.Fatalf("events = %v, want %v", events, wantEvents)
		}
		if got := respawnSchedulerLoadCrown4EC720(); got != 5 {
			t.Fatalf("Crown cache = %d, want 5", got)
		}
		if got := respawnAddLoadAllow4EC5E0(); got != 0 {
			t.Fatalf("allow = %d, want 0", got)
		}
		if rec.Pending != 0 || rec.Object != newObj {
			t.Fatalf("record pending/object = (%d, %p), want (0, %p)", rec.Pending, rec.Object, newObj)
		}
		if newObj.Direction1 != server.Dir16(rec.Direction) || newObj.Direction2 != server.Dir16(rec.Direction) {
			t.Fatalf("new directions = (%#x, %#x), want %#x", newObj.Direction1, newObj.Direction2, rec.Direction)
		}
		if useData[1] != rec.Charge1 || useData[0] != rec.Charge0 {
			t.Fatalf("new charges = (%#x, %#x), want (%#x, %#x)", useData[1], useData[0], rec.Charge1, rec.Charge0)
		}
	})
}

func TestRespawnSchedulerNative4EC720RestoresMovedWand(t *testing.T) {
	withRespawnSchedulerNative4EC720(t, func() {
		rec, freeRec := alloc.New(respawnRecord4EC5E0{})
		defer freeRec()
		obj, freeObj := alloc.New(server.Object{})
		defer freeObj()
		data, freeData := alloc.New(respawnRechargeData4EC720{})
		defer freeData()
		health, freeHealth := alloc.New(server.HealthData{})
		defer freeHealth()

		rec.TypeInd = 8
		rec.Object = obj
		rec.X = 51
		rec.Y = 0
		obj.TypeInd = 8
		obj.ObjClass = object.ClassWand
		obj.Field32 = 0
		obj.PosVec = types.Pointf{}
		obj.UseData.Ptr = unsafe.Pointer(data)
		obj.HealthData = health
		data.MaxCharge = 12
		health.Max = 77
		respawnAddStoreHead4EC5E0(rec)
		respawnAddStoreAllow4EC5E0(1)
		respawnSchedulerStoreCrown4EC720(8)

		var events []string
		respawnSchedulerNative4EC720(respawnSchedulerRuntime4EC720{
			LookupTypeByName: func(string) uint32 {
				t.Fatal("populated Crown cache was looked up")
				return 0
			},
			GameFlagsCheck: func(uint32) uint32 { return 0 },
			UnitDefAllowed: func(ind uint32) bool {
				events = append(events, fmt.Sprintf("unit-def:%d", ind))
				return true
			},
			NewObjectByTypeInd: func(uint32) *server.Object {
				t.Fatal("nonpending wand was respawned")
				return nil
			},
			SpecialModeCheck: func(uint32) uint32 { return 0 },
			FPS:              func() uint32 { return 1 },
			Frame:            func() uint32 { return 6 },
			PointFX: func(code uint32, pos types.Pointf) {
				events = append(events, fmt.Sprintf("point:%d:%.0f:%.0f", code, pos.X, pos.Y))
			},
			AudioAtPosition: func(code uint32, pos types.Pointf) {
				events = append(events, fmt.Sprintf("audio-pos:%d:%.0f:%.0f", code, pos.X, pos.Y))
			},
			AudioOnObject: func(uint32, *server.Object) {
				t.Fatal("nonpending wand emitted object audio")
			},
			MoveTo: func(got *server.Object, pos types.Pointf) {
				events = append(events, fmt.Sprintf("move:%t:%.0f:%.0f", got == obj, pos.X, pos.Y))
				got.PosVec = pos
			},
			WeaponEquipFlags: func(*server.Object) uint32 {
				t.Fatal("wand recharge path checked weapon flags")
				return 0
			},
			SetHP: func(got *server.Object, hp uint16) {
				events = append(events, fmt.Sprintf("set-hp:%t:%d", got == obj, hp))
			},
			CreateAt: func(*server.Object, *server.Object, types.Pointf) {
				t.Fatal("nonpending wand was created")
			},
			CopyModifierAttrs: func(*server.Object, *server.ModifierInitData) {
				t.Fatal("nonpending wand copied modifiers")
			},
			DelayedDelete: func(*server.Object) {
				t.Fatal("nonpending wand was deleted")
			},
		})

		wantEvents := []string{
			"unit-def:8", "point:129:0:0", "audio-pos:283:0:0", "move:true:51:0",
			"set-hp:true:77", "point:129:51:0", "audio-pos:283:51:0",
		}
		if !reflect.DeepEqual(events, wantEvents) {
			t.Fatalf("events = %v, want %v", events, wantEvents)
		}
		if data.Progress != 100 || data.Charge != data.MaxCharge {
			t.Fatalf("recharge progress/charge = (%d, %d), want (100, %d)", data.Progress, data.Charge, data.MaxCharge)
		}
		if rec.Pending != 0 || rec.Object != obj {
			t.Fatalf("record changed = pending:%d object:%p, want 0/%p", rec.Pending, rec.Object, obj)
		}
	})
}
