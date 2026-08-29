package legacy

import (
	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/server"
)

type abilityRewardXferNativeDeps4F6240 struct {
	abilityName       func(uint8) string
	abilityID         func(string) int32
	transferInventory func(uint16, *server.Object, int32) int32
}

func abilityRewardXferRuntimeDeps4F6240() abilityRewardXferNativeDeps4F6240 {
	return abilityRewardXferNativeDeps4F6240{
		abilityName: func(id uint8) string {
			return server.Ability(id).String()
		},
		abilityID: func(name string) int32 {
			for id, candidate := range server.AbilityNames {
				if candidate == name {
					return int32(id)
				}
			}
			return int32(server.AbilityInvalid)
		},
		transferInventory: func(version uint16, object *server.Object, count int32) int32 {
			return xferInventoryCall4F3E30(object, version, count)
		},
	}
}

func abilityRewardXferNative4F6240(
	cf *cryptfile.CryptFile,
	object *server.Object,
	deps abilityRewardXferNativeDeps4F6240,
) int32 {
	return abilityRewardXfer4F6240(
		object,
		abilityRewardXferDeps4F6240[*server.Object, *server.AbilityRewardUseData]{
			loadField34: func(object *server.Object) uint32 {
				return object.Field34
			},
			loadUseData: func(object *server.Object) *server.AbilityRewardUseData {
				return object.UseDataAbilityReward()
			},
			rwVersion: func(value uint16) uint16 {
				return objectReadOldRWU16Native4F4170(cf, value)
			},
			mapReadWrite: func(object *server.Object, mapVersion int32) int32 {
				return objectMapReadWriteNative4F4530(cf, object, mapVersion)
			},
			loadAbility: func(data *server.AbilityRewardUseData) uint8 {
				return data.Ability
			},
			abilityName: deps.abilityName,
			rwByte: func(value uint8) uint8 {
				return objectReadOldRWU8Native4F4170(cf, value)
			},
			rwBytes: func(value []byte) {
				_, _ = cf.ReadWrite(value)
			},
			abilityID: deps.abilityID,
			storeAbility: func(data *server.AbilityRewardUseData, value uint8) {
				data.Ability = value
			},
			readMode: func() int32 {
				if cf.ReadOnly() {
					return 1
				}
				return 0
			},
			transferInventory: deps.transferInventory,
			storeField34: func(object *server.Object, value uint32) {
				object.Field34 = value
			},
		},
	)
}

func Nox_xxx_XFerAbilityRewardNative4F6240(
	cf *cryptfile.CryptFile,
	object *server.Object,
) int32 {
	return abilityRewardXferNative4F6240(cf, object, abilityRewardXferRuntimeDeps4F6240())
}
