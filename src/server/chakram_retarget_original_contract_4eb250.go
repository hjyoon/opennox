package server

import "github.com/opennox/libs/types"

// ChakramRetargetState4EB250 is the original shared search scratch. Reentrant
// raw entries intentionally replace it while callers retain their local data.
// This reference contract is frozen before adding the raw native engine.
type ChakramRetargetState4EB250 struct {
	Source, Owner, LastHit, Selected *Object
	BestDistance                     float32
}

// ChakramRetargetRuntime4EB250 names the three external world operations.
// Production uses the original startup PC53/toward-zero arithmetic. The
// second original FPU profile remains independently selectable for comparison.
type ChakramRetargetRuntime4EB250 struct {
	EachRect       func(types.Rectf, func(*Object, *types.Pointf), *types.Pointf)
	HasEnchant     func(*Object, uint32) bool
	CanSee         func(*Object, *Object) bool
	RoundToNearest bool
}
