package server

// TowerInit4F0440 preserves GAME.EXE 004F0440's single RET instruction.
// The original callback accepts one object pointer but does not read it,
// mutate memory, invoke another callback, or produce an observable result.
//
//go:noinline
func TowerInit4F0440(_ *Object) {}
