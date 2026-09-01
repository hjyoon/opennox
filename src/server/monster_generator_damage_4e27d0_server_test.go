package server

import (
	"math"
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func TestMonsterGeneratorDamagePointFXPacket523150(t *testing.T) {
	packet := monsterGeneratorDamagePointFXPacket523150(0xf0, 26, types.Ptf(2.5, -3.5))
	want := [6]byte{0xf0, 26, 2, 0, 0xfc, 0xff}
	if packet != want {
		t.Fatalf("packet = %v, want %v", packet, want)
	}

	packet = monsterGeneratorDamagePointFXPacket523150(0x12, 0x34, types.Ptf(3.5, -2.5))
	want = [6]byte{0x12, 0x34, 4, 0, 0xfe, 0xff}
	if packet != want {
		t.Fatalf("tie-to-even packet = %v, want %v", packet, want)
	}
}

func TestMonsterGeneratorDamageNativeKeepsHighPointersAndScriptSlot4E27D0(t *testing.T) {
	update := &MonsterGenUpdateData{Field48: 0x01020304, FuncInd52: 0xf1020304}
	health := &HealthData{Cur: 250, Max: 300}
	target := &Object{UpdateData: unsafe.Pointer(update), Frame134: 10, HealthData: health}
	source := &Object{}
	weapon := &Object{}
	var pin runtime.Pinner
	pin.Pin(target)
	pin.Pin(source)
	pin.Pin(weapon)
	pin.Pin(update)
	defer pin.Unpin()

	if unsafe.Sizeof(uintptr(0)) == 8 {
		for name, pointer := range map[string]unsafe.Pointer{
			"target": unsafe.Pointer(target), "source": unsafe.Pointer(source),
			"weapon": unsafe.Pointer(weapon), "update": unsafe.Pointer(update),
		} {
			if uintptr(pointer) <= math.MaxUint32 {
				t.Fatalf("%s pointer = %p, want native high address", name, pointer)
			}
		}
	}

	defaultCalls := 0
	scriptCalls := 0
	got := MonsterGeneratorDamage4E27D0(target, source, weapon, -31, object.DamageElectric, MonsterGeneratorDamageRuntime4E27D0{
		Frame: func() uint32 { return 11 },
		Default: func(gotTarget, gotSource, gotWeapon *Object, damage int32, typ object.DamageType) bool {
			defaultCalls++
			if gotTarget != target || gotSource != source || gotWeapon != weapon || damage != -31 || typ != object.DamageElectric {
				t.Fatalf("default args = %p/%p/%p/%d/%d", gotTarget, gotSource, gotWeapon, damage, typ)
			}
			health.Cur--
			return true
		},
		Script: func(block *ScriptCallback, caller, trigger *Object, event ScriptEventType) {
			scriptCalls++
			wantBlock := (*ScriptCallback)(unsafe.Pointer(&update.Field48))
			if block != wantBlock || caller != source || trigger != target || event != NoxEventGeneratorDamage {
				t.Fatalf("script args = %p/%p/%p/%d, want %p/%p/%p/%d",
					block, caller, trigger, event, wantBlock, source, target, NoxEventGeneratorDamage)
			}
			if block.Flags != 0x01020304 || uint32(block.Func) != 0xf1020304 {
				t.Fatalf("script block = %#x/%#x", block.Flags, uint32(block.Func))
			}
		},
	})
	if !got || defaultCalls != 1 || scriptCalls != 1 {
		t.Fatalf("result/default/script = %t/%d/%d, want true/1/1", got, defaultCalls, scriptCalls)
	}
	runtime.KeepAlive(target)
	runtime.KeepAlive(source)
	runtime.KeepAlive(weapon)
	runtime.KeepAlive(update)
}

func TestMonsterGeneratorDamageNativeBindsStatusMethods4E27D0(t *testing.T) {
	update := new(MonsterGenUpdateData)
	health := &HealthData{Cur: 100, Max: 300}
	target := &Object{
		ObjClass: object.ClassImmobile, Field5: 0x100, Frame134: 11,
		UpdateData: unsafe.Pointer(update), HealthData: health,
	}
	got := MonsterGeneratorDamage4E27D0(target, nil, nil, 1, object.DamageBlade, MonsterGeneratorDamageRuntime4E27D0{
		Frame:   func() uint32 { return 11 },
		Default: func(*Object, *Object, *Object, int32, object.DamageType) bool { return false },
	})
	if got || target.Field5 != 0x200 || target.Field38 != math.MaxUint32 {
		t.Fatalf("result/status/sync = %t/%#x/%#x, want false/0x200/max", got, target.Field5, target.Field38)
	}
	if !reflect.DeepEqual(target.Field140, [32]uint32{
		0x80000, 0x80000, 0x80000, 0x80000, 0x80000, 0x80000, 0x80000, 0x80000,
		0x80000, 0x80000, 0x80000, 0x80000, 0x80000, 0x80000, 0x80000, 0x80000,
		0x80000, 0x80000, 0x80000, 0x80000, 0x80000, 0x80000, 0x80000, 0x80000,
		0x80000, 0x80000, 0x80000, 0x80000, 0x80000, 0x80000, 0x80000, 0x80000,
	}) {
		t.Fatalf("sync fields = %v", target.Field140)
	}
}

func TestMonsterGeneratorDamageNativeNilTargetPreservesInitialFieldFault4E27D0(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("nil target did not preserve the original flags-load fault")
		}
	}()
	MonsterGeneratorDamage4E27D0(nil, nil, nil, 1, object.DamageBlade, MonsterGeneratorDamageRuntime4E27D0{})
}
