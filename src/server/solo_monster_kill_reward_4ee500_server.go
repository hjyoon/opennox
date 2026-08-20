package server

import (
	"github.com/opennox/libs/strman"

	noxflags "github.com/opennox/opennox/v1/common/flags"
)

// SoloMonsterKillRewardRuntime4EE500 supplies dependencies whose own native
// ports are separate audit units. GiveXP is GAME.EXE 004EF270; the line-message
// callback preserves 004D9EB0's variadic formatting boundary.
type SoloMonsterKillRewardRuntime4EE500 struct {
	GiveXP          func(*Object, float32) float64
	SendLineMessage func(*Object, string, uint32)
}

type soloMonsterKillRewardNativeDeps4EE500 struct {
	gameFlag        func(uint32) int32
	findParent      func(*Object) *Object
	isMonitored     func(*Object, *Object) bool
	giveXP          func(*Object, float32) float64
	loadString      func(string, string, int) string
	sendLineMessage func(*Object, string, uint32)
}

func soloMonsterKillRewardNative4EE500(
	killed *Object,
	deps soloMonsterKillRewardNativeDeps4EE500,
) {
	soloMonsterKillReward4EE500(killed, soloMonsterKillRewardHooks4EE500[*Object, string]{
		gameFlag: deps.gameFlag,
		loadAttribution: func(killed *Object) *Object {
			return killed.Obj130
		},
		findParent: deps.findParent,
		loadClassLow: func(obj *Object) uint8 {
			return uint8(obj.ObjClass)
		},
		isMonitored: func(player, monster *Object) int32 {
			if deps.isMonitored(player, monster) {
				return 1
			}
			return 0
		},
		loadOwner: func(obj *Object) *Object {
			return obj.ObjOwner
		},
		loadExperience: func(killed *Object) float32 {
			return killed.Experience
		},
		giveXP:          deps.giveXP,
		loadString:      deps.loadString,
		sendLineMessage: deps.sendLineMessage,
	})
}

func soloMonsterKillRewardServerDeps4EE500(
	s *Server,
	runtime SoloMonsterKillRewardRuntime4EE500,
) soloMonsterKillRewardNativeDeps4EE500 {
	return soloMonsterKillRewardNativeDeps4EE500{
		gameFlag: func(flag uint32) int32 {
			if noxflags.HasGame(noxflags.GameFlag(flag)) {
				return 1
			}
			return 0
		},
		findParent:  (*Object).FindOwnerChainPlayer,
		isMonitored: Nox_xxx_creatureIsMonitored_500CC0,
		giveXP:      runtime.GiveXP,
		loadString: func(key, path string, line int) string {
			_ = line // retained by the generic provenance contract
			return s.Strings().GetStringInFile(strman.ID(key), path)
		},
		sendLineMessage: runtime.SendLineMessage,
	}
}

// SoloMonsterKillReward4EE500 resolves all Object links and fields through the
// native pointer-width layout. XP mutation and variadic message delivery remain
// explicit runtime dependencies until their separately addressed ports land.
func (s *Server) SoloMonsterKillReward4EE500(
	killed *Object,
	runtime SoloMonsterKillRewardRuntime4EE500,
) {
	soloMonsterKillRewardNative4EE500(
		killed,
		soloMonsterKillRewardServerDeps4EE500(s, runtime),
	)
}
