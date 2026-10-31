package server

// MonsterNPCAnimationNativeRuntime533D00 retains the actual weapon and
// frame helper calls, selector boundary and shared scratch pointer.
type MonsterNPCAnimationNativeRuntime533D00 struct {
	WeaponAnimation func(uint32) int32
	Frames          func(int32) (int32, int32)
	Selector        func(*Object) int32
	Scratch         func() *MonsterAnim
}
