package server

import (
	"unsafe"

	"github.com/opennox/libs/types"
)

// AwardSpellCollideData is the fixed-width four-byte record registered for
// AwardSpellCollide. GAME.EXE passes its complete uint32 value to the spell
// grant routine.
type AwardSpellCollideData struct {
	Spell uint32
}

type AwardSpellCollideRuntime4EAD20 struct {
	GrantSpell func(*Object, uint32, int32, int32, int32) int32
}

type awardSpellCollideNativeDeps4EAD20 struct {
	grantSpell func(*Object, uint32, int32, int32, int32) int32
}

func awardSpellCollideNative4EAD20(
	source, target *Object,
	collision *types.Pointf,
	deps awardSpellCollideNativeDeps4EAD20,
) int32 {
	return awardSpellCollide4EAD20(
		source,
		target,
		collision,
		awardSpellCollideHooks4EAD20[*Object, *AwardSpellCollideData]{
			loadCollideData: func(obj *Object) *AwardSpellCollideData {
				return (*AwardSpellCollideData)(obj.CollideData)
			},
			loadSpell: func(data *AwardSpellCollideData) uint32 {
				return data.Spell
			},
			grantSpell: deps.grantSpell,
		},
	)
}

// AwardSpellCollide4EAD20 binds GAME.EXE's registered callback to native-width
// Object pointers while retaining the original four-byte spell record and
// five-argument grant contract supplied by the legacy boundary.
func (s *Server) AwardSpellCollide4EAD20(
	source, target *Object,
	collision *types.Pointf,
	runtime AwardSpellCollideRuntime4EAD20,
) int32 {
	return awardSpellCollideNative4EAD20(
		source,
		target,
		collision,
		awardSpellCollideNativeDeps4EAD20{
			grantSpell: runtime.GrantSpell,
		},
	)
}

var (
	_ = [1]struct{}{}[4-unsafe.Sizeof(AwardSpellCollideData{})]
	_ = [1]struct{}{}[0-unsafe.Offsetof(AwardSpellCollideData{}.Spell)]
)
