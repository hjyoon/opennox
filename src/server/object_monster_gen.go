package server

type MonsterGenUpdateData struct {
	Field0          [12]*Object    // 0, 0
	Field48         uint32         // 12, 48
	FuncInd52       uint32         // 13, 52
	Field56         uint32         // 14, 56
	FuncInd60       uint32         // 15, 60
	Field64         uint32         // 16, 64
	FuncInd68       uint32         // 17, 68
	ScriptCollision ScriptCallback // 18, 72
	SpawnRate       [3]uint8       // 20, 80; normal-mode selector per generator group
	QuestSpawnRate  [3]uint8       // 20, 83; Quest selector per generator group
	ActiveCount     uint8          // 21, 86
	MaxActive       uint8          // 21, 87
	Frame88         uint32         // 22, 88
	Field92         uint32         // 23, 92
	HealthSamples   [32]uint16     // 24, 96; per-recipient damage-number cache
	Field160        uint32         // 40, 160
}
