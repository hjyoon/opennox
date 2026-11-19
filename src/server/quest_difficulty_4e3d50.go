package server

// QuestDifficulty4E3D50 applies the stage and participating-player count with
// the original x87 arithmetic and binary32 store. The stage is unsigned even
// when the legacy ABI reports it as a signed integer.
func QuestDifficulty4E3D50(stage uint32, players int32, delta float32) float32 {
	scaled := objectForceMul53_52DF80(float64(players-1), objectForceMemory32_52DF80(delta))
	factor := objectForceAdd53_52DF80(scaled, 1)
	return experienceSpill32TowardZero(objectForceMul53_52DF80(float64(stage), factor))
}
