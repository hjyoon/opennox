package legacy

/*
#include "client__gui__tooltip.h"
#include "GAME1_1.h"
#include "GAME5_2.h"
#include "noxstring.h"

static void tooltip_format_4BF050(wchar2_t* dst, const wchar2_t* format, const void* arg) {
	nox_swprintf(dst, format, arg);
}
*/
import "C"
import (
	"unsafe"

	"github.com/opennox/libs/strman"
	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

// The factory is a test seam for exercising the real C entry without replacing
// its native pointer ABI. Dependencies are resolved only on the selected branch.
var clientAskInfoDepsFactory4BF050 = func() clientAskInfoDeps4BF050 {
	return clientAskInfoDeps4BF050{
		scratch:      (*uint16)(memmap.PtrOff(0x5D4594, 1317000)),
		initial:      (*uint16)(memmap.PtrOff(0x5D4594, 1319048)),
		space:        alloc.InternCString16(" "),
		prefixFormat: alloc.InternCString16("%s "),
		copyText: func(dst, src *uint16) {
			C.nox_wcscpy((*wchar2_t)(unsafe.Pointer(dst)), (*wchar2_t)(unsafe.Pointer(src)))
		},
		appendText: func(dst, src *uint16) {
			C.nox_wcscat((*wchar2_t)(unsafe.Pointer(dst)), (*wchar2_t)(unsafe.Pointer(src)))
		},
		formatText: func(dst, format *uint16, arg unsafe.Pointer) {
			C.tooltip_format_4BF050((*wchar2_t)(unsafe.Pointer(dst)), (*wchar2_t)(unsafe.Pointer(format)), arg)
		},
		prettyName: func(id uint32) *uint16 {
			return (*uint16)(unsafe.Pointer(nox_get_thing_pretty_name(int32(id))))
		},
		typeName: func(id uint32) *byte {
			return (*byte)(unsafe.Pointer(nox_get_thing_name(int32(id))))
		},
		weaponDef: func(id uint32) *server.Modifier {
			return GetServer().S().Modif.Nox_xxx_getProjectileClassById413250(int(int32(id)))
		},
		armorDef: func(id uint32) *server.Modifier {
			return GetServer().S().Modif.Nox_xxx_equipClothFindDefByTT413270(int(int32(id)))
		},
		language: func() int { return int(GetServer().S().Strings().Lang()) },
		loadString: func(id string, _ int) *uint16 {
			return alloc.InternCString16(GetServer().S().Strings().GetStringInFile(strman.ID(id), `C:\NoxPost\src\client\Gui\ToolTip.c`))
		},
		spellTitle: func(id uint32) *uint16 {
			return (*uint16)(unsafe.Pointer(nox_xxx_spellTitle_424930(int32(id))))
		},
		creatureName: func(id uint32) *uint16 {
			return (*uint16)(unsafe.Pointer(C.nox_xxx_guiCreatureGetName_427240(C.int(int32(id)))))
		},
		abilityName: func(id uint32) *uint16 {
			return (*uint16)(unsafe.Pointer(nox_xxx_abilityGetName_0_425260(int32(id))))
		},
		unitCode: func(dr *client.Drawable) uint16 {
			return uint16(C.nox_xxx_netGetUnitCodeCli_578B00((*nox_drawable)(dr.C())))
		},
		send: Nox_xxx_netClientSend2_4E53C0,
	}
}

//export nox_xxx_clientAskInfoMb_native_4BF050
func nox_xxx_clientAskInfoMb_native_4BF050(dr *nox_drawable) *wchar2_t {
	return (*wchar2_t)(unsafe.Pointer(clientAskInfoNative4BF050(asDrawable(dr), clientAskInfoDepsFactory4BF050())))
}
