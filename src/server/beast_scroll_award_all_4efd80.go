package server

const (
	beastScrollAwardAllAdminMask4EFD80  = uint8(0x10)
	beastScrollAwardAllLevelCount4EFD80 = int32(41)
	beastScrollAwardAllFirstIndex4EFD80 = int32(1)
	beastScrollAwardAllAdminLevel4EFD80 = int32(1)
)

type beastScrollAwardAllHooks4EFD80[P comparable] struct {
	loadPlayerArg    func() P
	loadProtection   func(P) uint32
	resetProtection  func(uint32, int32)
	loadEngineFlags  func() uint8
	storeScrollLevel func(P, int32, uint32)
	awardProtection  func(uint32, int32, int32)
	applyProtection  func(uint32, P, int32)
}

// beastScrollAwardAll4EFD80 preserves GAME.EXE 004EFD80. The Player argument
// and initial protection token are loaded before protection reset. Engine
// flags are read only after reset, and the Admin decision remains fixed for
// the selected loop. Every index 1..40 is stored before a live
// protection-token reload and award callback.
//
// Both paths finish with one more live protection-token load followed by
// application to all 41 levels. Level zero is never written. There is no nil
// guard.
func beastScrollAwardAll4EFD80[P comparable](hooks beastScrollAwardAllHooks4EFD80[P]) {
	player := hooks.loadPlayerArg()
	protection := hooks.loadProtection(player)
	hooks.resetProtection(protection, 0)

	flags := hooks.loadEngineFlags()
	level := int32(0)
	if flags&beastScrollAwardAllAdminMask4EFD80 != 0 {
		level = beastScrollAwardAllAdminLevel4EFD80
	}
	for index := beastScrollAwardAllFirstIndex4EFD80; index < beastScrollAwardAllLevelCount4EFD80; index++ {
		hooks.storeScrollLevel(player, index, uint32(level))
		protection = hooks.loadProtection(player)
		hooks.awardProtection(protection, index, level)
	}

	protection = hooks.loadProtection(player)
	hooks.applyProtection(protection, player, beastScrollAwardAllLevelCount4EFD80)
}
