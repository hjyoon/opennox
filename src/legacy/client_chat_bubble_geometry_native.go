package legacy

/*
#include "GAME2_1.h"
#include "GAME2_3.h"

extern int nox_win_width, nox_win_height;
*/
import "C"

import "unsafe"

func chatBubbleTouchesRect(box *C.int4, left, top, right, bottom C.int) bool {
	corner := func(x, y C.int) bool {
		return x >= left && x <= right && y >= top && y <= bottom
	}
	return corner(box.field_0, box.field_4) || corner(box.field_8, box.field_C)
}

//export sub_48E000
func sub_48E000(box *C.int4, drawTail *C.uint32_t) C.bool {
	width, height := C.nox_win_width, C.nox_win_height
	if box.field_0 < 0 || box.field_4 < 0 || box.field_8 > width || box.field_C > height {
		*drawTail = 0
	}
	panelBottom := C.int(55)
	if C.sub_467C80() != 0 {
		panelBottom = 279
	}
	blocked := chatBubbleTouchesRect(box, 0, 0, 563, panelBottom)
	if blocked {
		box.field_4, *drawTail = panelBottom, 0
	}
	if nox_client_getRenderGUI() == 0 {
		return C.bool(blocked)
	}
	for _, panel := range [...][3]C.int{
		{0, height - 127, 111},
		{width/2 - 160, height - 74, width/2 + 160},
		{width - 91, height - 201, width},
	} {
		if chatBubbleTouchesRect(box, panel[0], panel[1], panel[2], height) {
			box.field_4 += panel[1] - box.field_C
			*drawTail = 0
		}
	}
	blocked = sub_4C3260() != 0 && chatBubbleTouchesRect(box, width-87, 0, width, 145)
	if blocked {
		box.field_4, *drawTail = 145, 0
	}
	return C.bool(blocked)
}

var chatBubbleMoves = [69][2]byte{
	17: {0xa0, 0x08},
	18: {0x30, 0x8a},
	20: {0x90, 0x02},
	33: {0xc0, 0x2c},
	34: {0xf0, 0x0f},
	36: {0xc0, 0x13},
	65: {0x60, 0x04},
	66: {0x30, 0x45},
	68: {0x50, 0x01},
}

//export sub_48E240
func sub_48E240(unused C.int, ptr *C.uint32_t) C.char {
	bubble := (*C.nox_chat_bubble)(unsafe.Pointer(ptr))
	if bubble == chatBubbleHead {
		return 0
	}
	var result byte
	for earlier := chatBubbleHead; earlier != nil; earlier = earlier.next {
		result = byte(earlier.visible)
		if earlier.visible == 0 {
			continue
		}
		result = byte(earlier.order)
		if uint32(earlier.order) >= uint32(bubble.order) {
			return C.char(result)
		}
		result = 0
		if !chatBubblesOverlap(bubble, earlier) {
			continue
		}
		zone := sub_48E530(C.int(bubble.x+C.int32_t(uint32(bubble.width)/2)), C.int(bubble.y+C.int32_t(uint32(bubble.height)/2)))
		for _, moves := range chatBubbleMoves[zone] {
			for bit := range 8 {
				result = byte(1 << bit)
				if moves&result == 0 {
					continue
				}
				gap := C.int32_t(2 * nox_xxx_guiFontHeightMB_43F320(nil))
				x, y := chatBubbleCandidate(result, gap, bubble, earlier)
				if chatBubblePositionAvailable(bubble, x, y) {
					bubble.x, bubble.y = x, y
					return 1
				}
				result = 0
			}
		}
		break
	}
	return C.char(result)
}

func chatBubblesOverlap(a, b *C.nox_chat_bubble) bool {
	pad := C.int32_t(nox_xxx_guiFontHeightMB_43F320(nil))
	return uint32(a.x-pad) < uint32(b.x+b.width+pad) &&
		uint32(a.x+a.width+pad) > uint32(b.x-pad) &&
		uint32(a.y-pad) < uint32(b.y+b.height+pad) &&
		a.y+a.height+pad > b.y-pad
}

//export sub_48E480
func sub_48E480(a, b *C.uint32_t) C.int {
	return C.int(bool2int(chatBubblesOverlap((*C.nox_chat_bubble)(unsafe.Pointer(a)), (*C.nox_chat_bubble)(unsafe.Pointer(b)))))
}

//export sub_48E530
func sub_48E530(x, y C.int) C.int {
	column := C.int(16)
	if x >= 2*C.nox_win_width/3 {
		column = 64
	} else if x >= C.nox_win_width/3 {
		column = 32
	}
	if y < C.nox_win_height/3 {
		return column | 1
	}
	if y >= 2*C.nox_win_height/3 {
		return column | 4
	}
	return column | 2
}

func chatBubblePositionAvailable(bubble *C.nox_chat_bubble, x, y C.int32_t) bool {
	nox_xxx_guiFontHeightMB_43F320(nil)
	rect := C.int4{field_0: C.int(x), field_4: C.int(y), field_8: C.int(x + bubble.width), field_C: C.int(y + bubble.height)}
	tail := C.uint32_t(1)
	sub_48E000(&rect, &tail)
	if tail == 0 {
		return false
	}
	if bubble.visible == 0 {
		return true
	}
	for other := chatBubbleHead; other != nil; other = other.next {
		if uint32(other.order) >= uint32(bubble.order) {
			return true
		}
		oldX, oldY := bubble.x, bubble.y
		bubble.x, bubble.y = x, y
		overlaps := chatBubblesOverlap(bubble, other)
		bubble.x, bubble.y = oldX, oldY
		if overlaps {
			return false
		}
	}
	return true
}

//export sub_48E5C0
func sub_48E5C0(ptr *C.uint32_t, x, y C.int) C.int {
	return C.int(bool2int(chatBubblePositionAvailable((*C.nox_chat_bubble)(unsafe.Pointer(ptr)), C.int32_t(x), C.int32_t(y))))
}

func chatBubbleCandidate(direction byte, gap C.int32_t, moving, earlier *C.nox_chat_bubble) (C.int32_t, C.int32_t) {
	left := earlier.x - moving.width - gap
	right := earlier.x + earlier.width + gap
	top := earlier.y - moving.height - gap
	bottom := earlier.y + earlier.height + gap
	switch direction {
	case 0x01:
		return left, top
	case 0x02:
		return right, top
	case 0x04:
		return left, bottom
	case 0x08:
		return right, bottom
	case 0x10:
		return moving.x, top
	case 0x20:
		return moving.x, bottom
	case 0x40:
		return left, moving.y
	case 0x80:
		return right, moving.y
	default:
		return moving.x, moving.y
	}
}

//export nox_client_chat_candidate_word
func nox_client_chat_candidate_word(direction C.char, moving, earlier *C.uint32_t, x, y *C.int) C.uintptr_t {
	gap := C.int32_t(2 * nox_xxx_guiFontHeightMB_43F320(nil))
	dir := byte(direction)
	if dir == 0 || dir&(dir-1) != 0 {
		return C.uintptr_t(gap)
	}
	cx, cy := chatBubbleCandidate(dir, gap, (*C.nox_chat_bubble)(unsafe.Pointer(moving)), (*C.nox_chat_bubble)(unsafe.Pointer(earlier)))
	*x, *y = C.int(cx), C.int(cy)
	switch dir {
	case 1:
		return C.uintptr_t(gap)
	case 0x40:
		return C.uintptr_t(uintptr(unsafe.Pointer(x)))
	default:
		return C.uintptr_t(uintptr(unsafe.Pointer(y)))
	}
}
