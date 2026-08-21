package server

// ProjectileInit4F0380 preserves GAME.EXE 004F0380's single RET instruction.
// The original callback accepts one object pointer but does not read it,
// mutate memory, invoke another callback, or produce an observable result.
//
//go:noinline
func ProjectileInit4F0380(_ *Object) {}
