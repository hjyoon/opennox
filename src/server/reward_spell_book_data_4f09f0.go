package server

// rewardSpellDefinition4F09F0 is the native fixed-width meaning of one
// original twelve-byte row at GAME.EXE 005B9900. The creator reads only the
// low byte of the original weight dword; spell IDs and slot masks remain exact
// uint32 game values.
type rewardSpellDefinition4F09F0 struct {
	Weight  uint8
	SpellID uint32
	Slots   uint32
}

// rewardSpellDefinitions4F09F0 contains 46 weighted rows, ten zero-weight
// disabled rows, and the first zero-ID sentinel. Its first weight is also the
// final dword of the preceding reward-marker category table in GAME.EXE.
var rewardSpellDefinitions4F09F0 = [...]rewardSpellDefinition4F09F0{
	{Weight: 4, SpellID: 0x47, Slots: 0x1f},
	{Weight: 4, SpellID: 0x1b, Slots: 0x1f},
	{Weight: 4, SpellID: 0x32, Slots: 0x1f},
	{Weight: 1, SpellID: 0x22, Slots: 0x1f},
	{Weight: 1, SpellID: 0x81, Slots: 0x1f},
	{Weight: 4, SpellID: 0x86, Slots: 0x1f},
	{Weight: 4, SpellID: 0x0e, Slots: 0x1f},
	{Weight: 4, SpellID: 0x80, Slots: 0x1f},
	{Weight: 4, SpellID: 0x34, Slots: 0x1f},
	{Weight: 4, SpellID: 0x3a, Slots: 0x1f},
	{Weight: 2, SpellID: 0x3c, Slots: 0x1f},
	{Weight: 2, SpellID: 0x09, Slots: 0x1f},
	{Weight: 2, SpellID: 0x29, Slots: 0x1f},
	{Weight: 1, SpellID: 0x0d, Slots: 0x1f},
	{Weight: 1, SpellID: 0x36, Slots: 0x1f},
	{Weight: 1, SpellID: 0x26, Slots: 0x1f},
	{Weight: 1, SpellID: 0x2a, Slots: 0x1f},
	{Weight: 1, SpellID: 0x48, Slots: 0x1f},
	{Weight: 1, SpellID: 0x3d, Slots: 0x1f},
	{Weight: 1, SpellID: 0x3e, Slots: 0x1f},
	{Weight: 1, SpellID: 0x40, Slots: 0x1f},
	{Weight: 4, SpellID: 0x24, Slots: 0x1e},
	{Weight: 4, SpellID: 0x18, Slots: 0x1e},
	{Weight: 4, SpellID: 0x15, Slots: 0x1e},
	{Weight: 4, SpellID: 0x0a, Slots: 0x1e},
	{Weight: 1, SpellID: 0x01, Slots: 0x1e},
	{Weight: 1, SpellID: 0x08, Slots: 0x1e},
	{Weight: 4, SpellID: 0x23, Slots: 0x1e},
	{Weight: 4, SpellID: 0x82, Slots: 0x1e},
	{Weight: 2, SpellID: 0x04, Slots: 0x1e},
	{Weight: 2, SpellID: 0x05, Slots: 0x1e},
	{Weight: 4, SpellID: 0x2b, Slots: 0x1c},
	{Weight: 2, SpellID: 0x43, Slots: 0x1c},
	{Weight: 2, SpellID: 0x16, Slots: 0x1c},
	{Weight: 2, SpellID: 0x84, Slots: 0x1c},
	{Weight: 1, SpellID: 0x0c, Slots: 0x1c},
	{Weight: 1, SpellID: 0x1a, Slots: 0x1c},
	{Weight: 4, SpellID: 0x1d, Slots: 0x1c},
	{Weight: 2, SpellID: 0x25, Slots: 0x1c},
	{Weight: 1, SpellID: 0x4a, Slots: 0x1c},
	{Weight: 4, SpellID: 0x17, Slots: 0x18},
	{Weight: 4, SpellID: 0x33, Slots: 0x18},
	{Weight: 2, SpellID: 0x10, Slots: 0x18},
	{Weight: 2, SpellID: 0x27, Slots: 0x18},
	{Weight: 2, SpellID: 0x87, Slots: 0x18},
	{Weight: 2, SpellID: 0x88, Slots: 0x18},
	{SpellID: 0x13},
	{SpellID: 0x1f},
	{SpellID: 0x21},
	{SpellID: 0x28},
	{SpellID: 0x38},
	{SpellID: 0x41},
	{SpellID: 0x42},
	{SpellID: 0x46},
	{SpellID: 0x7f},
	{SpellID: 0x74},
	{Slots: 0x1f},
}
