package opennox

import (
	"fmt"
	"math"
	"time"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

// Scalar observations only: assertions must never release a briefing, clear
// observer flags, advance the stage, repair HP or create the expected minions.
type e2eQuestPlayerState struct {
	mapName                          string
	stage                            int
	quest, connected, loading        bool
	briefing                         bool
	serverCode, clientCode, drawCode uint32
	phase                            byte
	status                           uint32
	flags                            object.Flags
	health, maxHealth                uint16
	position                         types.Pointf
}

func (s e2eQuestPlayerState) validate(stage int, mapName string) error {
	if stage <= 0 || !s.quest || !s.connected || s.loading || s.briefing || s.stage != stage ||
		s.mapName == "" || mapName != "" && e2eMapBaseName(s.mapName) != e2eMapBaseName(mapName) ||
		s.serverCode == 0 || s.clientCode != s.serverCode || s.drawCode != s.serverCode ||
		s.phase != 3 || s.status&1 != 0 || s.flags.HasAny(object.FlagDead|object.FlagDestroyed) ||
		s.health == 0 || s.maxHealth == 0 || s.health > s.maxHealth ||
		math.IsNaN(float64(s.position.X)) || math.IsInf(float64(s.position.X), 0) ||
		math.IsNaN(float64(s.position.Y)) || math.IsInf(float64(s.position.Y), 0) {
		return fmt.Errorf("Quest player is not playable: want stage=%d map=%q, got %+v", stage, mapName, s)
	}
	return nil
}

func e2eReadQuestPlayer() (e2eQuestPlayerState, error) {
	host, drawable := noxServer.Players.HostUnit(), noxClient.ClientPlayerUnit()
	if host == nil || drawable == nil || !host.Class().Has(object.ClassPlayer) ||
		host.UpdateData == nil || host.HealthData == nil || host.ControllingPlayer() == nil {
		return e2eQuestPlayerState{}, fmt.Errorf("Quest native player/drawable is missing: player=%p drawable=%p", host, drawable)
	}
	player := host.ControllingPlayer()
	return e2eQuestPlayerState{
		mapName: legacy.Nox_xxx_mapGetMapName_409B40(), stage: noxServer.nox_game_getQuestStage_4E3CC0(),
		quest: noxflags.HasGame(noxflags.GameModeQuest), connected: nox_client_isConnected(),
		loading:    legacy.Get_dword_5d4594_1548524() != 0,
		briefing:   sub_450560() || noxClient.GUI.Captured() != nil,
		serverCode: host.NetCode, clientCode: uint32(legacy.ClientPlayerNetCode()), drawCode: drawable.NetCode32,
		phase: player.Field3676, status: player.Field3680, flags: host.Flags(),
		health: host.HealthData.Cur, maxHealth: host.HealthData.Max, position: host.PosVec,
	}, nil
}

func (sc *e2eScenario) AssertQuestStage(stage int, mapName, name string) {
	sc.add(0, name, func() {
		state, err := e2eReadQuestPlayer()
		if err == nil {
			err = state.validate(stage, mapName)
		}
		if err != nil {
			e2eError(err)
			return
		}
		e2eLog.Printf("QUEST STAGE PLAYABLE: map=%q stage=%d frame=%d wire=%d phase=%d status=%#x observer=false briefing=false health=%d/%d pos=%v",
			state.mapName, state.stage, noxServer.Frame(), state.serverCode, state.phase, state.status, state.health, state.maxHealth, state.position)
	})
}

func e2eQuestMovementDistance(before, after e2eQuestPlayerState) (float64, error) {
	if err := before.validate(before.stage, before.mapName); err != nil {
		return 0, err
	}
	if err := after.validate(before.stage, before.mapName); err != nil {
		return 0, err
	}
	distance := math.Hypot(float64(after.position.X)-float64(before.position.X), float64(after.position.Y)-float64(before.position.Y))
	if before.serverCode != after.serverCode || distance < 8 {
		return distance, fmt.Errorf("Quest real input did not move the same player: before=%+v after=%+v distance=%.3f", before, after, distance)
	}
	return distance, nil
}

func (sc *e2eScenario) WalkQuest(ang float64, dt time.Duration, name string) {
	var before e2eQuestPlayerState
	sc.add(0, name+" capture player", func() {
		var err error
		before, err = e2eReadQuestPlayer()
		if err == nil {
			err = before.validate(before.stage, before.mapName)
		}
		if err != nil {
			e2eError(err)
		}
	})
	// Ordinary mouse input and game ticks, not a SetPos movement fixture.
	sc.WalkFor(ang, dt, name)
	sc.add(0, name+" verify movement", func() {
		after, err := e2eReadQuestPlayer()
		var distance float64
		if err == nil {
			distance, err = e2eQuestMovementDistance(before, after)
		}
		if err != nil {
			e2eError(err)
			return
		}
		e2eLog.Printf("QUEST REAL MOVEMENT: map=%q stage=%d wire=%d from=%v to=%v distance=%.3f health=%d/%d",
			after.mapName, after.stage, after.serverCode, before.position, after.position, distance, after.health, after.maxHealth)
	})
}

func (sc *e2eScenario) AssertQuestMinion(typeID string, minimum int, name string) {
	sc.add(0, name, func() {
		if !noxflags.HasGame(noxflags.GameModeQuest) || noxServer.nox_game_getQuestStage_4E3CC0() < 5 ||
			minimum < 1 || typeID != "Hecubah" && typeID != "Necromancer" {
			e2eError(fmt.Errorf("Quest minion assertion has invalid stage/type/count: type=%q minimum=%d", typeID, minimum))
			return
		}
		count := 0
		for obj := noxServer.Objs.First(); obj != nil; obj = obj.Next() {
			if obj.ObjectTypeC().ID() != typeID {
				continue
			}
			if !obj.Class().Has(object.ClassMonster) || obj.UpdateData == nil || obj.HealthData == nil ||
				obj.HealthData.Cur == 0 || obj.HealthData.Max == 0 || obj.Flags().HasAny(object.FlagDead|object.FlagDestroyed) {
				e2eError(fmt.Errorf("Quest minion is not alive: type=%s object=%p health=%+v update=%p flags=%#x", typeID, obj, obj.HealthData, obj.UpdateData, uint32(obj.Flags())))
				return
			}
			update := obj.UpdateDataMonster()
			if update.MonsterDef == nil || obj.ObjOwner != nil {
				e2eError(fmt.Errorf("Quest minion lost native definition/ownership: type=%s object=%p definition=%p owner=%p", typeID, obj, update.MonsterDef, obj.ObjOwner))
				return
			}
			inventory := make(map[*server.Object]bool)
			for item := obj.FirstItem(); item != nil; item = item.NextItem() {
				if inventory[item] || item.InvHolder != obj || item.Flags().Has(object.FlagDestroyed) {
					e2eError(fmt.Errorf("Quest minion inventory is broken: type=%s object=%p item=%p holder=%p duplicate=%t", typeID, obj, item, item.InvHolder, inventory[item]))
					return
				}
				inventory[item] = true
			}
			count++
			// RewardActivate may legitimately return nil. Log inventory rather
			// than inventing a guaranteed drop count; unit tests check attempts.
			e2eLog.Printf("QUEST MINION ALIVE: type=%s object=%p update=%p definition=%p health=%d/%d power=%d aggression=%.6f skill=%.6f inventory=%d pos=%v",
				typeID, obj, obj.UpdateData, update.MonsterDef, obj.HealthData.Cur, obj.HealthData.Max, update.Field510, update.Aggression, update.Field330, len(inventory), obj.PosVec)
		}
		if count < minimum {
			e2eError(fmt.Errorf("Quest minion is missing: type=%s count=%d minimum=%d", typeID, count, minimum))
			return
		}
		e2eLog.Printf("QUEST MINION VERIFIED: type=%s count=%d minimum=%d stage=%d", typeID, count, minimum, noxServer.nox_game_getQuestStage_4E3CC0())
	})
}

// Put the player at a stock Quest exit and queue ordinary collision work. The
// fixture does not set exit/observer/next-map flags or invoke SwitchMap: the
// loaded exit callback and server/client ticks must perform the transition.
func (sc *e2eScenario) EnterQuestExit(name string) {
	var beforeStage int
	sc.addWhen(0, name+" locate stock exit", 1200, func() bool {
		return noxflags.HasGame(noxflags.GameModeQuest) &&
			legacy.Get_dword_5d4594_1548524() == 0 &&
			noxServer.Players.HostUnit() != nil && noxClient.ClientPlayerUnit() != nil && nox_client_isConnected()
	}, func() {
		host := noxServer.Players.HostUnit()
		var exit *server.Object
		for obj := noxServer.Objs.First(); obj != nil; obj = obj.Next() {
			if !obj.Class().Has(object.ClassExit) || obj.SubClass()&1 == 0 || obj.CollideData == nil ||
				obj.Flags().HasAny(object.FlagDead|object.FlagDestroyed) {
				continue
			}
			e2eLog.Printf("QUEST EXIT CANDIDATE: object=%p type=%s id=%q pos=%v flags=%#x callback=%p",
				obj, obj.ObjectTypeC().ID(), obj.ID(), obj.PosVec, uint32(obj.Flags()), obj.Collide)
			if exit == nil {
				exit = obj
			}
		}
		if exit == nil || exit.Collide == nil {
			e2eError(fmt.Errorf("Quest map %q has no stock stage exit", legacy.Nox_xxx_mapGetMapName_409B40()))
			return
		}
		beforeStage = noxServer.nox_game_getQuestStage_4E3CC0()
		pos := exit.PosVec
		asObjectS(host).SetPos(pos)
		host.NewPos = pos
		host.PrevPos = pos
		host.VelVec = types.Pointf{}
		host.ForceVec = types.Pointf{}
		host.Pos24 = types.Pointf{}
		legacy.Nox_xxx_unitHasCollideOrUpdateFn_537610(host)
		e2eLog.Printf("QUEST EXIT CONTACT ARMED: map=%q stage=%d frame=%d exit=%p player=%p pos=%v",
			legacy.Nox_xxx_mapGetMapName_409B40(), beforeStage, noxServer.Frame(), exit, host, pos)
	})
	sc.addWhen(0, name+" destination ready", 2400, func() bool {
		return noxServer.nox_game_getQuestStage_4E3CC0() == beforeStage+1 &&
			legacy.Get_dword_5d4594_1548524() == 0 &&
			noxServer.Players.HostUnit() != nil && noxClient.ClientPlayerUnit() != nil && nox_client_isConnected()
	}, func() {
		host := noxServer.Players.HostUnit()
		e2eLog.Printf("QUEST NEXT STAGE READY: map=%q stage=%d frame=%d player=%p drawable=%p pos=%v health=%d/%d",
			legacy.Nox_xxx_mapGetMapName_409B40(), noxServer.nox_game_getQuestStage_4E3CC0(), noxServer.Frame(),
			host, noxClient.ClientPlayerUnit(), host.PosVec, host.HealthData.Cur, host.HealthData.Max)
	})
}
