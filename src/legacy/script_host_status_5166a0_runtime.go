package legacy

import (
	"github.com/opennox/opennox/v1/server"
	"github.com/opennox/opennox/v1/server/noxscript"
)

func scriptHostStatusNative5166A0(
	hostPlayer func() *server.Player,
	loadStateNonzero func(*server.PlayerUpdateData) bool,
	push func(int32),
) int32 {
	return scriptHostStatus5166A0(scriptHostStatusDeps5166A0[
		*server.Player,
		*server.Object,
		*server.PlayerUpdateData,
	]{
		hostPlayer:  hostPlayer,
		playerIsNil: func(player *server.Player) bool { return player == nil },
		loadUnit:    func(player *server.Player) *server.Object { return player.PlayerUnit },
		loadUpdate: func(unit *server.Object) *server.PlayerUpdateData {
			return (*server.PlayerUpdateData)(unit.UpdateData)
		},
		loadStateNonzero: loadStateNonzero,
		push:             push,
	})
}

func scriptHostStatusBoolNative5166A0(
	s *server.Server,
	loadStateNonzero func(*server.PlayerUpdateData) bool,
) bool {
	value := int32(0)
	scriptHostStatusNative5166A0(
		func() *server.Player { return s.Players.ByInd(server.HostPlayerIndex) },
		loadStateNonzero,
		func(got int32) { value = got },
	)
	return value != 0
}

// NoxScriptIsTalkingNative5166A0 reports the original host-player DialogWith
// state through native-width Player, Object, and PlayerUpdateData pointers.
func NoxScriptIsTalkingNative5166A0(s *server.Server) bool {
	return scriptHostStatusBoolNative5166A0(s,
		func(update *server.PlayerUpdateData) bool { return update.DialogWith != nil },
	)
}

// NoxScriptPlayerIsTradingNative5166E0 reports the original host-player
// Trade70 state through native-width Player, Object, and PlayerUpdateData
// pointers.
func NoxScriptPlayerIsTradingNative5166E0(s *server.Server) bool {
	return scriptHostStatusBoolNative5166A0(s,
		func(update *server.PlayerUpdateData) bool { return update.Trade70 != nil },
	)
}

func noxScriptIsTalkingBuiltin5166A0(vm noxscript.VM) int {
	s := GetServer().S()
	return int(scriptHostStatusNative5166A0(
		func() *server.Player { return s.Players.ByInd(server.HostPlayerIndex) },
		func(update *server.PlayerUpdateData) bool { return update.DialogWith != nil },
		vm.PushI32,
	))
}

func noxScriptPlayerIsTradingBuiltin5166E0(vm noxscript.VM) int {
	s := GetServer().S()
	return int(scriptHostStatusNative5166A0(
		func() *server.Player { return s.Players.ByInd(server.HostPlayerIndex) },
		func(update *server.PlayerUpdateData) bool { return update.Trade70 != nil },
		vm.PushI32,
	))
}
