package server

// MonsterAnimationUpdateNativeRuntime50A850 keeps the actual animation
// record provider boundary. The parent owns cached sprite-byte updates.
type MonsterAnimationUpdateNativeRuntime50A850 struct {
	Animation func(*Object) *MonsterAnim
}
