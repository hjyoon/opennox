package legacy

import (
	"runtime"
	"unsafe"
)

// The nonnil diagnostic observer implements the independently declared parent
// boundary. Nil uses the actual raw parent. Ordinary bool damage callbacks and
// their admission rules remain separate from this signed raw result transport.
var rawScalarDamageObserveParent4E24B0 func(site uint32, target, source, weapon unsafe.Pointer, damage, typ int32) int32

func rawScalarDamageCallParent4E24B0(site uint32, target, source, weapon unsafe.Pointer, damage, typ int32) int32 {
	var result int32
	if observe := rawScalarDamageObserveParent4E24B0; observe != nil {
		result = observe(site, target, source, weapon, damage, typ)
	} else {
		result = rawDefaultDamageNative4E0B30(target, source, weapon, damage, typ)
	}
	runtime.KeepAlive(target)
	runtime.KeepAlive(source)
	runtime.KeepAlive(weapon)
	return result
}

func rawStoneDamageNative4E24B0(target, source, weapon unsafe.Pointer, damage, typ int32) int32 {
	return rawScalarDamageCallParent4E24B0(0x4e24c9, target, source, weapon, damage, typ)
}

func rawMechGolemDamageNative4E24E0(target, source, weapon unsafe.Pointer, damage, typ int32) int32 {
	if typ == 9 || typ == 17 {
		damage = int32(uint32(damage) * 2)
	}
	return rawScalarDamageCallParent4E24B0(0x4e250b, target, source, weapon, damage, typ)
}

func rawFlammableDamageNative4E2520(target, source, weapon unsafe.Pointer, damage, typ int32) int32 {
	if typ == 1 || typ == 12 || typ == 7 {
		damage = 9999999
	}
	return rawScalarDamageCallParent4E24B0(0x4e254d, target, source, weapon, damage, typ)
}

func rawBlackPowderDamageNative4E2560(target, source, weapon unsafe.Pointer, damage, typ int32) int32 {
	if typ != 1 && typ != 12 && typ != 2 && typ != 0 {
		return 0
	}
	return rawScalarDamageCallParent4E24B0(0x4e258f, target, source, weapon, 999999, typ)
}
