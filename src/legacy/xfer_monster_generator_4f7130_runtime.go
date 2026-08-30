package legacy

import (
	"unsafe"

	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/server"
)

type monsterGeneratorXferNativeDeps4F7130 struct {
	loadTypeName        func(*server.Object) []byte
	transferScript      func(*server.ScriptCallback, unsafe.Pointer) int32
	saveObject          func(*server.Object) int32
	newObjectByTypeName func([]byte) *server.Object
	callObjectXfer      func(*server.Object) int32
	transferInventory   func(uint16, *server.Object, int32) int32
}

func monsterGeneratorXferRuntimeDeps4F7130(
	cf *cryptfile.CryptFile,
) monsterGeneratorXferNativeDeps4F7130 {
	s := GetServer().S()
	scriptDeps := scriptHandlerXferRuntimeDeps4F5580(cf)
	return monsterGeneratorXferNativeDeps4F7130{
		loadTypeName: func(object *server.Object) []byte {
			return []byte(s.Types.ByInd(int(uint16(object.TypeInd))).ID())
		},
		transferScript: func(handler *server.ScriptCallback, context unsafe.Pointer) int32 {
			return scriptHandlerXferNative4F5580(handler, context, scriptDeps)
		},
		saveObject: func(object *server.Object) int32 {
			return int32(Nox_xxx_xfer_saveObj51DF90(cf, object))
		},
		newObjectByTypeName: func(name []byte) *server.Object {
			return s.NewObjectByTypeID(cStringBytes528DB0(name))
		},
		callObjectXfer: func(object *server.Object) int32 {
			if err := object.CallXfer(nil); err != nil {
				return 0
			}
			return 1
		},
		transferInventory: func(version uint16, object *server.Object, count int32) int32 {
			if err := xferInventoryNative4F3E30(cf, s, object, version, count); err != nil {
				mapLog.Printf("nox_xxx_XFerMonsterGen_4F7130 inventory: %v", err)
				return 0
			}
			return 1
		},
	}
}

func monsterGeneratorScriptHandler4F7130(
	data *server.MonsterGenUpdateData,
	slot monsterGeneratorScriptSlot4F7130,
) *server.ScriptCallback {
	switch slot {
	case monsterGeneratorScript48_4F7130:
		return (*server.ScriptCallback)(unsafe.Pointer(&data.Field48))
	case monsterGeneratorScript56_4F7130:
		return (*server.ScriptCallback)(unsafe.Pointer(&data.Field56))
	case monsterGeneratorScript72_4F7130:
		return &data.ScriptCollision
	case monsterGeneratorScript64_4F7130:
		return (*server.ScriptCallback)(unsafe.Pointer(&data.Field64))
	default:
		panic("invalid MonsterGeneratorXfer script slot")
	}
}

func monsterGeneratorScriptContext4F7130(
	scriptData unsafe.Pointer,
	offset uintptr,
) unsafe.Pointer {
	if scriptData == nil {
		return nil
	}
	return unsafe.Add(scriptData, offset)
}

func monsterGeneratorXferNative4F7130(
	cf *cryptfile.CryptFile,
	object *server.Object,
	deps monsterGeneratorXferNativeDeps4F7130,
) int32 {
	return monsterGeneratorXfer4F7130(
		object,
		monsterGeneratorXferDeps4F7130[
			*server.Object,
			*server.MonsterGenUpdateData,
			unsafe.Pointer,
		]{
			loadUpdateData: func(object *server.Object) *server.MonsterGenUpdateData {
				// Do not call UpdateDataMonsterGen: GAME.EXE has no class gate.
				return (*server.MonsterGenUpdateData)(object.UpdateData)
			},
			loadField34: func(object *server.Object) uint32 {
				return object.Field34
			},
			storeField34: func(object *server.Object, value uint32) {
				object.Field34 = value
			},
			loadScriptData: func(object *server.Object) unsafe.Pointer {
				return object.Field189
			},
			rwVersion: func(value uint16) uint16 {
				return objectReadOldRWU16Native4F4170(cf, value)
			},
			mapReadWrite: func(object *server.Object, version int32) int32 {
				return objectMapReadWriteNative4F4530(cf, object, version)
			},
			rwSpawnSelectorCount: func(value uint8) uint8 {
				return objectReadOldRWU8Native4F4170(cf, value)
			},
			rwSpawnSelector: func(data *server.MonsterGenUpdateData, index int) {
				data.SpawnRate[index] = objectReadOldRWU8Native4F4170(cf, data.SpawnRate[index])
			},
			rwActiveCount: func(data *server.MonsterGenUpdateData) {
				data.ActiveCount = objectReadOldRWU8Native4F4170(cf, data.ActiveCount)
			},
			rwMaxActive: func(data *server.MonsterGenUpdateData) {
				data.MaxActive = objectReadOldRWU8Native4F4170(cf, data.MaxActive)
			},
			rwFrame88: func(data *server.MonsterGenUpdateData) {
				data.Frame88 = objectReadOldRWU32Native4F4170(cf, data.Frame88)
			},
			transferScript: func(data *server.MonsterGenUpdateData, slot monsterGeneratorScriptSlot4F7130, scriptData unsafe.Pointer, offset uintptr) int32 {
				return deps.transferScript(
					monsterGeneratorScriptHandler4F7130(data, slot),
					monsterGeneratorScriptContext4F7130(scriptData, offset),
				)
			},
			readMode: func() int32 {
				if cf.ReadOnly() {
					return 1
				}
				return 0
			},
			rwPrototypeGroupCount: func(value uint8) uint8 {
				return objectReadOldRWU8Native4F4170(cf, value)
			},
			loadPrototype: func(data *server.MonsterGenUpdateData, index int) *server.Object {
				return data.Field0[index]
			},
			rwPrototypeCount: func(value uint8) uint8 {
				return objectReadOldRWU8Native4F4170(cf, value)
			},
			loadTypeName: deps.loadTypeName,
			rwNameLength: func(value uint8) uint8 {
				return objectReadOldRWU8Native4F4170(cf, value)
			},
			rwNameBytes: func(value []byte) {
				_, _ = cf.ReadWrite(value)
			},
			saveObject: deps.saveObject,
			rwPrototypeTag: func(value uint16) uint16 {
				return objectReadOldRWU16Native4F4170(cf, value)
			},
			readPrototypeCRC: func() {
				var crc [4]byte
				_ = cf.ReadMaybeAlign(crc[:])
			},
			newObjectByTypeName: deps.newObjectByTypeName,
			callObjectXfer:      deps.callObjectXfer,
			storePrototype: func(data *server.MonsterGenUpdateData, index int, object *server.Object) {
				data.Field0[index] = object
			},
			rwQuestSelectorCount: func(value uint8) uint8 {
				return objectReadOldRWU8Native4F4170(cf, value)
			},
			rwQuestSelector: func(data *server.MonsterGenUpdateData, index int) {
				data.QuestSpawnRate[index] = objectReadOldRWU8Native4F4170(cf, data.QuestSpawnRate[index])
			},
			rwField92: func(data *server.MonsterGenUpdateData) {
				data.Field92 = objectReadOldRWU32Native4F4170(cf, data.Field92)
			},
			transferInventory: deps.transferInventory,
		},
	)
}

func Nox_xxx_XFerMonsterGenNative4F7130(
	cf *cryptfile.CryptFile,
	object *server.Object,
) int32 {
	return monsterGeneratorXferNative4F7130(cf, object, monsterGeneratorXferRuntimeDeps4F7130(cf))
}
