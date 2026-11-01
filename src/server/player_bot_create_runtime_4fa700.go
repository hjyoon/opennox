package server

// PlayerBotCreateRuntime4FA700 supplies the original world services while
// keeping unit, player, monster and definition records native-width.
type PlayerBotCreateRuntime4FA700 struct {
	Allocate   func() *MonsterUpdateData
	TypeID     func(string) uint32
	Definition func(uint32) *MonsterDef
	FPS        func() uint32
	Random     func(int32, int32) int32
}
