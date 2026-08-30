package legacy

import (
	"bytes"

	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/server"
)

type teamXferNativeDeps4F6D20 struct {
	modifierIDByName  func(string) int32
	modifierByID      func(int32) *server.ModifierEff
	applyModifiers    func(*server.Object, *server.ModifierInitData)
	transferInventory func(uint16, *server.Object, int32) int32
}

func teamXferRuntimeDeps4F6D20() teamXferNativeDeps4F6D20 {
	srv := GetServer().S()
	return teamXferNativeDeps4F6D20{
		modifierIDByName: func(name string) int32 {
			return int32(srv.Modif.Nox_xxx_modifGetIdByName413290(name))
		},
		modifierByID: func(id int32) *server.ModifierEff {
			return srv.Modif.Nox_xxx_modifGetDescById413330(int(id))
		},
		applyModifiers: func(object *server.Object, attrs *server.ModifierInitData) {
			srv.ApplyModifierAttrs4E4990(object, attrs)
		},
		transferInventory: func(version uint16, object *server.Object, count int32) int32 {
			return xferInventoryCall4F3E30(object, version, count)
		},
	}
}

func teamXferReadModifierNameNative4F6D20(cf *cryptfile.CryptFile, size uint8) string {
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

func teamXferWriteModifierNameNative4F6D20(
	cf *cryptfile.CryptFile,
	modifier *server.ModifierEff,
	size uint8,
) {
	if size == 0 {
		_, _ = cf.ReadWrite(nil)
		return
	}
	name := []byte(modifier.Name())
	// A live descriptor replacement shorter than the already-written length
	// faults here, matching the original unchecked memory-read boundary.
	_, _ = cf.ReadWrite(name[:int(size)])
}

func teamXferNative4F6D20(
	cf *cryptfile.CryptFile,
	object *server.Object,
	deps teamXferNativeDeps4F6D20,
) int32 {
	return teamXfer4F6D20(
		object,
		teamXferDeps4F6D20[
			*server.Object,
			*server.ModifierInitData,
			*server.ModifierEff,
			*server.ModifierEff,
			*server.FlagUpdateData4EA490,
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
				teamXferWriteModifierNameNative4F6D20(cf, modifier, size)
			},
			readModifierName: func(size uint8) string {
				return teamXferReadModifierNameNative4F6D20(cf, size)
			},
			modifierIDByName: deps.modifierIDByName,
			modifierByID:     deps.modifierByID,
			applyModifiers: func(object *server.Object, modifiers [4]*server.ModifierEff, tail uint32) {
				deps.applyModifiers(object, &server.ModifierInitData{
					Modifiers: modifiers,
					Field16:   tail,
				})
			},
			loadObjClass: func(object *server.Object) uint32 {
				return uint32(object.ObjClass)
			},
			loadUpdateData: func(object *server.Object) *server.FlagUpdateData4EA490 {
				return (*server.FlagUpdateData4EA490)(object.UpdateData)
			},
			loadPositionX: func(object *server.Object) float32 {
				return object.PosVec.X
			},
			storeUpdatePositionX: func(update *server.FlagUpdateData4EA490, value float32) {
				update.Home.X = value
			},
			loadPositionY: func(object *server.Object) float32 {
				return object.PosVec.Y
			},
			storeUpdatePositionY: func(update *server.FlagUpdateData4EA490, value float32) {
				update.Home.Y = value
			},
			transferInventory: deps.transferInventory,
		},
	)
}

func Nox_xxx_XFerTeamNative4F6D20(
	cf *cryptfile.CryptFile,
	object *server.Object,
) int32 {
	return teamXferNative4F6D20(cf, object, teamXferRuntimeDeps4F6D20())
}
