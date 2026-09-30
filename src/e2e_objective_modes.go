package opennox

import (
	"encoding/binary"
	"fmt"
	"image"
	"math"
	"strings"
	"unsafe"

	"github.com/opennox/libs/client/seat"
	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/internal/netlist"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

func (sc *e2eScenario) CreateObjectiveTeams(name string) {
	sc.add(0, name, func() {
		if noxServer.Teams.Count() != 0 {
			e2eError(fmt.Errorf("objective team creation requires an empty roster"))
			return
		}
		legacy.Nox_xxx_wndGuiTeamCreate_4185B0()
		if noxServer.Teams.Count() != 2 {
			e2eError(fmt.Errorf("objective team count = %d, want 2", noxServer.Teams.Count()))
			return
		}
		for team := noxServer.Teams.First(); team != nil; team = noxServer.Teams.Next(team) {
			flag := (*server.Object)(team.Field_72)
			if flag == nil || !flag.Class().Has(object.ClassFlag) || flag.TeamVal.ID != team.ID() {
				e2eError(fmt.Errorf("team %d has no valid native flag: %p", team.ID(), flag))
				return
			}
			e2eLog.Printf("OBJECTIVE TEAM CREATED: team=%p id=%d flag=%p color=%d frame=%d", team, team.ID(), flag, team.ColorInd, noxServer.Frame())
		}
	})
}

func (sc *e2eScenario) ResetObjectiveTeams(name string) {
	sc.add(0, name, func() {
		noxServer.TeamsRemoveActive(true)
		if noxServer.Teams.Count() != 0 || noxflags.HasGamePlay(noxflags.GameplayFlag4) {
			e2eError(fmt.Errorf("objective teams were not removed"))
			return
		}
		sc.objective = e2eObjectiveFixture{}
		e2eLog.Printf("OBJECTIVE TEAMS RESET: frame=%d", noxServer.Frame())
	})
}

// JoinObjectiveTeam sends the ordinary team-join request through the host's
// network queue. Solo test hosts initially remain teamless; assigning the
// TeamVal field directly would miss team links, notifications, and masks.
func (sc *e2eScenario) JoinObjectiveTeam(name string) {
	checks := 0
	sc.add(0, name, func() {
		player := noxServer.Players.Host()
		team := noxServer.Teams.First()
		if player == nil || player.PlayerUnit == nil || team == nil || player.PlayerUnit.TeamVal.Has() {
			e2eError(fmt.Errorf("invalid team-join fixture: player=%p team=%p", player, team))
			return
		}
		sc.objective.team = team
		var packet [10]byte
		packet[0] = byte(netmsg.MSG_TEAM_MSG)
		packet[1] = 10
		binary.LittleEndian.PutUint32(packet[2:], uint32(team.ID()))
		binary.LittleEndian.PutUint16(packet[6:], uint16(player.NetCode()))
		if !noxServer.NetList.AddToMsgListCli(server.HostPlayerIndex, netlist.Kind0, packet[:]) {
			e2eError(fmt.Errorf("cannot queue objective team-join request"))
			return
		}
		e2eLog.Printf("OBJECTIVE TEAM JOIN REQUEST: player=%p unit=%p netcode=%d team=%d frame=%d", player, player.PlayerUnit, player.NetCode(), team.ID(), noxServer.Frame())
	})
	sc.addWhen(1, name+" applied", 120, func() bool {
		player := noxServer.Players.Host()
		if player == nil || player.PlayerUnit == nil || sc.objective.team == nil {
			return false
		}
		drawable := noxClient.Objs.ByNetCodeDynamic(player.NetCode())
		// The stock attachment packet uses recipient 159 (all except the
		// local host). Its drawable is not an independently replicated team
		// member; host gameplay observes the server object's live team.
		ready := noxServer.Teams.ContainsObject(player.PlayerUnit.TeamPtr(), sc.objective.team.ID()) &&
			drawable != nil
		checks++
		if !ready && (checks == 1 || checks%30 == 0) {
			var clientTeam server.TeamID
			if drawable != nil {
				clientTeam = drawable.TeamPtr().ID
			}
			e2eLog.Printf("OBJECTIVE TEAM JOIN WAIT: checks=%d server_team=%d client_team=%d drawable=%p pending=%d flags=%v frame=%d", checks, player.PlayerUnit.TeamVal.ID, clientTeam, drawable, noxServer.NetList.ByInd(server.HostPlayerIndex, netlist.Kind0).Count(), noxflags.GetGame(), noxServer.Frame())
		}
		return ready
	}, func() {
		e2eLog.Printf("OBJECTIVE TEAM JOINED: team=%d unit=%p frame=%d", sc.objective.team.ID(), noxServer.Players.HostUnit(), noxServer.Frame())
	})
}

type e2eObjectiveFixture struct {
	ball        *server.Object
	goal        *server.Object
	team        *server.Team
	playerScore int32
	teamScore   int32
	crown       *server.Object
}

func e2eFindObjective(typeID string) (*server.Object, error) {
	typ := noxServer.Types.ByID(typeID)
	if typ == nil {
		return nil, fmt.Errorf("unknown objective type %q", typeID)
	}
	var found *server.Object
	for obj := noxServer.Objs.First(); obj != nil; obj = obj.Next() {
		if int(obj.TypeInd) != typ.Ind() || obj.Flags().HasAny(object.FlagDead|object.FlagDestroyed) {
			continue
		}
		if found != nil {
			return nil, fmt.Errorf("more than one live %s objective", typeID)
		}
		found = obj
	}
	if found == nil || found.UpdateData == nil {
		return nil, fmt.Errorf("no live %s objective with update data", typeID)
	}
	return found, nil
}

func e2eObjectiveGameMode(mode string) (noxflags.GameFlag, error) {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "ctf":
		return noxflags.GameModeCTF, nil
	case "flagball":
		return noxflags.GameModeFlagBall, nil
	case "kotr":
		return noxflags.GameModeKOTR, nil
	default:
		return 0, fmt.Errorf("unknown E2E objective game mode %q", mode)
	}
}

// SetObjectiveGameMode changes the same mode preference used by server
// options. The subsequent ordinary map switch still performs map validation,
// objective setup, player respawn, and client map loading.
func (sc *e2eScenario) SetObjectiveGameMode(mode, name string) {
	sc.add(0, name, func() {
		flag, err := e2eObjectiveGameMode(mode)
		if err != nil {
			e2eError(err)
			return
		}
		settings := getCurrentSettings2()
		settings.Field52 = (settings.Field52 &^ uint16(noxflags.GameModeMask)) | uint16(flag)
		e2eLog.Printf("OBJECTIVE MODE PREFERENCE: mode=%s flags=%#x", flag, settings.Field52)
	})
}

func (sc *e2eScenario) AssertObjectiveGameMode(mode, name string) {
	sc.add(0, name, func() {
		flag, err := e2eObjectiveGameMode(mode)
		if err != nil {
			e2eError(err)
			return
		}
		if got := noxflags.GetGame() & noxflags.GameModeMask; got != flag {
			e2eError(fmt.Errorf("objective mode = %s, want %s", got, flag))
			return
		}
		player := noxServer.Players.Host()
		if player == nil || player.PlayerUnit == nil {
			e2eError(fmt.Errorf("objective mode has no live host player"))
			return
		}
		var objectiveType uint16
		switch flag {
		case noxflags.GameModeFlagBall:
			objectiveType = uint16(noxServer.Types.IndByID("GameBall"))
		case noxflags.GameModeKOTR:
			objectiveType = uint16(noxServer.Types.IndByID("Crown"))
		}
		count := 0
		for obj := noxServer.Objs.First(); obj != nil; obj = obj.Next() {
			if obj.Flags().HasAny(object.FlagDead | object.FlagDestroyed) {
				continue
			}
			if flag == noxflags.GameModeCTF {
				if !obj.Class().Has(object.ClassFlag) {
					continue
				}
			} else if obj.TypeInd != objectiveType {
				continue
			}
			count++
			if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(obj.CObj()) <= math.MaxUint32 {
				e2eError(fmt.Errorf("objective pointer %p is not above 4 GiB", obj))
				return
			}
			if flag != noxflags.GameModeCTF && !player.MinimapTracks(obj) {
				e2eError(fmt.Errorf("host minimap does not track %s objective %p", flag, obj))
				return
			}
			if flag == noxflags.GameModeFlagBall {
				if obj.UpdateData == nil {
					e2eError(fmt.Errorf("FlagBall has no native update data"))
					return
				}
				update := (*server.GameBallUpdateData4EA800)(obj.UpdateData)
				if update.PossessionDuration == 0 || update.ResetVelocity <= 0 {
					e2eError(fmt.Errorf("FlagBall setup parameters are missing: duration=%d velocity=%g", update.PossessionDuration, update.ResetVelocity))
					return
				}
			}
			e2eLog.Printf("OBJECTIVE OBJECT: mode=%s obj=%p update=%p team=%d minimap=%t pos=(%.3f,%.3f)",
				flag, obj, obj.UpdateData, obj.TeamVal.ID, player.MinimapTracks(obj), obj.PosVec.X, obj.PosVec.Y)
		}
		if count == 0 || (flag == noxflags.GameModeCTF && count < 2) || (flag == noxflags.GameModeFlagBall && count != 1) {
			e2eError(fmt.Errorf("objective mode %s has unexpected objective count %d", flag, count))
			return
		}
		e2eLog.Printf("OBJECTIVE MODE READY: mode=%s map=%q frame=%d objectives=%d player=%p",
			flag, legacy.Nox_xxx_mapGetMapName_409B40(), noxServer.Frame(), count, player.PlayerUnit)
	})
}

func (sc *e2eScenario) ResetFlagball(name string) {
	sc.add(0, name, func() {
		if !noxflags.HasGame(noxflags.GameModeFlagBall) {
			e2eError(fmt.Errorf("cannot reset FlagBall outside FlagBall mode"))
			return
		}
		ballType := uint16(noxServer.Types.IndByID("GameBall"))
		for obj := noxServer.Objs.First(); obj != nil; obj = obj.Next() {
			if obj.TypeInd != ballType || obj.Flags().HasAny(object.FlagDead|object.FlagDestroyed) {
				continue
			}
			if noxServer.GameBallReset417F50(obj) != 1 {
				e2eError(fmt.Errorf("FlagBall reset failed for %p", obj))
				return
			}
			e2eLog.Printf("FLAGBALL RESET: old=%p frame=%d", obj, noxServer.Frame())
			return
		}
		e2eError(fmt.Errorf("no live FlagBall to reset"))
	})
}

// ApproachFlagball positions the host at the stock ball without calling its
// collision handler or modifying ownership. Ordinary server physics must pick
// it up during the following ticks.
func (sc *e2eScenario) ApproachFlagball(name string) {
	sc.add(0, name, func() {
		if !noxflags.HasGame(noxflags.GameModeFlagBall) {
			e2eError(fmt.Errorf("cannot approach FlagBall outside FlagBall mode"))
			return
		}
		ball, err := e2eFindObjective("GameBall")
		if err != nil {
			e2eError(err)
			return
		}
		player := noxServer.Players.HostUnit()
		if player == nil || noxServer.Teams.ByID(player.TeamVal.ID) == nil || ball.ObjOwner != nil {
			e2eError(fmt.Errorf("invalid FlagBall pickup fixture: player=%p ball=%p owner=%p", player, ball, ball.ObjOwner))
			return
		}
		sc.objective.ball = ball
		asObjectS(player).SetPos(ball.PosVec)
		e2eLog.Printf("FLAGBALL APPROACH: ball=%p player=%p team=%d pos=%v frame=%d", ball, player, player.TeamVal.ID, ball.PosVec, noxServer.Frame())
	})
}

func (sc *e2eScenario) AssertFlagballCarried(name string) {
	sc.addWhen(0, name, 120, func() bool {
		ball := sc.objective.ball
		player := noxServer.Players.HostUnit()
		return ball != nil && player != nil && ball.ObjOwner == player &&
			(*server.GameBallUpdateData4EA800)(ball.UpdateData).Carrier == player
	}, func() {
		ball := sc.objective.ball
		player := noxServer.Players.HostUnit()
		update := (*server.GameBallUpdateData4EA800)(ball.UpdateData)
		owned := false
		for obj := player.FirstOwned516(); obj != nil; obj = obj.NextOwned512() {
			if obj == ball {
				owned = true
				break
			}
		}
		if !owned || !ball.Flags().Has(object.FlagNoCollide) || ball.TeamVal.ID != player.TeamVal.ID ||
			update.TeamID != uint32(player.TeamVal.ID) || update.CarrierFrame == 0 || ball.InvHolder != nil {
			e2eError(fmt.Errorf("invalid carried FlagBall: owner=%p carrier=%p team=%d carrier_team=%d owned=%t flags=%v holder=%p", ball.ObjOwner, update.Carrier, ball.TeamVal.ID, update.TeamID, owned, ball.Flags(), ball.InvHolder))
			return
		}
		e2eLog.Printf("FLAGBALL CARRIED: ball=%p owner=%p carrier=%p update=%p team=%d carrier_frame=%d frame=%d", ball, ball.ObjOwner, update.Carrier, update, ball.TeamVal.ID, update.CarrierFrame, noxServer.Frame())
	})
}

func (sc *e2eScenario) ApproachFlagballGoal(name string) {
	sc.add(0, name, func() {
		player := noxServer.Players.Host()
		ball := sc.objective.ball
		if player == nil || player.PlayerUnit == nil || ball == nil || ball.ObjOwner != player.PlayerUnit {
			e2eError(fmt.Errorf("cannot score without a carried FlagBall"))
			return
		}
		team := noxServer.Teams.ByID(player.PlayerUnit.TeamVal.ID)
		if team == nil {
			e2eError(fmt.Errorf("FlagBall carrier has no live team"))
			return
		}
		var goal *server.Object
		for obj := noxServer.Objs.First(); obj != nil; obj = obj.Next() {
			if obj.Class().Has(object.ClassFlag) && obj.TeamVal.ID != team.ID() &&
				noxServer.Teams.ByID(obj.TeamVal.ID) != nil && !obj.Flags().HasAny(object.FlagDead|object.FlagDestroyed) {
				goal = obj
				break
			}
		}
		if goal == nil {
			e2eError(fmt.Errorf("FlagBall map has no opposing goal for team %d", team.ID()))
			return
		}
		sc.objective.goal = goal
		sc.objective.team = team
		sc.objective.playerScore = player.Lessons
		sc.objective.teamScore = team.Lessons
		asObjectS(player.PlayerUnit).SetPos(goal.PosVec)
		e2eLog.Printf("FLAGBALL GOAL APPROACH: goal=%p goal_team=%d carrier_team=%d player_score=%d team_score=%d pos=%v frame=%d", goal, goal.TeamVal.ID, team.ID(), player.Lessons, team.Lessons, goal.PosVec, noxServer.Frame())
	})
}

func (sc *e2eScenario) AssertFlagballScored(name string) {
	sc.addWhen(0, name, 120, func() bool {
		player := noxServer.Players.Host()
		return player != nil && sc.objective.team != nil && player.Lessons == sc.objective.playerScore+1 &&
			sc.objective.team.Lessons == sc.objective.teamScore+1
	}, func() {
		ball := sc.objective.ball
		update := (*server.GameBallUpdateData4EA800)(ball.UpdateData)
		atStart := false
		startType := uint16(noxServer.Types.IndByID("GameBallStart"))
		for obj := noxServer.Objs.First(); obj != nil; obj = obj.Next() {
			if obj.TypeInd == startType && obj.PosVec == ball.PosVec {
				atStart = true
				break
			}
		}
		if ball.ObjOwner != nil || update.Carrier != nil || !atStart ||
			ball.Flags().HasAny(object.FlagDead|object.FlagDestroyed|object.FlagNoCollide) ||
			ball.HealthData == nil || ball.HealthData.Cur != ball.HealthData.Max || ball.VelVec != (types.Pointf{}) {
			e2eError(fmt.Errorf("scored FlagBall did not respawn: ball=%p owner=%p carrier=%p start=%t flags=%v pos=%v velocity=%v", ball, ball.ObjOwner, update.Carrier, atStart, ball.Flags(), ball.PosVec, ball.VelVec))
			return
		}
		e2eLog.Printf("FLAGBALL SCORED: ball=%p goal=%p player_score=%d team_score=%d owner=%p carrier=%p pos=%v frame=%d", ball, sc.objective.goal, noxServer.Players.Host().Lessons, sc.objective.team.Lessons, ball.ObjOwner, update.Carrier, ball.PosVec, noxServer.Frame())
	})
}

// Collide with the map's existing crown through ordinary server physics, then
// reuse the real inventory/drop fixture with its unchanged stock callbacks.
func (sc *e2eScenario) ApproachStockCrown(name string) {
	sc.add(0, name, func() {
		if !noxflags.HasGame(noxflags.GameModeKOTR) {
			e2eError(fmt.Errorf("cannot approach crown outside KOTR mode"))
			return
		}
		crown, err := e2eFindObjective("Crown")
		if err != nil {
			e2eError(err)
			return
		}
		player := noxServer.Players.HostUnit()
		before, err := e2eInventoryItemCount("Crown")
		if err != nil || player == nil || crown.InvHolder != nil || crown.ObjOwner != nil || crown.Pickup.Ptr == nil || crown.Drop.Ptr == nil {
			e2eError(fmt.Errorf("invalid stock crown fixture: crown=%p player=%p inventory=%d error=%v", crown, player, before, err))
			return
		}
		pickup, pickupOK := server.ObjectPickupHandler("CrownPickup")
		drop, dropOK := server.ObjectDropHandler("CrownDrop")
		if !pickupOK || !dropOK || crown.Pickup.Ptr != pickup.Ptr || crown.Drop.Ptr != drop.Ptr {
			e2eError(fmt.Errorf("stock crown callbacks differ: pickup=%p want=%p drop=%p want=%p", crown.Pickup.Ptr, pickup.Ptr, crown.Drop.Ptr, drop.Ptr))
			return
		}
		sc.objective.crown = crown
		e2e.groundItem = crown
		e2e.groundItemTypeID = "Crown"
		e2e.groundItemPickupName = "CrownPickup"
		e2e.groundItemPickupPtr = crown.Pickup.Ptr
		e2e.groundItemDropName = "CrownDrop"
		e2e.groundItemDropPtr = crown.Drop.Ptr
		e2e.groundItemOwned = false
		e2e.groundItemBefore = before
		e2e.groundItemWireCode = uint16(noxServer.GetUnitNetCode(crown))
		e2e.groundItemDropped = nil
		e2e.groundItemDropChecks = 0
		asObjectS(player).SetPos(crown.PosVec)
		e2eLog.Printf("CROWN APPROACH: crown=%p update=%p player=%p pickup=%p drop=%p pos=%v frame=%d", crown, crown.UpdateData, player, crown.Pickup.Ptr, crown.Drop.Ptr, crown.PosVec, noxServer.Frame())
	})
}

func (sc *e2eScenario) AssertCrownState(carried bool, name string) {
	sc.add(0, name, func() {
		crown := sc.objective.crown
		player := noxServer.Players.Host()
		if crown == nil || player == nil || player.PlayerUnit == nil {
			e2eError(fmt.Errorf("crown fixture is missing"))
			return
		}
		unit := player.PlayerUnit
		update := (*server.CrownUpdateData)(crown.UpdateData)
		if carried != unit.HasEnchant(server.ENCHANT_CROWN) || carried == player.MinimapTracks(crown) ||
			(carried && (crown.ObjOwner != unit || crown.InvHolder != unit || unit.UpdateDataPlayer().Field66 == 0)) ||
			(!carried && (crown.ObjOwner != nil || crown.InvHolder != nil)) || update.PickupTarget != nil {
			e2eError(fmt.Errorf("invalid crown state: carried=%t owner=%p holder=%p enchant=%t minimap=%t pending=%p", carried, crown.ObjOwner, crown.InvHolder, unit.HasEnchant(server.ENCHANT_CROWN), player.MinimapTracks(crown), update.PickupTarget))
			return
		}
		e2eLog.Printf("CROWN STATE: crown=%p update=%p carried=%t owner=%p holder=%p enchant=%t minimap=%t pickup_frame=%d frame=%d", crown, update, carried, crown.ObjOwner, crown.InvHolder, unit.HasEnchant(server.ENCHANT_CROWN), player.MinimapTracks(crown), unit.UpdateDataPlayer().Field66, noxServer.Frame())
	})
}

// ClickObserverMode uses the visible escape-menu button. In KOTR the ordinary
// inventory drop request is rejected; entering observer mode returns the crown
// through the game's owned-object/drop workflow instead.
func (sc *e2eScenario) ClickObserverMode(name string) {
	sc.addWhen(0, name, 120, func() bool {
		button := noxClient.GUI.ChildByID(9007)
		return button != nil && !button.GetFlags().IsHidden() && button.GetFlags().IsEnabled() &&
			button.Parent() != nil && !button.Parent().GetFlags().IsHidden()
	}, func() {
		button := noxClient.GUI.ChildByID(9007)
		size := button.Size()
		pos := button.GlobalPos().Add(image.Pt(size.X/2, size.Y/2))
		e2eLog.Printf("OBSERVER MODE CLICK: button=%p point=%v frame=%d", button, pos, noxServer.Frame())
		e2eQueueInput(&seat.MouseMoveEvent{Pos: pos, Relative: false})
	})
	sc.Input(1, "", &seat.MouseButtonEvent{Button: seat.MouseButtonLeft, Pressed: true})
	sc.Input(1, "", &seat.MouseButtonEvent{Button: seat.MouseButtonLeft, Pressed: false})
}
