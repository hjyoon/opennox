package server

// GruntInit4F0360 preserves GAME.EXE 004F0360's single RET instruction.
// The original callback accepts one object pointer but does not read it,
// mutate memory, invoke another callback, or produce an observable result.
//
//go:noinline
func GruntInit4F0360(_ *Object) {}
