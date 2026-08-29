package legacy

import (
	"encoding/binary"
	"math"
)

const (
	invLightXferCurrentVersion4F5AA0 = uint16(60)
	invLightXferPayloadSize4F5AA0    = 140
	invLightXferGameMask4F5AA0       = uint32(0x00600000)
	invLightXferStaticMask4F5AA0     = uint32(0x20400000)
)

type invLightXferPart4F5AA0 struct {
	offset int
	size   int
}

var invLightXferParts4F5AA0 = [...]invLightXferPart4F5AA0{
	{0, 4}, {4, 4}, {8, 4}, {12, 4}, {16, 12}, {28, 2}, {30, 2}, {32, 4},
	{40, 2}, {42, 48}, {90, 16}, {106, 16}, {122, 2}, {124, 2}, {126, 2},
	{128, 4}, {134, 2}, {136, 2}, {138, 1},
}

var invLightXferLegacyParts4F5AA0 = [...]invLightXferPart4F5AA0{
	{0, 4}, {4, 4}, {8, 4}, {12, 4}, {16, 12}, {28, 2}, {30, 2}, {32, 4},
}

// invLightXferDeps4F5AA0 exposes every observable object, drawable, mode,
// game-flag, stream, and inventory access in GAME.EXE 004F5AA0. Object and
// drawable identities remain generic so the contract cannot inherit PE32
// pointer truncation.
type invLightXferDeps4F5AA0[O any, D comparable] struct {
	loadField34  func(O) uint32
	rwVersion    func(uint16) uint16
	mapReadWrite func(O, int32) int32

	readMode          func() int32
	gameFlags         func(uint32) int32
	firstDrawable     func() D
	nextDrawable      func(D) D
	loadDrawableCode  func(D) uint32
	loadExtent        func(O) uint32
	loadClass         func(O) uint32
	loadNetCode       func(O) uint32
	staticDrawable    func(uint32) D
	dynamicDrawable   func(uint32) D
	copyDrawableLight func(D, *[invLightXferPayloadSize4F5AA0]byte)
	rwLight           func(*[invLightXferPayloadSize4F5AA0]byte, int, int)
	rwLegacyField43   func(uint8) uint8
	legacyTooBright   func(*[invLightXferPayloadSize4F5AA0]byte) bool
	clampLegacyLight  func(*[invLightXferPayloadSize4F5AA0]byte)
	copyObjectLight   func(O, *[invLightXferPayloadSize4F5AA0]byte)
	transferInventory func(uint16, O, int32) int32
	storeField34      func(O, uint32)
}

// invLightXfer4F5AA0 preserves the entry Field34 cache, signed version gates,
// exact 140-byte PE32 light stream, mode and game-flag reload points, legacy
// defaults and clamp, and the inventory-failure prefix. Static/dynamic
// drawable and object light destinations deliberately have no nil guard.
func invLightXfer4F5AA0[O any, D comparable](
	object O,
	deps invLightXferDeps4F5AA0[O, D],
) int32 {
	originalField34 := deps.loadField34(object)
	var light [invLightXferPayloadSize4F5AA0]byte

	versionWord := deps.rwVersion(invLightXferCurrentVersion4F5AA0)
	version := int16(versionWord)
	if version > int16(invLightXferCurrentVersion4F5AA0) {
		return 0
	}
	if deps.mapReadWrite(object, int32(version)) == 0 {
		return 0
	}

	if deps.readMode() == 0 {
		if deps.gameFlags(invLightXferGameMask4F5AA0) != 0 {
			var zero D
			for drawable := deps.firstDrawable(); drawable != zero; drawable = deps.nextDrawable(drawable) {
				if deps.loadDrawableCode(drawable) == deps.loadExtent(object) {
					deps.copyDrawableLight(drawable, &light)
					break
				}
			}
		} else if deps.loadClass(object)&invLightXferStaticMask4F5AA0 != 0 {
			deps.copyDrawableLight(deps.staticDrawable(deps.loadExtent(object)), &light)
		} else {
			deps.copyDrawableLight(deps.dynamicDrawable(deps.loadNetCode(object)), &light)
		}
	}

	apply := false
	if version >= 2 {
		for _, part := range invLightXferParts4F5AA0 {
			deps.rwLight(&light, part.offset, part.size)
		}
		if version > 40 {
			if version >= 42 {
				deps.rwLight(&light, 36, 4)
			} else {
				value := deps.rwLegacyField43(0)
				binary.LittleEndian.PutUint32(light[36:40], uint32(value))
			}
			apply = deps.readMode() == 1
		} else if deps.readMode() == 1 {
			binary.LittleEndian.PutUint32(light[36:40], 0)
			apply = true
		}
	} else {
		for _, part := range invLightXferLegacyParts4F5AA0 {
			deps.rwLight(&light, part.offset, part.size)
		}
		mode := deps.readMode()
		clear(light[40:42])
		clear(light[122:124])
		clear(light[124:126])
		clear(light[126:128])
		clear(light[128:132])
		clear(light[134:136])
		light[138] = 0x80
		if mode == 1 {
			apply = true
			if deps.legacyTooBright(&light) {
				deps.clampLegacyLight(&light)
				apply = deps.readMode() == 1
			}
		}
	}

	if apply && deps.gameFlags(invLightXferGameMask4F5AA0) != 0 {
		deps.copyObjectLight(object, &light)
	}

	liveField34 := deps.loadField34(object)
	if liveField34 != 0 && deps.readMode() == 1 {
		if deps.transferInventory(versionWord, object, int32(liveField34)) == 0 {
			return 0
		}
	}
	deps.storeField34(object, originalField34)
	return 1
}

func invLightLegacyTooBright4F5AA0(light *[invLightXferPayloadSize4F5AA0]byte) bool {
	intensity := math.Float32frombits(binary.LittleEndian.Uint32(light[4:8]))
	fixedIntensity := int32(binary.LittleEndian.Uint32(light[12:16]))
	return intensity > 63 || float64(fixedIntensity)*(1.0/65536.0) > 63
}

func invLightClampLegacy4F5AA0(
	light *[invLightXferPayloadSize4F5AA0]byte,
	radius func(float32) uint32,
) {
	binary.LittleEndian.PutUint32(light[4:8], math.Float32bits(63))
	binary.LittleEndian.PutUint32(light[8:12], radius(63))
}
