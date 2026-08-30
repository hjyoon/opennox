package legacy

import (
	"unsafe"

	"github.com/opennox/libs/spell"

	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/server"
)

type rewardMarkerXferNativeDeps4F74D0 struct {
	spellName         func(int) string
	spellID           func(string) int
	abilityName       func(int) string
	abilityID         func(string) int
	guideName         func(int) string
	guideID           func(string) int
	transferInventory func(uint16, *server.Object, int32) int32
}

func rewardMarkerXferRuntimeDeps4F74D0(
	cf *cryptfile.CryptFile,
) rewardMarkerXferNativeDeps4F74D0 {
	return rewardMarkerXferNativeDeps4F74D0{
		spellName: func(id int) string {
			return spell.ID(id).String()
		},
		spellID: func(name string) int {
			return int(spell.ParseID(name))
		},
		abilityName: func(id int) string {
			return server.Ability(id).String()
		},
		abilityID: func(name string) int {
			for id, candidate := range server.AbilityNames {
				if candidate == name {
					return id
				}
			}
			return int(server.AbilityInvalid)
		},
		guideName: server.RewardFieldGuideName4F0D20,
		guideID:   server.RewardFieldGuideID4F0D20,
		transferInventory: func(version uint16, object *server.Object, count int32) int32 {
			if err := xferInventoryNative4F3E30(cf, GetServer().S(), object, version, count); err != nil {
				mapLog.Printf("nox_xxx_XFerRewardMarker_4F74D0 inventory: %v", err)
				return 0
			}
			return 1
		},
	}
}

func rewardMarkerXferListValue4F74D0(
	data *server.RewardMarkerInitData,
	list rewardMarkerXferList4F74D0,
	index int,
) *uint8 {
	switch list {
	case rewardMarkerXferSpells4F74D0:
		return &data.Spells[index]
	case rewardMarkerXferAbilities4F74D0:
		return &data.Abilities[index]
	case rewardMarkerXferGuides4F74D0:
		return &data.Guides[index]
	default:
		panic("invalid RewardMarkerXfer list")
	}
}

func rewardMarkerXferNative4F74D0(
	cf *cryptfile.CryptFile,
	object *server.Object,
	deps rewardMarkerXferNativeDeps4F74D0,
) int32 {
	return rewardMarkerXfer4F74D0(
		object,
		rewardMarkerXferDeps4F74D0[*server.Object, *server.RewardMarkerInitData]{
			loadInitData: func(object *server.Object) *server.RewardMarkerInitData {
				return (*server.RewardMarkerInitData)(object.InitData)
			},
			loadField34: func(object *server.Object) uint32 {
				return object.Field34
			},
			storeField34: func(object *server.Object, value uint32) {
				object.Field34 = value
			},
			rwVersion: func(value uint16) uint16 {
				return objectReadOldRWU16Native4F4170(cf, value)
			},
			mapReadWrite: func(object *server.Object, version int32) int32 {
				return objectMapReadWriteNative4F4530(cf, object, version)
			},
			rwHeader: func(data *server.RewardMarkerInitData, field rewardMarkerXferHeader4F74D0) {
				switch field {
				case rewardMarkerXferCategoryMask4F74D0:
					data.CategoryMask = objectReadOldRWU32Native4F4170(cf, data.CategoryMask)
				case rewardMarkerXferRewardFlags4F74D0:
					_, _ = cf.ReadWrite(unsafe.Slice(&data.RewardFlags, 4))
				default:
					panic("invalid RewardMarkerXfer header")
				}
			},
			loadListValue: func(data *server.RewardMarkerInitData, list rewardMarkerXferList4F74D0, index int) uint8 {
				return *rewardMarkerXferListValue4F74D0(data, list, index)
			},
			storeListValue: func(data *server.RewardMarkerInitData, list rewardMarkerXferList4F74D0, index int, value uint8) {
				*rewardMarkerXferListValue4F74D0(data, list, index) = value
			},
			rwCount: func(value uint16) uint16 {
				return objectReadOldRWU16Native4F4170(cf, value)
			},
			readMode: func() int32 {
				if cf.ReadOnly() {
					return 1
				}
				return 0
			},
			rwNameLength: func(value uint8) uint8 {
				return objectReadOldRWU8Native4F4170(cf, value)
			},
			rwNameBytes: func(value []byte) {
				_, _ = cf.ReadWrite(value)
			},
			resolveName: func(list rewardMarkerXferList4F74D0, name []byte) int {
				switch list {
				case rewardMarkerXferSpells4F74D0:
					return deps.spellID(string(name))
				case rewardMarkerXferAbilities4F74D0:
					return deps.abilityID(string(name))
				case rewardMarkerXferGuides4F74D0:
					return deps.guideID(string(name))
				default:
					panic("invalid RewardMarkerXfer list")
				}
			},
			loadName: func(list rewardMarkerXferList4F74D0, index int) []byte {
				switch list {
				case rewardMarkerXferSpells4F74D0:
					return []byte(deps.spellName(index))
				case rewardMarkerXferAbilities4F74D0:
					return []byte(deps.abilityName(index))
				case rewardMarkerXferGuides4F74D0:
					return []byte(deps.guideName(index))
				default:
					panic("invalid RewardMarkerXfer list")
				}
			},
			rwField: func(data *server.RewardMarkerInitData, field rewardMarkerXferField4F74D0) {
				switch field {
				case rewardMarkerXferField196_4F74D0:
					data.Field196 = objectReadOldRWU32Native4F4170(cf, data.Field196)
				case rewardMarkerXferField192_4F74D0:
					data.Field192 = objectReadOldRWU32Native4F4170(cf, data.Field192)
				case rewardMarkerXferField200_4F74D0:
					data.Field200 = objectReadOldRWU32Native4F4170(cf, data.Field200)
				case rewardMarkerXferField204_4F74D0:
					data.Field204 = objectReadOldRWU32Native4F4170(cf, data.Field204)
				case rewardMarkerXferField208_4F74D0:
					data.Field208 = objectReadOldRWU32Native4F4170(cf, data.Field208)
				case rewardMarkerXferField212_4F74D0:
					data.ChanceMode = objectReadOldRWU32Native4F4170(cf, data.ChanceMode)
				case rewardMarkerXferField216Low_4F74D0:
					low := objectReadOldRWU8Native4F4170(cf, uint8(data.Field216))
					data.Field216 = data.Field216&^0xff | uint32(low)
				default:
					panic("invalid RewardMarkerXfer field")
				}
			},
			transferInventory: deps.transferInventory,
		},
	)
}

func Nox_xxx_XFerRewardMarkerNative4F74D0(
	cf *cryptfile.CryptFile,
	object *server.Object,
) int32 {
	return rewardMarkerXferNative4F74D0(cf, object, rewardMarkerXferRuntimeDeps4F74D0(cf))
}

var (
	_ = [1]struct{}{}[220-unsafe.Sizeof(server.RewardMarkerInitData{})]
	_ = [1]struct{}{}[0-unsafe.Offsetof(server.RewardMarkerInitData{}.CategoryMask)]
	_ = [1]struct{}{}[4-unsafe.Offsetof(server.RewardMarkerInitData{}.RewardFlags)]
	_ = [1]struct{}{}[5-unsafe.Offsetof(server.RewardMarkerInitData{}.Field5)]
	_ = [1]struct{}{}[8-unsafe.Offsetof(server.RewardMarkerInitData{}.Spells)]
	_ = [1]struct{}{}[145-unsafe.Offsetof(server.RewardMarkerInitData{}.Abilities)]
	_ = [1]struct{}{}[151-unsafe.Offsetof(server.RewardMarkerInitData{}.Guides)]
	_ = [1]struct{}{}[192-unsafe.Offsetof(server.RewardMarkerInitData{}.Field192)]
	_ = [1]struct{}{}[196-unsafe.Offsetof(server.RewardMarkerInitData{}.Field196)]
	_ = [1]struct{}{}[200-unsafe.Offsetof(server.RewardMarkerInitData{}.Field200)]
	_ = [1]struct{}{}[204-unsafe.Offsetof(server.RewardMarkerInitData{}.Field204)]
	_ = [1]struct{}{}[208-unsafe.Offsetof(server.RewardMarkerInitData{}.Field208)]
	_ = [1]struct{}{}[212-unsafe.Offsetof(server.RewardMarkerInitData{}.ChanceMode)]
	_ = [1]struct{}{}[216-unsafe.Offsetof(server.RewardMarkerInitData{}.Field216)]
)
