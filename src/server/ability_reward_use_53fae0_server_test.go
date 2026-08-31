package server

import (
	"math"
	"runtime"
	"testing"
	"unsafe"
)

func defaultAbilityRewardUseNativeDeps53FAE0() abilityRewardUseNativeDeps53FAE0 {
	return abilityRewardUseNativeDeps53FAE0{
		primaryMessage: func(*Object, string, uint8) {},
		audit:          func(int32, *Object, int32, uint32) {},
		gameFlagsCheck: func(uint32) int32 { return 0 },
		rewardAbility:  func(*Object, int32, int32) int32 { return 0 },
		delayedDelete:  func(*Object) {},
	}
}

func TestUseAbilityRewardNative53FAE0PreservesPointersAndScalars(t *testing.T) {
	player := &Player{}
	update := &PlayerUpdateData{Player: player}
	owner := &Object{ObjClass: 0xf4, NetCode: 0x89abcdef, UpdateData: unsafe.Pointer(update)}
	data := &AbilityRewardUseData{Ability: 5}
	item := &Object{}
	item.UseData.SetPtr(unsafe.Pointer(data))

	var (
		rewardOwner *Object
		rewardArgs  [2]int32
		deletedItem *Object
	)
	deps := defaultAbilityRewardUseNativeDeps53FAE0()
	deps.gameFlagsCheck = func(mask uint32) int32 {
		if mask != abilityRewardUseQuestMask53FAE0 {
			t.Fatalf("flag mask = %#x, want %#x", mask, abilityRewardUseQuestMask53FAE0)
		}
		return math.MinInt32
	}
	deps.rewardAbility = func(gotOwner *Object, ability, rewardArg int32) int32 {
		rewardOwner = gotOwner
		rewardArgs = [2]int32{ability, rewardArg}
		return math.MinInt32
	}
	deps.delayedDelete = func(gotItem *Object) {
		deletedItem = gotItem
	}
	deps.audit = func(int32, *Object, int32, uint32) {
		t.Fatal("successful reward emitted audit")
	}

	if got := useAbilityRewardNative53FAE0(owner, item, deps); got != 1 {
		t.Fatalf("result = %d, want 1", got)
	}
	if rewardOwner != owner || rewardArgs != [2]int32{5, 1} {
		t.Fatalf("reward = %p/%v, want %p/[5 1]", rewardOwner, rewardArgs, owner)
	}
	if deletedItem != item {
		t.Fatalf("deleted item = %p, want %p", deletedItem, item)
	}

	if unsafe.Sizeof(uintptr(0)) == 8 {
		for name, ptr := range map[string]uintptr{
			"player": uintptr(unsafe.Pointer(player)),
			"update": uintptr(unsafe.Pointer(update)),
			"owner":  uintptr(unsafe.Pointer(owner)),
			"data":   uintptr(unsafe.Pointer(data)),
			"item":   uintptr(unsafe.Pointer(item)),
		} {
			if ptr <= math.MaxUint32 {
				t.Fatalf("%s pointer = %#x, want native address above 4 GiB", name, ptr)
			}
		}
	}
	runtime.KeepAlive(player)
	runtime.KeepAlive(update)
	runtime.KeepAlive(owner)
	runtime.KeepAlive(data)
	runtime.KeepAlive(item)
}

func TestUseAbilityRewardNative53FAE0ClassFailureUsesLiveOwner(t *testing.T) {
	player := &Player{}
	player.info[66] = 0xff
	update := &PlayerUpdateData{Player: player}
	owner := &Object{ObjClass: 4, NetCode: 0x10203040, UpdateData: unsafe.Pointer(update)}
	data := &AbilityRewardUseData{Ability: 2}
	item := &Object{}
	item.UseData.SetPtr(unsafe.Pointer(data))

	deps := defaultAbilityRewardUseNativeDeps53FAE0()
	deps.primaryMessage = func(gotOwner *Object, message string, value uint8) {
		if gotOwner != owner || message != abilityRewardUseClassMessage53FAE0 || value != 0 {
			t.Fatalf("message = %p/%q/%d", gotOwner, message, value)
		}
		owner.NetCode = math.MaxUint32
	}
	audited := false
	deps.audit = func(id int32, gotOwner *Object, kind int32, code uint32) {
		audited = true
		if id != abilityRewardUseAuditSound53FAE0 || gotOwner != owner || kind != abilityRewardUseAuditKind53FAE0 || code != math.MaxUint32 {
			t.Fatalf("audit = %d/%p/%d/%#x", id, gotOwner, kind, code)
		}
	}
	deps.gameFlagsCheck = func(uint32) int32 {
		t.Fatal("class failure checked game flags")
		return 0
	}

	if got := useAbilityRewardNative53FAE0(owner, item, deps); got != 0 || !audited {
		t.Fatalf("result/audited = %d/%t, want 0/true", got, audited)
	}
	runtime.KeepAlive(player)
	runtime.KeepAlive(update)
	runtime.KeepAlive(owner)
	runtime.KeepAlive(data)
	runtime.KeepAlive(item)
}
