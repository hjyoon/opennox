package opennox

import (
	"fmt"
	"math"
	"unsafe"

	"github.com/opennox/libs/object"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/legacy"
)

type e2ePlayerQuestKeysReportOutcome struct {
	inventory       [2]int
	clientInventory [2]uint32
	clientFound     [2]bool
	cached, client  [2]byte
}

func (o e2ePlayerQuestKeysReportOutcome) validate(want [2]int) error {
	for kind, count := range want {
		if count < 0 {
			return fmt.Errorf("negative expected Quest key count: %v", want)
		}
		presence := byte(0)
		if count != 0 {
			presence = 1
		}
		if o.inventory[kind] != count || o.clientInventory[kind] != uint32(count) || o.clientFound[kind] != (count != 0) ||
			o.cached[kind] != presence || o.client[kind] != presence {
			return fmt.Errorf("Quest key report: kind=%d want=%d inventory=%v client_inventory=%v found=%v cache=%v receiver=%v",
				kind, count, o.inventory, o.clientInventory, o.clientFound, o.cached, o.client)
		}
	}
	return nil
}

func readE2EPlayerQuestKeysReport() (e2ePlayerQuestKeysReportOutcome, error) {
	var out e2ePlayerQuestKeysReportOutcome
	if !noxflags.HasGame(noxflags.GameModeQuest) || noxServer.Players.HostUnit() == nil ||
		noxClient.ClientPlayerUnit() == nil || !nox_client_isConnected() {
		return out, fmt.Errorf("Quest key report requires the real connected Quest host")
	}
	if active, amount, maximum := legacy.Nox_gui_itemAmountState(); active || legacy.Nox_client_inventoryHasDragged() {
		return out, fmt.Errorf("Quest key input remains unfinished: amount_dialog=%t amount=%d max=%d dragged=%t",
			active, amount, maximum, legacy.Nox_client_inventoryHasDragged())
	}
	unit := noxServer.Players.HostUnit()
	if unit.UpdateData == nil || unit.ControllingPlayer() == nil {
		return out, fmt.Errorf("Quest key report lost its native player/update binding")
	}
	player := unit.ControllingPlayer()
	code := uint16(unit.NetCode)
	if player.PlayerInd >= 32 || noxServer.Players.ByID(int(code)) != player ||
		legacy.ClientPlayerNetCode() != int(code) {
		return out, fmt.Errorf("Quest key report player identity mismatch: source=%d player=%d client=%d", code, player.NetCodeVal, legacy.ClientPlayerNetCode())
	}
	if unsafe.Sizeof(uintptr(0)) == 8 {
		for _, ptr := range []unsafe.Pointer{unsafe.Pointer(unit), unit.UpdateData, unsafe.Pointer(player)} {
			if uintptr(ptr) <= math.MaxUint32 {
				return out, fmt.Errorf("Quest key report pointer %p must exceed 4 GiB", ptr)
			}
		}
	}
	for kind, typeID := range []string{"SilverKey", "GoldKey"} {
		count, err := e2eInventoryItemCount(typeID)
		if err != nil {
			return out, err
		}
		out.inventory[kind] = count
		typ := noxServer.Types.ByID(typeID)
		out.clientFound[kind], out.clientInventory[kind], _, _ = legacy.Nox_client_inventoryItemState(uint32(typ.Ind()))
		if count != 0 {
			item, _, err := e2eInventoryItem(typeID)
			if err != nil || item.InvHolder != unit || !item.Class().Has(object.ClassKey) ||
				unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(item)) <= math.MaxUint32 {
				return out, fmt.Errorf("Quest key %q lost its high native inventory binding: %v", typeID, err)
			}
		}
	}
	update := unit.UpdateDataPlayer()
	out.cached = [2]byte{update.QuestPlayerFlagsA[player.PlayerInd], update.QuestPlayerFlagsB[player.PlayerInd]}
	out.client = legacy.Nox_client_playerQuestKeysReceived(player)
	return out, nil
}

// CheckPlayerQuestKeysReport only reads genuine inventory, self-report cache,
// and client decoder results. A normal grant, queued inventory input, or stock
// Quest exit is scheduled by the public scenario, never supplied by this check.
func (sc *e2eScenario) CheckPlayerQuestKeysReport(silver, gold int, name string) {
	want := [2]int{silver, gold}
	observe := func(stage string) bool {
		state, err := readE2EPlayerQuestKeysReport()
		if err == nil {
			err = state.validate(want)
		}
		if err != nil {
			if noxServer.Frame()%30 == 0 {
				e2eLog.Printf("QUEST KEY REPORT WAIT: %s: %v", stage, err)
			}
			return false
		}
		unit := noxServer.Players.HostUnit()
		e2eLog.Printf("QUEST KEY REPORT PASS: %s frame=%d unit=%p update=%p player=%p inventory=%v client_inventory=%v cache=%v receiver=%v",
			stage, noxServer.Frame(), unit, unit.UpdateData, unit.ControllingPlayer(), state.inventory, state.clientInventory, state.cached, state.client)
		return true
	}
	sc.addWhen(0, name+" report", 1200, func() bool { return observe(name + " report") }, func() {})
	sc.Wait(12, name+" settle")
	sc.addWhen(0, name+" stable report", 120, func() bool { return observe(name + " stable report") }, func() {})
}
