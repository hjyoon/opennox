package legacy

/*
#include "GAME2_3.h"
#include "common/alloc/classes/alloc_class.h"

*/
import "C"

import (
	"encoding/binary"
	"unsafe"
)

var chatBubbleHead, chatBubbleTail *C.nox_chat_bubble

func chatBubblePool() *C.nox_alloc_class {
	return (*C.nox_alloc_class)(Get_nox_alloc_chat_1197364())
}

func resetChatBubbleList() {
	chatBubbleHead, chatBubbleTail = nil, nil
	*g2ServerWord(1197368) = 0
}

//export sub_48D800
func sub_48D800() C.int {
	if pool := chatBubblePool(); pool != nil {
		nox_free_alloc_class(pool)
	}
	Set_nox_alloc_chat_1197364(nil)
	resetChatBubbleList()
	return 1
}

//export nox_xxx_netCode2ChatBubble_48D850
func nox_xxx_netCode2ChatBubble_48D850(code C.int) *C.nox_chat_bubble {
	for bubble := chatBubbleHead; bubble != nil; bubble = bubble.next {
		if bubble.net_code == C.uint32_t(code) {
			return bubble
		}
	}
	return nil
}

const chatBubblePacketHeaderSize = 11

//export nox_xxx_createTextBubble_48D880
func nox_xxx_createTextBubble_48D880(packet unsafe.Pointer, text *wchar2_t) {
	if packet != nil {
		createChatBubble(unsafe.Slice((*byte)(packet), chatBubblePacketHeaderSize), text)
	}
}

func createChatBubble(packet []byte, text *wchar2_t) {
	code := C.int(binary.LittleEndian.Uint16(packet[1:3]))
	bubble := nox_xxx_netCode2ChatBubble_48D850(code)
	isNew := bubble == nil
	if isNew {
		bubble = (*C.nox_chat_bubble)(nox_alloc_class_new_obj_zero(chatBubblePool()))
		if bubble == nil {
			return
		}
	}
	copyChatBubbleText(&bubble.text, text)
	bubble.duration_hint = C.uint32_t(packet[8])
	bubble.drawable_x = C.uint16_t(binary.LittleEndian.Uint16(packet[4:6]))
	bubble.drawable_y = C.uint16_t(binary.LittleEndian.Uint16(packet[6:8]))
	bubble.net_code = C.uint32_t(code)
	duration := uint32(binary.LittleEndian.Uint16(packet[9:11]))
	if duration == 0 {
		duration = gameFPS() * (min(uint32(bubble.duration_hint)/8, 8) + 2)
	}
	bubble.expire_frame = C.uint32_t(gameFrame() + duration)

	if isNew {
		tail := chatBubbleTail
		bubble.prev, bubble.next = tail, nil
		if tail == nil {
			chatBubbleHead = bubble
		} else {
			tail.next = bubble
		}
		chatBubbleTail = bubble
	}
}

func copyChatBubbleText(dst *[318]wchar2_t, text *wchar2_t) {
	if text == nil {
		dst[0] = 0
		return
	}
	for i := range len(dst) - 1 {
		value := *(*wchar2_t)(unsafe.Add(unsafe.Pointer(text), i*2))
		dst[i] = value
		if value == 0 {
			return
		}
	}
	dst[len(dst)-1] = 0
}

func nox_chat_bubble_unlink_48DCF0(bubble *C.nox_chat_bubble) {
	if bubble.prev != nil {
		bubble.prev.next = bubble.next
	} else {
		chatBubbleHead = bubble.next
	}
	if bubble.next != nil {
		bubble.next.prev = bubble.prev
	} else {
		chatBubbleTail = bubble.prev
	}
}

//export sub_48E8E0
func sub_48E8E0(code C.int) {
	bubble := nox_xxx_netCode2ChatBubble_48D850(code)
	if bubble != nil {
		nox_chat_bubble_unlink_48DCF0(bubble)
		nox_alloc_class_free_obj_first(chatBubblePool(), unsafe.Pointer(bubble))
	}
}

//export sub_48E940
func sub_48E940() {
	nox_alloc_class_free_all(chatBubblePool())
	resetChatBubbleList()
}
