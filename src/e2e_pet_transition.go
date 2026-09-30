package opennox

import (
	"fmt"
	"math"
	"strings"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/types"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

type e2eTransitionPet struct {
	object *server.Object
	typeID string
	health uint16
	wire   uint16
}

var e2eTransitionPets struct {
	pets     []e2eTransitionPet
	pixies   []e2eTransitionPet
	from, to string
}

// Use the real summon allocator, ownership/monitor reports and ordinary
// update/network ticks. The fixture never injects migration/status flags.
func (sc *e2eScenario) CreateTransitionSummons(name string) {
	sc.add(0, name, func() {
		host := noxServer.Players.HostUnit()
		if host == nil || !noxflags.HasGame(noxflags.GameModeCoop) {
			e2eError(fmt.Errorf("pet transition: a live campaign player is required"))
			return
		}
		e2eTransitionPets.pets = nil
		e2eTransitionPets.pixies = nil
		for i, typeID := range []string{"Wolf", "Wolf", "Urchin"} {
			pos := types.Pointf{X: host.PosVec.X + float32(20+12*i), Y: host.PosVec.Y}
			typ := noxServer.Types.ByID(typeID)
			if typ == nil {
				e2eError(fmt.Errorf("pet transition: stock type %q is missing", typeID))
				return
			}
			pet := nox_xxx_unitDoSummonAt_5016C0(int32(typ.Ind()), &pos, host, uint8(host.Direction1))
			if pet == nil || pet.ObjOwner != host || pet.HealthData == nil || pet.UpdateData == nil {
				e2eError(fmt.Errorf("pet transition: summon %s is not initialized/owned: %p", typeID, pet))
				return
			}
			e2eTransitionPets.pets = append(e2eTransitionPets.pets, e2eTransitionPet{object: pet, typeID: typeID})
			e2eLog.Printf("PET CREATED: map=%q type=%s object=%p owner=%p subclass=%#x status=%#x flags=%#x",
				legacy.Nox_xxx_mapGetMapName_409B40(), typeID, pet, pet.ObjOwner,
				uint32(pet.ObjSubClass), uint32(pet.UpdateDataMonster().StatusFlags), uint32(pet.ObjFlags))
		}
	})
	sc.waitTransitionSummonsReady(name)
}

// Exercise duration-spell completion as well as the summon allocator. Guides
// and spells are awarded through gameplay services, not by editing migration,
// ownership or completion flags. The final Urchin is born unowned and acquired
// only by the real Charm spell.
func (sc *e2eScenario) CreateTransitionSpellPets(name string) {
	sc.add(0, name+" prepare spell awards", func() {
		host := noxServer.Players.HostUnit()
		if host == nil || !noxflags.HasGame(noxflags.GameModeCoop) {
			e2eError(fmt.Errorf("pet transition: a live campaign player is required"))
			return
		}
		e2eTransitionPets.pets = nil
		e2eTransitionPets.pixies = nil
		for _, id := range []spell.ID{spell.SPELL_SUMMON_WOLF, spell.SPELL_SUMMON_URCHIN} {
			guide := int32(id) - 74
			noxServer.AwardBeastGuide4FAE80(host, guide, 1)
			if host.UpdateDataPlayer().Player.BeastScrollLvl[guide] == 0 {
				e2eError(fmt.Errorf("pet transition: guide %d award failed", guide))
				return
			}
		}
		if legacy.Nox_xxx_spellGrantToPlayer_4FB550(host, spell.SPELL_CHARM, 0, 0, 1) != 1 {
			e2eError(fmt.Errorf("pet transition: Charm award failed"))
		}
	})
	for i, cast := range []struct {
		id     spell.ID
		typeID string
	}{
		{spell.SPELL_SUMMON_WOLF, "Wolf"},
		{spell.SPELL_SUMMON_URCHIN, "Urchin"},
		{spell.SPELL_CHARM, "Urchin"},
	} {
		var (
			before  map[*server.Object]bool
			charmed *server.Object
			created *server.Object
		)
		sc.add(0, fmt.Sprintf("%s cast %s", name, cast.id), func() {
			host := noxServer.Players.HostUnit()
			before = make(map[*server.Object]bool)
			for obj := noxServer.Objs.First(); obj != nil; obj = obj.Next() {
				before[obj] = true
			}
			pos := host.PosVec.Add(types.Ptf(float32(20+12*i), 0))
			if cast.id == spell.SPELL_CHARM {
				charmed = noxServer.NewObjectByTypeID(cast.typeID)
				if charmed == nil {
					e2eError(fmt.Errorf("pet transition: cannot create unowned Charm target"))
					return
				}
				noxServer.CreateObjectAt(charmed, nil, pos)
				noxServer.ObjectsAddPending()
				if charmed.ObjOwner != nil {
					e2eError(fmt.Errorf("pet transition: Charm target was already owned"))
					return
				}
			}
			if !noxServer.castSpellBy(cast.id, 1, host, charmed, pos) {
				e2eError(fmt.Errorf("pet transition: %s cast was rejected", cast.id))
				return
			}
			e2eLog.Printf("PET SPELL CAST: spell=%s target=%p frame=%d", cast.id, charmed, noxServer.Frame())
		})
		sc.addWhen(0, fmt.Sprintf("%s complete %s", name, cast.id), 1200, func() bool {
			host := noxServer.Players.HostUnit()
			for obj := noxServer.Objs.First(); obj != nil; obj = obj.Next() {
				if before[obj] || obj.ObjOwner != host || obj.ObjectTypeC().ID() != cast.typeID ||
					!obj.Class().Has(object.ClassMonster) || obj.UpdateData == nil || obj.HealthData == nil {
					continue
				}
				if charmed != nil && obj != charmed {
					continue
				}
				if !obj.SubClass().AsMonster().Has(object.MonsterMigrate) || !obj.UpdateDataMonster().StatusFlags.Has(object.MonStatusSummoned) {
					continue
				}
				created = obj
				return true
			}
			return false
		}, func() {
			e2eTransitionPets.pets = append(e2eTransitionPets.pets, e2eTransitionPet{object: created, typeID: cast.typeID})
			e2eLog.Printf("PET SPELL COMPLETED: spell=%s object=%p owner=%p frame=%d", cast.id, created, created.ObjOwner, noxServer.Frame())
		})
		sc.Wait(3, name+" retire completed duration")
	}
	sc.waitTransitionSummonsReady(name)
}

func (sc *e2eScenario) waitTransitionSummonsReady(name string) {
	sc.addWhen(0, name+" wait for synchronized summons", 240, func() bool {
		if len(e2eTransitionPets.pets) != 3 {
			return false
		}
		for _, pet := range e2eTransitionPets.pets {
			code := noxServer.GetUnitNetCode(pet.object)
			if code <= 0 || code > math.MaxUint16 || noxClient.Objs.ByNetCode(uint16(code)) == nil {
				return false
			}
		}
		return true
	}, func() {
		for i := range e2eTransitionPets.pets {
			pet := &e2eTransitionPets.pets[i]
			pet.health = pet.object.HealthData.Cur
			pet.wire = uint16(noxServer.GetUnitNetCode(pet.object))
			if pet.health == 0 || pet.object.ObjOwner != noxServer.Players.HostUnit() ||
				!pet.object.SubClass().AsMonster().Has(object.MonsterMigrate) {
				e2eError(fmt.Errorf("pet transition: invalid ready summon %s: health=%d owner=%p subclass=%#x",
					pet.typeID, pet.health, pet.object.ObjOwner, uint32(pet.object.ObjSubClass)))
				return
			}
			e2eLog.Printf("PET READY: type=%s object=%p wire=%#x health=%d pos=%v", pet.typeID, pet.object, pet.wire, pet.health, pet.object.PosVec)
		}
	})
}

func (sc *e2eScenario) CreateTransitionPixies(name string) {
	sc.add(0, name, func() {
		host := noxServer.Players.HostUnit()
		if host == nil || !noxServer.castSpellBy(spell.SPELL_PIXIE_SWARM, 1, host, host, host.PosVec) {
			e2eError(fmt.Errorf("pet transition: Pixie Swarm cast was rejected"))
			return
		}
		e2eLog.Printf("PET PIXIE CAST: owner=%p frame=%d desired=%g", host, noxServer.Frame(), noxServer.Balance.FloatInd("PixieCount", 0))
	})
	sc.addWhen(0, name+" wait for pixies", 240, func() bool {
		host := noxServer.Players.HostUnit()
		var pets []e2eTransitionPet
		for obj := noxServer.Objs.MissileList; obj != nil; obj = obj.Next() {
			if obj.ObjOwner != host || int(obj.TypeInd) != noxServer.Types.PixieID() {
				continue
			}
			wire := noxServer.GetUnitNetCode(obj)
			if wire <= 0 || wire > math.MaxUint16 || noxClient.Objs.ByNetCode(uint16(wire)) == nil {
				return false
			}
			pets = append(pets, e2eTransitionPet{object: obj, typeID: "Pixie", wire: uint16(wire)})
		}
		if len(pets) != int(noxServer.Balance.FloatInd("PixieCount", 0)) || len(pets) == 0 {
			return false
		}
		e2eTransitionPets.pixies = pets
		return true
	}, func() {
		for _, pet := range e2eTransitionPets.pixies {
			e2eLog.Printf("PET PIXIE READY: object=%p wire=%#x pos=%v deadline=%d frame=%d", pet.object, pet.wire, pet.object.PosVec, pet.object.UpdateDataPixie().Deadline, noxServer.Frame())
		}
	})
}

// Invoke the loaded stock ExitCollide handler, then let its normal cooperative
// save, map-load and owner/position restoration run. This is not SwitchMap or
// a replacement persistence implementation.
func (sc *e2eScenario) EnterPetTransitionExit(mapID, name string) {
	sc.enterPetTransitionExit(mapID, name, false)
}

// Put the player in contact with the loaded stock exit, then queue ordinary
// collision work. Pets, exit flags and migration/save state are not modified.
func (sc *e2eScenario) ContactPetTransitionExit(mapID, name string) {
	sc.enterPetTransitionExit(mapID, name, true)
}

func (sc *e2eScenario) enterPetTransitionExit(mapID, name string, contact bool) {
	sc.add(0, name, func() {
		host := noxServer.Players.HostUnit()
		exit := e2eExitWithDestination()
		if mapID != "" {
			exit = nil
			for obj := noxServer.Objs.First(); obj != nil; obj = obj.Next() {
				if obj.Xfer != legacy.Get_nox_xxx_XFerExit_4F4B90() || obj.CollideData == nil {
					continue
				}
				data := exitCollideData4DB600(obj.CObj())
				destination := string(data.MapName[:])
				if i := strings.IndexByte(destination, 0); i >= 0 {
					destination = destination[:i]
				}
				destination, _, _ = strings.Cut(destination, ":")
				e2eLog.Printf("PET EXIT CANDIDATE: object=%p destination=%q", obj, destination)
				if e2eMapBaseName(destination) == e2eMapBaseName(mapID) {
					exit = obj
					break
				}
			}
		}
		if host == nil || exit == nil || len(e2eTransitionPets.pets) != 3 {
			e2eError(fmt.Errorf("pet transition: player/exit(%q)/summons unavailable: %p/%p/%d", mapID, host, exit, len(e2eTransitionPets.pets)))
			return
		}
		data := exitCollideData4DB600(exit.CObj())
		target := string(data.MapName[:])
		if i := strings.IndexByte(target, 0); i >= 0 {
			target = target[:i]
		}
		target, _, _ = strings.Cut(target, ":")
		e2eTransitionPets.from = e2eMapBaseName(legacy.Nox_xxx_mapGetMapName_409B40())
		e2eTransitionPets.to = e2eMapBaseName(target)
		if e2eTransitionPets.to == "" || e2eTransitionPets.to == e2eTransitionPets.from {
			e2eError(fmt.Errorf("pet transition: invalid stock exit destination %q", target))
			return
		}
		if contact {
			pos := exit.PosVec
			asObjectS(host).SetPos(pos)
			host.NewPos, host.PrevPos = pos, pos
			host.VelVec, host.ForceVec, host.Pos24 = types.Pointf{}, types.Pointf{}, types.Pointf{}
			legacy.Nox_xxx_unitHasCollideOrUpdateFn_537610(host)
			e2eLog.Printf("PET TRANSITION CONTACT ARMED: %s -> %s exit=%p player=%p pos=%v frame=%d",
				e2eTransitionPets.from, e2eTransitionPets.to, exit, host, pos, noxServer.Frame())
		} else {
			server.CallObjectCollide(exit.Collide, exit, host, nil)
			if !sub_4DCC00() || !nox_xxx_gameGet_4DB1B0() {
				e2eError(fmt.Errorf("pet transition: ExitCollide did not queue cooperative migration/save: migrate=%t save=%t", sub_4DCC00(), nox_xxx_gameGet_4DB1B0()))
				return
			}
		}
		e2eLog.Printf("PET TRANSITION EXIT: %s -> %s exit=%p collision=%p pets=%d game-flags=%#x", e2eTransitionPets.from, e2eTransitionPets.to, exit, exit.Collide, len(e2eTransitionPets.pets), uint32(noxflags.GetGame()))
	})
	sc.addWhen(0, name+" wait for destination map", 2400, func() bool {
		return e2eMapBaseName(legacy.Nox_xxx_mapGetMapName_409B40()) == e2eTransitionPets.to &&
			legacy.Get_dword_5d4594_1548524() == 0 && noxServer.Players.HostUnit() != nil &&
			noxClient.ClientPlayerUnit() != nil && nox_client_isConnected()
	}, func() {
		e2eLog.Printf("PET TRANSITION MAP READY: %s frame=%d", e2eTransitionPets.to, noxServer.Frame())
	})
}

func (sc *e2eScenario) AssertTransitionSummons(name string) {
	sc.add(0, name, func() {
		host := noxServer.Players.HostUnit()
		if host == nil || len(e2eTransitionPets.pets) != 3 {
			e2eError(fmt.Errorf("pet transition: missing host/summons at assertion: %p/%d", host, len(e2eTransitionPets.pets)))
			return
		}
		owned := make(map[*server.Object]bool)
		for pet := host.FirstOwned516(); pet != nil; pet = pet.NextOwned512() {
			owned[pet] = true
		}
		world := make(map[*server.Object]bool)
		for pet := noxServer.Objs.First(); pet != nil; pet = pet.Next() {
			world[pet] = true
		}
		for _, pet := range e2eTransitionPets.pets {
			// Never dereference a potentially freed pointer before membership.
			if !owned[pet.object] || !world[pet.object] {
				e2eError(fmt.Errorf("pet transition: %s vanished after %s -> %s: object=%p owned=%t world=%t owned-count=%d",
					pet.typeID, e2eTransitionPets.from, e2eTransitionPets.to, pet.object, owned[pet.object], world[pet.object], len(owned)))
				return
			}
			unit := pet.object
			if unit.HealthData == nil {
				e2eError(fmt.Errorf("pet transition: %s lost health data after %s -> %s", pet.typeID, e2eTransitionPets.from, e2eTransitionPets.to))
				return
			}
			dx, dy := float64(unit.PosVec.X-host.PosVec.X), float64(unit.PosVec.Y-host.PosVec.Y)
			distance := math.Hypot(dx, dy)
			drawable := noxClient.Objs.ByNetCode(pet.wire)
			if unit.ObjOwner != host || unit.HealthData.Cur != pet.health ||
				unit.Flags().HasAny(object.FlagDead|object.FlagDestroyed) || distance > 80 || drawable == nil {
				e2eError(fmt.Errorf("pet transition: %s did not restore beside player: owner=%p/%p health=%d/%d distance=%.3f drawable=%p flags=%#x",
					pet.typeID, unit.ObjOwner, host, unit.HealthData.Cur, pet.health, distance, drawable, uint32(unit.ObjFlags)))
				return
			}
			clientDistance := math.Hypot(float64(drawable.PosVec.X)-float64(unit.PosVec.X), float64(drawable.PosVec.Y)-float64(unit.PosVec.Y))
			if clientDistance > 10 || !drawable.ObjFlags.Has(object.FlagEnabled) || drawable.ObjFlags.HasAny(object.FlagDead|object.FlagDestroyed) {
				e2eError(fmt.Errorf("pet transition: %s client disappeared/stale after %s -> %s: server=%v client=%v delta=%.3f flags=%#x",
					pet.typeID, e2eTransitionPets.from, e2eTransitionPets.to, unit.PosVec, drawable.PosVec, clientDistance, uint32(drawable.ObjFlags)))
				return
			}
			e2eLog.Printf("PET TRANSITION PRESERVED: type=%s object=%p owner=%p health=%d wire=%#x distance=%.3f drawable=%p server-pos=%v client-pos=%v client-delta=%.3f client-flags=%#x", pet.typeID, unit, unit.ObjOwner, unit.HealthData.Cur, pet.wire, distance, drawable, unit.PosVec, drawable.PosVec, clientDistance, uint32(drawable.ObjFlags))
		}
		missiles := make(map[*server.Object]bool)
		for obj := noxServer.Objs.MissileList; obj != nil; obj = obj.Next() {
			missiles[obj] = true
		}
		for _, pet := range e2eTransitionPets.pixies {
			if !owned[pet.object] || !missiles[pet.object] {
				e2eError(fmt.Errorf("pet transition: Pixie vanished: object=%p owned=%t missile=%t", pet.object, owned[pet.object], missiles[pet.object]))
				return
			}
			obj := pet.object
			if obj.UpdateData == nil {
				e2eError(fmt.Errorf("pet transition: Pixie lost update data after %s -> %s", e2eTransitionPets.from, e2eTransitionPets.to))
				return
			}
			distance := math.Hypot(float64(obj.PosVec.X-host.PosVec.X), float64(obj.PosVec.Y-host.PosVec.Y))
			drawable := noxClient.Objs.ByNetCode(pet.wire)
			if obj.ObjOwner != host || obj.UpdateDataPixie().Owner != host || distance > 150 || drawable == nil ||
				obj.Flags().HasAny(object.FlagDead|object.FlagDestroyed) {
				e2eError(fmt.Errorf("pet transition: Pixie did not restore beside player: object=%p owner=%p/%p pos=%v player=%v distance=%.3f drawable=%p deadline=%d frame=%d", obj, obj.ObjOwner, obj.UpdateDataPixie().Owner, obj.PosVec, host.PosVec, distance, drawable, obj.UpdateDataPixie().Deadline, noxServer.Frame()))
				return
			}
			clientDistance := math.Hypot(float64(drawable.PosVec.X)-float64(obj.PosVec.X), float64(drawable.PosVec.Y)-float64(obj.PosVec.Y))
			if clientDistance > 10 || !drawable.ObjFlags.Has(object.FlagEnabled) || drawable.ObjFlags.HasAny(object.FlagDead|object.FlagDestroyed) {
				e2eError(fmt.Errorf("pet transition: Pixie client disappeared/stale after %s -> %s: server=%v client=%v delta=%.3f flags=%#x",
					e2eTransitionPets.from, e2eTransitionPets.to, obj.PosVec, drawable.PosVec, clientDistance, uint32(drawable.ObjFlags)))
				return
			}
			e2eLog.Printf("PET PIXIE PRESERVED: object=%p owner=%p wire=%#x distance=%.3f server-pos=%v client-pos=%v client-delta=%.3f client-flags=%#x deadline=%d frame=%d", obj, obj.ObjOwner, pet.wire, distance, obj.PosVec, drawable.PosVec, clientDistance, uint32(drawable.ObjFlags), obj.UpdateDataPixie().Deadline, noxServer.Frame())
		}
	})
}
