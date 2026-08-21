package server

// SkeletonInit4F0370 preserves GAME.EXE 004F0370's single RET instruction.
// The original callback accepts one object pointer but does not read it,
// mutate memory, invoke another callback, or produce an observable result.
//
//go:noinline
func SkeletonInit4F0370(_ *Object) {}
