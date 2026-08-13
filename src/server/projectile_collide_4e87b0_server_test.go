package server

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/common/ntype"
)

func TestProjectileCollide4E87B0NativeWallFieldsAndLayout(t *testing.T) {
	data := &ProjectileCollideData{Damage: -31, Field4: 0x12345678}
	projectile := &Object{TypeInd: 7, CollideData: unsafe.Pointer(data)}
	point := &ntype.Point32{X: -101, Y: 202}
	var gotMap [4]int32
	var gotSource *Object
	var deleted []*Object

	projectileCollideNative4E87B0(projectile, nil, unsafe.Pointer(uintptr(0x1234)), projectileCollideNativeDeps4E87B0{
		loadThrowingStoneType: func() uint32 { return 5 },
		lookupType: func(string) uint32 {
			t.Fatal("lookup called with populated cache")
			return 0
		},
		storeThrowingStone: func(uint32) { t.Fatal("ThrowingStone cache stored") },
		storeImpShot:       func(uint32) { t.Fatal("ImpShot cache stored") },
		loadImpShotType:    func() uint32 { return 6 },
		gameDataFloat: func(string) float64 {
			t.Fatal("balance override called for ordinary projectile")
			return 0
		},
		floatToInt: func(float32) int32 {
			t.Fatal("float conversion called for ordinary projectile")
			return 0
		},
		traceHitPoint: func() *ntype.Point32 { return point },
		damageMap: func(x, y, damage int32, damageType object.DamageType, source *Object) {
			gotMap = [4]int32{x, y, damage, int32(damageType)}
			gotSource = source
		},
		delayedDelete: func(obj *Object) { deleted = append(deleted, obj) },
	})

	if gotMap != [4]int32{-101, 202, -31, int32(object.DamageImpact)} {
		t.Fatalf("map args = %v", gotMap)
	}
	if gotSource != projectile {
		t.Fatalf("map source = %p, want %p", gotSource, projectile)
	}
	if len(deleted) != 1 || deleted[0] != projectile {
		t.Fatalf("deleted = %v, want projectile once", deleted)
	}
	if data.Field4 != 0x12345678 {
		t.Fatalf("unused collide field changed to %#x", data.Field4)
	}
	if unsafe.Sizeof(*data) != 8 || unsafe.Offsetof(data.Damage) != 0 || unsafe.Offsetof(data.Field4) != 4 {
		t.Fatalf("collide layout = size %d offsets %d/%d", unsafe.Sizeof(*data), unsafe.Offsetof(data.Damage), unsafe.Offsetof(data.Field4))
	}
	wantCollideData := uintptr(700)
	wantDamage := uintptr(716)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantCollideData = 776
		wantDamage = 808
	}
	if got := unsafe.Offsetof(Object{}.CollideData); got != wantCollideData {
		t.Fatalf("Object.CollideData offset = %d, want %d", got, wantCollideData)
	}
	if got := unsafe.Offsetof(Object{}.Damage); got != wantDamage {
		t.Fatalf("Object.Damage offset = %d, want %d", got, wantDamage)
	}
}
