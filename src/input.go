package opennox

import (
	"github.com/opennox/libs/client/keybind"

	"github.com/opennox/opennox/v1/client/input"
	"github.com/opennox/opennox/v1/common/memmap"
)

var (
	keyBinding *keybind.Binding
)

var noxMouseSelectOpt = []string{
	"Left",
	"Right",
	"Middle",
	"Wheel",
}

func inputInitMouse() {
	// 0047D8D0 publishes the device capability BYTE before the present DWORD.
	// SDL's input contract accepts three logical buttons; publish the handler's
	// capacity here instead of leaving both InputCfg constructors at zero.
	*memmap.PtrUint8(0x5D4594, 1193128) = byte(input.MouseButtonCount)
	// indicates that mouse is present so cursor should be drawn
	*memmap.PtrUint32(0x5D4594, 1193108) = 1
}
