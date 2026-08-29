package legacy

import "math"

const (
	triggerXferCurrentVersion4F4E50 = uint16(61)
	triggerXferColorsVersion4F4E50  = int16(41)
	triggerXferScriptsVersion4F4E50 = int16(3)
	triggerXferCollideVersion4F4E50 = int16(31)
	triggerXferTeamsVersion4F4E50   = int16(21)
	triggerXferStateVersion4F4E50   = int16(61)
)

type triggerXferCallback4F4E50 uint8

const (
	triggerXferActivate4F4E50 triggerXferCallback4F4E50 = iota
	triggerXferDeactivate4F4E50
	triggerXferCollide4F4E50
)

// triggerXferDeps4F4E50 exposes every observable object/update-data access,
// transfer, global-mode read, and external call in GAME.EXE 004F4E50. Object,
// update-data, and script-data identities remain generic so the contract
// cannot inherit PE32 pointer width.
type triggerXferDeps4F4E50[O comparable, U any, S any] struct {
	loadField34        func(O) uint32
	loadUpdateData     func(O) U
	rwVersion          func(uint16) uint16
	mapReadWrite       func(O, int32) int32
	readOnly           func() int32
	loadBoxWidth       func(O) float32
	loadBoxHeight      func(O) float32
	truncFloatDword    func(float32) int32
	rwBoxWidth         func(int32) int32
	rwBoxHeight        func(int32) int32
	storeBoxWidth      func(O, float32)
	storeBoxHeight     func(O, float32)
	calcBox            func(O)
	rwLegacyScratch3   func(uint32) uint32
	rwColor            func(U, int)
	rwFlags            func(U)
	loadScriptData     func(O) S
	transferScript     func(U, triggerXferCallback4F4E50, S, uintptr)
	initLegacyScript   func(U, triggerXferCallback4F4E50)
	rwLegacyCount      func(uint8) uint8
	seekCurrent        func(int32)
	rwClassInclude     func(U)
	rwClassExclude     func(U)
	storeTeamInclude   func(U, uint8)
	storeTeamExclude   func(U, uint8)
	rwTeamInclude      func(U)
	rwTeamExclude      func(U)
	rwState            func(U)
	rwField9           func(U)
	rwField33          func(O)
	loadField33        func(O) uint32
	markAnimationFrame func(O, uint32)
	transferInventory  func(uint16, O, int32) int32
	storeField34       func(O, uint32)
}

// triggerXferTruncFloatDword4F4E50 models the low dword returned by the x87
// signed-qword conversion helper at 00566DCC. Invalid conversions produce the
// integer-indefinite qword, whose low dword is zero.
func triggerXferTruncFloatDword4F4E50(value float32) int32 {
	v := float64(value)
	if math.IsNaN(v) || v >= 0x1p63 || v < -0x1p63 {
		return 0
	}
	return int32(int64(math.Trunc(v)))
}

// triggerXfer4F4E50 preserves the original entry-time Field34 and UpdateData
// caches, signed version branches, truthy shape-read mode, exact-one legacy
// skips/team/animation/inventory gates, shared scratch reuse, callback order,
// live Field33/Field34 reads, and failure prefixes. There are deliberately no
// object, update-data, shape, or class guards. The native adapter separately
// preserves the original null script-data to null callback-context mapping.
func triggerXfer4F4E50[O comparable, U any, S any](
	object O,
	deps triggerXferDeps4F4E50[O, U, S],
) int32 {
	originalField34 := deps.loadField34(object)
	update := deps.loadUpdateData(object)

	versionWord := deps.rwVersion(triggerXferCurrentVersion4F4E50)
	version := int16(versionWord)
	if version > int16(triggerXferCurrentVersion4F4E50) {
		return 0
	}
	if deps.mapReadWrite(object, int32(version)) == 0 {
		return 0
	}

	var scratch uint32
	var heightScratch uint32
	if deps.readOnly() == 0 {
		scratch = uint32(deps.truncFloatDword(deps.loadBoxWidth(object)))
		heightScratch = uint32(deps.truncFloatDword(deps.loadBoxHeight(object)))
		scratch = uint32(deps.rwBoxWidth(int32(scratch)))
		heightScratch = uint32(deps.rwBoxHeight(int32(heightScratch)))
	} else {
		scratch = uint32(deps.rwBoxWidth(int32(scratch)))
		heightScratch = uint32(deps.rwBoxHeight(int32(heightScratch)))

		width := float32(int32(scratch))
		scratch = math.Float32bits(width)
		deps.storeBoxWidth(object, width)
		height := float32(int32(heightScratch))
		deps.storeBoxHeight(object, height)
		if width > 60.0 {
			deps.storeBoxWidth(object, 60.0)
		}
		if height > 60.0 {
			deps.storeBoxHeight(object, 60.0)
		}
	}
	deps.calcBox(object)

	if version < triggerXferColorsVersion4F4E50 {
		for range 3 {
			scratch = deps.rwLegacyScratch3(scratch)
		}
	} else {
		for index := range 6 {
			deps.rwColor(update, index)
		}
	}
	deps.rwFlags(update)

	if version < triggerXferScriptsVersion4F4E50 {
		deps.initLegacyScript(update, triggerXferActivate4F4E50)
		deps.initLegacyScript(update, triggerXferDeactivate4F4E50)
	} else {
		scriptData := deps.loadScriptData(object)
		deps.transferScript(update, triggerXferActivate4F4E50, scriptData, 256)
		deps.transferScript(update, triggerXferDeactivate4F4E50, scriptData, 384)
		if version >= triggerXferCollideVersion4F4E50 {
			deps.transferScript(update, triggerXferCollide4F4E50, scriptData, 512)
		}
	}

	if deps.readOnly() == 1 && version < triggerXferCollideVersion4F4E50 {
		for range 4 {
			count := deps.rwLegacyCount(uint8(scratch))
			scratch = scratch&0xffffff00 | uint32(count)
			deps.seekCurrent(int32(uint32(count) * 4))
		}
	}

	deps.rwClassInclude(update)
	deps.rwClassExclude(update)
	if deps.readOnly() == 1 {
		deps.storeTeamInclude(update, 0)
		deps.storeTeamExclude(update, 0)
	}
	if deps.readOnly() == 0 || version >= triggerXferTeamsVersion4F4E50 {
		deps.rwTeamInclude(update)
		deps.rwTeamExclude(update)
	}

	if version >= triggerXferStateVersion4F4E50 {
		deps.rwState(update)
		deps.rwField9(update)
		deps.rwField33(object)
		if deps.readOnly() == 1 {
			deps.markAnimationFrame(object, deps.loadField33(object))
		}
	}

	inventoryCount := deps.loadField34(object)
	if inventoryCount != 0 && deps.readOnly() == 1 {
		if deps.transferInventory(versionWord, object, int32(inventoryCount)) == 0 {
			return 0
		}
	}
	deps.storeField34(object, originalField34)
	return 1
}
