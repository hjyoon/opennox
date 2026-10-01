package opennox

import (
	"encoding/json"
	"reflect"
	"testing"
)

func e2ePetSnapshotFixture() e2eCampaignPetSnapshot {
	return e2eCampaignPetSnapshot{Version: 1, Map: "con02a", Pets: []e2eSavedPet{
		{TypeID: "Wolf", ScriptID: 0x11223344, Health: 40, MaxHealth: 40},
		{TypeID: "Urchin", ScriptID: 0x11223345, Health: 8, MaxHealth: 8},
		{TypeID: "Urchin", ScriptID: 0x11223346, Health: 8, MaxHealth: 8},
	}}
}

func TestE2ECampaignPetSnapshotRoundTrip(t *testing.T) {
	snap := e2ePetSnapshotFixture()
	data, err := json.Marshal(snap)
	if err != nil {
		t.Fatal(err)
	}
	var loaded e2eCampaignPetSnapshot
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(snap, loaded) {
		t.Fatalf("snapshot changed: %+v / %+v", snap, loaded)
	}
	actual := []e2eSavedPet{snap.Pets[2], snap.Pets[0], snap.Pets[1]}
	if err := loaded.compare("Con02a.map", actual); err != nil {
		t.Fatal(err)
	}
}

func TestE2ECampaignPetSnapshotRejectsIncompleteSeed(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(*e2eCampaignPetSnapshot)
	}{
		{"version", func(s *e2eCampaignPetSnapshot) { s.Version++ }},
		{"empty map", func(s *e2eCampaignPetSnapshot) { s.Map = "" }},
		{"unnormalized map", func(s *e2eCampaignPetSnapshot) { s.Map = "Con02a.map" }},
		{"missing pet", func(s *e2eCampaignPetSnapshot) { s.Pets = s.Pets[:2] }},
		{"zero identity", func(s *e2eCampaignPetSnapshot) { s.Pets[0].ScriptID = 0 }},
		{"duplicate identity", func(s *e2eCampaignPetSnapshot) { s.Pets[1].ScriptID = s.Pets[0].ScriptID }},
		{"dead pet", func(s *e2eCampaignPetSnapshot) { s.Pets[0].Health = 0 }},
		{"invalid health", func(s *e2eCampaignPetSnapshot) { s.Pets[0].MaxHealth = 1 }},
		{"wrong type", func(s *e2eCampaignPetSnapshot) { s.Pets[0].TypeID = "Spider" }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			snap := e2ePetSnapshotFixture()
			tc.mutate(&snap)
			if err := snap.validate(); err == nil {
				t.Fatal("invalid seed accepted")
			}
		})
	}
}

func TestE2ECampaignPetSnapshotDetectsReloadLoss(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(*string, *[]e2eSavedPet)
	}{
		{"wrong map", func(m *string, _ *[]e2eSavedPet) { *m = "con01a" }},
		{"all pets missing", func(_ *string, p *[]e2eSavedPet) { *p = nil }},
		{"one pet missing", func(_ *string, p *[]e2eSavedPet) { *p = (*p)[:2] }},
		{"extra pet", func(_ *string, p *[]e2eSavedPet) { *p = append(*p, (*p)[0]) }},
		{"duplicate pet", func(_ *string, p *[]e2eSavedPet) { (*p)[1] = (*p)[0] }},
		{"recreated identity", func(_ *string, p *[]e2eSavedPet) { (*p)[0].ScriptID++ }},
		{"changed type", func(_ *string, p *[]e2eSavedPet) { (*p)[0].TypeID = "Urchin" }},
		{"lost health", func(_ *string, p *[]e2eSavedPet) { (*p)[0].Health-- }},
		{"changed max health", func(_ *string, p *[]e2eSavedPet) { (*p)[0].MaxHealth++ }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			snap := e2ePetSnapshotFixture()
			actual := append([]e2eSavedPet(nil), snap.Pets...)
			mapName := snap.Map
			tc.mutate(&mapName, &actual)
			if err := snap.compare(mapName, actual); err == nil {
				t.Fatal("corrupted reload accepted")
			}
		})
	}
}
