package legacy

/*
#include "GAME1_1.h"
#include "cgo_const.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

type fieldGuideEntry struct {
	name        *wchar2_t
	description *wchar2_t
	thingType   uint32
	cageImage   *nox_video_bag_image_t
	bookImage   *nox_video_bag_image_t
	unitSize    uint8
}

var fieldGuides [server.FieldGuideCount]fieldGuideEntry

const fieldGuideSource = "C:\\NoxPost\\src\\common\\Magic\\ComGuide.c"

func loadFieldGuides(typeByID func(string) *client.ObjectType, translate func(strman.ID, string) string, loadImage func(string) *noxrender.Image) bool {
	clear(fieldGuides[:])
	for id := 1; id < len(fieldGuides); id++ {
		creature := server.RewardFieldGuideName4F0D20(id)
		typ := typeByID(creature)
		if typ == nil {
			return false
		}
		entry := &fieldGuides[id]
		entry.name = internWStr(translate(strman.ID("creature:"+creature), fieldGuideSource))
		if typ.ID() != "Bomber" {
			entry.thingType = uint32(typ.Index())
		}
		entry.description = internWStr(translate(strman.ID("creature_desc:"+creature), fieldGuideSource))
		entry.cageImage = (*nox_video_bag_image_t)(loadImage("CreatureCage" + creature).C())
		entry.bookImage = (*nox_video_bag_image_t)(loadImage("Spellbook" + creature).C())
		switch {
		case typ.ObjSubClass&object.SubClass(object.MonsterSmall) != 0:
			entry.unitSize = 1
		case typ.ObjSubClass&object.SubClass(object.MonsterMedium) != 0:
			entry.unitSize = 2
		default:
			entry.unitSize = 4
		}
		*memmap.PtrUint8(0x5D4594, uintptr(740100+28*id)) = entry.unitSize
	}
	return true
}

func Nox_xxx_loadGuides_427070() int {
	things := &GetClient().Cli().Things
	if !loadFieldGuides(things.TypeByID, GetServer().S().Strings().GetStringInFile, Nox_xxx_gLoadImg) {
		return 0
	}
	return 1
}

//export nox_xxx_loadGuides_427070
func nox_xxx_loadGuides_427070() C.int {
	return C.int(Nox_xxx_loadGuides_427070())
}

//export nox_xxx_guide_427010
func nox_xxx_guide_427010(name *C.nox_cgo_const_char) C.int {
	return C.int(server.RewardFieldGuideID4F0D20(GoString((*C.char)(unsafe.Pointer(name)))))
}

//export nox_xxx_guideNameByN_427230
func nox_xxx_guideNameByN_427230(id C.int) *C.char {
	if name := server.RewardFieldGuideName4F0D20(int(id)); name != "" {
		return internCStr(name)
	}
	return nil
}

//export nox_xxx_guiCreatureGetName_427240
func nox_xxx_guiCreatureGetName_427240(id C.int) *wchar2_t {
	if id > 0 && int(id) < len(fieldGuides) && fieldGuides[id].thingType != 0 {
		return fieldGuides[id].name
	}
	return nil
}

//export nox_xxx_creatureIsCharmableByTT_4272B0
func nox_xxx_creatureIsCharmableByTT_4272B0(typ C.int) C.int {
	for id := 1; id < len(fieldGuides); id++ {
		if fieldGuides[id].thingType == uint32(typ) {
			return C.int(id)
		}
	}
	return 0
}

//export nox_xxx_guideGetDescById_4272E0
func nox_xxx_guideGetDescById_4272E0(id C.int) *wchar2_t {
	if id > 0 && int(id) < len(fieldGuides) {
		return fieldGuides[id].description
	}
	return nil
}

//export nox_xxx_bookGetFirstCreMB_427300
func nox_xxx_bookGetFirstCreMB_427300() C.int {
	return nox_xxx_bookGetNextCre_427320(0)
}

//export nox_xxx_bookGetNextCre_427320
func nox_xxx_bookGetNextCre_427320(id C.int) C.int {
	for next := max(1, int64(id)+1); next < int64(len(fieldGuides)); next++ {
		if fieldGuides[next].thingType != 0 {
			return C.int(next)
		}
	}
	return 0
}

//export nox_xxx_bookGetCreatureImg_427400
func nox_xxx_bookGetCreatureImg_427400(id C.int) *nox_video_bag_image_t {
	if id > 0 && int(id) < len(fieldGuides) {
		return fieldGuides[id].bookImage
	}
	return nil
}

//export sub_427430
func sub_427430(id C.int) *nox_video_bag_image_t {
	if id > 0 && int(id) < len(fieldGuides) {
		return fieldGuides[id].cageImage
	}
	return nil
}

//export nox_xxx_guideGetUnitSize_427460
func nox_xxx_guideGetUnitSize_427460(id C.int) C.uchar {
	if id > 0 && int(id) < len(fieldGuides) {
		return C.uchar(fieldGuides[id].unitSize)
	}
	return 0
}
