package opennox

import "unsafe"

// summonAnimateDrawData4B7D00 mirrors nox_animate_draw_data_t. Go's native
// pointer alignment intentionally matches the C parser on both 32- and
// 64-bit targets. It is shared by animation drawing and network lifetime.
type summonAnimateDrawData4B7D00 struct {
	size          uint32
	images        unsafe.Pointer
	frameCount    uint8
	frameDelay    uint8
	animationKind uint32
}

func animateDrawKindNative48EA70(data unsafe.Pointer) uint32 {
	return (*summonAnimateDrawData4B7D00)(data).animationKind
}
