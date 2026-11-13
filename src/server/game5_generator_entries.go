package server

import "github.com/opennox/libs/types"

// The public raw generator entries share the established native world
// services rather than reading PE32 Object/update offsets in C.
func (s *Server) MonsterGeneratorPick54EBA0(generator, template *Object, destination *types.Pointf, rt MonsterGeneratorRuntime54E930) bool {
	return monsterGeneratorPick54EBA0(generator, template, generator.UpdateDataMonsterGen(), destination, s.monsterGeneratorNativeDeps54E930(rt))
}
func (s *Server) MonsterGeneratorSetCreaturePos54ED50(generator, template, target *Object, destination *types.Pointf, rt MonsterGeneratorRuntime54E930) bool {
	return monsterGeneratorSetCreaturePos54ED50(generator, template, target, destination, generator.UpdateDataMonsterGen(), s.monsterGeneratorNativeDeps54E930(rt))
}
func (s *Server) MonsterGeneratorRandomPosition54EF90(radius float32, center types.Pointf, template *Object, destination *types.Pointf, rt MonsterGeneratorRuntime54E930) bool {
	return monsterGeneratorRandomPosition54EF90(radius, center, template, destination, s.monsterGeneratorNativeDeps54E930(rt))
}
func (s *Server) MonsterGeneratorSpawn54F070(generator, template *Object, destination types.Pointf, rt MonsterGeneratorRuntime54E930) *Object {
	return monsterGeneratorSpawn54F070(generator, template, destination, &s.monsterGenerator54E930, s.monsterGeneratorNativeDeps54E930(rt))
}
func (s *Server) MonsterGeneratorCopyCreature54F2B0(template, spawned *Object, rt MonsterGeneratorRuntime54E930) {
	monsterGeneratorCopyCreature54F2B0(template, spawned, s.monsterGeneratorNativeDeps54E930(rt))
}
