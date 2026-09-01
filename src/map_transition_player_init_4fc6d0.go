package opennox

import (
	"path/filepath"

	"github.com/opennox/libs/common"
	"github.com/opennox/libs/datapath"
	"github.com/opennox/libs/ifs"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/server"
)

func (s *Server) mapTransitionPlayerInit4FC6D0() {
	s.Server.MapTransitionPlayerInit4FC6D0(server.MapTransitionPlayerInitRuntime4FC6D0{
		GameFlag: func(mask uint32) int32 {
			return int32(bool2int(noxflags.HasGame(noxflags.GameFlag(mask))))
		},

		QuestStage: func() int32 {
			return int32(s.nox_game_getQuestStage_4E3CC0())
		},
		RestorePredicate: func() int32 {
			return int32(sub_4D6F30())
		},
		RestoreReady: func() int32 {
			return int32(legacy.Sub_4D7430())
		},
		QueuedRestore: func() int32 {
			return int32(legacy.Sub_4D76F0())
		},
		SendQuestStage: func(index uint8) {
			legacy.Nox_game_sendQuestStage_4D6960(ntype.PlayerInd(index))
		},
		SendQuestRestore: func(index uint8, value int32) {
			legacy.Sub_4D6880(int(index), int(value))
		},
		StoreQueuedRestore: func(value int32) {
			legacy.Sub_4D76E0(int(value))
		},
		MarkQuestReady: func(value int32) {
			legacy.Sub_4D7440(int(value))
		},
		FinishQuestTransition: legacy.Sub_4D60B0,
		FadeBegin: func(out, menu int32) {
			s.Nox_xxx_netMsgFadeBegin_4D9800(out != 0, menu != 0)
		},

		DataRoot: func() string {
			return datapath.Data()
		},
		FormatTempSavePath: func(root string) string {
			return filepath.Join(root, common.SaveDir, "_temp_.dat")
		},
		DeleteTempFile: func(path string) {
			_ = ifs.Remove(path)
		},

		SavePlayerData: func(path string, index uint8) int32 {
			return int32(bool2int(savePlayerServerData(path, ntype.PlayerInd(index))))
		},
		PreparePlayerData: func(index uint8) int32 {
			return int32(bool2int(sub_419EE0(ntype.PlayerInd(index))))
		},
		SendGauntlet: func(index uint8, value int32) {
			s.Nox_xxx_sendGauntlet_4DCF80(ntype.PlayerInd(index), byte(value))
		},
		RestorePlayerData: func(path string, index uint8) int32 {
			return int32(bool2int(sub41CFA0(path, ntype.PlayerInd(index))))
		},
		FinishPlayerData: func(index uint8) {
			legacy.Sub_4D6770(ntype.PlayerInd(index))
		},
		ApplyEnchant: func(unit *server.Object, enchant server.EnchantID, duration, power int32) {
			asObjectS(unit).ApplyEnchant(enchant, int(duration), int(power))
		},
	})
}
