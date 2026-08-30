package legacy

import (
	"bytes"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/server"
)

type ammoXferNativeDeps4F6B20 struct {
	modifierIDByName  func(string) int32
	modifierByID      func(int32) *server.ModifierEff
	applyModifiers    func(*server.Object, *server.ModifierInitData)
	gameFlag4096      func() bool
	transferInventory func(uint16, *server.Object, int32) int32
}

func ammoXferRuntimeDeps4F6B20() ammoXferNativeDeps4F6B20 {
	srv := GetServer().S()
	return ammoXferNativeDeps4F6B20{
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
		transferInventory: func(version uint16, object *server.Object, count int32) int32 {
			return xferInventoryCall4F3E30(object, version, count)
		},
	}
}

func ammoXferReadModifierNameNative4F6B20(cf *cryptfile.CryptFile, size uint8) string {
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

func ammoXferWriteModifierNameNative4F6B20(
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

func ammoXferLoadUseByteNative4F6B20(data *server.AmmoUseData, index int) uint8 {
	switch index {
	case 0:
		return data.Charge0
	case 1:
		return data.Charge1
	case 2:
		return data.Field2
	default:
		panic("AmmoXfer 004F6B20: invalid use-data byte index")
	}
}

func ammoXferStoreUseByteNative4F6B20(data *server.AmmoUseData, index int, value uint8) {
	switch index {
	case 0:
		data.Charge0 = value
	case 1:
		data.Charge1 = value
	case 2:
		data.Field2 = value
	default:
		panic("AmmoXfer 004F6B20: invalid use-data byte index")
	}
}

func ammoXferNative4F6B20(
	cf *cryptfile.CryptFile,
	object *server.Object,
	deps ammoXferNativeDeps4F6B20,
) int32 {
	return ammoXfer4F6B20(
		object,
		ammoXferDeps4F6B20[
			*server.Object,
			*server.ModifierInitData,
			*server.ModifierEff,
			*server.ModifierEff,
			*server.AmmoUseData,
		]{
			loadUseData: func(object *server.Object) *server.AmmoUseData {
				return object.UseDataAmmo()
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
			mapReadWrite: func(object *server.Object, mapVersion int32) int32 {
				return objectMapReadWriteNative4F4530(cf, object, mapVersion)
			},
			readMode: func() int32 {
				if cf.ReadOnly() {
					return 1
				}
				return 0
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
				ammoXferWriteModifierNameNative4F6B20(cf, modifier, size)
			},
			readModifierName: func(size uint8) string {
				return ammoXferReadModifierNameNative4F6B20(cf, size)
			},
			modifierIDByName: deps.modifierIDByName,
			modifierByID:     deps.modifierByID,
			applyModifiers: func(object *server.Object, modifiers [4]*server.ModifierEff, tail uint32) {
				deps.applyModifiers(object, &server.ModifierInitData{
					Modifiers: modifiers,
					Field16:   tail,
				})
			},
			loadUseByte:       ammoXferLoadUseByteNative4F6B20,
			storeUseByte:      ammoXferStoreUseByteNative4F6B20,
			gameFlag4096:      deps.gameFlag4096,
			transferInventory: deps.transferInventory,
		},
	)
}

func Nox_xxx_XFerAmmoNative4F6B20(
	cf *cryptfile.CryptFile,
	object *server.Object,
) int32 {
	return ammoXferNative4F6B20(cf, object, ammoXferRuntimeDeps4F6B20())
}
