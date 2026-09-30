package legacy

import (
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

// GAME.EXE 005C4028 stores 53 PE32 records, each containing creature name,
// creature ID, generator name, and generator ID. Native pointers cannot occupy
// those four-byte slots: publishing an ID would overwrite half of a name.
// Blob initialization already resolves these sealed names into PtrPtr's native
// side slots. Use those live slots, not the raw PE32 bytes underneath them.

func questGeneratorName51A550(index int) *byte {
	return (*byte)(*memmap.PtrPtr(0x587000, uintptr(249904+16*index)))
}

func questGeneratorCreatureName51A550(index int) *byte {
	return (*byte)(*memmap.PtrPtr(0x587000, uintptr(249896+16*index)))
}

type questGeneratorTypesInitDeps51A550 struct {
	generatorName, creatureName   func(int) *byte
	lookupType                    func(*byte) uint32
	storeGenerator, storeCreature func(int, uint32)
	storeReady                    func(uint32)
}

func questGeneratorTypesInitNative51A550(d questGeneratorTypesInitDeps51A550) *byte {
	name := d.generatorName(0)
	for index := 0; name != nil; index++ {
		// 0051A565 publishes the generator result before reading the live
		// creature name at 0051A568. The next name is read after both calls.
		d.storeGenerator(index, d.lookupType(name))
		d.storeCreature(index, d.lookupType(d.creatureName(index)))
		name = d.generatorName(index + 1)
	}
	// This function has no ready guard and also sets ready for an empty
	// table or unresolved names. Its normal return is the nil terminator.
	d.storeReady(1)
	return name
}

var questGeneratorTypesInitDepsFactory51A550 = func() questGeneratorTypesInitDeps51A550 {
	return questGeneratorTypesInitDeps51A550{
		generatorName: questGeneratorName51A550,
		creatureName:  questGeneratorCreatureName51A550,
		lookupType: func(name *byte) uint32 {
			return uint32(GetServer().S().Types.IndByID(alloc.GoString(name)))
		},
		storeGenerator: func(index int, id uint32) {
			*memmap.PtrUint32(0x587000, uintptr(249908+16*index)) = id
		},
		storeCreature: func(index int, id uint32) {
			*memmap.PtrUint32(0x587000, uintptr(249900+16*index)) = id
		},
		storeReady: func(ready uint32) {
			*memmap.PtrUint32(0x5D4594, 2388664) = ready
		},
	}
}
