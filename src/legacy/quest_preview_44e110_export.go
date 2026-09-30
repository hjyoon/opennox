package legacy

/*
#include "quest_preview_44e110.h"
#include "GAME2.h"
#include "client__drawable__drawable.h"
extern uintptr_t dword_5d4594_832484;
extern uintptr_t dword_5d4594_832492;
extern uintptr_t dword_5d4594_832496;
extern uintptr_t dword_5d4594_832500;
extern uintptr_t dword_5d4594_832504;
extern uintptr_t dword_5d4594_832508;
extern uintptr_t dword_5d4594_832512;
extern uintptr_t dword_5d4594_832516;
extern uintptr_t dword_5d4594_832520;
extern uintptr_t dword_5d4594_832524;
extern uintptr_t dword_5d4594_832528;
extern uintptr_t dword_5d4594_832532;
extern uintptr_t dword_5d4594_832536;
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/client"
)

func questPreviewFontSlot44E110() *unsafe.Pointer {
	return (*unsafe.Pointer)(unsafe.Pointer(&C.dword_5d4594_832484))
}

// Use the actual native-width C storage, shared with the existing cleanup.
// memmap's four-byte original address descriptors are not these C variables.
func questPreviewSlots44E110() [12]**client.Drawable {
	return [12]**client.Drawable{
		(**client.Drawable)(unsafe.Pointer(&C.dword_5d4594_832492)),
		(**client.Drawable)(unsafe.Pointer(&C.dword_5d4594_832496)),
		(**client.Drawable)(unsafe.Pointer(&C.dword_5d4594_832500)),
		(**client.Drawable)(unsafe.Pointer(&C.dword_5d4594_832504)),
		(**client.Drawable)(unsafe.Pointer(&C.dword_5d4594_832508)),
		(**client.Drawable)(unsafe.Pointer(&C.dword_5d4594_832512)),
		(**client.Drawable)(unsafe.Pointer(&C.dword_5d4594_832516)),
		(**client.Drawable)(unsafe.Pointer(&C.dword_5d4594_832520)),
		(**client.Drawable)(unsafe.Pointer(&C.dword_5d4594_832524)),
		(**client.Drawable)(unsafe.Pointer(&C.dword_5d4594_832528)),
		(**client.Drawable)(unsafe.Pointer(&C.dword_5d4594_832532)),
		(**client.Drawable)(unsafe.Pointer(&C.dword_5d4594_832536)),
	}
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
