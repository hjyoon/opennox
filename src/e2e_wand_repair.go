package opennox

import (
	"fmt"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/legacy"
)

// SetInventoryItemCharges prepares a depleted stock wand. The subsequent quote
// and repair must pass through the ordinary shop mouse input and wire protocol.
func (sc *e2eScenario) SetInventoryItemCharges(typeID string, charge int, name string) {
	sc.add(0, name, func() {
		item, _, err := e2eInventoryItem(typeID)
		if err != nil {
			e2eError(err)
			return
		}
		if !item.Class().Has(object.ClassWand) || item.UseData.Ptr == nil {
			e2eError(fmt.Errorf("inventory item %q is not a wand", typeID))
			return
		}
		data := item.UseData.AsWand()
		if charge < 0 || charge >= int(data.MaxCharge) {
			e2eError(fmt.Errorf("charge fixture for %q must be in 0..%d, got %d", typeID, int(data.MaxCharge)-1, charge))
			return
		}
		data.Charge = uint8(charge)
		data.Progress = uint32(charge) * 100 / uint32(data.MaxCharge)
		player := noxServer.Players.HostUnit().UpdateDataPlayer().Player
		legacy.Nox_xxx_netReportCharges_4D82B0(player.PlayerInd, item, data.Charge, data.MaxCharge)
		e2eLog.Printf("WAND CHARGE FIXTURE: item=%s netcode=%d charge=%d/%d progress=%d worth=%d repair_coefficient=%g",
			typeID, item.NetCode, data.Charge, data.MaxCharge, data.Progress, item.Worth, noxServer.Balance.Float("RepairCoefficient"))
	})
}

func (sc *e2eScenario) AssertInventoryItemCharges(typeID string, charge int, full bool, name string) {
	sc.add(0, name, func() {
		item, typeInd, err := e2eInventoryItem(typeID)
		if err != nil {
			e2eError(err)
			return
		}
		if !item.Class().Has(object.ClassWand) || item.UseData.Ptr == nil {
			e2eError(fmt.Errorf("inventory item %q is not a wand", typeID))
			return
		}
		data := item.UseData.AsWand()
		found, column, row, netCode := legacy.Nox_client_inventoryItemLocation(uint32(typeInd))
		drawable := legacy.InventoryCellDrawable(column, row)
		if !found || drawable == nil || netCode != item.NetCode {
			e2eError(fmt.Errorf("client wand %q not found: found=%t cell=(%d,%d) drawable=%p netcode=%d/%d", typeID, found, column, row, drawable, netCode, item.NetCode))
			return
		}
		want := charge
		if full {
			want = int(data.MaxCharge)
		}
		clientData := drawable.UnionItem()
		if want < 0 || want > int(data.MaxCharge) || int(data.Charge) != want || int(clientData.Field_112_0) != want || int(clientData.Field_112_2) != int(data.MaxCharge) || (full && data.Progress != 100) {
			e2eError(fmt.Errorf("inventory charges %q = server:%d/%d progress:%d client:%d/%d, want %d/%d", typeID, data.Charge, data.MaxCharge, data.Progress, clientData.Field_112_0, clientData.Field_112_2, want, data.MaxCharge))
			return
		}
		e2eLog.Printf("INVENTORY CHARGES: item=%s server=%d/%d progress=%d client=%d/%d gold=%d",
			typeID, data.Charge, data.MaxCharge, data.Progress, clientData.Field_112_0, clientData.Field_112_2, noxServer.Players.HostUnit().UpdateDataPlayer().Player.GoldVal)
	})
	sc.CaptureMagicFrame(name)
}
