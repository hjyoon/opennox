package legacy

/*
#include "defs.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

// Keep process-lifetime pointer slots in native memory, including slots passed
// by address and slots that carry legacy values outside the Go heap.
var legacyPointers, _ = alloc.New(struct {
	audioGlobal          unsafe.Pointer
	dialogTimer          unsafe.Pointer
	ordinaryMusicTimer   unsafe.Pointer
	audioPlaybackContext unsafe.Pointer

	inventoryDragItem            *nox_drawable
	inventoryTooltipItem         *nox_drawable
	inventoryPreviousTooltipItem *nox_drawable
	shopSellInstructions         *C.wchar2_t
	shopRepairInstructions       *C.wchar2_t
	shopBuyInstructions          *C.wchar2_t
	shopDialogFile               *C.char
	entryFieldState              unsafe.Pointer
	entryFieldBuffer             unsafe.Pointer

	pixelSpanAllocator      unsafe.Pointer
	pixelSpanBuffer         unsafe.Pointer
	videoRows               **C.uchar
	highlightedSprite       unsafe.Pointer
	spriteUnderCursor       unsafe.Pointer
	flagTeamBorder          *C.nox_video_bag_image_t
	screenParticleAllocator unsafe.Pointer
	screenParticleHead      *C.nox_screenParticle
	screenParticleTail      *C.nox_screenParticle

	wolWindowAnimation  *C.nox_gui_animation
	wolChannelHead      *wolChannel
	wolGameHead         *wolGameListNode
	wolServiceHead      *wolServiceEntry
	friendListAllocator unsafe.Pointer
	friendListHead      *clientFriendNode
	wolServerList       C.nox_list_item_t
	wolSelectedServer   unsafe.Pointer

	collisionHitAllocator unsafe.Pointer
	cursorCollisionSource *C.nox_object_t
	mapgenLastWaypoint    *nox_waypoint_t
	polygonAngleRemapHead unsafe.Pointer
	playerProtectionHead  unsafe.Pointer
	playerProtectionTail  unsafe.Pointer
	respawnHead           *respawnRecord4EC5E0
	respawnAllocator      unsafe.Pointer
	magicEntityAllocator  unsafe.Pointer
	chatBubbleAllocator   unsafe.Pointer
	lastFlagStart         unsafe.Pointer
}{})
