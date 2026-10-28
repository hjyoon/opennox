package server

import "testing"

func TestTypePredicates534A10KeepIndependentWordCaches(t *testing.T) {
	s := new(Server)
	s.Types.fast.plant = 300
	s.Types.fast.zombie = 100
	s.Types.fast.zombieVile = 200
	view := s.TypePredicateRuntime534A10()
	if view.Read(0) != 0 || view.Read(1) != 0 || view.Read(2) != 0 {
		t.Fatal("general ID lookups warmed the independent original caches")
	}
	view.Write(0, 0x10000+300)
	view.Write(1, 0x10000+100)
	view.Write(2, 0xffffffff)
	if s.IsPlant(&Object{TypeInd: 300}) || s.IsZombie(&Object{TypeInd: 100}) || s.IsZombie(&Object{TypeInd: 65535}) {
		t.Fatal("original DWORD cache was truncated to a type WORD")
	}
	view.Write(0, 300)
	view.Write(1, 100)
	view.Write(2, 200)
	if !s.IsPlant(&Object{TypeInd: 300}) || !s.IsZombie(&Object{TypeInd: 100}) || !s.IsZombie(&Object{TypeInd: 200}) {
		t.Fatal("shared native predicate owner does not retain the original identities")
	}
	if s.Types.CarnivorousPlantID() != 300 || s.Types.ZombieID() != 100 || s.Types.VileZombieID() != 200 {
		t.Fatal("independent predicate writes changed general type caches")
	}
}
