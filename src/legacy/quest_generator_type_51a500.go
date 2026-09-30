package legacy

import (
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

type questGeneratorTypeDeps51A500 struct {
	ready         func() uint32
	initialize    func()
	generatorName func(int) *byte
	objectType    func(*server.Object) uint16
	creatureType  func(int) uint32
	generatorType func(int) uint32
}

func questGeneratorTypeNative51A500(unit *server.Object, d questGeneratorTypeDeps51A500) int32 {
	// 0051A509 initializes even when the supplied object is nil. A nonzero
	// ready flag suppresses initialization; there is no second ready check.
	if d.ready() == 0 {
		d.initialize()
	}
	if unit == nil || d.generatorName(0) == nil {
		return 0
	}
	// 0051A525 reads and zero-extends the type index once, before the loop.
	// The recovered C instead reloads it inside the loop; follow GAME.EXE.
	typeID := uint32(d.objectType(unit))
	for index := 0; ; index++ {
		if d.creatureType(index) == typeID {
			return int32(d.generatorType(index))
		}
		if d.generatorName(index+1) == nil {
			return 0
		}
	}
}

var questGeneratorTypeDepsFactory51A500 = func() questGeneratorTypeDeps51A500 {
	return questGeneratorTypeDeps51A500{
		ready: func() uint32 {
			return memmap.Uint32(0x5D4594, 2388664)
		},
		initialize: func() {
			questGeneratorTypesInitCEntry51A550()
		},
		generatorName: questGeneratorName51A550,
		objectType: func(unit *server.Object) uint16 {
			return unit.TypeInd
		},
		creatureType: func(index int) uint32 {
			return memmap.Uint32(0x587000, uintptr(249900+16*index))
		},
		generatorType: func(index int) uint32 {
			return memmap.Uint32(0x587000, uintptr(249908+16*index))
		},
	}
}
