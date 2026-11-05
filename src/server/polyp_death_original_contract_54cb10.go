package server

import "github.com/opennox/libs/types"

// PolypDeathState54CB10 is the shared DWORD type cache in GAME.EXE 0054CB10.
// The raw C entry's contract is separate from ordinary dispatcher admission.
type PolypDeathState54CB10 struct {
	TypeIndex uint32
}

// PolypDeathRawRuntime54CB10 declares only the original world boundaries.
// Source/update loads, cached identity, arithmetic and stores belong to the
// raw native body. RoundToNearest selects the independent reference control;
// the production startup contract uses its zero value.
type PolypDeathRawRuntime54CB10 struct {
	SimpleObjectDeathRuntime54CAE0
	Resolve        func() uint32
	New            func(uint32) *Object
	Create         func(*Object, *Object, types.Pointf, uint32)
	Balance        func() float64
	FPS            func() uint32
	RoundToNearest bool
}
