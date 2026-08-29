package legacy

import (
	"unsafe"

	"github.com/opennox/libs/spell"

	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/server"
)

type glyphXferNativeDeps4F5890 struct {
	spellID           func(string) uint32
	spellName         func(uint32) string
	transferInventory func(uint16, *server.Object, int32) int32
}

func glyphXferRuntimeDeps4F5890() glyphXferNativeDeps4F5890 {
	return glyphXferNativeDeps4F5890{
		spellID: func(name string) uint32 {
			id := spell.ParseID(name)
			if id <= 0 {
				return 0
			}
			return uint32(id)
		},
		spellName: func(id uint32) string {
			return spell.ID(id).String()
		},
		transferInventory: func(version uint16, object *server.Object, count int32) int32 {
			return xferInventoryCall4F3E30(object, version, count)
		},
	}
}

func glyphXferReadWriteNative4F5890(
	cf *cryptfile.CryptFile,
	pointer unsafe.Pointer,
	size int,
) {
	_, _ = cf.ReadWrite(unsafe.Slice((*byte)(pointer), size))
}

func glyphXferReadWriteBytesNative4F5890(cf *cryptfile.CryptFile, data []byte) {
	if len(data) == 0 {
		return
	}
	glyphXferReadWriteNative4F5890(cf, unsafe.Pointer(&data[0]), len(data))
}

func glyphXferNative4F5890(
	cf *cryptfile.CryptFile,
	object *server.Object,
	deps glyphXferNativeDeps4F5890,
) int32 {
	return glyphXfer4F5890(
		object,
		glyphXferDeps4F5890[*server.Object, *server.GlyphInitData, string]{
			loadGlyphData: func(object *server.Object) *server.GlyphInitData {
				// Preserve the entry pointer without allocation or class validation.
				return object.InitDataGlyph()
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
			rwLegacyDword: func() {
				var value uint32
				glyphXferReadWriteNative4F5890(cf, unsafe.Pointer(&value), 4)
			},
			rwDirection1: func(object *server.Object) {
				glyphXferReadWriteNative4F5890(cf, unsafe.Pointer(&object.Direction1), 1)
			},
			rwTargetX: func(data *server.GlyphInitData) {
				glyphXferReadWriteNative4F5890(cf, unsafe.Pointer(&data.SpellArg.Pos.X), 4)
			},
			rwTargetY: func(data *server.GlyphInitData) {
				glyphXferReadWriteNative4F5890(cf, unsafe.Pointer(&data.SpellArg.Pos.Y), 4)
			},
			rwSpellCount: func(data *server.GlyphInitData) {
				glyphXferReadWriteNative4F5890(cf, unsafe.Pointer(&data.SpellsCnt), 1)
			},
			readOnly: func() int32 {
				if cf.ReadOnly() {
					return 1
				}
				return 0
			},
			loadSpellCount: func(data *server.GlyphInitData) uint8 {
				return uint8(data.SpellsCnt)
			},
			rwLegacySpells: func(data *server.GlyphInitData) {
				glyphXferReadWriteNative4F5890(cf, unsafe.Pointer(&data.Spells[0]), 20)
			},
			rwNameLength: func(value uint8) uint8 {
				glyphXferReadWriteNative4F5890(cf, unsafe.Pointer(&value), 1)
				return value
			},
			rwNameBytes: func(value []byte) {
				glyphXferReadWriteBytesNative4F5890(cf, value)
			},
			spellID: deps.spellID,
			storeSpell: func(data *server.GlyphInitData, index int, value uint32) {
				data.Spells[index] = value
			},
			loadSpell: func(data *server.GlyphInitData, index int) uint32 {
				return data.Spells[index]
			},
			spellName: deps.spellName,
			spellNameLength: func(name string) uint8 {
				if name == "" {
					panic("invalid GlyphXfer spell ID")
				}
				return uint8(len(name))
			},
			rwSpellNameBytes: func(name string, length uint8) {
				glyphXferReadWriteBytesNative4F5890(cf, []byte(name)[:int(length)])
			},
			copyDirection: func(object *server.Object) {
				object.Direction2 = object.Direction1
			},
			clearSpellTargetObject: func(data *server.GlyphInitData) {
				data.SpellArg.Obj = nil
			},
			transferInventory: deps.transferInventory,
			storeField34: func(object *server.Object, value uint32) {
				object.Field34 = value
			},
		},
	)
}

func Nox_xxx_XFerGlyphNative4F5890(
	cf *cryptfile.CryptFile,
	object *server.Object,
) int32 {
	return glyphXferNative4F5890(cf, object, glyphXferRuntimeDeps4F5890())
}
