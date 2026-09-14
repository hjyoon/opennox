package server

import (
	"math"

	"github.com/opennox/libs/types"
)

// SparkCreateRuntime54FD80 supplies object creation without passing a native
// object pointer through GAME.EXE's 32-bit integer arguments.
type SparkCreateRuntime54FD80 struct {
	NewObject func(string) *Object
	CreateAt  func(*Object, *Object, types.Pointf)
	Frame     func() uint32
	Raise     func(*Object, float32)
}

// CreateSpark54FD80 restores the Spark factory's object and update-data writes
// using native-width pointers. Its caller owns the map insertion operation.
func CreateSpark54FD80(
	runtime SparkCreateRuntime54FD80, pos types.Pointf, kind, lifetime int,
	velocity types.Pointf, z float32, owner *Object,
) *Object {
	spark := runtime.NewObject("Spark")
	if spark == nil || spark.CollideData == nil || spark.UpdateData == nil {
		return nil
	}
	runtime.CreateAt(spark, owner, pos)
	spark.Field34 = runtime.Frame()
	update := (*SparkUpdateData)(spark.UpdateData)
	update.LifetimeInitial = uint32(lifetime)
	update.LifetimeRemaining = uint32(lifetime)
	update.Kind = uint32(kind)
	spark.ObjClass = spark.ObjClass&^0x2000 | 0x80000
	spark.ObjFlags &^= 0x800040
	mode := uint32(0)
	float29 := float32(0)
	switch kind {
	case 0:
		spark.ObjFlags |= 0x40
		spark.ObjClass &^= 0x80000
	case 1:
		spark.ObjClass |= 0x2000
		spark.ObjFlags |= 0x800000
		mode = 3
		float29 = 7
	case 2:
		spark.ObjFlags |= 0x800040
		float29 = 7
	case 4:
		spark.ObjClass &^= 0x80000
	}
	*(*uint32)(spark.CollideData) = mode
	runtime.Raise(spark, 28)
	spark.Field27 = z
	spark.Field29 = math.Float32bits(float29)
	spark.VelVec = velocity
	return spark
}
