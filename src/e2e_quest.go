package opennox

import (
	"fmt"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"
	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

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
