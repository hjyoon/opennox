package legacy

/*
#include <stdint.h>
#include "regeneration_raw_4e01d0.h"
static void* regeneration_raw_product_entry_4e01d0(void) {
 return (void*)&nox_xxx_effectRegeneration_4E01D0;
}
*/
import "C"

import (
	"reflect"
	"runtime"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

// Capture the ordinary service identities during package initialization,
// before a test can replace a variable. The diagnostic never assigns them.
var regenerationRawProductDefaultServices4E01D0 = [4]uintptr{
	reflect.ValueOf(regenerationRawFrame4E01D0).Pointer(),
	reflect.ValueOf(regenerationRawFPS4E01D0).Pointer(),
	reflect.ValueOf(regenerationRawArmorService4E01D0).Pointer(),
	reflect.ValueOf(regenerationRawAdjustService4E01D0).Pointer(),
}

type RegenerationRawProductFacts4E01D0 struct {
	Frame, FPS, ArmorFlags uint32
	RawEntry               unsafe.Pointer
	RegisteredUpdate       unsafe.Pointer
}

// RegenerationRawActualServices4E01D0 reports live production facts. It
// refuses raw-root and HP-getter observers, substituted world/armor/HP services, an old integer
// public ABI or a missing native engine. It changes no world state.
func RegenerationRawActualServices4E01D0(item *server.Object) (RegenerationRawProductFacts4E01D0, bool) {
	if regenerationRawABI4E01D0() != 2 || regenerationRawEngine4E01D0 == nil ||
		regenerationRawObserveRead4E01D0 != nil || regenerationRawObserveCall4E01D0 != nil || regenerationRawObserveArithmetic4E01D0 != nil ||
		server.ObserveUnitHPRead4EE780 != nil {
		return RegenerationRawProductFacts4E01D0{}, false
	}
	current := [4]uintptr{
		reflect.ValueOf(regenerationRawFrame4E01D0).Pointer(),
		reflect.ValueOf(regenerationRawFPS4E01D0).Pointer(),
		reflect.ValueOf(regenerationRawArmorService4E01D0).Pointer(),
		reflect.ValueOf(regenerationRawAdjustService4E01D0).Pointer(),
	}
	if current != regenerationRawProductDefaultServices4E01D0 ||
		current[3] != reflect.ValueOf(unitAdjustHPCall4EE460).Pointer() ||
		reflect.ValueOf(regenerationRawEngine4E01D0).Pointer() != reflect.ValueOf(regenerationRawNative4E01D0).Pointer() {
		return RegenerationRawProductFacts4E01D0{}, false
	}
	srv := GetServer().S()
	var armor uint32
	if item != nil && uint32(item.ObjClass)&0x02000000 != 0 {
		armor = srv.Armor.Nox_xxx_unitArmorInventoryEquipFlags_415C70(item)
	}
	return RegenerationRawProductFacts4E01D0{
		Frame: srv.Frame(), FPS: srv.TickRate(), ArmorFlags: armor,
		RawEntry:         C.regeneration_raw_product_entry_4e01d0(),
		RegisteredUpdate: regenerationEffectPointerNative4E01D0(),
	}, true
}

// RegenerationRawActualDefaultC4E01D0 directly transports the two native
// pointers through the real public C entry. The root is void: callers check
// real health effects themselves. No callback or expected result is supplied.
func RegenerationRawActualDefaultC4E01D0(effect *server.ModifierEff, item *server.Object) bool {
	if _, ok := RegenerationRawActualServices4E01D0(item); !ok {
		return false
	}
	regenerationRawPublic4E01D0(unsafe.Pointer(effect), unsafe.Pointer(item))
	runtime.KeepAlive(effect)
	runtime.KeepAlive(item)
	return true
}
