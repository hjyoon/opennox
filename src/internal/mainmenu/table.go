// Package mainmenu contains native representations of the packed PE32 data
// used while drawing the first menu screen.
package mainmenu

// EyeSpec mirrors one 48-byte record at GAME.EXE VA 0x5B0380. Pointer fields
// are deliberately omitted: callers attach native image pointers separately.
type EyeSpec struct {
	Name                   string
	X, Y                   int
	Hidden                 bool
	VisibleMin, VisibleMax uint32
	HiddenMin, HiddenMax   uint32
	InitialPhaseTicks      uint32
	InitialBlinkTicks      uint32
	InitialBlinkCooldown   uint32
}

var eyeSpecs = [...]EyeSpec{
	{Name: "BlinkEyeF", X: 184, Y: 237, VisibleMin: 120, VisibleMax: 230, HiddenMin: 105, HiddenMax: 200, InitialPhaseTicks: 1},
	{Name: "BlinkEyeD", X: 126, Y: 377, Hidden: true, VisibleMin: 100, VisibleMax: 206, HiddenMin: 105, HiddenMax: 196, InitialPhaseTicks: 3},
	{Name: "BlinkEyeF", X: 455, Y: 219, Hidden: true, VisibleMin: 100, VisibleMax: 242, HiddenMin: 87, HiddenMax: 182, InitialPhaseTicks: 5},
	{Name: "BlinkEyeD", X: 569, Y: 313, Hidden: true, VisibleMin: 100, VisibleMax: 197, HiddenMin: 105, HiddenMax: 188, InitialPhaseTicks: 7},
	{Name: "BlinkEyeF", X: 548, Y: 64, VisibleMin: 80, VisibleMax: 180, HiddenMin: 132, HiddenMax: 224, InitialPhaseTicks: 8},
}

// EyeSpecs returns a copy so runtime animation state cannot alter the oracle
// constants.
func EyeSpecs() [len(eyeSpecs)]EyeSpec {
	return eyeSpecs
}
