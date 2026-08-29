package legacy

import (
	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/server"
)

type fieldGuideXferNativeDeps4F6390 struct {
	transferInventory func(uint16, *server.Object, int32) int32
}

func fieldGuideXferRuntimeDeps4F6390() fieldGuideXferNativeDeps4F6390 {
	return fieldGuideXferNativeDeps4F6390{
		transferInventory: func(version uint16, object *server.Object, count int32) int32 {
			return xferInventoryCall4F3E30(object, version, count)
		},
	}
}

func fieldGuideCreatureLength4F6390(data *server.FieldGuideUseData) uint32 {
	for i, value := range data.CreatureBuf {
		if value == 0 {
			return uint32(i)
		}
	}
	// GAME.EXE uses an unbounded strlen here. The native Go record has no
	// addressable storage beyond its exact 64 bytes, so fail explicitly
	// instead of inventing a bounded zero result or reading unrelated memory.
	panic("FieldGuideXfer 004F6390: creature name has no NUL terminator")
}

func fieldGuideXferNative4F6390(
	cf *cryptfile.CryptFile,
	object *server.Object,
	deps fieldGuideXferNativeDeps4F6390,
) int32 {
	return fieldGuideXfer4F6390(
		object,
		fieldGuideXferDeps4F6390[*server.Object, *server.FieldGuideUseData]{
			loadUseData: func(object *server.Object) *server.FieldGuideUseData {
				return object.UseDataFieldGuide()
			},
			loadField34: func(object *server.Object) uint32 {
				return object.Field34
			},
			rwVersion: func(value uint16) uint16 {
				return objectReadOldRWU16Native4F4170(cf, value)
			},
			mapReadWrite: func(object *server.Object, mapVersion int32) int32 {
				return objectMapReadWriteNative4F4530(cf, object, mapVersion)
			},
			readMode: func() int32 {
				if cf.ReadOnly() {
					return 1
				}
				return 0
			},
			creatureLength: fieldGuideCreatureLength4F6390,
			rwByte: func(value uint8) uint8 {
				return objectReadOldRWU8Native4F4170(cf, value)
			},
			rwCreature: func(data *server.FieldGuideUseData, size uint8) {
				if size == 0 {
					_, _ = cf.ReadWrite(nil)
					return
				}
				_, _ = cf.ReadWrite(data.CreatureBuf[:int(size)])
			},
			storeCreatureTerminator: func(data *server.FieldGuideUseData, index uint8) {
				data.CreatureBuf[index] = 0
			},
			transferInventory: deps.transferInventory,
			storeField34: func(object *server.Object, value uint32) {
				object.Field34 = value
			},
		},
	)
}

func Nox_xxx_XFerFieldGuideNative4F6390(
	cf *cryptfile.CryptFile,
	object *server.Object,
) int32 {
	return fieldGuideXferNative4F6390(cf, object, fieldGuideXferRuntimeDeps4F6390())
}
