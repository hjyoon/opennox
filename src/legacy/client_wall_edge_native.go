package legacy

/*
#include "GAME2_2.h"

extern uint8_t** nox_pixbuffer_rows_3798784;
*/
import "C"

import (
	"encoding/binary"
	"image"
	"unsafe"
)

//export nox_xxx_edgeDraw_480EF0
func nox_xxx_edgeDraw_480EF0(sprite *nox_video_bag_image_t, x, y C.int, startLight, endLight *C.int,
	bottom, left, right, trimBottom, unused C.int,
) *C.int {
	if sprite == nil || asImage(sprite).Type()&0x3f != 3 {
		return (*C.int)(unsafe.Pointer(sprite))
	}
	pixels := asImage(sprite).Pixdata()
	if len(pixels) == 0 {
		return nil
	}
	result := (*C.int)(unsafe.Pointer(&pixels[0]))
	if len(pixels) < 17 {
		return result
	}
	sourceWidth := int32(binary.LittleEndian.Uint32(pixels[:4]))
	sourceHeight := int32(binary.LittleEndian.Uint32(pixels[4:8])) - int32(trimBottom)
	xPos := int32(binary.LittleEndian.Uint32(pixels[8:12])) + int32(x)
	yPos := int32(binary.LittleEndian.Uint32(pixels[12:16])) + int32(y)
	clipLeft, clipRight := int32(legacyWords.dword_5d4594_3807140), int32(legacyWords.dword_5d4594_3807116)
	clipTop, clipBottom := int32(legacyWords.dword_5d4594_3807136), int32(legacyWords.dword_5d4594_3807152)

	drawX, drawY := max(xPos, clipLeft), max(yPos, clipTop)
	skipX, skipY := drawX-xPos, drawY-yPos
	width := min(sourceWidth-skipX, clipRight-drawX)
	height := min(sourceHeight-skipY, clipBottom-drawY)
	if width <= 0 || height <= 0 {
		return result
	}
	height = min(height, max(0, int32(bottom)-drawY))
	moveX := max(0, int32(left)-drawX-skipX)
	skipX += moveX
	width = min(width-moveX, int32(right)-drawX-skipX)
	drawX += moveX

	pitch := int(nox_getBackbufferPitch()) / 2
	frameHeight := int(nox_getBackbufHeight())
	if width <= 0 || height <= 0 || drawX < 0 || int(drawX+width) > pitch || drawY < 0 ||
		int(drawY+height) > frameHeight || C.nox_pixbuffer_rows_3798784 == nil {
		return result
	}

	rows := unsafe.Slice(C.nox_pixbuffer_rows_3798784, frameHeight)
	if rows[drawY] == nil {
		return result
	}
	rowStart := unsafe.Add(unsafe.Pointer(rows[drawY]), 2*int(drawX))
	buffer := unsafe.Slice((*uint16)(rowStart), pitch*(frameHeight-int(drawY))-int(drawX))
	var base, step [3]uint32
	light0, light1 := unsafe.Slice(startLight, 3), unsafe.Slice(endLight, 3)
	for i := range base {
		value := int32(light0[i]) << 8
		base[i] = uint32(value)
		step[i] = uint32(((int32(light1[i]) << 8) - value) / sourceWidth)
	}
	parity := Sub_473970(image.Pt(0, int(drawY))).Y
	highRes := Get_nox_client_highResFrontWalls_80820() != 0
	pixels = pixels[17:]
	var previous []uint16
	repeat := 0
	for row := -skipY; row < height; row++ {
		var destination []uint16
		draw := false
		if row >= 0 {
			destination = buffer[int(row)*pitch:]
			draw = highRes || (parity+int(row))&1 == 0
			if !draw && repeat != 0 {
				copied := min(repeat, len(destination))
				copy(destination[:copied], previous[:copied])
			}
		}
		repeat = 0
		for column := int32(0); column < sourceWidth; {
			if len(pixels) < 2 {
				return result
			}
			op, size := pixels[0], int32(pixels[1])
			pixels = pixels[2:]
			end := column + size
			if size == 0 || end > sourceWidth {
				return result
			}
			if op == 2 && int(size)*2 > len(pixels) {
				return result
			}
			if draw && column < skipX+width && end > skipX {
				from := max(column, skipX)
				to := min(end, skipX+width)
				if op == 2 {
					count := int(to - from)
					source := (*uint16)(unsafe.Pointer(&pixels[2*(from-column)]))
					light := base
					for i := range light {
						light[i] += uint32(from) * step[i]
					}
					start := int(from - skipX)
					Sub_480860(destination[start:start+count], unsafe.Slice(source, count), count, light[:], step[:])
					repeat += count
				} else if column >= skipX {
					// A transparent run starting in the clip contributes its whole
					// length to the row duplicated in interlaced mode.
					repeat += int(size)
				}
			}
			if op == 2 {
				pixels = pixels[2*size:]
			}
			column = end
		}
		previous = destination
	}
	return result
}
