package server

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
)

func TestProjectileSparkCollide4E8880NativeMapFieldsAndLayout(t *testing.T) {
	data := &ProjectileCollideData{Damage: -31, Field4: 0x12345678}
	projectile := &Object{NewPos: types.Ptf(69, -46), CollideData: unsafe.Pointer(data)}
	collision := uint32(0x7fa12345)
	var gotMap [4]int32
	var gotSource *Object
	var deleted []*Object

	projectileSparkCollideNative4E8880(projectile, nil, unsafe.Pointer(&collision), projectileSparkCollideNativeDeps4E8880{
		floatToInt: playerCollideRound4E8460,
		damageMap: func(x, y, damage int32, damageType object.DamageType, source *Object) {
			gotMap = [4]int32{x, y, damage, int32(damageType)}
			gotSource = source
		},
		delayedDelete: func(obj *Object) { deleted = append(deleted, obj) },
	})

	if gotMap != [4]int32{3, -2, -31, int32(object.DamageImpact)} {
		t.Fatalf("map args = %v", gotMap)
	}
	if gotSource != projectile {
		t.Fatalf("map source = %p, want %p", gotSource, projectile)
	}
	if len(deleted) != 1 || deleted[0] != projectile {
		t.Fatalf("deleted = %v, want projectile once", deleted)
	}
	if collision != 0x7fa12345 || data.Field4 != 0x12345678 || projectile.NewPos != types.Ptf(69, -46) {
		t.Fatal("ignored collision, collide-data field, or position changed")
	}

	ptrSize := unsafe.Sizeof(uintptr(0))
	wantNewPos, wantCollideData, wantDamage := uintptr(68), uintptr(776), uintptr(808)
	if ptrSize == 4 {
		wantNewPos, wantCollideData, wantDamage = 64, 700, 716
	}
	if got := unsafe.Offsetof(Object{}.NewPos); got != wantNewPos {
		t.Fatalf("Object.NewPos offset = %d, want %d", got, wantNewPos)
	}
	if got := unsafe.Offsetof(Object{}.CollideData); got != wantCollideData {
		t.Fatalf("Object.CollideData offset = %d, want %d", got, wantCollideData)
	}
	if got := unsafe.Offsetof(Object{}.Damage); got != wantDamage {
		t.Fatalf("Object.Damage offset = %d, want %d", got, wantDamage)
	}
	if unsafe.Sizeof(*data) != 8 || unsafe.Offsetof(data.Damage) != 0 || unsafe.Offsetof(data.Field4) != 4 {
		t.Fatalf("collide layout = size %d offsets %d/%d", unsafe.Sizeof(*data), unsafe.Offsetof(data.Damage), unsafe.Offsetof(data.Field4))
	}
}
