package server

import (
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func defaultPixieCollideNativeDeps4EA080() pixieCollideNativeDeps4EA080 {
	return pixieCollideNativeDeps4EA080{
		isEnemy:        func(*Object, *Object) int32 { return 1 },
		checkInversion: func(*Object, *Object) int32 { return 0 },
		changeOwner:    func(*Object, *Object) {},
		checkDirection: func(types.Pointf, int16, types.Pointf) int32 { return 0 },
		findParent:     func(obj *Object) *Object { return obj.ObjOwner },
		targetDamage:   func(*Object, *Object, *Object, int32, object.DamageType) int32 { return 0 },
		audio:          func(uint32, *Object) {},
		delayedDelete:  func(*Object) {},
		wallReflect:    spellProjectileWallReflect57B810,
		floatToInt:     playerCollideRound4E8460,
		damageMap:      func(int32, int32, int32, object.DamageType, *Object) {},
	}
}

func TestPixieCollide4EA080NativeLayout(t *testing.T) {
	wantClass := uintptr(8)
	wantFlags := uintptr(16)
	wantPos := uintptr(56)
	wantNewPos := uintptr(64)
	wantVelocity := uintptr(80)
	wantDirection1 := uintptr(124)
	wantDirection2 := uintptr(126)
	wantOwner := uintptr(508)
	wantCollideData := uintptr(700)
	wantDamage := uintptr(716)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantClass = 12
		wantFlags = 20
		wantPos = 60
		wantNewPos = 68
		wantVelocity = 84
		wantDirection1 = 128
		wantDirection2 = 130
		wantOwner = 552
		wantCollideData = 776
		wantDamage = 808
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"ProjectileCollideData size", unsafe.Sizeof(ProjectileCollideData{}), 8},
		{"ProjectileCollideData.Damage", unsafe.Offsetof(ProjectileCollideData{}.Damage), 0},
		{"ProjectileCollideData.Field4", unsafe.Offsetof(ProjectileCollideData{}.Field4), 4},
		{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), wantClass},
		{"Object.ObjFlags", unsafe.Offsetof(Object{}.ObjFlags), wantFlags},
		{"Object.PosVec", unsafe.Offsetof(Object{}.PosVec), wantPos},
		{"Object.NewPos", unsafe.Offsetof(Object{}.NewPos), wantNewPos},
		{"Object.VelVec", unsafe.Offsetof(Object{}.VelVec), wantVelocity},
		{"Object.Direction1", unsafe.Offsetof(Object{}.Direction1), wantDirection1},
		{"Object.Direction2", unsafe.Offsetof(Object{}.Direction2), wantDirection2},
		{"Object.ObjOwner", unsafe.Offsetof(Object{}.ObjOwner), wantOwner},
		{"Object.CollideData", unsafe.Offsetof(Object{}.CollideData), wantCollideData},
		{"Object.Damage", unsafe.Offsetof(Object{}.Damage), wantDamage},
		{"Pointf size", unsafe.Sizeof(types.Pointf{}), 8},
		{"Pointf.X", unsafe.Offsetof(types.Pointf{}.X), 0},
		{"Pointf.Y", unsafe.Offsetof(types.Pointf{}.Y), 4},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestPixieCollideNative4EA080TargetUsesCachedDataAndLiveDamage(t *testing.T) {
	oldData := &ProjectileCollideData{Damage: -31, Field4: 0x12345678}
	newData := &ProjectileCollideData{Damage: 99}
	parent := &Object{}
	source := &Object{CollideData: unsafe.Pointer(oldData), ObjOwner: parent}
	target := &Object{ObjClass: 2}
	events := make([]string, 0, 6)
	deps := defaultPixieCollideNativeDeps4EA080()
	deps.isEnemy = func(first, second *Object) int32 {
		events = append(events, "enemy")
		if first != source || second != target {
			t.Fatalf("enemy = %p/%p", first, second)
		}
		source.CollideData = unsafe.Pointer(newData)
		return 1
	}
	deps.findParent = func(got *Object) *Object {
		events = append(events, "parent")
		if got != source {
			t.Fatalf("parent source = %p", got)
		}
		target.ObjFlags = 7
		return parent
	}
	deps.targetDamage = func(gotTarget, gotParent, gotSource *Object, damage int32, damageType object.DamageType) int32 {
		events = append(events, "damage")
		if gotTarget != target || gotParent != parent || gotSource != source || damage != -31 || damageType != object.DamageType(11) {
			t.Fatalf("damage = %p/%p/%p/%d/%d", gotTarget, gotParent, gotSource, damage, damageType)
		}
		if target.ObjFlags != 7 {
			t.Fatal("Damage callback was not observed after parent lookup")
		}
		return math.MinInt32
	}
	deps.audio = func(id uint32, got *Object) {
		events = append(events, "audio")
		if id != 96 || got != source {
			t.Fatalf("audio = %d/%p", id, got)
		}
	}
	deps.delayedDelete = func(got *Object) {
		events = append(events, "delete")
		if got != source {
			t.Fatalf("delete = %p", got)
		}
	}
	pixieCollideNative4EA080(source, target, nil, deps)
	if want := []string{"enemy", "parent", "damage", "audio", "delete"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
	if oldData.Field4 != 0x12345678 {
		t.Fatalf("unused collide-data word = %#x", oldData.Field4)
	}
}

func TestPixieCollideNative4EA080ReflectiveShieldUsesNativeFields(t *testing.T) {
	source := &Object{CollideData: unsafe.Pointer(&ProjectileCollideData{}), PosVec: types.Ptf(-4, 9)}
	target := &Object{
		ObjClass:   4,
		Buffs:      1 << 27,
		PosVec:     types.Ptf(7, -2),
		Direction1: Dir16(0x8001),
	}
	events := make([]string, 0, 4)
	deps := defaultPixieCollideNativeDeps4EA080()
	deps.checkInversion = func(gotTarget, gotSource *Object) int32 {
		events = append(events, "inversion")
		if gotTarget != target || gotSource != source {
			t.Fatalf("inversion = %p/%p", gotTarget, gotSource)
		}
		return 0
	}
	deps.checkDirection = func(first types.Pointf, direction int16, second types.Pointf) int32 {
		events = append(events, "direction")
		if first != target.PosVec || direction != -32767 || second != source.PosVec {
			t.Fatalf("direction = %v/%d/%v", first, direction, second)
		}
		return 1
	}
	deps.changeOwner = func(gotSource, gotTarget *Object) {
		events = append(events, "owner")
		if gotSource != source || gotTarget != target {
			t.Fatalf("owner = %p/%p", gotSource, gotTarget)
		}
	}
	deps.audio = func(id uint32, got *Object) {
		events = append(events, "audio")
		if id != 122 || got != target {
			t.Fatalf("audio = %d/%p", id, got)
		}
	}
	deps.targetDamage = func(*Object, *Object, *Object, int32, object.DamageType) int32 {
		t.Fatal("reflected collision damaged target")
		return 0
	}
	deps.delayedDelete = func(*Object) { t.Fatal("reflected collision deleted source") }
	pixieCollideNative4EA080(source, target, nil, deps)
	if want := []string{"inversion", "direction", "owner", "audio"}; !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %v, want %v", events, want)
	}
}

func TestPixieCollideNative4EA080WallReflectsAdvancesAndDamagesMap(t *testing.T) {
	data := &ProjectileCollideData{Damage: 17, Field4: -1}
	source := &Object{
		CollideData: unsafe.Pointer(data),
		NewPos:      types.Ptf(10, 20),
		VelVec:      types.Ptf(2, 3),
	}
	collision := &types.Pointf{X: 1, Y: -1}
	deps := defaultPixieCollideNativeDeps4EA080()
	inputs := make([]uint32, 0, 2)
	results := []int32{-5, 7}
	deps.floatToInt = func(value float32) int32 {
		inputs = append(inputs, math.Float32bits(value))
		result := results[0]
		results = results[1:]
		return result
	}
	var got [4]int32
	var gotSource *Object
	deps.damageMap = func(x, y, damage int32, damageType object.DamageType, source *Object) {
		got = [4]int32{x, y, damage, int32(damageType)}
		gotSource = source
	}
	deps.delayedDelete = func(*Object) { t.Fatal("wall path deleted source") }
	pixieCollideNative4EA080(source, nil, collision, deps)
	if source.VelVec != (types.Ptf(3, 2)) || source.NewPos != (types.Ptf(13, 22)) {
		t.Fatalf("velocity/new position = %v/%v", source.VelVec, source.NewPos)
	}
	if want := Dir16(directionFromVector509ED0(3, 2)); source.Direction2 != want {
		t.Fatalf("Direction2 = %d, want %d", source.Direction2, want)
	}
	gridInverse := float64(math.Float32frombits(pixieGridInverseBits4EA080))
	wantInputs := []uint32{
		math.Float32bits(float32(float64(22) * gridInverse)),
		math.Float32bits(float32(float64(13) * gridInverse)),
	}
	if !reflect.DeepEqual(inputs, wantInputs) {
		t.Fatalf("conversion inputs = %#v, want %#v", inputs, wantInputs)
	}
	if got != [4]int32{7, -5, 17, 11} || gotSource != source {
		t.Fatalf("map = %#v/%p", got, gotSource)
	}
}
