package legacy

import (
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

type abilityRewardLegacyServer4FB9C0 struct {
	Server
	rewardUnit   *server.Object
	rewardArgs   [2]int32
	rewardResult int32
	useOwner     *server.Object
	useItem      *server.Object
	useResult    int32
}

func (s *abilityRewardLegacyServer4FB9C0) AbilityRewardServ4FB9C0(
	unit *server.Object,
	ability, rewardArg int32,
) int32 {
	s.rewardUnit = unit
	s.rewardArgs = [2]int32{ability, rewardArg}
	return s.rewardResult
}

func (s *abilityRewardLegacyServer4FB9C0) UseAbilityReward53FAE0(
	owner, item *server.Object,
) int32 {
	s.useOwner = owner
	s.useItem = item
	return s.useResult
}

func TestAbilityRewardExportsPreserveNativePointersAndScalars(t *testing.T) {
	fake := &abilityRewardLegacyServer4FB9C0{
		rewardResult: -0x2345678,
		useResult:    math.MinInt32 + 0x4321,
	}
	oldGetServer := GetServer
	GetServer = func() Server { return fake }
	t.Cleanup(func() { GetServer = oldGetServer })

	unit := new(server.Object)
	item := new(server.Object)
	var pin runtime.Pinner
	pin.Pin(unit)
	pin.Pin(item)
	defer pin.Unpin()
	if unsafe.Sizeof(uintptr(0)) == 8 {
		if uintptr(unsafe.Pointer(unit)) <= math.MaxUint32 || uintptr(unsafe.Pointer(item)) <= math.MaxUint32 {
			t.Fatalf("test pointers do not exercise native high addresses: unit=%p item=%p", unit, item)
		}
	}

	wantRewardArgs := [2]int32{math.MaxInt32, math.MinInt32 + 0x1234}
	if got := abilityRewardExportCall4FB9C0(unit, wantRewardArgs[0], wantRewardArgs[1]); got != fake.rewardResult {
		t.Fatalf("reward export result = %d, want %d", got, fake.rewardResult)
	}
	if fake.rewardUnit != unit || fake.rewardArgs != wantRewardArgs {
		t.Fatalf("reward export call = %p/%v, want %p/%v", fake.rewardUnit, fake.rewardArgs, unit, wantRewardArgs)
	}

	fake.rewardResult = 0x1234567
	wantRewardArgs = [2]int32{5, -7}
	if got := Nox_xxx_abilityRewardServ_4FB9C0_ability(unit, wantRewardArgs[0], wantRewardArgs[1]); got != fake.rewardResult {
		t.Fatalf("reward Go wrapper result = %d, want %d", got, fake.rewardResult)
	}
	if fake.rewardUnit != unit || fake.rewardArgs != wantRewardArgs {
		t.Fatalf("reward Go wrapper call = %p/%v, want %p/%v", fake.rewardUnit, fake.rewardArgs, unit, wantRewardArgs)
	}

	if got := useAbilityRewardExportCall53FAE0(unit, item); got != fake.useResult {
		t.Fatalf("use export result = %d, want %d", got, fake.useResult)
	}
	if fake.useOwner != unit || fake.useItem != item {
		t.Fatalf("use export call = %p/%p, want %p/%p", fake.useOwner, fake.useItem, unit, item)
	}
	runtime.KeepAlive(unit)
	runtime.KeepAlive(item)
}
