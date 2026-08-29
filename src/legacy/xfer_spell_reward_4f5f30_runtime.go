package legacy

import (
	"github.com/opennox/libs/spell"

	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/server"
)

type spellRewardXferNativeDeps4F5F30 struct {
	spellName         func(uint8) string
	spellID           func(string) uint8
	transferInventory func(uint16, *server.Object, int32) int32
}

func spellRewardXferRuntimeDeps4F5F30() spellRewardXferNativeDeps4F5F30 {
	return spellRewardXferNativeDeps4F5F30{
		spellName: func(id uint8) string {
			name := spell.ID(id).String()
			if name == "" {
				panic("invalid SpellRewardXfer spell ID")
			}
			return name
		},
		spellID: func(name string) uint8 {
			id := spell.ParseID(name)
			if id <= 0 || id > 0xff {
				return 0
			}
			return uint8(id)
		},
		transferInventory: func(version uint16, object *server.Object, count int32) int32 {
			return xferInventoryCall4F3E30(object, version, count)
		},
	}
}

func spellRewardXferNative4F5F30(
	cf *cryptfile.CryptFile,
	object *server.Object,
	deps spellRewardXferNativeDeps4F5F30,
) int32 {
	return spellRewardXfer4F5F30(
		object,
		spellRewardXferDeps4F5F30[*server.Object, *server.SpellRewardUseData]{
			loadField34: func(object *server.Object) uint32 {
				return object.Field34
			},
			loadUseData: func(object *server.Object) *server.SpellRewardUseData {
				return object.UseDataSpellReward()
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
			rwByte: func(value uint8) uint8 {
				return objectReadOldRWU8Native4F4170(cf, value)
			},
			rwBytes: func(value []byte) {
				_, _ = cf.ReadWrite(value)
			},
			loadSpell: func(data *server.SpellRewardUseData) uint8 {
				return data.Spell
			},
			storeSpell: func(data *server.SpellRewardUseData, value uint8) {
				data.Spell = value
			},
			spellName:         deps.spellName,
			spellID:           deps.spellID,
			transferInventory: deps.transferInventory,
			storeField34: func(object *server.Object, value uint32) {
				object.Field34 = value
			},
		},
	)
}

func Nox_xxx_XFerSpellRewardNative4F5F30(
	cf *cryptfile.CryptFile,
	object *server.Object,
) int32 {
	return spellRewardXferNative4F5F30(cf, object, spellRewardXferRuntimeDeps4F5F30())
}
