package server

import (
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"
)

func TestQuestNecroSpawn51A7A0NativePointersAndExactAIFields(t *testing.T) {
	srv := new(Server)
	definition := &MonsterDef{HealthQuest72: 25}
	update := &MonsterUpdateData{MonsterDef: definition, Field388: 1234, Field330: 0.75}
	health := &HealthData{Cur: 7, Max: 9, Field2: 0x1234}
	unit := &Object{UpdateData: unsafe.Pointer(update), HealthData: health}
	marker, reward := new(Object), new(Object)
	position := &types.Pointf{X: 41, Y: 43}
	if unsafe.Sizeof(uintptr(0)) > 4 {
		for _, p := range []unsafe.Pointer{unsafe.Pointer(unit), unsafe.Pointer(update), unsafe.Pointer(definition), unsafe.Pointer(health), unsafe.Pointer(position), unsafe.Pointer(marker), unsafe.Pointer(reward)} {
			if uintptr(p) <= math.MaxUint32 {
				t.Fatalf("native pointer %p is not above 4 GiB", p)
			}
		}
	}
	var trace []string
	h := srv.questNecroSpawnNativeHooks51A7A0(position, QuestNecroSpawnRuntime51A7A0{
		HealthScale: func() float64 { return 1.5 },
		SetHP: func(got *Object, hp uint16) {
			if got != unit || hp != 37 {
				t.Fatalf("set identity/hp=%p/%d", got, hp)
			}
			health.Cur = hp
			unit.UpdateData = unsafe.Pointer(new(MonsterUpdateData))
			position.X = 47
		},
		CreateAt: func(got *Object, p types.Pointf) {
			if got != unit || p != (types.Pointf{X: 47, Y: 43}) {
				t.Fatalf("create identity/position=%p/%v", got, p)
			}
			trace = append(trace, "create")
		},
		Stage: func() uint32 { trace = append(trace, "stage"); return 5 },
		InventoryPut: func(owner, item *Object, report int32) {
			if owner != unit || item != reward || report != 0 {
				t.Fatal("inventory identity/report")
			}
			trace = append(trace, "inventory")
		},
	})
	h.newObject = func(name string) *Object {
		if name == "Necromancer" {
			return unit
		}
		if name == "RewardMarker" {
			return marker
		}
		panic(name)
	}
	h.activateReward = func(got *Object, stage uint32) *Object {
		if got != marker || stage != 7 {
			t.Fatal("reward identity/stage")
		}
		trace = append(trace, "reward")
		return reward
	}
	h.freeObject = func(got *Object) {
		if got != marker {
			t.Fatal("free identity")
		}
		trace = append(trace, "free")
	}
	questNecroSpawn51A7A0(h)
	if health.Cur != 37 || health.Max != 38 || health.Field2 != 0x1234 || update.MonsterDef != definition ||
		update.AIAction340 != 4 || update.Field411 != 0x10000000 || update.Field423 != 0x10000000 ||
		math.Float32bits(update.Aggression) != 0x3f547ae1 || update.Field510 != 1 || update.Field410 != 0x08000000 ||
		update.Field444 != 0x20000000 || update.Field415 != 0x40000000 || update.Field388 != 1234 || update.Field330 != 0.75 ||
		(*MonsterUpdateData)(unit.UpdateData).AIAction340 != 0 || !reflect.DeepEqual(trace, []string{"create", "stage", "reward", "inventory", "free"}) {
		t.Fatalf("native fields health=%+v update=%+v trace=%v", health, update, trace)
	}
}

func TestQuestNecroSpawn51A7A0NativeLateNullPosition(t *testing.T) {
	srv := new(Server)
	unit := &Object{HealthData: new(HealthData), UpdateData: unsafe.Pointer(&MonsterUpdateData{MonsterDef: &MonsterDef{HealthQuest72: 25}})}
	h := srv.questNecroSpawnNativeHooks51A7A0(nil, QuestNecroSpawnRuntime51A7A0{
		HealthScale: func() float64 { return 1 },
		SetHP:       func(o *Object, hp uint16) { o.HealthData.Cur = hp },
	})
	h.newObject = func(string) *Object { return nil }
	questNecroSpawn51A7A0(h) // Failed allocation must not dereference nil position.
	h.newObject = func(string) *Object { return unit }
	defer func() {
		if recover() == nil || unit.HealthData.Cur != 25 || unit.HealthData.Max != 25 || (*MonsterUpdateData)(unit.UpdateData).AIAction340 != 4 {
			t.Fatal("missing late position fault or incomplete preceding stores")
		}
	}()
	questNecroSpawn51A7A0(h)
}
