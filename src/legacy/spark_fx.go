package legacy

import (
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/server"
)

// Nox_xxx_createSpark_54FD80 keeps Spark creation on the native object path.
// The old C factory interpreted object pointers as 32-bit integers.
func Nox_xxx_createSpark_54FD80(
	x, y float32,
	kind, lifetime int,
	velocityX, velocityY, z float32,
	owner *server.Object,
) *server.Object {
	outer := GetServer()
	world := outer.S()
	return server.CreateSpark54FD80(server.SparkCreateRuntime54FD80{
		NewObject: world.NewObjectByTypeID,
		CreateAt: func(spark, owner *server.Object, pos types.Pointf) {
			outer.CreateObjectAt(spark, owner, pos)
		},
		Frame: world.Frame,
		Raise: func(spark *server.Object, height float32) {
			spark.Raise(height)
		},
	}, types.Ptf(x, y), kind, lifetime, types.Ptf(velocityX, velocityY), z, owner)
}
