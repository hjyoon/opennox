package legacy

/*
#include "defs.h"
#include "noxstring.h"

static void nox_item_name_format_no_info_4E77E0(wchar2_t* buffer, const wchar2_t* format, const char* unit_name) {
	nox_swprintf(buffer, format, unit_name);
}
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/strman"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

const itemNameSourceFile4E77E0 = "C:\\NoxPost\\src\\Server\\Object\\objutil.c"

type objectItemNameRuntime4E77E0 struct {
	weaponDef     func(uint16) *server.Modifier
	armorDef      func(uint16) *server.Modifier
	unitName      func(*server.Object) *byte
	noInfo        func() *uint16
	noDescription func() *uint16
	buffer        func() *uint16
	clear         func()
	copy          func(*uint16)
	formatNoInfo  func(*uint16, *byte)
	append        func(*uint16)
	appendSpace   func()
}

func itemNameWidePointer4E77E0(s string) *uint16 {
	return (*uint16)(unsafe.Pointer(internWStr(s)))
}

func itemNameNarrowPointer4E77E0(s string) *byte {
	return (*byte)(unsafe.Pointer(internCStr(s)))
}

func itemNameBuffer4E77E0() *wchar2_t {
	return (*wchar2_t)(memmap.PtrOff(0x5D4594, 1565660))
}

var objectItemNameRuntime = objectItemNameRuntime4E77E0{
	weaponDef: func(ind uint16) *server.Modifier {
		return GetServer().S().Modif.Nox_xxx_getProjectileClassById413250(int(ind))
	},
	armorDef: func(ind uint16) *server.Modifier {
		return GetServer().S().Modif.Nox_xxx_equipClothFindDefByTT413270(int(ind))
	},
	unitName: func(obj *server.Object) *byte {
		return itemNameNarrowPointer4E77E0(GetServer().S().Types.ByInd(int(obj.TypeInd)).ID())
	},
	noInfo: func() *uint16 {
		s := GetServer().S().Strings().GetStringInFile(strman.ID("NoInfo"), itemNameSourceFile4E77E0)
		return itemNameWidePointer4E77E0(s)
	},
	noDescription: func() *uint16 {
		s := GetServer().S().Strings().GetStringInFile(strman.ID("NoDescription"), itemNameSourceFile4E77E0)
		return itemNameWidePointer4E77E0(s)
	},
	buffer: func() *uint16 {
		return (*uint16)(unsafe.Pointer(itemNameBuffer4E77E0()))
	},
	clear: func() {
		empty := (*wchar2_t)(memmap.PtrOff(0x5D4594, 1567732))
		copyWideString(itemNameBuffer4E77E0(), empty)
	},
	copy: func(src *uint16) {
		copyWideString(itemNameBuffer4E77E0(), (*wchar2_t)(unsafe.Pointer(src)))
	},
	formatNoInfo: func(format *uint16, name *byte) {
		C.nox_item_name_format_no_info_4E77E0(
			itemNameBuffer4E77E0(),
			(*C.wchar2_t)(unsafe.Pointer(format)),
			(*C.char)(unsafe.Pointer(name)),
		)
	},
	append: func(src *uint16) {
		appendWideString(itemNameBuffer4E77E0(), (*wchar2_t)(unsafe.Pointer(src)))
	},
	appendSpace: func() {
		appendWideString(itemNameBuffer4E77E0(), internWStr(" "))
	},
}

func objectItemNameNative4E77E0(obj *server.Object) *uint16 {
	rt := objectItemNameRuntime
	objectItemName4E77E0(obj, objectItemNameHooks4E77E0[*server.Object, *server.ModifierInitData, *server.Modifier, *server.ModifierEff, *uint16, *byte]{
		class: func(obj *server.Object) uint32 {
			return uint32(obj.ObjClass)
		},
		initData: func(obj *server.Object) *server.ModifierInitData {
			return (*server.ModifierInitData)(obj.InitData)
		},
		typeInd: func(obj *server.Object) uint16 {
			return obj.TypeInd
		},
		weaponDef:     rt.weaponDef,
		armorDef:      rt.armorDef,
		unitName:      rt.unitName,
		noInfo:        rt.noInfo,
		noDescription: rt.noDescription,
		clear:         rt.clear,
		copy:          rt.copy,
		formatNoInfo:  rt.formatNoInfo,
		modifier: func(attrs *server.ModifierInitData, slot int) *server.ModifierEff {
			return attrs.Modifiers[slot]
		},
		modifierDesc: func(mod *server.ModifierEff) *uint16 {
			return mod.Description16()
		},
		modifierIdent: func(mod *server.ModifierEff) *uint16 {
			return mod.IdentificationDescription16()
		},
		definitionDesc: func(def *server.Modifier) *uint16 {
			return def.Description16()
		},
		append:      rt.append,
		appendSpace: rt.appendSpace,
	})
	return rt.buffer()
}

//export nox_xxx_itemGetName_4E77E0_obj_util
func nox_xxx_itemGetName_4E77E0_obj_util(obj *nox_object_t) *wchar2_t {
	return (*wchar2_t)(unsafe.Pointer(objectItemNameNative4E77E0(asObjectS(obj))))
}
