package opennox

import (
	"unsafe"

	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

var spellRuntimeObjectTypeOffsets4FC9B0 = [...]uintptr{
	1569676,
	1569680,
	1569684,
	1569688,
	1569692,
	1569696,
	1569700,
}

type spellRuntimeInitNativeDeps4FC9B0 struct {
	initDurations          func() int32
	newMagicClass          func(name string, recordSize uintptr, capacity int) alloc.ClassT[server.MagicEntityClass]
	storeMagicClass        func(alloc.ClassT[server.MagicEntityClass])
	newObjectByTypeID      func(name string) *server.Object
	storeImaginaryCaster   func(*server.Object)
	createObjectAt         func(object, owner *server.Object, pos types.Pointf)
	objectTypeIDByName     func(name string) int32
	storeSpellObjectTypeID func(index int, value uint32)
}

func spellRuntimeInitNative4FC9B0(deps spellRuntimeInitNativeDeps4FC9B0) int32 {
	return spellRuntimeInit4FC9B0(
		unsafe.Sizeof(server.MagicEntityClass{}),
		spellRuntimeInitHooks4FC9B0[alloc.ClassT[server.MagicEntityClass], *server.Object]{
			initDurations:        deps.initDurations,
			newMagicClass:        deps.newMagicClass,
			storeMagicClass:      deps.storeMagicClass,
			newObjectByTypeID:    deps.newObjectByTypeID,
			storeImaginaryCaster: deps.storeImaginaryCaster,
			createObjectAt: func(object, owner *server.Object, x, y float32) {
				deps.createObjectAt(object, owner, types.Pointf{X: x, Y: y})
			},
			objectTypeIDByName: func(name string) uint32 {
				return uint32(deps.objectTypeIDByName(name))
			},
			storeSpellObjectTypeID: deps.storeSpellObjectTypeID,
		},
	)
}

// nox_xxx_allocSpellRelatedArrays_4FC9B0 binds the original initializer to
// native-width allocator and Object handles. There is no public C entrypoint:
// the only decoded GAME.EXE caller is the Go-owned new-session initializer.
func nox_xxx_allocSpellRelatedArrays_4FC9B0() int32 {
	s := noxServer
	return spellRuntimeInitNative4FC9B0(spellRuntimeInitNativeDeps4FC9B0{
		initDurations: func() int32 {
			if s.Spells.Init() {
				return 1
			}
			return 0
		},
		newMagicClass: func(name string, recordSize uintptr, capacity int) alloc.ClassT[server.MagicEntityClass] {
			return alloc.ClassT[server.MagicEntityClass]{
				Class: alloc.NewClass(name, recordSize, capacity),
			}
		},
		storeMagicClass: func(value alloc.ClassT[server.MagicEntityClass]) {
			magicEntityAlloc = value
		},
		newObjectByTypeID: s.NewObjectByTypeID,
		storeImaginaryCaster: func(value *server.Object) {
			nox_xxx_imagCasterUnit_1569664 = value
		},
		createObjectAt: func(object, owner *server.Object, pos types.Pointf) {
			s.CreateObjectAt(object, owner, pos)
		},
		objectTypeIDByName: func(name string) int32 {
			return int32(s.Types.IndByID(name))
		},
		storeSpellObjectTypeID: func(index int, value uint32) {
			if index == 0 {
				noxPixieObjID = int(value)
			}
			*memmap.PtrUint32(0x5D4594, spellRuntimeObjectTypeOffsets4FC9B0[index]) = value
		},
	})
}
