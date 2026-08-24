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
func ParseAnimStateDrawData(dd *AnimationStateDrawData, f *binfile.MemFile, imageRef func(id int, typ byte, name string) noxrender.ImageHandle) (err error) {
	dd.Size = uint32(unsafe.Sizeof(*dd))
	var frameAllocs []func()
	defer func() {
		if err != nil {
			for i := len(frameAllocs) - 1; i >= 0; i-- {
				frameAllocs[i]()
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

		cnt := int(ani.Cnt40)
		if cnt == 0 {
			continue
		}
		if imageRef == nil {
			return fmt.Errorf("animate-state image resolver is nil")
		}
		frames, free := alloc.Make([]noxrender.ImageHandle{}, cnt)
		frameAllocs = append(frameAllocs, free)
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
