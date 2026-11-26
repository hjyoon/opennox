package legacy

/*
#include "GAME4_2.h"
#include "client__gui__guigen.h"

extern int nox_win_width;
extern int nox_win_height;

// The packed compass table stores 32-bit image keys. The bag resolves them
// against its live image handles without dereferencing the key itself.
static inline nox_video_bag_image_t* nox_mapgenCompassImageKey(uint32_t value) {
	return (nox_video_bag_image_t*)(uintptr_t)value;
}
*/
import "C"

import (
	"encoding/binary"
	"unsafe"

	"github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
)

type mapgenProgressServices struct {
	ticks     func() uint64
	poll      func() int32
	threshold func() int32
	sequence  func() *uint32
	display   func([3]byte)
}

func mapgenUpdateProgress(code byte, current, previous *uint32, services mapgenProgressServices) {
	*current = uint32(services.ticks())
	if int32(*previous) > int32(*current) {
		*previous = 0
	}
	services.poll()
	if int32(*current-*previous) <= services.threshold() {
		return
	}
	sequence := services.sequence()
	packet := [3]byte{code}
	binary.LittleEndian.PutUint16(packet[1:], uint16(*sequence))
	*sequence += 1
	services.display(packet)
	*previous = *current
}

//export nox_xxx_mapGenSetFlags_5235F0
func nox_xxx_mapGenSetFlags_5235F0(code C.char) {
	if noxflags.HasEngine(noxflags.EngineNoRendering) || noxflags.HasGame(noxflags.GameFlag22) {
		return
	}
	mapgenUpdateProgress(byte(code), (*uint32)(unsafe.Pointer(&legacyAddressableWords.dword_5d4594_2487564)),
		(*uint32)(unsafe.Pointer(&legacyAddressableWords.dword_5d4594_2487568)), mapgenProgressServices{
			ticks:     PlatformTicks,
			poll:      nox_input_pollEvents_4453A0,
			threshold: func() int32 { return memmap.Int32(0x587000, 254948) },
			sequence:  func() *uint32 { return memmap.PtrUint32(0x5D4594, 2487572) },
			display:   func(packet [3]byte) { nox_xxx_mapGenClientText_4A9D00((*C.uchar)(unsafe.Pointer(&packet[0]))) },
		})
}

var mapgenProgressLabels = [...]string{"Generating", "Assembling", "Populating"}

//export nox_xxx_mapGenClientText_4A9D00
func nox_xxx_mapGenClientText_4A9D00(raw *C.uchar) C.short {
	packet := unsafe.Slice((*byte)(unsafe.Pointer(raw)), 3)
	sequence := binary.LittleEndian.Uint16(packet[1:])
	last := memmap.PtrUint32(0x5D4594, 1309668)
	if uint32(sequence) == *last {
		return C.short(sequence)
	}
	*last = uint32(sequence)

	nox_xxx_clientPlaySoundSpecial_452D80(897, 50)
	Sub_430B50(0, 0, int(C.nox_win_width)-1, int(C.nox_win_height)-1)
	nox_client_clearScreen_440900()
	x := int32(C.nox_win_width)/2 - 160
	y := int32(C.nox_win_height)/2 - 120
	frame := memmap.PtrUint32(0x5D4594, 1309672)
	background := memmap.Uint32(0x5D4594, 1309644+4*uintptr(*frame%4))
	nox_client_drawImageAt_47D2C0(C.nox_mapgenCompassImageKey(C.uint32_t(background)), x, y)
	arrow := memmap.Uint32(0x5D4594, uintptr(1309516+4*(*frame)))
	nox_client_drawImageAt_47D2C0(C.nox_mapgenCompassImageKey(C.uint32_t(arrow)), x, y)

	label := memmap.PtrPtr(0x5D4594, 1309660)
	if index := int(packet[0]) - 0x9B; index >= 0 && index < len(mapgenProgressLabels) {
		source := internCStr(`C:\NoxPost\src\client\Gui\guigen.c`)
		text := nox_strman_loadString_40F1D0(internCStr(mapgenProgressLabels[index]), nil, source, int32(89+3*index))
		*label = unsafe.Pointer(text)
	}
	var textWidth C.int
	nox_xxx_drawGetStringSize_43F840(nil, (*wchar2_t)(*label), &textWidth, nil, 0)
	centerX := int32(C.nox_win_width-textWidth) / 2
	centerY := int32(C.nox_win_height)/2 - (2*nox_xxx_guiFontHeightMB_43F320(nil) + 70)
	nox_xxx_drawSetTextColor_434390(int32(Get_nox_color_white_2523948()))
	nox_xxx_drawSetColor_4343E0(memmap.Int32(0x852978, 4))
	nox_draw_drawStringHL_43F730(nil, (*wchar2_t)(*label), centerX, centerY)
	*frame += 1
	if *frame >= 32 {
		*frame = 0
	}
	nox_video_callCopyBackBuffer_4AD170()
	return C.short(sequence)
}
