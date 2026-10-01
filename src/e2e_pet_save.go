package opennox

import (
	"encoding/json"
	"fmt"
	"math"
	"os"

	"github.com/opennox/libs/datapath"
	"github.com/opennox/libs/object"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

// This test-only file lives outside the actual save slot. The seed process
// records scalar identities; the load process must recover pets from the stock
// map/player save, not from this snapshot or a stale native pointer.
const e2eCampaignPetSnapshotFile = ".e2e-campaign-pets.json"

type e2eSavedPet struct {
	TypeID    string `json:"type"`
	ScriptID  int32  `json:"script_id"`
	Health    uint16 `json:"health"`
	MaxHealth uint16 `json:"max_health"`
}

type e2eCampaignPetSnapshot struct {
	Version int           `json:"version"`
	Map     string        `json:"map"`
	Pets    []e2eSavedPet `json:"pets"`
}

func (snap e2eCampaignPetSnapshot) validate() error {
	if snap.Version != 1 || snap.Map == "" || snap.Map != e2eMapBaseName(snap.Map) || len(snap.Pets) != 3 {
		return fmt.Errorf("invalid campaign pet snapshot: version=%d map=%q count=%d", snap.Version, snap.Map, len(snap.Pets))
	}
	ids := make(map[int32]bool)
	types := make(map[string]int)
	for _, pet := range snap.Pets {
		if pet.ScriptID == 0 || ids[pet.ScriptID] || pet.Health == 0 || pet.Health > pet.MaxHealth {
			return fmt.Errorf("invalid/duplicate saved pet: %+v", pet)
		}
		ids[pet.ScriptID] = true
		types[pet.TypeID]++
	}
	if types["Wolf"] != 1 || types["Urchin"] != 2 || len(types) != 2 {
		return fmt.Errorf("campaign pet snapshot has unexpected creature types: %v", types)
	}
	return nil
}

func (snap e2eCampaignPetSnapshot) compare(mapName string, actual []e2eSavedPet) error {
	if err := snap.validate(); err != nil {
		return err
	}
	if got := e2eMapBaseName(mapName); got != snap.Map {
		return fmt.Errorf("loaded pet map=%q, want %q", got, snap.Map)
	}
	if len(actual) != len(snap.Pets) {
		return fmt.Errorf("loaded summoned creature count=%d, want %d: %+v", len(actual), len(snap.Pets), actual)
	}
	want := make(map[int32]e2eSavedPet)
	for _, pet := range snap.Pets {
		want[pet.ScriptID] = pet
	}
	for _, pet := range actual {
		if expected, ok := want[pet.ScriptID]; !ok || pet != expected {
			return fmt.Errorf("loaded pet=%+v, want saved identity=%+v (present=%t)", pet, expected, ok)
		}
		delete(want, pet.ScriptID)
	}
	return nil
}

func e2ePetSaveRecord(pet *server.Object) e2eSavedPet {
	return e2eSavedPet{
		TypeID: pet.ObjectTypeC().ID(), ScriptID: pet.ScriptIDVal,
		Health: pet.HealthData.Cur, MaxHealth: pet.HealthData.Max,
	}
}

func (sc *e2eScenario) CaptureCampaignPetSave(name string) {
	sc.add(0, name, func() {
		if !noxflags.HasGame(noxflags.GameModeCoop) || nox_xxx_gameGet_4DB1B0() || len(e2eTransitionPets.pets) != 3 {
			e2eError(fmt.Errorf("campaign pet snapshot: completed autosave and three live fixture pets are required"))
			return
		}
		world := make(map[*server.Object]bool)
		for obj := noxServer.Objs.First(); obj != nil; obj = obj.Next() {
			world[obj] = true
		}
		snap := e2eCampaignPetSnapshot{Version: 1, Map: e2eMapBaseName(legacy.Nox_xxx_mapGetMapName_409B40())}
		for _, pet := range e2eTransitionPets.pets {
			if !world[pet.object] || pet.object.HealthData == nil {
				e2eError(fmt.Errorf("campaign pet snapshot: %s is no longer live", pet.typeID))
				return
			}
			snap.Pets = append(snap.Pets, e2ePetSaveRecord(pet.object))
		}
		if err := snap.validate(); err != nil {
			e2eError(err)
			return
		}
		data, err := json.Marshal(snap)
		if err == nil {
			err = os.WriteFile(datapath.Save(e2eCampaignPetSnapshotFile), data, 0600)
		}
		if err != nil {
			e2eError(fmt.Errorf("write campaign pet snapshot: %w", err))
			return
		}
		e2eLog.Printf("CAMPAIGN PET SAVE SNAPSHOT: map=%s pets=%+v", snap.Map, snap.Pets)
	})
}

func e2eLoadedCampaignPets() []*server.Object {
	host := noxServer.Players.HostUnit()
	if host == nil {
		return nil
	}
	var pets []*server.Object
	for obj := noxServer.Objs.First(); obj != nil; obj = obj.Next() {
		if obj.ObjOwner == host && obj.Class().Has(object.ClassMonster) && obj.UpdateData != nil &&
			obj.UpdateDataMonster().StatusFlags.Has(object.MonStatusSummoned) {
			pets = append(pets, obj)
		}
	}
	return pets
}

func (sc *e2eScenario) AssertCampaignPetsReloaded(name string) {
	var snap e2eCampaignPetSnapshot
	sc.add(0, name+" read seed identities", func() {
		data, err := os.ReadFile(datapath.Save(e2eCampaignPetSnapshotFile))
		if err == nil {
			err = json.Unmarshal(data, &snap)
		}
		if err == nil {
			err = snap.validate()
		}
		if err != nil {
			e2eError(fmt.Errorf("read campaign pet snapshot: %w", err))
		}
	})
	sc.addWhen(0, name+" synchronize loaded pets", 240, func() bool {
		pets := e2eLoadedCampaignPets()
		if len(pets) != len(snap.Pets) {
			return true // Report missing/duplicate server pets without hiding it behind a client timeout.
		}
		for _, pet := range pets {
			code := noxServer.GetUnitNetCode(pet)
			if code <= 0 || code > math.MaxUint16 || noxClient.Objs.ByNetCode(uint16(code)) == nil {
				return false
			}
		}
		return true
	}, func() {
		host := noxServer.Players.HostUnit()
		if host == nil || !noxflags.HasGame(noxflags.GameModeCoop) {
			e2eError(fmt.Errorf("campaign pet reload: a live campaign player is required"))
			return
		}
		pets := e2eLoadedCampaignPets()
		var records []e2eSavedPet
		for _, pet := range pets {
			if pet.HealthData == nil {
				e2eError(fmt.Errorf("campaign pet reload: pet %p lost health data", pet))
				return
			}
			records = append(records, e2ePetSaveRecord(pet))
		}
		if err := snap.compare(legacy.Nox_xxx_mapGetMapName_409B40(), records); err != nil {
			e2eError(err)
			return
		}
		owned := make(map[*server.Object]bool)
		for obj := host.FirstOwned516(); obj != nil; obj = obj.NextOwned512() {
			owned[obj] = true
		}
		// These are four pointer-free, 32-byte nox_gui_summon_records. Stock
		// reload reports summoned creatures through Acquire, not MONITOR bit 0x80.
		// Check the actual HUD records and minimap membership instead of requiring
		// a server subclass bit that is not part of the reload contract.
		hud := make(map[uint32]uint32)
		for i := uintptr(0); i < 4; i++ {
			base := uintptr(1321052) + 32*i
			if *memmap.PtrUint32(0x5D4594, base+8) != 0 {
				hud[*memmap.PtrUint32(0x5D4594, base)] = *memmap.PtrUint32(0x5D4594, base+4)
			}
		}
		if len(hud) != len(pets) {
			e2eError(fmt.Errorf("campaign pet reload: HUD entries=%v, want %d pets", hud, len(pets)))
			return
		}
		minimap := make(map[uint32]bool)
		for dr := noxClient.Objs.FirstMinimapList(); dr != nil; dr = dr.Nox_xxx_cliNextMinimapObj_459EC0(dr) {
			minimap[dr.NetCode32] = true
		}
		e2eTransitionPets.pets, e2eTransitionPets.pixies = nil, nil
		e2eTransitionPets.from, e2eTransitionPets.to = snap.Map, snap.Map
		for _, pet := range pets {
			wire := uint16(noxServer.GetUnitNetCode(pet))
			drawable := noxClient.Objs.ByNetCode(wire)
			distance := math.Hypot(float64(pet.PosVec.X-host.PosVec.X), float64(pet.PosVec.Y-host.PosVec.Y))
			if !owned[pet] || !pet.SubClass().AsMonster().Has(object.MonsterMigrate) ||
				pet.Flags().HasAny(object.FlagDead|object.FlagDestroyed) || distance > 80 || drawable == nil ||
				!drawable.ObjFlags.Has(object.FlagEnabled) || drawable.ObjFlags.HasAny(object.FlagDead|object.FlagDestroyed) {
				e2eError(fmt.Errorf("campaign pet reload: pet not restored beside player: object=%p owned=%t subclass=%#x flags=%#x distance=%.3f drawable=%p", pet, owned[pet], uint32(pet.ObjSubClass), uint32(pet.ObjFlags), distance, drawable))
				return
			}
			if hud[uint32(wire)] != uint32(pet.TypeInd) || !minimap[drawable.NetCode32] {
				e2eError(fmt.Errorf("campaign pet reload: missing HUD/minimap pet: wire=%#x type=%d HUD=%v minimap=%t", wire, pet.TypeInd, hud, minimap[drawable.NetCode32]))
				return
			}
			clientDistance := math.Hypot(float64(drawable.PosVec.X)-float64(pet.PosVec.X), float64(drawable.PosVec.Y)-float64(pet.PosVec.Y))
			if clientDistance > 10 {
				e2eError(fmt.Errorf("campaign pet reload: stale client position: object=%p delta=%.3f", pet, clientDistance))
				return
			}
			e2eTransitionPets.pets = append(e2eTransitionPets.pets, e2eTransitionPet{object: pet, typeID: pet.ObjectTypeC().ID(), health: pet.HealthData.Cur, wire: wire})
			e2eLog.Printf("CAMPAIGN PET RELOADED: type=%s script-id=%d object=%p owner=%p health=%d/%d wire=%#x distance=%.3f client-delta=%.3f HUD=true minimap=true", pet.ObjectTypeC().ID(), pet.ScriptIDVal, pet, host, pet.HealthData.Cur, pet.HealthData.Max, wire, distance, clientDistance)
		}
	})
}
