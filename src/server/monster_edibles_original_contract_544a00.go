package server

import (
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/unit/ai"
)

// MonsterEdibleSearchState544A00 describes the original shared search scratch.
// The reference contract is frozen before adding native engine bindings.
type MonsterEdibleSearchState544A00 struct {
	BestDistance float32
	BestObject   *Object
}

// MonsterEdibleNativeRuntime544A00 keeps world operations explicit. Flag
// observers must delegate to the real getter; they do not supply fake answers.
type MonsterEdibleNativeRuntime544A00 struct {
	EachCircle  func(*types.Pointf, float32, func(*Object, *Object), *Object)
	CanInteract func(*Object, *Object, int) bool
	Quest       func() bool
	Online      func() bool
	Push        func(*Object, ai.ActionType) *AIStackItem
}
