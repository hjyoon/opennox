package server

const ObjectSpecialMask527E50 = uint32(0x0fff0000)

// ObjectSpecialRuntime527E50 names the nine child boundaries of GAME.EXE
// 00527E50. Report return words are ignored by the original root.
type ObjectSpecialRuntime527E50 struct {
	Animation func(int32, *Object) int32
	Enemy     func(*Object, *Object) int32
	Health    func(int32, *Object) int32
	Hidden    func(int32, *Object) int32
	XStatus   func(int32, *Object) int32
	Height    func(int32, *Object) int32
	Enchant   func(int32, *Object) int32
	TeamBase  func(int32, *Object) int32
	NPC       func(int32, *Object) int32
}

type ObjectSpecialFlagHooks527E50 struct {
	Animation, Health, Hidden, XStatus, Height, Enchant, TeamBase, NPC func()
}

// ConsumeObjectSpecialFlagsNative527E50 tests each live word in original
// order and clears its bit after the callback, retaining callback mutations.
func ConsumeObjectSpecialFlagsNative527E50(state *uint32, hooks ObjectSpecialFlagHooks527E50) bool {
	if state == nil || *state&ObjectSpecialMask527E50 == 0 {
		return false
	}
	steps := [...]struct {
		bit  uint32
		call func()
	}{
		{0x00010000, hooks.Animation},
		{0x00020000, hooks.Health},
		{0x00040000, hooks.Hidden},
		{0x00080000, hooks.XStatus},
		{0x00400000, hooks.Height},
		{0x00800000, hooks.Enchant},
		{0x02000000, hooks.TeamBase},
		{0x04000000, hooks.NPC},
	}
	for _, step := range steps {
		if *state&step.bit != 0 {
			if step.call != nil {
				step.call()
			}
			*state &^= step.bit
		}
	}
	return true
}

// ObjectSpecialUpdateNative527E50 preserves the raw public entry: recipient,
// update and Player are read before the nil-unit and unsigned byte-index gates.
// This entry does not apply the ordinary sender's class/nil policies.
func ObjectSpecialUpdateNative527E50(recipient, unit *Object, runtime ObjectSpecialRuntime527E50) int32 {
	update := (*PlayerUpdateData)(recipient.UpdateData)
	index := update.Player.PlayerInd
	return ObjectSpecialUpdateForIndexNative527E50(index, recipient, unit, runtime)
}

// ObjectSpecialUpdateForIndexNative527E50 accepts the already cached index
// used by the raw entry and by the ordinary sender after its existing checks.
func ObjectSpecialUpdateForIndexNative527E50(index byte, recipient, unit *Object, runtime ObjectSpecialRuntime527E50) int32 {
	if unit == nil || index >= 32 {
		return 1
	}
	i := int32(index)
	if !ConsumeObjectSpecialFlagsNative527E50(&unit.Field140[index], ObjectSpecialFlagHooks527E50{
		Animation: func() { runtime.Animation(i, unit) },
		Health: func() {
			if runtime.Enemy(recipient, unit) == 0 {
				runtime.Health(i, unit)
			}
		},
		Hidden:   func() { runtime.Hidden(i, unit) },
		XStatus:  func() { runtime.XStatus(i, unit) },
		Height:   func() { runtime.Height(i, unit) },
		Enchant:  func() { runtime.Enchant(i, unit) },
		TeamBase: func() { runtime.TeamBase(i, unit) },
		NPC:      func() { runtime.NPC(i, unit) },
	}) {
		return 0
	}
	return 1
}
