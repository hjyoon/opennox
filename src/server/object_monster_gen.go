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
	Field96         uint32         // 24, 96
	Field100        uint32         // 25, 100
	Field104        uint32         // 26, 104
	Field108        uint32         // 27, 108
	Field112        uint32         // 28, 112
	Field116        uint32         // 29, 116
	Field120        uint32         // 30, 120
	Field124        uint32         // 31, 124
	Field128        uint32         // 32, 128
	Field132        uint32         // 33, 132
	Field136        uint32         // 34, 136
	Field140        uint32         // 35, 140
	Field144        uint32         // 36, 144
	Field148        uint32         // 37, 148
	Field152        uint32         // 38, 152
	Field156        uint32         // 39, 156
	Field160        uint32         // 40, 160
}
