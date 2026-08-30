package legacy

import (
	"bytes"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/server"
)

type armorXferNativeDeps4F6860 struct {
	modifierIDByName  func(string) int32
	modifierByID      func(int32) *server.ModifierEff
	applyModifiers    func(*server.Object, *server.ModifierInitData)
	gameFlag4096      func() bool
	unitSetHP         func(*server.Object, uint16)
	switchToSolo      func() int32
	notMultiplayer    func() int32
	anyTrackedPlayers func() int32
	armorClass        func(uint16) *server.Modifier
	transferInventory func(uint16, *server.Object, int32) int32
}

func armorXferRuntimeDeps4F6860() armorXferNativeDeps4F6860 {
	srv := GetServer().S()
	return armorXferNativeDeps4F6860{
		modifierIDByName: func(name string) int32 {
			return int32(srv.Modif.Nox_xxx_modifGetIdByName413290(name))
		},
		modifierByID: func(id int32) *server.ModifierEff {
			return srv.Modif.Nox_xxx_modifGetDescById413330(int(id))
		},
		applyModifiers: func(object *server.Object, attrs *server.ModifierInitData) {
			srv.ApplyModifierAttrs4E4990(object, attrs)
		},
		gameFlag4096: func() bool {
			return noxflags.HasGame(noxflags.GameModeQuest)
		},
		unitSetHP: func(object *server.Object, value uint16) {
			Nox_xxx_unitSetHP_4E4560(object, value)
		},
		switchToSolo: func() int32 {
			return int32(bool2int(Nox_xxx_gameIsSwitchToSolo_4DB240()))
		},
		notMultiplayer: func() int32 {
			return int32(bool2int(Nox_xxx_gameIsNotMultiplayer_4DB250()))
		},
		anyTrackedPlayers: func() int32 {
			return int32(bool2int(srv.Players.AnyXxx()))
		},
		armorClass: func(typeIndex uint16) *server.Modifier {
			return srv.Modif.Nox_xxx_equipClothFindDefByTT413270(int(typeIndex))
		},
		transferInventory: func(version uint16, object *server.Object, count int32) int32 {
			return xferInventoryCall4F3E30(object, version, count)
		},
	}
}

func armorXferReadModifierNameNative4F6860(cf *cryptfile.CryptFile, size uint8) string {
	var buffer [256]byte
	if size != 0 {
		_, _ = cf.ReadWrite(buffer[:int(size)])
	} else {
		_, _ = cf.ReadWrite(nil)
	}
	buffer[int(size)] = 0
	end := int(size)
	if index := bytes.IndexByte(buffer[:end], 0); index >= 0 {
		end = index
	}
	return string(buffer[:end])
}

func armorXferWriteModifierNameNative4F6860(
	cf *cryptfile.CryptFile,
	modifier *server.ModifierEff,
	size uint8,
) {
	if size == 0 {
		_, _ = cf.ReadWrite(nil)
		return
	}
	name := []byte(modifier.Name())
	// A live descriptor replacement that is shorter than the already-written
	// length faults here, matching the original unchecked memory read boundary.
	_, _ = cf.ReadWrite(name[:int(size)])
}

func armorXferNative4F6860(
	cf *cryptfile.CryptFile,
	object *server.Object,
	deps armorXferNativeDeps4F6860,
) int32 {
	return armorXfer4F6860(
		object,
		armorXferDeps4F6860[
			*server.Object,
			*server.ModifierInitData,
			*server.ModifierEff,
			*server.ModifierEff,
			*server.HealthData,
			*server.Modifier,
			*server.WeaponArmorUpdateData,
		]{
			loadField34: func(object *server.Object) uint32 {
				return object.Field34
			},
			storeField34: func(object *server.Object, value uint32) {
				object.Field34 = value
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
			applyLegacyEmptyModifiers: func(object *server.Object) {
				// Armor modifier application observes the four slots only. The PE32
				// routine left Field16 uninitialized in this legacy branch.
				deps.applyModifiers(object, &server.ModifierInitData{})
			},
			loadModifierData: func(object *server.Object) *server.ModifierInitData {
				return object.InitDataModifier()
			},
			loadModifierSlot: func(data *server.ModifierInitData, index int) *server.ModifierEff {
				return data.Modifiers[index]
			},
			modifierNameLength: func(modifier *server.ModifierEff) uint32 {
				return uint32(len(modifier.Name()))
			},
			rwByte: func(value uint8) uint8 {
				return objectReadOldRWU8Native4F4170(cf, value)
			},
			rwModifierName: func(modifier *server.ModifierEff, size uint8) {
				armorXferWriteModifierNameNative4F6860(cf, modifier, size)
			},
			readModifierName: func(size uint8) string {
				return armorXferReadModifierNameNative4F6860(cf, size)
			},
			modifierIDByName: deps.modifierIDByName,
			modifierByID:     deps.modifierByID,
			applyModifiers: func(object *server.Object, modifiers [4]*server.ModifierEff, tail uint32) {
				deps.applyModifiers(object, &server.ModifierInitData{
					Modifiers: modifiers,
					Field16:   tail,
				})
			},
			unitGetHP: func(object *server.Object) uint16 {
				return server.UnitGetHP4EE780(object)
			},
			rwWord: func(value uint16) uint16 {
				return objectReadOldRWU16Native4F4170(cf, value)
			},
			loadHealthData: func(object *server.Object) *server.HealthData {
				return object.HealthData
			},
			loadHealthMaximum: func(health *server.HealthData) uint16 {
				return health.Max
			},
			switchToSolo:      deps.switchToSolo,
			notMultiplayer:    deps.notMultiplayer,
			gameFlag4096:      deps.gameFlag4096,
			anyTrackedPlayers: deps.anyTrackedPlayers,
			unitSetHP:         deps.unitSetHP,
			loadTypeIndex: func(object *server.Object) uint16 {
				return object.TypeInd
			},
			armorClass: deps.armorClass,
			loadArmorHP: func(armor *server.Modifier) uint16 {
				return uint16(armor.Durability52)
			},
			storeHealthMaximum: func(health *server.HealthData, value uint16) {
				health.Max = value
			},
			storeHealthField2: func(health *server.HealthData, value uint16) {
				health.Field2 = value
			},
			rwDummyByte: func() {
				// GAME.EXE used an uninitialized stack byte here. Preserve the one-byte
				// stream event with a deterministic write value on native platforms.
				_ = objectReadOldRWU8Native4F4170(cf, 0)
			},
			loadUpdateData: func(object *server.Object) *server.WeaponArmorUpdateData {
				return object.UpdateDataWeaponArmor()
			},
			rwUpdateField4: func(data *server.WeaponArmorUpdateData) {
				data.Field4 = objectReadOldRWU32Native4F4170(cf, data.Field4)
			},
			transferInventory: deps.transferInventory,
		},
	)
}

func Nox_xxx_XFerArmorNative4F6860(
	cf *cryptfile.CryptFile,
	object *server.Object,
) int32 {
	return armorXferNative4F6860(cf, object, armorXferRuntimeDeps4F6860())
}
