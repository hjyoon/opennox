package legacy

/*
#include "quest_preview_44e110.h"
#include "GAME2.h"
#include "client__drawable__drawable.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/client"
)

var questPreviewDrawables [12]*client.Drawable

func questPreviewFontSlot44E110() *unsafe.Pointer { return &briefingState.preview_font }

// The constructor, renderer, and cleanup use the same drawable slots.
func questPreviewSlots44E110() [12]**client.Drawable {
	var slots [12]**client.Drawable
	for i := range slots {
		slots[i] = &questPreviewDrawables[i]
	}
	return slots
}

func questPreviewNativeHooks44E110() questPreviewHooks44E110[*client.Drawable, unsafe.Pointer] {
	slots := questPreviewSlots44E110()
	return questPreviewHooks44E110[*client.Drawable, unsafe.Pointer]{
		loadFont:   func() unsafe.Pointer { return *questPreviewFontSlot44E110() },
		fontByName: func(name string) unsafe.Pointer { return GetClient().R2().GetFonts().FontPtrByName(name) },
		storeFont:  func(font unsafe.Pointer) { *questPreviewFontSlot44E110() = font },
		load:       func(slot int) *client.Drawable { return *slots[slot] },
		thingByName: func(name string) int32 {
			return int32(C.nox_xxx_getTTByNameSpriteMB_44CFC0(internCStr(name)))
		},
		create: func(typ int32) *client.Drawable {
			return (*client.Drawable)(unsafe.Pointer(C.nox_new_drawable_for_thing(C.int(typ))))
		},
		store: func(slot int, dr *client.Drawable) { *slots[slot] = dr },
		mark:  func(dr *client.Drawable) { dr.ObjFlags |= object.Flags(0x1000000) },
	}
}

var questPreviewCall44E110 = func() *client.Drawable {
	return questPreview44E110(questPreviewNativeHooks44E110())
}

//export nox_client_questPreview_native_44E110
func nox_client_questPreview_native_44E110() *C.uint32_t {
	return (*C.uint32_t)(unsafe.Pointer(questPreviewCall44E110()))
}

func questPreviewCEntry44E110() *client.Drawable {
	return (*client.Drawable)(unsafe.Pointer(C.sub_44E110()))
}
