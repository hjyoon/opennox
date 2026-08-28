package legacy

import (
	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/server"
)

func Nox_xxx_XFerSpellPagePedestalNative4F4A20(
	cf *cryptfile.CryptFile,
	object *server.Object,
) int32 {
	return spellPagePedestalXfer4F4A20(
		object,
		spellPagePedestalXferDeps4F4A20[*server.Object, *server.AwardSpellCollideData]{
			loadField34: func(object *server.Object) uint32 {
				return object.Field34
			},
			rwVersion: func(value uint16) uint16 {
				return objectReadOldRWU16Native4F4170(cf, value)
			},
			mapReadWrite: func(object *server.Object, mapVersion int32) int32 {
				return objectMapReadWriteNative4F4530(cf, object, mapVersion)
			},
			loadCollideData: func(object *server.Object) *server.AwardSpellCollideData {
				return (*server.AwardSpellCollideData)(object.CollideData)
			},
			rwSpellPayload: func(data *server.AwardSpellCollideData) {
				data.Spell = objectReadOldRWU32Native4F4170(cf, data.Spell)
			},
			readOnly: func() int32 {
				if cf.ReadOnly() {
					return 1
				}
				return 0
			},
			transferInventory: func(version uint16, object *server.Object, count int32) int32 {
				return xferInventoryCall4F3E30(object, version, count)
			},
			storeField34: func(object *server.Object, value uint32) {
				object.Field34 = value
			},
		},
	)
}
