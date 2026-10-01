package server

import (
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/balance"
	"github.com/opennox/libs/types"
)

func TestQuestHecubahSpawn51A5A0NativePointersBalanceAndFourRewards(t *testing.T) {
	srv := new(Server)
	srv.Balance.file = &balance.File{Global: balance.Config{"hecubahquestskill": balance.Float(123)}}
	definition := &MonsterDef{HealthQuest72: 25}
	update := &MonsterUpdateData{MonsterDef: definition, Field330: 0.75, Field389: 0x11223344}
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
	stage := uint32(5)
	h := srv.questHecubahSpawnNativeHooks51A5A0(position, QuestHecubahSpawnRuntime51A5A0{
		HealthScale: func() float64 { return 1.5 },
		SetHP: func(got *Object, hp uint16) {
			if got != unit || hp != 37 {
				t.Fatalf("set identity/hp=%p/%d", got, hp)
			}
			health.Cur = hp
			unit.UpdateData = unsafe.Pointer(new(MonsterUpdateData))
		},
		CreateAt: func(got *Object, p types.Pointf) {
			if got != unit || p != (types.Pointf{X: 47, Y: 43}) {
				t.Fatalf("create identity/position=%p/%v", got, p)
			}
			trace = append(trace, "create")
		},
		Stage: func() uint32 { v := stage; stage++; trace = append(trace, "stage"); return v },
		InventoryPut: func(owner, item *Object, report int32) {
			if owner != unit || item != reward || report != 0 {
				t.Fatal("inventory identity/report")
			}
			trace = append(trace, "inventory")
		},
	})
	lookup := h.balanceFloat
	h.balanceFloat = func(key string) float32 {
		if key != "HecubahQuestSkill" || update.Field330 != 0.75 || update.Field388 != 0x40000000 || lookup(key) != 123 {
			t.Fatal("lost ordered native balance lookup")
		}
		position.X = 47
		trace = append(trace, "balance")
		return 123
	}
	h.newObject = func(name string) *Object {
		if name == "Hecubah" {
			return unit
		}
		if name == "RewardMarker" {
			return marker
		}
		panic(name)
	}
	h.activateReward = func(got *Object, rewardStage uint32) *Object {
		if got != marker || rewardStage != stage+1 {
			t.Fatal("reward identity/live stage")
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
	questHecubahSpawn51A5A0(h)
	want := []string{"balance", "create", "stage", "reward", "inventory", "stage", "reward", "inventory", "stage", "reward", "inventory", "stage", "reward", "inventory", "free"}
	if health.Cur != 37 || health.Max != 38 || health.Field2 != 0x1234 || update.MonsterDef != definition || update.Field389 != 0x11223344 ||
		update.AIAction340 != 4 || update.Field411 != 0x10000000 || update.Field423 != 0x10000000 ||
		math.Float32bits(update.Aggression) != 0x3f547ae1 || update.Field510 != 3 || update.Field410 != 0x08000000 ||
		update.Field444 != 0x20000000 || update.Field415 != 0x40000000 || update.Field388 != 0x40000000 || math.Float32bits(update.Field330) != 0x3f59999a ||
		(*MonsterUpdateData)(unit.UpdateData).AIAction340 != 0 || stage != 9 || !reflect.DeepEqual(trace, want) {
		t.Fatalf("native fields health=%+v update=%+v trace=%v", health, update, trace)
	}
}

func TestQuestHecubahSpawn51A5A0NativeLateNullPosition(t *testing.T) {
	srv := new(Server)
	srv.Balance.file = &balance.File{Global: balance.Config{"hecubahquestskill": balance.Float(123)}}
	update := &MonsterUpdateData{MonsterDef: &MonsterDef{HealthQuest72: 25}}
	unit := &Object{HealthData: new(HealthData), UpdateData: unsafe.Pointer(update)}
	h := srv.questHecubahSpawnNativeHooks51A5A0(nil, QuestHecubahSpawnRuntime51A5A0{
		HealthScale: func() float64 { return 1 },
		SetHP:       func(o *Object, hp uint16) { o.HealthData.Cur = hp },
	})
	h.newObject = func(string) *Object { return nil }
	questHecubahSpawn51A5A0(h)
	h.newObject = func(string) *Object { return unit }
	defer func() {
		if recover() == nil || unit.HealthData.Cur != 25 || unit.HealthData.Max != 25 || math.Float32bits(update.Field330) != 0x3f59999a {
			t.Fatal("missing late position fault or incomplete preceding stores")
		}
	}()
	questHecubahSpawn51A5A0(h)
}
