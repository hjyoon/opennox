//go:build !server

package opennox

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/legacy"
)

func TestMonsterGeneratorDrawData4BC750UsesNativePointerLayout(t *testing.T) {
	var data monsterGeneratorDrawData4BC750
	pointerSize := unsafe.Sizeof(uintptr(0))
	wantImages := uintptr(4)
	if pointerSize == 8 {
		wantImages = 8
	}
	if got := unsafe.Offsetof(data.images); got != wantImages {
		t.Fatalf("images offset = %d, want %d", got, wantImages)
	}
	wantCounts := wantImages + 5*pointerSize
	if got := unsafe.Offsetof(data.frameCount); got != wantCounts {
		t.Fatalf("frame-count offset = %d, want %d", got, wantCounts)
	}
	if got := unsafe.Offsetof(data.frameDelay); got != wantCounts+5 {
		t.Fatalf("frame-delay offset = %d, want %d", got, wantCounts+5)
	}
	wantKinds := (wantCounts + 10 + 3) &^ 3
	if got := unsafe.Offsetof(data.animationKind); got != wantKinds {
		t.Fatalf("animation-kind offset = %d, want %d", got, wantKinds)
	}
	wantSize := wantKinds + 5*unsafe.Sizeof(uint32(0))
	if rem := wantSize % pointerSize; rem != 0 {
		wantSize += pointerSize - rem
	}
	if got := unsafe.Sizeof(data); got != wantSize {
		t.Fatalf("draw-data size = %d, want %d", got, wantSize)
	}
}

func TestMonsterGeneratorState4BC750(t *testing.T) {
	for _, tc := range []struct {
		flags uint32
		want  int
	}{
		{0, 0},
		{0x100, 1},
		{0x200, 2},
		{0x400, 3},
		{0x800, 3},
		{0x300, 1},
	} {
		if got := monsterGeneratorState4BC750(tc.flags); got != tc.want {
			t.Errorf("state(%#x) = %d, want %d", tc.flags, got, tc.want)
		}
	}
}

func TestMonsterGeneratorFrames4BC750HighAddressAndOverlay(t *testing.T) {
	dr := &client.Drawable{NetCode32: 1}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(dr)) <= uintptr(^uint32(0)) {
		t.Skipf("allocator returned a low address: %p", dr)
	}
	var imageBytes [5]byte
	mainImages := [...]noxrender.ImageHandle{
		noxrender.ImageHandle(unsafe.Pointer(&imageBytes[0])),
		noxrender.ImageHandle(unsafe.Pointer(&imageBytes[1])),
		noxrender.ImageHandle(unsafe.Pointer(&imageBytes[2])),
	}
	overlayImages := [...]noxrender.ImageHandle{
		noxrender.ImageHandle(unsafe.Pointer(&imageBytes[3])),
		noxrender.ImageHandle(unsafe.Pointer(&imageBytes[4])),
	}
	data := monsterGeneratorDrawData4BC750{}
	data.images[0] = &mainImages[0]
	data.frameCount[0] = uint8(len(mainImages))
	data.frameDelay[0] = 1
	data.animationKind[0] = 2
	data.images[4] = &overlayImages[0]
	data.frameCount[4] = uint8(len(overlayImages))
	data.animationKind[4] = 2

	frames, ok := monsterGeneratorDrawFramesFor4BC750(dr, &data, 5, nil)
	if !ok {
		t.Fatal("native frame selection rejected valid high-address drawable")
	}
	if frames.mainState != 0 || frames.mainFrame != 0 || !frames.hasOverlay || frames.overlayFrame != 0 || frames.terminal {
		t.Fatalf("frames = %+v, want state 0/frame 0 and overlay frame 0", frames)
	}
	if img, ok := monsterGeneratorDrawImage4BC750(&data, frames.mainState, frames.mainFrame); !ok || img != mainImages[0] {
		t.Fatalf("main image = (%p, %t), want (%p, true)", img, ok, mainImages[0])
	}
	if img, ok := monsterGeneratorDrawImage4BC750(&data, 4, frames.overlayFrame); !ok || img != overlayImages[0] {
		t.Fatalf("overlay image = (%p, %t), want (%p, true)", img, ok, overlayImages[0])
	}
}

func TestMonsterGeneratorFrames4BC750AnimationKindsAndBounds(t *testing.T) {
	var imageBytes [4]byte
	images := [...]noxrender.ImageHandle{
		noxrender.ImageHandle(unsafe.Pointer(&imageBytes[0])),
		noxrender.ImageHandle(unsafe.Pointer(&imageBytes[1])),
		noxrender.ImageHandle(unsafe.Pointer(&imageBytes[2])),
		noxrender.ImageHandle(unsafe.Pointer(&imageBytes[3])),
	}
	data := monsterGeneratorDrawData4BC750{}
	data.images[0] = &images[0]
	data.frameCount[0] = uint8(len(images))
	dr := &client.Drawable{}

	data.animationKind[0] = 4
	randomCalled := false
	frames, ok := monsterGeneratorDrawFramesFor4BC750(dr, &data, 0, func(minimum, maximum int) int {
		randomCalled = true
		if minimum != 0 || maximum != len(images)-1 {
			t.Fatalf("random bounds = %d..%d, want 0..%d", minimum, maximum, len(images)-1)
		}
		return maximum
	})
	if !ok || !randomCalled || frames.mainFrame != len(images)-1 {
		t.Fatalf("random frames = (%+v, %t), called=%t", frames, ok, randomCalled)
	}

	data.animationKind[0] = 5
	dr.AnimFrameSlave = 2
	frames, ok = monsterGeneratorDrawFramesFor4BC750(dr, &data, 0, nil)
	if !ok || frames.mainFrame != 2 {
		t.Fatalf("slave frames = (%+v, %t), want frame 2", frames, ok)
	}
	dr.AnimFrameSlave = uint32(len(images))
	if _, ok := monsterGeneratorDrawFramesFor4BC750(dr, &data, 0, nil); ok {
		t.Fatal("out-of-range slave frame was accepted")
	}
	data.animationKind[0] = 0
	if _, ok := monsterGeneratorDrawFramesFor4BC750(dr, &data, 0, nil); ok {
		t.Fatal("unsafe kind-0 address-as-index animation was accepted")
	}
}

func TestMonsterGeneratorFrames4BC750TimerAndTerminalFlags(t *testing.T) {
	var imageBytes [4]byte
	images := [...]noxrender.ImageHandle{
		noxrender.ImageHandle(unsafe.Pointer(&imageBytes[0])),
		noxrender.ImageHandle(unsafe.Pointer(&imageBytes[1])),
		noxrender.ImageHandle(unsafe.Pointer(&imageBytes[2])),
		noxrender.ImageHandle(unsafe.Pointer(&imageBytes[3])),
	}
	data := monsterGeneratorDrawData4BC750{}
	data.images[3] = &images[0]
	data.frameCount[3] = uint8(len(images))
	data.frameDelay[3] = 1
	data.animationKind[3] = 2

	dr := &client.Drawable{
		Flags70Val: 0x400,
		ObjClass:   object.ClassMonsterGenerator | object.ClassLight,
		ObjFlags:   object.FlagActive | object.FlagFlicker,
	}
	dr.UnionEffect().Field_108 = 1
	frames, ok := monsterGeneratorDrawFramesFor4BC750(dr, &data, 0, nil)
	if !ok || frames.mainFrame != len(images)-1 || !frames.terminal {
		t.Fatalf("timer frames = (%+v, %t), want final terminal frame", frames, ok)
	}
	if dr.UnionEffect().Field_108 != 0 || dr.Flags70Val&0xC00 != 0x800 {
		t.Fatalf("timer transition = timer:%d flags:%#x, want timer 0/flags 0x800", dr.UnionEffect().Field_108, dr.Flags70Val)
	}
	if !dr.ObjClass.Has(object.ClassLight) || !dr.ObjFlags.Has(object.FlagFlicker) {
		t.Fatal("timer-completed frame prematurely cleared initial-state class/flags")
	}
	finishMonsterGeneratorDraw4BC750(dr, frames)
	if !dr.ObjFlags.Has(object.FlagBelow) {
		t.Fatal("terminal draw did not set FlagBelow")
	}

	dr.Flags70Val = 0x800
	dr.ObjClass |= object.ClassLight
	dr.ObjFlags |= object.FlagFlicker
	frames, ok = monsterGeneratorDrawFramesFor4BC750(dr, &data, 0, nil)
	if !ok || frames.mainFrame != len(images)-1 || !frames.terminal {
		t.Fatalf("forced-terminal frames = (%+v, %t)", frames, ok)
	}
	if dr.ObjClass.Has(object.ClassLight) || dr.ObjFlags.Has(object.FlagFlicker) {
		t.Fatal("forced-terminal draw did not clear light/flicker flags")
	}
}

func TestMonsterGeneratorDrawDispatch4BC750(t *testing.T) {
	fn := legacy.Get_nox_thing_monster_gen_draw()
	if fn == nil {
		t.Fatal("monster-generator callback pointer is nil")
	}
	dr := &client.Drawable{DrawFuncPtr: fn}
	if got, ok := (*Client)(nil).callMonsterGeneratorDraw4BC750(dr, nil); !ok || got != 0 {
		t.Fatalf("monster-generator dispatch = (%d, %t), want (0, true)", got, ok)
	}
	dr.DrawFuncPtr = nil
	if got, ok := (*Client)(nil).callMonsterGeneratorDraw4BC750(dr, nil); ok || got != 0 {
		t.Fatalf("non-generator dispatch = (%d, %t), want (0, false)", got, ok)
	}
}
