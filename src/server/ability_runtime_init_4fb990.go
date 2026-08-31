package server

const (
	abilityRuntimePlayerSlots4FB990         = 32
	abilityRuntimeAbilitySlots4FB990        = int(AbilityMax)
	abilityRuntimeCooldownWordBytes4FB990   = 4
	abilityRuntimeCooldownBytes4FB990       = abilityRuntimePlayerSlots4FB990 * abilityRuntimeAbilitySlots4FB990 * abilityRuntimeCooldownWordBytes4FB990
	executingAbilityClassName4FB990         = "executingAbilityClass"
	executingAbilityClassRecordBytes4FB990  = 24
	executingAbilityClassPoolCapacity4FB990 = 64
)

// Init4FB990 is the native-width replacement for
// nox_xxx_allocArrayExecAbilities_4FB990. GAME.EXE cleared a fixed 32-player
// by six-ability PE32 cooldown matrix and installed a new 64-element allocator
// for 24-byte executingAbilityClass records. The Go representation preserves
// that fixed int32 matrix and one global execution-list head while widening
// record object and link pointers to native width. Nodes are managed by the Go
// allocator.
func (a *serverAbilities) Init4FB990() {
	a.cooldowns = [abilityRuntimePlayerSlots4FB990][AbilityMax]int32{}
	a.execList = nil
}

// Reset remains the package-level compatibility name used by focused server
// tests and non-session initialization paths.
func (a *serverAbilities) Reset() {
	a.Init4FB990()
}
