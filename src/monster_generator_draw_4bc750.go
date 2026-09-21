//go:build !server

package opennox

import (
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/legacy"
)

// monsterGeneratorDrawData4BC750 mirrors nox_cond_animate_draw_data_t. The
// image table uses native pointers, so Go's normal alignment is intentional.
type monsterGeneratorDrawData4BC750 struct {
	size          uint32
	images        [5]*noxrender.ImageHandle
	frameCount    [5]uint8
	frameDelay    [5]uint8
	animationKind [5]uint32
}

type monsterGeneratorDrawFrames4BC750 struct {
	mainState    int
	mainFrame    int
	overlayFrame int
	hasOverlay   bool
	terminal     bool
}

func monsterGeneratorState4BC750(flags uint32) int {
	switch {
	case flags&0x100 != 0:
		return 1
	case flags&0x200 != 0:
		return 2
	case flags&0xC00 != 0:
		return 3
	default:
		return 0
	}
}

// monsterGeneratorDrawFramesFor4BC750 reproduces the original frame and
// state transitions without interpreting Drawable through PE32 byte offsets.
func monsterGeneratorDrawFramesFor4BC750(dr *client.Drawable, data *monsterGeneratorDrawData4BC750, gameFrame uint32, random func(int, int) int) (monsterGeneratorDrawFrames4BC750, bool) {
	if dr == nil || data == nil {
		return monsterGeneratorDrawFrames4BC750{}, false
	}
	state := monsterGeneratorState4BC750(dr.Flags70Val)
	count := int(data.frameCount[state])
	if count == 0 || data.images[state] == nil {
		return monsterGeneratorDrawFrames4BC750{}, false
	}

	frame := 0
	switch data.animationKind[state] {
	case 2:
		frame = int((gameFrame + dr.NetCode32) / (uint32(data.frameDelay[state]) + 1) % uint32(count))
	case 4:
		if random == nil {
			return monsterGeneratorDrawFrames4BC750{}, false
		}
		// GAME.EXE passed frameCount as an inclusive upper bound here, unlike
		// ConditionalAnimateDraw. Restrict it to the last valid element so a
		// random animation can never walk one image beyond the native table.
		frame = random(0, count-1)
	case 5:
		frame = int(dr.AnimFrameSlave)
	default:
		// Kind 0 in GAME.EXE accidentally used the address of frameDelay as
		// an image index. It cannot produce a valid portable frame.
		return monsterGeneratorDrawFrames4BC750{}, false
	}
	if frame < 0 || frame >= count {
		return monsterGeneratorDrawFrames4BC750{}, false
	}

	if dr.Flags70Val&0x800 != 0 {
		frame = count - 1
		dr.ObjClass &^= object.ClassLight
		dr.ObjFlags &^= object.FlagFlicker
	}
	if timer := dr.UnionEffect().Field_108; timer != 0 {
		delay := uint32(data.frameDelay[state]) + 1
		computed := int32((delay*uint32(count) - timer) / delay)
		switch {
		case computed >= int32(count):
			frame = count - 1
		case computed < 0:
			frame = 0
		default:
			frame = int(computed)
		}
		timer--
		dr.UnionEffect().Field_108 = timer
		if timer == 0 {
			dr.Flags70Val = dr.Flags70Val&^0x400 | 0x800
		}
	}

	frames := monsterGeneratorDrawFrames4BC750{
		mainState: state,
		mainFrame: frame,
		terminal:  dr.Flags70Val&0x800 != 0,
	}
	if dr.Flags70Val&0xC00 == 0 && data.images[4] != nil && data.frameCount[4] != 0 {
		count := uint32(data.frameCount[4])
		frames.overlayFrame = int((gameFrame + dr.NetCode32) / (uint32(data.frameDelay[4]) + 1) % count)
		frames.hasOverlay = true
	}
	return frames, true
}

func monsterGeneratorDrawImage4BC750(data *monsterGeneratorDrawData4BC750, state, frame int) (noxrender.ImageHandle, bool) {
	if data == nil || state < 0 || state >= len(data.images) || frame < 0 {
		return nil, false
	}
	count := int(data.frameCount[state])
	if data.images[state] == nil || frame >= count {
		return nil, false
	}
	img := unsafe.Slice(data.images[state], count)[frame]
	return img, img != nil
}

func finishMonsterGeneratorDraw4BC750(dr *client.Drawable, frames monsterGeneratorDrawFrames4BC750) {
	if dr != nil && frames.terminal {
		dr.ObjFlags |= object.FlagBelow
	}
}

// drawMonsterGenerator4BC750 replaces the legacy drawer, which copied both
// dr and dr->field_76 into 32-bit integers before reading dr+432. That exact
// truncation faults for normal high-address drawables on 64-bit hosts.
func (c *Client) drawMonsterGenerator4BC750(dr *client.Drawable, vp *noxrender.Viewport) int {
	if dr == nil || dr.DrawData == nil {
		return 0
	}
	data := (*monsterGeneratorDrawData4BC750)(dr.DrawData)
	frames, ok := monsterGeneratorDrawFramesFor4BC750(dr, data, c.srv.Frame(), c.srv.Rand.Other.Int)
	if !ok {
		return 0
	}
	img, ok := monsterGeneratorDrawImage4BC750(data, frames.mainState, frames.mainFrame)
	if !ok {
		return 0
	}
	legacy.Nox_xxx_drawObject_4C4770_draw(vp, dr, img)
	if frames.hasOverlay {
		if img, ok := monsterGeneratorDrawImage4BC750(data, 4, frames.overlayFrame); ok {
			legacy.Nox_xxx_drawObject_4C4770_draw(vp, dr, img)
		}
	}
	finishMonsterGeneratorDraw4BC750(dr, frames)
	return 1
}

func (c *Client) callMonsterGeneratorDraw4BC750(dr *client.Drawable, vp *noxrender.Viewport) (int, bool) {
	if dr == nil || dr.DrawFuncPtr != legacy.Get_nox_thing_monster_gen_draw() {
		return 0, false
	}
	return c.drawMonsterGenerator4BC750(dr, vp), true
}
