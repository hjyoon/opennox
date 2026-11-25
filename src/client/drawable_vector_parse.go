package client

import (
	"fmt"
	"unsafe"

	"github.com/opennox/opennox/v1/client/noxrender"
	"github.com/opennox/opennox/v1/internal/binfile"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

// ParseAnimStateDrawData ports nox_things_animate_state_draw_parse and its
// single-sequence frame loader. The binary input remains the original PE32
// format, while dd uses the target architecture's native pointer width.
func ParseAnimStateDrawData(dd *AnimationStateDrawData, f *binfile.MemFile, imageRef func(id int, typ byte, name string) noxrender.ImageHandle) error {
	return ParseAnimStateDrawDataWithAllocator(dd, f, imageRef, func(count int) ([]noxrender.ImageHandle, func()) {
		return alloc.Make([]noxrender.ImageHandle{}, count)
	})
}

// ParseAnimStateDrawDataHigh uses the same binary parser and failure cleanup
// while requesting high foreign frame arrays for an explicit native owner.
func ParseAnimStateDrawDataHigh(dd *AnimationStateDrawData, f *binfile.MemFile, imageRef func(id int, typ byte, name string) noxrender.ImageHandle) error {
	return ParseAnimStateDrawDataWithAllocator(dd, f, imageRef, func(count int) ([]noxrender.ImageHandle, func()) {
		return alloc.MakeHigh([]noxrender.ImageHandle{}, count)
	})
}

// ParseAnimStateDrawDataWithAllocator lets the C ABI supply its current
// allocation policy while using the same layout, parser and cleanup.
func ParseAnimStateDrawDataWithAllocator(
	dd *AnimationStateDrawData, f *binfile.MemFile,
	imageRef func(id int, typ byte, name string) noxrender.ImageHandle,
	allocate func(count int) ([]noxrender.ImageHandle, func()),
) (err error) {
	dd.Size = uint32(unsafe.Sizeof(*dd))
	var frameAllocs [3]func()
	defer func() {
		if err != nil {
			for i, free := range frameAllocs {
				if free != nil {
					free()
					dd.Anim[i].Frames[0] = nil
				}
			}
		}
	}()

	for {
		cmd := f.ReadU32()
		if cmd == 0x454E4420 { // "END "
			return nil
		}
		params := f.ReadU32()
		if params&0xE == 0 {
			return fmt.Errorf("invalid animate-state parameters: %#x", params)
		}

		// These are legacy selector strings (normally "NULL" and empty).
		f.SkipString8()
		f.SkipString8()

		ind := 0
		if params&2 == 0 {
			if params&4 != 0 {
				ind = 1
			} else {
				ind = 2
			}
		}
		ani := &dd.Anim[ind]
		ani.Cnt40 = uint16(f.ReadU8())
		ani.Val42 = uint16(f.ReadU8())
		kind, readErr := f.ReadString8()
		if readErr != nil {
			return readErr
		}
		ani.Kind = ParseAnimKind(kind)

		if free := frameAllocs[ind]; free != nil {
			free()
			frameAllocs[ind] = nil
			ani.Frames[0] = nil
		}
		cnt := int(ani.Cnt40)
		if cnt == 0 {
			continue
		}
		if imageRef == nil {
			return fmt.Errorf("animate-state image resolver is nil")
		}
		frames, free := allocate(cnt)
		frameAllocs[ind] = free
		if len(frames) != cnt {
			return fmt.Errorf("cannot allocate %d animate-state frames", cnt)
		}
		ani.Frames[0] = &frames[0]
		for i := range frames {
			id := f.ReadI32()
			var (
				typ  byte
				name string
			)
			if id == -1 {
				typ = f.ReadU8()
				name, readErr = f.ReadString8()
				if readErr != nil {
					return readErr
				}
			}
			frames[i] = imageRef(int(id), typ, name)
		}
	}
}
